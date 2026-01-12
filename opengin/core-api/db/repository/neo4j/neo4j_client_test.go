package neo4jrepository

import (
	"context"
	"log"
	"os"
	"testing"

	"lk/datafoundation/core-api/db/config"
	pb "lk/datafoundation/core-api/lk/datafoundation/core-api"

	"github.com/stretchr/testify/assert"
)

var repository *Neo4jRepository

// cleanupDatabase deletes all nodes and relationships in the database
// Note this functioncannot be called in the test functions as due to Go running tests in parallel it can cause issues
// SEE: https://github.com/LDFLK/nexoan/issues/348
func cleanupDatabase(ctx context.Context, repo *Neo4jRepository) error {
	session := repo.getSession(ctx)
	defer session.Close(ctx)

	// Delete all relationships first
	_, err := session.Run(ctx, "MATCH ()-[r]-() DELETE r", nil)
	if err != nil {
		return err
	}

	// Then delete all nodes
	_, err = session.Run(ctx, "MATCH (n) DELETE n", nil)
	if err != nil {
		return err
	}

	return nil
}

// TestMain initializes the Neo4jRepository before running the tests and closes it afterward.
func TestMain(m *testing.M) {
	// Setup: Initialize the Neo4j repository with the config
	ctx := context.Background()
	cfg := &config.Neo4jConfig{
		URI:      os.Getenv("NEO4J_URI"),
		Username: os.Getenv("NEO4J_USER"),
		Password: os.Getenv("NEO4J_PASSWORD"),
	}
	log.Printf("Connecting to Neo4j at %s", cfg.URI)
	var err error
	repository, err = NewNeo4jRepository(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create Neo4j repository: %v", err)
	}

	// Clean up the database before running tests
	if err := cleanupDatabase(ctx, repository); err != nil {
		log.Printf("Warning: Failed to clean up database before tests: %v", err)
	}

	// Run the tests
	code := m.Run()

	// Teardown: Close the repository after tests
	repository.Close(ctx)

	// Exit with the test result code
	os.Exit(code)
}

// TestCreateEntity tests the CreateGraphEntity method of the Neo4jRepository
func TestCreateEntity(t *testing.T) {
	ctx := context.Background()
	// defer cleanupDatabase(ctx, repository)

	// Prepare the kind parameter
	kind := &pb.Kind{
		Major: "Person",
		Minor: "Minister",
	}

	// Prepare the entity data as a map
	entity := map[string]interface{}{
		"Id":         "1",
		"Name":       "John Doe",
		"Created":    "2025-03-18T00:00:00Z",
		"Terminated": nil,
	}

	// Call the CreateGraphEntity method and capture the returned entity
	createdEntity, err := repository.CreateGraphEntity(ctx, kind, entity)
	log.Printf("Created entity: %v", createdEntity)

	// Verify that no error occurred during creation
	assert.Nil(t, err, "Expected no error when creating an entity")

	// Verify that the returned entity has the correct values
	assert.Equal(t, "1", createdEntity["Id"], "Expected entity to have the correct Id")
	assert.Equal(t, "John Doe", createdEntity["Name"], "Expected entity to have the correct Name")
	assert.Equal(t, "2025-03-18T00:00:00Z", createdEntity["Created"], "Expected entity to have the correct Created date")
	assert.Equal(t, "Minister", createdEntity["MinorKind"], "Expected entity to have the correct MinorKind")
	assert.Nil(t, createdEntity["Terminated"], "Expected entity to have no Terminated field")
}

// TestCreateRelationship tests the CreateRelationship method of the Neo4jRepository
func TestCreateRelationship(t *testing.T) {
	// Prepare the context
	ctx := context.Background()

	kind := &pb.Kind{
		Major: "Person",
		Minor: "Minister",
	}

	// Create two entities first
	entity1 := map[string]interface{}{
		"Id":      "2",
		"Name":    "Alice",
		"Created": "2025-03-18",
	}
	entity2 := map[string]interface{}{
		"Id":      "3",
		"Name":    "Bob",
		"Created": "2025-03-18",
	}

	// Create entities
	_, err := repository.CreateGraphEntity(ctx, kind, entity1)
	assert.Nil(t, err, "Expected no error when creating first entity")

	_, err = repository.CreateGraphEntity(ctx, kind, entity2)
	assert.Nil(t, err, "Expected no error when creating second entity")

	// Prepare relationship data
	relationship := &pb.Relationship{
		Id:              "101",
		RelatedEntityId: "3",
		Name:            "KNOWS",
		StartTime:       "2025-03-18",
	}

	// Create the relationship
	createdRelationship, err := repository.CreateRelationship(ctx, "2", relationship)
	assert.Nil(t, err, "Expected no error when creating the relationship")
	log.Printf("Created relationship: %v", createdRelationship)

	// Verify that the returned relationship has the correct values
	assert.Equal(t, "101", createdRelationship["Id"], "Expected relationship to have the correct Id")
	assert.Equal(t, "2025-03-18T00:00:00Z", createdRelationship["Created"], "Expected relationship to have the correct Created date")
	assert.Equal(t, "KNOWS", createdRelationship["relationshipType"], "Expected relationship to have the correct type")
}

// TestReadEntity tests the ReadGraphEntity method of the Neo4jRepository
func TestReadEntity(t *testing.T) {
	ctx := context.Background()

	kind := &pb.Kind{
		Major: "Person",
		Minor: "Minister",
	}

	// Create an entity for testing
	entity := map[string]interface{}{
		"Id":      "6",
		"Name":    "Charlie",
		"Created": "2025-03-18T00:00:00Z",
	}

	// Create the entity
	createdEntity, err := repository.CreateGraphEntity(ctx, kind, entity)
	assert.Nil(t, err, "Expected no error when creating the entity")
	assert.Equal(t, entity["Id"], createdEntity["Id"], "Expected created entity to have the correct Id")
	assert.Equal(t, entity["Name"], createdEntity["Name"], "Expected created entity to have the correct Name")
	assert.Equal(t, "2025-03-18T00:00:00Z", createdEntity["Created"], "Expected created entity to have the correct Created date")

	// Read the entity by Id
	readEntity, err := repository.ReadGraphEntity(ctx, "6")
	assert.Nil(t, err, "Expected no error when reading the entity")

	// Verify the content of the entity
	assert.Equal(t, entity["Id"], readEntity["Id"], "Expected entity to have the correct Id")
	assert.Equal(t, entity["Name"], readEntity["Name"], "Expected entity to have the correct Name")
	assert.Equal(t, "2025-03-18T00:00:00Z", readEntity["Created"], "Expected entity to have the correct Created date")
}

