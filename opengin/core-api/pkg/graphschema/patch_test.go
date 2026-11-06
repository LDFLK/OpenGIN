// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package graphschema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Values from testdata/lk_regions_government.json.
const (
	fixtureRelationships   = 12
	localGovernmentKind    = "region/country-level-3+lk-local-government"
	administrativeDistrict = "region/country-level-2+lk-administrative-district"
)

func findKindDef(schema *Schema, major string) *KindDef {
	for i := range schema.KindDefs {
		if schema.KindDefs[i].Major == major {
			return &schema.KindDefs[i]
		}
	}
	return nil
}

func findRelationshipDef(schema *Schema, name string) *RelationshipDef {
	for i := range schema.RelationshipDefs {
		if schema.RelationshipDefs[i].Name == name {
			return &schema.RelationshipDefs[i]
		}
	}
	return nil
}

func TestApplyPatches(t *testing.T) {
	tests := map[string]struct {
		patch Patch
		check func(t *testing.T, result *Schema)
	}{
		"empty patch leaves the schema unchanged": {
			patch: Patch{},
			check: func(t *testing.T, result *Schema) {
				assert.Equal(t, parseFixture(t, "lk_regions_government.json"), result)
			},
		},
		"description replaces the schema description": {
			patch: Patch{Description: "Updated"},
			check: func(t *testing.T, result *Schema) {
				assert.Equal(t, "Updated", result.Description)
			},
		},
		"minors are merged into an existing Major": {
			patch: Patch{Add: PatchAdditions{KindDefs: []KindDef{
				{Major: "Organization", Minors: []string{"Agency", "Board"}},
			}}},
			check: func(t *testing.T, result *Schema) {
				assert.Len(t, result.KindDefs, 3)
				assert.Equal(t, []string{"Government", "Minister", "Department", "Agency", "Board"},
					findKindDef(result, "Organization").Minors)
			},
		},
		"a new Major is appended": {
			patch: Patch{Add: PatchAdditions{KindDefs: []KindDef{
				{Major: "Place", Minors: []string{"City"}},
			}}},
			check: func(t *testing.T, result *Schema) {
				assert.Len(t, result.KindDefs, 4)
				assert.Equal(t, KindDef{Major: "Place", Minors: []string{"City"}}, result.KindDefs[3])
			},
		},
		"pairs are merged into an existing relationship": {
			patch: Patch{Add: PatchAdditions{RelationshipDefs: []RelationshipDef{
				{Name: "located_in", Endpoints: []Endpoint{{From: "Organization/Minister", To: "region/country"}}},
			}}},
			check: func(t *testing.T, result *Schema) {
				assert.Len(t, result.RelationshipDefs, fixtureRelationships)
				assert.Equal(t, []Endpoint{
					{From: "Organization/Department", To: administrativeDistrict},
					{From: "Organization/Government", To: "region/country"},
					{From: "Organization/Minister", To: "region/country"},
				}, result.AllowedPairs("located_in"))
			},
		},
		"a new relationship is appended": {
			patch: Patch{Add: PatchAdditions{RelationshipDefs: []RelationshipDef{
				{Name: "reports_to", Endpoints: []Endpoint{{From: "Organization/Department", To: "Organization/Minister"}}},
			}}},
			check: func(t *testing.T, result *Schema) {
				assert.Len(t, result.RelationshipDefs, fixtureRelationships+1)
				assert.Equal(t, "reports_to", result.RelationshipDefs[fixtureRelationships].Name)
			},
		},
		"one of several pairs is removed": {
			patch: Patch{Remove: PatchRemovals{Endpoints: []EndpointRef{
				{Name: "located_in", From: "Organization/Government", To: "region/country"},
			}}},
			check: func(t *testing.T, result *Schema) {
				assert.Equal(t, []Endpoint{{From: "Organization/Department", To: administrativeDistrict}},
					result.AllowedPairs("located_in"))
			},
		},
		"a relationship and the kind only it used are removed together": {
			patch: Patch{Remove: PatchRemovals{
				Kinds:         []string{localGovernmentKind},
				Relationships: []string{"local_government"},
			}},
			check: func(t *testing.T, result *Schema) {
				assert.Len(t, result.RelationshipDefs, fixtureRelationships-1)
				assert.Nil(t, findRelationshipDef(result, "local_government"))
				assert.NotContains(t, findKindDef(result, "region").Minors, "country-level-3+lk-local-government")
			},
		},
		"removing a Major's last minor drops the Major": {
			patch: Patch{Remove: PatchRemovals{
				Kinds:         []string{"Person/Citizen"},
				Relationships: []string{"appointed_minister"},
			}},
			check: func(t *testing.T, result *Schema) {
				assert.Len(t, result.KindDefs, 2)
				assert.Nil(t, findKindDef(result, "Person"))
			},
		},
		"the only pair can be replaced in the same patch": {
			patch: Patch{
				Remove: PatchRemovals{Endpoints: []EndpointRef{
					{Name: "has_department", From: "Organization/Minister", To: "Organization/Department"},
				}},
				Add: PatchAdditions{RelationshipDefs: []RelationshipDef{
					{Name: "has_department", Endpoints: []Endpoint{{From: "Organization/Government", To: "Organization/Department"}}},
				}},
			},
			check: func(t *testing.T, result *Schema) {
				assert.Equal(t, []Endpoint{{From: "Organization/Government", To: "Organization/Department"}},
					result.AllowedPairs("has_department"))
			},
		},
		"removing a relationship's pairs and the relationship together": {
			patch: Patch{Remove: PatchRemovals{
				Endpoints:     []EndpointRef{{Name: "has_department", From: "Organization/Minister", To: "Organization/Department"}},
				Relationships: []string{"has_department"},
			}},
			check: func(t *testing.T, result *Schema) {
				assert.Nil(t, findRelationshipDef(result, "has_department"))
			},
		},
		"plan example: removals and additions in one patch": {
			patch: Patch{
				Description: "Drop local government; add a second province pair",
				Add: PatchAdditions{
					KindDefs: []KindDef{{Major: "region", Minors: []string{"country-level-1+lk-western-province-special"}}},
					RelationshipDefs: []RelationshipDef{{Name: "province", Endpoints: []Endpoint{
						{From: "region/country", To: "region/country-level-1+lk-western-province-special"},
					}}},
				},
				Remove: PatchRemovals{
					Kinds:         []string{localGovernmentKind},
					Relationships: []string{"local_government"},
					Endpoints:     []EndpointRef{{Name: "located_in", From: "Organization/Government", To: "region/country"}},
				},
			},
			check: func(t *testing.T, result *Schema) {
				assert.Len(t, result.AllowedPairs("province"), 2)
				assert.Len(t, result.AllowedPairs("located_in"), 1)
				assert.Nil(t, findRelationshipDef(result, "local_government"))
				region := findKindDef(result, "region").Minors
				assert.Len(t, region, 9)
				assert.Contains(t, region, "country-level-1+lk-western-province-special")
				assert.NotContains(t, region, "country-level-3+lk-local-government")
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, violations := Apply(parseFixture(t, "lk_regions_government.json"), &tc.patch)
			if assert.Empty(t, violations) && assert.NotNil(t, result) {
				tc.check(t, result)
			}
		})
	}
}

