// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"fmt"
	"log"
	"time"

	"lk/datafoundation/core-api/commons"
	dbcommons "lk/datafoundation/core-api/commons/db"
	pb "lk/datafoundation/core-api/lk/datafoundation/core-api"
	"lk/datafoundation/core-api/pkg/storageinference"

	"google.golang.org/protobuf/types/known/anypb"
)

// DatasetType represents the major type for datasets
const DatasetType = "Dataset"

// DatasetMinorTypes represents the minor types for different storage types
const (
	TabularDataset  = "Tabular"
	GraphDataset    = "Graph"
	DocumentDataset = "Document"
	BlobDataset     = "Blob"
)

// IS_ATTRIBUTE relationship type
const IS_ATTRIBUTE_RELATIONSHIP = "IS_ATTRIBUTE"

// IS_ATTRIBUTE relationship direction
// The reason for outgoing is the attribute we create here is an attribute of the parent entity.
const IS_ATTRIBUTE_RELATIONSHIP_DIRECTION = "OUTGOING"

// GraphMetadataManager handles the reference graph for tracking attributes
type GraphMetadataManager struct {
	// This would typically connect to Neo4j or another graph database
	// For now, we'll define the interface and structure
}

// NewGraphMetadataManager creates a new graph metadata manager
func NewGraphMetadataManager() *GraphMetadataManager {
	return &GraphMetadataManager{}
}

// AttributeMetadata represents metadata for an attribute in the graph
type AttributeMetadata struct {
	EntityID      string
	AttributeID   string
	AttributeName string
	StorageType   storageinference.StorageType
	StoragePath   string // Path/location in the specific storage system
	Created       time.Time
	Updated       time.Time
	EndTime       time.Time
	Schema        map[string]interface{} // Schema information
}

// CreateAttributeNode creates a node in the graph for an attribute
func (g *GraphMetadataManager) CreateAttribute(ctx context.Context, metadata *AttributeMetadata) error {
	fmt.Printf("Creating attribute node: Entity=%s, Attribute=%s, StorageType=%s, Path=%s\n",
		metadata.EntityID, metadata.AttributeName, metadata.StorageType, metadata.StoragePath)
	// create the attribute look up graph
	err := g.createAttributeLookUpGraph(ctx, metadata)
	if err != nil {
		return err
	}
	// TODO: save the attribute values in the matching storage system
	return nil
}