// TestReadRelatedEntityIds tests the ReadRelatedGraphEntityIds method of the Neo4jRepository
func TestReadRelatedEntityIds(t *testing.T) {
	ctx := context.Background()

	kind := &pb.Kind{
		Major: "Person",
		Minor: "Minister",
	}

	// Create two entities
	entity1 := map[string]interface{}{
		"Id":      "4",
		"Name":    "Alice",
		"Created": "2025-03-18T00:00:00Z",
	}
	entity2 := map[string]interface{}{
		"Id":      "5",
		"Name":    "Bob",
		"Created": "2025-03-18T00:00:00Z",
	}

	// Create entities
	_, err := repository.CreateGraphEntity(ctx, kind, entity1)
	assert.Nil(t, err, "Expected no error when creating the first entity")

	_, err = repository.CreateGraphEntity(ctx, kind, entity2)
	assert.Nil(t, err, "Expected no error when creating the second entity")

	// Create a relationship between the entities
	relationship := &pb.Relationship{
		Id:              "102",
		Name:            "KNOWS",
		RelatedEntityId: "5",
		StartTime:       "2025-03-18T00:00:00Z",
		EndTime:         "2025-12-31T00:00:00Z",
	}

	_, err = repository.CreateRelationship(ctx, "4", relationship)
	assert.Nil(t, err, "Expected no error when creating the relationship")

	// Step 3: Prepare the test data for fetching related relationships
	entityID := "4"              // ID of the entity to get related relationships for
	relationshipType := "KNOWS"  // Relationship type
	ts := "2025-03-18T00:00:00Z" // Timestamp (YYYY-MM-DD)

	// Step 4: Call the function to fetch related relationships
	relatedRelationships, err := repository.ReadRelatedGraphEntityIds(ctx, entityID, relationshipType, ts)
	assert.Nil(t, err, "Expected no error when getting related relationships")
	assert.NotNil(t, relatedRelationships, "Expected related relationships to be returned")

	// Step 5: Verify the response
	assert.Equal(t, 1, len(relatedRelationships), "Expected exactly one related relationship")
	relationshipData := relatedRelationships[0]

	// Verify the structure and content of the relationship
	assert.Equal(t, "102", relationshipData["Id"], "Expected relationship ID to match")
	assert.Equal(t, "KNOWS", relationshipData["Name"], "Expected relationship Name to match")
	assert.Equal(t, "5", relationshipData["RelatedEntityId"], "Expected RelatedEntityId to match")
	assert.Equal(t, "2025-03-18T00:00:00Z", relationshipData["StartTime"], "Expected StartTime to match")
	assert.Equal(t, "2025-12-31T00:00:00Z", relationshipData["EndTime"], "Expected EndTime to match")
}

func TestReadRelationships(t *testing.T) {
	ctx := context.Background()

	kind := &pb.Kind{
		Major: "Person",
		Minor: "Minister",
	}

	// Create two entities
	entityMap1 := map[string]interface{}{
		"Id":      "7",
		"Name":    "David",
		"Created": "2025-03-18",
	}
	entityMap2 := map[string]interface{}{
		"Id":      "8",
		"Name":    "Eve",
		"Created": "2025-03-18",
	}

	// Create entities in the repository
	_, err := repository.CreateGraphEntity(ctx, kind, entityMap1)
	assert.Nil(t, err, "Expected no error when creating first entity")

	_, err = repository.CreateGraphEntity(ctx, kind, entityMap2)
	assert.Nil(t, err, "Expected no error when creating second entity")

	// Create a relationship between the entities
	relationship := &pb.Relationship{
		Id:              "103",
		RelatedEntityId: "8", // ID of the related entity
		Name:            "KNOWS",
		StartTime:       "2025-03-18",
	}
	_, err = repository.CreateRelationship(ctx, "7", relationship)
	assert.Nil(t, err, "Expected no error when creating the relationship")

	// Fetch relationships for entity 7
	relationships, err := repository.ReadRelationships(ctx, "7")
	assert.Nil(t, err, "Expected no error when fetching relationships")
	log.Printf("Relationships for entity 7: %v", relationships)

	// Verify that the relationship exists
	relationshipFound := false
	for _, relationship := range relationships {
		if relationship["relatedID"] == "8" {
			relationshipFound = true
			break
		}
	}

	// Assert that the relationship to the entity 8 exists
	assert.True(t, relationshipFound, "Expected relationship to include the correct related entity ID")
}

func TestReadRelationship(t *testing.T) {
	ctx := context.Background()

	kind := &pb.Kind{
		Major: "Person",
		Minor: "Minister",
	}

	// Create two entities
	entityMap1 := map[string]interface{}{
		"Id":      "9",
		"Name":    "David",
		"Created": "2025-03-18",
	}
	entityMap2 := map[string]interface{}{
		"Id":      "10",
		"Name":    "Eve",
		"Created": "2025-03-18",
	}

	// Create entities in the repository
	_, err := repository.CreateGraphEntity(ctx, kind, entityMap1)
	assert.Nil(t, err, "Expected no error when creating first entity")

	_, err = repository.CreateGraphEntity(ctx, kind, entityMap2)
	assert.Nil(t, err, "Expected no error when creating second entity")

	// Create a relationship between the entities
	relationship := &pb.Relationship{
		Id:              "rel_9_10",
		RelatedEntityId: "10", // ID of the related entity (Eve)
		Name:            "KNOWS",
		StartTime:       "2025-03-18",
	}
	_, err = repository.CreateRelationship(ctx, "9", relationship) // "9" is David's ID
	assert.Nil(t, err, "Expected no error when creating the relationship")

	// Fetch the relationship by ID
	relationshipMap, err := repository.ReadRelationship(ctx, "rel_9_10")
	assert.Nil(t, err, "Expected no error when fetching the relationship")
	log.Printf("Fetched relationship: %v", relationshipMap)

	// Verify that the relationship data is correct
	assert.Equal(t, "KNOWS", relationshipMap["type"], "Expected relationship type to be KNOWS")
	assert.Equal(t, "9", relationshipMap["startEntityID"], "Expected start entity ID to be 9 (David's ID)")
	assert.Equal(t, "10", relationshipMap["endEntityID"], "Expected end entity ID to be 10 (Eve's ID)")
	assert.Equal(t, "2025-03-18T00:00:00Z", relationshipMap["Created"], "Expected start date to be 2025-03-18T00:00:00Z")

	// Optional: Assert the endDate is nil (since it wasn't set in the creation)
	assert.Nil(t, relationshipMap["Terminated"], "Expected end date to be nil")
}

func TestUpdateEntity(t *testing.T) {
	ctx := context.Background()

	kind := &pb.Kind{
		Major: "Person",
		Minor: "Minister",
	}

	// Create a test entity
	entityData := map[string]interface{}{
		"Id":      "11",
		"Name":    "Mary",
		"Created": "2025-03-18",
	}
	_, err := repository.CreateGraphEntity(ctx, kind, entityData)
	assert.Nil(t, err, "Expected no error when creating entity")

	// Update the entity
	updateData := map[string]interface{}{
		"Name":       "Mary Updated",
		"Terminated": "2025-12-31T00:00:00Z",
	}

	updatedEntity, err := repository.UpdateGraphEntity(ctx, "11", updateData)
	log.Printf("Updated entity: %v", updatedEntity)
	assert.Nil(t, err, "Expected no error when updating entity")
	assert.NotNil(t, updatedEntity, "Expected updated entity to be returned")

	// Verify that the entity was updated correctly in the return value
	assert.Equal(t, "Mary Updated", updatedEntity["Name"], "Expected updated name")
	assert.Equal(t, "2025-12-31T00:00:00Z", updatedEntity["Terminated"], "Expected updated dateEnded")

	// Fetch the entity from the database and verify
	entity, err := repository.ReadGraphEntity(ctx, "11")
	log.Printf("Fetched entity: %v", entity)
	assert.Nil(t, err, "Expected no error when reading updated entity")
	assert.Equal(t, "Mary Updated", entity["Name"], "Expected database to have updated name")
	assert.Equal(t, "2025-12-31T00:00:00Z", entity["Terminated"], "Expected database to have updated dateEnded")
}

