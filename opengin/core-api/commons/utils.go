// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package commons

import (
	"encoding/json"
	"fmt"
	pb "lk/datafoundation/core-api/lk/datafoundation/core-api"
	"lk/datafoundation/core-api/pkg/storageinference"
	"log"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// CreateTimeBasedValue creates a TimeBasedValue with a string value
func CreateTimeBasedValue(startTime, endTime, value string) *pb.TimeBasedValue {
	return &pb.TimeBasedValue{
		StartTime: startTime,
		EndTime:   endTime,
		Value:     ConvertStringToAny(value),
	}
}

func ConvertStringToAny(value string) *anypb.Any {
	any, _ := anypb.New(&wrapperspb.StringValue{
		Value: value,
	})
	return any
}

// ExtractStringFromAny extracts a string value from *anypb.Any
func ExtractStringFromAny(anyValue *anypb.Any) string {
	if anyValue == nil {
		return ""
	}
	var wrapper wrapperspb.StringValue
	if err := anyValue.UnmarshalTo(&wrapper); err == nil {
		return wrapper.Value
	}
	return ""
}

// ConvertStorageTypeStringToEnum converts a storage type string to StorageType enum
func ConvertStorageTypeStringToEnum(storageTypeStr string) storageinference.StorageType {
	switch storageTypeStr {
	case "tabular":
		return storageinference.TabularData
	case "graph":
		return storageinference.GraphData
	case "map":
		return storageinference.MapData
	case "list":
		return storageinference.ListData
	case "scalar":
		return storageinference.ScalarData
	default:
		return storageinference.UnknownData
	}
}

// ExtractAttributeMetadataFields extracts attribute metadata fields from a MongoDB entity
func ExtractAttributeMetadataFields(entity *pb.Entity) (storageTypeStr, storagePath, updatedStr string, schemaMap map[string]interface{}) {
	// attribute_id, attribute_name, storage_type, storage_path, updated, schema
	if entity == nil || entity.Metadata == nil {
		return "", "", "", make(map[string]interface{})
	}

	metadataMap := entity.Metadata

	// Extract basic fields
	storageTypeStr = ExtractStringFromAny(metadataMap["storage_type"])
	storagePath = ExtractStringFromAny(metadataMap["storage_path"])
	updatedStr = ExtractStringFromAny(metadataMap["updated"])
	schemaStr := ExtractStringFromAny(metadataMap["schema"])

	log.Printf("[Commons.ExtractAttributeMetadataFields] storageTypeStr: %s, storagePath: %s, updatedStr: %s, schemaStr: %s", storageTypeStr, storagePath, updatedStr, schemaStr)

	// Convert schema JSON string to map
	schemaMap, err := ConvertJSONStringToMap(schemaStr)
	if err != nil {
		// Return empty map if conversion fails
		schemaMap = make(map[string]interface{})
	}

	return storageTypeStr, storagePath, updatedStr, schemaMap
}

// ConvertMapToAnyMap converts map[string]interface{} to map[string]*anypb.Any
func ConvertMapToAnyMap(input map[string]interface{}) map[string]*anypb.Any {
	result := make(map[string]*anypb.Any)

	for key, value := range input {
		switch v := value.(type) {
		case string:
			result[key] = ConvertStringToAny(v)
		case int, int32, int64:
			result[key] = ConvertStringToAny(fmt.Sprintf("%v", v))
		case float32, float64:
			result[key] = ConvertStringToAny(fmt.Sprintf("%v", v))
		case bool:
			result[key] = ConvertStringToAny(fmt.Sprintf("%v", v))
		case map[string]interface{}:
			// Handle nested maps recursively
			nestedMap := ConvertMapToAnyMap(v)
			// For nested maps, we'll store them as individual key-value pairs with a prefix
			for nestedKey, nestedValue := range nestedMap {
				result[key+"_"+nestedKey] = nestedValue
			}
		case []interface{}:
			// Handle slices by converting to JSON string
			result[key] = ConvertStringToAny(fmt.Sprintf("%v", v))
		default:
			// For complex types, convert to JSON string
			result[key] = ConvertStringToAny(fmt.Sprintf("%v", v))
		}
	}

	return result
}

// ConvertMapToAny converts map[string]interface{} to *anypb.Any
func ConvertMapToAny(input map[string]interface{}) *anypb.Any {
	// For now, convert the map to a JSON-like string representation
	// In a more sophisticated implementation, you could use a proper protobuf message
	jsonStr := fmt.Sprintf("%v", input)
	return ConvertStringToAny(jsonStr)
}

// ConvertJSONStringToMap converts a JSON string to map[string]interface{}
func ConvertJSONStringToMap(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" {
		return make(map[string]interface{}), nil
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON string: %w", err)
	}
	return result, nil
}

// ParseTimestamp parses a timestamp string to time.Time, returns zero value if empty or invalid
func ParseTimestamp(timestampStr string, context string) time.Time {
	if timestampStr == "" {
		return time.Time{} // Return zero value for empty strings
	}

	parsed, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		log.Printf("[Commons.ParseTimestamp] Warning: Failed to parse timestamp '%s' for %s, using zero value", timestampStr, context)
		return time.Time{} // Return zero value for invalid timestamps
	}

	return parsed
}

// SanitizeIdentifier makes a string safe for use as a PostgreSQL identifier
// IMPROVEME: https://github.com/LDFLK/nexoan/issues/160
func SanitizeIdentifier(s string) string {
	// Replace invalid characters with underscores
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, strings.ToLower(s))

	// Ensure it doesn't start with a number
	if len(safe) > 0 && safe[0] >= '0' && safe[0] <= '9' {
		safe = "_" + safe
	}

	return safe
}
