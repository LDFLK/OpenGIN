// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"encoding/json"
	"fmt"
	"log"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"

	"lk/datafoundation/core-api/pkg/typeinference"
)

// Storage type constants
const (
	ScalarData  = "scalar"
	ListData    = "list"
	MapData     = "map"
	GraphData   = "graph"
	TabularData = "tabular"
)

// SchemaInfoJSON represents the JSON structure of SchemaInfo
type SchemaInfoJSON struct {
	StorageType string                     `json:"storage_type"`
	TypeInfo    *TypeInfoJSON              `json:"type_info"`
	Fields      map[string]*SchemaInfoJSON `json:"fields,omitempty"`
	Items       *SchemaInfoJSON            `json:"items,omitempty"`
	Properties  map[string]*SchemaInfoJSON `json:"properties,omitempty"`
}

// TypeInfoJSON represents the JSON structure of TypeInfo
type TypeInfoJSON struct {
	Type       string        `json:"type"`
	IsNullable bool          `json:"is_nullable,omitempty"`
	IsArray    bool          `json:"is_array,omitempty"`
	ArrayType  *TypeInfoJSON `json:"array_type,omitempty"`
}

// SchemaInfoToJSON converts a SchemaInfo to its JSON representation
func SchemaInfoToJSON(schema *SchemaInfo) (*SchemaInfoJSON, error) {
	if schema == nil {
		return nil, nil
	}

	jsonSchema := &SchemaInfoJSON{
		StorageType: string(schema.StorageType),
		TypeInfo:    TypeInfoToJSON(schema.TypeInfo),
	}

	if schema.Fields != nil {
		jsonSchema.Fields = make(map[string]*SchemaInfoJSON)
		for k, v := range schema.Fields {
			fieldJSON, err := SchemaInfoToJSON(v)
			if err != nil {
				return nil, err
			}
			jsonSchema.Fields[k] = fieldJSON
		}
	}

	if schema.Items != nil {
		itemsJSON, err := SchemaInfoToJSON(schema.Items)
		if err != nil {
			return nil, err
		}
		jsonSchema.Items = itemsJSON
	}

	if schema.Properties != nil {
		jsonSchema.Properties = make(map[string]*SchemaInfoJSON)
		for k, v := range schema.Properties {
			propJSON, err := SchemaInfoToJSON(v)
			if err != nil {
				return nil, err
			}
			jsonSchema.Properties[k] = propJSON
		}
	}

	return jsonSchema, nil
}

// TypeInfoToJSON converts a TypeInfo to its JSON representation
func TypeInfoToJSON(typeInfo *typeinference.TypeInfo) *TypeInfoJSON {
	if typeInfo == nil {
		return nil
	}

	jsonTypeInfo := &TypeInfoJSON{
		Type:       string(typeInfo.Type),
		IsNullable: typeInfo.IsNullable,
		IsArray:    typeInfo.IsArray,
	}

	if typeInfo.ArrayType != nil {
		jsonTypeInfo.ArrayType = TypeInfoToJSON(typeInfo.ArrayType)
	}

	return jsonTypeInfo
}

// JSONToAny converts a JSON string to a protobuf Any value
func JSONToAny(jsonStr string) (*anypb.Any, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
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
				return nil, err
			}
			return anypb.New(structValue)
		}
		structValue, err := structpb.NewStruct(map[string]interface{}{
			"value": v,
		})
		if err != nil {
			return nil, err
		}
		return anypb.New(structValue)
	case string:
		structValue, err := structpb.NewStruct(map[string]interface{}{
			"value": v,
		})
		if err != nil {
			return nil, err
		}
		return anypb.New(structValue)
	case bool:
		structValue, err := structpb.NewStruct(map[string]interface{}{
			"value": v,
		})
		if err != nil {
			return nil, err
		}
		return anypb.New(structValue)
	case nil:
		structValue, err := structpb.NewStruct(map[string]interface{}{
			"value": nil,
		})
		if err != nil {
			return nil, err
		}
		return anypb.New(structValue)
	case []interface{}:
		// For arrays, create a struct with a "value" field containing the array
		structValue, err := structpb.NewStruct(map[string]interface{}{
			"value": v,
		})
		if err != nil {
			return nil, err
		}
		return anypb.New(structValue)
	}

	// Handle objects
	if obj, ok := data.(map[string]interface{}); ok {
		// For objects, create a struct directly
		structValue, err := structpb.NewStruct(obj)
		if err != nil {
			return nil, err
		}
		return anypb.New(structValue)
	}

	return nil, fmt.Errorf("unsupported data type: %T", data)
}

// AnyToJSON converts a protobuf Any value to a JSON string
func AnyToJSON(anyValue *anypb.Any) (string, error) {
	if anyValue == nil {
		return "null", nil
	}

	// Unpack the Any value to get the underlying message
	message, err := anyValue.UnmarshalNew()
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal Any value: %v", err)
	}

	// Convert the message to JSON
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message to JSON: %v", err)
	}

	return string(jsonBytes), nil
}

// ValidateSchema validates a JSON value against a schema
func ValidateSchema(value interface{}, schema *SchemaInfo) error {
	if schema == nil {
		return fmt.Errorf("schema is nil")
	}

	switch schema.StorageType {
	case ScalarData:
		return validateScalarValue(value, schema.TypeInfo)
	case ListData:
		return validateListValue(value, schema)
	case MapData:
		return validateMapValue(value, schema)
	case GraphData:
		return validateGraphValue(value, schema)
	case TabularData:
		return validateTabularValue(value, schema)
	default:
		return fmt.Errorf("unsupported storage type: %v", schema.StorageType)
	}
}

