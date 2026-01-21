// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"

	"lk/datafoundation/core-api/pkg/typeinference"
)

func TestSchemaInfoToJSON(t *testing.T) {
	// Create a sample schema
	schema := &SchemaInfo{
		StorageType: ScalarData,
		TypeInfo: &typeinference.TypeInfo{
			Type:       typeinference.StringType,
			IsNullable: true,
		},
	}

	// Convert to JSON
	jsonSchema, err := SchemaInfoToJSON(schema)
	assert.NoError(t, err)
	assert.NotNil(t, jsonSchema)
	assert.Equal(t, ScalarData, jsonSchema.StorageType)
	assert.Equal(t, string(typeinference.StringType), jsonSchema.TypeInfo.Type)
	assert.True(t, jsonSchema.TypeInfo.IsNullable)
}

func TestTypeInfoToJSON(t *testing.T) {
	// Create a sample type info
	typeInfo := &typeinference.TypeInfo{
		Type:       typeinference.IntType,
		IsArray:    true,
		IsNullable: false,
		ArrayType: &typeinference.TypeInfo{
			Type: typeinference.StringType,
		},
	}

	// Convert to JSON
	jsonTypeInfo := TypeInfoToJSON(typeInfo)
	assert.NotNil(t, jsonTypeInfo)
	assert.Equal(t, string(typeinference.IntType), jsonTypeInfo.Type)
	assert.True(t, jsonTypeInfo.IsArray)
	assert.False(t, jsonTypeInfo.IsNullable)
	assert.Equal(t, string(typeinference.StringType), jsonTypeInfo.ArrayType.Type)
}

func TestJSONToAny(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		validate func(t *testing.T, any *anypb.Any, err error)
	}{
		{
			name:  "integer value",
			input: `42`,
			validate: func(t *testing.T, any *anypb.Any, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, any)
				// Unpack to struct
				var structValue structpb.Struct
				err = any.UnmarshalTo(&structValue)
				assert.NoError(t, err)
				// Check value field
				value, exists := structValue.Fields["value"]
				assert.True(t, exists)
				assert.Equal(t, float64(42), value.GetNumberValue())
			},
		},
		{
			name:  "string value",
			input: `"hello"`,
			validate: func(t *testing.T, any *anypb.Any, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, any)
				// Unpack to struct
				var structValue structpb.Struct
				err = any.UnmarshalTo(&structValue)
				assert.NoError(t, err)
				// Check value field
				value, exists := structValue.Fields["value"]
				assert.True(t, exists)
				assert.Equal(t, "hello", value.GetStringValue())
			},
		},
		{
			name:  "boolean value",
			input: `true`,
			validate: func(t *testing.T, any *anypb.Any, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, any)
				// Unpack to struct
				var structValue structpb.Struct
				err = any.UnmarshalTo(&structValue)
				assert.NoError(t, err)
				// Check value field
				value, exists := structValue.Fields["value"]
				assert.True(t, exists)
				assert.Equal(t, true, value.GetBoolValue())
			},
		},
		{
			name:  "null value",
			input: `null`,
			validate: func(t *testing.T, any *anypb.Any, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, any)
				// Unpack to struct
				var structValue structpb.Struct
				err = any.UnmarshalTo(&structValue)
				assert.NoError(t, err)
				// Check value field
				value, exists := structValue.Fields["value"]
				assert.True(t, exists)
				assert.Equal(t, structpb.NullValue_NULL_VALUE, value.GetNullValue())
			},
		},
		{
			name:  "array value",
			input: `[1, 2, 3]`,
			validate: func(t *testing.T, any *anypb.Any, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, any)
				// Unpack to struct
				var structValue structpb.Struct
				err = any.UnmarshalTo(&structValue)
				assert.NoError(t, err)
				// Check value field
				value, exists := structValue.Fields["value"]
				assert.True(t, exists)
				listValue := value.GetListValue()
				assert.NotNil(t, listValue)
				assert.Equal(t, 3, len(listValue.Values))
				assert.Equal(t, float64(1), listValue.Values[0].GetNumberValue())
				assert.Equal(t, float64(2), listValue.Values[1].GetNumberValue())
				assert.Equal(t, float64(3), listValue.Values[2].GetNumberValue())
			},
		},
		{
			name:  "object value",
			input: `{"name": "John", "age": 30}`,
			validate: func(t *testing.T, any *anypb.Any, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, any)
				// Unpack to struct
				var structValue structpb.Struct
				err = any.UnmarshalTo(&structValue)
				assert.NoError(t, err)
				// Check fields
				name, exists := structValue.Fields["name"]
				assert.True(t, exists)
				assert.Equal(t, "John", name.GetStringValue())
				age, exists := structValue.Fields["age"]
				assert.True(t, exists)
				assert.Equal(t, float64(30), age.GetNumberValue())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			any, err := JSONToAny(tc.input)
			tc.validate(t, any, err)
		})
	}
}

