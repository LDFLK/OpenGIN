// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package graphschema

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func parseFixture(t *testing.T, name string) *Schema {
	t.Helper()
	schema, err := Parse(readFixture(t, name))
	if err != nil {
		t.Fatalf("failed to parse fixture %s: %v", name, err)
	}
	return schema
}

func TestValidateValidFixtures(t *testing.T) {
	for _, fixture := range []string{"lk_regions_government.json", "valid_wildcards.json"} {
		t.Run(fixture, func(t *testing.T) {
			assert.Empty(t, Validate(parseFixture(t, fixture)))
		})
	}
}

// ruleFixtures lists every fixture in testdata/invalid that parses but breaks one schema rule,
// with the single violation it must produce.
var ruleFixtures = map[string]struct {
	path            string
	messageContains string
}{
	// kindDefs
	"empty_kind_defs.json":    {path: "kindDefs", messageContains: "at least one kind definition is required"},
	"missing_major.json":      {path: "kindDefs[1].major", messageContains: "major is required"},
	"invalid_major_name.json": {path: "kindDefs[0].major", messageContains: `"lk-region" must match`},
	"reserved_major.json":     {path: "kindDefs[1].major", messageContains: `"Dataset" is reserved`},
	"duplicate_major.json":    {path: "kindDefs[1].major", messageContains: `duplicate Major "region"; add its minors to kindDefs[0]`},
	"missing_minors.json":     {path: "kindDefs[1].minors", messageContains: "at least one minor kind is required"},
	"empty_minor.json":        {path: "kindDefs[0].minors[1]", messageContains: "must not be empty"},
	"minor_with_slash.json":   {path: "kindDefs[0].minors[1]", messageContains: `"province/east" must not contain "/"`},
	"duplicate_minor.json":    {path: "kindDefs[0].minors[2]", messageContains: `duplicate minor kind "country"`},

	// relationshipDefs
	"missing_relationship_name.json":  {path: "relationshipDefs[0].name", messageContains: "name is required"},
	"invalid_relationship_name.json":  {path: "relationshipDefs[0].name", messageContains: `"lk-province" must match`},
	"reserved_relationship_name.json": {path: "relationshipDefs[1].name", messageContains: `"IS_ATTRIBUTE" is reserved`},
	"duplicate_relationship_name.json": {
		path:            "relationshipDefs[1].name",
		messageContains: `duplicate relationship "located_in"; add the pair to relationshipDefs[0].endpoints`,
	},
	"empty_endpoints.json": {path: "relationshipDefs[0].endpoints", messageContains: "at least one endpoint pair is required"},
	"duplicate_endpoint_pair.json": {
		path:            "relationshipDefs[0].endpoints[1]",
		messageContains: "duplicate pair region/country -> region/province; same as endpoints[0]",
	},

	// kind references
	"malformed_kind_reference.json": {path: "relationshipDefs[0].endpoints[0].to", messageContains: `"region" is not in Major/Minor form`},
	"unknown_kind_reference.json":   {path: "relationshipDefs[0].endpoints[0].to", messageContains: "unknown kind region/district"},
	"wildcard_unknown_major.json": {
		path:            "relationshipDefs[0].endpoints[0].from",
		messageContains: `unknown kind Organization/*: Major "Organization" is not declared`,
	},
}

// Fixtures in testdata/invalid that Parse rejects, so they never reach Validate.
var parseErrorFixtures = map[string]bool{
	"unknown_top_level_key.json": true,
	"category_field.json":        true,
	"cardinality_field.json":     true,
}

func TestValidateRuleFixtures(t *testing.T) {
	for fixture, expected := range ruleFixtures {
		t.Run(fixture, func(t *testing.T) {
			violations := Validate(parseFixture(t, filepath.Join("invalid", fixture)))
			if assert.Len(t, violations, 1, "expected exactly one violation, got %v", violations) {
				assert.Equal(t, expected.path, violations[0].Path)
				assert.Contains(t, violations[0].Message, expected.messageContains)
			}
		})
	}
}

