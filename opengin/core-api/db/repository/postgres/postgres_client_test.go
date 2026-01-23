package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	pb "lk/datafoundation/core-api/lk/datafoundation/core-api"
	schema "lk/datafoundation/core-api/pkg/schema"
	"lk/datafoundation/core-api/pkg/typeinference"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestNewPostgresRepository(t *testing.T) {
	// Build database URI from main PostgreSQL configuration
	dbURI := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_SSL_MODE"))
	t.Logf("dbURI: %s", dbURI)

	// Create new repository using connection string
	repo, err := NewPostgresRepositoryFromDSN(dbURI)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	// Test basic query
	rows, err := repo.DB().Query("SELECT NOW()")
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("Expected at least one row")
	}

	var currentTime time.Time
	if err := rows.Scan(&currentTime); err != nil {
		t.Fatalf("Failed to scan result: %v", err)
	}

	// Verify the time is recent
	if time.Since(currentTime) > time.Minute {
		t.Errorf("Database time seems incorrect: %v", currentTime)
	}
}

// Helper function to create tabular data struct
func createTabularDataStruct(columns []string, rows [][]interface{}) (*anypb.Any, error) {
	// Create columns value
	colValues := make([]*structpb.Value, len(columns))
	for i, col := range columns {
		colValues[i] = structpb.NewStringValue(col)
	}

	// Create rows value
	rowValues := make([]*structpb.Value, len(rows))
	for i, row := range rows {
		rowData := make([]*structpb.Value, len(row))
		for j, val := range row {
			switch v := val.(type) {
			case string:
				rowData[j] = structpb.NewStringValue(v)
			case float64:
				rowData[j] = structpb.NewNumberValue(v)
			case int:
				rowData[j] = structpb.NewNumberValue(float64(v))
			case bool:
				rowData[j] = structpb.NewBoolValue(v)
			case time.Time:
				rowData[j] = structpb.NewStringValue(v.Format(time.RFC3339))
			default:
				// Try to convert to number if it's a numeric type
				if i, ok := val.(int64); ok {
					rowData[j] = structpb.NewNumberValue(float64(i))
				} else if i, ok := val.(int32); ok {
					rowData[j] = structpb.NewNumberValue(float64(i))
				} else if f, ok := val.(float32); ok {
					rowData[j] = structpb.NewNumberValue(float64(f))
				} else {
					rowData[j] = structpb.NewStringValue(fmt.Sprintf("%v", v))
				}
			}
		}
		rowValues[i] = structpb.NewListValue(&structpb.ListValue{Values: rowData})
	}

	// Create the struct
	tabularStruct := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"columns": structpb.NewListValue(&structpb.ListValue{Values: colValues}),
			"rows":    structpb.NewListValue(&structpb.ListValue{Values: rowValues}),
		},
	}

	// Pack into Any
	return anypb.New(tabularStruct)
}

func TestValidateAndReturnTabularDataTypes(t *testing.T) {
	tests := []struct {
		name          string
		columns       []string
		rows          [][]interface{}
		expectedTypes map[string]typeinference.TypeInfo
		expectError   bool
	}{
		{
			name:    "basic types",
			columns: []string{"int_col", "float_col", "string_col", "bool_col", "date_col"},
			rows: [][]interface{}{
				{42, 3.14, "hello", true, "2024-03-14T15:30:45Z"},
				{123, 2.718, "world", false, "2024-03-14"},
			},
			expectedTypes: map[string]typeinference.TypeInfo{
				"int_col": {
					Type: typeinference.IntType,
				},
				"float_col": {
					Type: typeinference.FloatType,
				},
				"string_col": {
					Type: typeinference.StringType,
				},
				"bool_col": {
					Type: typeinference.BoolType,
				},
				"date_col": {
					Type: typeinference.DateTimeType,
				},
			},
			expectError: false,
		},
		{
			name:    "mixed types become strings",
			columns: []string{"mixed_col"},
			rows: [][]interface{}{
				{42},
				{"not a number"},
			},
			expectedTypes: map[string]typeinference.TypeInfo{
				"mixed_col": {
					Type:       typeinference.StringType,
					IsNullable: true,
				},
			},
			expectError: false,
		},
		{
			name:    "all integers",
			columns: []string{"int_col"},
			rows: [][]interface{}{
				{1},
				{2},
				{3},
			},
			expectedTypes: map[string]typeinference.TypeInfo{
				"int_col": {
					Type: typeinference.IntType,
				},
			},
			expectError: false,
		},
		{
			name:    "mixed numbers become float",
			columns: []string{"num_col"},
			rows: [][]interface{}{
				{1},
				{2.5},
			},
			expectedTypes: map[string]typeinference.TypeInfo{
				"num_col": {
					Type: typeinference.FloatType,
				},
			},
			expectError: false,
		},
		{
			name:    "mixed dates become strings",
			columns: []string{"date_col"},
			rows: [][]interface{}{
				{"2024-03-14T15:30:45Z"},
				{"not a date"},
			},
			expectedTypes: map[string]typeinference.TypeInfo{
				"date_col": {
					Type:       typeinference.StringType,
					IsNullable: true,
				},
			},
			expectError: false,
		},
		{
			name:          "empty table",
			columns:       []string{"col1", "col2"},
			rows:          [][]interface{}{},
			expectedTypes: map[string]typeinference.TypeInfo{},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the tabular data struct
			tabularData, err := createTabularDataStruct(tt.columns, tt.rows)
			if err != nil {
				t.Fatalf("Failed to create tabular data struct: %v", err)
			}
			assert.NotNil(t, tabularData, "Failed to create tabular data struct")

			// Unmarshal to structpb.Struct
			var dataStruct structpb.Struct
			err = tabularData.UnmarshalTo(&dataStruct)
			assert.NoError(t, err, "Failed to unmarshal tabular data")

			// Call validateAndReturnTabularDataTypes
			columnTypes, err := validateAndReturnTabularDataTypes(&dataStruct)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedTypes), len(columnTypes),
					"Number of column types mismatch")

				for colName, expectedType := range tt.expectedTypes {
					actualType, exists := columnTypes[colName]
					assert.True(t, exists, "Column %s not found in results", colName)
					assert.Equal(t, expectedType.Type, actualType.Type,
						"Type mismatch for column %s", colName)
					assert.Equal(t, expectedType.IsNullable, actualType.IsNullable,
						"Nullable mismatch for column %s", colName)
				}
			}
		})
	}
}

