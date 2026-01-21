// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package typeinference

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

// DataType represents the primitive or complex data type that can be inferred from a value.
// This type system is designed to map directly to common database column types.
type DataType string

const (
	// Primitive Types
	IntType    DataType = "int"    // Integer values (e.g., 42, -1)
	FloatType  DataType = "float"  // Floating-point numbers (e.g., 3.14, -0.001)
	StringType DataType = "string" // Text data
	BoolType   DataType = "bool"   // Boolean values (true/false)
	NullType   DataType = "null"   // Null values

	// Special Types
	DateType     DataType = "date"     // Date values (e.g., "2024-03-20")
	TimeType     DataType = "time"     // Time values (e.g., "14:30:00")
	DateTimeType DataType = "datetime" // Date and time values (e.g., "2024-03-20T14:30:00Z")
)

// TypeInfo contains both the data type and additional metadata about the type.
// This structure is used to provide detailed type information for schema generation.
type TypeInfo struct {
	Type       DataType             // The inferred data type
	IsNullable bool                 // Whether the type can be null
	IsArray    bool                 // Whether the type is an array
	ArrayType  *TypeInfo            // For array elements, contains the type of array elements
	Properties map[string]*TypeInfo // For map types, contains property types
}

// TypeInferrer provides functionality to infer data types from protobuf Any values.
// It analyzes the structure and content of values to determine their types.
type TypeInferrer struct{}

// InferType attempts to determine the data type from a protobuf Any value.
// It first unpacks the Any value to get the underlying message, then analyzes its structure
// to determine the appropriate type. The function handles both primitive and complex types.
//
// Parameters:
//   - anyValue: A protobuf Any value containing the data to analyze
//
// Returns:
//   - *TypeInfo: A structure containing the inferred type and metadata
//   - error: Any error that occurred during type inference
func (ti *TypeInferrer) InferType(anyValue *anypb.Any) (*TypeInfo, error) {
	// Unpack the Any value to get the underlying message
	message, err := anyValue.UnmarshalNew()
	if err != nil {
		return nil, err
	}

	// Get the reflection value of the message
	rv := reflect.ValueOf(message)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// Handle structpb.Struct
	if structValue, ok := message.(*structpb.Struct); ok {
		// If it's a struct with a single field, use that field's value
		if len(structValue.Fields) == 1 {
			for _, value := range structValue.Fields {
				return ti.inferTypeFromValue(value)
			}
		}
		// Otherwise, treat the entire struct as a map
		return &TypeInfo{Type: StringType}, nil
	}

	// If not a structpb.Struct, infer type directly
	return ti.inferTypeFromReflection(rv)
}

// inferTypeFromValue determines the type from a structpb.Value.
// This function handles the actual type inference logic for different value kinds.
//
// Parameters:
//   - value: A protobuf Value to analyze
//
// Returns:
//   - *TypeInfo: The inferred type information
//   - error: Any error that occurred during inference
func (ti *TypeInferrer) inferTypeFromValue(value *structpb.Value) (*TypeInfo, error) {
	if value == nil {
		return &TypeInfo{Type: NullType, IsNullable: true}, nil
	}

	switch v := value.GetKind().(type) {
	case *structpb.Value_NullValue:
		return &TypeInfo{Type: NullType, IsNullable: true}, nil

	case *structpb.Value_NumberValue:
		num := v.NumberValue
		// Check if the number is a float by comparing it with its integer conversion
		// This works because float64(int64(num)) truncates any decimal part
		// If the original number has a decimal part, the comparison will be false
		if num != float64(int64(num)) {
			return &TypeInfo{Type: FloatType}, nil
		}

		// For zero, check if it was originally a float by looking at the string representation
		// This is necessary because 0.0 and 0 have the same numeric value but different types
		// We convert to string to check if the original format had a decimal point
		if num == 0 {
			// Use %g format to preserve decimal point for zero values
			str := fmt.Sprintf("%g", num)
			if strings.Contains(str, ".") {
				return &TypeInfo{Type: FloatType}, nil
			}
		}
		return &TypeInfo{Type: IntType}, nil

	case *structpb.Value_StringValue:
		str := v.StringValue
		// Check for special string types
		if isDate(str) {
			return &TypeInfo{Type: DateType}, nil
		}
		if isTime(str) {
			return &TypeInfo{Type: TimeType}, nil
		}
		if isDateTime(str) {
			return &TypeInfo{Type: DateTimeType}, nil
		}
		return &TypeInfo{Type: StringType}, nil

	case *structpb.Value_BoolValue:
		return &TypeInfo{Type: BoolType}, nil

	case *structpb.Value_ListValue:
		list := v.ListValue
		if list == nil || len(list.Values) == 0 {
			return &TypeInfo{Type: StringType, IsArray: true}, nil
		}

		// Infer type from first element
		// Assumption: All elements in the list are of the same type
		// This is a common assumption in type systems, though it may not always hold true
		// For example, [1, "two", 3] would be inferred as an array of integers based on the first element
		// A more robust implementation might check all elements or use a union type
		elemType, err := ti.inferTypeFromValue(list.Values[0])
		if err != nil {
			return nil, err
		}

		return &TypeInfo{
			Type:      StringType,
			IsArray:   true,
			ArrayType: elemType,
		}, nil

	case *structpb.Value_StructValue:
		structVal := v.StructValue
		if structVal == nil {
			return &TypeInfo{Type: StringType}, nil
		}

		// For map types, we use StringType as a base type because:
		// 1. In many systems, complex objects are serialized to strings (e.g., JSON)
		// 2. The actual structure and types of the map's contents are handled separately
		//    through the Properties field in the schema
		// 3. This allows for flexible handling of nested structures while maintaining
		//    a simple type system at the top level
		return &TypeInfo{
			Type: StringType,
		}, nil

	default:
		return &TypeInfo{Type: StringType}, nil
	}
}

