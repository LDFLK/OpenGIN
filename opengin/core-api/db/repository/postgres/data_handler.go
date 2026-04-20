// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	pb "lk/datafoundation/core-api/lk/datafoundation/core-api"
	"lk/datafoundation/core-api/pkg/schema"
	"lk/datafoundation/core-api/pkg/typeinference"

	commons "lk/datafoundation/core-api/commons"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// filterInternalColumns removes internal columns that shouldn't be returned to the client by default
// unless they are explicitly requested in the fields parameter
func filterInternalColumns(columns []string, requestedFields []string) ([]string, []int) {
	var filteredColumns []string
	var columnIndices []int

	// Define internal columns that should be filtered out by default
	internalColumns := map[string]bool{
		"created_at":          true,
		"entity_attribute_id": true,
		// Note: "id" is NOT filtered out as it's user data
	}

	// Create a set of requested fields for quick lookup
	requestedFieldsSet := make(map[string]bool)
	for _, field := range requestedFields {
		requestedFieldsSet[field] = true
	}

	for i, column := range columns {
		// Keep the column if:
		// 1. It's not an internal column, OR
		// 2. It's an internal column but was explicitly requested
		if !internalColumns[column] || requestedFieldsSet[column] {
			filteredColumns = append(filteredColumns, column)
			columnIndices = append(columnIndices, i)
		}
	}

	return filteredColumns, columnIndices
}

// UnmarshalAnyToString attempts to unmarshal an Any protobuf message to a string value
func UnmarshalAnyToString(anyValue *anypb.Any) (string, error) {
	if anyValue == nil {
		return "", nil
	}

	var stringValue wrapperspb.StringValue
	if err := anyValue.UnmarshalTo(&stringValue); err != nil {
		return "", fmt.Errorf("error unmarshaling to string value: %v", err)
	}
	return stringValue.Value, nil
}

// UnmarshalTimeBasedValueList attempts to unmarshal a TimeBasedValueList from an Any protobuf message
func UnmarshalTimeBasedValueList(anyValue *anypb.Any) ([]interface{}, error) {
	if anyValue == nil {
		return nil, nil
	}

	var timeBasedValueList pb.TimeBasedValueList
	if err := anyValue.UnmarshalTo(&timeBasedValueList); err != nil {
		return nil, fmt.Errorf("error unmarshaling to TimeBasedValueList: %v", err)
	}

	// Convert TimeBasedValueList to []interface{}
	result := make([]interface{}, len(timeBasedValueList.Values))
	for i, v := range timeBasedValueList.Values {
		result[i] = v
	}
	return result, nil
}

// UnmarshalEntityAttributes unmarshals the attributes map from a protobuf Entity
func UnmarshalEntityAttributes(attributes map[string]*anypb.Any) (map[string]interface{}, error) {
	if attributes == nil {
		return nil, nil
	}

	result := make(map[string]interface{})
	for key, value := range attributes {
		if value == nil {
			continue
		}

		// Try to unmarshal as string first
		if strValue, err := UnmarshalAnyToString(value); err == nil {
			result[key] = strValue
			continue
		}

		// Try to unmarshal as TimeBasedValueList
		if timeBasedValues, err := UnmarshalTimeBasedValueList(value); err == nil {
			result[key] = timeBasedValues
			continue
		}

		log.Printf("Warning: Could not unmarshal attribute %s with type %s", key, value.TypeUrl)
	}

	return result, nil
}