// Helper functions for schema validation
func validateScalarValue(value interface{}, typeInfo *typeinference.TypeInfo) error {
	if value == nil {
		if !typeInfo.IsNullable {
			return fmt.Errorf("value cannot be null")
		}
		return nil
	}

	switch typeInfo.Type {
	case typeinference.StringType:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case typeinference.IntType:
		if _, ok := value.(int); !ok {
			return fmt.Errorf("expected int, got %T", value)
		}
	case typeinference.FloatType:
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("expected float, got %T", value)
		}
	case typeinference.BoolType:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
	default:
		return fmt.Errorf("unsupported type: %v", typeInfo.Type)
	}

	return nil
}

func validateListValue(value interface{}, schema *SchemaInfo) error {
	if value == nil {
		return fmt.Errorf("list value cannot be null")
	}

	list, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected array, got %T", value)
	}

	for i, item := range list {
		if err := ValidateSchema(item, schema.Items); err != nil {
			return fmt.Errorf("invalid item at index %d: %v", i, err)
		}
	}

	return nil
}

func validateMapValue(value interface{}, schema *SchemaInfo) error {
	if value == nil {
		return fmt.Errorf("map value cannot be null")
	}

	obj, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected object, got %T", value)
	}

	for key, propSchema := range schema.Properties {
		if val, exists := obj[key]; exists {
			if err := ValidateSchema(val, propSchema); err != nil {
				return fmt.Errorf("invalid value for key %s: %v", key, err)
			}
		} else if !propSchema.TypeInfo.IsNullable {
			return fmt.Errorf("required key %s is missing", key)
		}
	}

	return nil
}

func validateGraphValue(value interface{}, schema *SchemaInfo) error {
	if value == nil {
		return fmt.Errorf("graph value cannot be null")
	}

	obj, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected object, got %T", value)
	}

	// Validate nodes if present
	if nodes, exists := obj["nodes"]; exists {
		if err := ValidateSchema(nodes, schema.Fields["nodes"]); err != nil {
			return fmt.Errorf("invalid nodes: %v", err)
		}
	}

	// Validate edges if present
	if edges, exists := obj["edges"]; exists {
		if err := ValidateSchema(edges, schema.Fields["edges"]); err != nil {
			return fmt.Errorf("invalid edges: %v", err)
		}
	}

	return nil
}

// validateTabularValue validates a value against a tabular schema
func validateTabularValue(value interface{}, schema *SchemaInfo) error {
	// Convert value to map
	valueMap, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("tabular data must be a map")
	}

	// Check for table field
	tableValue, ok := valueMap["table"]
	if !ok {
		return fmt.Errorf("tabular data must contain a 'table' field")
	}

	// Validate table structure
	tableMap, ok := tableValue.(map[string]interface{})
	if !ok {
		return fmt.Errorf("table field must be a map")
	}

	// Check for columns
	columns, ok := tableMap["columns"]
	if !ok {
		return fmt.Errorf("table must contain 'columns' field")
	}

	// Check for rows
	rows, ok := tableMap["rows"]
	if !ok {
		return fmt.Errorf("table must contain 'rows' field")
	}

	// Validate columns
	columnsList, ok := columns.([]interface{})
	if !ok {
		return fmt.Errorf("columns must be a list")
	}

	// Validate rows
	rowsList, ok := rows.([]interface{})
	if !ok {
		return fmt.Errorf("rows must be a list")
	}

	// Validate that each row has the same number of columns
	expectedColumns := len(columnsList)
	for i, row := range rowsList {
		rowList, ok := row.([]interface{})
		if !ok {
			return fmt.Errorf("row %d must be a list", i)
		}
		if len(rowList) != expectedColumns {
			return fmt.Errorf("row %d has %d columns, expected %d", i, len(rowList), expectedColumns)
		}

		// Validate each cell against its column's schema
		for j, cell := range rowList {
			columnName := columnsList[j].(string)
			if fieldSchema, exists := schema.Fields[columnName]; exists {
				if err := validateScalarValue(cell, fieldSchema.TypeInfo); err != nil {
					return fmt.Errorf("invalid value in row %d, column %s: %v", i, columnName, err)
				}
			}
		}
	}

	return nil
}

// generateSchema generates schema information for a value
func GenerateSchema(value *anypb.Any) (*SchemaInfo, error) {
	// Generate schema directly from the Any value
	schemaGenerator := NewSchemaGenerator()
	return schemaGenerator.GenerateSchema(value)
}

// logSchemaInfo logs schema information in a readable format
func LogSchemaInfo(schemaInfo *SchemaInfo) {
	if schemaInfo == nil {
		log.Printf("Schema is nil")
		return
	}

	// Log the schema information
	log.Printf("Schema: StorageType=%v, TypeInfo=%v",
		schemaInfo.StorageType,
		schemaInfo.TypeInfo)

	// Convert schema to JSON for logging
	schemaJSON, err := SchemaInfoToJSON(schemaInfo)
	if err != nil {
		log.Printf("Failed to convert schema to JSON: %v", err)
		return
	}

	// Marshal to pretty JSON for better readability
	prettyJSON, err := json.MarshalIndent(schemaJSON, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal schema to JSON: %v", err)
		return
	}

	log.Printf("Schema JSON:\n%s", string(prettyJSON))
}
