// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package typeinference

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

// JSONToAny converts a JSON string to a protobuf Any value
func JSONToAny(jsonStr string) (*anypb.Any, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Handle scalar values
	switch v := data.(type) {
	case float64:
		// Check if it's an integer
		if v == float64(int64(v)) {
			structValue, err := structpb.NewStruct(map[string]interface{}{
				"value": int64(v),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create struct: %v", err)
			}
			return anypb.New(structValue)
		}
		structValue, err := structpb.NewStruct(map[string]interface{}{
			"value": v,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create struct: %v", err)
		}
		return anypb.New(structValue)
	case string:
		structValue, err := structpb.NewStruct(map[string]interface{}{
			"value": v,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create struct: %v", err)
		}
		return anypb.New(structValue)
	case bool:
		structValue, err := structpb.NewStruct(map[string]interface{}{
			"value": v,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create struct: %v", err)
		}
		return anypb.New(structValue)
	case nil:
		structValue, err := structpb.NewStruct(map[string]interface{}{
			"value": nil,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create struct: %v", err)
		}
		return anypb.New(structValue)
	case []interface{}:
		structValue, err := structpb.NewStruct(map[string]interface{}{
			"value": v,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create struct: %v", err)
		}
		return anypb.New(structValue)
	}

	// Handle objects
	if obj, ok := data.(map[string]interface{}); ok {
		structValue, err := structpb.NewStruct(obj)
		if err != nil {
			return nil, fmt.Errorf("failed to create struct: %v", err)
		}
		return anypb.New(structValue)
	}

	return nil, fmt.Errorf("unsupported data type: %T", data)
}

// TestScalarTypes tests type inference for scalar data types
func TestScalarTypes(t *testing.T) {
	// FIXME: https://github.com/LDFLK/nexoan/issues/142
	testCases := map[string]struct {
		json     string
		expected DataType
	}{
		"empty_string": {
			json:     `""`,
			expected: StringType,
		},
		"zero_int": {
			json:     `0`,
			expected: IntType,
		},
		"zero_float": {
			json:     `0.1`,
			expected: FloatType,
		},
		"integer": {
			json:     `42`,
			expected: IntType,
		},
		"float": {
			json:     `3.14`,
			expected: FloatType,
		},
		"string": {
			json:     `"hello"`,
			expected: StringType,
		},
		"boolean": {
			json:     `true`,
			expected: BoolType,
		},
		"null": {
			json:     `null`,
			expected: NullType,
		},
		"empty_object": {
			json:     `{}`,
			expected: StringType,
		},
	}

	inferrer := &TypeInferrer{}
	for testName, tc := range testCases {
		t.Run(testName, func(t *testing.T) {
			anyValue, err := JSONToAny(tc.json)
			assert.NoError(t, err)

			typeInfo, err := inferrer.InferType(anyValue)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, typeInfo.Type)
		})
	}
}

// TestListTypes tests type inference for list data types
func TestListTypes(t *testing.T) {
	testCases := map[string]struct {
		json     string
		expected *TypeInfo
	}{
		"empty_list": {
			json: `[]`,
			expected: &TypeInfo{
				Type:    StringType,
				IsArray: true,
			},
		},
		"list_with_empty_values": {
			json: `["", 0, 0.0, null]`,
			expected: &TypeInfo{
				Type:    StringType,
				IsArray: true,
				ArrayType: &TypeInfo{
					Type: StringType,
				},
			},
		},
		"list_of_integers": {
			json: `[1, 2, 3]`,
			expected: &TypeInfo{
				Type:    StringType,
				IsArray: true,
				ArrayType: &TypeInfo{
					Type: IntType,
				},
			},
		},
		"list_of_strings": {
			json: `["a", "b", "c"]`,
			expected: &TypeInfo{
				Type:    StringType,
				IsArray: true,
				ArrayType: &TypeInfo{
					Type: StringType,
				},
			},
		},
		"list_with_null": {
			json: `null`,
			expected: &TypeInfo{
				Type:       NullType,
				IsNullable: true,
			},
		},
	}

	inferrer := &TypeInferrer{}
	for testName, tc := range testCases {
		t.Run(testName, func(t *testing.T) {
			anyValue, err := JSONToAny(tc.json)
			assert.NoError(t, err)

			typeInfo, err := inferrer.InferType(anyValue)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected.Type, typeInfo.Type)
			assert.Equal(t, tc.expected.IsArray, typeInfo.IsArray)
			if tc.expected.ArrayType != nil {
				assert.NotNil(t, typeInfo.ArrayType)
				assert.Equal(t, tc.expected.ArrayType.Type, typeInfo.ArrayType.Type)
			}
		})
	}
}

// TestMapTypes tests type inference for map data types
func TestMapTypes(t *testing.T) {
	testCases := map[string]struct {
		json     string
		expected *TypeInfo
	}{
		"empty_map": {
			json: `{}`,
			expected: &TypeInfo{
				Type: StringType,
			},
		},
		"map_with_empty_values": {
			json: `{"empty_str": "", "zero": 0, "null_val": null}`,
			expected: &TypeInfo{
				Type: StringType,
			},
		},
		"simple_map": {
			json: `{"key": "value"}`,
			expected: &TypeInfo{
				Type: StringType,
			},
		},
		"map_with_null": {
			json: `null`,
			expected: &TypeInfo{
				Type:       NullType,
				IsNullable: true,
			},
		},
	}

	inferrer := &TypeInferrer{}
	for testName, tc := range testCases {
		t.Run(testName, func(t *testing.T) {
			anyValue, err := JSONToAny(tc.json)
			assert.NoError(t, err)

			typeInfo, err := inferrer.InferType(anyValue)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected.Type, typeInfo.Type)
		})
	}
}

// TestSpecialTypes tests type inference for special data types
func TestSpecialTypes(t *testing.T) {
	testCases := map[string]struct {
		json     string
		expected DataType
	}{
		"date": {
			json:     `{"attributes": "2024-03-20"}`,
			expected: DateType,
		},
		"time": {
			json:     `{"attributes": "14:30:00"}`,
			expected: TimeType,
		},
		"datetime": {
			json:     `{"attributes": "2024-03-20T14:30:00Z"}`,
			expected: DateTimeType,
		},
	}

	inferrer := &TypeInferrer{}
	for testName, tc := range testCases {
		t.Run(testName, func(t *testing.T) {
			anyValue, err := JSONToAny(tc.json)
			assert.NoError(t, err)

			typeInfo, err := inferrer.InferType(anyValue)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, typeInfo.Type)
		})
	}
}