// isTabularData checks if the data has a valid tabular structure
// TODO: See if this is needed or else remove it. ,
func isTabularData(value *anypb.Any) (bool, *structpb.Struct, error) {
	// Try to unmarshal as struct
	var dataStruct structpb.Struct
	if err := value.UnmarshalTo(&dataStruct); err != nil {
		return false, nil, fmt.Errorf("failed to unmarshal as struct: %v", err)
	}

	// Check for required fields
	columnsField, hasColumns := dataStruct.Fields["columns"]
	rowsField, hasRows := dataStruct.Fields["rows"]
	if !hasColumns || !hasRows {
		return false, nil, nil
	}

	// Verify columns is a list
	columnsList := columnsField.GetListValue()
	if columnsList == nil {
		return false, nil, nil
	}

	// Verify rows is a list
	rowsList := rowsField.GetListValue()
	if rowsList == nil {
		return false, nil, nil
	}

	// Verify all columns are strings
	for i, col := range columnsList.Values {
		if col.GetStringValue() == "" {
			return false, nil, fmt.Errorf("column %d is not a string", i)
		}
	}

	// Verify all rows have the same number of columns
	columnCount := len(columnsList.Values)
	for i, row := range rowsList.Values {
		rowData := row.GetListValue()
		if rowData == nil {
			return false, nil, fmt.Errorf("row %d is not a list", i)
		}
		if len(rowData.Values) != columnCount {
			return false, nil, fmt.Errorf("row %d has incorrect number of columns", i)
		}
	}

	return true, &dataStruct, nil
}

// isStructpbNull reports whether a structpb scalar should be treated as SQL NULL.
// JSON null may deserialize as either Value_NullValue or a Value with an unset Kind.
func isStructpbNull(v *structpb.Value) bool {
	if v == nil {
		return true
	}
	if v.Kind == nil {
		return true
	}
	_, ok := v.Kind.(*structpb.Value_NullValue)
	return ok
}

// isDateTime checks if a string is a valid datetime
func isDateTime(val string) bool {
	_, err := time.Parse(time.RFC3339, val)
	if err == nil {
		return true
	}

	// IMPROVEME: https://github.com/LDFLK/nexoan/issues/159
	// Try other common formats
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006/01/02",
		"02/01/2006",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, val); err == nil {
			return true
		}
	}

	return false
}

func validateRowsAgainstSchema(data *structpb.Struct, schemaInfo *schema.SchemaInfo) error {
	columnsList := data.Fields["columns"].GetListValue()
	rowsList := data.Fields["rows"].GetListValue()

	if columnsList == nil || rowsList == nil {
		return fmt.Errorf("invalid tabular data: 'columns' and 'rows' must be valid lists")
	}

	for i, row := range rowsList.Values {
		rowData := row.GetListValue()
		if len(rowData.Values) != len(columnsList.Values) {
			return fmt.Errorf("row %d length (%d) does not match number of columns (%d)", i, len(rowData.Values), len(columnsList.Values))
		}
		for j, value := range rowData.Values {
			colName := columnsList.Values[j].GetStringValue()
			fieldSchema, exists := schemaInfo.Fields[colName]
			if !exists {
				return fmt.Errorf("row %d, column %q: not found in schema", i, colName)
			}

			// Null is always allowed (all columns are nullable)
			if isStructpbNull(value) {
				continue
			}

			// Validate type
			switch fieldSchema.TypeInfo.Type {
			case typeinference.NumericType:
				if _, ok := value.Kind.(*structpb.Value_NumberValue); !ok {
					return fmt.Errorf("row %d, column %q: expected numeric, got %T", i, colName, value.Kind)
				}
			case typeinference.IntType:
				if v, ok := value.Kind.(*structpb.Value_NumberValue); !ok || v.NumberValue != float64(int64(v.NumberValue)) {
					return fmt.Errorf("row %d, column %q: expected integer, got %T", i, colName, value.Kind)
				}
			case typeinference.FloatType:
				if _, ok := value.Kind.(*structpb.Value_NumberValue); !ok {
					return fmt.Errorf("row %d, column %q: expected float, got %T", i, colName, value.Kind)
				}
			case typeinference.BoolType:
				if _, ok := value.Kind.(*structpb.Value_BoolValue); !ok {
					return fmt.Errorf("row %d, column %q: expected bool, got %T", i, colName, value.Kind)
				}
			case typeinference.StringType:
				if _, ok := value.Kind.(*structpb.Value_StringValue); !ok {
					return fmt.Errorf("row %d, column %q: expected string, got %T", i, colName, value.Kind)
				}
			case typeinference.DateType:
				if v, ok := value.Kind.(*structpb.Value_StringValue); !ok || !isDateTime(v.StringValue) {
					return fmt.Errorf("row %d, column %q: expected date string, got %T", i, colName, value.Kind)
				}
			case typeinference.DateTimeType:
				if v, ok := value.Kind.(*structpb.Value_StringValue); !ok || !isDateTime(v.StringValue) {
					return fmt.Errorf("row %d, column %q: expected datetime string, got %T", i, colName, value.Kind)
				}
			}
		}
	}

	return nil
}

