// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package storageinference

import (
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

// StorageType represents the type of data structure
type StorageType string

const (
	TabularData StorageType = "tabular"
	ScalarData  StorageType = "scalar"
	ListData    StorageType = "list"
	MapData     StorageType = "map"
	GraphData   StorageType = "graph"
	UnknownData StorageType = "unknown"
)

// TypeInferrer provides functionality to infer data types from protobuf Any values
type StorageInferrer struct{}

// InferType attempts to determine the storage type from a protobuf Any value.
// The function follows a hierarchical approach to identify the storage type:
//
// 1. First, it unpacks the Any value to get the underlying message:
//
//   - Uses UnmarshalNew() to convert the Any value to its concrete type
//
//   - Handles any unmarshaling errors that might occur
//
// 2. For structpb.Struct messages:
//   - Checks for tabular structure (has both "columns" and "rows" fields)
//   - Checks for graph structure (has both "nodes" and "edges" fields)
//   - Checks for list structure (has "value" field with ListValue)
//   - Checks for scalar structure (single field with scalar value)
//   - If none of the above, defaults to MapData
//
// 3. For non-structpb.Struct messages:
//   - Uses reflection to determine the type:
//   - Slice/Array types return ListData
//   - Map types return MapData
//   - All other types return ScalarData
//
// The function returns one of the following StorageType values:
// - TabularData: For data with columns and rows structure
// - GraphData: For data with nodes and edges structure
// - ListData: For array-like data structures
// - MapData: For key-value pair structures
// - ScalarData: For single value data
//
// Example JSON structures for each type:
//
// TabularData:
//
//	{
//	  "columns": ["id", "name"],
//	  "rows": [[1, "John"], [2, "Jane"]]
//	}
//
// GraphData:
//
//	{
//	  "nodes": [{"id": "1", "type": "user"}],
//	  "edges": [{"source": "1", "target": "2"}]
//	}
//
// ListData:
//
//	[1, 2, 3]
//
// MapData:
//
//	{
//	  "key1": "value1",
//	  "key2": "value2"
//	}
//
// ScalarData:
//
//	42
//
// Error handling:
// - Returns error if Any value cannot be unmarshaled
// - Returns error if the message type is not supported
// - Returns error if the structure is invalid
//
// Note: The function prioritizes specific structures over generic ones.
// For example, if a structure has both "items" and "nodes"/"edges",
// it will be classified based on the more specific structure first.
func (si *StorageInferrer) InferType(anyValue *anypb.Any) (StorageType, error) {
	// Unpack the Any value to get the underlying message
	message, err := anyValue.UnmarshalNew()
	if err != nil {
		return UnknownData, err
	}

	// Get the struct value from the message
	structValue, ok := message.(*structpb.Struct)
	if !ok {
		return UnknownData, fmt.Errorf("expected struct value")
	}

	// Check storage types in order of precedence:
	// 1. Tabular (highest priority)
	// 2. Graph
	// 3. List
	// 4. Scalar
	// 5. Map
	// 6. Unknown (lowest priority)

	// Check for tabular data first (highest priority)
	if isTabular(structValue) {
		return TabularData, nil
	}

	// Check for graph data (second priority)
	if isGraph(structValue) {
		return GraphData, nil
	}

	// Check for list data (third priority)
	if isList(structValue) {
		return ListData, nil
	}

	// Check for scalar data (fourth priority)
	if isScalar(structValue) {
		return ScalarData, nil
	}

	// Check for map data (fifth priority)
	if len(structValue.Fields) > 0 {
		return MapData, nil
	}

	// If none of the above, return an error
	return UnknownData, fmt.Errorf("unable to determine storage type: data structure does not match tabular, graph, list, scalar, or map formats")
}

// isTabular checks if a struct represents tabular data
func isTabular(structValue *structpb.Struct) bool {
	// Check if the struct has both columns and rows fields
	columnsField, hasColumns := structValue.Fields["columns"]
	rowsField, hasRows := structValue.Fields["rows"]
	if !hasColumns || !hasRows {
		return false
	}

	// Verify columns is a list
	_, isColumnsList := columnsField.GetKind().(*structpb.Value_ListValue)
	if !isColumnsList {
		return false
	}

	// Verify rows is a list
	_, isRowsList := rowsField.GetKind().(*structpb.Value_ListValue)
	return isRowsList
}

// isGraph checks if a struct represents graph data
func isGraph(structValue *structpb.Struct) bool {
	// A struct is considered a graph if it has both nodes and edges fields
	_, hasNodes := structValue.Fields["nodes"]
	_, hasEdges := structValue.Fields["edges"]
	return hasNodes && hasEdges
}

// isList checks if a struct represents list data
func isList(structValue *structpb.Struct) bool {
	// A struct is considered a list if:
	// 1. It has exactly one field
	// 2. That field contains a list value
	if len(structValue.Fields) != 1 {
		return false
	}

	// Get the single field value
	var listValue *structpb.Value
	for _, v := range structValue.Fields {
		listValue = v
		break
	}

	// Check if the value is a list
	_, ok := listValue.GetKind().(*structpb.Value_ListValue)
	return ok
}

// isScalar checks if a struct represents scalar data
func isScalar(structValue *structpb.Struct) bool {
	// Check if the struct has a single field with a scalar value
	if len(structValue.Fields) == 1 {
		for _, value := range structValue.Fields {
			switch value.GetKind().(type) {
			case *structpb.Value_NumberValue, *structpb.Value_StringValue, *structpb.Value_BoolValue, *structpb.Value_NullValue:
				return true
			}
		}
	}
	return false
}
