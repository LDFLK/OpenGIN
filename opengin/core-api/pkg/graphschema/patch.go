// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package graphschema

import (
	"fmt"
	"slices"
	"sort"
)

// Patch is an atomic set of changes to a schema: everything in it is applied together, or the
// patch is rejected as a whole.
type Patch struct {
	// BaseVersion is the schema version the patch was written against. Apply does not use it;
	// the caller checks it against the latest stored version.
	BaseVersion int `json:"baseVersion"`
	// Description, when set, replaces the schema's description.
	Description string         `json:"description,omitempty"`
	Add         PatchAdditions `json:"add"`
	Remove      PatchRemovals  `json:"remove"`
}

// PatchAdditions lists what a patch adds. A KindDef whose Major already exists adds its Minor
// kinds to that Major, and a RelationshipDef whose name already exists adds its endpoint pairs
// to that relationship.
type PatchAdditions struct {
	KindDefs         []KindDef         `json:"kindDefs,omitempty"`
	RelationshipDefs []RelationshipDef `json:"relationshipDefs,omitempty"`
}

// PatchRemovals lists what a patch removes: kinds as "Major/Minor" references, whole
// relationships by name, and individual endpoint pairs.
type PatchRemovals struct {
	Kinds         []string      `json:"kinds,omitempty"`
	Relationships []string      `json:"relationships,omitempty"`
	Endpoints     []EndpointRef `json:"endpoints,omitempty"`
}

// EndpointRef identifies one endpoint pair of a named relationship.
type EndpointRef struct {
	Name string `json:"name"`
	From string `json:"from"`
	To   string `json:"to"`
}

// Apply applies a patch to a copy of base and returns the resulting schema. The base schema is
// never modified.
//
// Removals are applied first (endpoint pairs, then relationships, then kinds), followed by
// additions, and the result is checked with Validate. If anything is wrong, Apply returns a nil
// schema and every violation found: patch violations have paths into the patch (e.g.
// "remove.kinds[0]"), and violations from Validate have paths into the resulting schema.
func Apply(base *Schema, patch *Patch) (*Schema, []Violation) {
	if base == nil {
		return nil, []Violation{{Path: "", Message: "there is no schema to patch"}}
	}
	if patch == nil {
		return nil, []Violation{{Path: "", Message: "patch is required"}}
	}

	p := &patcher{schema: cloneSchema(base), emptiedBy: map[string]int{}}

	p.removeEndpoints(patch.Remove.Endpoints)
	p.removeRelationships(patch.Remove.Relationships)
	p.removeKinds(patch.Remove.Kinds)
	p.addKindDefs(patch.Add.KindDefs)
	p.addRelationshipDefs(patch.Add.RelationshipDefs)
	p.checkEmptiedRelationships()

	if patch.Description != "" {
		p.schema.Description = patch.Description
	}

	violations := append(p.violations, Validate(p.schema)...)
	if len(violations) > 0 {
		return nil, violations
	}
	return p.schema, nil
}

// patcher holds the schema being patched and the violations found so far.
type patcher struct {
	schema     *Schema
	violations []Violation
	// emptiedBy records, for each relationship whose endpoint pairs were all removed, the index
	// of the last remove.endpoints entry that removed one of them.
	emptiedBy map[string]int
}

func (p *patcher) addViolation(path string, format string, args ...interface{}) {
	p.violations = append(p.violations, Violation{Path: path, Message: fmt.Sprintf(format, args...)})
}

func (p *patcher) findRelationship(name string) int {
	for i, relationshipDef := range p.schema.RelationshipDefs {
		if relationshipDef.Name == name {
			return i
		}
	}
	return -1
}

func (p *patcher) findKindDef(major string) int {
	for i, kindDef := range p.schema.KindDefs {
		if kindDef.Major == major {
			return i
		}
	}
	return -1
}

func (p *patcher) removeEndpoints(refs []EndpointRef) {
	for k, ref := range refs {
		path := fmt.Sprintf("remove.endpoints[%d]", k)

		i := p.findRelationship(ref.Name)
		if i < 0 {
			p.addViolation(path, "unknown relationship %q", ref.Name)
			continue
		}

		relationshipDef := &p.schema.RelationshipDefs[i]
		pair := Endpoint{From: ref.From, To: ref.To}
		j := slices.Index(relationshipDef.Endpoints, pair)
		if j < 0 {
			p.addViolation(path, "relationship %q has no pair %s -> %s", ref.Name, ref.From, ref.To)
			continue
		}

		relationshipDef.Endpoints = append(relationshipDef.Endpoints[:j], relationshipDef.Endpoints[j+1:]...)
		if len(relationshipDef.Endpoints) == 0 {
			p.emptiedBy[ref.Name] = k
		}
	}
}

func (p *patcher) removeRelationships(names []string) {
	for k, name := range names {
		i := p.findRelationship(name)
		if i < 0 {
			p.addViolation(fmt.Sprintf("remove.relationships[%d]", k), "unknown relationship %q", name)
			continue
		}
		p.schema.RelationshipDefs = append(p.schema.RelationshipDefs[:i], p.schema.RelationshipDefs[i+1:]...)
	}
}

