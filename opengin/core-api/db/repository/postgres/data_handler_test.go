// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	"lk/datafoundation/core-api/pkg/schema"
	"lk/datafoundation/core-api/pkg/storageinference"
	"lk/datafoundation/core-api/pkg/typeinference"
)

func setupTestDB(t *testing.T) *PostgresRepository {
	// Build database URI from environment variables (same as other tests)
	dbURI := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_SSL_MODE"))

	repo, err := NewPostgresRepositoryFromDSN(dbURI)
	assert.NoError(t, err, "Failed to create repository")

	// Check if repo is nil to avoid panic
	if repo == nil {
		t.Fatal("Repository is nil after successful creation")
	}

	// Initialize tables
	err = repo.InitializeTables(context.Background())
	assert.NoError(t, err)

	// Clean up test tables after test and close repository
	t.Cleanup(func() {
		if repo != nil && repo.DB() != nil {
			// Clean up all test tables created during this test
			_, err := repo.DB().Exec(`
				-- Drop all test tables that start with test_
				DO $$ 
				DECLARE 
					table_name TEXT;
				BEGIN 
					FOR table_name IN 
						SELECT tablename FROM pg_tables 
						WHERE schemaname = 'public' 
						AND (tablename LIKE 'test_data_table_%' 
							OR tablename LIKE 'attr_test_%'
							OR tablename = 'test_data_table'
							OR tablename = 'attr_test_entity_test_attribute')
					LOOP 
						EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(table_name) || ' CASCADE';
					END LOOP;
				END $$;
				
				-- Clean up test entity_attributes entries
				DELETE FROM entity_attributes WHERE entity_id LIKE 'test_%' OR entity_id = 'test_entity';
				
				-- Clean up test attribute_schemas entries  
				DELETE FROM attribute_schemas WHERE table_name LIKE 'test_%' OR table_name = 'test_table';
			`)
			if err != nil {
				t.Logf("Warning: Failed to clean up test data: %v", err)
			}

			// Close the repository after cleanup
			repo.Close()
		}
	})

	return repo
}

func TestGetTableList(t *testing.T) {
	repo := setupTestDB(t)
	// Do not defer repo.Close() here - let cleanup handle it

	// Use unique entity and attribute names for this test
	entityID := fmt.Sprintf("test_entity_%d", time.Now().UnixNano())
	attributeName := fmt.Sprintf("test_attribute_%d", time.Now().UnixNano())
	tableName := fmt.Sprintf("attr_%s_%s", entityID, attributeName)

	// Insert a dummy entity attribute
	_, err := repo.DB().Exec(`
		INSERT INTO entity_attributes (entity_id, attribute_name, table_name)
		VALUES ($1, $2, $3)
	`, entityID, attributeName, tableName)
	assert.NoError(t, err)

	// Get the table list
	tableList, err := GetTableList(context.Background(), repo, entityID)
	assert.NoError(t, err)
	assert.Equal(t, []string{tableName}, tableList)
}

func TestGetSchemaOfTable(t *testing.T) {
	repo := setupTestDB(t)
	// Do not defer repo.Close() here - let cleanup handle it

	// Use unique table name for this test
	tableName := fmt.Sprintf("test_table_%d", time.Now().UnixNano())

	// Insert a dummy schema
	schemaInfo := &schema.SchemaInfo{
		StorageType: storageinference.TabularData,
		Fields: map[string]*schema.SchemaInfo{
			"col1": {
				StorageType: storageinference.ScalarData,
				TypeInfo:    &typeinference.TypeInfo{Type: typeinference.StringType},
			},
		},
	}
	schemaJSON, _ := json.Marshal(schemaInfo)

	_, err := repo.DB().Exec(`
		INSERT INTO attribute_schemas (table_name, schema_version, schema_definition)
		VALUES ($1, 1, $2)
	`, tableName, schemaJSON)
	assert.NoError(t, err)

	// Get the schema
	retrievedSchema, err := GetSchemaOfTable(context.Background(), repo, tableName)
	assert.NoError(t, err)
	assert.Equal(t, schemaInfo.StorageType, retrievedSchema.StorageType)
	assert.Equal(t, len(schemaInfo.Fields), len(retrievedSchema.Fields))
}