func TestApplyViolations(t *testing.T) {
	tests := map[string]struct {
		patch    Patch
		expected []Violation // Message is matched as a substring
	}{
		"removing a pair of an unknown relationship": {
			patch: Patch{Remove: PatchRemovals{Endpoints: []EndpointRef{{Name: "missing", From: "a/b", To: "c/d"}}}},
			expected: []Violation{
				{Path: "remove.endpoints[0]", Message: `unknown relationship "missing"`},
			},
		},
		"removing a pair the relationship does not have": {
			patch: Patch{Remove: PatchRemovals{Endpoints: []EndpointRef{
				{Name: "located_in", From: "Person/Citizen", To: "region/country"},
			}}},
			expected: []Violation{
				{Path: "remove.endpoints[0]", Message: `relationship "located_in" has no pair Person/Citizen -> region/country`},
			},
		},
		"removing an unknown relationship": {
			patch: Patch{Remove: PatchRemovals{Relationships: []string{"missing"}}},
			expected: []Violation{
				{Path: "remove.relationships[0]", Message: `unknown relationship "missing"`},
			},
		},
		"removing an unknown kind": {
			patch: Patch{Remove: PatchRemovals{Kinds: []string{"region/district"}}},
			expected: []Violation{
				{Path: "remove.kinds[0]", Message: "unknown kind region/district"},
			},
		},
		"removing a malformed kind reference": {
			patch: Patch{Remove: PatchRemovals{Kinds: []string{"region"}}},
			expected: []Violation{
				{Path: "remove.kinds[0]", Message: `"region" is not in Major/Minor form`},
			},
		},
		"removing a kind that is still used": {
			patch: Patch{Remove: PatchRemovals{Kinds: []string{"region/country-level-3+lk-medical-officer-of-health"}}},
			expected: []Violation{
				{
					Path: "remove.kinds[0]",
					Message: `kind region/country-level-3+lk-medical-officer-of-health is still used by relationship ` +
						`"medical_officer_of_health"`,
				},
			},
		},
		"removing the last pair of a relationship": {
			patch: Patch{Remove: PatchRemovals{Endpoints: []EndpointRef{
				{Name: "has_department", From: "Organization/Minister", To: "Organization/Department"},
			}}},
			expected: []Violation{
				{Path: "remove.endpoints[0]", Message: `removes the last pair of relationship "has_department"; use remove.relationships instead`},
			},
		},
		"adding a pair the relationship already has": {
			patch: Patch{Add: PatchAdditions{RelationshipDefs: []RelationshipDef{
				{Name: "province", Endpoints: []Endpoint{{From: "region/country", To: "region/country-level-1+lk-province"}}},
			}}},
			expected: []Violation{
				{Path: "add.relationshipDefs[0].endpoints[0]", Message: `relationship "province" already has the pair`},
			},
		},
		"adding a kind that is already declared": {
			patch: Patch{Add: PatchAdditions{KindDefs: []KindDef{{Major: "Person", Minors: []string{"Citizen"}}}}},
			expected: []Violation{
				{Path: "add.kindDefs[0].minors[0]", Message: "kind Person/Citizen is already declared"},
			},
		},
		"the result is checked with Validate": {
			patch: Patch{Add: PatchAdditions{RelationshipDefs: []RelationshipDef{
				{Name: "located_in_city", Endpoints: []Endpoint{{From: "Organization/Department", To: "Place/City"}}},
			}}},
			expected: []Violation{
				{Path: "relationshipDefs[12].endpoints[0].to", Message: `Major "Place" is not declared`},
			},
		},
		"all violations are reported together": {
			patch: Patch{
				Add: PatchAdditions{KindDefs: []KindDef{{Major: "Person", Minors: []string{"Citizen"}}}},
				Remove: PatchRemovals{
					Kinds:         []string{"region/district"},
					Relationships: []string{"missing"},
				},
			},
			expected: []Violation{
				{Path: "remove.relationships[0]", Message: `unknown relationship "missing"`},
				{Path: "remove.kinds[0]", Message: "unknown kind region/district"},
				{Path: "add.kindDefs[0].minors[0]", Message: "kind Person/Citizen is already declared"},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, violations := Apply(parseFixture(t, "lk_regions_government.json"), &tc.patch)
			assert.Nil(t, result)
			if assert.Len(t, violations, len(tc.expected), "got %v", violations) {
				for i, expected := range tc.expected {
					assert.Equal(t, expected.Path, violations[i].Path)
					assert.Contains(t, violations[i].Message, expected.Message)
				}
			}
		})
	}
}