// createAttributeLookUpGraph creates the graph node, relationship, and metadata for an attribute
//
// This function creates a complete attribute representation in the graph database with:
//   - A graph node representing the attribute
//   - A relationship connecting the attribute to its parent entity
//   - Metadata containing attribute information
//
// Graph Node Properties:
//   - attribute_id: Unique identifier for the attribute
//   - attribute_name: Name of the attribute
//   - storage_type: Type of storage (tabular, graph, document, blob)
//   - created: Creation timestamp
//
// Relationship:
//   - Creates an IS_ATTRIBUTE relationship between the attribute and its parent entity
//   - Used for lookups by entity ID and attribute name
//   - Direction: INCOMING (attribute belongs to parent entity)
//
// Metadata Structure:
//   - attribute_id: The attribute ID
//   - storage_path: Path in the storage system
//   - storage_type: Type of storage
//   - updated: Last update timestamp
//   - schema: Schema information as a dictionary
//
// Storage Path Details:
//   - storage_database: Connection details or access method
//   - identifier: Database-specific identifier
//   - Tables: table name
//   - Graphs: root node or map of nodes/relationships
//   - Documents: document name
//   - Blobs: blob name
//
// Additional Metadata:
//   - Data source information
//   - Version details
//   - Verification status
//   - Any other key-value pairs as needed
//
// Note: This method creates the attribute node and relationship but does not create
// the parent entity node itself.
func (g *GraphMetadataManager) createAttributeLookUpGraph(ctx context.Context, metadata *AttributeMetadata) error {
	fmt.Printf("Creating attribute look up graph: Entity=%s, Attribute=%s, StorageType=%s, Path=%s\n",
		metadata.EntityID, metadata.AttributeName, metadata.StorageType, metadata.StoragePath)
	// TODO: Explore a way to update the Look up graph
	// FIXME: https://github.com/LDFLK/nexoan/issues/288

	// create the attribute node in the graph
	// attribute core data is stored in the graph
	// attribute metadata is stored in the mongo database

	// create the attribute node in the graph
	// stored parameters: id, kind, name, created
	attributeNode := &pb.Entity{
		Id: metadata.AttributeID,
		Kind: &pb.Kind{
			Major: "Dataset",
			Minor: string(metadata.StorageType),
		},
		Name:          commons.CreateTimeBasedValue(metadata.Created.Format(time.RFC3339), "", metadata.AttributeName),
		Created:       metadata.Created.Format(time.RFC3339), // contains the data object's time relation with the world
		Terminated:    "",                                    // TODO: Implement invalidating a dataset for a specific time range
		Metadata:      MakeMetadataOfAttributeMetadata(metadata),
		Attributes:    make(map[string]*pb.TimeBasedValueList),
		Relationships: make(map[string]*pb.Relationship),
	}

	// the relationships map needs a unique key for each relationship
	// since the attribute id and the name of the attribute is unique for each attribute
	// among all entities, we can use this to form a unique key for the relationship
	relationshipId := GenerateAttributeRelationshipID(metadata.EntityID, metadata.AttributeName)

	parentNode := &pb.Entity{
		Id:         metadata.EntityID,
		Metadata:   make(map[string]*anypb.Any),
		Attributes: make(map[string]*pb.TimeBasedValueList),
		Relationships: map[string]*pb.Relationship{
			relationshipId: MakeRelationshipFromAttributeMetadata(metadata),
		},
	}

	neo4jRepository, err := dbcommons.GetNeo4jRepository(ctx)
	if err != nil {
		log.Printf("[GraphMetadataManager.CreateAttribute] Error getting Neo4j repository: %v", err)
		return err
	}

	// Check if the attribute node already exists
	existingEntity, err := neo4jRepository.ReadGraphEntity(ctx, metadata.AttributeID)
	if err == nil && existingEntity != nil {
		log.Printf("[GraphMetadataManager.CreateAttribute] Attribute node already exists: %s, skipping creation", metadata.AttributeID)
		// Node already exists, we can still proceed to create/update the relationship
	} else {
		// Node doesn't exist, create it
		success, err := neo4jRepository.HandleGraphEntityCreation(ctx, attributeNode)
		if !success {
			log.Printf("[GraphMetadataManager.CreateAttribute] Error creating attributeNode as a graph entity: %v", err)
			return err
		}
		log.Printf("[GraphMetadataManager.CreateAttribute] Successfully created attribute node for entity: %s, attribute: %s", metadata.EntityID, metadata.AttributeName)

		// FIXME: This means that when updating an attribute we cannot update the relationship
		// FIXME: https://github.com/LDFLK/nexoan/issues/346
		// create the relationship between the entity and the attribute
		err = neo4jRepository.HandleGraphRelationshipsUpdate(ctx, parentNode)
		if err != nil {
			log.Printf("[GraphMetadataManager.CreateAttribute] Error creating relationship between entity and attribute: %v", err)
			return err
		}
	}

	log.Printf("[GraphMetadataManager.CreateAttribute] Successfully created relationship for entity: %s, attribute: %s", metadata.EntityID, metadata.AttributeName)

	// create the attribute metadata in the mongo database
	// stored parameters: attribute_id, attribute_name, storage_type, storage_path, updated, schema
	mongoRepository := dbcommons.GetMongoRepository(ctx)

	// Check if the attribute metadata already exists
	existingMetadata, err := mongoRepository.ReadEntity(ctx, metadata.AttributeID)
	if err == nil && existingMetadata != nil {
		log.Printf("[GraphMetadataManager.CreateAttribute] Attribute metadata already exists: %s, skipping creation", metadata.AttributeID)
	} else {
		// Metadata doesn't exist, create it
		_, err = mongoRepository.CreateEntity(ctx, attributeNode)
		if err != nil {
			log.Printf("[GraphMetadataManager.CreateAttribute] Error creating attribute metadata: %v", err)
			return err
		}
		log.Printf("[GraphMetadataManager.CreateAttribute] Successfully created attribute metadata for entity: %s, attribute: %s", metadata.EntityID, metadata.AttributeName)
	}

	log.Print("Lookup graph created successfully!")

	return nil
}