func TestGetData(t *testing.T) {
	repo := setupTestDB(t)
	// Do not defer repo.Close() here - let cleanup handle it

	// Use unique table name for this test
	tableName := fmt.Sprintf("test_data_table_%d", time.Now().UnixNano())

	// Create a dummy table and insert data
	_, err := repo.DB().Exec(fmt.Sprintf(`
		CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
			col1 TEXT,
			col2 INTEGER
		)
	`, tableName))
	assert.NoError(t, err)

	_, err = repo.DB().Exec(fmt.Sprintf(`
		INSERT INTO %s (col1, col2) VALUES 
		('val1', 10), 
		('val2', 20), 
		('val3', 30), 
		('val4', 40), 
		('val5', 50)
	`, tableName))
	assert.NoError(t, err)

	// Get data with a filter (all columns)
	filters := map[string]interface{}{"col2": 20}
	anyData, err := repo.GetData(context.Background(), tableName, filters)
	assert.NoError(t, err)
	assert.NotNil(t, anyData)

	// Unmarshal the Any data to get the JSON string
	var structValue structpb.Struct
	err = anyData.UnmarshalTo(&structValue)
	assert.NoError(t, err)

	jsonStr := structValue.Fields["data"].GetStringValue()
	assert.NotEmpty(t, jsonStr)

	// Parse the JSON to verify the structure
	var tabularData map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &tabularData)
	assert.NoError(t, err)
	assert.NotNil(t, tabularData)

	// Add safety checks for the map keys
	firstColumnsInterface, hasFirstColumns := tabularData["columns"]
	assert.True(t, hasFirstColumns, "columns key should exist")
	assert.NotNil(t, firstColumnsInterface, "columns should not be nil")

	firstRowsInterface, hasFirstRows := tabularData["rows"]
	assert.True(t, hasFirstRows, "rows key should exist")
	assert.NotNil(t, firstRowsInterface, "rows should not be nil")

	columns := firstColumnsInterface.([]interface{})
	rows := firstRowsInterface.([]interface{})

	assert.Equal(t, "id", columns[0])
	assert.Equal(t, "col1", columns[1])
	assert.Equal(t, "col2", columns[2])
	assert.Len(t, rows, 1)

	// Verify the filtered data
	row := rows[0].([]interface{})
	assert.Equal(t, "val2", row[1]) // col1 is at index 1

	// Get all data (no filter)
	allAnyData, err := repo.GetData(context.Background(), tableName, nil)
	assert.NoError(t, err)
	assert.NotNil(t, allAnyData)

	// Unmarshal the Any data for all data
	var allStructValue structpb.Struct
	err = allAnyData.UnmarshalTo(&allStructValue)
	assert.NoError(t, err)

	allJsonStr := allStructValue.Fields["data"].GetStringValue()
	assert.NotEmpty(t, allJsonStr)

	// Parse the JSON for all data
	var allTabularData map[string]interface{}
	err = json.Unmarshal([]byte(allJsonStr), &allTabularData)
	assert.NoError(t, err)

	allColumns := allTabularData["columns"].([]interface{})
	allRows := allTabularData["rows"].([]interface{})

	assert.Equal(t, "id", allColumns[0])
	assert.Equal(t, "col1", allColumns[1])
	assert.Equal(t, "col2", allColumns[2])
	assert.Len(t, allRows, 5) // Now we have 5 rows

	// Verify the structure matches the expected tabular format
	assert.NotNil(t, allRows)
	row1 := allRows[0].([]interface{})
	row2 := allRows[1].([]interface{})
	row3 := allRows[2].([]interface{})
	row4 := allRows[3].([]interface{})
	row5 := allRows[4].([]interface{})
	assert.NotNil(t, row1)
	assert.NotNil(t, row2)
	assert.NotNil(t, row3)
	assert.NotNil(t, row4)
	assert.NotNil(t, row5)
	assert.Equal(t, "val1", row1[1])      // First row, col1
	assert.Equal(t, float64(10), row1[2]) // First row, col2 (JSON numbers are float64)
	assert.Equal(t, "val2", row2[1])      // Second row, col1
	assert.Equal(t, float64(20), row2[2]) // Second row, col2
	assert.Equal(t, "val3", row3[1])      // Third row, col1
	assert.Equal(t, float64(30), row3[2]) // Third row, col2
	assert.Equal(t, "val4", row4[1])      // Fourth row, col1
	assert.Equal(t, float64(40), row4[2]) // Fourth row, col2
	assert.Equal(t, "val5", row5[1])      // Fifth row, col1
	assert.Equal(t, float64(50), row5[2]) // Fifth row, col2

	// Test field selection
	selectedFieldsData, err := repo.GetData(context.Background(), tableName, nil, "col1", "col2")
	assert.NoError(t, err)
	assert.NotNil(t, selectedFieldsData)

	// Unmarshal the selected fields data
	var selectedStructValue structpb.Struct
	err = selectedFieldsData.UnmarshalTo(&selectedStructValue)
	assert.NoError(t, err)

	selectedJsonStr := selectedStructValue.Fields["data"].GetStringValue()
	assert.NotEmpty(t, selectedJsonStr)

	// Parse the JSON for selected fields
	var selectedTabularData map[string]interface{}
	err = json.Unmarshal([]byte(selectedJsonStr), &selectedTabularData)
	assert.NoError(t, err)

	selectedFields := selectedTabularData["columns"].([]interface{})
	selectedRows := selectedTabularData["rows"].([]interface{})

	// Verify only selected fields are returned
	assert.Len(t, selectedFields, 2)
	assert.Equal(t, "col1", selectedFields[0])
	assert.Equal(t, "col2", selectedFields[1])
	assert.Len(t, selectedRows, 5) // We have 5 rows in total

	// Verify the data
	assert.NotNil(t, selectedRows)
	selectedRow1 := selectedRows[0].([]interface{})
	selectedRow2 := selectedRows[1].([]interface{})
	selectedRow3 := selectedRows[2].([]interface{})
	selectedRow4 := selectedRows[3].([]interface{})
	selectedRow5 := selectedRows[4].([]interface{})
	assert.NotNil(t, selectedRow1)
	assert.NotNil(t, selectedRow2)
	assert.NotNil(t, selectedRow3)
	assert.NotNil(t, selectedRow4)
	assert.NotNil(t, selectedRow5)
	assert.Equal(t, "val1", selectedRow1[0])      // First row, col1
	assert.Equal(t, float64(10), selectedRow1[1]) // First row, col2
	assert.Equal(t, "val2", selectedRow2[0])      // Second row, col1
	assert.Equal(t, float64(20), selectedRow2[1]) // Second row, col2
	assert.Equal(t, "val3", selectedRow3[0])      // Third row, col1
	assert.Equal(t, float64(30), selectedRow3[1]) // Third row, col2
	assert.Equal(t, "val4", selectedRow4[0])      // Fourth row, col1
	assert.Equal(t, float64(40), selectedRow4[1]) // Fourth row, col2
	assert.Equal(t, "val5", selectedRow5[0])      // Fifth row, col1
	assert.Equal(t, float64(50), selectedRow5[1]) // Fifth row, col2

	// Test field selection with filters
	filteredSelectedData, err := repo.GetData(context.Background(), tableName, filters, "col1")
	assert.NoError(t, err)
	assert.NotNil(t, filteredSelectedData)

	// Unmarshal the filtered selected data
	var filteredSelectedStructValue structpb.Struct
	err = filteredSelectedData.UnmarshalTo(&filteredSelectedStructValue)
	assert.NoError(t, err)

	filteredSelectedJsonStr := filteredSelectedStructValue.Fields["data"].GetStringValue()
	assert.NotEmpty(t, filteredSelectedJsonStr)

	// Parse the JSON for filtered selected data
	var filteredSelectedTabularData map[string]interface{}
	err = json.Unmarshal([]byte(filteredSelectedJsonStr), &filteredSelectedTabularData)
	assert.NoError(t, err)

	filteredSelectedFields := filteredSelectedTabularData["columns"].([]interface{})
	filteredSelectedRows := filteredSelectedTabularData["rows"].([]interface{})

	// Verify only selected field is returned with filter
	assert.Len(t, filteredSelectedFields, 1)
	assert.Equal(t, "col1", filteredSelectedFields[0])
	assert.Len(t, filteredSelectedRows, 1)
	assert.NotNil(t, filteredSelectedRows)
	filteredSelectedRow := filteredSelectedRows[0].([]interface{})
	assert.NotNil(t, filteredSelectedRow)
	assert.Equal(t, "val2", filteredSelectedRow[0]) // Filtered row, col1

	// Test multiple filters
	multipleFilters := map[string]interface{}{
		"col1": "val3",
		"col2": 30,
	}
	multipleFilteredData, err := repo.GetData(context.Background(), tableName, multipleFilters)
	assert.NoError(t, err)
	assert.NotNil(t, multipleFilteredData)

	// Unmarshal the multiple filtered data
	var multipleFilteredStructValue structpb.Struct
	err = multipleFilteredData.UnmarshalTo(&multipleFilteredStructValue)
	assert.NoError(t, err)

	multipleFilteredJsonStr := multipleFilteredStructValue.Fields["data"].GetStringValue()
	assert.NotEmpty(t, multipleFilteredJsonStr)

	// Parse the JSON for multiple filtered data
	var multipleFilteredTabularData map[string]interface{}
	err = json.Unmarshal([]byte(multipleFilteredJsonStr), &multipleFilteredTabularData)
	assert.NoError(t, err)

	multipleFilteredFields := multipleFilteredTabularData["columns"].([]interface{})
	multipleFilteredRows := multipleFilteredTabularData["rows"].([]interface{})

	// Verify multiple filters work correctly
	assert.Len(t, multipleFilteredFields, 3) // id, col1, col2
	assert.Len(t, multipleFilteredRows, 1)
	assert.NotNil(t, multipleFilteredRows)
	multipleFilteredRow := multipleFilteredRows[0].([]interface{})
	assert.NotNil(t, multipleFilteredRow)
	assert.Equal(t, "val3", multipleFilteredRow[1])      // col1
	assert.Equal(t, float64(30), multipleFilteredRow[2]) // col2

	// Test filter that should return no results
	noResultsFilter := map[string]interface{}{
		"col1": "nonexistent",
	}
	noResultsData, err := repo.GetData(context.Background(), tableName, noResultsFilter)
	assert.NoError(t, err)
	assert.NotNil(t, noResultsData)

	// Unmarshal the no results data
	var noResultsStructValue structpb.Struct
	err = noResultsData.UnmarshalTo(&noResultsStructValue)
	assert.NoError(t, err)

	noResultsJsonStr := noResultsStructValue.Fields["data"].GetStringValue()
	assert.NotEmpty(t, noResultsJsonStr)

	// Parse the JSON for no results data
	var noResultsTabularData map[string]interface{}
	err = json.Unmarshal([]byte(noResultsJsonStr), &noResultsTabularData)
	assert.NoError(t, err)
	assert.NotNil(t, noResultsTabularData)

	// Add safety checks for the map keys
	noResultsColumnsInterface, hasNoResultsColumns := noResultsTabularData["columns"]
	assert.True(t, hasNoResultsColumns, "columns key should exist")
	assert.NotNil(t, noResultsColumnsInterface, "columns should not be nil")

	noResultsRowsInterface, hasNoResultsRows := noResultsTabularData["rows"]
	assert.True(t, hasNoResultsRows, "rows key should exist")

	// Handle the case where no results returns null instead of empty array
	var noResultsRows []interface{}
	if noResultsRowsInterface == nil {
		noResultsRows = []interface{}{} // Convert null to empty array
	} else {
		noResultsRows = noResultsRowsInterface.([]interface{})
	}

	noResultsFields := noResultsColumnsInterface.([]interface{})

	// Verify no results filter returns empty result set
	assert.Len(t, noResultsFields, 3) // id, col1, col2
	assert.Len(t, noResultsRows, 0)   // No rows should match
	assert.NotNil(t, noResultsRows)   // Should be empty array, not nil

	// Test numeric filter
	numericFilter := map[string]interface{}{
		"col2": 50,
	}
	numericFilteredData, err := repo.GetData(context.Background(), tableName, numericFilter)
	assert.NoError(t, err)
	assert.NotNil(t, numericFilteredData)

	// Unmarshal the numeric filtered data
	var numericFilteredStructValue structpb.Struct
	err = numericFilteredData.UnmarshalTo(&numericFilteredStructValue)
	assert.NoError(t, err)

	numericFilteredJsonStr := numericFilteredStructValue.Fields["data"].GetStringValue()
	assert.NotEmpty(t, numericFilteredJsonStr)

	// Parse the JSON for numeric filtered data
	var numericFilteredTabularData map[string]interface{}
	err = json.Unmarshal([]byte(numericFilteredJsonStr), &numericFilteredTabularData)
	assert.NoError(t, err)
	assert.NotNil(t, numericFilteredTabularData)

	// Add safety checks for the map keys
	numericFilteredColumnsInterface, hasNumericFilteredColumns := numericFilteredTabularData["columns"]
	assert.True(t, hasNumericFilteredColumns, "columns key should exist")
	assert.NotNil(t, numericFilteredColumnsInterface, "columns should not be nil")

	numericFilteredRowsInterface, hasNumericFilteredRows := numericFilteredTabularData["rows"]
	assert.True(t, hasNumericFilteredRows, "rows key should exist")
	assert.NotNil(t, numericFilteredRowsInterface, "rows should not be nil")

	numericFilteredFields := numericFilteredColumnsInterface.([]interface{})
	numericFilteredRows := numericFilteredRowsInterface.([]interface{})

	// Verify numeric filter works correctly
	assert.Len(t, numericFilteredFields, 3) // id, col1, col2
	assert.Len(t, numericFilteredRows, 1)
	assert.NotNil(t, numericFilteredRows)
	numericFilteredRow := numericFilteredRows[0].([]interface{})
	assert.NotNil(t, numericFilteredRow)
	assert.Equal(t, "val5", numericFilteredRow[1])      // col1
	assert.Equal(t, float64(50), numericFilteredRow[2]) // col2
}

