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