func TestUpdateRelationship(t *testing.T) {
	ctx := context.Background()

	kind := &pb.Kind{
		Major: "Person",
		Minor: "Minister",
	}

	// Create two entities first
	entity1 := map[string]interface{}{
		"Id":      "update_rel_entity_1",
		"Name":    "Alice",
		"Created": "2025-03-18",
	}
	entity2 := map[string]interface{}{
		"Id":      "update_rel_entity_2",
		"Name":    "Bob",
		"Created": "2025-03-18",
	}

	_, err := repository.CreateGraphEntity(ctx, kind, entity1)
	assert.Nil(t, err, "Expected no error when creating first entity")

	_, err = repository.CreateGraphEntity(ctx, kind, entity2)
	assert.Nil(t, err, "Expected no error when creating second entity")

	// Create a relationship
	relationship := &pb.Relationship{
		Id:              "update_rel_test",
		RelatedEntityId: "update_rel_entity_2",
		Name:            "KNOWS",
		StartTime:       "2025-03-18",
	}

	_, err = repository.CreateRelationship(ctx, "update_rel_entity_1", relationship)
	assert.Nil(t, err, "Expected no error when creating the relationship")

	// Update the relationship
	updateData := map[string]interface{}{
		"Terminated": "2025-12-31T00:00:00Z",
	}

	// Call the function to update the relationship
	updatedRelationship, err := repository.UpdateRelationship(ctx, "update_rel_test", updateData)
	log.Printf("Updated relationship: %v", updatedRelationship)
	assert.Nil(t, err, "Expected no error when updating relationship")
	assert.NotNil(t, updatedRelationship, "Expected updated relationship to be returned")

	// Verify that the relationship was updated correctly in the return value
	assert.Equal(t, "2025-12-31T00:00:00Z", updatedRelationship["Terminated"], "Expected updated endDate")

	// Fetch the relationship from the database using getRelationship
	fetchedRelationship, err := repository.ReadRelationship(ctx, "update_rel_test")
	log.Printf("Fetched relationship: %v", fetchedRelationship)
	assert.Nil(t, err, "Expected no error when reading updated relationship")

	// Check if the relationship has the updated endDate
	assert.Equal(t, "2025-12-31T00:00:00Z", fetchedRelationship["Terminated"], "Expected relationship to have updated endDate")
}

func TestDeleteRelationship(t *testing.T) {
	ctx := context.Background()

	kind := &pb.Kind{
		Major: "Person",
		Minor: "Minister",
	}

	// Create two entities first
	entity1 := map[string]interface{}{
		"Id":      "delete_rel_entity_1",
		"Name":    "Alice",
		"Created": "2025-03-18",
	}
	entity2 := map[string]interface{}{
		"Id":      "delete_rel_entity_2",
		"Name":    "Bob",
		"Created": "2025-03-18",
	}

	_, err := repository.CreateGraphEntity(ctx, kind, entity1)
	assert.Nil(t, err, "Expected no error when creating first entity")

	_, err = repository.CreateGraphEntity(ctx, kind, entity2)
	assert.Nil(t, err, "Expected no error when creating second entity")

	// Create a relationship
	relationship := &pb.Relationship{
		Id:              "delete_rel_test",
		RelatedEntityId: "delete_rel_entity_2",
		Name:            "KNOWS",
		StartTime:       "2025-03-18",
	}

	_, err = repository.CreateRelationship(ctx, "delete_rel_entity_1", relationship)
	assert.Nil(t, err, "Expected no error when creating the relationship")

	// Call the function to delete the relationship
	err = repository.DeleteRelationship(ctx, "delete_rel_test")
	assert.Nil(t, err, "Expected no error when deleting relationship")

	// Fetch the relationship to ensure it was deleted
	deletedRelationship, err := repository.ReadRelationship(ctx, "delete_rel_test")
	assert.NotNil(t, err, "Expected error when fetching deleted relationship")
	assert.Contains(t, err.Error(), "not found", "Expected error message to indicate relationship not found")
	assert.Nil(t, deletedRelationship, "Expected relationship to be nil after deletion")
}

func TestDeleteEntity(t *testing.T) {
	ctx := context.Background()

	kind := &pb.Kind{
		Major: "Person",
		Minor: "Minister",
	}

	// Create a test entity that can be deleted
	entity1 := map[string]interface{}{
		"Id":      "delete_entity_1",
		"Name":    "John Smith",
		"Created": "2025-03-18",
	}
	_, err := repository.CreateGraphEntity(ctx, kind, entity1)
	assert.Nil(t, err, "Expected no error when creating entity 1")

	// Delete the entity
	err = repository.DeleteGraphEntity(ctx, "delete_entity_1")
	assert.Nil(t, err, "Expected no error when deleting entity")

	// Verify the entity was deleted
	_, err = repository.ReadGraphEntity(ctx, "delete_entity_1")
	assert.NotNil(t, err, "Expected error when fetching deleted entity")
	assert.Contains(t, err.Error(), "not found", "Expected error message to indicate entity not found")

	// Test deleting an entity with relationships (should fail)
	// Create two entities
	entity2 := map[string]interface{}{
		"Id":      "delete_entity_2",
		"Name":    "Alice",
		"Created": "2025-03-18",
	}
	entity3 := map[string]interface{}{
		"Id":      "delete_entity_3",
		"Name":    "Bob",
		"Created": "2025-03-18",
	}

	_, err = repository.CreateGraphEntity(ctx, kind, entity2)
	assert.Nil(t, err, "Expected no error when creating entity 2")

	_, err = repository.CreateGraphEntity(ctx, kind, entity3)
	assert.Nil(t, err, "Expected no error when creating entity 3")

	// Create a relationship between them
	relationship := &pb.Relationship{
		Id:              "delete_entity_rel",
		RelatedEntityId: "delete_entity_3",
		Name:            "KNOWS",
		StartTime:       "2025-03-18",
	}
	_, err = repository.CreateRelationship(ctx, "delete_entity_2", relationship)
	assert.Nil(t, err, "Expected no error when creating relationship")

	// Try to delete entity with relationships (should fail)
	err = repository.DeleteGraphEntity(ctx, "delete_entity_2")
	assert.NotNil(t, err, "Expected error when deleting entity with relationships")
	assert.Contains(t, err.Error(), "entity has relationships and cannot be deleted", "Expected error message to indicate relationships prevent deletion")
}

