// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package typeinference

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

// JSONToAny converts a JSON string to a protobuf Any value.
func JSONToAny(jsonStr string) (*anypb.Any, error) {
	var data interface{}
	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	decoder.UseNumber()
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	switch v := data.(type) {
	case map[string]interface{}:
		structValue, err := buildStruct(v)
		if err != nil {
			return nil, err
		}
		return anypb.New(structValue)
	case []interface{}:
		value, err := convertJSONValue(v)
		if err != nil {
			return nil, err
		}
		return anypb.New(&structpb.Struct{Fields: map[string]*structpb.Value{"value": value}})
	default:
		value, err := convertJSONValue(v)
		if err != nil {
			return nil, err
		}
		return anypb.New(&structpb.Struct{Fields: map[string]*structpb.Value{"value": value}})
	}
}

func buildStruct(data map[string]interface{}) (*structpb.Struct, error) {
	fields := make(map[string]*structpb.Value, len(data))
	for key, value := range data {
		converted, err := convertJSONValue(value)
		if err != nil {
			return nil, err
		}
		fields[key] = converted
	}
	return &structpb.Struct{Fields: fields}, nil
}

func convertJSONValue(value interface{}) (*structpb.Value, error) {
	switch v := value.(type) {
	case nil:
		return &structpb.Value{Kind: &structpb.Value_NullValue{}}, nil
	case string:
		return &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: v}}, nil
	case bool:
		return &structpb.Value{Kind: &structpb.Value_BoolValue{BoolValue: v}}, nil
	case json.Number:
		return &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: v.String()}}, nil
	case float64:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: v}}, nil
	case float32:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(v)}}, nil
	case int:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(v)}}, nil
	case int8:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(v)}}, nil
	case int16:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(v)}}, nil
	case int32:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(v)}}, nil
	case int64:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(v)}}, nil
	case uint:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(v)}}, nil
	case uint8:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(v)}}, nil
	case uint16:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(v)}}, nil
	case uint32:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(v)}}, nil
	case uint64:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(v)}}, nil
	case []interface{}:
		listValues := make([]*structpb.Value, len(v))
		for i, item := range v {
			converted, err := convertJSONValue(item)
			if err != nil {
				return nil, err
			}
			listValues[i] = converted
		}
		return &structpb.Value{Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: listValues}}}, nil
	case map[string]interface{}:
		fields := make(map[string]*structpb.Value, len(v))
		for key, item := range v {
			converted, err := convertJSONValue(item)
			if err != nil {
				return nil, err
			}
			fields[key] = converted
		}
		return &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: fields}}}, nil
	default:
		return nil, fmt.Errorf("unsupported data type: %T", value)
	}
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
		"zero_float_literal": {
			json:     `0.0`,
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