func TestApplyKindRemovalWithWildcardReference(t *testing.T) {
	base := &Schema{
		KindDefs: []KindDef{
			{Major: "region", Minors: []string{"country"}},
			{Major: "Organization", Minors: []string{"Minister", "Department"}},
		},
		RelationshipDefs: []RelationshipDef{
			{Name: "located_in", Endpoints: []Endpoint{{From: "Organization/*", To: "region/country"}}},
		},
	}

	// Organization/* still matches Organization/Department, so Minister can go.
	result, violations := Apply(base, &Patch{Remove: PatchRemovals{Kinds: []string{"Organization/Minister"}}})
	if assert.Empty(t, violations) {
		assert.Equal(t, []string{"Department"}, findKindDef(result, "Organization").Minors)
	}

	// Removing both minors would drop the Major that Organization/* refers to.
	result, violations = Apply(base, &Patch{Remove: PatchRemovals{
		Kinds: []string{"Organization/Minister", "Organization/Department"},
	}})
	assert.Nil(t, result)
	if assert.Len(t, violations, 1, "got %v", violations) {
		assert.Equal(t, "remove.kinds[1]", violations[0].Path)
		assert.Contains(t, violations[0].Message,
			`kind Organization/Department is still used by relationship "located_in" pair Organization/* -> region/country`)
	}
}

