// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"lk/datafoundation/core-api/pkg/schema"

	_ "github.com/lib/pq"
)

// Config holds the database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// PostgresRepository represents a PostgreSQL database repository
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(cfg Config) (*PostgresRepository, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	return NewPostgresRepositoryFromDSN(dsn)
}

// NewPostgresRepositoryFromDSN creates a new PostgreSQL repository from a connection string
func NewPostgresRepositoryFromDSN(dsn string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %v", err)
	}

	return &PostgresRepository{db: db}, nil
}

// Close closes the database connection
func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

// DB returns the underlying *sql.DB instance
func (r *PostgresRepository) DB() *sql.DB {
	return r.db
}

// InitializeTables creates the necessary tables if they don't exist
// The entity_attributes table serves as the core mapping between entities and their dynamic attributes.
// Purpose:
// 1. Entity-Attribute Relationships: Maps each entity (entity_id) to its attributes (attribute_name)
// 2. Dynamic Table References: Stores the actual table name (table_name) where the attribute data is stored
// 3. Schema Versioning: Maintains the current schema version for each attribute
// 4. Relationship Tracking: Enables efficient querying of which attributes belong to which entities
//
// This design allows for:
// - Dynamic attribute storage with different schemas per attribute
// - Efficient querying without having to scan all attribute tables
// - Schema evolution while maintaining backward compatibility
// - Separation of metadata (in this table) from the actual attribute values (in dynamic tables)
//
// Each attribute's actual data is stored in a separate dynamic table (named in table_name)
// which is created on-demand when new attributes are added to an entity.
func (r *PostgresRepository) InitializeTables(ctx context.Context) error {
	// Create entity_attributes table
	entityAttributesSQL := `
	CREATE TABLE IF NOT EXISTS entity_attributes (
		id SERIAL PRIMARY KEY,
		entity_id VARCHAR(255) NOT NULL,
		attribute_name VARCHAR(255) NOT NULL,
		table_name VARCHAR(255) NOT NULL,
		schema_version INT NOT NULL DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(entity_id, attribute_name)
	);`

	// Create attribute_schemas table
	// The attribute_schemas table stores the schema definitions for each dynamic attribute table.
	// Purpose:
	// 1. Schema Storage: Maintains JSON representation of each attribute's data structure
	// 2. Version Control: Tracks schema versions to support schema evolution
	// 3. Documentation: Provides a self-documenting system for attribute data types
	//
	// This table works together with entity_attributes to provide a complete
	// dynamic attribute management system that can handle diverse data types
	// and evolve over time without breaking existing data.
	attributeSchemasSQL := `
	CREATE TABLE IF NOT EXISTS attribute_schemas (
		id SERIAL PRIMARY KEY,
		table_name VARCHAR(255) NOT NULL,
		schema_version INT NOT NULL,
		schema_definition JSONB NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(table_name, schema_version)
	);`

	// Execute the creation queries
	if _, err := r.db.ExecContext(ctx, entityAttributesSQL); err != nil {
		return fmt.Errorf("error creating entity_attributes table: %v", err)
	}

	if _, err := r.db.ExecContext(ctx, attributeSchemasSQL); err != nil {
		return fmt.Errorf("error creating attribute_schemas table: %v", err)
	}

	return nil
}

// TableExists checks if a table exists in the database
func (r *PostgresRepository) TableExists(ctx context.Context, tableName string) (bool, error) {
	query := `
	SELECT EXISTS (
		SELECT FROM pg_tables
		WHERE schemaname = 'public'
		AND tablename = $1
	);`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking table existence: %v", err)
	}

	return exists, nil
}

// CreateDynamicTable creates a new table for storing attribute data
func (r *PostgresRepository) CreateDynamicTable(ctx context.Context, tableName string, columns []Column) error {
	// Build column definitions
	var columnDefs []string

	// Add primary key and entity_attribute_id first
	columnDefs = append(columnDefs, "id SERIAL PRIMARY KEY")
	columnDefs = append(columnDefs, "entity_attribute_id INTEGER REFERENCES entity_attributes(id)")

	// Add the rest of the columns
	for _, col := range columns {
		columnDefs = append(columnDefs, fmt.Sprintf("%s %s", col.Name, col.Type))
	}

	// Add created_at timestamp at the end
	columnDefs = append(columnDefs, "created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP")

	// Create table query
	createTableSQL := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		%s
	);`, tableName, strings.Join(columnDefs, ",\n"))

	// Execute the creation query
	if _, err := r.db.ExecContext(ctx, createTableSQL); err != nil {
		return fmt.Errorf("error creating dynamic table: %v", err)
	}

	return nil
}

// InsertTabularData inserts rows into a dynamic table
func (r *PostgresRepository) InsertTabularData(ctx context.Context, tableName string, entityAttributeID int, columns []string, rows [][]interface{}) error {
	// Build the INSERT query
	columnNames := append([]string{"entity_attribute_id"}, columns...)
	placeholders := make([]string, len(rows))
	valuesPerRow := len(columns) + 1 // +1 for entity_attribute_id

	for i := range rows {
		rowPlaceholders := make([]string, valuesPerRow)
		for j := range rowPlaceholders {
			rowPlaceholders[j] = fmt.Sprintf("$%d", i*valuesPerRow+j+1)
		}
		placeholders[i] = fmt.Sprintf("(%s)", strings.Join(rowPlaceholders, ", "))
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		tableName,
		strings.Join(columnNames, ", "),
		strings.Join(placeholders, ", "),
	)

	// Flatten values for the query
	values := make([]interface{}, 0, len(rows)*valuesPerRow)
	for _, row := range rows {
		values = append(values, entityAttributeID) // Add entity_attribute_id first
		values = append(values, row...)
	}

	// Execute the query
	_, err := r.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("error inserting data: %v", err)
	}

	return nil
}

// GetTableList retrieves a list of attribute tables for a given entity ID.
func (r *PostgresRepository) GetTableList(ctx context.Context, entityID string) ([]string, error) {
	return GetTableList(ctx, r, entityID)
}

// GetSchemaOfTable retrieves the schema for a given attribute table.
func (r *PostgresRepository) GetSchemaOfTable(ctx context.Context, tableName string) (*schema.SchemaInfo, error) {
	return GetSchemaOfTable(ctx, r, tableName)
}