// removeKinds removes kinds that no remaining endpoint uses. A kind that is still used is kept,
// so the only violation reported for it is the one that says why.
func (p *patcher) removeKinds(refs []string) {
	for k, ref := range refs {
		path := fmt.Sprintf("remove.kinds[%d]", k)

		major, minor, err := ParseKindRef(ref)
		if err != nil {
			p.addViolation(path, "%q is not in Major%sMinor form", ref, KindSeparator)
			continue
		}

		i := p.findKindDef(major)
		if i < 0 || slices.Index(p.schema.KindDefs[i].Minors, minor) < 0 {
			p.addViolation(path, "unknown kind %s", ref)
			continue
		}

		dropsMajor := len(p.schema.KindDefs[i].Minors) == 1
		if user, found := p.findEndpointUsing(ref, major, dropsMajor); found {
			p.addViolation(path, "kind %s is still used by %s; remove that pair or relationship in the same patch", ref, user)
			continue
		}

		kindDef := &p.schema.KindDefs[i]
		j := slices.Index(kindDef.Minors, minor)
		kindDef.Minors = append(kindDef.Minors[:j], kindDef.Minors[j+1:]...)
		if len(kindDef.Minors) == 0 {
			p.schema.KindDefs = append(p.schema.KindDefs[:i], p.schema.KindDefs[i+1:]...)
		}
	}
}

// findEndpointUsing describes the first endpoint that references ref. When removing ref would
// drop its whole Major, a "Major/*" reference also counts as a use.
func (p *patcher) findEndpointUsing(ref string, major string, dropsMajor bool) (string, bool) {
	wildcard := FormatKindRef(major, AnyMinor)
	for _, relationshipDef := range p.schema.RelationshipDefs {
		for _, endpoint := range relationshipDef.Endpoints {
			for _, side := range []string{endpoint.From, endpoint.To} {
				if side == ref || (dropsMajor && side == wildcard) {
					return fmt.Sprintf("relationship %q pair %s -> %s", relationshipDef.Name, endpoint.From, endpoint.To), true
				}
			}
		}
	}
	return "", false
}

func (p *patcher) addKindDefs(kindDefs []KindDef) {
	for k, kindDef := range kindDefs {
		i := p.findKindDef(kindDef.Major)
		if kindDef.Major == "" || i < 0 {
			// A new Major (or one with no name, which Validate reports).
			p.schema.KindDefs = append(p.schema.KindDefs, KindDef{
				Major:  kindDef.Major,
				Minors: slices.Clone(kindDef.Minors),
			})
			continue
		}

		existing := &p.schema.KindDefs[i]
		for j, minor := range kindDef.Minors {
			if slices.Index(existing.Minors, minor) >= 0 {
				p.addViolation(fmt.Sprintf("add.kindDefs[%d].minors[%d]", k, j),
					"kind %s is already declared", FormatKindRef(kindDef.Major, minor))
				continue
			}
			existing.Minors = append(existing.Minors, minor)
		}
	}
}

func (p *patcher) addRelationshipDefs(relationshipDefs []RelationshipDef) {
	for k, relationshipDef := range relationshipDefs {
		i := p.findRelationship(relationshipDef.Name)
		if relationshipDef.Name == "" || i < 0 {
			// A new relationship (or one with no name, which Validate reports).
			p.schema.RelationshipDefs = append(p.schema.RelationshipDefs, RelationshipDef{
				Name:      relationshipDef.Name,
				Endpoints: slices.Clone(relationshipDef.Endpoints),
			})
			continue
		}

		existing := &p.schema.RelationshipDefs[i]
		for j, endpoint := range relationshipDef.Endpoints {
			if slices.Index(existing.Endpoints, endpoint) >= 0 {
				p.addViolation(fmt.Sprintf("add.relationshipDefs[%d].endpoints[%d]", k, j),
					"relationship %q already has the pair %s -> %s", relationshipDef.Name, endpoint.From, endpoint.To)
				continue
			}
			existing.Endpoints = append(existing.Endpoints, endpoint)
		}
	}
}

// checkEmptiedRelationships reports relationships left with no endpoint pairs because the patch
// removed them all, without removing the relationship itself or adding a replacement pair. The
// empty relationship is dropped from the result so Validate does not report it a second time.
func (p *patcher) checkEmptiedRelationships() {
	names := make([]string, 0, len(p.emptiedBy))
	for name := range p.emptiedBy {
		names = append(names, name)
	}
	// report in patch order
	sort.Slice(names, func(a, b int) bool { return p.emptiedBy[names[a]] < p.emptiedBy[names[b]] })

	for _, name := range names {
		k := p.emptiedBy[name]
		i := p.findRelationship(name)
		if i < 0 || len(p.schema.RelationshipDefs[i].Endpoints) > 0 {
			continue
		}
		p.addViolation(fmt.Sprintf("remove.endpoints[%d]", k),
			"removes the last pair of relationship %q; use remove.relationships instead", name)
		p.schema.RelationshipDefs = append(p.schema.RelationshipDefs[:i], p.schema.RelationshipDefs[i+1:]...)
	}
}

// cloneSchema returns a deep copy of s, so patching never changes the caller's schema.
func cloneSchema(s *Schema) *Schema {
	clone := &Schema{Description: s.Description}
	for _, kindDef := range s.KindDefs {
		clone.KindDefs = append(clone.KindDefs, KindDef{
			Major:  kindDef.Major,
			Minors: slices.Clone(kindDef.Minors),
		})
	}
	for _, relationshipDef := range s.RelationshipDefs {
		clone.RelationshipDefs = append(clone.RelationshipDefs, RelationshipDef{
			Name:      relationshipDef.Name,
			Endpoints: slices.Clone(relationshipDef.Endpoints),
		})
	}
	return clone
}