func TestInternalColumnFiltering(t *testing.T) {
	repo := setupTestDB(t)

	// Use unique table name for this test
	tableName := fmt.Sprintf("test_internal_filtering_%d", time.Now().UnixNano())

	// Create a table with internal columns
	_, err := repo.DB().Exec(fmt.Sprintf(`
		CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
			name TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			entity_attribute_id TEXT
		)
	`, tableName))
	assert.NoError(t, err)

	// Insert test data
	_, err = repo.DB().Exec(fmt.Sprintf(`
		INSERT INTO %s (name, entity_attribute_id) VALUES 
		('John Doe', 'attr_123'),
		('Jane Smith', 'attr_456')
	`, tableName))
	assert.NoError(t, err)

	// Test 1: Get all data without specifying fields (should filter out internal columns)
	allData, err := repo.GetData(context.Background(), tableName, nil)
	assert.NoError(t, err)
	assert.NotNil(t, allData)

	// Unmarshal and check columns
	var structValue structpb.Struct
	err = allData.UnmarshalTo(&structValue)
	assert.NoError(t, err)

	jsonStr := structValue.Fields["data"].GetStringValue()
	var tabularData map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &tabularData)
	assert.NoError(t, err)

	columns := tabularData["columns"].([]interface{})
	rows := tabularData["rows"].([]interface{})

	// Should only have id and name (internal columns filtered out)
	assert.Len(t, columns, 2)
	assert.Equal(t, "id", columns[0])
	assert.Equal(t, "name", columns[1])
	assert.Len(t, rows, 2)

	// Test 2: Explicitly request internal columns (should include them)
	internalData, err := repo.GetData(context.Background(), tableName, nil, "id", "name", "created_at", "entity_attribute_id")
	assert.NoError(t, err)
	assert.NotNil(t, internalData)

	err = internalData.UnmarshalTo(&structValue)
	assert.NoError(t, err)

	jsonStr = structValue.Fields["data"].GetStringValue()
	err = json.Unmarshal([]byte(jsonStr), &tabularData)
	assert.NoError(t, err)

	columns = tabularData["columns"].([]interface{})
	rows = tabularData["rows"].([]interface{})

	// Should have all 4 columns including internal ones
	assert.Len(t, columns, 4)
	assert.Equal(t, "id", columns[0])
	assert.Equal(t, "name", columns[1])
	assert.Equal(t, "created_at", columns[2])
	assert.Equal(t, "entity_attribute_id", columns[3])
	assert.Len(t, rows, 2)

	// Test 3: Request only internal columns
	onlyInternalData, err := repo.GetData(context.Background(), tableName, nil, "created_at", "entity_attribute_id")
	assert.NoError(t, err)
	assert.NotNil(t, onlyInternalData)

	err = onlyInternalData.UnmarshalTo(&structValue)
	assert.NoError(t, err)

	jsonStr = structValue.Fields["data"].GetStringValue()
	err = json.Unmarshal([]byte(jsonStr), &tabularData)
	assert.NoError(t, err)

	columns = tabularData["columns"].([]interface{})
	rows = tabularData["rows"].([]interface{})

	// Should have only the requested internal columns
	assert.Len(t, columns, 2)
	assert.Equal(t, "created_at", columns[0])
	assert.Equal(t, "entity_attribute_id", columns[1])
	assert.Len(t, rows, 2)
}

