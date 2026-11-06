// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package graphschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseKindRef(t *testing.T) {
	tests := map[string]struct {
		input         string
		expectedMajor string
		expectedMinor string
		expectError   bool
	}{
		"simple kind": {
			input:         "Organization/Department",
			expectedMajor: "Organization",
			expectedMinor: "Department",
		},
		"minor with plus and dashes": {
			input:         "region/country-level-1+lk-province",
			expectedMajor: "region",
			expectedMinor: "country-level-1+lk-province",
		},
		"wildcard minor": {
			input:         "Person/*",
			expectedMajor: "Person",
			expectedMinor: AnyMinor,
		},
		"splits on first separator only": {
			input:         "region/a/b",
			expectedMajor: "region",
			expectedMinor: "a/b",
		},
		"missing separator": {
			input:       "region",
			expectError: true,
		},
		"empty major": {
			input:       "/country",
			expectError: true,
		},
		"empty minor": {
			input:       "region/",
			expectError: true,
		},
		"empty string": {
			input:       "",
			expectError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			major, minor, err := ParseKindRef(tc.input)
			if tc.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedMajor, major)
			assert.Equal(t, tc.expectedMinor, minor)
		})
	}
}

func TestFormatKindRef(t *testing.T) {
	assert.Equal(t, "region/country", FormatKindRef("region", "country"))
	assert.Equal(t, "Person/*", FormatKindRef("Person", AnyMinor))
}

func TestKindRefRoundTrip(t *testing.T) {
	refs := []string{
		"region/country",
		"region/country-level-3+lk-divisional-secretariat-division",
		"Organization/Minister",
	}
	for _, ref := range refs {
		major, minor, err := ParseKindRef(ref)
		assert.NoError(t, err)
		assert.Equal(t, ref, FormatKindRef(major, minor))
	}
}

func TestAllowedPairs(t *testing.T) {
	schema := parseFixture(t, "lk_regions_government.json")

	tests := map[string]struct {
		name     string
		expected []Endpoint
	}{
		"single pair": {
			name: "province",
			expected: []Endpoint{
				{From: "region/country", To: "region/country-level-1+lk-province"},
			},
		},
		"several pairs for one name": {
			name: "located_in",
			expected: []Endpoint{
				{From: "Organization/Department", To: "region/country-level-2+lk-administrative-district"},
				{From: "Organization/Government", To: "region/country"},
			},
		},
		"undeclared name": {
			name:     "not_declared",
			expected: nil,
		},
		"names are case-sensitive": {
			name:     "Located_In",
			expected: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, schema.AllowedPairs(tc.name))
		})
	}
}

func TestAllowedPairsReturnsCopy(t *testing.T) {
	schema := parseFixture(t, "lk_regions_government.json")

	pairs := schema.AllowedPairs("located_in")
	pairs[0] = Endpoint{From: "changed", To: "changed"}

	assert.Equal(t, "Organization/Department", schema.AllowedPairs("located_in")[0].From)
}

func TestAllowedPairsNilSchema(t *testing.T) {
	var schema *Schema
	assert.Nil(t, schema.AllowedPairs("province"))
}