func TestAddMinistriesAndDepartments(t *testing.T) {
	ctx := context.Background()

	// Define ministries and their departments
	ministries := []struct {
		id          string
		name        string
		departments []struct {
			id   string
			name string
		}
	}{
		{
			id:   "ministry1",
			name: "Ministry of Education",
			departments: []struct {
				id   string
				name string
			}{
				{id: "dept1", name: "Department of Schools"},
				{id: "dept2", name: "Department of Higher Education"},
				{id: "dept3", name: "Department of Research"},
			},
		},
		{
			id:   "ministry2",
			name: "Ministry of Health",
			departments: []struct {
				id   string
				name string
			}{
				{id: "dept4", name: "Department of Hospitals"},
				{id: "dept5", name: "Department of Public Health"},
				{id: "dept6", name: "Department of Medical Research"},
			},
		},
		{
			id:   "ministry3",
			name: "Ministry of Finance",
			departments: []struct {
				id   string
				name string
			}{
				{id: "dept7", name: "Department of Budget"},
				{id: "dept8", name: "Department of Taxation"},
				{id: "dept9", name: "Department of Audits"},
			},
		},
	}

	// Start time for the relationships
	startTime := "2022-07-22"

	kindMinistry := &pb.Kind{
		Major: "Organization",
		Minor: "Ministry",
	}

	kindDept := &pb.Kind{
		Major: "Organization",
		Minor: "Department",
	}

	// Create ministries and departments, and establish relationships
	for _, ministry := range ministries {

		// Create the ministry
		ministryEntity := map[string]interface{}{
			"Id":      ministry.id,
			"Name":    ministry.name,
			"Created": "2022-07-22",
		}

		_, err := repository.CreateGraphEntity(ctx, kindMinistry, ministryEntity)
		assert.Nil(t, err, "Failed to create ministry: %s", ministry.name)

		// Create departments and relationships
		for _, department := range ministry.departments {
			// Create the department
			departmentEntity := map[string]interface{}{
				"Id":      department.id,
				"Name":    department.name,
				"Created": "2022-07-22",
			}

			_, err := repository.CreateGraphEntity(ctx, kindDept, departmentEntity)
			assert.Nil(t, err, "Failed to create department: %s", department.name)

			// Establish the is_department relationship
			relationship := &pb.Relationship{
				Id:              ministry.id + "_to_" + department.id,
				Name:            "is_department",
				RelatedEntityId: department.id,
				StartTime:       startTime,
			}

			_, err = repository.CreateRelationship(ctx, ministry.id, relationship)
			assert.Nil(t, err, "Failed to create relationship between ministry %s and department %s", ministry.name, department.name)
		}
	}
}

func TestCreateRelationshipWithDuplicateId(t *testing.T) {
	ctx := context.Background()

	kind := &pb.Kind{
		Major: "Person",
		Minor: "Employee",
	}

	// Create two entities
	entity1 := map[string]interface{}{
		"Id":      "dup_entity_1",
		"Name":    "DuplicateTestEntity1",
		"Created": "2025-04-01T00:00:00Z",
	}
	entity2 := map[string]interface{}{
		"Id":      "dup_entity_2",
		"Name":    "DuplicateTestEntity2",
		"Created": "2025-04-01T00:00:00Z",
	}
	_, err := repository.CreateGraphEntity(ctx, kind, entity1)
	assert.Nil(t, err, "Expected no error when creating entity 1")
	_, err = repository.CreateGraphEntity(ctx, kind, entity2)
	assert.Nil(t, err, "Expected no error when creating entity 2")

	// Create a relationship with a specific ID
	relationship := &pb.Relationship{
		Id:              "duplicate_rel_id",
		Name:            "WORKS_WITH",
		RelatedEntityId: "dup_entity_2",
		StartTime:       "2025-04-01T00:00:00Z",
	}
	createdRel, err := repository.CreateRelationship(ctx, "dup_entity_1", relationship)
	assert.Nil(t, err, "Expected no error when creating the first relationship")
	assert.NotNil(t, createdRel, "Expected relationship to be created")
	log.Printf("First relationship created: %v", createdRel)

	// Read the original relationship to verify its properties
	originalRel, err := repository.ReadRelationship(ctx, "duplicate_rel_id")
	assert.Nil(t, err, "Expected no error reading original relationship")
	originalCreated := originalRel["Created"]
	log.Printf("Original relationship: %v", originalRel)

	// Attempt to create another relationship with the SAME ID (should fail)
	duplicateRelationship := &pb.Relationship{
		Id:              "duplicate_rel_id", // Same ID as above
		Name:            "MANAGES",          // Different name
		RelatedEntityId: "dup_entity_2",
		StartTime:       "2025-05-01T00:00:00Z", // Different start time
	}
	createdDup, err := repository.CreateRelationship(ctx, "dup_entity_1", duplicateRelationship)

	// Verify that the creation failed
	assert.NotNil(t, err, "Expected error when creating relationship with duplicate ID")
	assert.Contains(t, err.Error(), "already exists", "Expected error message to indicate relationship already exists")
	assert.Nil(t, createdDup, "Expected no relationship to be returned when creation fails")
	log.Printf("Duplicate creation error (expected): %v", err)

	// Verify the original relationship was NOT modified
	verifyRel, err := repository.ReadRelationship(ctx, "duplicate_rel_id")
	assert.Nil(t, err, "Expected no error reading relationship after duplicate attempt")
	assert.Equal(t, "WORKS_WITH", verifyRel["type"], "Expected relationship type to remain unchanged")
	assert.Equal(t, originalCreated, verifyRel["Created"], "Expected relationship Created date to remain unchanged")
	log.Printf("Verified relationship unchanged: %v", verifyRel)

	// Verify only one relationship exists between the entities
	relationships, err := repository.ReadRelationships(ctx, "dup_entity_1")
	assert.Nil(t, err, "Expected no error reading relationships")

	// Count relationships with our test ID
	count := 0
	for _, rel := range relationships {
		if rel["relationshipID"] == "duplicate_rel_id" {
			count++
		}
	}
	assert.Equal(t, 1, count, "Expected exactly one relationship with the duplicate ID")
}

func TestUpdateRelationshipWithNonExistentId(t *testing.T) {
	ctx := context.Background()

	// Attempt to update a relationship that doesn't exist
	updateData := map[string]interface{}{
		"Terminated": "2025-12-31T00:00:00Z",
	}

	updatedRel, err := repository.UpdateRelationship(ctx, "non_existent_rel_id", updateData)

	// Verify that the update failed
	assert.NotNil(t, err, "Expected error when updating non-existent relationship")
	assert.Contains(t, err.Error(), "does not exist", "Expected error message to indicate relationship does not exist")
	assert.Nil(t, updatedRel, "Expected no relationship to be returned when update fails")
	log.Printf("Update non-existent relationship error (expected): %v", err)
}

