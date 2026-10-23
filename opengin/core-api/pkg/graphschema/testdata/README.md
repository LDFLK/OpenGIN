# Graph schema test fixtures

The graph schema is submitted as JSON, so every fixture is JSON. JSON has no comments,
so the purpose of each file is recorded here.

## Valid

| File | Purpose |
|---|---|
| `lk_regions_government.json` | Sri Lanka region hierarchy plus government structure: 13 kinds, 12 relationship names, 13 endpoint pairs (`located_in` has two). |
| `valid_wildcards.json` | `"*"` as a minor kind, and `Major/*` kind references in endpoints. |

## Rejected

| File | Expected result |
|---|---|
| `tree_with_loader_keys.json` | Parse error: a hierarchical loader-style document (`major`, `minor`, `file`, `geojson`, `children`, ...) is not the graph schema format, so its keys are unknown. |
| `invalid/cardinality_field.json` | Parse error, unknown key "cardinality". Cardinality is not part of the v1 schema. |
| `invalid/category_field.json` | Parse error, unknown key "category". Category is not part of the v1 schema. |
| `invalid/duplicate_endpoint_pair.json` | Violation at `relationshipDefs[0].endpoints[1]`: duplicate pair region/country -> region/province. |
| `invalid/duplicate_major.json` | Violation at `kindDefs[1].major`: duplicate Major "region"; add minors to kindDefs[0]. |
| `invalid/duplicate_minor.json` | Violation at `kindDefs[0].minors[2]`: duplicate minor kind "country". |
| `invalid/duplicate_relationship_name.json` | Violation at `relationshipDefs[1].name`: duplicate "located_in"; add the pair to relationshipDefs[0].endpoints instead. |
| `invalid/empty_endpoints.json` | Violation at `relationshipDefs[0].endpoints`: required, non-empty. |
| `invalid/empty_minor.json` | Violation at `kindDefs[0].minors[1]`: empty minor kind. |
| `invalid/invalid_major_name.json` | Violation at `kindDefs[0].major`: "lk-region" is not a safe Neo4j label. |
| `invalid/invalid_relationship_name.json` | Violation at `relationshipDefs[0].name`: "lk-province" is not a safe Neo4j relationship type. |
| `invalid/malformed_kind_reference.json` | Violation at `relationshipDefs[0].endpoints[0].to`: "region" is not in Major/Minor form. |
| `invalid/minor_with_slash.json` | Violation at `kindDefs[0].minors[1]`: minor kind contains "/". |
| `invalid/missing_major.json` | Violation at `kindDefs[1].major`: required. |
| `invalid/missing_minors.json` | Violation at `kindDefs[1].minors`: required, non-empty. |
| `invalid/missing_relationship_name.json` | Violation at `relationshipDefs[0].name`: required. |
| `invalid/reserved_major.json` | Violation at `kindDefs[1].major`: "Dataset" is reserved for system attribute nodes. |
| `invalid/reserved_relationship_name.json` | Violation at `relationshipDefs[1].name`: "IS_ATTRIBUTE" is reserved for system attribute edges. |
| `invalid/unknown_kind_reference.json` | Violation at `relationshipDefs[0].endpoints[0].to`: unknown kind region/district. |
| `invalid/unknown_top_level_key.json` | Parse error, unknown key "namespace". |
| `invalid/wildcard_unknown_major.json` | Violation at `relationshipDefs[0].endpoints[0].from`: Organization/* used but Major "Organization" is not declared. |
