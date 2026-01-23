package engine

import (
	"context"
	"testing"
	"time"

	"lk/datafoundation/core-api/pkg/storageinference"

	"github.com/stretchr/testify/assert"
)

// TestGraphMetadataManager tests the graph metadata manager functionality
func TestGraphMetadataManager(t *testing.T) {
	manager := NewGraphMetadataManager()
	assert.NotNil(t, manager)

	ctx := context.Background()

	parentEntityID := "engine-test-entity-1"

	createdTime := time.Now()

	// Test creating attribute metadata
	metadata := &AttributeMetadata{
		EntityID:      parentEntityID,
		AttributeID:   "test-attribute-1",
		AttributeName: "test-attribute",
		StorageType:   storageinference.TabularData,
		StoragePath:   "tables/attr_test-entity-1_test-attribute",
		Created:       createdTime,
		Updated:       createdTime,
		Schema: map[string]interface{}{
			"columns": []string{"id", "name"},
			"types":   []string{"int", "string"},
		},
	}

	entity, err := createEntityWithAttributes(parentEntityID, "Engine Test Entity 1", map[string]string{
		"test-attribute": `{"columns": ["id", "name"], "types": ["int", "string"]}`,
	})
	assert.NoError(t, err)
	err = saveEntityToDatabase(ctx, entity)
	assert.NoError(t, err)

	// Test creating attribute node
	err = manager.CreateAttribute(ctx, metadata)
	assert.NoError(t, err)

	// Test getting attribute metadata
	retrievedMetadata, err := manager.GetAttribute(ctx, metadata.EntityID, metadata.AttributeName, createdTime)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedMetadata)
	assert.Equal(t, metadata.EntityID, retrievedMetadata.EntityID)
	assert.Equal(t, metadata.AttributeName, retrievedMetadata.AttributeName)

	// Test updating attribute metadata
	metadata.Updated = time.Now()
	metadata.Schema["new_field"] = "new_value"
	err = manager.UpdateAttribute(ctx, metadata)
	assert.NoError(t, err)

	// Test listing entity attributes
	attributes, err := manager.ListAttributes(ctx, metadata.EntityID)
	assert.NoError(t, err)
	assert.NotNil(t, attributes)
	assert.Equal(t, 1, len(attributes))

	// Test deleting attribute node
	err = manager.DeleteAttribute(ctx, metadata.EntityID, metadata.AttributeName)
	assert.NoError(t, err)
}

// TestDatasetTypeMapping tests the mapping between storage types and dataset types
func TestDatasetTypeMapping(t *testing.T) {
	testCases := map[storageinference.StorageType]string{
		storageinference.TabularData: TabularDataset,
		storageinference.GraphData:   GraphDataset,
		storageinference.MapData:     DocumentDataset,
		storageinference.ListData:    DocumentDataset,
		storageinference.ScalarData:  DocumentDataset,
		storageinference.UnknownData: BlobDataset,
	}

	for storageType, expectedDatasetType := range testCases {
		t.Run(string(storageType), func(t *testing.T) {
			datasetType := GetDatasetType(storageType)
			assert.Equal(t, expectedDatasetType, datasetType)
		})
	}
}


// TestStoragePathGeneration tests the storage path generation
func TestStoragePathGeneration(t *testing.T) {
	entityID := "engine-test-entity-123"
	attributeName := "user_profile"

	testCases := map[storageinference.StorageType]string{
		storageinference.TabularData: "tables/attr_engine-test-entity-123_user_profile",
		storageinference.GraphData:   "graphs/attr_engine-test-entity-123_user_profile",
		storageinference.MapData:     "documents/attr_engine-test-entity-123_user_profile",
		storageinference.ListData:    "documents/attr_engine-test-entity-123_user_profile",
		storageinference.ScalarData:  "documents/attr_engine-test-entity-123_user_profile",
		storageinference.UnknownData: "unknown/attr_engine-test-entity-123_user_profile",
	}

	for storageType, expectedPath := range testCases {
		t.Run(string(storageType), func(t *testing.T) {
			path := GenerateStoragePath(entityID, attributeName, storageType)
			assert.Equal(t, expectedPath, path)
		})
	}
}

