// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"fmt"
	commons "lk/datafoundation/core-api/commons"
	dbcommons "lk/datafoundation/core-api/commons/db"
	pb "lk/datafoundation/core-api/lk/datafoundation/core-api"
	schema "lk/datafoundation/core-api/pkg/schema"
	storageinference "lk/datafoundation/core-api/pkg/storageinference"
	"log"

	"time"

	"google.golang.org/protobuf/types/known/anypb"
)

// Result represents the result of an attribute resolver operation
type Result struct {
	Data    interface{} // Can hold any type of data (TimeBasedValue, error details, etc.)
	Success bool        // Indicates if the operation was successful
	Error   error       // Error if the operation failed
}

// AttributeResolver interface defines the contract for all attribute resolvers
type AttributeResolver interface {
	Initialize() error
	CreateResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) *Result
	ReadResolve(ctx context.Context, entityID, attrName string, filters map[string]interface{}, fields ...string) *Result
	UpdateResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) *Result
	DeleteResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) *Result
	Finalize() error
}

// BaseAttributeResolver provides common functionality for all resolvers
type BaseAttributeResolver struct {
	storageInferrer *storageinference.StorageInferrer
}

func (r *BaseAttributeResolver) Initialize() error {
	r.storageInferrer = &storageinference.StorageInferrer{}
	return nil
}

func (r *BaseAttributeResolver) Finalize() error {
	return nil
}

// EntityAttributeProcessor handles the main processing of Entity objects
type EntityAttributeProcessor struct {
	resolvers    map[storageinference.StorageType]AttributeResolver
	graphManager *GraphMetadataManager
}

// NewEntityAttributeProcessor creates a new processor with all resolvers initialized
func NewEntityAttributeProcessor() *EntityAttributeProcessor {
	processor := &EntityAttributeProcessor{
		resolvers:    make(map[storageinference.StorageType]AttributeResolver),
		graphManager: NewGraphMetadataManager(),
	}

	// Initialize all resolvers
	processor.resolvers[storageinference.GraphData] = &GraphAttributeResolver{}
	processor.resolvers[storageinference.TabularData] = &TabularAttributeResolver{}
	processor.resolvers[storageinference.MapData] = &DocumentAttributeResolver{}

	// Initialize each resolver
	for _, resolver := range processor.resolvers {
		if err := resolver.Initialize(); err != nil {
			fmt.Printf("Warning: failed to initialize resolver: %v\n", err)
		}
	}

	return processor
}

// GetResolver returns the resolver for a specific storage type
func (p *EntityAttributeProcessor) GetResolver(storageType storageinference.StorageType) (AttributeResolver, bool) {
	resolver, exists := p.resolvers[storageType]
	return resolver, exists
}

