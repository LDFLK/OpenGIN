// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package graphschema

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", name, err)
	}
	return data
}

func TestParseRegionsGovernmentFixture(t *testing.T) {
	schema, err := Parse(readFixture(t, "lk_regions_government.json"))
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "Sri Lanka regions and central government", schema.Description)

	// 3 Majors declaring 13 kinds in total
	assert.Len(t, schema.KindDefs, 3)
	kindCount := 0
	for _, kindDef := range schema.KindDefs {
		kindCount += len(kindDef.Minors)
	}
	assert.Equal(t, 13, kindCount)
	assert.Equal(t, KindDef{Major: "Organization", Minors: []string{"Government", "Minister", "Department"}}, schema.KindDefs[1])

	// 12 relationship names covering 13 endpoint pairs
	assert.Len(t, schema.RelationshipDefs, 12)
	pairCount := 0
	for _, relationshipDef := range schema.RelationshipDefs {
		pairCount += len(relationshipDef.Endpoints)
	}
	assert.Equal(t, 13, pairCount)

	assert.Equal(t, RelationshipDef{
		Name: "province",
		Endpoints: []Endpoint{
			{From: "region/country", To: "region/country-level-1+lk-province"},
		},
	}, schema.RelationshipDefs[0])

	// one name, several pairs
	assert.Equal(t, RelationshipDef{
		Name: "located_in",
		Endpoints: []Endpoint{
			{From: "Organization/Department", To: "region/country-level-2+lk-administrative-district"},
			{From: "Organization/Government", To: "region/country"},
		},
	}, schema.RelationshipDefs[11])
}

func TestParseWildcardsFixture(t *testing.T) {
	schema, err := Parse(readFixture(t, "valid_wildcards.json"))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []string{AnyMinor}, schema.KindDefs[1].Minors)
	assert.Equal(t, Endpoint{From: "Person/*", To: "Organization/*"}, schema.RelationshipDefs[0].Endpoints[0])
}

func TestParseRejectsUnknownFields(t *testing.T) {
	tests := map[string]struct {
		fixture      string
		unknownField string
	}{
		"hierarchical loader document": {fixture: "tree_with_loader_keys.json", unknownField: "major"},
		"unknown top-level key":        {fixture: "invalid/unknown_top_level_key.json", unknownField: "namespace"},
		"category on a relationship":   {fixture: "invalid/category_field.json", unknownField: "category"},
		"cardinality on an endpoint":   {fixture: "invalid/cardinality_field.json", unknownField: "cardinality"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			schema, err := Parse(readFixture(t, tc.fixture))
			assert.Nil(t, schema)
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), `unknown field "`+tc.unknownField+`"`)
			}
		})
	}
}

// The remaining invalid fixtures break schema rules, not the document format, so they must
// parse successfully and be left for Validate to reject.
func TestParseAcceptsRuleViolationFixtures(t *testing.T) {
	parseErrorFixtures := map[string]bool{
		"unknown_top_level_key.json": true,
		"category_field.json":        true,
		"cardinality_field.json":     true,
	}

	files, err := filepath.Glob(filepath.Join("testdata", "invalid", "*.json"))
	if !assert.NoError(t, err) || !assert.NotEmpty(t, files) {
		return
	}

	for _, file := range files {
		name := filepath.Base(file)
		if parseErrorFixtures[name] {
			continue
		}
		t.Run(name, func(t *testing.T) {
			_, err := Parse(readFixture(t, filepath.Join("invalid", name)))
			assert.NoError(t, err)
		})
	}
}

func TestParseMissingFieldsAreLeftEmpty(t *testing.T) {
	schema, err := Parse(readFixture(t, "invalid/missing_major.json"))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "", schema.KindDefs[1].Major)
	assert.Equal(t, []string{"Department"}, schema.KindDefs[1].Minors)
}

func TestParseMalformedInput(t *testing.T) {
	tests := map[string]struct {
		input         string
		errorContains string
	}{
		"empty input": {
			input:         "",
			errorContains: "document is empty",
		},
		"whitespace only": {
			input:         "  \n\t ",
			errorContains: "document is empty",
		},
		"truncated JSON": {
			input:         `{"kindDefs": [`,
			errorContains: "invalid graph schema",
		},
		"top-level array": {
			input:         `[{"major": "region"}]`,
			errorContains: "cannot unmarshal array",
		},
		"wrong field type": {
			input:         `{"kindDefs": [{"major": "region", "minors": "country"}]}`,
			errorContains: "cannot unmarshal string",
		},
		"second JSON object": {
			input:         `{"kindDefs": []} {"kindDefs": []}`,
			errorContains: "unexpected data after the JSON object",
		},
		"trailing garbage": {
			input:         `{"kindDefs": []} x`,
			errorContains: "unexpected data after the JSON object",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			schema, err := Parse([]byte(tc.input))
			assert.Nil(t, schema)
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), tc.errorContains)
			}
		})
	}
}

func TestParseAllowsTrailingWhitespace(t *testing.T) {
	_, err := Parse([]byte("{\"kindDefs\": [], \"relationshipDefs\": []}\n\n  "))
	assert.NoError(t, err)
}