// inferTypeFromReflection determines the type from a reflection value.
// This function handles type inference for Go native types using reflection.
//
// Parameters:
//   - rv: A reflection Value to analyze
//
// Returns:
//   - *TypeInfo: The inferred type information
//   - error: Any error that occurred during inference
func (ti *TypeInferrer) inferTypeFromReflection(rv reflect.Value) (*TypeInfo, error) {
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &TypeInfo{Type: IntType}, nil

	case reflect.Float32, reflect.Float64:
		return &TypeInfo{Type: FloatType}, nil

	case reflect.String:
		return &TypeInfo{Type: StringType}, nil

	case reflect.Bool:
		return &TypeInfo{Type: BoolType}, nil

	case reflect.Slice, reflect.Array:
		return &TypeInfo{Type: StringType, IsArray: true}, nil

	case reflect.Map:
		return &TypeInfo{Type: StringType}, nil

	case reflect.Struct:
		return &TypeInfo{Type: StringType}, nil

	default:
		return &TypeInfo{Type: StringType}, nil
	}
}

// isDate checks if a string represents a valid date.
// It supports multiple common date formats including:
// - YYYY-MM-DD (e.g., "2024-03-20")
// - DD/MM/YYYY (e.g., "20/03/2024")
// - MM/DD/YYYY (e.g., "03/20/2024")
// - YYYY.MM.DD (e.g., "2024.03.20")
// - DD-MM-YYYY (e.g., "20-03-2024")
// - MM-DD-YYYY (e.g., "03-20-2024")
// - YYYY/MM/DD (e.g., "2024/03/20")
//
// Parameters:
//   - str: The string to check
//
// Returns:
//   - bool: True if the string matches any of the supported date formats
func isDate(str string) bool {
	dateFormats := []string{
		"2006-01-02", // YYYY-MM-DD
		"02/01/2006", // DD/MM/YYYY
		"01/02/2006", // MM/DD/YYYY
		"2006.01.02", // YYYY.MM.DD
		"02-01-2006", // DD-MM-YYYY
		"01-02-2006", // MM-DD-YYYY
		"2006/01/02", // YYYY/MM/DD
	}

	for _, format := range dateFormats {
		if _, err := time.Parse(format, str); err == nil {
			return true
		}
	}
	return false
}

// isTime checks if a string represents a valid time.
// It supports multiple common time formats including:
// - HH:MM:SS (e.g., "14:30:00")
// - HH:MM (e.g., "14:30")
// - h:MM AM/PM (e.g., "2:30 PM")
// - HH:MM:SS.mmm (e.g., "14:30:00.000")
// - HH:MM:SS±HH:MM (e.g., "14:30:00-07:00")
// - HH:MM:SSZ (e.g., "14:30:00Z")
//
// Parameters:
//   - str: The string to check
//
// Returns:
//   - bool: True if the string matches any of the supported time formats
func isTime(str string) bool {
	timeFormats := []string{
		"15:04:05",       // HH:MM:SS
		"15:04",          // HH:MM
		"3:04 PM",        // h:MM AM/PM
		"15:04:05.000",   // HH:MM:SS.mmm
		"15:04:05-07:00", // HH:MM:SS±HH:MM
		"15:04:05Z",      // HH:MM:SSZ
	}

	for _, format := range timeFormats {
		if _, err := time.Parse(format, str); err == nil {
			return true
		}
	}
	return false
}

// isDateTime checks if a string represents a valid datetime.
// It supports multiple common datetime formats including:
// - RFC3339 (e.g., "2024-03-20T14:30:00Z07:00")
// - YYYY-MM-DD HH:MM:SS (e.g., "2024-03-20 14:30:00")
// - YYYY-MM-DDTHH:MM:SS (e.g., "2024-03-20T14:30:00")
// - DD/MM/YYYY HH:MM:SS (e.g., "20/03/2024 14:30:00")
// - MM/DD/YYYY HH:MM:SS (e.g., "03/20/2024 14:30:00")
// - YYYY.MM.DD HH:MM:SS (e.g., "2024.03.20 14:30:00")
// - YYYY-MM-DD HH:MM:SS.mmm (e.g., "2024-03-20 14:30:00.000")
//
// Parameters:
//   - str: The string to check
//
// Returns:
//   - bool: True if the string matches any of the supported datetime formats
func isDateTime(str string) bool {
	datetimeFormats := []string{
		time.RFC3339,              // 2006-01-02T15:04:05Z07:00
		"2006-01-02 15:04:05",     // YYYY-MM-DD HH:MM:SS
		"2006-01-02T15:04:05",     // YYYY-MM-DDTHH:MM:SS
		"02/01/2006 15:04:05",     // DD/MM/YYYY HH:MM:SS
		"01/02/2006 15:04:05",     // MM/DD/YYYY HH:MM:SS
		"2006.01.02 15:04:05",     // YYYY.MM.DD HH:MM:SS
		"2006-01-02 15:04:05.000", // YYYY-MM-DD HH:MM:SS.mmm
	}

	for _, format := range datetimeFormats {
		if _, err := time.Parse(format, str); err == nil {
			return true
		}
	}
	return false
}