func TestUpdateRelationshipFields(t *testing.T) {
	ctx := context.Background()

	kind := &pb.Kind{
		Major: "Person",
		Minor: "Employee",
	}

	// Create two entities
	entity1 := map[string]interface{}{
		"Id":      "update_test_entity_1",
		"Name":    "UpdateTestEntity1",
		"Created": "2025-04-01T00:00:00Z",
	}
	entity2 := map[string]interface{}{
		"Id":      "update_test_entity_2",
		"Name":    "UpdateTestEntity2",
		"Created": "2025-04-01T00:00:00Z",
	}
	_, err := repository.CreateGraphEntity(ctx, kind, entity1)
	assert.Nil(t, err, "Expected no error when creating entity 1")
	_, err = repository.CreateGraphEntity(ctx, kind, entity2)
	assert.Nil(t, err, "Expected no error when creating entity 2")

	// Create a relationship
	relationship := &pb.Relationship{
		Id:              "update_test_rel",
		Name:            "REPORTS_TO",
		RelatedEntityId: "update_test_entity_2",
		StartTime:       "2025-04-01T00:00:00Z",
	}
	createdRel, err := repository.CreateRelationship(ctx, "update_test_entity_1", relationship)
	assert.Nil(t, err, "Expected no error when creating relationship")
	log.Printf("Created relationship for update test: %v", createdRel)

	// Count total relationships before updates
	allRelsBefore, err := repository.ReadRelationships(ctx, "update_test_entity_1")
	assert.Nil(t, err, "Expected no error reading relationships before update")
	totalRelsBefore := len(allRelsBefore)
	log.Printf("Total relationships before update: %d", totalRelsBefore)

	// Test 1: Update only Created field
	updateCreated := map[string]interface{}{
		"Created": "2025-03-15T00:00:00Z",
	}
	updatedRel, err := repository.UpdateRelationship(ctx, "update_test_rel", updateCreated)
	assert.Nil(t, err, "Expected no error when updating Created field")
	assert.NotNil(t, updatedRel, "Expected updated relationship to be returned")
	assert.Equal(t, "2025-03-15T00:00:00Z", updatedRel["Created"], "Expected Created date to be updated")
	log.Printf("After updating Created: %v", updatedRel)

	// Verify no new relationships were created
	allRelsAfterCreated, err := repository.ReadRelationships(ctx, "update_test_entity_1")
	assert.Nil(t, err, "Expected no error reading relationships after Created update")
	assert.Equal(t, totalRelsBefore, len(allRelsAfterCreated), "Expected same number of relationships after Created update")

	// Test 2: Update only Terminated field
	updateTerminated := map[string]interface{}{
		"Terminated": "2025-12-31T00:00:00Z",
	}
	updatedRel, err = repository.UpdateRelationship(ctx, "update_test_rel", updateTerminated)
	assert.Nil(t, err, "Expected no error when updating Terminated field")
	assert.NotNil(t, updatedRel, "Expected updated relationship to be returned")
	assert.Equal(t, "2025-12-31T00:00:00Z", updatedRel["Terminated"], "Expected Terminated date to be updated")
	log.Printf("After updating Terminated: %v", updatedRel)

	// Verify no new relationships were created
	allRelsAfterTerminated, err := repository.ReadRelationships(ctx, "update_test_entity_1")
	assert.Nil(t, err, "Expected no error reading relationships after Terminated update")
	assert.Equal(t, totalRelsBefore, len(allRelsAfterTerminated), "Expected same number of relationships after Terminated update")

	// Test 3: Update both Created and Terminated fields
	updateBoth := map[string]interface{}{
		"Created":    "2025-02-01T00:00:00Z",
		"Terminated": "2025-11-30T00:00:00Z",
	}
	updatedRel, err = repository.UpdateRelationship(ctx, "update_test_rel", updateBoth)
	assert.Nil(t, err, "Expected no error when updating both fields")
	assert.NotNil(t, updatedRel, "Expected updated relationship to be returned")
	assert.Equal(t, "2025-02-01T00:00:00Z", updatedRel["Created"], "Expected Created date to be updated")
	assert.Equal(t, "2025-11-30T00:00:00Z", updatedRel["Terminated"], "Expected Terminated date to be updated")
	log.Printf("After updating both Created and Terminated: %v", updatedRel)

	// Verify no new relationships were created
	allRelsAfterBoth, err := repository.ReadRelationships(ctx, "update_test_entity_1")
	assert.Nil(t, err, "Expected no error reading relationships after both fields update")
	assert.Equal(t, totalRelsBefore, len(allRelsAfterBoth), "Expected same number of relationships after both fields update")

	// Test 4: Try to update Name field (should fail - only Created and Terminated allowed)
	updateName := map[string]interface{}{
		"Name": "UpdatedName",
	}
	updatedRel, err = repository.UpdateRelationship(ctx, "update_test_rel", updateName)
	assert.NotNil(t, err, "Expected error when trying to update Name field")
	assert.Contains(t, err.Error(), "unsupported field", "Expected error message about unsupported field")
	assert.Nil(t, updatedRel, "Expected no relationship to be returned when updating unsupported field")
	log.Printf("Update Name error (expected): %v", err)

	// Test 5: Verify the relationship type (Neo4j type) hasn't changed
	finalRel, err := repository.ReadRelationship(ctx, "update_test_rel")
	assert.Nil(t, err, "Expected no error reading final relationship state")
	assert.Equal(t, "REPORTS_TO", finalRel["type"], "Expected relationship type to remain unchanged")
	assert.Equal(t, "update_test_entity_1", finalRel["startEntityID"], "Expected start entity to remain unchanged")
	assert.Equal(t, "update_test_entity_2", finalRel["endEntityID"], "Expected end entity to remain unchanged")
	log.Printf("Final relationship state: %v", finalRel)

	// Test 6: Try updating with no valid fields (should fail)
	updateEmpty := map[string]interface{}{}
	updatedRel, err = repository.UpdateRelationship(ctx, "update_test_rel", updateEmpty)
	assert.NotNil(t, err, "Expected error when updating with no valid fields")
	assert.Contains(t, err.Error(), "no valid fields", "Expected error message about no valid fields")
	assert.Nil(t, updatedRel, "Expected no relationship to be returned when no fields provided")

	// Test 7: Try to update other unsupported fields (should fail)
	updateUnsupported := map[string]interface{}{
		"RelatedEntityId": "some_other_entity",
	}
	updatedRel, err = repository.UpdateRelationship(ctx, "update_test_rel", updateUnsupported)
	assert.NotNil(t, err, "Expected error when trying to update RelatedEntityId field")
	assert.Contains(t, err.Error(), "unsupported field", "Expected error message about unsupported field")
	assert.Nil(t, updatedRel, "Expected no relationship to be returned when updating unsupported field")
	log.Printf("Update unsupported field error (expected): %v", err)

	// Test 8: Update with both supported fields (Created and Terminated)
	updateAll := map[string]interface{}{
		"Created":    "2025-01-01T00:00:00Z",
		"Terminated": "2025-12-31T23:59:59Z",
	}
	updatedRel, err = repository.UpdateRelationship(ctx, "update_test_rel", updateAll)
	assert.Nil(t, err, "Expected no error when updating both supported fields")
	assert.NotNil(t, updatedRel, "Expected updated relationship to be returned")
	assert.Equal(t, "2025-01-01T00:00:00Z", updatedRel["Created"], "Expected Created date to be updated")
	assert.Equal(t, "2025-12-31T23:59:59Z", updatedRel["Terminated"], "Expected Terminated date to be updated")
	log.Printf("After updating both supported fields: %v", updatedRel)

	// Final verification: no new relationships were created
	allRelsFinal, err := repository.ReadRelationships(ctx, "update_test_entity_1")
	assert.Nil(t, err, "Expected no error reading relationships at end")
	assert.Equal(t, totalRelsBefore, len(allRelsFinal), "Expected same number of relationships at end of all updates")
}