func TestApplyReplacingLastMinorUsedByWildcard(t *testing.T) {
	newBase := func() *Schema {
		return &Schema{
			KindDefs: []KindDef{
				{Major: "Organization", Minors: []string{"Minister"}},
				{Major: "region", Minors: []string{"country"}},
			},
			RelationshipDefs: []RelationshipDef{
				{Name: "located_in", Endpoints: []Endpoint{{From: "Organization/*", To: "region/country"}}},
			},
		}
	}

	t.Run("replacement minor added in the same patch", func(t *testing.T) {
		result, violations := Apply(newBase(), &Patch{
			Remove: PatchRemovals{Kinds: []string{"Organization/Minister"}},
			Add:    PatchAdditions{KindDefs: []KindDef{{Major: "Organization", Minors: []string{"Ministry"}}}},
		})
		if assert.Empty(t, violations) {
			assert.Equal(t, []string{"Ministry"}, findKindDef(result, "Organization").Minors)
			assert.Equal(t, []Endpoint{{From: "Organization/*", To: "region/country"}}, result.AllowedPairs("located_in"))
		}
	})

	t.Run("no replacement minor", func(t *testing.T) {
		result, violations := Apply(newBase(), &Patch{
			Remove: PatchRemovals{Kinds: []string{"Organization/Minister"}},
		})
		assert.Nil(t, result)
		// Only the patch violation: Validate must not also report Organization/* as unknown.
		if assert.Len(t, violations, 1, "got %v", violations) {
			assert.Equal(t, "remove.kinds[0]", violations[0].Path)
			assert.Contains(t, violations[0].Message,
				`kind Organization/Minister is still used by relationship "located_in" pair Organization/* -> region/country, as it is the last minor of Organization`)
		}
	})

	t.Run("wildcard pair removed in the same patch", func(t *testing.T) {
		result, violations := Apply(newBase(), &Patch{
			Remove: PatchRemovals{
				Kinds:         []string{"Organization/Minister"},
				Relationships: []string{"located_in"},
			},
		})
		if assert.Empty(t, violations) {
			assert.Nil(t, findKindDef(result, "Organization"))
		}
	})

	t.Run("exact reference is still rejected even with a replacement", func(t *testing.T) {
		base := newBase()
		base.RelationshipDefs = append(base.RelationshipDefs, RelationshipDef{
			Name: "reports_to", Endpoints: []Endpoint{{From: "Organization/Minister", To: "region/country"}},
		})

		result, violations := Apply(base, &Patch{
			Remove: PatchRemovals{Kinds: []string{"Organization/Minister"}},
			Add:    PatchAdditions{KindDefs: []KindDef{{Major: "Organization", Minors: []string{"Ministry"}}}},
		})
		assert.Nil(t, result)
		if assert.Len(t, violations, 1, "got %v", violations) {
			assert.Equal(t, "remove.kinds[0]", violations[0].Path)
			assert.Contains(t, violations[0].Message,
				`kind Organization/Minister is still used by relationship "reports_to" pair Organization/Minister -> region/country; remove that pair`)
		}
	})
}