// compareSchemas compares two schemas and returns true if they are compatible
func compareSchemas(existing, newSchema *schema.SchemaInfo) (bool, error) {
	if existing.StorageType != newSchema.StorageType {
		return false, fmt.Errorf("storage type mismatch: existing=%s, newSchema=%s",
			existing.StorageType, newSchema.StorageType)
	}

	// Check all existing columns are present in newSchema
	for fieldName, existingField := range existing.Fields {
		newField, exists := newSchema.Fields[fieldName]
		if !exists {
			// Missing column in newSchema
			return false, fmt.Errorf("column %s missing in newSchema", fieldName)
		}

		// Check type compatibility
		if !isTypeCompatible(existingField.TypeInfo.Type, newField.TypeInfo.Type) {
			return false, fmt.Errorf("incompatible type for column %s: existing=%s, newSchema=%s",
				fieldName, existingField.TypeInfo.Type, newField.TypeInfo.Type)
		}

		// Check nullability
		if !existingField.TypeInfo.IsNullable && newField.TypeInfo.IsNullable {
			return false, fmt.Errorf("column %s cannot be changed from NOT NULL to NULL", fieldName)
		}
	}

	return true, nil
}

// isTypeCompatible checks if two types are compatible for schema evolution
func isTypeCompatible(existingType, newType typeinference.DataType) bool {
	// Same type is always compatible
	if existingType == newType {
		return true
	}
	// NullType means the incoming batch only had nulls for this column.
	// For existing tables, that's compatible with any existing column type.
	if newType == typeinference.NullType {
		return true
	}

	// Type promotion rules
	switch existingType {
	case typeinference.IntType:
		// Int can be promoted to float
		return newType == typeinference.FloatType
	case typeinference.StringType:
		// String can accept any type
		return true
	case typeinference.DateTimeType:
		// DateTime can only be datetime or string
		return newType == typeinference.StringType
	}

	return false
}

func hasNullOnlyColumns(schemaInfo *schema.SchemaInfo) []string {
	if schemaInfo == nil {
		return nil
	}
	var cols []string
	for fieldName, field := range schemaInfo.Fields {
		if field != nil && field.TypeInfo != nil && field.TypeInfo.Type == typeinference.NullType {
			cols = append(cols, fieldName)
		}
	}
	return cols
}