// TestGraphMetadataIntegration tests the integration of graph metadata with attribute processing
func TestGraphMetadataIntegration(t *testing.T) {
	// Create an entity with mixed data types
	entity, err := createEntityWithAttributes("engine-id-integration-test-entity-1", "integration-test-entity-1", map[string]string{
		"tabular_data": `{
			"columns": ["id", "name"],
			"rows": [[1, "John"], [2, "Jane"]]
		}`,
		"graph_data": `{
			"nodes": [{"n_id": "user1", "type": "user"}],
			"edges": [{"source": "user1", "target": "user2"}]
		}`,
		"document_data": `{
			"user_profile": {"name": "John", "age": 30}
		}`,
	})
	assert.NoError(t, err)

	processor := NewEntityAttributeProcessor()
	ctx := context.Background()

	// save the parent entity in the database
	err = saveEntityToDatabase(ctx, entity)
	assert.NoError(t, err)

	// Test create operation - this should create graph metadata
	options := getOptionsForOperation("create")
	attributeResults := processor.ProcessEntityAttributes(ctx, entity, "create", options)

	// Check that all attributes were processed successfully
	for attrName, result := range attributeResults {
		assert.True(t, result.Success, "Attribute %s should be processed successfully in create operation", attrName)
		assert.NoError(t, result.Error, "Attribute %s should not have errors in create operation", attrName)
	}

	// Test read operation - this should verify graph metadata
	options = getOptionsForOperation("read")
	attributeResults = processor.ProcessEntityAttributes(ctx, entity, "read", options)

	// Check that all attributes were processed successfully
	for attrName, result := range attributeResults {
		assert.True(t, result.Success, "Attribute %s should be processed successfully in read operation", attrName)
		assert.NoError(t, result.Error, "Attribute %s should not have errors in read operation", attrName)
	}

	// Test update operation - this should update graph metadata
	options = getOptionsForOperation("update")
	attributeResults = processor.ProcessEntityAttributes(ctx, entity, "update", options)

	// Check that all attributes were processed successfully
	for attrName, result := range attributeResults {
		assert.True(t, result.Success, "Attribute %s should be processed successfully in update operation", attrName)
		assert.NoError(t, result.Error, "Attribute %s should not have errors in update operation", attrName)
	}

	// Test delete operation - this should delete graph metadata
	options = getOptionsForOperation("delete")
	attributeResults = processor.ProcessEntityAttributes(ctx, entity, "delete", options)

	// Check that all attributes were processed successfully
	for attrName, result := range attributeResults {
		assert.True(t, result.Success, "Attribute %s should be processed successfully in delete operation", attrName)
		assert.NoError(t, result.Error, "Attribute %s should not have errors in delete operation", attrName)
	}
}

// TestAttributeMetadataStructure tests the AttributeMetadata structure
func TestAttributeMetadataStructure(t *testing.T) {
	metadata := &AttributeMetadata{
		EntityID:      "engine-test-entity-1",
		AttributeID:   "engine-test-entity-1_attr_user_profile",
		AttributeName: "user_profile",
		StorageType:   storageinference.TabularData,
		StoragePath:   "tables/engine-test-entity-1_attr_user_profile",
		Created:       time.Now(),
		Updated:       time.Now(),
		Schema:        make(map[string]interface{}),
	}

	assert.Equal(t, "engine-test-entity-1", metadata.EntityID)
	assert.Equal(t, "engine-test-entity-1_attr_user_profile", metadata.AttributeID)
	assert.Equal(t, "user_profile", metadata.AttributeName)
	assert.Equal(t, storageinference.TabularData, metadata.StorageType)
	assert.Equal(t, "tables/engine-test-entity-1_attr_user_profile", metadata.StoragePath)
	assert.NotZero(t, metadata.Created)
	assert.NotZero(t, metadata.Updated)
	assert.NotNil(t, metadata.Schema)
}