func TestApplyNilInputs(t *testing.T) {
	result, violations := Apply(nil, &Patch{})
	assert.Nil(t, result)
	if assert.Len(t, violations, 1) {
		assert.Equal(t, "there is no schema to patch", violations[0].Message)
	}

	result, violations = Apply(parseFixture(t, "lk_regions_government.json"), nil)
	assert.Nil(t, result)
	if assert.Len(t, violations, 1) {
		assert.Equal(t, "patch is required", violations[0].Message)
	}
}

func TestApplyDoesNotModifyBase(t *testing.T) {
	patches := map[string]Patch{
		"successful patch": {
			Description: "changed",
			Add: PatchAdditions{
				KindDefs:         []KindDef{{Major: "region", Minors: []string{"new-minor"}}},
				RelationshipDefs: []RelationshipDef{{Name: "located_in", Endpoints: []Endpoint{{From: "Organization/Minister", To: "region/country"}}}},
			},
			Remove: PatchRemovals{
				Kinds:         []string{localGovernmentKind},
				Relationships: []string{"local_government"},
				Endpoints:     []EndpointRef{{Name: "located_in", From: "Organization/Government", To: "region/country"}},
			},
		},
		"rejected patch": {
			Remove: PatchRemovals{Endpoints: []EndpointRef{
				{Name: "has_department", From: "Organization/Minister", To: "Organization/Department"},
			}},
		},
	}

	for name, patch := range patches {
		t.Run(name, func(t *testing.T) {
			base := parseFixture(t, "lk_regions_government.json")
			Apply(base, &patch)
			assert.Equal(t, parseFixture(t, "lk_regions_government.json"), base)
		})
	}
}

func TestApplyResultDoesNotShareMemoryWithPatch(t *testing.T) {
	patch := Patch{Add: PatchAdditions{
		KindDefs: []KindDef{{Major: "Place", Minors: []string{"City"}}},
		RelationshipDefs: []RelationshipDef{
			{Name: "located_in_city", Endpoints: []Endpoint{{From: "Organization/Department", To: "Place/City"}}},
		},
	}}
	result, violations := Apply(parseFixture(t, "lk_regions_government.json"), &patch)
	if !assert.Empty(t, violations) {
		return
	}

	patch.Add.KindDefs[0].Minors[0] = "changed"
	patch.Add.RelationshipDefs[0].Endpoints[0].To = "changed"

	assert.Equal(t, []string{"City"}, findKindDef(result, "Place").Minors)
	assert.Equal(t, "Place/City", result.AllowedPairs("located_in_city")[0].To)
}

func TestPatchJSON(t *testing.T) {
	body := `{
		"baseVersion": 1,
		"description": "Drop local government",
		"add": {
			"kindDefs": [{ "major": "region", "minors": ["country-level-1+lk-western-province-special"] }],
			"relationshipDefs": [{ "name": "province", "endpoints": [{ "from": "region/country", "to": "region/country-level-1+lk-western-province-special" }] }]
		},
		"remove": {
			"kinds": ["region/country-level-3+lk-local-government"],
			"relationships": ["local_government"],
			"endpoints": [{ "name": "located_in", "from": "Organization/Government", "to": "region/country" }]
		}
	}`

	var patch Patch
	if !assert.NoError(t, json.Unmarshal([]byte(body), &patch)) {
		return
	}
	assert.Equal(t, 1, patch.BaseVersion)
	assert.Equal(t, "Drop local government", patch.Description)
	assert.Len(t, patch.Add.KindDefs, 1)
	assert.Len(t, patch.Add.RelationshipDefs, 1)
	assert.Equal(t, []string{localGovernmentKind}, patch.Remove.Kinds)
	assert.Equal(t, []string{"local_government"}, patch.Remove.Relationships)
	assert.Equal(t, EndpointRef{Name: "located_in", From: "Organization/Government", To: "region/country"}, patch.Remove.Endpoints[0])
}