func TestDateTimeDetection(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"2024-03-01T10:00:00Z", true},
		{"2024-03-01", true},
		{"2024-03-01 15:04:05", true},
		{"2024/03/01", true},
		{"01/03/2024", true},
		{"not a date", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := isDateTime(tc.input)
		assert.Equal(t, tc.expected, result, "DateTime detection failed for %s", tc.input)
	}
}

func TestIsDateTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "RFC3339 format",
			input:    "2024-03-14T15:30:45Z",
			expected: true,
		},
		{
			name:     "RFC3339 with timezone offset",
			input:    "2024-03-14T15:30:45+05:30",
			expected: true,
		},
		{
			name:     "RFC3339 with nanoseconds",
			input:    "2024-03-14T15:30:45.123456789Z",
			expected: true,
		},
		{
			name:     "simple date format",
			input:    "2024-03-14",
			expected: true,
		},
		{
			name:     "date time without timezone",
			input:    "2024-03-14 15:30:45",
			expected: true,
		},
		{
			name:     "date with forward slashes",
			input:    "2024/03/14",
			expected: true,
		},
		{
			name:     "UK date format",
			input:    "14/03/2024",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid format",
			input:    "not a date",
			expected: false,
		},
		{
			name:     "invalid date",
			input:    "2024-13-45",
			expected: false,
		},
		{
			name:     "partial date",
			input:    "2024-03",
			expected: false,
		},
		{
			name:     "numbers only",
			input:    "20240314",
			expected: false,
		},
		{
			name:     "invalid time",
			input:    "2024-03-14 25:70:99",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDateTime(tt.input)
			assert.Equal(t, tt.expected, result, "Test case '%s' failed: input '%s'",
				tt.name, tt.input)
		})
	}
}