// Every invalid fixture must be covered by either the parse tests or the rule table above.
func TestValidateRuleFixturesAreComplete(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "invalid", "*.json"))
	if !assert.NoError(t, err) {
		return
	}
	for _, file := range files {
		name := filepath.Base(file)
		_, inRuleTable := ruleFixtures[name]
		assert.True(t, inRuleTable || parseErrorFixtures[name], "fixture %s has no expected result", name)
	}
}

func TestValidateNilSchema(t *testing.T) {
	violations := Validate(nil)
	if assert.Len(t, violations, 1) {
		assert.Equal(t, "schema is required", violations[0].Message)
	}
}

func TestValidateCollectsAllViolationsInOrder(t *testing.T) {
	schema := &Schema{
		KindDefs: []KindDef{
			{Major: "region", Minors: []string{"country", "country"}},
			{Major: "Dataset", Minors: []string{"tabular"}},
		},
		RelationshipDefs: []RelationshipDef{
			{Name: "bad-name", Endpoints: []Endpoint{{From: "region/country", To: "region/district"}}},
			{Name: "empty"},
		},
	}

	violations := Validate(schema)

	var paths []string
	for _, violation := range violations {
		paths = append(paths, violation.Path)
	}
	assert.Equal(t, []string{
		"kindDefs[0].minors[1]",
		"kindDefs[1].major",
		"relationshipDefs[0].name",
		"relationshipDefs[0].endpoints[0].to",
		"relationshipDefs[1].endpoints",
	}, paths)
}

func TestValidateKindReferences(t *testing.T) {
	kindDefs := []KindDef{
		{Major: "Organization", Minors: []string{"Minister", "Department"}},
		{Major: "Person", Minors: []string{AnyMinor}},
	}

	tests := map[string]struct {
		from         string
		to           string
		expectedPath string
		expected     string
	}{
		"declared kinds": {
			from: "Organization/Minister", to: "Organization/Department",
		},
		"wildcard minor kind accepts any minor": {
			from: "Person/Anything", to: "Organization/Minister",
		},
		"wildcard reference to a Major with explicit minors": {
			from: "Person/Citizen", to: "Organization/*",
		},
		"undeclared minor of a declared Major": {
			from: "Organization/Ministry", to: "Organization/Department",
			expectedPath: "relationshipDefs[0].endpoints[0].from", expected: "unknown kind Organization/Ministry",
		},
		"undeclared Major": {
			from: "Organization/Minister", to: "Place/City",
			expectedPath: "relationshipDefs[0].endpoints[0].to", expected: `Major "Place" is not declared`,
		},
		"empty from": {
			from: "", to: "Organization/Minister",
			expectedPath: "relationshipDefs[0].endpoints[0].from", expected: "kind is required",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			schema := &Schema{
				KindDefs: kindDefs,
				RelationshipDefs: []RelationshipDef{
					{Name: "linked_to", Endpoints: []Endpoint{{From: tc.from, To: tc.to}}},
				},
			}
			violations := Validate(schema)
			if tc.expected == "" {
				assert.Empty(t, violations)
				return
			}
			if assert.Len(t, violations, 1) {
				assert.Equal(t, tc.expectedPath, violations[0].Path)
				assert.Contains(t, violations[0].Message, tc.expected)
			}
		})
	}
}

func TestValidateAllowsSchemaWithoutRelationships(t *testing.T) {
	schema := &Schema{KindDefs: []KindDef{{Major: "region", Minors: []string{"country"}}}}
	assert.Empty(t, Validate(schema))
}

func TestViolationString(t *testing.T) {
	assert.Equal(t, "kindDefs[0].major: major is required",
		Violation{Path: "kindDefs[0].major", Message: "major is required"}.String())
	assert.Equal(t, "schema is required", Violation{Message: "schema is required"}.String())
}