// StoreTabularData persists tabular attribute data to Postgres, creating the backing table if needed and validating schema compatibility on subsequent writes.
func (repo *PostgresRepository) StoreTabularData(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue, schemaInfo *schema.SchemaInfo) error {
	// Generate table name - UUID without hyphens (32 chars) + prefix (5 chars) = 37 chars total
	name := fmt.Sprintf("%s:%s", entityID, attrName)
	namespace := commons.GetNamespace("attributes")
	unique_id := uuid.NewSHA1(namespace, []byte(name)).String()
	unique_id = strings.ReplaceAll(unique_id, "-", "") // Remove hyphens for database compatibility
	tableName := fmt.Sprintf("attr_%s", unique_id)

	// Unmarshal tabular data once at the beginning to avoid redundant work
	var tabularStruct structpb.Struct
	if err := value.Value.UnmarshalTo(&tabularStruct); err != nil {
		return fmt.Errorf("error unmarshaling tabular data: %v", err)
	}

	// Check if table exists
	exists, err := repo.TableExists(ctx, tableName)
	if err != nil {
		return fmt.Errorf("error checking table existence: %v", err)
	}

	if exists {
		// Get existing schema
		var schemaJSON []byte
		err = repo.DB().QueryRowContext(ctx,
			`SELECT schema_definition FROM attribute_schemas WHERE table_name = $1 ORDER BY schema_version DESC LIMIT 1`,
			tableName).Scan(&schemaJSON)
		if err != nil {
			return fmt.Errorf("error getting existing schema: %v", err)
		}

		var existingSchema schema.SchemaInfo
		if err := json.Unmarshal(schemaJSON, &existingSchema); err != nil {
			return fmt.Errorf("error unmarshaling existing schema: %v", err)
		}

		// Compare schemas only if caller provided one.
		if schemaInfo != nil {
			compatible, err := compareSchemas(&existingSchema, schemaInfo)
			if err != nil {
				return fmt.Errorf("schema compatibility check failed: %v", err)
			}

			if !compatible {
				return fmt.Errorf("incompatible schema changes detected")
			}
		}

		// Validate data against existing schema
		if err := validateRowsAgainstSchema(&tabularStruct, &existingSchema); err != nil {
			return fmt.Errorf("data validation failed: %v", err)
		}
	} else {
		if schemaInfo == nil {
			return fmt.Errorf("schema is required to create a new tabular attribute")
		}
		if nullOnlyCols := hasNullOnlyColumns(schemaInfo); len(nullOnlyCols) > 0 {
			return fmt.Errorf("pre-insert schema validation failed: columns %v have only null values; cannot create new table with null-only inferred types", nullOnlyCols)
		}

		// Validate all rows against the inferred schema before creating the table
		if err := validateRowsAgainstSchema(&tabularStruct, schemaInfo); err != nil {
			return fmt.Errorf("pre-insert schema validation failed: %v", err)
		}

		// Convert schema to columns
		columns := schemaToColumns(schemaInfo)

		// Create new table
		if err := repo.CreateDynamicTable(ctx, tableName, columns); err != nil {
			return fmt.Errorf("error creating table: %v", err)
		}

		// Store schema information
		schemaJSON, err := json.Marshal(schemaInfo)
		if err != nil {
			return fmt.Errorf("error marshaling schema: %v", err)
		}

		// Insert schema record
		_, err = repo.DB().ExecContext(ctx,
			`INSERT INTO attribute_schemas (table_name, schema_version, schema_definition)
			VALUES ($1, $2, $3)`,
			tableName, 1, schemaJSON)
		if err != nil {
			return fmt.Errorf("error storing schema: %v", err)
		}
	}

	// Create entity attribute record if it doesn't exist
	var attributeID int
	err = repo.DB().QueryRowContext(ctx,
		`INSERT INTO entity_attributes (entity_id, attribute_name, table_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (entity_id, attribute_name) DO UPDATE
		SET table_name = EXCLUDED.table_name
		RETURNING id`,
		entityID, attrName, tableName).Scan(&attributeID)
	if err != nil {
		return fmt.Errorf("error creating entity attribute record: %v", err)
	}

	// Extract columns and rows
	columnsValue := tabularStruct.Fields["columns"].GetListValue()
	rowsValue := tabularStruct.Fields["rows"].GetListValue()

	if columnsValue == nil || rowsValue == nil {
		return fmt.Errorf("invalid tabular data format")
	}

	// Convert columns to string slice
	columnNames := make([]string, len(columnsValue.Values))
	for i, col := range columnsValue.Values {
		columnNames[i] = commons.SanitizeIdentifier(col.GetStringValue())
	}

	// Convert rows to [][]interface{}
	rows := make([][]interface{}, len(rowsValue.Values))
	for i, row := range rowsValue.Values {
		rowList := row.GetListValue()
		if rowList == nil {
			return fmt.Errorf("invalid row format at index %d", i)
		}

		rows[i] = make([]interface{}, len(rowList.Values))
		for j, cell := range rowList.Values {
			if isStructpbNull(cell) {
				rows[i][j] = nil
				continue
			}
			switch cell.Kind.(type) {
			case *structpb.Value_StringValue:
				rows[i][j] = cell.GetStringValue()
			case *structpb.Value_NumberValue:
				rows[i][j] = cell.GetNumberValue()
			case *structpb.Value_BoolValue:
				rows[i][j] = cell.GetBoolValue()
			default:
				rows[i][j] = cell.GetStringValue()
			}
		}
	}

	// Insert the data
	if err := repo.InsertTabularData(ctx, tableName, attributeID, columnNames, rows); err != nil {
		return fmt.Errorf("error inserting tabular data: %v", err)
	}

	return nil
}