// MakeMetadataOfAttributeMetadata converts AttributeMetadata to Entity Metadata map
func MakeMetadataOfAttributeMetadata(metadata *AttributeMetadata) map[string]*anypb.Any {
	entityMetadata := make(map[string]*anypb.Any)

	// Add attribute_id
	entityMetadata["attribute_id"] = commons.ConvertStringToAny(metadata.AttributeID)

	// Add storage_path
	entityMetadata["storage_path"] = commons.ConvertStringToAny(metadata.StoragePath)

	// Add storage_type
	entityMetadata["storage_type"] = commons.ConvertStringToAny(string(metadata.StorageType))

	// Add updated timestamp
	entityMetadata["updated"] = commons.ConvertStringToAny(metadata.Updated.Format(time.RFC3339))

	// Add schema information
	if len(metadata.Schema) > 0 {
		entityMetadata["schema"] = commons.ConvertMapToAny(metadata.Schema)
	}

	return entityMetadata
}

// MakeRelationshipProto creates a Relationship protobuf object for IS_ATTRIBUTE relationship
func MakeRelationshipFromAttributeMetadata(metadata *AttributeMetadata) *pb.Relationship {
	return &pb.Relationship{
		Id:              GenerateAttributeRelationshipID(metadata.EntityID, metadata.AttributeName),
		RelatedEntityId: metadata.AttributeID,
		Name:            IS_ATTRIBUTE_RELATIONSHIP,
		StartTime:       metadata.Created.Format(time.RFC3339),
		EndTime:         "", // TODO: Implement invalidating a relationship for a specific time range
		Direction:       IS_ATTRIBUTE_RELATIONSHIP_DIRECTION,
	}
}

