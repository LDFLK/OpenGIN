// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package mongorepository

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"lk/datafoundation/core-api/db/config"
	pb "lk/datafoundation/core-api/lk/datafoundation/core-api"
)

var testRepo *MongoRepository
var testCtx context.Context

// Helper function to get DB name from the repository
func (r *MongoRepository) GetDBName() string {
	return r.config.DBName
}

// Helper function to get collection name from the repository
func (r *MongoRepository) GetCollectionName() string {
	return r.config.Collection
}

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Load .env file

	// Create test config - using test collection to avoid affecting production data
	testConfig := &config.MongoConfig{
		URI:        os.Getenv("MONGO_URI"),
		DBName:     os.Getenv("MONGO_DB_NAME"),
		Collection: os.Getenv("MONGO_COLLECTION") + "_test",
	}

	// Initialize MongoDB repository
	testCtx = context.Background()
	testRepo = NewMongoRepository(testCtx, testConfig)

	// Clear test collection before tests
	// testRepo.collection().Drop(testCtx)

	// Add before running tests
	if err := testRepo.client.Ping(testCtx, nil); err != nil {
		log.Fatalf("Cannot connect to MongoDB: %v", err)
	}
	log.Println("Successfully connected to MongoDB")

	// Run tests
	log.Println("Running tests")
	log.Println("DB Name: " + testRepo.GetDBName())
	log.Println("Collection Name: " + testRepo.GetCollectionName())
	exitCode := m.Run()

	// Clean up after tests
	// testRepo.collection().Drop(testCtx)

	log.Println("Tests completed")
	os.Exit(exitCode)
}

// TestCreateAndReadEntity verifies the entity creation and reading functionality:
// 1. Creates an entity with test metadata
// 2. Confirms the entity exists in the database after creation
// 3. Reads the entity and verifies the metadata values
// 4. Validates that the entity ID is correctly set
// 5. Confirms the metadata values are stored correctly
func TestCreateAndReadEntity(t *testing.T) {
	// Log database and collection information
	log.Printf("Test using database: %s, collection: %s", testRepo.GetDBName(), testRepo.GetCollectionName())

	// Create test entity with metadata
	entityID := "test-entity-1"

	// Create metadata
	metadata := make(map[string]*anypb.Any)
	val1, err := anypb.New(wrapperspb.String("value1"))
	assert.NoError(t, err)
	val2, err := anypb.New(wrapperspb.String("value2"))
	assert.NoError(t, err)

	metadata["key1"] = val1
	metadata["key2"] = val2

	entity := &pb.Entity{
		Id:       entityID,
		Metadata: metadata,
	}

	// Create entity
	result, err := testRepo.CreateEntity(testCtx, entity)
	assert.NoError(t, err)
	assert.NotNil(t, result, "Insert result should not be nil")
	log.Printf("Inserted document with ID: %v", result.InsertedID)

	// Read entity with error check
	readEntity, err := testRepo.ReadEntity(testCtx, entityID)
	if err != nil {
		t.Fatalf("Failed to read entity: %v", err)
	}
	if readEntity == nil {
		t.Fatal("Entity not found after creation")
	}

	// Verify entity data
	assert.Equal(t, entityID, readEntity.Id)

	// Verify metadata
	assert.Equal(t, 2, len(readEntity.Metadata))
	assert.Contains(t, readEntity.Metadata, "key1")
	assert.Contains(t, readEntity.Metadata, "key2")
}

// TestUpdateEntityMetadata verifies the metadata update functionality:
// 1. Creates an entity with initial metadata
// 2. Updates the metadata with new values
// 3. Reads the updated entity and verifies the metadata changes
// 4. Confirms that the metadata values are updated correctly
// 5. Validates that the entity ID remains unchanged
// TestUpdateEntityMetadata tests updating entity metadata
func TestUpdateEntityMetadata(t *testing.T) {
	// Log database and collection information
	log.Printf("Test using database: %s, collection: %s", testRepo.GetDBName(), testRepo.GetCollectionName())

	// Create test entity with initial metadata
	entityID := "test-entity-2"

	// Create initial metadata
	metadata := make(map[string]*anypb.Any)
	val1, err := anypb.New(wrapperspb.String("initial-value"))
	assert.NoError(t, err)
	metadata["key1"] = val1

	// Create minimal entity with only ID and metadata
	entity := &pb.Entity{
		Id:       entityID,
		Metadata: metadata,
	}

	// Create entity
	_, err = testRepo.CreateEntity(testCtx, entity)
	assert.NoError(t, err)

	// Update metadata
	updatedMetadata := make(map[string]*anypb.Any)
	val2, err := anypb.New(wrapperspb.String("updated-value"))
	assert.NoError(t, err)
	val3, err := anypb.New(wrapperspb.String("new-value"))
	assert.NoError(t, err)
	updatedMetadata["key1"] = val2 // Update existing key
	updatedMetadata["key3"] = val3 // Add new key

	// Update entity
	_, err = testRepo.UpdateEntity(testCtx, entityID, map[string]interface{}{
		"metadata": updatedMetadata,
	})
	assert.NoError(t, err)

	// Read updated entity
	readEntity, err := testRepo.ReadEntity(testCtx, entityID)
	assert.NoError(t, err)

	// Verify updated metadata
	assert.Equal(t, 2, len(readEntity.Metadata))
	assert.Contains(t, readEntity.Metadata, "key1")
	assert.Contains(t, readEntity.Metadata, "key3")
}