// schemaToColumns converts a schema to database columns
func schemaToColumns(schemaInfo *schema.SchemaInfo) []Column {
	var columns []Column

	for fieldName, field := range schemaInfo.Fields {
		// Skip "id" columns as they conflict with the auto-generated primary key
		if strings.ToLower(fieldName) == "id" {
			continue
		}

		var colType string
		switch field.TypeInfo.Type {
		case typeinference.NumericType:
			colType = "NUMERIC"
		case typeinference.IntType:
			colType = "INTEGER"
		case typeinference.FloatType:
			colType = "DOUBLE PRECISION"
		case typeinference.StringType:
			colType = "TEXT"
		case typeinference.BoolType:
			colType = "BOOLEAN"
		case typeinference.DateType:
			colType = "DATE"
		case typeinference.DateTimeType:
			colType = "TIMESTAMP WITH TIME ZONE"
		default:
			colType = "TEXT"
		}

		if field.TypeInfo.IsNullable {
			colType += " NULL"
		} else {
			colType += " NOT NULL"
		}

		columns = append(columns, Column{
			Name: commons.SanitizeIdentifier(fieldName),
			Type: colType,
		})
	}

	return columns
}

// Column represents a database column definition
type Column struct {
	Name string
	Type string
}

// GetTableList retrieves a list of attribute tables for a given entity ID.
func GetTableList(ctx context.Context, repo *PostgresRepository, entityID string) ([]string, error) {
	query := `
		SELECT table_name
		FROM entity_attributes
		WHERE entity_id = $1
	`
	rows, err := repo.DB().QueryContext(ctx, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("error querying for table list: %v", err)
	}
	defer rows.Close()

	var tableList []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("error scanning table name: %v", err)
		}
		tableList = append(tableList, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over table list rows: %v", err)
	}

	return tableList, nil
}

// GetSchemaOfTable retrieves the schema for a given attribute table.
func GetSchemaOfTable(ctx context.Context, repo *PostgresRepository, tableName string) (*schema.SchemaInfo, error) {
	query := `
		SELECT schema_definition
		FROM attribute_schemas
		WHERE table_name = $1
		ORDER BY schema_version DESC
		LIMIT 1
	`
	var schemaJSON []byte
	err := repo.DB().QueryRowContext(ctx, query, tableName).Scan(&schemaJSON)
	if err != nil {
		return nil, fmt.Errorf("error getting schema for table %s: %v", tableName, err)
	}

	var schemaInfo schema.SchemaInfo
	if err := json.Unmarshal(schemaJSON, &schemaInfo); err != nil {
		return nil, fmt.Errorf("error unmarshaling schema for table %s: %v", tableName, err)
	}

	return &schemaInfo, nil
}

type columnCaster func(interface{}) (interface{}, bool)