// ProcessEntityAttributes processes all attributes in an Entity with operation options
// Returns a map of attribute names to their processing results
func (p *EntityAttributeProcessor) ProcessEntityAttributes(ctx context.Context, entity *pb.Entity, operation string, options *Options) map[string]*Result {
	log.Printf("Processing entity attributes at [processor.ProcessEntityAttributes] [operation: %s] [entity: %+v]", operation, entity)
	if entity == nil || entity.Attributes == nil {
		return make(map[string]*Result)
	}

	// Map to store results for each attribute
	attributeResults := make(map[string]*Result)

	// Process each attribute
	for attrName, timeBasedValueList := range entity.Attributes {
		fmt.Printf("DEBUG: Processing attribute[%s] %s\n", operation, attrName)
		if timeBasedValueList == nil {
			log.Printf("DEBUG: Time-based value list is nil for attribute %s", attrName)
			attributeResults[attrName] = &Result{
				Success: true,
				Data:    nil,
				Error:   nil,
			}
			continue
		}

		log.Printf("DEBUG: Time-based value list is not nil for attribute %s, length: %d", attrName, len(timeBasedValueList.Values))

		// Process each time-based value
		for _, value := range timeBasedValueList.Values {
			if value == nil || value.Value == nil {
				continue
			}

			log.Printf("DEBUG: Processing time-based value for attribute %s: %+v", attrName, value)

			// Determine storage type
			storageType, err := p.determineStorageType(value.Value)
			fmt.Printf("DEBUG: Determined storage type[%s] for attribute %s: %s\n", operation, attrName, storageType)
			if err != nil {
				attributeResults[attrName] = &Result{
					Success: false,
					Data:    nil,
					Error:   fmt.Errorf("error determining storage type for attribute %s: %v", attrName, err),
				}
				continue
			}

			// Create or update graph metadata BEFORE processing the attribute
			// NOTE: for the attribute the timestamp is always the value carried at the attribute level
			// not the entity level. The entity level timestamp is used for the entity itself.
			attributeStartTime, _ := time.Parse(time.RFC3339, value.StartTime)
			if err := p.handleAttributeLookUp(ctx, entity.Id, attrName, storageType, operation, attributeStartTime); err != nil {
				attributeResults[attrName] = &Result{
					Success: false,
					Data:    nil,
					Error:   fmt.Errorf("error handling graph metadata for attribute %s: %v", attrName, err),
				}
				continue
			}

			// Get appropriate resolver
			resolver, exists := p.resolvers[storageType]
			if !exists {
				fmt.Printf("Warning: no resolver found for storage type %s, skipping attribute %s\n", storageType, attrName)
				attributeResults[attrName] = &Result{
					Success: false,
					Data:    nil,
					Error:   fmt.Errorf("no resolver found for storage type %s", storageType),
				}
				continue
			}

			// Execute the appropriate operation
			var operationOptions *Options
			if operation == "read" {
				// Use provided options or default to empty filters
				// TODO: Limitation in multi-value attribute reads.
				// FIXME: https://github.com/LDFLK/nexoan/issues/285
				if options != nil {
					operationOptions = options
				} else {
					operationOptions = &Options{
						ReadOptions: &ReadOptions{
							Filters: make(map[string]interface{}),
							Fields:  []string{}, // Empty means all fields
						},
					}
				}
			} else {
				// For non-read operations, pass the options as-is
				operationOptions = options
			}
			result := p.executeOperation(ctx, resolver, operation, entity.Id, attrName, value, operationOptions)

			log.Printf("DEBUG: Result for attribute %s: %+v", attrName, result)

			// Store the result for this attribute
			attributeResults[attrName] = result

			// For read operations, we might want to do something with the result
			if operation == "read" && result.Data != nil {
				fmt.Printf("Read operation completed for attribute %s\n", attrName)
				// TODO: Handle the read result (e.g., store it, return it, etc.)
			}
		}
	}

	return attributeResults
}

// handleAttributeLookUp handles the attribute look up operations
// This is the first step in the attribute processing pipeline.
// It creates the attribute look up metadata and the attribute node in the graph.
// It also creates the IS_ATTRIBUTE relationship between the entity and the attribute.
// It also creates the attribute metadata in the document database.
func (p *EntityAttributeProcessor) handleAttributeLookUp(ctx context.Context, entityID, attrName string, storageType storageinference.StorageType, operation string, startTime time.Time) error {
	// Generate attribute metadata
	fmt.Printf("DEBUG: Handling graph metadata for attribute %s\n", attrName)
	attributeID := GenerateAttributeID(entityID, attrName)
	storagePath := GenerateStoragePath(entityID, attrName, storageType)

	metadata := &AttributeMetadata{
		EntityID:      entityID,
		AttributeID:   attributeID,
		AttributeName: attrName,
		StorageType:   storageType,
		StoragePath:   storagePath,
		Created:       startTime,
		Updated:       time.Now(),
		Schema:        make(map[string]interface{}), // TODO: Extract schema from value
	}

	// Note: endTime parameter is optional and available for future use if needed

	switch operation {
	case "create":
		// Create attribute node in graph
		if err := p.graphManager.CreateAttribute(ctx, metadata); err != nil {
			return fmt.Errorf("failed to create attribute node: %v", err)
		}

	case "update":
		// Update attribute metadata in graph
		if err := p.graphManager.UpdateAttribute(ctx, metadata); err != nil {
			metadata.Updated = time.Now()
			return fmt.Errorf("failed to update attribute metadata: %v", err)
		}

	case "delete":
		// Delete attribute node and relationships from graph
		if err := p.graphManager.DeleteAttribute(ctx, entityID, attrName); err != nil {
			metadata.Updated = time.Now()
			return fmt.Errorf("failed to delete attribute node: %v", err)
		}

	case "read":
		// For read operations, retrieve the attribute metadata from the graph
		attributeMetadata, err := p.graphManager.GetAttribute(ctx, entityID, attrName, startTime)
		if err != nil {
			fmt.Printf("Warning: attribute %s not found in graph metadata for entity %s\n", attrName, entityID)
		} else if attributeMetadata != nil {
			// Store the retrieved metadata for potential use
			fmt.Printf("DEBUG: Retrieved attribute metadata for %s: %+v\n", attrName, attributeMetadata)
		}
	}

	return nil
}