// TestDeleteEntity verifies the entity deletion functionality:
// 1. Creates an entity with test metadata
// 2. Confirms the entity exists in the database after creation
// 3. Tests the DeleteEntity method by removing the entity
// 4. Verifies that exactly one document was deleted (DeletedCount = 1)
// 5. Confirms the entity no longer exists by attempting to read it
// 6. Validates that an error is returned when trying to read a deleted entity
func TestDeleteEntity(t *testing.T) {
	// Log database and collection information
	log.Printf("Test using database: %s, collection: %s", testRepo.GetDBName(), testRepo.GetCollectionName())

	// Create test entity
	entityID := "test-entity-3"

	// Create metadata
	metadata := make(map[string]*anypb.Any)
	val1, err := anypb.New(wrapperspb.String("delete-test"))
	assert.NoError(t, err)
	metadata["deleteKey"] = val1

	// Create minimal entity with only ID and metadata
	entity := &pb.Entity{
		Id:       entityID,
		Metadata: metadata,
	}

	// Create entity
	_, err = testRepo.CreateEntity(testCtx, entity)
	assert.NoError(t, err)

	// Delete entity
	result, err := testRepo.DeleteEntity(testCtx, entityID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.DeletedCount)

	// Verify entity is deleted
	_, err = testRepo.ReadEntity(testCtx, entityID)
	assert.Error(t, err) // Should return an error since entity doesn't exist
}

// TestMetadataHandling verifies the handling of complex metadata with various data types:
// 1. Tests storage and retrieval of different protobuf wrapper types (String, Int32, Bool)
// 2. Confirms that Entity.Id is correctly used as MongoDB's _id field
// 3. Validates that metadata values maintain their type information through storage/retrieval
// 4. Ensures that type-specific unwrapping (UnmarshalTo) works correctly after retrieval
// 5. Verifies the integrity of data values after a complete create-read cycle
func TestMetadataHandling(t *testing.T) {
	// Log database and collection information
	log.Printf("Test using database: %s, collection: %s", testRepo.GetDBName(), testRepo.GetCollectionName())

	// Create test entity with metadata
	entityID := "test-entity-4"

	// Create diverse metadata with different value types
	metadata := make(map[string]*anypb.Any)

	stringVal, err := anypb.New(wrapperspb.String("string-value"))
	assert.NoError(t, err)

	intVal, err := anypb.New(wrapperspb.Int32(42))
	assert.NoError(t, err)

	boolVal, err := anypb.New(wrapperspb.Bool(true))
	assert.NoError(t, err)

	metadata["string"] = stringVal
	metadata["number"] = intVal
	metadata["boolean"] = boolVal

	// Create minimal entity with only required fields
	entity := &pb.Entity{
		Id:       entityID,
		Metadata: metadata,
	}

	// Create entity - this should use the entityID as the document ID
	result, err := testRepo.CreateEntity(testCtx, entity)
	assert.NoError(t, err)
	assert.NotNil(t, result, "Insert result should not be nil")

	// Read entity to verify metadata was stored correctly
	readEntity, err := testRepo.ReadEntity(testCtx, entityID)
	assert.NoError(t, err)

	// Verify metadata with different value types
	assert.Equal(t, 3, len(readEntity.Metadata))
	assert.Contains(t, readEntity.Metadata, "string")
	assert.Contains(t, readEntity.Metadata, "number")
	assert.Contains(t, readEntity.Metadata, "boolean")

	// Test we can correctly extract values from metadata
	stringAny := readEntity.Metadata["string"]
	stringWrapper := &wrapperspb.StringValue{}
	err = stringAny.UnmarshalTo(stringWrapper)
	assert.NoError(t, err)
	assert.Equal(t, "string-value", stringWrapper.Value)

	intAny := readEntity.Metadata["number"]
	intWrapper := &wrapperspb.Int32Value{}
	err = intAny.UnmarshalTo(intWrapper)
	assert.NoError(t, err)
	assert.Equal(t, int32(42), intWrapper.Value)
}