// GetAttributeMetadata retrieves metadata for an attribute
func (g *GraphMetadataManager) GetAttribute(ctx context.Context, entityID string, attributeName string, startTime time.Time) (*AttributeMetadata, error) {
	fmt.Printf("Getting attribute metadata: EntityID=%s, AttributeName=%s\n", entityID, attributeName)

	neo4jRepository, err := dbcommons.GetNeo4jRepository(ctx)
	if err != nil {
		log.Printf("[GraphMetadataManager.GetAttribute] Error getting Neo4j repository: %v", err)
		return nil, err
	}

	// Get all IS_ATTRIBUTE relationships for the entity
	filteredRelationships, err := neo4jRepository.ReadFilteredRelationships(ctx, entityID, map[string]interface{}{"name": IS_ATTRIBUTE_RELATIONSHIP, "direction": IS_ATTRIBUTE_RELATIONSHIP_DIRECTION, "startTime": startTime.Format(time.RFC3339)}, "")
	if err != nil {
		log.Printf("[GraphMetadataManager.GetAttribute] Error getting relationships: %v", err)
		return nil, err
	}

	if len(filteredRelationships) == 0 {
		log.Printf("[GraphMetadataManager.GetAttribute] No attributes found for entity %s", entityID)
		return nil, fmt.Errorf("no attributes found for entity %s", entityID)
	}

	fmt.Printf("Number of related entities: %v\n", len(filteredRelationships))

	// Find the specific attribute by name
	var targetAttributeID string
	found := false

	for _, relationship := range filteredRelationships {
		attributeID, ok := relationship["relatedEntityId"].(string)
		if !ok {
			continue
		}

		// Get the attribute entity from Neo4j to check its name
		_, attributeNameTimeBased, _, _, err := neo4jRepository.GetGraphEntity(ctx, attributeID)
		if err != nil {
			log.Printf("[GraphMetadataManager.GetAttribute] Error getting attribute entity %s: %v", attributeID, err)
			continue
		}

		attributeNameStr := commons.ExtractStringFromAny(attributeNameTimeBased.Value)
		fmt.Printf("Search name: '%s' vs Entity name: '%s'\n", attributeName, attributeNameStr)

		// Check if this entity has the target attribute name
		if attributeNameStr == attributeName {
			targetAttributeID = attributeID
			found = true
			break
		}
	}

	if !found {
		log.Printf("[GraphMetadataManager.GetAttribute] Attribute '%s' not found for entity %s", attributeName, entityID)
		return nil, fmt.Errorf("attribute '%s' not found for entity %s", attributeName, entityID)
	}

	// Get the attribute metadata from MongoDB
	mongoRepository := dbcommons.GetMongoRepository(ctx)
	attributeMetadataEntity, err := mongoRepository.ReadEntity(ctx, targetAttributeID)
	if err != nil {
		log.Printf("[GraphMetadataManager.GetAttribute] Error getting attribute metadata from MongoDB for attribute %s (entity %s): %v", targetAttributeID, entityID, err)
		return nil, fmt.Errorf("failed to get attribute metadata from MongoDB for attribute %s (entity %s): %w", targetAttributeID, entityID, err)
	}

	// Extract metadata fields
	storageTypeStr, storagePathStr, updatedStr, schemaMap := commons.ExtractAttributeMetadataFields(attributeMetadataEntity)

	// Convert storage type string to StorageType enum
	storageType := commons.ConvertStorageTypeStringToEnum(storageTypeStr)
	log.Printf("[GraphMetadataManager.GetAttribute] storageType: %s", storageType)

	// Get creation time from the attribute entity
	_, _, createdTimeStr, _, err := neo4jRepository.GetGraphEntity(ctx, targetAttributeID)
	if err != nil {
		log.Printf("[GraphMetadataManager.GetAttribute] Error getting creation time for attribute %s: %v", targetAttributeID, err)
		createdTimeStr = ""
	}

	createdTime := commons.ParseTimestamp(createdTimeStr, fmt.Sprintf("attribute %s (entity %s) creation time", targetAttributeID, entityID))
	updatedTime := commons.ParseTimestamp(updatedStr, fmt.Sprintf("attribute %s (entity %s) update time", targetAttributeID, entityID))

	return &AttributeMetadata{
		EntityID:      entityID,
		AttributeID:   targetAttributeID,
		AttributeName: attributeName,
		StorageType:   storageType,
		StoragePath:   storagePathStr,
		Created:       createdTime,
		Updated:       updatedTime,
		Schema:        schemaMap,
	}, nil
}

// ListEntityAttributes lists all attributes for an entity
func (g *GraphMetadataManager) ListAttributes(ctx context.Context, entityID string) ([]*AttributeMetadata, error) {
	fmt.Printf("Listing attributes for entity: %s\n", entityID)

	neo4jRepository, err := dbcommons.GetNeo4jRepository(ctx)
	if err != nil {
		log.Printf("[GraphMetadataManager.ListAttributes] Error getting Neo4j repository: %v", err)
		return nil, err
	}

	filteredRelationships, err := neo4jRepository.ReadFilteredRelationships(ctx, entityID, map[string]interface{}{"name": IS_ATTRIBUTE_RELATIONSHIP, "direction": IS_ATTRIBUTE_RELATIONSHIP_DIRECTION}, "")
	if err != nil {
		log.Printf("[GraphMetadataManager.ListAttributes] Error getting relationships: %v", err)
		return nil, err
	}

	var attributes []*AttributeMetadata
	for _, relationship := range filteredRelationships {
		attributeID, ok := relationship["relatedEntityId"].(string)
		if !ok {
			continue
		}

		// Verify the attribute exists in Neo4j graph and get creation time
		// stored parameters: id, kind, name, created
		//  out of that the GetGraphEntity returns name and createdTime only and we ignore the terminated in this context.
		// TODO: determine if an attribute needs to be teriminated based on various conditions.
		_, attributeName, createdTimeStr, _, err := neo4jRepository.GetGraphEntity(ctx, attributeID)
		if err != nil {
			log.Printf("[GraphMetadataManager.ListAttributes] Error verifying attribute %s in graph for entity %s: %v", attributeID, entityID, err)
			return nil, fmt.Errorf("failed to verify attribute %s in graph for entity %s: %w", attributeID, entityID, err)
		}

		attributeNameStr := commons.ExtractStringFromAny(attributeName.Value)

		// Get the attribute metadata from the mongo database
		mongoRepository := dbcommons.GetMongoRepository(ctx)
		attributeMetadataEntity, err := mongoRepository.ReadEntity(ctx, attributeID)
		if err != nil {
			log.Printf("[GraphMetadataManager.ListAttributes] Error getting attribute metadata from MongoDB for attribute %s (entity %s): %v", attributeID, entityID, err)
			return nil, fmt.Errorf("failed to get attribute metadata from MongoDB for attribute %s (entity %s): %w", attributeID, entityID, err)
		}

		// Extract attribute metadata fields using utility function
		storageTypeStr, storagePathStr, updatedStr, schemaMap := commons.ExtractAttributeMetadataFields(attributeMetadataEntity)

		// Convert storage type string to StorageType enum using utility function
		storageType := commons.ConvertStorageTypeStringToEnum(storageTypeStr)

		// Parse timestamps using utility function
		createdTime := commons.ParseTimestamp(createdTimeStr, fmt.Sprintf("attribute %s (entity %s) creation time", attributeID, entityID))
		updatedTime := commons.ParseTimestamp(updatedStr, fmt.Sprintf("attribute %s (entity %s) update time", attributeID, entityID))

		attrMetadata := &AttributeMetadata{
			EntityID:      entityID,
			AttributeID:   attributeID,
			AttributeName: attributeNameStr,
			StorageType:   storageType,
			StoragePath:   storagePathStr,
			Created:       createdTime,
			Updated:       updatedTime,
			Schema:        schemaMap,
		}
		attributes = append(attributes, attrMetadata)
	}

	return attributes, nil
}

