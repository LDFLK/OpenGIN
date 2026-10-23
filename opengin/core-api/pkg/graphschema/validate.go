// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package graphschema

import (
	"fmt"
	"regexp"
	"strings"
)

// ReservedMajor is the Major kind used for system attribute nodes (engine.DatasetType).
// User schemas may not declare it.
const ReservedMajor = "Dataset"

// ReservedRelationshipName is the relationship used to link entities to their attribute nodes
// (engine.IS_ATTRIBUTE_RELATIONSHIP). User schemas may not declare it.
const ReservedRelationshipName = "IS_ATTRIBUTE"

// identifierPattern restricts Major kinds and relationship names, which become Neo4j labels and
// relationship types and are concatenated into Cypher queries.
// must start with a letter or _, followed only by letters, digits and _
var identifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Violation describes one schema rule that a document breaks. Path locates the offending value
// in the document, e.g. "relationshipDefs[11].endpoints[0].to".
type Violation struct {
	Path    string
	Message string
}

// String formats the violation as "path: message".
func (v Violation) String() string {
	if v.Path == "" {
		return v.Message
	}
	return v.Path + ": " + v.Message
}

// Validate checks a parsed schema against the schema rules and returns every violation found,
// in document order. An empty result means the schema is valid.
func Validate(schema *Schema) []Violation {
	if schema == nil {
		return []Violation{{Path: "", Message: "schema is required"}}
	}

	var violations []Violation
	addViolation := func(path string, format string, args ...interface{}) {
		violations = append(violations, Violation{Path: path, Message: fmt.Sprintf(format, args...)})
	}

	declaredKinds := validateKindDefs(schema.KindDefs, addViolation)
	validateRelationshipDefs(schema.RelationshipDefs, declaredKinds, addViolation)

	return violations
}

type violationFunc func(path string, format string, args ...interface{})

// declaredKinds maps each declared Major kind to the set of its Minor kinds.
type declaredKinds map[string]map[string]bool

// validateKindDefs checks kindDefs and returns the kinds they declare. Every Major that has a
// name is registered, even one that breaks a rule, and duplicate Majors have their Minor kinds
// merged, so that a single mistake is not reported again by every endpoint that uses the kind.
func validateKindDefs(kindDefs []KindDef, addViolation violationFunc) declaredKinds {
	declared := declaredKinds{}
	firstDeclaration := map[string]int{}

	if len(kindDefs) == 0 {
		addViolation("kindDefs", "at least one kind definition is required")
	}

	for i, kindDef := range kindDefs {
		majorPath := fmt.Sprintf("kindDefs[%d].major", i)

		switch {
		case kindDef.Major == "":
			addViolation(majorPath, "major is required")
		case !identifierPattern.MatchString(kindDef.Major):
			addViolation(majorPath, "%q must match %s", kindDef.Major, identifierPattern.String())
		case kindDef.Major == ReservedMajor:
			addViolation(majorPath, "%q is reserved for system attribute nodes", kindDef.Major)
		}

		if kindDef.Major != "" {
			if first, exists := firstDeclaration[kindDef.Major]; exists {
				addViolation(majorPath, "duplicate Major %q; add its minors to kindDefs[%d] instead", kindDef.Major, first)
			} else {
				firstDeclaration[kindDef.Major] = i
				declared[kindDef.Major] = map[string]bool{}
			}
		}

		if len(kindDef.Minors) == 0 {
			addViolation(fmt.Sprintf("kindDefs[%d].minors", i), "at least one minor kind is required")
		}

		seenMinors := map[string]bool{}
		for j, minor := range kindDef.Minors {
			minorPath := fmt.Sprintf("kindDefs[%d].minors[%d]", i, j)

			switch {
			case minor == "":
				addViolation(minorPath, "minor kind must not be empty")
			case strings.Contains(minor, KindSeparator):
				addViolation(minorPath, "minor kind %q must not contain %q", minor, KindSeparator)
			case seenMinors[minor]:
				addViolation(minorPath, "duplicate minor kind %q", minor)
			}

			seenMinors[minor] = true
			if kindDef.Major != "" {
				declared[kindDef.Major][minor] = true
			}
		}
	}

	return declared
}

// validateRelationshipDefs checks relationshipDefs, resolving every endpoint against the
// declared kinds.
func validateRelationshipDefs(relationshipDefs []RelationshipDef, declared declaredKinds, addViolation violationFunc) {
	firstDeclaration := map[string]int{}

	for i, relationshipDef := range relationshipDefs {
		namePath := fmt.Sprintf("relationshipDefs[%d].name", i)

		switch {
		case relationshipDef.Name == "":
			addViolation(namePath, "name is required")
		case !identifierPattern.MatchString(relationshipDef.Name):
			addViolation(namePath, "%q must match %s", relationshipDef.Name, identifierPattern.String())
		case relationshipDef.Name == ReservedRelationshipName:
			addViolation(namePath, "%q is reserved for system attribute relationships", relationshipDef.Name)
		}

		if relationshipDef.Name != "" {
			if first, exists := firstDeclaration[relationshipDef.Name]; exists {
				addViolation(namePath, "duplicate relationship %q; add the pair to relationshipDefs[%d].endpoints instead",
					relationshipDef.Name, first)
			} else {
				firstDeclaration[relationshipDef.Name] = i
			}
		}

		if len(relationshipDef.Endpoints) == 0 {
			addViolation(fmt.Sprintf("relationshipDefs[%d].endpoints", i), "at least one endpoint pair is required")
		}

		seenPairs := map[Endpoint]int{}
		for j, endpoint := range relationshipDef.Endpoints {
			endpointPath := fmt.Sprintf("relationshipDefs[%d].endpoints[%d]", i, j)

			validateKindRef(endpointPath+".from", endpoint.From, declared, addViolation)
			validateKindRef(endpointPath+".to", endpoint.To, declared, addViolation)

			if first, exists := seenPairs[endpoint]; exists {
				addViolation(endpointPath, "duplicate pair %s -> %s; same as endpoints[%d]", endpoint.From, endpoint.To, first)
			} else {
				seenPairs[endpoint] = j
			}
		}
	}
}

// validateKindRef checks that ref is a "Major/Minor" reference to a declared kind. "Major/*"
// matches any minor of a declared Major, and a Major that lists "*" accepts any minor.
func validateKindRef(path string, ref string, declared declaredKinds, addViolation violationFunc) {
	if ref == "" {
		addViolation(path, "kind is required")
		return
	}

	major, minor, err := ParseKindRef(ref)
	if err != nil {
		addViolation(path, "%q is not in Major%sMinor form", ref, KindSeparator)
		return
	}

	minors, majorDeclared := declared[major]
	if !majorDeclared {
		addViolation(path, "unknown kind %s: Major %q is not declared", ref, major)
		return
	}

	if minor == AnyMinor || minors[AnyMinor] || minors[minor] {
		return
	}
	addViolation(path, "unknown kind %s", ref)
}