func TestReadFilteredRelationships(t *testing.T) {
	ctx := context.Background()

	kind := &pb.Kind{
		Major: "Person",
		Minor: "Tester",
	}

	// Create two entities
	entityA := map[string]interface{}{
		"Id":      "A",
		"Name":    "EntityA",
		"Created": "2025-04-01T00:00:00Z",
	}
	entityB := map[string]interface{}{
		"Id":      "B",
		"Name":    "EntityB",
		"Created": "2025-04-01T00:00:00Z",
	}
	_, err := repository.CreateGraphEntity(ctx, kind, entityA)
	assert.Nil(t, err, "Expected no error when creating entity A")
	_, err = repository.CreateGraphEntity(ctx, kind, entityB)
	assert.Nil(t, err, "Expected no error when creating entity B")

	// Create relationships
	rel1 := &pb.Relationship{
		Id:              "rel1",
		Name:            "FRIEND",
		RelatedEntityId: "B",
		StartTime:       "2025-04-01T00:00:00Z",
	}
	rel2 := &pb.Relationship{
		Id:              "rel2",
		Name:            "COLLEAGUE",
		RelatedEntityId: "B",
		StartTime:       "2025-04-02T00:00:00Z",
		EndTime:         "2025-05-01T00:00:00Z",
	}
	_, err = repository.CreateRelationship(ctx, "A", rel1)
	assert.Nil(t, err, "Expected no error when creating FRIEND relationship")
	_, err = repository.CreateRelationship(ctx, "A", rel2)
	assert.Nil(t, err, "Expected no error when creating COLLEAGUE relationship")

	// Also create an incoming relationship to A from B
	rel3 := &pb.Relationship{
		Id:              "rel3",
		Name:            "MENTOR",
		RelatedEntityId: "A",
		StartTime:       "2025-04-03T00:00:00Z",
	}
	_, err = repository.CreateRelationship(ctx, "B", rel3)
	assert.Nil(t, err, "Expected no error when creating MENTOR relationship")

	// 1. No filters (should return all relationships for A)
	rels, err := repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{}, "")
	log.Printf("ReadFilteredRelationships response (no filters): %+v", rels)
	assert.Nil(t, err, "Expected no error when reading filtered relationships with no filters")
	assert.Equal(t, 3, len(rels), "Expected exactly 3 relationships for entity A with no filters (2 outgoing + 1 incoming)")

	// // 2. Filter by relationship type (name)
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{"name": "FRIEND"}, "")
	log.Printf("ReadFilteredRelationships response (only name): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by name")
	assert.Equal(t, 1, len(rels), "Expected 1 FRIEND relationship")
	assert.Equal(t, "FRIEND", rels[0]["name"])

	// // 3. Filter by relatedEntityId
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{"relatedEntityId": "B"}, "")
	log.Printf("ReadFilteredRelationships response (only relatedEntityId): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by relatedEntityId")
	assert.Equal(t, 3, len(rels), "Expected exactly 3 relationships where B is the related entity (2 outgoing to B, 1 incoming from B)")

	// // 4. Filter by direction OUTGOING
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{"direction": "OUTGOING"}, "")
	log.Printf("ReadFilteredRelationships response (only direction: OUTGOING): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by direction OUTGOING")
	for _, r := range rels {
		assert.Equal(t, "OUTGOING", r["direction"])
	}

	// // 5. Filter by direction INCOMING
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{"direction": "INCOMING"}, "")
	log.Printf("ReadFilteredRelationships response (only direction: INCOMING): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by direction INCOMING")
	for _, r := range rels {
		assert.Equal(t, "INCOMING", r["direction"])
	}

	// // 6. Filter by startTime (exact match)
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{"startTime": "2025-04-01T00:00:00Z"}, "")
	log.Printf("ReadFilteredRelationships response (only startTime): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by startTime")
	assert.Equal(t, 1, len(rels), "Expected 1 relationship with startTime 2025-04-01T00:00:00Z")
	assert.Equal(t, "2025-04-01T00:00:00Z", rels[0]["startTime"])

	// // 7. Filter by endTime (exact match)
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{"endTime": "2025-05-01T00:00:00Z"}, "")
	log.Printf("ReadFilteredRelationships response (only endTime): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by endTime")
	assert.Equal(t, 1, len(rels), "Expected 1 relationship with endTime 2025-05-01T00:00:00Z")
	assert.Equal(t, "2025-05-01T00:00:00Z", rels[0]["endTime"])

	// 	// 8. Filter by activeAt (should match rel1 and rel2 if activeAt is between their start/end)
	activeAt := "2025-05-03T00:00:00Z"
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{}, activeAt)
	log.Printf("ReadFilteredRelationships response (only activeAt): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by activeAt")
	var foundRel1, foundRel3 bool
	for _, r := range rels {
		if r["id"] == "rel1" {
			foundRel1 = true
		}
		if r["id"] == "rel3" {
			foundRel3 = true
		}
	}
	assert.True(t, foundRel1, "Expected rel1 to be active at 2025-05-03T00:00:00Z")
	assert.True(t, foundRel3, "Expected rel3 to be active at 2025-05-03T00:00:00Z")

	// 9. Filter by multiple criteria: name AND relatedEntityId
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{
		"name":            "FRIEND",
		"relatedEntityId": "B",
	}, "")
	log.Printf("ReadFilteredRelationships response (name AND relatedEntityId): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by name AND relatedEntityId")
	assert.Equal(t, 1, len(rels), "Expected 1 relationship with name FRIEND and relatedEntityId B")
	assert.Equal(t, "FRIEND", rels[0]["name"])
	assert.Equal(t, "B", rels[0]["relatedEntityId"])

	// 10. Filter by name AND direction
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{
		"name":      "COLLEAGUE",
		"direction": "OUTGOING",
	}, "")
	log.Printf("ReadFilteredRelationships response (name AND direction): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by name AND direction")
	assert.Equal(t, 1, len(rels), "Expected 1 relationship with name COLLEAGUE and direction OUTGOING")
	assert.Equal(t, "COLLEAGUE", rels[0]["name"])
	assert.Equal(t, "OUTGOING", rels[0]["direction"])

	// 11. Filter by startTime AND endTime (should match rel2)
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{
		"startTime": "2025-04-02T00:00:00Z",
		"endTime":   "2025-05-01T00:00:00Z",
	}, "")
	log.Printf("ReadFilteredRelationships response (startTime AND endTime): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by startTime AND endTime")
	assert.Equal(t, 1, len(rels), "Expected 1 relationship with specific startTime and endTime")
	assert.Equal(t, "rel2", rels[0]["id"])

	// 12. Filter by name AND startTime
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{
		"name":      "FRIEND",
		"startTime": "2025-04-01T00:00:00Z",
	}, "")
	log.Printf("ReadFilteredRelationships response (name AND startTime): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by name AND startTime")
	assert.Equal(t, 1, len(rels), "Expected 1 relationship with name FRIEND and startTime 2025-04-01T00:00:00Z")
	assert.Equal(t, "rel1", rels[0]["id"])

	// 13. Filter by direction AND activeAt
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{
		"direction": "OUTGOING",
	}, "2025-04-15T00:00:00Z")
	log.Printf("ReadFilteredRelationships response (direction AND activeAt): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by direction AND activeAt")
	for _, r := range rels {
		assert.Equal(t, "OUTGOING", r["direction"])
		// Verify that the relationship is active at the specified time
		startTime := r["startTime"].(string)
		endTimeVal, hasEndTime := r["endTime"]
		endTime, isString := endTimeVal.(string)
		if hasEndTime && isString && endTime != "" {
			assert.True(t, startTime <= "2025-04-15T00:00:00Z" && "2025-04-15T00:00:00Z" <= endTime,
				"Relationship should be active at 2025-04-15T00:00:00Z")
		} else {
			assert.True(t, startTime <= "2025-04-15T00:00:00Z",
				"Relationship without endTime should be active if startTime <= activeAt")
		}
	}

	// 14. Filter by name AND relatedEntityId AND direction
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{
		"name":            "FRIEND",
		"relatedEntityId": "B",
		"direction":       "OUTGOING",
	}, "")
	log.Printf("ReadFilteredRelationships response (name AND relatedEntityId AND direction): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by name AND relatedEntityId AND direction")
	assert.Equal(t, 1, len(rels), "Expected 1 relationship with all three filters")
	assert.Equal(t, "rel1", rels[0]["id"])

	// 15. Test with non-existent filters (should return empty)
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{
		"name": "NONEXISTENT",
	}, "")
	log.Printf("ReadFilteredRelationships response (non-existent name): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by non-existent name")
	assert.Equal(t, 0, len(rels), "Expected 0 relationships with non-existent name")

	// 16. Test with non-existent relatedEntityId
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{
		"relatedEntityId": "NONEXISTENT",
	}, "")
	log.Printf("ReadFilteredRelationships response (non-existent relatedEntityId): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by non-existent relatedEntityId")
	assert.Equal(t, 0, len(rels), "Expected 0 relationships with non-existent relatedEntityId")

	// 17. Test activeAt with a time when no relationships are active
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{}, "2025-06-01T00:00:00Z")
	log.Printf("ReadFilteredRelationships response (activeAt future): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by activeAt in future")
	// Should only return relationships without endTime or with endTime > 2025-06-01T00:00:00Z
	for _, r := range rels {
		endTimeVal, hasEndTime := r["endTime"]
		endTime, isString := endTimeVal.(string)
		if hasEndTime && isString && endTime != "" {
			assert.True(t, endTime > "2025-06-01T00:00:00Z",
				"Relationship should be active if endTime is in the future")
		}
	}

	// 18. Test activeAt with a time before all relationships started
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{}, "2025-03-01T00:00:00Z")
	log.Printf("ReadFilteredRelationships response (activeAt past): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by activeAt in past")
	assert.Equal(t, 0, len(rels), "Expected 0 relationships active before any started")

	// 19. Test combination of filters with activeAt
	rels, err = repository.ReadFilteredRelationships(ctx, "A", map[string]interface{}{
		"name":      "FRIEND",
		"direction": "OUTGOING",
	}, "2025-04-15T00:00:00Z")
	log.Printf("ReadFilteredRelationships response (name AND direction AND activeAt): %+v", rels)
	assert.Nil(t, err, "Expected no error when filtering by name AND direction AND activeAt")
	assert.Equal(t, 1, len(rels), "Expected 1 FRIEND relationship that is OUTGOING and active at 2025-04-15T00:00:00Z")
	assert.Equal(t, "rel1", rels[0]["id"])
}