func makeColumnCaster(schemaInfo *schema.SchemaInfo, colName string) columnCaster {
	if schemaInfo == nil {
		return nil
	}
	fieldInfo, exists := schemaInfo.Fields[colName]
	if !exists || fieldInfo == nil || fieldInfo.TypeInfo == nil {
		return nil
	}

	switch fieldInfo.TypeInfo.Type {
	case typeinference.IntType:
		return func(val interface{}) (interface{}, bool) {
			if val == nil {
				return nil, false
			}
			var strVal string
			switch v := val.(type) {
			case []byte:
				strVal = string(v)
			case string:
				strVal = v
			default:
				return nil, false
			}
			intVal, err := strconv.ParseInt(strVal, 10, 64)
			if err != nil {
				return nil, false
			}
			return intVal, true
		}
	case typeinference.NumericType, typeinference.FloatType:
		return func(val interface{}) (interface{}, bool) {
			if val == nil {
				return nil, false
			}
			var strVal string
			switch v := val.(type) {
			case []byte:
				strVal = string(v)
			case string:
				strVal = v
			default:
				return nil, false
			}
			floatVal, err := strconv.ParseFloat(strVal, 64)
			if err != nil {
				return nil, false
			}
			return floatVal, true
		}
	case typeinference.BoolType:
		return func(val interface{}) (interface{}, bool) {
			if val == nil {
				return nil, false
			}
			var strVal string
			switch v := val.(type) {
			case []byte:
				strVal = string(v)
			case string:
				strVal = v
			default:
				return nil, false
			}
			boolVal, err := strconv.ParseBool(strVal)
			if err != nil {
				return nil, false
			}
			return boolVal, true
		}
	default:
		return nil
	}
}

// TabularData represents the structure of tabular data with columns and rows
type TabularData struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
}

// RecordFilter represents a filter applied on the values in a particular column of the table
type RecordFilter struct {
	FieldName string `json:"field_name"`
	Operator  string `json:"operator"` //accepted operators: 'eq','neq','gt','lt','gte','lte','contains','notcontains'
	Value     string `json:"value"`
}