// determineStorageType determines the storage type of a TimeBasedValue
func (p *EntityAttributeProcessor) determineStorageType(anyValue *anypb.Any) (storageinference.StorageType, error) {
	if anyValue == nil {
		return storageinference.UnknownData, fmt.Errorf("anyValue is nil")
	}

	// Use the storage inference logic to determine type
	storageInferrer := &storageinference.StorageInferrer{}
	return storageInferrer.InferType(anyValue)
}

// executeOperation executes the appropriate CORE operation
// Returns a Result object containing operation-specific data and status
// Options contains operation-specific options for all CORE operations
type Options struct {
	// Read operation options
	ReadOptions *ReadOptions

	// Create operation options
	CreateOptions *CreateOptions

	// Update operation options
	UpdateOptions *UpdateOptions

	// Delete operation options
	DeleteOptions *DeleteOptions
}

// ReadOptions contains options for read operations
type ReadOptions struct {
	Filters map[string]interface{}
	Fields  []string
}

// CreateOptions contains options for create operations
type CreateOptions struct {
	// Future: validation options, conflict resolution, etc.
}

// UpdateOptions contains options for update operations
type UpdateOptions struct {
	// Future: partial updates, conflict resolution, etc.
}

// DeleteOptions contains options for delete operations
type DeleteOptions struct {
	// Future: cascade delete, soft delete, etc.
}

// Helper functions to create options for different operations

// NewReadOptions creates options for read operations
func NewReadOptions(filters map[string]interface{}, fields ...string) *Options {
	return &Options{
		ReadOptions: &ReadOptions{
			Filters: filters,
			Fields:  fields,
		},
	}
}

// NewCreateOptions creates options for create operations
func NewCreateOptions(createOpts *CreateOptions) *Options {
	return &Options{
		CreateOptions: createOpts,
	}
}

// NewUpdateOptions creates options for update operations
func NewUpdateOptions(updateOpts *UpdateOptions) *Options {
	return &Options{
		UpdateOptions: updateOpts,
	}
}

// NewDeleteOptions creates options for delete operations
func NewDeleteOptions(deleteOpts *DeleteOptions) *Options {
	return &Options{
		DeleteOptions: deleteOpts,
	}
}

// executeOperation executes the appropriate operation on the given resolver
func (p *EntityAttributeProcessor) executeOperation(ctx context.Context, resolver AttributeResolver, operation, entityID, attrName string, value *pb.TimeBasedValue, options *Options) *Result {
	if resolver == nil {
		return &Result{
			Data:    nil,
			Success: false,
			Error:   fmt.Errorf("resolver is nil"),
		}
	}

	switch operation {
	case "create":
		// TODO: Use CreateOptions when implemented
		log.Printf("Creating attribute %s for entity %s\n", attrName, entityID)
		return resolver.CreateResolve(ctx, entityID, attrName, value)
	case "read":
		// Use provided options or default to empty filters
		log.Printf("Reading attribute %s for entity %s\n", attrName, entityID)
		var filters map[string]interface{}
		var fields []string
		if options != nil && options.ReadOptions != nil {
			filters = options.ReadOptions.Filters
			fields = options.ReadOptions.Fields
		} else {
			filters = make(map[string]interface{})
		}
		return resolver.ReadResolve(ctx, entityID, attrName, filters, fields...)
	case "update":
		log.Printf("Updating attribute %s for entity %s\n", attrName, entityID)
		// TODO: Use UpdateOptions when implemented
		return resolver.UpdateResolve(ctx, entityID, attrName, value)
	case "delete":
		log.Printf("Deleting attribute %s for entity %s\n", attrName, entityID)
		// TODO: Use DeleteOptions when implemented
		return resolver.DeleteResolve(ctx, entityID, attrName, value)
	default:
		return &Result{
			Data:    nil,
			Success: false,
			Error:   fmt.Errorf("unknown operation: %s", operation),
		}
	}
}

