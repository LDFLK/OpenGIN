// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	"lk/datafoundation/core-api/pkg/typeinference"
)

// TestSchemaGeneration tests the schema generation for different data structures
func TestSchemaGeneration(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected string
	}{
		"scalar_data": {
			input: `{
				"value": 42
			}`,
			expected: `{
				"storage_type": "scalar",
				"type_info": {
					"type": "int"
				}
			}`,
		},
		"list_data": {
			input: `{
				"numbers": [1, 2, 3]
			}`,
			expected: `{
				"storage_type": "list",
				"type_info": {
					"type": "string",
					"is_array": true,
					"array_type": {
						"type": "int"
					}
				},
				"items": {
					"storage_type": "scalar",
					"type_info": {
						"type": "int"
					}
				}
			}`,
		},
		"list_data_with_different_name": {
			input: `{
				"values": [1, 2, 3]
			}`,
			expected: `{
				"storage_type": "list",
				"type_info": {
					"type": "string",
					"is_array": true,
					"array_type": {
						"type": "int"
					}
				},
				"items": {
					"storage_type": "scalar",
					"type_info": {
						"type": "int"
					}
				}
			}`,
		},
		"map_data": {
			input: `{
				"properties": {
					"name": "John",
					"age": 30,
					"active": true
				}
			}`,
			expected: `{
				"storage_type": "map",
				"type_info": {
					"type": "string"
				},
				"properties": {
					"name": {
						"storage_type": "scalar",
						"type_info": {
							"type": "string"
						}
					},
					"age": {
						"storage_type": "scalar",
						"type_info": {
							"type": "int"
						}
					},
					"active": {
						"storage_type": "scalar",
						"type_info": {
							"type": "bool"
						}
					}
				}
			}`,
		},
		"empty_values": {
			input: `{
				"properties": {
					"empty_str": "",
					"zero": 0,
					"null_val": null
				}
			}`,
			expected: `{
				"storage_type": "map",
				"type_info": {
					"type": "string"
				},
				"properties": {
					"empty_str": {
						"storage_type": "scalar",
						"type_info": {
							"type": "string"
						}
					},
					"zero": {
						"storage_type": "scalar",
						"type_info": {
							"type": "int"
						}
					},
					"null_val": {
						"storage_type": "scalar",
						"type_info": {
							"type": "null",
							"is_nullable": true
						}
					}
				}
			}`,
		},
		"graph_data_with_nodes": {
			input: `{
				"nodes": {
					"person": {
						"name": "John",
						"age": 30,
						"active": true
					}
				}
			}`,
			expected: `{
				"storage_type": "graph",
				"type_info": {
					"type": "string"
				},
				"fields": {
					"nodes": {
						"storage_type": "map",
						"type_info": {
							"type": "string"
						},
						"properties": {
							"person": {
								"storage_type": "map",
								"type_info": {
									"type": "string"
								},
								"properties": {
									"name": {
										"storage_type": "scalar",
										"type_info": {
											"type": "string"
										}
									},
									"age": {
										"storage_type": "scalar",
										"type_info": {
											"type": "int"
										}
									},
									"active": {
										"storage_type": "scalar",
										"type_info": {
											"type": "bool"
										}
									}
								}
							}
						}
					}
				}
			}`,
		},
		"graph_data_with_edges": {
			input: `{
				"edges": {
					"knows": {
						"since": "2020-01-01",
						"strength": 0.8
					}
				}
			}`,
			expected: `{
				"storage_type": "graph",
				"type_info": {
					"type": "string"
				},
				"fields": {
					"edges": {
						"storage_type": "map",
						"type_info": {
							"type": "string"
						},
						"properties": {
							"knows": {
								"storage_type": "map",
								"type_info": {
									"type": "string"
								},
								"properties": {
									"since": {
										"storage_type": "scalar",
										"type_info": {
											"type": "date"
										}
									},
									"strength": {
										"storage_type": "scalar",
										"type_info": {
											"type": "float"
										}
									}
								}
							}
						}
					}
				}
			}`,
		},
		"graph_data_with_both": {
			input: `{
				"nodes": {
					"person": {
						"name": "John",
						"age": 30
					}
				},
				"edges": {
					"knows": {
						"since": "2020-01-01"
					}
				}
			}`,
			expected: `{
				"storage_type": "graph",
				"type_info": {
					"type": "string"
				},
				"fields": {
					"nodes": {
						"storage_type": "map",
						"type_info": {
							"type": "string"
						},
						"properties": {
							"person": {
								"storage_type": "map",
								"type_info": {
									"type": "string"
								},
								"properties": {
									"name": {
										"storage_type": "scalar",
										"type_info": {
											"type": "string"
										}
									},
									"age": {
										"storage_type": "scalar",
										"type_info": {
											"type": "int"
										}
									}
								}
							}
						}
					},
					"edges": {
						"storage_type": "map",
						"type_info": {
							"type": "string"
						},
						"properties": {
							"knows": {
								"storage_type": "map",
								"type_info": {
									"type": "string"
								},
								"properties": {
									"since": {
										"storage_type": "scalar",
										"type_info": {
											"type": "date"
										}
									}
								}
							}
						}
					}
				}
			}`,
		},
		"social_network": {
			input: `{
				"nodes": [
					{"id": "user1", "type": "user", "properties": {"name": "Alice", "age": 30, "location": "NY"}},
					{"id": "user2", "type": "user", "properties": {"name": "Bob", "age": 25, "location": "SF"}},
					{"id": "user3", "type": "user", "properties": {"name": "Charlie", "age": 35, "location": "LA"}},
					{"id": "post1", "type": "post", "properties": {"title": "Hello", "content": "World", "created": "2024-03-20"}},
					{"id": "post2", "type": "post", "properties": {"title": "Graph", "content": "DB", "created": "2024-03-21"}}
				],
				"edges": [
					{"source": "user1", "target": "user2", "type": "follows", "properties": {"since": "2024-01-01"}},
					{"source": "user2", "target": "user3", "type": "follows", "properties": {"since": "2024-02-01"}},
					{"source": "user1", "target": "post1", "type": "created", "properties": {"timestamp": "2024-03-20T10:00:00Z"}},
					{"source": "user2", "target": "post1", "type": "likes", "properties": {"timestamp": "2024-03-20T11:00:00Z"}},
					{"source": "user3", "target": "post2", "type": "created", "properties": {"timestamp": "2024-03-21T09:00:00Z"}}
				]
			}`,
			expected: `{
				"storage_type": "graph",
				"type_info": {
					"type": "string"
				},
				"fields": {
					"nodes": {
						"storage_type": "map",
						"type_info": {
							"type": "string"
						},
						"properties": {
							"user": {
								"storage_type": "map",
								"type_info": {
									"type": "string"
								},
								"properties": {
									"name": {
										"storage_type": "scalar",
										"type_info": {
											"type": "string"
										}
									},
									"age": {
										"storage_type": "scalar",
										"type_info": {
											"type": "int"
										}
									},
									"location": {
										"storage_type": "scalar",
										"type_info": {
											"type": "string"
										}
									}
								}
							},
							"post": {
								"storage_type": "map",
								"type_info": {
									"type": "string"
								},
								"properties": {
									"title": {
										"storage_type": "scalar",
										"type_info": {
											"type": "string"
										}
									},
									"content": {
										"storage_type": "scalar",
										"type_info": {
											"type": "string"
										}
									},
									"created": {
										"storage_type": "scalar",
										"type_info": {
											"type": "date"
										}
									}
								}
							}
						}
					},
					"edges": {
						"storage_type": "map",
						"type_info": {
							"type": "string"
						},
						"properties": {
							"follows": {
								"storage_type": "map",
								"type_info": {
									"type": "string"
								},
								"properties": {
									"since": {
										"storage_type": "scalar",
										"type_info": {
											"type": "date"
										}
									}
								}
							},
							"created": {
								"storage_type": "map",
								"type_info": {
									"type": "string"
								},
								"properties": {
									"timestamp": {
										"storage_type": "scalar",
										"type_info": {
											"type": "datetime"
										}
									}
								}
							},
							"likes": {
								"storage_type": "map",
								"type_info": {
									"type": "string"
								},
								"properties": {
									"timestamp": {
										"storage_type": "scalar",
										"type_info": {
											"type": "datetime"
										}
									}
								}
							}
						}
					}
				}
			}`,
		},
		"dependency_graph": {
			input: `{
				"nodes": [
					{"id": "pkg1", "type": "package", "properties": {"name": "core", "version": "1.0.0"}},
					{"id": "pkg2", "type": "package", "properties": {"name": "utils", "version": "2.1.0"}},
					{"id": "pkg3", "type": "package", "properties": {"name": "network", "version": "1.5.0"}},
					{"id": "pkg4", "type": "package", "properties": {"name": "database", "version": "3.0.0"}}
				],
				"edges": [
					{"source": "pkg2", "target": "pkg1", "type": "depends_on", "properties": {"version": ">=1.0.0"}},
					{"source": "pkg3", "target": "pkg1", "type": "depends_on", "properties": {"version": ">=1.0.0"}},
					{"source": "pkg4", "target": "pkg2", "type": "depends_on", "properties": {"version": ">=2.0.0"}},
					{"source": "pkg4", "target": "pkg3", "type": "depends_on", "properties": {"version": ">=1.5.0"}}
				]
			}`,
			expected: `{
				"storage_type": "graph",
				"type_info": {
					"type": "string"
				},
				"fields": {
					"nodes": {
						"storage_type": "map",
						"type_info": {
							"type": "string"
						},
						"properties": {
							"package": {
								"storage_type": "map",
								"type_info": {
									"type": "string"
								},
								"properties": {
									"name": {
										"storage_type": "scalar",
										"type_info": {
											"type": "string"
										}
									},
									"version": {
										"storage_type": "scalar",
										"type_info": {
											"type": "string"
										}
									}
								}
							}
						}
					},
					"edges": {
						"storage_type": "map",
						"type_info": {
							"type": "string"
						},
						"properties": {
							"depends_on": {
								"storage_type": "map",
								"type_info": {
									"type": "string"
								},
								"properties": {
									"version": {
										"storage_type": "scalar",
										"type_info": {
											"type": "string"
										}
									}
								}
							}
						}
					}
				}
			}`,
		},
		"simple_tabular_data": {
			input: `{
				"columns": ["id", "name", "age"],
				"rows": [
					[1, "Alice", 30],
					[2, "Bob", 25]
				]
			}`,
			expected: `{
				"storage_type": "tabular",
				"type_info": {
					"type": "string"
				},
				"fields": {
					"id": {
						"storage_type": "scalar",
						"type_info": {
							"type": "numeric",
							"is_nullable": true
						}
					},
					"name": {
						"storage_type": "scalar",
						"type_info": {
							"type": "string",
							"is_nullable": true
						}
					},
					"age": {
						"storage_type": "scalar",
						"type_info": {
							"type": "numeric",
							"is_nullable": true
						}
					}
				}
			}`,
		},
	}

	generator := NewSchemaGenerator()
	for testName, tc := range testCases {
		t.Run(testName, func(t *testing.T) {
			// Parse input
			anyValue, err := JSONToAny(tc.input)
			fmt.Println(">>>>>>> Go Converted JSON to Any")
			fmt.Println(anyValue)
			fmt.Println(">>>>>>>")

			assert.NoError(t, err)

			// Generate schema
			schema, err := generator.GenerateSchema(anyValue)
			fmt.Println(">>>>>>> Go Generated Schema")
			fmt.Println(schema)
			fmt.Println(">>>>>>>")
			assert.NoError(t, err)

			// Convert schema to JSON
			schemaJSON, err := SchemaInfoToJSON(schema)
			assert.NoError(t, err)

			// Parse expected JSON
			var expectedJSON SchemaInfoJSON
			err = json.Unmarshal([]byte(tc.expected), &expectedJSON)
			assert.NoError(t, err)

			// Compare schemas
			assert.Equal(t, expectedJSON, *schemaJSON)
		})
	}
}