// GetData retrieves data from a table with optional field selection and filters, returns it as pb.Any with JSON-formatted tabular data.
func (repo *PostgresRepository) GetData(ctx context.Context, tableName string, filters map[string]interface{}, fields ...string) (*anypb.Any, error) {
	log.Printf("DEBUG: GetData: tableName=%s, \t\nfilters=%v, \t\nfields=%v", tableName, filters, fields)
	// Build the SELECT clause
	var selectClause string
	if len(fields) > 0 {
		log.Printf("DEBUG: [DataHandler.GetData] selectClause: %v", fields)
		// Sanitize and quote field names
		sanitizedFields := make([]string, len(fields))
		for i, field := range fields {
			sanitizedFields[i] = commons.SanitizeIdentifier(field)
		}
		selectClause = strings.Join(sanitizedFields, ", ")
	} else {
		log.Printf("DEBUG: [DataHandler.GetData] selectClause: *")
		selectClause = "*"
	}

	log.Printf("DEBUG: [DataHandler.GetData] selectClause: %s", selectClause)
	// Base query
	query := fmt.Sprintf("SELECT %s FROM %s", selectClause, commons.SanitizeIdentifier(tableName))

	log.Printf("DEBUG: [DataHandler.GetData] query: %s", query)

	var args []interface{}
	var whereClauses []string
	argCount := 1

	// Add filters to the query
	for key, value := range filters {
		// Handle RecordFilter slices (operator-based row filtering)
		if key == "record_filters" {
			if recordFilters, ok := value.([]RecordFilter); ok {
				for _, rf := range recordFilters {
					sqlOp := "="
					switch rf.Operator {
					case "eq":
						sqlOp = "="
					case "neq":
						sqlOp = "!="
					case "gt":
						sqlOp = ">"
					case "lt":
						sqlOp = "<"
					case "lte":
						sqlOp = "<="
					case "gte":
						sqlOp = ">="
					case "contains":
						sqlOp = "ILIKE"
					case "notcontains":
						sqlOp = "NOT ILIKE"
					}

					if rf.Operator == "contains" || rf.Operator == "notcontains" {
						whereClauses = append(whereClauses, fmt.Sprintf("%s %s $%d", commons.SanitizeIdentifier(rf.FieldName), sqlOp, argCount))
						args = append(args, "%"+rf.Value+"%")
					} else {
						whereClauses = append(whereClauses, fmt.Sprintf("%s %s $%d", commons.SanitizeIdentifier(rf.FieldName), sqlOp, argCount))
						args = append(args, rf.Value)
					}
					argCount++
				}
			}
			continue
		}

		// Default: simple equality filter
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", commons.SanitizeIdentifier(key), argCount))
		args = append(args, value)
		argCount++
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Execute the query
	rows, err := repo.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying data from %s: %v", tableName, err)
	}
	defer rows.Close()

	// Get column names from the result set
	resultColumns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting columns from %s: %v", tableName, err)
	}

	// Filter out internal columns that shouldn't be returned by default
	// unless they are explicitly requested in the fields parameter
	filteredColumns, columnIndices := filterInternalColumns(resultColumns, fields)
	log.Printf("DEBUG: [DataHandler.GetData] Original columns: %v", resultColumns)
	log.Printf("DEBUG: [DataHandler.GetData] Filtered columns: %v", filteredColumns)
	log.Printf("DEBUG: [DataHandler.GetData] Column indices to keep: %v", columnIndices)

	// Log which internal columns were filtered out or included
	internalColumns := map[string]bool{
		"created_at":          true,
		"entity_attribute_id": true,
	}
	for _, column := range resultColumns {
		if internalColumns[column] {
			if len(columnIndices) > 0 && columnIndices[len(columnIndices)-1] >= 0 {
				// Check if this column is in the filtered columns
				found := false
				for _, filteredCol := range filteredColumns {
					if filteredCol == column {
						found = true
						break
					}
				}
				if found {
					log.Printf("INFO: [DataHandler.GetData] Internal column '%s' included (explicitly requested)", column)
				} else {
					log.Printf("INFO: [DataHandler.GetData] Internal column '%s' filtered out (not requested)", column)
				}
			}
		}
	}

	// Get schema for the table to cast values correctly
	schemaInfo, err := GetSchemaOfTable(ctx, repo, tableName)
	if err != nil {
		log.Printf("Warning: failed to get schema for table %s: %v", tableName, err)
	}

	// Precompute per-column casters once to avoid repeated map lookups and type switches per cell.
	columnCasters := make([]columnCaster, len(filteredColumns))
	for i, colName := range filteredColumns {
		columnCasters[i] = makeColumnCaster(schemaInfo, colName)
	}

	var tabularRows [][]interface{}
	for rows.Next() {
		rowValues := make([]interface{}, len(resultColumns))
		rowPointers := make([]interface{}, len(resultColumns))

		for i := range rowValues {
			rowPointers[i] = &rowValues[i]
		}

		if err := rows.Scan(rowPointers...); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		// Convert row values to interface{} slice, but only include filtered columns
		row := make([]interface{}, len(filteredColumns))
		for i, colIndex := range columnIndices {
			val := rowValues[colIndex]

			var assigned bool
			if caster := columnCasters[i]; caster != nil {
				if castedVal, ok := caster(val); ok {
					row[i] = castedVal
					assigned = true
				}
			}

			if !assigned {
				// Handle byte slices (common for text, json, etc.)
				if b, ok := val.([]byte); ok {
					row[i] = string(b)
				} else {
					row[i] = val
				}
			}
		}
		tabularRows = append(tabularRows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	// Create the tabular data structure
	tabularData := map[string]interface{}{
		"columns": filteredColumns,
		"rows":    tabularRows,
	}

	// Convert to JSON string
	jsonData, err := json.Marshal(tabularData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling tabular data to JSON: %v", err)
	}

	log.Printf("DEBUG: [DataHandler.GetData] jsonData: %s", string(jsonData))

	// Create a struct with the JSON string
	structValue, err := structpb.NewStruct(map[string]interface{}{
		"data": string(jsonData),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating struct for JSON data: %v", err)
	}

	// Convert to Any
	anyValue, err := anypb.New(structValue)
	if err != nil {
		return nil, fmt.Errorf("error converting struct to Any: %v", err)
	}

	return anyValue, nil
}