// GraphAttributeResolver handles graph data structures with nodes and edges
type GraphAttributeResolver struct {
	BaseAttributeResolver
}

func (r *GraphAttributeResolver) CreateResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) *Result {
	// TODO: implement graph-specific create logic
	// - Validate graph structure (nodes and edges)
	// - Store in graph database (Neo4j)
	// - Handle graph relationships
	fmt.Printf("Creating graph attribute %s for entity %s\n", attrName, entityID)
	return &Result{
		Data:    nil,
		Success: true,
		Error:   nil,
	}
}

func (r *GraphAttributeResolver) ReadResolve(ctx context.Context, entityID, attrName string, filters map[string]interface{}, fields ...string) *Result {
	// TODO: implement graph-specific read logic
	// - Query graph database
	// - Retrieve nodes and edges
	// - Return graph structure
	fmt.Printf("Reading graph attribute %s for entity %s with filters: %+v and fields: %+v\n", attrName, entityID, filters, fields)

	// TODO: Return actual graph data from Neo4j
	// For now, return empty TimeBasedValue
	timeBasedValue := &pb.TimeBasedValue{
		StartTime: "",
		EndTime:   "",
		Value:     nil, // TODO: Convert graph data to Any
	}

	return &Result{
		Data:    timeBasedValue,
		Success: true,
		Error:   nil,
	}
}

func (r *GraphAttributeResolver) UpdateResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) *Result {
	// TODO: implement graph-specific update logic
	// - Update nodes and edges
	// - Handle graph modifications
	// - Maintain graph consistency
	fmt.Printf("Updating graph attribute %s for entity %s\n", attrName, entityID)
	return &Result{
		Data:    nil,
		Success: true,
		Error:   nil,
	}
}

func (r *GraphAttributeResolver) DeleteResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) *Result {
	// TODO: implement graph-specific delete logic
	// - Remove nodes and edges
	// - Clean up relationships
	// - Handle cascading deletes
	fmt.Printf("Deleting graph attribute %s for entity %s\n", attrName, entityID)
	return &Result{
		Data:    nil,
		Success: true,
		Error:   nil,
	}
}

// TabularAttributeResolver handles tabular data structures with columns and rows
type TabularAttributeResolver struct {
	BaseAttributeResolver
}

func (r *TabularAttributeResolver) CreateResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) *Result {
	// TODO: startDate and endDate must be stored in somewhere in the tabular database.
	//  this will be useful for schema evolution setup.
	startDate := value.StartTime
	endDate := value.EndTime

	// validate the data are in tabular shape
	values := value.Value
	if values == nil {
		return &Result{
			Data:    nil,
			Success: false,
			Error:   fmt.Errorf("values are nil"),
		}
	}

	fmt.Printf("Creating tabular attribute %s for entity %s (validated as tabular) from %v to %v\n", attrName, entityID, startDate, endDate)

	repo, err := dbcommons.GetPostgresRepository(ctx)
	if err != nil {
		return &Result{
			Data:    nil,
			Success: false,
			Error:   fmt.Errorf("failed to get Postgres repository: %v", err),
		}
	}

	// Initialize database tables if they don't exist
	if err := repo.InitializeTables(ctx); err != nil {
		return &Result{
			Data:    nil,
			Success: false,
			Error:   fmt.Errorf("failed to initialize database tables: %v", err),
		}
	}

	schemaInfo, err := schema.GenerateSchema(value.Value)
	if err != nil {
		return &Result{
			Data:    nil,
			Success: false,
			Error:   fmt.Errorf("failed to generate schema: %v", err),
		}
	}

	err = repo.HandleTabularData(ctx, entityID, attrName, value, schemaInfo)
	if err != nil {
		return &Result{
			Data:    nil,
			Success: false,
			Error:   fmt.Errorf("failed to handle tabular data: %v", err),
		}
	}

	return &Result{
		Data:    nil,
		Success: true,
		Error:   nil,
	}
}

