// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

// Package graphschema defines the user-supplied schema that describes the shape of the
// entity graph: which kinds (Major/Minor) exist and which relationships may connect them.
//
// The same types are used for the document users write (YAML or JSON) and for the form
// stored in MongoDB, so there is no conversion step between the two.
package graphschema

import (
	"fmt"
	"strings"
)

// KindSeparator separates the Major and Minor parts of a kind reference, e.g. "region/country".
const KindSeparator = "/"

// AnyMinor is the wildcard Minor kind. In KindDef.Minors it allows any minor kind under the
// Major, and in a kind reference ("Organization/*") it matches every minor kind of that Major.
const AnyMinor = "*"

// Schema is the complete graph schema for a deployment.
type Schema struct {
	Description      string            `yaml:"description,omitempty" json:"description,omitempty" bson:"description,omitempty"`
	KindDefs         []KindDef         `yaml:"kindDefs" json:"kindDefs" bson:"kindDefs"`
	RelationshipDefs []RelationshipDef `yaml:"relationshipDefs" json:"relationshipDefs" bson:"relationshipDefs"`
}

// KindDef declares a Major kind and the Minor kinds allowed under it.
type KindDef struct {
	Major  string   `yaml:"major" json:"major" bson:"major"`
	Minors []string `yaml:"minors" json:"minors" bson:"minors"`
}

// RelationshipDef declares a relationship name and every (from, to) kind pair it may connect.
// A name is declared once; additional valid pairs are added to its Endpoints.
type RelationshipDef struct {
	Name      string     `yaml:"name" json:"name" bson:"name"`
	Endpoints []Endpoint `yaml:"endpoints" json:"endpoints" bson:"endpoints"`
}

// Endpoint is one allowed (from, to) pair for a relationship. Both sides are kind references
// in "Major/Minor" form. From is the entity that holds the relationship and To is its
// relatedEntityId, matching the direction edges are created in Neo4j.
type Endpoint struct {
	From string `yaml:"from" json:"from" bson:"from"`
	To   string `yaml:"to" json:"to" bson:"to"`
}

// ParseKindRef splits a kind reference such as "region/country-level-1+lk-province" into its
// Major and Minor parts. It splits on the first separator, so only the Major part is guaranteed
// to be free of "/". Both parts must be non-empty.
func ParseKindRef(ref string) (major string, minor string, err error) {
	major, minor, found := strings.Cut(ref, KindSeparator)
	if !found {
		return "", "", fmt.Errorf("invalid kind reference %q: expected Major%sMinor", ref, KindSeparator)
	}
	if major == "" {
		return "", "", fmt.Errorf("invalid kind reference %q: Major is empty", ref)
	}
	if minor == "" {
		return "", "", fmt.Errorf("invalid kind reference %q: Minor is empty", ref)
	}
	return major, minor, nil
}

// FormatKindRef joins a Major and Minor kind into a kind reference, e.g. "region/country".
func FormatKindRef(major string, minor string) string {
	return major + KindSeparator + minor
}