// UpdateAttributeMetadata updates metadata for an attribute
func (g *GraphMetadataManager) UpdateAttribute(ctx context.Context, metadata *AttributeMetadata) error {
	// TODO: Implement Neo4j or graph database connection
	// This would update the attribute node properties
	fmt.Printf("Updating attribute metadata: Entity=%s, Attribute=%s\n", metadata.EntityID, metadata.AttributeName)

	return nil
}

// DeleteAttributeNode deletes an attribute node and its relationships
func (g *GraphMetadataManager) DeleteAttribute(ctx context.Context, entityID, attributeName string) error {
	// TODO: Implement Neo4j or graph database connection
	// This would delete the attribute node and its IS_ATTRIBUTE relationship

	fmt.Printf("Deleting attribute node: Entity=%s, Attribute=%s\n", entityID, attributeName)

	return nil
}

// GetDatasetType returns the appropriate dataset type for a storage type
func GetDatasetType(storageType storageinference.StorageType) string {
	switch storageType {
	case storageinference.TabularData:
		return TabularDataset
	case storageinference.GraphData:
		return GraphDataset
	case storageinference.MapData, storageinference.ListData, storageinference.ScalarData:
		return DocumentDataset
	default:
		return BlobDataset
	}
}

// GenerateAttributeID generates a unique ID for an attribute
func GenerateAttributeRelationshipID(entityID, attributeName string) string {
	return fmt.Sprintf("%s_is_attribute_of_%s", attributeName, entityID)
}

func GenerateAttributeID(entityID, attributeName string) string {
	// attribute name should be unique within an entity
	return fmt.Sprintf("%s_attr_%s", entityID, attributeName)
}

// GenerateStoragePath generates a storage path for an attribute
func GenerateStoragePath(entityID, attributeName string, storageType storageinference.StorageType) string {
	switch storageType {
	case storageinference.TabularData:
		// TODO: This is a placeholder if data are stored in multiple databases.
		// 	This needs to be thought through and implemented such that we define a schema
		//   for the user to define that kind of information.
		return fmt.Sprintf("tables/attr_%s_%s", entityID, attributeName)
	case storageinference.GraphData:
		return fmt.Sprintf("graphs/attr_%s_%s", entityID, attributeName)
	case storageinference.MapData, storageinference.ListData, storageinference.ScalarData:
		return fmt.Sprintf("documents/attr_%s_%s", entityID, attributeName)
	default:
		return fmt.Sprintf("unknown/attr_%s_%s", entityID, attributeName)
	}
}