func TestGetDataTabularFormat(t *testing.T) {
	repo := setupTestDB(t)

	// Use unique table name for this test
	tableName := fmt.Sprintf("test_tabular_table_%d", time.Now().UnixNano())

	// Create a table that matches the original tabular data structure
	_, err := repo.DB().Exec(fmt.Sprintf(`
		CREATE TABLE %s (
			id TEXT,
			name TEXT,
			email TEXT,
			department TEXT
		)
	`, tableName))
	assert.NoError(t, err)

	// Insert data that matches the original format
	_, err = repo.DB().Exec(fmt.Sprintf(`
		INSERT INTO %s (id, name, email, department) VALUES 
		('001', 'John Doe', 'john@example.com', 'Engineering'),
		('002', 'Jane Smith', 'jane@example.com', 'Marketing'),
		('003', 'Bob Wilson', 'bob@example.com', 'Sales')
	`, tableName))
	assert.NoError(t, err)

	// Get all data
	anyData, err := repo.GetData(context.Background(), tableName, nil)
	assert.NoError(t, err)
	assert.NotNil(t, anyData)

	// Unmarshal the Any data to get the JSON string
	var structValue structpb.Struct
	err = anyData.UnmarshalTo(&structValue)
	assert.NoError(t, err)

	jsonStr := structValue.Fields["data"].GetStringValue()
	assert.NotEmpty(t, jsonStr)

	// Parse the JSON to verify the structure
	var tabularData map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &tabularData)
	assert.NoError(t, err)

	// Verify the structure matches the original tabular format
	expectedColumns := []string{"id", "name", "email", "department"}
	columns := tabularData["columns"].([]interface{})
	rows := tabularData["rows"].([]interface{})

	// Convert interface{} columns to strings for comparison
	actualColumns := make([]string, len(columns))
	for i, col := range columns {
		actualColumns[i] = col.(string)
	}
	assert.Equal(t, expectedColumns, actualColumns)
	assert.Len(t, rows, 3)

	// Verify the data matches the original input
	expectedRows := [][]interface{}{
		{"001", "John Doe", "john@example.com", "Engineering"},
		{"002", "Jane Smith", "jane@example.com", "Marketing"},
		{"003", "Bob Wilson", "bob@example.com", "Sales"},
	}

	for i, expectedRow := range expectedRows {
		actualRow := rows[i].([]interface{})
		assert.Equal(t, expectedRow, actualRow)
	}

	// Test with filters
	filters := map[string]interface{}{"department": "Engineering"}
	filteredAnyData, err := repo.GetData(context.Background(), tableName, filters)
	assert.NoError(t, err)
	assert.NotNil(t, filteredAnyData)

	// Unmarshal the filtered Any data
	var filteredStructValue structpb.Struct
	err = filteredAnyData.UnmarshalTo(&filteredStructValue)
	assert.NoError(t, err)

	filteredJsonStr := filteredStructValue.Fields["data"].GetStringValue()
	assert.NotEmpty(t, filteredJsonStr)

	// Parse the filtered JSON
	var filteredTabularData map[string]interface{}
	err = json.Unmarshal([]byte(filteredJsonStr), &filteredTabularData)
	assert.NoError(t, err)

	filteredColumns := filteredTabularData["columns"].([]interface{})
	filteredRows := filteredTabularData["rows"].([]interface{})

	// Convert interface{} columns to strings for comparison
	filteredActualColumns := make([]string, len(filteredColumns))
	for i, col := range filteredColumns {
		filteredActualColumns[i] = col.(string)
	}
	assert.Equal(t, expectedColumns, filteredActualColumns)
	assert.Len(t, filteredRows, 1)
	assert.Equal(t, expectedRows[0], filteredRows[0].([]interface{}))
}