func TestAnyToJSON(t *testing.T) {
	// Create a sample Any value
	structValue, err := structpb.NewStruct(map[string]interface{}{
		"key": "value",
	})
	assert.NoError(t, err)

	any, err := anypb.New(structValue)
	assert.NoError(t, err)

	// Convert to JSON
	jsonStr, err := AnyToJSON(any)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonStr)

	// Parse the JSON string
	var result map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &result)
	assert.NoError(t, err)
	assert.Equal(t, "value", result["key"])
}

func TestValidateSchema(t *testing.T) {
	testCases := []struct {
		name     string
		value    interface{}
		schema   *SchemaInfo
		validate func(*testing.T, error)
	}{
		{
			name:  "valid scalar string",
			value: "hello",
			schema: &SchemaInfo{
				StorageType: ScalarData,
				TypeInfo: &typeinference.TypeInfo{
					Type: typeinference.StringType,
				},
			},
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:  "invalid scalar type",
			value: 42,
			schema: &SchemaInfo{
				StorageType: ScalarData,
				TypeInfo: &typeinference.TypeInfo{
					Type: typeinference.StringType,
				},
			},
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "expected string")
			},
		},
		{
			name:  "valid nullable scalar",
			value: nil,
			schema: &SchemaInfo{
				StorageType: ScalarData,
				TypeInfo: &typeinference.TypeInfo{
					Type:       typeinference.StringType,
					IsNullable: true,
				},
			},
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:  "invalid non-nullable scalar",
			value: nil,
			schema: &SchemaInfo{
				StorageType: ScalarData,
				TypeInfo: &typeinference.TypeInfo{
					Type:       typeinference.StringType,
					IsNullable: false,
				},
			},
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot be null")
			},
		},
		{
			name:  "valid list",
			value: []interface{}{"a", "b", "c"},
			schema: &SchemaInfo{
				StorageType: ListData,
				Items: &SchemaInfo{
					StorageType: ScalarData,
					TypeInfo: &typeinference.TypeInfo{
						Type: typeinference.StringType,
					},
				},
			},
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:  "invalid list item type",
			value: []interface{}{"a", 42, "c"},
			schema: &SchemaInfo{
				StorageType: ListData,
				Items: &SchemaInfo{
					StorageType: ScalarData,
					TypeInfo: &typeinference.TypeInfo{
						Type: typeinference.StringType,
					},
				},
			},
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid item at index 1")
			},
		},
		{
			name: "valid map",
			value: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			schema: &SchemaInfo{
				StorageType: MapData,
				Properties: map[string]*SchemaInfo{
					"name": {
						StorageType: ScalarData,
						TypeInfo: &typeinference.TypeInfo{
							Type: typeinference.StringType,
						},
					},
					"age": {
						StorageType: ScalarData,
						TypeInfo: &typeinference.TypeInfo{
							Type: typeinference.IntType,
						},
					},
				},
			},
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "missing required map property",
			value: map[string]interface{}{
				"name": "John",
			},
			schema: &SchemaInfo{
				StorageType: MapData,
				Properties: map[string]*SchemaInfo{
					"name": {
						StorageType: ScalarData,
						TypeInfo: &typeinference.TypeInfo{
							Type: typeinference.StringType,
						},
					},
					"age": {
						StorageType: ScalarData,
						TypeInfo: &typeinference.TypeInfo{
							Type: typeinference.IntType,
						},
					},
				},
			},
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "required key age is missing")
			},
		},
		{
			name: "valid graph",
			value: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{
						"id":   "node1",
						"type": "user",
					},
				},
				"edges": []interface{}{
					map[string]interface{}{
						"source": "node1",
						"target": "node2",
						"type":   "follows",
					},
				},
			},
			schema: &SchemaInfo{
				StorageType: GraphData,
				Fields: map[string]*SchemaInfo{
					"nodes": {
						StorageType: ListData,
						Items: &SchemaInfo{
							StorageType: MapData,
							Properties: map[string]*SchemaInfo{
								"id": {
									StorageType: ScalarData,
									TypeInfo: &typeinference.TypeInfo{
										Type: typeinference.StringType,
									},
								},
								"type": {
									StorageType: ScalarData,
									TypeInfo: &typeinference.TypeInfo{
										Type: typeinference.StringType,
									},
								},
							},
						},
					},
					"edges": {
						StorageType: ListData,
						Items: &SchemaInfo{
							StorageType: MapData,
							Properties: map[string]*SchemaInfo{
								"source": {
									StorageType: ScalarData,
									TypeInfo: &typeinference.TypeInfo{
										Type: typeinference.StringType,
									},
								},
								"target": {
									StorageType: ScalarData,
									TypeInfo: &typeinference.TypeInfo{
										Type: typeinference.StringType,
									},
								},
								"type": {
									StorageType: ScalarData,
									TypeInfo: &typeinference.TypeInfo{
										Type: typeinference.StringType,
									},
								},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSchema(tc.value, tc.schema)
			tc.validate(t, err)
		})
	}
}

func TestValidateTabularData(t *testing.T) {
	// Create a schema for tabular data
	schema := &SchemaInfo{
		StorageType: TabularData,
		TypeInfo:    &typeinference.TypeInfo{Type: typeinference.StringType},
		Fields: map[string]*SchemaInfo{
			"id": {
				StorageType: ScalarData,
				TypeInfo:    &typeinference.TypeInfo{Type: typeinference.IntType},
			},
			"name": {
				StorageType: ScalarData,
				TypeInfo:    &typeinference.TypeInfo{Type: typeinference.StringType},
			},
			"age": {
				StorageType: ScalarData,
				TypeInfo:    &typeinference.TypeInfo{Type: typeinference.IntType},
			},
		},
	}

	testCases := []struct {
		name        string
		input       interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid tabular data",
			input: map[string]interface{}{
				"table": map[string]interface{}{
					"columns": []interface{}{"id", "name", "age"},
					"rows": []interface{}{
						[]interface{}{1, "Alice", 30},
						[]interface{}{2, "Bob", 25},
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing table field",
			input: map[string]interface{}{
				"data": map[string]interface{}{
					"columns": []interface{}{"id", "name"},
					"rows":    []interface{}{[]interface{}{1, "Alice"}},
				},
			},
			expectError: true,
			errorMsg:    "tabular data must contain a 'table' field",
		},
		{
			name: "table field not a map",
			input: map[string]interface{}{
				"table": "not a map",
			},
			expectError: true,
			errorMsg:    "table field must be a map",
		},
		{
			name: "missing columns field",
			input: map[string]interface{}{
				"table": map[string]interface{}{
					"rows": []interface{}{[]interface{}{1, "Alice"}},
				},
			},
			expectError: true,
			errorMsg:    "table must contain 'columns' field",
		},
		{
			name: "missing rows field",
			input: map[string]interface{}{
				"table": map[string]interface{}{
					"columns": []interface{}{"id", "name"},
				},
			},
			expectError: true,
			errorMsg:    "table must contain 'rows' field",
		},
		{
			name: "columns not a list",
			input: map[string]interface{}{
				"table": map[string]interface{}{
					"columns": "not a list",
					"rows":    []interface{}{[]interface{}{1, "Alice"}},
				},
			},
			expectError: true,
			errorMsg:    "columns must be a list",
		},
		{
			name: "rows not a list",
			input: map[string]interface{}{
				"table": map[string]interface{}{
					"columns": []interface{}{"id", "name"},
					"rows":    "not a list",
				},
			},
			expectError: true,
			errorMsg:    "rows must be a list",
		},
		{
			name: "row not a list",
			input: map[string]interface{}{
				"table": map[string]interface{}{
					"columns": []interface{}{"id", "name"},
					"rows": []interface{}{
						"not a list",
					},
				},
			},
			expectError: true,
			errorMsg:    "row 0 must be a list",
		},
		{
			name: "row length mismatch",
			input: map[string]interface{}{
				"table": map[string]interface{}{
					"columns": []interface{}{"id", "name", "age"},
					"rows": []interface{}{
						[]interface{}{1, "Alice"}, // Missing age
						[]interface{}{2, "Bob", 25},
					},
				},
			},
			expectError: true,
			errorMsg:    "row 0 has 2 columns, expected 3",
		},
		{
			name: "invalid column type",
			input: map[string]interface{}{
				"table": map[string]interface{}{
					"columns": []interface{}{"id", "name", "age"},
					"rows": []interface{}{
						[]interface{}{"not an int", "Alice", 30}, // id should be int
						[]interface{}{2, "Bob", 25},
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid value in row 0, column id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSchema(tc.input, schema)
			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