func TestFilterEntitiesWithPartialNameMatch(t *testing.T) {
	ctx := context.Background()

	kind := &pb.Kind{
		Major: "Person",
		Minor: "Employee",
	}

	// Create test entities with unique realistic names for testing partial matches
	testEntities := []map[string]interface{}{
		{
			"Id":      "filter_partial_name_test_alpha",
			"Name":    "Alice Martinez",
			"Created": "2025-04-01T00:00:00Z",
		},
		{
			"Id":      "filter_partial_name_test_alpha_gamma",
			"Name":    "Alice Morrison",
			"Created": "2025-04-01T00:00:00Z",
		},
		{
			"Id":      "filter_partial_name_test_beta_delta",
			"Name":    "Benjamin Martinez",
			"Created": "2025-04-01T00:00:00Z",
		},
		{
			"Id":      "filter_partial_name_test_gamma_epsilon",
			"Name":    "Charlotte Morrison",
			"Created": "2025-04-01T00:00:00Z",
		},
		{
			"Id":      "filter_partial_name_test_alpha_beta_gamma",
			"Name":    "Alice Martinez Morrison",
			"Created": "2025-04-01T00:00:00Z",
		},
		// Add entities with special regex characters to test regex injection prevention
		{
			"Id":      "filter_partial_name_test_regex_1",
			"Name":    "Alice (test)",
			"Created": "2025-04-01T00:00:00Z",
		},
		{
			"Id":      "filter_partial_name_test_regex_2",
			"Name":    "Ben.jamin",
			"Created": "2025-04-01T00:00:00Z",
		},
	}

	// Create all test entities
	for _, entity := range testEntities {
		_, err := repository.CreateGraphEntity(ctx, kind, entity)
		assert.Nil(t, err, "Expected no error when creating test entity: %s", entity["Name"])
	}

	// Helper function to extract names from entities
	getNames := func(entities []map[string]interface{}) []string {
		names := []string{}
		for _, e := range entities {
			names = append(names, e["name"].(string))
		}
		return names
	}

	t.Run("partial match at beginning", func(t *testing.T) {
		filters := map[string]interface{}{
			"name": "Alice",
		}
		entities, err := repository.FilterEntities(ctx, kind, filters)
		assert.Nil(t, err, "Expected no error when filtering by partial name 'Alice'")
		assert.Equal(t, 4, len(entities), "Expected 4 entities with 'Alice' in the name")
		names := getNames(entities)
		assert.Contains(t, names, "Alice Martinez", "Expected 'Alice Martinez' to be in results")
		assert.Contains(t, names, "Alice Morrison", "Expected 'Alice Morrison' to be in results")
		assert.Contains(t, names, "Alice Martinez Morrison", "Expected 'Alice Martinez Morrison' to be in results")
		assert.Contains(t, names, "Alice (test)", "Expected 'Alice (test)' to be in results")
	})

	t.Run("partial match in middle", func(t *testing.T) {
		filters := map[string]interface{}{
			"name": "Martinez",
		}
		entities, err := repository.FilterEntities(ctx, kind, filters)
		assert.Nil(t, err, "Expected no error when filtering by partial name 'Martinez'")
		assert.Equal(t, 3, len(entities), "Expected 3 entities with 'Martinez' in the name")
		names := getNames(entities)
		assert.Contains(t, names, "Alice Martinez", "Expected 'Alice Martinez' to be in results")
		assert.Contains(t, names, "Benjamin Martinez", "Expected 'Benjamin Martinez' to be in results")
		assert.Contains(t, names, "Alice Martinez Morrison", "Expected 'Alice Martinez Morrison' to be in results")
	})

	t.Run("case-insensitive match lowercase", func(t *testing.T) {
		filters := map[string]interface{}{
			"name": "alice",
		}
		entities, err := repository.FilterEntities(ctx, kind, filters)
		assert.Nil(t, err, "Expected no error when filtering by lowercase partial name 'alice'")
		assert.Equal(t, 4, len(entities), "Expected 4 entities with 'alice' (case-insensitive) in the name")
		names := getNames(entities)
		assert.Contains(t, names, "Alice Martinez", "Expected 'Alice Martinez' to be in results (case-insensitive match)")
		assert.Contains(t, names, "Alice Morrison", "Expected 'Alice Morrison' to be in results (case-insensitive match)")
		assert.Contains(t, names, "Alice (test)", "Expected 'Alice (test)' to be in results (case-insensitive match)")
	})

	t.Run("case-insensitive match uppercase", func(t *testing.T) {
		filters := map[string]interface{}{
			"name": "MORRISON",
		}
		entities, err := repository.FilterEntities(ctx, kind, filters)
		assert.Nil(t, err, "Expected no error when filtering by uppercase partial name 'MORRISON'")
		assert.Equal(t, 3, len(entities), "Expected 3 entities with 'MORRISON' (case-insensitive) in the name")
		names := getNames(entities)
		assert.Contains(t, names, "Alice Morrison", "Expected 'Alice Morrison' to be in results (case-insensitive match)")
		assert.Contains(t, names, "Charlotte Morrison", "Expected 'Charlotte Morrison' to be in results (case-insensitive match)")
		assert.Contains(t, names, "Alice Martinez Morrison", "Expected 'Alice Martinez Morrison' to be in results (case-insensitive match)")
	})

	t.Run("partial match in middle of multi-word name", func(t *testing.T) {
		filters := map[string]interface{}{
			"name": "Morrison",
		}
		entities, err := repository.FilterEntities(ctx, kind, filters)
		assert.Nil(t, err, "Expected no error when filtering by partial name 'Morrison'")
		assert.Equal(t, 3, len(entities), "Expected 3 entities with 'Morrison' in the name")
		names := getNames(entities)
		assert.Contains(t, names, "Alice Morrison", "Expected 'Alice Morrison' to be in results")
		assert.Contains(t, names, "Charlotte Morrison", "Expected 'Charlotte Morrison' to be in results")
		assert.Contains(t, names, "Alice Martinez Morrison", "Expected 'Alice Martinez Morrison' to be in results")
	})

	t.Run("no match returns empty", func(t *testing.T) {
		filters := map[string]interface{}{
			"name": "Zachary Wellington",
		}
		entities, err := repository.FilterEntities(ctx, kind, filters)
		assert.Nil(t, err, "Expected no error when filtering by non-existent name")
		assert.Equal(t, 0, len(entities), "Expected 0 entities with 'Zachary Wellington' in the name")
	})

	t.Run("combined with MinorKind filter", func(t *testing.T) {
		filters := map[string]interface{}{
			"name": "Alice",
		}
		kindWithMinor := &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		}
		entities, err := repository.FilterEntities(ctx, kindWithMinor, filters)
		assert.Nil(t, err, "Expected no error when filtering by name and MinorKind")
		assert.Equal(t, 4, len(entities), "Expected 4 entities with 'Alice' in the name and Employee minor kind")
	})

	t.Run("combined with created date filter", func(t *testing.T) {
		filters := map[string]interface{}{
			"name":    "Alice",
			"created": "2025-04-01T00:00:00Z",
		}
		entities, err := repository.FilterEntities(ctx, kind, filters)
		assert.Nil(t, err, "Expected no error when filtering by name and created date")
		assert.Equal(t, 4, len(entities), "Expected 4 entities with 'Alice' in the name and matching created date")
		names := getNames(entities)
		expectedNames := []string{"Alice Martinez", "Alice Morrison", "Alice Martinez Morrison", "Alice (test)"}
		for _, name := range names {
			assert.Contains(t, expectedNames, name, "Expected entity name to be one of the Alice entities")
		}
	})

	t.Run("single character partial match", func(t *testing.T) {
		filters := map[string]interface{}{
			"name": "A",
		}
		entities, err := repository.FilterEntities(ctx, kind, filters)
		assert.Nil(t, err, "Expected no error when filtering by single character 'A'")
		assert.GreaterOrEqual(t, len(entities), 4, "Expected at least 4 entities with 'A' in the name (Alice entities)")
	})

	t.Run("regex injection prevention - parentheses", func(t *testing.T) {
		// Test that special regex characters like parentheses are treated as literal characters
		filters := map[string]interface{}{
			"name": "Alice (test)",
		}
		entities, err := repository.FilterEntities(ctx, kind, filters)
		assert.Nil(t, err, "Expected no error when filtering by name with parentheses")
		assert.Equal(t, 1, len(entities), "Expected 1 entity with 'Alice (test)' in the name")
		names := getNames(entities)
		assert.Contains(t, names, "Alice (test)", "Expected 'Alice (test)' to be in results")
		// Verify it doesn't match other Alice entities (regex injection would cause false matches)
		assert.NotContains(t, names, "Alice Martinez", "Should not match 'Alice Martinez' when searching for 'Alice (test)'")
	})

	t.Run("regex injection prevention - dot", func(t *testing.T) {
		// Test that special regex characters like dot are treated as literal characters
		filters := map[string]interface{}{
			"name": "Ben.jamin",
		}
		entities, err := repository.FilterEntities(ctx, kind, filters)
		assert.Nil(t, err, "Expected no error when filtering by name with dot")
		assert.Equal(t, 1, len(entities), "Expected 1 entity with 'Ben.jamin' in the name")
		names := getNames(entities)
		assert.Contains(t, names, "Ben.jamin", "Expected 'Ben.jamin' to be in results")
		// Verify it doesn't match "Benjamin" (regex injection would cause false matches)
		assert.NotContains(t, names, "Benjamin Martinez", "Should not match 'Benjamin Martinez' when searching for 'Ben.jamin'")
	})

	t.Run("regex injection prevention - partial match with special chars", func(t *testing.T) {
		// Test partial matching with special characters
		filters := map[string]interface{}{
			"name": "(test)",
		}
		entities, err := repository.FilterEntities(ctx, kind, filters)
		assert.Nil(t, err, "Expected no error when filtering by partial name with parentheses")
		assert.Equal(t, 1, len(entities), "Expected 1 entity with '(test)' in the name")
		names := getNames(entities)
		assert.Contains(t, names, "Alice (test)", "Expected 'Alice (test)' to be in results")
	})
}