func (r *TabularAttributeResolver) ReadResolve(ctx context.Context, entityID, attrName string, filters map[string]interface{}, fields ...string) *Result {
	// TODO: implement tabular-specific read logic
	// - Query database table
	// - Retrieve rows and columns
	// - Return tabular structure
	fmt.Printf("[TabularAttributeResolver.ReadResolve] Reading tabular attribute %s for entity %s with filters: %+v and fields: %+v\n", attrName, entityID, filters, fields)

	repo, err := dbcommons.GetPostgresRepository(ctx)
	if err != nil {
		return &Result{
			Data:    nil,
			Success: false,
			Error:   fmt.Errorf("failed to get Postgres repository: %v", err),
		}
	}

	// Get the table name for this attribute
	tableName := fmt.Sprintf("attr_%s_%s", commons.SanitizeIdentifier(entityID), commons.SanitizeIdentifier(attrName))
	log.Printf("[TabularAttributeResolver.ReadResolve] tableName: %s", tableName)

	// Use the GetData method from the repository to retrieve data with filters and fields
	anyData, err := repo.GetData(ctx, tableName, filters, fields...)
	if err != nil {
		return &Result{
			Data:    nil,
			Success: false,
			Error:   fmt.Errorf("failed to get data: %v", err),
		}
	}

	fmt.Printf("Retrieved data from table %s\n", tableName)

	// The data is already in the correct format (pb.Any with JSON)
	timeBasedValue := &pb.TimeBasedValue{
		StartTime: "",
		EndTime:   "",
		Value:     anyData,
	}

	return &Result{
		Data:    timeBasedValue,
		Success: true,
		Error:   nil,
	}
}

func (r *TabularAttributeResolver) UpdateResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) *Result {
	// TODO: implement tabular-specific update logic
	// - Update table schema if needed
	// - Update data rows
	// - Handle schema evolution
	fmt.Printf("Updating tabular attribute %s for entity %s\n", attrName, entityID)
	return &Result{
		Data:    nil,
		Success: true,
		Error:   nil,
	}
}

func (r *TabularAttributeResolver) DeleteResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) *Result {
	// TODO: implement tabular-specific delete logic
	// - Delete data rows
	// - Optionally drop table
	// - Clean up schema
	fmt.Printf("Deleting tabular attribute %s for entity %s\n", attrName, entityID)
	return &Result{
		Data:    nil,
		Success: true,
		Error:   nil,
	}
}

// DocumentAttributeResolver handles document/map data structures with key-value pairs
type DocumentAttributeResolver struct {
	BaseAttributeResolver
}

func (r *DocumentAttributeResolver) CreateResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) *Result {
	// TODO: implement document-specific create logic
	// - Validate document structure
	// - Store in document database (MongoDB)
	// - Handle document indexing
	fmt.Printf("Creating document attribute %s for entity %s\n", attrName, entityID)
	return &Result{
		Data:    nil,
		Success: true,
		Error:   nil,
	}
}

func (r *DocumentAttributeResolver) ReadResolve(ctx context.Context, entityID, attrName string, filters map[string]interface{}, fields ...string) *Result {
	// TODO: implement document-specific read logic
	// - Query document database
	// - Retrieve document structure
	// - Return key-value pairs
	fmt.Printf("Reading document attribute %s for entity %s with filters: %+v and fields: %+v\n", attrName, entityID, filters, fields)

	// TODO: Return actual document data from MongoDB
	// For now, return empty TimeBasedValue
	timeBasedValue := &pb.TimeBasedValue{
		StartTime: "",
		EndTime:   "",
		Value:     nil, // TODO: Convert document data to Any
	}

	return &Result{
		Data:    timeBasedValue,
		Success: true,
		Error:   nil,
	}
}

func (r *DocumentAttributeResolver) UpdateResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) *Result {
	// TODO: implement document-specific update logic
	// - Update document fields
	// - Handle partial updates
	// - Maintain document consistency
	fmt.Printf("Updating document attribute %s for entity %s\n", attrName, entityID)
	return &Result{
		Data:    nil,
		Success: true,
		Error:   nil,
	}
}

func (r *DocumentAttributeResolver) DeleteResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) *Result {
	// TODO: implement document-specific delete logic
	// - Remove document
	// - Clean up indexes
	// - Handle cascading deletes
	fmt.Printf("Deleting document attribute %s for entity %s\n", attrName, entityID)
	return &Result{
		Data:    nil,
		Success: true,
		Error:   nil,
	}
}