// makeTabularStruct builds a *structpb.Struct in {columns:[...], rows:[[...],...]} format.
func makeTabularStruct(t *testing.T, columns []string, rows [][]interface{}) *structpb.Struct {
	t.Helper()
	colVals := make([]*structpb.Value, len(columns))
	for i, c := range columns {
		colVals[i] = structpb.NewStringValue(c)
	}
	rowVals := make([]*structpb.Value, len(rows))
	for i, row := range rows {
		cells := make([]*structpb.Value, len(row))
		for j, cell := range row {
			switch v := cell.(type) {
			case nil:
				cells[j] = structpb.NewNullValue()
			case string:
				cells[j] = structpb.NewStringValue(v)
			case float64:
				cells[j] = structpb.NewNumberValue(v)
			case int:
				cells[j] = structpb.NewNumberValue(float64(v))
			case bool:
				cells[j] = structpb.NewBoolValue(v)
			}
		}
		rowVals[i] = structpb.NewListValue(&structpb.ListValue{Values: cells})
	}
	return &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"columns": structpb.NewListValue(&structpb.ListValue{Values: colVals}),
			"rows":    structpb.NewListValue(&structpb.ListValue{Values: rowVals}),
		},
	}
}

// TestInferColumnTypes tests the inferColumnTypes function directly.
func TestInferColumnTypes(t *testing.T) {
	tests := []struct {
		name          string
		columns       []string
		rows          [][]interface{}
		expectError   bool
		expectedTypes map[string]typeinference.DataType
	}{
		{
			name:    "numbers infer as NumericType",
			columns: []string{"val"},
			rows:    [][]interface{}{{42}, {3.14}},
			expectedTypes: map[string]typeinference.DataType{
				"val": typeinference.NumericType,
			},
		},
		{
			name:    "null in row 1 resolved by row 2",
			columns: []string{"val"},
			rows:    [][]interface{}{{nil}, {99}},
			expectedTypes: map[string]typeinference.DataType{
				"val": typeinference.NumericType,
			},
		},
		{
			name:    "string column",
			columns: []string{"name"},
			rows:    [][]interface{}{{"Alice"}, {"Bob"}},
			expectedTypes: map[string]typeinference.DataType{
				"name": typeinference.StringType,
			},
		},
		{
			name:    "bool column",
			columns: []string{"active"},
			rows:    [][]interface{}{{true}, {false}},
			expectedTypes: map[string]typeinference.DataType{
				"active": typeinference.BoolType,
			},
		},
		{
			name:    "datetime string column",
			columns: []string{"ts"},
			rows:    [][]interface{}{{"2024-03-20T10:00:00Z"}, {"2024-03-21T10:00:00Z"}},
			expectedTypes: map[string]typeinference.DataType{
				"ts": typeinference.DateTimeType,
			},
		},
		{
			name:    "date string column",
			columns: []string{"d"},
			rows:    [][]interface{}{{"2024-03-20"}, {"2024-03-21"}},
			expectedTypes: map[string]typeinference.DataType{
				"d": typeinference.DateType,
			},
		},
		{
			name:    "all nulls returns error",
			columns: []string{"val"},
			rows:    [][]interface{}{{nil}, {nil}},
			expectError: true,
		},
		{
			name:    "type set from first non-null ignores later values",
			columns: []string{"val"},
			rows:    [][]interface{}{{nil}, {42}, {"not a number"}},
			expectedTypes: map[string]typeinference.DataType{
				"val": typeinference.NumericType, // set by row 1 (42), row 2 irrelevant to inference
			},
		},
		{
			name:    "multiple columns inferred independently",
			columns: []string{"num", "str", "flag"},
			rows: [][]interface{}{
				{nil, "hello", nil},
				{7, nil, true},
			},
			expectedTypes: map[string]typeinference.DataType{
				"num":  typeinference.NumericType,
				"str":  typeinference.StringType,
				"flag": typeinference.BoolType,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := makeTabularStruct(t, tt.columns, tt.rows)
			colTypes, err := inferColumnTypes(data)

			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			for col, expectedType := range tt.expectedTypes {
				actual, ok := colTypes[col]
				assert.True(t, ok, "column %q missing from result", col)
				assert.Equal(t, expectedType, actual.Type, "type mismatch for column %q", col)
				assert.True(t, actual.IsNullable, "column %q should be nullable", col)
			}
		})
	}
}