// TestInsertSampleData tests inserting various types of sample data
func TestInsertSampleData(t *testing.T) {
	// Build database URI from main PostgreSQL configuration
	dbURI := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_SSL_MODE"))

	// Create new repository using connection string
	repo, err := NewPostgresRepositoryFromDSN(dbURI)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	// Initialize tables
	err = repo.InitializeTables(ctx)
	assert.NoError(t, err, "Failed to initialize tables")

	// Test cases with different types of data
	tests := []struct {
		name       string
		entityID   string
		attrName   string
		columns    []string
		sampleData [][]interface{}
	}{
		{
			name:     "Employee Records",
			entityID: "emp_data",
			attrName: "employee_records",
			columns:  []string{"emp_id", "name", "salary", "join_date", "is_active"},
			sampleData: [][]interface{}{
				{1001, "John Doe", 75000.50, "2024-01-15T09:00:00Z", true},
				{1002, "Jane Smith", 82000.75, "2024-02-01T09:00:00Z", true},
				{1003, "Bob Wilson", 65000.25, "2024-03-01T09:00:00Z", false},
			},
		},
		{
			name:     "Product Inventory",
			entityID: "inventory",
			attrName: "product_stock",
			columns:  []string{"product_id", "name", "quantity", "price", "last_updated"},
			sampleData: [][]interface{}{
				{"P001", "Laptop", 50, 999.99, "2024-03-15T10:30:00Z"},
				{"P002", "Mouse", 200, 29.99, "2024-03-15T10:30:00Z"},
				{"P003", "Keyboard", 150, 59.99, "2024-03-15T10:30:00Z"},
			},
		},
		{
			name:     "Sensor Readings",
			entityID: "sensor_data",
			attrName: "temperature_readings",
			columns:  []string{"sensor_id", "location", "temperature", "humidity", "timestamp"},
			sampleData: [][]interface{}{
				{"S001", "Room A", 23.5, 45.0, "2024-03-15T10:00:00Z"},
				{"S002", "Room B", 24.2, 47.0, "2024-03-15T10:05:00Z"},
				{"S003", "Room C", 22.8, 44.0, "2024-03-15T10:10:00Z"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create TimeBasedValue with the sample data
			dataStruct, err := createTabularDataStruct(tt.columns, tt.sampleData)
			assert.NoError(t, err, "Failed to create tabular data struct")

			timeBasedValue := &pb.TimeBasedValue{
				StartTime: time.Now().Format(time.RFC3339),
				EndTime:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				Value:     dataStruct,
			}

			schemaInfo, err := schema.GenerateSchema(dataStruct)
			assert.NoError(t, err, "Failed to generate schema")

			// Handle attributes (this will create table and insert data)
			err = repo.HandleTabularData(ctx, tt.entityID, tt.attrName, timeBasedValue, schemaInfo)
			assert.NoError(t, err, "Failed to handle attributes")

			// Verify table exists by retrieving the actual table name from entity_attributes
			var tableName string
			err = repo.DB().QueryRowContext(ctx,
				`SELECT table_name FROM entity_attributes WHERE entity_id = $1 AND attribute_name = $2`,
				tt.entityID, tt.attrName).Scan(&tableName)
			assert.NoError(t, err, "Failed to retrieve table name from entity_attributes")
			assert.NotEmpty(t, tableName, "Table name should not be empty")

			exists, err := repo.TableExists(ctx, tableName)
			assert.NoError(t, err, "Failed to check table existence")
			assert.True(t, exists, "Table should exist")

			// Query the data to verify insertion
			query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
			var count int
			err = repo.DB().QueryRowContext(ctx, query).Scan(&count)
			assert.NoError(t, err, "Failed to query data")
			assert.Equal(t, len(tt.sampleData), count, "Row count should match")

			// Query and verify a sample value
			query = fmt.Sprintf("SELECT %s FROM %s LIMIT 1", tt.columns[1], tableName)
			var sampleValue string
			err = repo.DB().QueryRowContext(ctx, query).Scan(&sampleValue)
			assert.NoError(t, err, "Failed to query sample value")
			assert.NotEmpty(t, sampleValue, "Sample value should not be empty")
		})
	}
}

// TestQuerySampleData tests querying the inserted sample data
func TestQuerySampleData(t *testing.T) {
	// Build database URI from main PostgreSQL configuration
	dbURI := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_SSL_MODE"))

	// Create new repository using connection string
	repo, err := NewPostgresRepositoryFromDSN(dbURI)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	// Test queries for different tables
	queries := []struct {
		name       string
		tableName  string
		query      string
		expectRows bool
	}{
		{
			name:       "Query Employee Salaries",
			tableName:  "attr_emp_data_employee_records",
			query:      "SELECT name, salary FROM attr_emp_data_employee_records WHERE salary > 70000",
			expectRows: true,
		},
		{
			name:       "Query Active Employees",
			tableName:  "attr_emp_data_employee_records",
			query:      "SELECT name FROM attr_emp_data_employee_records WHERE is_active = true",
			expectRows: true,
		},
		{
			name:       "Query Product Stock",
			tableName:  "attr_inventory_product_stock",
			query:      "SELECT name, quantity FROM attr_inventory_product_stock WHERE quantity > 100",
			expectRows: true,
		},
		{
			name:       "Query Sensor Temperature",
			tableName:  "attr_sensor_data_temperature_readings",
			query:      "SELECT location, temperature FROM attr_sensor_data_temperature_readings WHERE temperature > 23",
			expectRows: true,
		},
	}

	for _, tt := range queries {
		t.Run(tt.name, func(t *testing.T) {
			// Check if table exists
			exists, err := repo.TableExists(ctx, tt.tableName)
			if !exists {
				t.Skipf("Skipping test: table %s does not exist", tt.tableName)
			}
			assert.NoError(t, err, "Failed to check table existence")

			// Execute query
			rows, err := repo.DB().QueryContext(ctx, tt.query)
			assert.NoError(t, err, "Failed to execute query")
			defer rows.Close()

			// Count rows
			rowCount := 0
			for rows.Next() {
				rowCount++
			}

			if tt.expectRows {
				assert.Greater(t, rowCount, 0, "Expected to find matching rows")
			}
		})
	}
}
