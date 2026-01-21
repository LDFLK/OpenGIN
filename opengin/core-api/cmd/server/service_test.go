// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"os"
// 	"strings"
// 	"testing"

// 	"google.golang.org/protobuf/types/known/anypb"
// 	"google.golang.org/protobuf/types/known/structpb"
// 	"google.golang.org/protobuf/types/known/wrapperspb"

// 	"lk/datafoundation/core-api/db/config"
// 	mongorepository "lk/datafoundation/core-api/db/repository/mongo"
// 	neo4jrepository "lk/datafoundation/core-api/db/repository/neo4j"
// 	postgres "lk/datafoundation/core-api/db/repository/postgres"
// 	pb "lk/datafoundation/core-api/lk/datafoundation/core-api"
// )

// var server *Server

// // createNameValue is a helper function to properly create a TimeBasedValue for Name field
// func createNameValue(startTime, name string) *pb.TimeBasedValue {
// 	value, _ := anypb.New(&wrapperspb.StringValue{Value: name})
// 	return &pb.TimeBasedValue{
// 		StartTime: startTime,
// 		Value:     value,
// 	}
// }

// // TestMain sets up the actual MongoDB, Neo4j, and PostgreSQL repositories before running the tests
// func TestMain(m *testing.M) {
// 	// Load environment variables for database configurations
// 	neo4jConfig := &config.Neo4jConfig{
// 		URI:      os.Getenv("NEO4J_URI"),
// 		Username: os.Getenv("NEO4J_USER"),
// 		Password: os.Getenv("NEO4J_PASSWORD"),
// 	}

// 	mongoConfig := &config.MongoConfig{
// 		URI:        os.Getenv("MONGO_URI"),
// 		DBName:     os.Getenv("MONGO_DB_NAME"),
// 		Collection: os.Getenv("MONGO_COLLECTION"),
// 	}

// 	postgresConfig := &postgres.Config{
// 		Host:     os.Getenv("POSTGRES_HOST"),
// 		Port:     os.Getenv("POSTGRES_PORT"),
// 		User:     os.Getenv("POSTGRES_USER"),
// 		Password: os.Getenv("POSTGRES_PASSWORD"),
// 		DBName:   os.Getenv("POSTGRES_DB"),
// 		SSLMode:  os.Getenv("POSTGRES_SSL_MODE"),
// 	}

// 	// Initialize Neo4j repository
// 	ctx := context.Background()
// 	neo4jRepo, err := neo4jrepository.NewNeo4jRepository(ctx, neo4jConfig)
// 	if err != nil {
// 		log.Fatalf("Failed to initialize Neo4j repository: %v", err)
// 	}
// 	defer neo4jRepo.Close(ctx)

// 	// Initialize MongoDB repository
// 	mongoRepo := mongorepository.NewMongoRepository(ctx, mongoConfig)
// 	if mongoRepo == nil {
// 		log.Fatalf("Failed to initialize MongoDB repository")
// 	}

// 	// Initialize PostgreSQL repository
// 	postgresRepo, err := postgres.NewPostgresRepository(*postgresConfig)
// 	if err != nil {
// 		log.Fatalf("Failed to initialize PostgreSQL repository: %v", err)
// 	}
// 	defer postgresRepo.Close()

// 	// Create the server with the initialized repositories
// 	server = &Server{
// 		mongoRepo:    mongoRepo,
// 		neo4jRepo:    neo4jRepo,
// 		postgresRepo: postgresRepo,
// 	}

// 	// Run the tests
// 	code := m.Run()

// 	// Exit with the test result code
// 	os.Exit(code)
// }

// // TestServiceCreateEntity tests creating an entity through the service layer
// func TestServiceCreateEntity(t *testing.T) {
// 	ctx := context.Background()

// 	// Create a simple entity
// 	entity := &pb.Entity{
// 		Id:      "service_test_entity_1",
// 		Kind:    &pb.Kind{Major: "Person", Minor: "Minister"},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "John Doe"),
// 		Created: "2025-03-18T00:00:00Z",
// 	}

// 	// Create the entity
// 	resp, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}
// 	if resp == nil {
// 		t.Fatal("CreateEntity() returned nil response")
// 	}
// 	if resp.Id != entity.Id {
// 		t.Errorf("CreateEntity() response ID = %v, want %v", resp.Id, entity.Id)
// 	}

// 	log.Printf("Successfully created entity: %v", resp.Id)
// }

// // TestServiceCreateEntityWithValidFields tests creating a graph entity with all valid required fields
// func TestServiceCreateEntityWithValidFields(t *testing.T) {
// 	ctx := context.Background()

// 	// Create entity with all required fields
// 	entity := &pb.Entity{
// 		Id: "service_valid_fields_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Organization",
// 			Minor: "Department",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Engineering Department"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	resp, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() with valid fields error = %v", err)
// 	}
// 	if resp == nil {
// 		t.Fatal("CreateEntity() returned nil response")
// 	}
// 	if resp.Id != entity.Id {
// 		t.Errorf("CreateEntity() response ID = %v, want %v", resp.Id, entity.Id)
// 	}

// 	log.Printf("Successfully created entity with valid fields: %v", resp.Id)
// }

// // TestServiceCreateEntityWithMissingKindMajor tests that creating entity without Kind.Major fails
// func TestServiceCreateEntityWithMissingKindMajor(t *testing.T) {
// 	ctx := context.Background()

// 	// Try to create entity without Kind.Major
// 	entity := &pb.Entity{
// 		Id: "service_missing_major_entity_1",
// 		Kind: &pb.Kind{
// 			Minor: "Employee",
// 			// Major is missing
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Invalid User"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err == nil {
// 		t.Error("Expected error when creating entity without Kind.Major, but got none")
// 	} else {
// 		log.Printf("CreateEntity correctly failed without Kind.Major: %v", err)
// 	}

// 	log.Printf("Successfully verified that Kind.Major is required")
// }

// // TestServiceCreateEntityWithMissingKindMinor tests that creating entity without Kind.Minor fails
// func TestServiceCreateEntityWithMissingKindMinor(t *testing.T) {
// 	ctx := context.Background()

// 	// Try to create entity without Kind.Minor
// 	entity := &pb.Entity{
// 		Id: "service_missing_minor_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			// Minor is missing
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Invalid User"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err == nil {
// 		t.Error("Expected error when creating entity without Kind.Minor, but got none")
// 	} else {
// 		log.Printf("CreateEntity correctly failed without Kind.Minor: %v", err)
// 	}

// 	log.Printf("Successfully verified that Kind.Minor is required")
// }

// // TestServiceCreateEntityWithDuplicateId tests that creating entity with duplicate ID fails
// func TestServiceCreateEntityWithDuplicateId(t *testing.T) {
// 	ctx := context.Background()

// 	// Create first entity
// 	entity1 := &pb.Entity{
// 		Id: "service_duplicate_id_entity",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "First User"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(first entity) error = %v", err)
// 	}

// 	// Try to create second entity with same ID
// 	entity2 := &pb.Entity{
// 		Id: "service_duplicate_id_entity", // Same ID!
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Manager",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Second User"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err == nil {
// 		t.Error("Expected error when creating entity with duplicate ID, but got none")
// 	} else {
// 		log.Printf("CreateEntity correctly failed with duplicate ID: %v", err)
// 	}

// 	// Verify original entity wasn't modified
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_duplicate_id_entity"},
// 		Output: []string{},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	var stringValue wrapperspb.StringValue
// 	err = readResp.Name.GetValue().UnmarshalTo(&stringValue)
// 	if err != nil {
// 		t.Fatalf("Error unpacking Name value: %v", err)
// 	}
// 	if stringValue.Value != "First User" {
// 		t.Errorf("Entity was modified, name = %v, should remain 'First User'", stringValue.Value)
// 	}

// 	log.Printf("Successfully verified that duplicate IDs are rejected")
// }

// // TestServiceCreateEntityWithDuplicateRelationshipId tests that creating relationship with duplicate ID fails
// func TestServiceCreateEntityWithDuplicateRelationshipId(t *testing.T) {
// 	ctx := context.Background()

// 	// Create two target entities
// 	entity1 := &pb.Entity{
// 		Id: "service_dup_rel_target_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Target 1"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_dup_rel_target_2",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Target 2"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(target1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(target2) error = %v", err)
// 	}

// 	// Create entity with first relationship using a specific ID
// 	entity3 := &pb.Entity{
// 		Id: "service_dup_rel_source_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Manager",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Manager 1"),
// 		Created: "2025-04-01T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"duplicate_relationship_id": {
// 				Id:              "duplicate_relationship_id",
// 				Name:            "MANAGES",
// 				RelatedEntityId: "service_dup_rel_target_1",
// 				StartTime:       "2025-04-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err = server.CreateEntity(ctx, entity3)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(first relationship) error = %v", err)
// 	}

// 	// Try to create another entity with a relationship using the SAME ID
// 	entity4 := &pb.Entity{
// 		Id: "service_dup_rel_source_2",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Manager",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Manager 2"),
// 		Created: "2025-04-01T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"duplicate_relationship_id": { // Same relationship ID!
// 				Id:              "duplicate_relationship_id",
// 				Name:            "SUPERVISES",
// 				RelatedEntityId: "service_dup_rel_target_2",
// 				StartTime:       "2025-04-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err = server.CreateEntity(ctx, entity4)
// 	if err == nil {
// 		t.Error("Expected error when creating entity with duplicate relationship ID, but got none")
// 	} else {
// 		log.Printf("CreateEntity correctly failed with duplicate relationship ID: %v", err)
// 	}

// 	log.Printf("Successfully verified that duplicate relationship IDs are rejected")
// }

// // TestServiceCreateEntityWithMetadataAndAttributes tests creating an entity with both metadata and attributes
// func TestServiceCreateEntityWithMetadataAndAttributes(t *testing.T) {
// 	ctx := context.Background()

// 	// Create metadata
// 	metadata := make(map[string]*anypb.Any)
// 	metadata["department"], _ = anypb.New(&wrapperspb.StringValue{Value: "Research"})
// 	metadata["clearance_level"], _ = anypb.New(&wrapperspb.StringValue{Value: "Level 3"})

// 	// Create attributes with all three types
// 	attributes := make(map[string]*pb.TimeBasedValueList)

// 	// Tabular attribute
// 	salaryData := map[string]interface{}{
// 		"columns": []interface{}{"amount", "currency"},
// 		"rows":    []interface{}{[]interface{}{"130000", "USD"}},
// 	}
// 	salaryStruct, _ := structpb.NewStruct(salaryData)
// 	salaryValue, _ := anypb.New(salaryStruct)

// 	attributes["salary"] = &pb.TimeBasedValueList{
// 		Values: []*pb.TimeBasedValue{
// 			{
// 				StartTime: "2025-04-01T00:00:00Z",
// 				EndTime:   "",
// 				Value:     salaryValue,
// 			},
// 		},
// 	}

// 	// Graph attribute
// 	projectData := map[string]interface{}{
// 		"nodes": []interface{}{
// 			map[string]interface{}{"id": "proj1", "name": "AI Research"},
// 		},
// 		"edges": []interface{}{},
// 	}
// 	projectStruct, _ := structpb.NewStruct(projectData)
// 	projectValue, _ := anypb.New(projectStruct)

// 	attributes["projects"] = &pb.TimeBasedValueList{
// 		Values: []*pb.TimeBasedValue{
// 			{
// 				StartTime: "2025-04-01T00:00:00Z",
// 				EndTime:   "",
// 				Value:     projectValue,
// 			},
// 		},
// 	}

// 	// Map attribute
// 	credentialsData := map[string]interface{}{
// 		"username": "researcher123",
// 		"access":   "read-write",
// 	}
// 	credentialsStruct, _ := structpb.NewStruct(credentialsData)
// 	credentialsValue, _ := anypb.New(credentialsStruct)

// 	attributes["credentials"] = &pb.TimeBasedValueList{
// 		Values: []*pb.TimeBasedValue{
// 			{
// 				StartTime: "2025-04-01T00:00:00Z",
// 				EndTime:   "",
// 				Value:     credentialsValue,
// 			},
// 		},
// 	}

// 	// Create entity with both metadata and attributes
// 	entity := &pb.Entity{
// 		Id: "service_metadata_attributes_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Researcher",
// 		},
// 		Name:       createNameValue("2025-04-01T00:00:00Z", "Dr. Smith"),
// 		Created:    "2025-04-01T00:00:00Z",
// 		Metadata:   metadata,
// 		Attributes: attributes,
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() with metadata and attributes error = %v", err)
// 	}

// 	// Read entity back with both metadata and attributes
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Id:         "service_metadata_attributes_entity_1",
// 			Attributes: attributes,
// 		},
// 		Output: []string{"metadata", "attributes"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	// Verify metadata was stored
// 	if len(readResp.Metadata) != 2 {
// 		t.Errorf("Expected 2 metadata fields, got %d", len(readResp.Metadata))
// 	}
// 	if _, exists := readResp.Metadata["department"]; !exists {
// 		t.Error("Expected 'department' metadata field not found")
// 	}
// 	if _, exists := readResp.Metadata["clearance_level"]; !exists {
// 		t.Error("Expected 'clearance_level' metadata field not found")
// 	}

// 	// Verify all three attribute types were stored
// 	if len(readResp.Attributes) == 0 {
// 		t.Error("Expected attributes to be stored, but got empty attributes")
// 	}
// 	if _, exists := readResp.Attributes["salary"]; !exists {
// 		t.Error("Expected 'salary' (tabular) attribute not found")
// 	}
// 	if _, exists := readResp.Attributes["projects"]; !exists {
// 		t.Error("Expected 'projects' (graph) attribute not found")
// 	}
// 	if _, exists := readResp.Attributes["credentials"]; !exists {
// 		t.Error("Expected 'credentials' (map) attribute not found")
// 	}

// 	log.Printf("Successfully created entity with both metadata and all three attribute types")
// }

// // TestServiceCreateEntityWithMetadataAttributesAndRelationships tests creating a complete entity
// func TestServiceCreateEntityWithMetadataAttributesAndRelationships(t *testing.T) {
// 	ctx := context.Background()

// 	// Create a target entity for the relationship
// 	targetEntity := &pb.Entity{
// 		Id: "service_complete_target_entity",
// 		Kind: &pb.Kind{
// 			Major: "Organization",
// 			Minor: "Department",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Engineering Dept"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, targetEntity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(target) error = %v", err)
// 	}

// 	// Create metadata
// 	metadata := make(map[string]*anypb.Any)
// 	metadata["role"], _ = anypb.New(&wrapperspb.StringValue{Value: "Lead Engineer"})

// 	// Create attributes
// 	attributes := make(map[string]*pb.TimeBasedValueList)

// 	salaryData := map[string]interface{}{
// 		"columns": []interface{}{"amount", "currency"},
// 		"rows":    []interface{}{[]interface{}{"160000", "USD"}},
// 	}
// 	salaryStruct, _ := structpb.NewStruct(salaryData)
// 	salaryValue, _ := anypb.New(salaryStruct)

// 	attributes["salary"] = &pb.TimeBasedValueList{
// 		Values: []*pb.TimeBasedValue{
// 			{
// 				StartTime: "2025-04-01T00:00:00Z",
// 				EndTime:   "",
// 				Value:     salaryValue,
// 			},
// 		},
// 	}

// 	// Create entity with metadata, attributes, AND relationships
// 	entity := &pb.Entity{
// 		Id: "service_complete_entity",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Engineer",
// 		},
// 		Name:       createNameValue("2025-04-01T00:00:00Z", "Complete User"),
// 		Created:    "2025-04-01T00:00:00Z",
// 		Metadata:   metadata,
// 		Attributes: attributes,
// 		Relationships: map[string]*pb.Relationship{
// 			"service_complete_rel": {
// 				Id:              "service_complete_rel",
// 				Name:            "MEMBER_OF",
// 				RelatedEntityId: "service_complete_target_entity",
// 				StartTime:       "2025-04-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err = server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() with metadata, attributes, and relationships error = %v", err)
// 	}

// 	// Read entity back with all components
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Id:         "service_complete_entity",
// 			Attributes: attributes,
// 		},
// 		Output: []string{"metadata", "attributes", "relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	// Verify metadata
// 	if len(readResp.Metadata) == 0 {
// 		t.Error("Expected metadata to be stored")
// 	}
// 	if _, exists := readResp.Metadata["role"]; !exists {
// 		t.Error("Expected 'role' metadata field not found")
// 	}

// 	// Verify attributes
// 	if len(readResp.Attributes) == 0 {
// 		t.Error("Expected attributes to be stored")
// 	}
// 	if _, exists := readResp.Attributes["salary"]; !exists {
// 		t.Error("Expected 'salary' attribute not found")
// 	}

// 	// Verify relationships
// 	if len(readResp.Relationships) == 0 {
// 		t.Error("Expected relationships to be stored")
// 	}
// 	if _, exists := readResp.Relationships["service_complete_rel"]; !exists {
// 		t.Error("Expected 'service_complete_rel' relationship not found")
// 	}

// 	log.Printf("Successfully created complete entity with metadata, attributes, and relationships")
// }

// // TestServiceReadEntity tests reading an entity through the service layer
// func TestServiceReadEntity(t *testing.T) {
// 	ctx := context.Background()

// 	// First create an entity to read
// 	entity := &pb.Entity{
// 		Id: "service_read_test_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Minister",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Jane Doe"),
// 		Created: "2025-03-18T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}

// 	// Read the entity back
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_read_test_entity_1"},
// 		Output: []string{}, // Request basic info only
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}
// 	if readResp == nil {
// 		t.Fatal("ReadEntity() returned nil response")
// 	}
// 	if readResp.Id != entity.Id {
// 		t.Errorf("ReadEntity() response ID = %v, want %v", readResp.Id, entity.Id)
// 	}
// 	if readResp.Kind.Major != "Person" {
// 		t.Errorf("ReadEntity() response Kind.Major = %v, want Person", readResp.Kind.Major)
// 	}

// 	log.Printf("Successfully read entity: %v", readResp.Id)
// }

// // TestServiceCreateEntityWithRelationships tests creating entities with relationships
// func TestServiceCreateEntityWithRelationships(t *testing.T) {
// 	ctx := context.Background()

// 	// Create first entity
// 	entity1 := &pb.Entity{
// 		Id: "service_test_entity_3",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Alice"),
// 		Created: "2025-03-18T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	// Create second entity
// 	entity2 := &pb.Entity{
// 		Id: "service_test_entity_4",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Bob"),
// 		Created: "2025-03-18T00:00:00Z",
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	// Create third entity with relationship to entity2
// 	entity3 := &pb.Entity{
// 		Id: "service_test_entity_5",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Charlie"),
// 		Created: "2025-03-18T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_test_rel_1": {
// 				Id:              "service_test_rel_1",
// 				Name:            "KNOWS",
// 				RelatedEntityId: "service_test_entity_4",
// 				StartTime:       "2025-03-18T00:00:00Z",
// 			},
// 		},
// 	}

// 	resp, err := server.CreateEntity(ctx, entity3)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity3 with relationships) error = %v", err)
// 	}
// 	if resp == nil {
// 		t.Fatal("CreateEntity() returned nil response")
// 	}

// 	log.Printf("Successfully created entity with relationships: %v", resp.Id)
// }

// // TestServiceReadEntityWithRelationships tests reading entities with relationships
// func TestServiceReadEntityWithRelationships(t *testing.T) {
// 	ctx := context.Background()

// 	// Create two entities
// 	entity1 := &pb.Entity{
// 		Id: "service_test_entity_6",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "David"),
// 		Created: "2025-03-18T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_test_entity_7",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Eve"),
// 		Created: "2025-03-18T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_test_rel_2": {
// 				Id:              "service_test_rel_2",
// 				Name:            "REPORTS_TO",
// 				RelatedEntityId: "service_test_entity_6",
// 				StartTime:       "2025-03-18T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2 with relationship) error = %v", err)
// 	}

// 	// Read entity with relationships
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_test_entity_7"},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}
// 	if readResp == nil {
// 		t.Fatal("ReadEntity() returned nil response")
// 	}
// 	if len(readResp.Relationships) == 0 {
// 		t.Error("ReadEntity() returned no relationships")
// 	}

// 	// Verify the relationship exists
// 	if _, exists := readResp.Relationships["service_test_rel_2"]; !exists {
// 		t.Error("Expected relationship 'service_test_rel_2' not found")
// 	}

// 	log.Printf("Successfully read entity with relationships: %v", readResp.Id)
// }

// // TestServiceReadNonExistentEntity tests that reading a non-existent entity returns an error
// func TestServiceReadNonExistentEntity(t *testing.T) {
// 	ctx := context.Background()

// 	// Try to read an entity that doesn't exist
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "non_existent_entity_12345"},
// 		Output: []string{},
// 	}

// 	_, err := server.ReadEntity(ctx, readReq)
// 	if err == nil {
// 		t.Error("Expected error when reading non-existent entity, but got none")
// 	} else {
// 		log.Printf("ReadEntity correctly failed for non-existent entity: %v", err)
// 	}

// 	log.Printf("Successfully verified that reading non-existent entity fails")
// }

// // TestServiceReadEntityWithMetadata tests reading an entity with metadata output
// func TestServiceReadEntityWithMetadata(t *testing.T) {
// 	ctx := context.Background()

// 	// Create entity with metadata
// 	metadata := make(map[string]*anypb.Any)
// 	metadata["position"], _ = anypb.New(&wrapperspb.StringValue{Value: "Software Engineer"})
// 	metadata["team"], _ = anypb.New(&wrapperspb.StringValue{Value: "Platform"})

// 	entity := &pb.Entity{
// 		Id: "service_read_metadata_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:     createNameValue("2025-04-01T00:00:00Z", "Metadata Reader"),
// 		Created:  "2025-04-01T00:00:00Z",
// 		Metadata: metadata,
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}

// 	// Read entity with metadata output
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_read_metadata_entity_1"},
// 		Output: []string{"metadata"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	// Verify metadata was returned
// 	if len(readResp.Metadata) != 2 {
// 		t.Errorf("Expected 2 metadata fields, got %d", len(readResp.Metadata))
// 	}
// 	if _, exists := readResp.Metadata["position"]; !exists {
// 		t.Error("Expected 'position' metadata field not found")
// 	}
// 	if _, exists := readResp.Metadata["team"]; !exists {
// 		t.Error("Expected 'team' metadata field not found")
// 	}

// 	log.Printf("Successfully read entity with metadata")
// }

// // TestServiceReadEntityWithFilteredRelationships tests reading entity with filtered relationships
// func TestServiceReadEntityWithFilteredRelationships(t *testing.T) {
// 	ctx := context.Background()

// 	// Create three entities
// 	entity1 := &pb.Entity{
// 		Id: "service_read_filtered_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Person A"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_read_filtered_entity_2",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Person B"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	// Create entity with multiple relationships
// 	entity3 := &pb.Entity{
// 		Id: "service_read_filtered_entity_3",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Manager",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Manager"),
// 		Created: "2025-04-01T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_read_filtered_rel_1": {
// 				Id:              "service_read_filtered_rel_1",
// 				Name:            "MANAGES",
// 				RelatedEntityId: "service_read_filtered_entity_1",
// 				StartTime:       "2025-04-01T00:00:00Z",
// 			},
// 			"service_read_filtered_rel_2": {
// 				Id:              "service_read_filtered_rel_2",
// 				Name:            "SUPERVISES",
// 				RelatedEntityId: "service_read_filtered_entity_2",
// 				StartTime:       "2025-04-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity3)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity3) error = %v", err)
// 	}

// 	// Read with filtered relationships (filter by relationship ID)
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Id: "service_read_filtered_entity_3",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_read_filtered_rel_1": {
// 					Id: "service_read_filtered_rel_1",
// 				},
// 			},
// 		},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() with filtered relationships error = %v", err)
// 	}

// 	// Should only get the filtered relationship
// 	if _, exists := readResp.Relationships["service_read_filtered_rel_1"]; !exists {
// 		t.Error("Expected filtered relationship 'service_read_filtered_rel_1' not found")
// 	}

// 	log.Printf("Successfully read entity with filtered relationships")
// }

// // TestServiceReadEntityWithAttributes tests reading entity with attributes
// func TestServiceReadEntityWithAttributes(t *testing.T) {
// 	ctx := context.Background()

// 	// Create attributes
// 	attributes := make(map[string]*pb.TimeBasedValueList)

// 	salaryData := map[string]interface{}{
// 		"columns": []interface{}{"amount", "currency"},
// 		"rows":    []interface{}{[]interface{}{"95000", "USD"}},
// 	}
// 	salaryStruct, _ := structpb.NewStruct(salaryData)
// 	salaryValue, _ := anypb.New(salaryStruct)

// 	attributes["salary"] = &pb.TimeBasedValueList{
// 		Values: []*pb.TimeBasedValue{
// 			{
// 				StartTime: "2025-04-01T00:00:00Z",
// 				EndTime:   "",
// 				Value:     salaryValue,
// 			},
// 		},
// 	}

// 	// Create entity with attributes
// 	entity := &pb.Entity{
// 		Id: "service_read_attributes_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:       createNameValue("2025-04-01T00:00:00Z", "Attr Reader"),
// 		Created:    "2025-04-01T00:00:00Z",
// 		Attributes: attributes,
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}

// 	// Read entity with attributes
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Id:         "service_read_attributes_entity_1",
// 			Attributes: attributes,
// 		},
// 		Output: []string{"attributes"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() with attributes error = %v", err)
// 	}

// 	// Verify attributes were returned
// 	if len(readResp.Attributes) == 0 {
// 		t.Error("Expected attributes to be returned")
// 	}
// 	if _, exists := readResp.Attributes["salary"]; !exists {
// 		t.Error("Expected 'salary' attribute not found")
// 	}

// 	log.Printf("Successfully read entity with attributes")
// }

// // TestServiceReadEntityWithMultipleOutputFields tests reading with multiple output fields at once
// func TestServiceReadEntityWithMultipleOutputFields(t *testing.T) {
// 	ctx := context.Background()

// 	// Create target entity for relationship
// 	targetEntity := &pb.Entity{
// 		Id: "service_read_multi_target",
// 		Kind: &pb.Kind{
// 			Major: "Organization",
// 			Minor: "Company",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "TechCorp"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, targetEntity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(target) error = %v", err)
// 	}

// 	// Create entity with metadata, attributes, and relationships
// 	metadata := make(map[string]*anypb.Any)
// 	metadata["employee_id"], _ = anypb.New(&wrapperspb.StringValue{Value: "EMP001"})

// 	attributes := make(map[string]*pb.TimeBasedValueList)
// 	// Use tabular structure for skills
// 	skillsData := map[string]interface{}{
// 		"columns": []interface{}{"skill", "level"},
// 		"rows":    []interface{}{[]interface{}{"Java", "Expert"}, []interface{}{"Kubernetes", "Intermediate"}},
// 	}
// 	skillsStruct, _ := structpb.NewStruct(skillsData)
// 	skillsValue, _ := anypb.New(skillsStruct)

// 	attributes["skills"] = &pb.TimeBasedValueList{
// 		Values: []*pb.TimeBasedValue{
// 			{
// 				StartTime: "2025-04-01T00:00:00Z",
// 				EndTime:   "",
// 				Value:     skillsValue,
// 			},
// 		},
// 	}

// 	entity := &pb.Entity{
// 		Id: "service_read_multi_entity",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:       createNameValue("2025-04-01T00:00:00Z", "Multi Output User"),
// 		Created:    "2025-04-01T00:00:00Z",
// 		Metadata:   metadata,
// 		Attributes: attributes,
// 		Relationships: map[string]*pb.Relationship{
// 			"service_read_multi_rel": {
// 				Id:              "service_read_multi_rel",
// 				Name:            "WORKS_FOR",
// 				RelatedEntityId: "service_read_multi_target",
// 				StartTime:       "2025-04-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err = server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}

// 	// Read with multiple output fields
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Id:         "service_read_multi_entity",
// 			Attributes: attributes,
// 		},
// 		Output: []string{"metadata", "relationships", "attributes"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() with multiple outputs error = %v", err)
// 	}

// 	// Verify all requested components were returned
// 	if len(readResp.Metadata) == 0 {
// 		t.Error("Expected metadata to be returned")
// 	}
// 	if len(readResp.Relationships) == 0 {
// 		t.Error("Expected relationships to be returned")
// 	}
// 	if len(readResp.Attributes) == 0 {
// 		t.Error("Expected attributes to be returned")
// 	}

// 	log.Printf("Successfully read entity with multiple output fields")
// }

// // TestServiceReadEntityWithNonExistentRelationship tests reading entity looking for relationship that doesn't exist
// func TestServiceReadEntityWithNonExistentRelationship(t *testing.T) {
// 	ctx := context.Background()

// 	// Create an entity without relationships
// 	entity := &pb.Entity{
// 		Id: "service_read_no_rel_entity",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "No Rel User"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}

// 	// Try to read with filter for non-existent relationship
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Id: "service_read_no_rel_entity",
// 			Relationships: map[string]*pb.Relationship{
// 				"non_existent_relationship_id": {
// 					Id: "non_existent_relationship_id",
// 				},
// 			},
// 		},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v (should succeed with empty relationships)", err)
// 	}

// 	// Should return entity with empty relationships
// 	if len(readResp.Relationships) > 0 {
// 		t.Errorf("Expected no relationships, but got %d", len(readResp.Relationships))
// 	}

// 	// Verify entity itself was returned correctly
// 	if readResp.Id != "service_read_no_rel_entity" {
// 		t.Errorf("Entity ID = %v, want 'service_read_no_rel_entity'", readResp.Id)
// 	}

// 	log.Printf("Successfully read entity with filter for non-existent relationship (returned empty)")
// }

// // TestServiceReadEntityWithRelationshipBelongingToDifferentEntity tests that relationship of another entity is not returned
// func TestServiceReadEntityWithRelationshipBelongingToDifferentEntity(t *testing.T) {
// 	ctx := context.Background()

// 	// Create three entities
// 	entity1 := &pb.Entity{
// 		Id: "service_read_other_rel_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Employee A"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_read_other_rel_entity_2",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Employee B"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	// Create entity3 with a relationship to entity1
// 	entity3 := &pb.Entity{
// 		Id: "service_read_other_rel_entity_3",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Manager",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Manager"),
// 		Created: "2025-04-01T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_read_other_rel_belongs_to_3": {
// 				Id:              "service_read_other_rel_belongs_to_3",
// 				Name:            "SUPERVISES",
// 				RelatedEntityId: "service_read_other_rel_entity_1",
// 				StartTime:       "2025-04-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity3)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity3) error = %v", err)
// 	}

// 	// Try to read entity2 and look for the relationship that belongs to entity3
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Id: "service_read_other_rel_entity_2", // Reading entity2
// 			Relationships: map[string]*pb.Relationship{
// 				"service_read_other_rel_belongs_to_3": { // This relationship belongs to entity3, not entity2
// 					Id: "service_read_other_rel_belongs_to_3",
// 				},
// 			},
// 		},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v (should succeed with empty relationships)", err)
// 	}

// 	// Should NOT return the relationship since it doesn't belong to this entity
// 	if _, exists := readResp.Relationships["service_read_other_rel_belongs_to_3"]; exists {
// 		t.Error("Should not return relationship that belongs to a different entity")
// 	}

// 	if len(readResp.Relationships) > 0 {
// 		t.Errorf("Expected no relationships, but got %d", len(readResp.Relationships))
// 	}

// 	// Verify the correct entity was returned
// 	if readResp.Id != "service_read_other_rel_entity_2" {
// 		t.Errorf("Entity ID = %v, want 'service_read_other_rel_entity_2'", readResp.Id)
// 	}

// 	log.Printf("Successfully verified that relationship belonging to different entity is not returned")
// }

// // TestServiceUpdateEntity tests updating an entity through the service layer
// func TestServiceUpdateEntity(t *testing.T) {
// 	ctx := context.Background()

// 	// Create an entity first
// 	entity := &pb.Entity{
// 		Id: "service_test_entity_8",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Mary"),
// 		Created: "2025-03-18T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}

// 	// Update the entity
// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_test_entity_8",
// 		Entity: &pb.Entity{
// 			Id:         "service_test_entity_8",
// 			Name:       createNameValue("2025-03-18T00:00:00Z", "Mary Updated"),
// 			Terminated: "2025-12-31T00:00:00Z",
// 		},
// 	}

// 	updateResp, err := server.UpdateEntity(ctx, updateReq)
// 	if err != nil {
// 		t.Fatalf("UpdateEntity() error = %v", err)
// 	}
// 	if updateResp == nil {
// 		t.Fatal("UpdateEntity() returned nil response")
// 	}
// 	if updateResp.Terminated != "2025-12-31T00:00:00Z" {
// 		t.Errorf("UpdateEntity() Terminated = %v, want 2025-12-31T00:00:00Z", updateResp.Terminated)
// 	}

// 	log.Printf("Successfully updated entity: %v", updateResp.Id)
// }

// // TestServiceDeleteEntity tests deleting an entity successfully
// func TestServiceDeleteEntityMetadata(t *testing.T) {
// 	ctx := context.Background()

// 	// Create an entity with metadata to delete
// 	metadata := make(map[string]*anypb.Any)
// 	metadata["department"], _ = anypb.New(&wrapperspb.StringValue{Value: "Finance"})

// 	entity := &pb.Entity{
// 		Id: "service_delete_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:     createNameValue("2025-04-01T00:00:00Z", "To Be Deleted"),
// 		Created:  "2025-04-01T00:00:00Z",
// 		Metadata: metadata,
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}

// 	// Verify entity exists with metadata
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_delete_entity_1"},
// 		Output: []string{"metadata"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() before delete error = %v", err)
// 	}
// 	if len(readResp.Metadata) == 0 {
// 		t.Error("Expected metadata before deletion")
// 	}

// 	// Delete the entity
// 	deleteReq := &pb.EntityId{Id: "service_delete_entity_1"}
// 	deleteResp, err := server.DeleteEntity(ctx, deleteReq)
// 	if err != nil {
// 		t.Fatalf("DeleteEntity() error = %v", err)
// 	}
// 	if deleteResp == nil {
// 		t.Fatal("DeleteEntity() returned nil response")
// 	}

// 	// Verify metadata was deleted (entity should still exist in Neo4j per TODO comments)
// 	readRespAfter, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() after delete error = %v", err)
// 	}

// 	// Metadata should be empty now
// 	if len(readRespAfter.Metadata) > 0 {
// 		t.Errorf("Expected metadata to be deleted, but found %d metadata fields", len(readRespAfter.Metadata))
// 	}

// 	log.Printf("Successfully deleted entity metadata")
// }

// // TODO: TestServiceDeleteNonExistentEntity tests deleting an entity that doesn't exist
// // 	FIXME: Once the delete functionality is fully implemented, this test should be added.

// // TestServiceReadEntities tests filtering entities through the service layer
// func TestServiceReadEntities(t *testing.T) {
// 	ctx := context.Background()

// 	// Create multiple entities of the same kind
// 	for i := 1; i <= 3; i++ {
// 		entity := &pb.Entity{
// 			Id: fmt.Sprintf("service_test_ministry_%d", i),
// 			Kind: &pb.Kind{
// 				Major: "Organization",
// 				Minor: "Ministry",
// 			},
// 			Name:    createNameValue("2025-03-18T00:00:00Z", fmt.Sprintf("Ministry %d", i)),
// 			Created: "2025-03-18T00:00:00Z",
// 		}

// 		_, err := server.CreateEntity(ctx, entity)
// 		if err != nil {
// 			t.Fatalf("CreateEntity(ministry %d) error = %v", i, err)
// 		}
// 	}

// 	// Filter entities by Kind
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Kind: &pb.Kind{
// 				Major: "Organization",
// 				Minor: "Ministry",
// 			},
// 		},
// 	}

// 	listResp, err := server.ReadEntities(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntities() error = %v", err)
// 	}
// 	if listResp == nil {
// 		t.Fatal("ReadEntities() returned nil response")
// 	}
// 	if len(listResp.Entities) < 3 {
// 		t.Errorf("ReadEntities() returned %d entities, want at least 3", len(listResp.Entities))
// 	}

// 	log.Printf("Successfully filtered entities: found %d entities", len(listResp.Entities))
// }

// // TestServiceReadEntityById tests filtering a single entity by ID
// func TestServiceReadEntityById(t *testing.T) {
// 	ctx := context.Background()

// 	// Create an entity
// 	entity := &pb.Entity{
// 		Id: "service_test_entity_9",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Frank"),
// 		Created: "2025-03-18T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}

// 	// Filter by ID
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Id: "service_test_entity_9",
// 		},
// 	}

// 	listResp, err := server.ReadEntities(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntities() error = %v", err)
// 	}
// 	if listResp == nil {
// 		t.Fatal("ReadEntities() returned nil response")
// 	}
// 	if len(listResp.Entities) != 1 {
// 		t.Errorf("ReadEntities() returned %d entities, want 1", len(listResp.Entities))
// 	}
// 	if len(listResp.Entities) > 0 && listResp.Entities[0].Id != "service_test_entity_9" {
// 		t.Errorf("ReadEntities() returned entity with ID %v, want service_test_entity_9", listResp.Entities[0].Id)
// 	}

// 	log.Printf("Successfully filtered entity by ID: %v", listResp.Entities[0].Id)
// }

// // TestServiceReadEntitiesByKindMajorOnly tests filtering entities by Kind.Major only
// func TestServiceReadEntitiesByKindMajorOnly(t *testing.T) {
// 	ctx := context.Background()

// 	// Create entities with same Major but different Minor
// 	entity1 := &pb.Entity{
// 		Id: "service_kind_filter_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Vehicle",
// 			Minor: "Car",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Tesla Model 3"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_kind_filter_entity_2",
// 		Kind: &pb.Kind{
// 			Major: "Vehicle",
// 			Minor: "Truck",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Ford F-150"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	// Filter by Kind.Major only (should return both)
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Kind: &pb.Kind{
// 				Major: "Vehicle",
// 				// Minor not specified
// 			},
// 		},
// 	}

// 	listResp, err := server.ReadEntities(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntities() error = %v", err)
// 	}
// 	if listResp == nil {
// 		t.Fatal("ReadEntities() returned nil response")
// 	}
// 	if len(listResp.Entities) < 2 {
// 		t.Errorf("ReadEntities() returned %d entities, want at least 2", len(listResp.Entities))
// 	}

// 	log.Printf("Successfully filtered entities by Kind.Major only: found %d entities", len(listResp.Entities))
// }

// // TestServiceReadEntitiesWithNoResults tests filtering with criteria that matches nothing
// func TestServiceReadEntitiesWithNoResults(t *testing.T) {
// 	ctx := context.Background()

// 	// Filter by a Kind that doesn't exist
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Kind: &pb.Kind{
// 				Major: "NonExistentKind",
// 				Minor: "AlsoNonExistent",
// 			},
// 		},
// 	}

// 	listResp, err := server.ReadEntities(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntities() error = %v (should succeed with empty list)", err)
// 	}
// 	if listResp == nil {
// 		t.Fatal("ReadEntities() returned nil response")
// 	}

// 	// Should return empty list, not error
// 	if len(listResp.Entities) > 0 {
// 		t.Errorf("Expected empty entity list, but got %d entities", len(listResp.Entities))
// 	}

// 	log.Printf("Successfully verified that no matches returns empty list")
// }

// // TestServiceReadEntitiesWithMissingEntity tests that missing entity parameter returns error
// func TestServiceReadEntitiesWithMissingEntity(t *testing.T) {
// 	ctx := context.Background()

// 	// Try to filter without providing entity
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: nil, // Missing entity
// 	}

// 	_, err := server.ReadEntities(ctx, readReq)
// 	if err == nil {
// 		t.Error("Expected error when entity is nil, but got none")
// 	} else {
// 		log.Printf("ReadEntities correctly failed with missing entity: %v", err)
// 	}

// 	log.Printf("Successfully verified that missing entity parameter fails")
// }

// // TestServiceReadEntitiesWithMissingIdAndKind tests that missing both ID and Kind.Major returns error
// func TestServiceReadEntitiesWithMissingIdAndKind(t *testing.T) {
// 	ctx := context.Background()

// 	// Try to filter without ID and without Kind.Major
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			// No ID
// 			Kind: &pb.Kind{
// 				// No Major
// 				Minor: "Employee",
// 			},
// 		},
// 	}

// 	_, err := server.ReadEntities(ctx, readReq)
// 	if err == nil {
// 		t.Error("Expected error when both ID and Kind.Major are missing, but got none")
// 	} else {
// 		log.Printf("ReadEntities correctly failed with missing ID and Kind.Major: %v", err)
// 	}

// 	log.Printf("Successfully verified that missing both ID and Kind.Major fails")
// }

// // TestServiceCreateEntityWithMetadata tests creating an entity with metadata
// func TestServiceCreateEntityWithMetadata(t *testing.T) {
// 	ctx := context.Background()

// 	metadata := make(map[string]*anypb.Any)
// 	metadata["department"] = &anypb.Any{
// 		TypeUrl: "type.googleapis.com/google.protobuf.StringValue",
// 		Value:   []byte("Engineering"),
// 	}

// 	entity := &pb.Entity{
// 		Id: "service_test_entity_10",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:     createNameValue("2025-03-18T00:00:00Z", "Grace"),
// 		Created:  "2025-03-18T00:00:00Z",
// 		Metadata: metadata,
// 	}

// 	resp, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}
// 	if resp == nil {
// 		t.Fatal("CreateEntity() returned nil response")
// 	}

// 	// Read back with metadata
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_test_entity_10"},
// 		Output: []string{"metadata"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}
// 	if readResp == nil {
// 		t.Fatal("ReadEntity() returned nil response")
// 	}
// 	if len(readResp.Metadata) == 0 {
// 		t.Error("ReadEntity() returned no metadata")
// 	}

// 	log.Printf("Successfully created and read entity with metadata: %v", readResp.Id)
// }

// // TestServiceUpdateEntityAddRelationship tests adding a relationship to an existing entity via UpdateEntity
// func TestServiceUpdateEntityAddRelationship(t *testing.T) {
// 	ctx := context.Background()

// 	// Create two entities first
// 	entity1 := &pb.Entity{
// 		Id: "service_test_entity_11",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Henry"),
// 		Created: "2025-03-18T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_test_entity_12",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Iris"),
// 		Created: "2025-03-18T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	// Now update entity1 to add a relationship to entity2
// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_test_entity_11",
// 		Entity: &pb.Entity{
// 			Id: "service_test_entity_11",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_test_rel_3": {
// 					Id:              "service_test_rel_3",
// 					Name:            "MANAGES",
// 					RelatedEntityId: "service_test_entity_12",
// 					StartTime:       "2025-04-01T00:00:00Z",
// 				},
// 			},
// 		},
// 	}

// 	updateResp, err := server.UpdateEntity(ctx, updateReq)
// 	if err != nil {
// 		t.Fatalf("UpdateEntity() error = %v", err)
// 	}
// 	if updateResp == nil {
// 		t.Fatal("UpdateEntity() returned nil response")
// 	}

// 	// Read back to verify relationship was added
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_test_entity_11"},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}
// 	if len(readResp.Relationships) == 0 {
// 		t.Error("ReadEntity() returned no relationships after update")
// 	}
// 	if _, exists := readResp.Relationships["service_test_rel_3"]; !exists {
// 		t.Error("Expected relationship 'service_test_rel_3' not found after update")
// 	}

// 	log.Printf("Successfully added relationship via UpdateEntity: %v", readResp.Id)
// }

// // TestServiceUpdateEntityModifyRelationship tests updating an existing relationship via UpdateEntity
// func TestServiceUpdateEntityModifyRelationship(t *testing.T) {
// 	ctx := context.Background()

// 	// Create two entities with a relationship
// 	entity1 := &pb.Entity{
// 		Id: "service_test_entity_13",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Jack"),
// 		Created: "2025-03-18T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_test_entity_14",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Kate"),
// 		Created: "2025-03-18T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_test_rel_4": {
// 				Id:              "service_test_rel_4",
// 				Name:            "WORKS_WITH",
// 				RelatedEntityId: "service_test_entity_13",
// 				StartTime:       "2025-03-18T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2 with relationship) error = %v", err)
// 	}

// 	// Update the relationship to add an EndTime date
// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_test_entity_14",
// 		Entity: &pb.Entity{
// 			Id: "service_test_entity_14",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_test_rel_4": {
// 					Id:      "service_test_rel_4",
// 					EndTime: "2025-12-31T00:00:00Z",
// 				},
// 			},
// 		},
// 	}

// 	updateResp, err := server.UpdateEntity(ctx, updateReq)
// 	if err != nil {
// 		t.Fatalf("UpdateEntity() error = %v", err)
// 	}
// 	if updateResp == nil {
// 		t.Fatal("UpdateEntity() returned nil response")
// 	}

// 	// Read back to verify relationship was updated
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Id: "service_test_entity_14",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_test_rel_4": {Id: "service_test_rel_4"},
// 			},
// 		},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}
// 	if len(readResp.Relationships) == 0 {
// 		t.Error("ReadEntity() returned no relationships")
// 	}

// 	rel, exists := readResp.Relationships["service_test_rel_4"]
// 	if !exists {
// 		t.Fatal("Expected relationship 'service_test_rel_4' not found")
// 	}
// 	if rel.EndTime != "2025-12-31T00:00:00Z" {
// 		t.Errorf("Relationship EndTime = %v, want 2025-12-31T00:00:00Z", rel.EndTime)
// 	}

// 	log.Printf("Successfully updated relationship via UpdateEntity: %v", readResp.Id)
// }

// // TestServiceUpdateEntityMultipleRelationships tests adding multiple relationships via UpdateEntity
// func TestServiceUpdateEntityMultipleRelationships(t *testing.T) {
// 	ctx := context.Background()

// 	// Create three entities
// 	entity1 := &pb.Entity{
// 		Id: "service_test_entity_15",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Manager",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Leo"),
// 		Created: "2025-03-18T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_test_entity_16",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Mia"),
// 		Created: "2025-03-18T00:00:00Z",
// 	}

// 	entity3 := &pb.Entity{
// 		Id: "service_test_entity_17",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Nina"),
// 		Created: "2025-03-18T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity3)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity3) error = %v", err)
// 	}

// 	// Update entity1 to add relationships to both entity2 and entity3
// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_test_entity_15",
// 		Entity: &pb.Entity{
// 			Id: "service_test_entity_15",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_test_rel_5": {
// 					Id:              "service_test_rel_5",
// 					Name:            "SUPERVISES",
// 					RelatedEntityId: "service_test_entity_16",
// 					StartTime:       "2025-04-01T00:00:00Z",
// 				},
// 				"service_test_rel_6": {
// 					Id:              "service_test_rel_6",
// 					Name:            "SUPERVISES",
// 					RelatedEntityId: "service_test_entity_17",
// 					StartTime:       "2025-04-01T00:00:00Z",
// 				},
// 			},
// 		},
// 	}

// 	updateResp, err := server.UpdateEntity(ctx, updateReq)
// 	if err != nil {
// 		t.Fatalf("UpdateEntity() error = %v", err)
// 	}
// 	if updateResp == nil {
// 		t.Fatal("UpdateEntity() returned nil response")
// 	}

// 	// Read back to verify both relationships were added
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_test_entity_15"},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}
// 	if len(readResp.Relationships) < 2 {
// 		t.Errorf("ReadEntity() returned %d relationships, want at least 2", len(readResp.Relationships))
// 	}

// 	// Verify both relationships exist
// 	if _, exists := readResp.Relationships["service_test_rel_5"]; !exists {
// 		t.Error("Expected relationship 'service_test_rel_5' not found")
// 	}
// 	if _, exists := readResp.Relationships["service_test_rel_6"]; !exists {
// 		t.Error("Expected relationship 'service_test_rel_6' not found")
// 	}

// 	log.Printf("Successfully added multiple relationships via UpdateEntity: %v", readResp.Id)
// }

// // TestServiceCreateEntityWithMultipleRelationships tests creating an entity with multiple relationships at once
// func TestServiceCreateEntityWithMultipleRelationships(t *testing.T) {
// 	ctx := context.Background()

// 	// Create target entities first
// 	entity1 := &pb.Entity{
// 		Id: "service_test_entity_18",
// 		Kind: &pb.Kind{
// 			Major: "Organization",
// 			Minor: "Department",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Engineering"),
// 		Created: "2025-03-18T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_test_entity_19",
// 		Kind: &pb.Kind{
// 			Major: "Organization",
// 			Minor: "Department",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Sales"),
// 		Created: "2025-03-18T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	// Create a person entity with relationships to both departments
// 	entity3 := &pb.Entity{
// 		Id: "service_test_entity_20",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-03-18T00:00:00Z", "Oscar"),
// 		Created: "2025-03-18T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_test_rel_7": {
// 				Id:              "service_test_rel_7",
// 				Name:            "MEMBER_OF",
// 				RelatedEntityId: "service_test_entity_18",
// 				StartTime:       "2025-01-01T00:00:00Z",
// 				EndTime:         "2025-06-30T00:00:00Z",
// 			},
// 			"service_test_rel_8": {
// 				Id:              "service_test_rel_8",
// 				Name:            "MEMBER_OF",
// 				RelatedEntityId: "service_test_entity_19",
// 				StartTime:       "2025-07-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	resp, err := server.CreateEntity(ctx, entity3)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity with multiple relationships) error = %v", err)
// 	}
// 	if resp == nil {
// 		t.Fatal("CreateEntity() returned nil response")
// 	}

// 	// Read back to verify both relationships were created
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_test_entity_20"},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}
// 	if len(readResp.Relationships) < 2 {
// 		t.Errorf("ReadEntity() returned %d relationships, want at least 2", len(readResp.Relationships))
// 	}

// 	// Verify both relationships exist
// 	rel1, exists1 := readResp.Relationships["service_test_rel_7"]
// 	if !exists1 {
// 		t.Error("Expected relationship 'service_test_rel_7' not found")
// 	} else if rel1.EndTime != "2025-06-30T00:00:00Z" {
// 		t.Errorf("Relationship 'service_test_rel_7' EndTime = %v, want 2025-06-30T00:00:00Z", rel1.EndTime)
// 	}

// 	if _, exists2 := readResp.Relationships["service_test_rel_8"]; !exists2 {
// 		t.Error("Expected relationship 'service_test_rel_8' not found")
// 	}

// 	log.Printf("Successfully created entity with multiple relationships: %v", readResp.Id)
// }

// // TestServiceCreateRelationshipWithDuplicateId tests that creating a relationship with a duplicate ID fails
// func TestServiceCreateRelationshipWithDuplicateId(t *testing.T) {
// 	ctx := context.Background()

// 	// Create two entities
// 	entity1 := &pb.Entity{
// 		Id: "service_dup_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Paul"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_dup_entity_2",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Quinn"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	// Create a relationship with a specific ID
// 	entity3 := &pb.Entity{
// 		Id: "service_dup_entity_3",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Rachel"),
// 		Created: "2025-04-01T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_duplicate_rel_id": {
// 				Id:              "service_duplicate_rel_id",
// 				Name:            "WORKS_WITH",
// 				RelatedEntityId: "service_dup_entity_2",
// 				StartTime:       "2025-04-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err = server.CreateEntity(ctx, entity3)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity3 with relationship) error = %v", err)
// 	}

// 	// Read the original relationship to verify its properties
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Id: "service_dup_entity_3",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_duplicate_rel_id": {Id: "service_duplicate_rel_id"},
// 			},
// 		},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	originalRel := readResp.Relationships["service_duplicate_rel_id"]
// 	originalStartTime := originalRel.StartTime

// 	// Attempt to create another entity with a relationship with the SAME ID (should fail)
// 	entity4 := &pb.Entity{
// 		Id: "service_dup_entity_4",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Sam"),
// 		Created: "2025-04-01T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_duplicate_rel_id": {
// 				Id:              "service_duplicate_rel_id", // Same ID!
// 				Name:            "MANAGES",                  // Different type
// 				RelatedEntityId: "service_dup_entity_2",
// 				StartTime:       "2025-05-01T00:00:00Z", // Different start time
// 			},
// 		},
// 	}

// 	_, err = server.CreateEntity(ctx, entity4)
// 	if err == nil {
// 		t.Error("Expected error when creating entity with duplicate relationship ID, but got none")
// 	} else {
// 		log.Printf("Duplicate relationship creation failed as expected: %v", err)
// 	}

// 	// Verify the original relationship was NOT modified
// 	verifyResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error after duplicate attempt = %v", err)
// 	}

// 	verifyRel := verifyResp.Relationships["service_duplicate_rel_id"]
// 	if verifyRel.Name != "WORKS_WITH" {
// 		t.Errorf("Relationship type changed from WORKS_WITH to %v", verifyRel.Name)
// 	}
// 	if verifyRel.StartTime != originalStartTime {
// 		t.Errorf("Relationship StartTime changed from %v to %v", originalStartTime, verifyRel.StartTime)
// 	}

// 	log.Printf("Successfully verified that duplicate relationship IDs are rejected")
// }

// // TestServiceUpdateNonExistentRelationship tests that updating a non-existent relationship creates it
// func TestServiceUpdateNonExistentRelationship(t *testing.T) {
// 	ctx := context.Background()

// 	// Create two entities
// 	entity1 := &pb.Entity{
// 		Id: "service_nonexistent_test_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Tom"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_nonexistent_test_entity_2",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Uma"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	// Count relationships before update (should be 0)
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_nonexistent_test_entity_1"},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}
// 	countBefore := len(readResp.Relationships)

// 	// Try to update a relationship that doesn't exist (should create it)
// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_nonexistent_test_entity_1",
// 		Entity: &pb.Entity{
// 			Id: "service_nonexistent_test_entity_1",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_nonexistent_rel_id": {
// 					Id:              "service_nonexistent_rel_id",
// 					Name:            "MENTORS",
// 					RelatedEntityId: "service_nonexistent_test_entity_2",
// 					StartTime:       "2025-04-01T00:00:00Z",
// 					EndTime:         "2025-12-31T00:00:00Z",
// 				},
// 			},
// 		},
// 	}

// 	_, err = server.UpdateEntity(ctx, updateReq)
// 	if err != nil {
// 		t.Fatalf("UpdateEntity() error = %v (expected to succeed and create relationship)", err)
// 	}

// 	// Verify the relationship was created
// 	readResp, err = server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() after update error = %v", err)
// 	}

// 	countAfter := len(readResp.Relationships)
// 	if countAfter != countBefore+1 {
// 		t.Errorf("Expected 1 new relationship, but count changed from %d to %d", countBefore, countAfter)
// 	}

// 	// Verify the new relationship exists with correct properties
// 	rel := readResp.Relationships["service_nonexistent_rel_id"]
// 	if rel == nil {
// 		t.Fatal("Relationship 'service_nonexistent_rel_id' not found after update")
// 	}
// 	if rel.Name != "MENTORS" {
// 		t.Errorf("Relationship Name = %v, want MENTORS", rel.Name)
// 	}
// 	if rel.RelatedEntityId != "service_nonexistent_test_entity_2" {
// 		t.Errorf("Relationship RelatedEntityId = %v, want service_nonexistent_test_entity_2", rel.RelatedEntityId)
// 	}
// 	if rel.EndTime != "2025-12-31T00:00:00Z" {
// 		t.Errorf("Relationship EndTime = %v, want 2025-12-31T00:00:00Z", rel.EndTime)
// 	}

// 	log.Printf("Successfully verified that updating non-existent relationship creates it")
// }

// // TestServiceUpdateRelationshipValidFields tests updating valid fields (StartTime/EndTime) on relationships
// func TestServiceUpdateRelationshipValidFields(t *testing.T) {
// 	ctx := context.Background()

// 	// Create two entities with a relationship
// 	entity1 := &pb.Entity{
// 		Id: "service_update_valid_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Uma"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_update_valid_entity_2",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Victor"),
// 		Created: "2025-04-01T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_update_valid_rel": {
// 				Id:              "service_update_valid_rel",
// 				Name:            "REPORTS_TO",
// 				RelatedEntityId: "service_update_valid_entity_1",
// 				StartTime:       "2025-04-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	// Count relationships before update
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_update_valid_entity_2"},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}
// 	countBefore := len(readResp.Relationships)

// 	// Update only StartTime (Created date)
// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_update_valid_entity_2",
// 		Entity: &pb.Entity{
// 			Id: "service_update_valid_entity_2",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_update_valid_rel": {
// 					Id:        "service_update_valid_rel",
// 					StartTime: "2025-03-15T00:00:00Z",
// 				},
// 			},
// 		},
// 	}

// 	_, err = server.UpdateEntity(ctx, updateReq)
// 	if err != nil {
// 		t.Fatalf("UpdateEntity(update StartTime) error = %v", err)
// 	}

// 	// Verify the relationship was updated
// 	readResp, err = server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() after StartTime update error = %v", err)
// 	}

// 	if len(readResp.Relationships) != countBefore {
// 		t.Errorf("Relationship count changed from %d to %d after update", countBefore, len(readResp.Relationships))
// 	}

// 	rel := readResp.Relationships["service_update_valid_rel"]
// 	if rel.StartTime != "2025-03-15T00:00:00Z" {
// 		t.Errorf("Relationship StartTime = %v, want 2025-03-15T00:00:00Z", rel.StartTime)
// 	}

// 	// Update only EndTime (Terminated date)
// 	updateReq2 := &pb.UpdateEntityRequest{
// 		Id: "service_update_valid_entity_2",
// 		Entity: &pb.Entity{
// 			Id: "service_update_valid_entity_2",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_update_valid_rel": {
// 					Id:      "service_update_valid_rel",
// 					EndTime: "2025-12-31T00:00:00Z",
// 				},
// 			},
// 		},
// 	}

// 	_, err = server.UpdateEntity(ctx, updateReq2)
// 	if err != nil {
// 		t.Fatalf("UpdateEntity(update EndTime) error = %v", err)
// 	}

// 	// Verify EndTime was updated
// 	readResp, err = server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() after EndTime update error = %v", err)
// 	}

// 	rel = readResp.Relationships["service_update_valid_rel"]
// 	if rel.EndTime != "2025-12-31T00:00:00Z" {
// 		t.Errorf("Relationship EndTime = %v, want 2025-12-31T00:00:00Z", rel.EndTime)
// 	}

// 	// Verify Name and RelatedEntityId haven't changed
// 	if rel.Name != "REPORTS_TO" {
// 		t.Errorf("Relationship Name changed to %v, want REPORTS_TO", rel.Name)
// 	}
// 	if rel.RelatedEntityId != "service_update_valid_entity_1" {
// 		t.Errorf("Relationship RelatedEntityId changed to %v, want service_update_valid_entity_1", rel.RelatedEntityId)
// 	}

// 	// Verify no new relationships were created
// 	if len(readResp.Relationships) != countBefore {
// 		t.Errorf("Relationship count changed from %d to %d after updates", countBefore, len(readResp.Relationships))
// 	}

// 	log.Printf("Successfully updated relationship with valid fields (StartTime/EndTime)")
// }

// // TestServiceUpdateRelationshipNoNewCreations tests that updating relationships doesn't create new ones
// func TestServiceUpdateRelationshipNoNewCreations(t *testing.T) {
// 	ctx := context.Background()

// 	// Create entity with relationships
// 	entity1 := &pb.Entity{
// 		Id: "service_no_new_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Wendy"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_no_new_entity_2",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Xander"),
// 		Created: "2025-04-01T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_no_new_rel_1": {
// 				Id:              "service_no_new_rel_1",
// 				Name:            "COLLABORATES_WITH",
// 				RelatedEntityId: "service_no_new_entity_1",
// 				StartTime:       "2025-04-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	// Count relationships before update
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_no_new_entity_2"},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}
// 	countBefore := len(readResp.Relationships)

// 	// Perform multiple updates
// 	for i := 0; i < 3; i++ {
// 		updateReq := &pb.UpdateEntityRequest{
// 			Id: "service_no_new_entity_2",
// 			Entity: &pb.Entity{
// 				Id: "service_no_new_entity_2",
// 				Relationships: map[string]*pb.Relationship{
// 					"service_no_new_rel_1": {
// 						Id:      "service_no_new_rel_1",
// 						EndTime: fmt.Sprintf("2025-12-%02dT00:00:00Z", 10+i),
// 					},
// 				},
// 			},
// 		}

// 		_, err = server.UpdateEntity(ctx, updateReq)
// 		if err != nil {
// 			t.Fatalf("UpdateEntity(iteration %d) error = %v", i, err)
// 		}
// 	}

// 	// Verify no new relationships were created
// 	readResp, err = server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() after updates error = %v", err)
// 	}

// 	countAfter := len(readResp.Relationships)
// 	if countAfter != countBefore {
// 		t.Errorf("Relationship count changed from %d to %d after multiple updates", countBefore, countAfter)
// 	}

// 	// Verify the relationship still exists with the latest update
// 	rel := readResp.Relationships["service_no_new_rel_1"]
// 	if rel == nil {
// 		t.Fatal("Relationship 'service_no_new_rel_1' not found after updates")
// 	}
// 	if rel.EndTime != "2025-12-12T00:00:00Z" {
// 		t.Errorf("Relationship EndTime = %v, want 2025-12-12T00:00:00Z (last update)", rel.EndTime)
// 	}

// 	log.Printf("Successfully verified that updating relationships doesn't create new ones")
// }

// // TestServiceUpdateRelationshipBothFields tests updating both StartTime and EndTime together
// func TestServiceUpdateRelationshipBothFields(t *testing.T) {
// 	ctx := context.Background()

// 	// Create two entities with a relationship
// 	entity1 := &pb.Entity{
// 		Id: "service_both_fields_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Yara"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_both_fields_entity_2",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Zane"),
// 		Created: "2025-04-01T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_both_fields_rel": {
// 				Id:              "service_both_fields_rel",
// 				Name:            "WORKS_WITH",
// 				RelatedEntityId: "service_both_fields_entity_1",
// 				StartTime:       "2025-04-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	// Count relationships before update
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_both_fields_entity_2"},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}
// 	countBefore := len(readResp.Relationships)

// 	// Update both StartTime and EndTime
// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_both_fields_entity_2",
// 		Entity: &pb.Entity{
// 			Id: "service_both_fields_entity_2",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_both_fields_rel": {
// 					Id:        "service_both_fields_rel",
// 					StartTime: "2025-02-01T00:00:00Z",
// 					EndTime:   "2025-11-30T00:00:00Z",
// 				},
// 			},
// 		},
// 	}

// 	_, err = server.UpdateEntity(ctx, updateReq)
// 	if err != nil {
// 		t.Fatalf("UpdateEntity(update both fields) error = %v", err)
// 	}

// 	// Verify both fields were updated
// 	readResp, err = server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() after both fields update error = %v", err)
// 	}

// 	rel := readResp.Relationships["service_both_fields_rel"]
// 	if rel.StartTime != "2025-02-01T00:00:00Z" {
// 		t.Errorf("Relationship StartTime = %v, want 2025-02-01T00:00:00Z", rel.StartTime)
// 	}
// 	if rel.EndTime != "2025-11-30T00:00:00Z" {
// 		t.Errorf("Relationship EndTime = %v, want 2025-11-30T00:00:00Z", rel.EndTime)
// 	}

// 	// Verify Name and RelatedEntityId haven't changed
// 	if rel.Name != "WORKS_WITH" {
// 		t.Errorf("Relationship Name changed to %v, want WORKS_WITH", rel.Name)
// 	}
// 	if rel.RelatedEntityId != "service_both_fields_entity_1" {
// 		t.Errorf("Relationship RelatedEntityId changed to %v, want service_both_fields_entity_1", rel.RelatedEntityId)
// 	}

// 	// Verify no new relationships were created
// 	if len(readResp.Relationships) != countBefore {
// 		t.Errorf("Relationship count changed from %d to %d after update", countBefore, len(readResp.Relationships))
// 	}

// 	log.Printf("Successfully updated both StartTime and EndTime fields")
// }

// // TestServiceUpdateRelationshipInvalidFields tests that updating invalid fields (Name, RelatedEntityId) fails
// func TestServiceUpdateRelationshipInvalidFields(t *testing.T) {
// 	ctx := context.Background()

// 	// Create two entities with a relationship
// 	entity1 := &pb.Entity{
// 		Id: "service_invalid_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Alpha"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_invalid_entity_2",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Beta"),
// 		Created: "2025-04-01T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_invalid_rel": {
// 				Id:              "service_invalid_rel",
// 				Name:            "MANAGES",
// 				RelatedEntityId: "service_invalid_entity_1",
// 				StartTime:       "2025-04-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	entity3 := &pb.Entity{
// 		Id: "service_invalid_entity_3",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Gamma"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity3)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity3) error = %v", err)
// 	}

// 	// Store original relationship properties
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Id: "service_invalid_entity_2",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_invalid_rel": {Id: "service_invalid_rel"},
// 			},
// 		},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() before update error = %v", err)
// 	}
// 	originalRel := readResp.Relationships["service_invalid_rel"]
// 	originalName := originalRel.Name
// 	originalRelatedEntityId := originalRel.RelatedEntityId

// 	// Try to update the relationship with invalid fields (Name and RelatedEntityId)
// 	// This should FAIL because only StartTime/EndTime are allowed
// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_invalid_entity_2",
// 		Entity: &pb.Entity{
// 			Id: "service_invalid_entity_2",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_invalid_rel": {
// 					Id:              "service_invalid_rel",
// 					Name:            "SUPERVISES",               // Invalid: try to change relationship type
// 					RelatedEntityId: "service_invalid_entity_3", // Invalid: try to change target entity
// 				},
// 			},
// 		},
// 	}

// 	_, err = server.UpdateEntity(ctx, updateReq)
// 	if err == nil {
// 		t.Error("Expected error when trying to update invalid fields (Name, RelatedEntityId), but got none")
// 	} else {
// 		log.Printf("UpdateEntity failed when trying to update invalid fields (expected): %v", err)
// 	}

// 	// Read back to verify relationship hasn't changed
// 	readResp, err = server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() after failed update error = %v", err)
// 	}

// 	rel := readResp.Relationships["service_invalid_rel"]
// 	if rel == nil {
// 		t.Fatal("Relationship 'service_invalid_rel' not found after failed update")
// 	}

// 	// Verify Name (relationship type) hasn't changed
// 	if rel.Name != originalName {
// 		t.Errorf("Relationship Name changed from %v to %v (should remain unchanged after failed update)", originalName, rel.Name)
// 	}

// 	// Verify RelatedEntityId hasn't changed
// 	if rel.RelatedEntityId != originalRelatedEntityId {
// 		t.Errorf("Relationship RelatedEntityId changed from %v to %v (should remain unchanged after failed update)", originalRelatedEntityId, rel.RelatedEntityId)
// 	}

// 	log.Printf("Successfully verified that updating invalid fields (Name/RelatedEntityId) fails")
// }

// // TestServiceCreateEntityWithIncompleteRelationship tests that creating an entity with incomplete relationship fails
// func TestServiceCreateEntityWithIncompleteRelationship(t *testing.T) {
// 	ctx := context.Background()

// 	// Create a target entity first
// 	entity1 := &pb.Entity{
// 		Id: "service_incomplete_create_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Delta"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(target entity) error = %v", err)
// 	}

// 	// Try to create an entity with incomplete relationship (missing Name field)
// 	entity2 := &pb.Entity{
// 		Id: "service_incomplete_create_entity_2",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Epsilon"),
// 		Created: "2025-04-01T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_incomplete_create_rel_1": {
// 				Id: "service_incomplete_create_rel_1",
// 				// Name is missing (required field)
// 				RelatedEntityId: "service_incomplete_create_entity_1",
// 				StartTime:       "2025-04-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err == nil {
// 		t.Error("Expected error when creating entity with incomplete relationship (missing Name), but got none")
// 	} else {
// 		log.Printf("CreateEntity failed with incomplete relationship as expected: %v", err)
// 	}

// 	// Try to create an entity with incomplete relationship (missing RelatedEntityId)
// 	entity3 := &pb.Entity{
// 		Id: "service_incomplete_create_entity_3",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Zeta"),
// 		Created: "2025-04-01T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_incomplete_create_rel_2": {
// 				Id:   "service_incomplete_create_rel_2",
// 				Name: "REPORTS_TO",
// 				// RelatedEntityId is missing (required field)
// 				StartTime: "2025-04-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err = server.CreateEntity(ctx, entity3)
// 	if err == nil {
// 		t.Error("Expected error when creating entity with incomplete relationship (missing RelatedEntityId), but got none")
// 	} else {
// 		log.Printf("CreateEntity failed with incomplete relationship as expected: %v", err)
// 	}

// 	// Try to create an entity with incomplete relationship (missing StartTime)
// 	entity4 := &pb.Entity{
// 		Id: "service_incomplete_create_entity_4",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Eta"),
// 		Created: "2025-04-01T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_incomplete_create_rel_3": {
// 				Id:              "service_incomplete_create_rel_3",
// 				Name:            "MANAGES",
// 				RelatedEntityId: "service_incomplete_create_entity_1",
// 				// StartTime is missing (required field)
// 			},
// 		},
// 	}

// 	_, err = server.CreateEntity(ctx, entity4)
// 	if err == nil {
// 		t.Error("Expected error when creating entity with incomplete relationship (missing StartTime), but got none")
// 	} else {
// 		log.Printf("CreateEntity failed with incomplete relationship as expected: %v", err)
// 	}

// 	log.Printf("Successfully verified that creating entities with incomplete relationships fails")
// }

// // TestServiceUpdateEntityAddIncompleteRelationship tests that adding incomplete relationship via update fails
// func TestServiceUpdateEntityAddIncompleteRelationship(t *testing.T) {
// 	ctx := context.Background()

// 	// Create two entities
// 	entity1 := &pb.Entity{
// 		Id: "service_incomplete_update_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Theta"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_incomplete_update_entity_2",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Iota"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	// Try to update entity to add incomplete relationship (missing Name)
// 	updateReq1 := &pb.UpdateEntityRequest{
// 		Id: "service_incomplete_update_entity_1",
// 		Entity: &pb.Entity{
// 			Id: "service_incomplete_update_entity_1",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_incomplete_update_rel_1": {
// 					Id: "service_incomplete_update_rel_1",
// 					// Name is missing (required field)
// 					RelatedEntityId: "service_incomplete_update_entity_2",
// 					StartTime:       "2025-04-01T00:00:00Z",
// 				},
// 			},
// 		},
// 	}

// 	_, err = server.UpdateEntity(ctx, updateReq1)
// 	if err == nil {
// 		t.Error("Expected error when updating entity with incomplete relationship (missing Name), but got none")
// 	} else {
// 		log.Printf("UpdateEntity failed with incomplete relationship as expected: %v", err)
// 	}

// 	// Try to update entity to add incomplete relationship (missing RelatedEntityId)
// 	updateReq2 := &pb.UpdateEntityRequest{
// 		Id: "service_incomplete_update_entity_1",
// 		Entity: &pb.Entity{
// 			Id: "service_incomplete_update_entity_1",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_incomplete_update_rel_2": {
// 					Id:   "service_incomplete_update_rel_2",
// 					Name: "COLLABORATES_WITH",
// 					// RelatedEntityId is missing (required field)
// 					StartTime: "2025-04-01T00:00:00Z",
// 				},
// 			},
// 		},
// 	}

// 	_, err = server.UpdateEntity(ctx, updateReq2)
// 	if err == nil {
// 		t.Error("Expected error when updating entity with incomplete relationship (missing RelatedEntityId), but got none")
// 	} else {
// 		log.Printf("UpdateEntity failed with incomplete relationship as expected: %v", err)
// 	}

// 	// Try to update entity to add incomplete relationship (missing StartTime)
// 	updateReq3 := &pb.UpdateEntityRequest{
// 		Id: "service_incomplete_update_entity_1",
// 		Entity: &pb.Entity{
// 			Id: "service_incomplete_update_entity_1",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_incomplete_update_rel_3": {
// 					Id:              "service_incomplete_update_rel_3",
// 					Name:            "SUPERVISES",
// 					RelatedEntityId: "service_incomplete_update_entity_2",
// 					// StartTime is missing (required field)
// 				},
// 			},
// 		},
// 	}

// 	_, err = server.UpdateEntity(ctx, updateReq3)
// 	if err == nil {
// 		t.Error("Expected error when updating entity with incomplete relationship (missing StartTime), but got none")
// 	} else {
// 		log.Printf("UpdateEntity failed with incomplete relationship as expected: %v", err)
// 	}

// 	// Verify no relationships were created
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_incomplete_update_entity_1"},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	if len(readResp.Relationships) > 0 {
// 		t.Errorf("Expected no relationships to be created, but found %d", len(readResp.Relationships))
// 	}

// 	log.Printf("Successfully verified that adding incomplete relationships via update fails")
// }

// // TestServiceUpdateEntityCoreAttributesOnly tests updating only core attributes (Name, Terminated)
// func TestServiceUpdateEntityCoreAttributesOnly(t *testing.T) {
// 	ctx := context.Background()

// 	// Create an entity
// 	entity := &pb.Entity{
// 		Id: "service_update_core_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Original Name"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}

// 	// Update only core attributes (Name and Terminated)
// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_update_core_entity_1",
// 		Entity: &pb.Entity{
// 			Name:       createNameValue("2025-04-01T00:00:00Z", "Updated Name"),
// 			Terminated: "2025-12-31T00:00:00Z",
// 		},
// 	}

// 	updateResp, err := server.UpdateEntity(ctx, updateReq)
// 	if err != nil {
// 		t.Fatalf("UpdateEntity() error = %v", err)
// 	}
// 	if updateResp == nil {
// 		t.Fatal("UpdateEntity() returned nil response")
// 	}

// 	// Verify the updates
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_update_core_entity_1"},
// 		Output: []string{},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	// Unpack and verify Name
// 	var stringValue wrapperspb.StringValue
// 	err = readResp.Name.GetValue().UnmarshalTo(&stringValue)
// 	if err != nil {
// 		t.Fatalf("Error unpacking Name value: %v", err)
// 	}
// 	if stringValue.Value != "Updated Name" {
// 		t.Errorf("Name = %v, want 'Updated Name'", stringValue.Value)
// 	}

// 	// Verify Terminated
// 	if readResp.Terminated != "2025-12-31T00:00:00Z" {
// 		t.Errorf("Terminated = %v, want '2025-12-31T00:00:00Z'", readResp.Terminated)
// 	}

// 	// Verify Kind hasn't changed
// 	if readResp.Kind.Major != "Person" || readResp.Kind.Minor != "Employee" {
// 		t.Errorf("Kind changed to %v/%v, should remain Person/Employee", readResp.Kind.Major, readResp.Kind.Minor)
// 	}

// 	log.Printf("Successfully updated core attributes only")
// }

// // TestServiceUpdateEntityKindNotAllowed tests that updating Kind fails
// func TestServiceUpdateEntityKindNotAllowed(t *testing.T) {
// 	ctx := context.Background()

// 	// Create an entity
// 	entity := &pb.Entity{
// 		Id: "service_update_kind_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Test User"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}

// 	// Try to update Kind.Major
// 	updateReq1 := &pb.UpdateEntityRequest{
// 		Id: "service_update_kind_entity_1",
// 		Entity: &pb.Entity{
// 			Id: "service_update_kind_entity_1",
// 			Kind: &pb.Kind{
// 				Major: "Organization", // Try to change Major
// 			},
// 		},
// 	}

// 	_, err = server.UpdateEntity(ctx, updateReq1)
// 	if err == nil {
// 		t.Error("Expected error when trying to update Kind.Major, but got none")
// 	} else {
// 		log.Printf("UpdateEntity correctly rejected Kind.Major update: %v", err)
// 	}

// 	// Try to update Kind.Minor
// 	updateReq2 := &pb.UpdateEntityRequest{
// 		Id: "service_update_kind_entity_1",
// 		Entity: &pb.Entity{
// 			Id: "service_update_kind_entity_1",
// 			Kind: &pb.Kind{
// 				Minor: "Manager", // Try to change Minor
// 			},
// 		},
// 	}

// 	_, err = server.UpdateEntity(ctx, updateReq2)
// 	if err == nil {
// 		t.Error("Expected error when trying to update Kind.Minor, but got none")
// 	} else {
// 		log.Printf("UpdateEntity correctly rejected Kind.Minor update: %v", err)
// 	}

// 	// Try to update both
// 	updateReq3 := &pb.UpdateEntityRequest{
// 		Id: "service_update_kind_entity_1",
// 		Entity: &pb.Entity{
// 			Id: "service_update_kind_entity_1",
// 			Kind: &pb.Kind{
// 				Major: "Organization",
// 				Minor: "Department",
// 			},
// 		},
// 	}

// 	_, err = server.UpdateEntity(ctx, updateReq3)
// 	if err == nil {
// 		t.Error("Expected error when trying to update both Kind.Major and Kind.Minor, but got none")
// 	} else {
// 		log.Printf("UpdateEntity correctly rejected Kind update: %v", err)
// 	}

// 	// Verify Kind hasn't changed
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_update_kind_entity_1"},
// 		Output: []string{},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	if readResp.Kind.Major != "Person" || readResp.Kind.Minor != "Employee" {
// 		t.Errorf("Kind was modified to %v/%v, should remain Person/Employee", readResp.Kind.Major, readResp.Kind.Minor)
// 	}

// 	log.Printf("Successfully verified that Kind updates are rejected")
// }

// // TestServiceUpdateEntityCoreAttributesAndRelationships tests updating both core attributes and relationships
// func TestServiceUpdateEntityCoreAttributesAndRelationships(t *testing.T) {
// 	ctx := context.Background()

// 	// Create two entities
// 	entity1 := &pb.Entity{
// 		Id: "service_update_both_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Alice"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_update_both_entity_2",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Manager",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Bob"),
// 		Created: "2025-04-01T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_update_both_rel_1": {
// 				Id:              "service_update_both_rel_1",
// 				Name:            "REPORTS_TO",
// 				RelatedEntityId: "service_update_both_entity_1",
// 				StartTime:       "2025-04-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	// Update both core attributes and relationships successfully
// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_update_both_entity_2",
// 		Entity: &pb.Entity{
// 			Id:         "service_update_both_entity_2",
// 			Name:       createNameValue("2025-04-01T00:00:00Z", "Bob Updated"),
// 			Terminated: "2025-12-31T00:00:00Z",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_update_both_rel_1": {
// 					Id:      "service_update_both_rel_1",
// 					EndTime: "2025-12-31T00:00:00Z",
// 				},
// 			},
// 		},
// 	}

// 	updateResp, err := server.UpdateEntity(ctx, updateReq)
// 	if err != nil {
// 		t.Fatalf("UpdateEntity() error = %v", err)
// 	}
// 	if updateResp == nil {
// 		t.Fatal("UpdateEntity() returned nil response")
// 	}

// 	// Verify both core attributes and relationships were updated
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_update_both_entity_2"},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	// Verify core attributes
// 	var stringValue wrapperspb.StringValue
// 	err = readResp.Name.GetValue().UnmarshalTo(&stringValue)
// 	if err != nil {
// 		t.Fatalf("Error unpacking Name value: %v", err)
// 	}
// 	if stringValue.Value != "Bob Updated" {
// 		t.Errorf("Name = %v, want 'Bob Updated'", stringValue.Value)
// 	}
// 	if readResp.Terminated != "2025-12-31T00:00:00Z" {
// 		t.Errorf("Terminated = %v, want '2025-12-31T00:00:00Z'", readResp.Terminated)
// 	}

// 	// Verify relationship was updated
// 	rel := readResp.Relationships["service_update_both_rel_1"]
// 	if rel == nil {
// 		t.Fatal("Relationship 'service_update_both_rel_1' not found")
// 	}
// 	if rel.EndTime != "2025-12-31T00:00:00Z" {
// 		t.Errorf("Relationship EndTime = %v, want '2025-12-31T00:00:00Z'", rel.EndTime)
// 	}

// 	log.Printf("Successfully updated both core attributes and relationships")
// }

// // TestServiceUpdateEntityCoreAttributesSuccessRelationshipsFail tests when core attributes succeed but relationships fail
// func TestServiceUpdateEntityCoreAttributesSuccessRelationshipsFail(t *testing.T) {
// 	ctx := context.Background()

// 	// Create an entity
// 	entity := &pb.Entity{
// 		Id: "service_update_partial_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Charlie"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}

// 	// Try to update core attributes with invalid relationship (missing required fields)
// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_update_partial_entity_1",
// 		Entity: &pb.Entity{
// 			Id:         "service_update_partial_entity_1",
// 			Name:       createNameValue("2025-04-01T00:00:00Z", "Charlie Updated"),
// 			Terminated: "2025-12-31T00:00:00Z",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_update_partial_rel_1": {
// 					Id: "service_update_partial_rel_1",
// 					// Missing Name and RelatedEntityId (required fields)
// 					StartTime: "2025-04-01T00:00:00Z",
// 				},
// 			},
// 		},
// 	}

// 	_, err = server.UpdateEntity(ctx, updateReq)
// 	if err == nil {
// 		t.Error("Expected error when relationships update fails, but got none")
// 	} else {
// 		log.Printf("UpdateEntity correctly failed when relationship is invalid: %v", err)
// 	}

// 	// Verify that core attributes WERE updated successfully (no transaction rollback)
// 	// Since there's no transaction wrapping, core attributes succeed before relationships fail
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_update_partial_entity_1"},
// 		Output: []string{},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	// Verify Name WAS updated (core attributes succeed before relationship failure)
// 	var stringValue wrapperspb.StringValue
// 	err = readResp.Name.GetValue().UnmarshalTo(&stringValue)
// 	if err != nil {
// 		t.Fatalf("Error unpacking Name value: %v", err)
// 	}
// 	if stringValue.Value != "Charlie Updated" {
// 		t.Errorf("Name = %v, want 'Charlie Updated' (core attributes should be updated despite relationship failure)", stringValue.Value)
// 	}

// 	// Verify Terminated was also updated
// 	if readResp.Terminated != "2025-12-31T00:00:00Z" {
// 		t.Errorf("Terminated = %v, want '2025-12-31T00:00:00Z' (core attributes should be updated despite relationship failure)", readResp.Terminated)
// 	}

// 	log.Printf("Successfully verified that core attributes are updated even when relationships fail (no transaction rollback)")
// }

// // TestServiceUpdateEntityRelationshipsOnlyNoCore tests updating only relationships without core attributes
// func TestServiceUpdateEntityRelationshipsOnlyNoCore(t *testing.T) {
// 	ctx := context.Background()

// 	// Create two entities
// 	entity1 := &pb.Entity{
// 		Id: "service_update_rel_only_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "David"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	entity2 := &pb.Entity{
// 		Id: "service_update_rel_only_entity_2",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Manager",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Eve"),
// 		Created: "2025-04-01T00:00:00Z",
// 		Relationships: map[string]*pb.Relationship{
// 			"service_update_rel_only_rel_1": {
// 				Id:              "service_update_rel_only_rel_1",
// 				Name:            "WORKS_WITH",
// 				RelatedEntityId: "service_update_rel_only_entity_1",
// 				StartTime:       "2025-04-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	// Update only relationships, no core attributes
// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_update_rel_only_entity_2",
// 		Entity: &pb.Entity{
// 			Id: "service_update_rel_only_entity_2",
// 			Relationships: map[string]*pb.Relationship{
// 				"service_update_rel_only_rel_1": {
// 					Id:      "service_update_rel_only_rel_1",
// 					EndTime: "2025-06-30T00:00:00Z",
// 				},
// 			},
// 		},
// 	}

// 	updateResp, err := server.UpdateEntity(ctx, updateReq)
// 	if err != nil {
// 		t.Fatalf("UpdateEntity() error = %v", err)
// 	}
// 	if updateResp == nil {
// 		t.Fatal("UpdateEntity() returned nil response")
// 	}

// 	// Verify relationship was updated
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_update_rel_only_entity_2"},
// 		Output: []string{"relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	rel := readResp.Relationships["service_update_rel_only_rel_1"]
// 	if rel == nil {
// 		t.Fatal("Relationship 'service_update_rel_only_rel_1' not found")
// 	}
// 	if rel.EndTime != "2025-06-30T00:00:00Z" {
// 		t.Errorf("Relationship EndTime = %v, want '2025-06-30T00:00:00Z'", rel.EndTime)
// 	}

// 	// Verify core attributes remain unchanged
// 	var stringValue wrapperspb.StringValue
// 	err = readResp.Name.GetValue().UnmarshalTo(&stringValue)
// 	if err != nil {
// 		t.Fatalf("Error unpacking Name value: %v", err)
// 	}
// 	if stringValue.Value != "Eve" {
// 		t.Errorf("Name changed to %v, should remain 'Eve'", stringValue.Value)
// 	}

// 	log.Printf("Successfully updated only relationships without modifying core attributes")
// }

// // TestServiceCreateEntityWithMetadataFullFlow tests creating an entity with metadata
// func TestServiceCreateEntityWithMetadataFullFlow(t *testing.T) {
// 	ctx := context.Background()

// 	// Create metadata
// 	metadata := make(map[string]*anypb.Any)
// 	metadata["department"], _ = anypb.New(&wrapperspb.StringValue{Value: "Engineering"})
// 	metadata["level"], _ = anypb.New(&wrapperspb.StringValue{Value: "Senior"})

// 	// Create entity with metadata
// 	entity := &pb.Entity{
// 		Id: "service_metadata_create_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:     createNameValue("2025-04-01T00:00:00Z", "Metadata User"),
// 		Created:  "2025-04-01T00:00:00Z",
// 		Metadata: metadata,
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() with metadata error = %v", err)
// 	}

// 	// Read entity back with metadata
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_metadata_create_entity_1"},
// 		Output: []string{"metadata"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	// Verify metadata was stored
// 	if len(readResp.Metadata) == 0 {
// 		t.Error("Expected metadata to be stored, but got empty metadata")
// 	}

// 	if len(readResp.Metadata) != 2 {
// 		t.Errorf("Expected 2 metadata fields, got %d", len(readResp.Metadata))
// 	}

// 	// Verify specific metadata values
// 	if _, exists := readResp.Metadata["department"]; !exists {
// 		t.Error("Expected 'department' metadata field not found")
// 	}
// 	if _, exists := readResp.Metadata["level"]; !exists {
// 		t.Error("Expected 'level' metadata field not found")
// 	}

// 	log.Printf("Successfully created entity with metadata")
// }

// // TestServiceCreateEntityWithoutMetadataFullFlow tests creating an entity without metadata
// func TestServiceCreateEntityWithoutMetadataFullFlow(t *testing.T) {
// 	ctx := context.Background()

// 	// Create entity without metadata
// 	entity := &pb.Entity{
// 		Id: "service_no_metadata_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "No Metadata User"),
// 		Created: "2025-04-01T00:00:00Z",
// 		// No metadata field set
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() without metadata error = %v", err)
// 	}

// 	// Read entity back
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_no_metadata_entity_1"},
// 		Output: []string{"metadata"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	// Verify entity exists but has empty metadata
// 	if readResp.Id != "service_no_metadata_entity_1" {
// 		t.Errorf("Entity ID = %v, want 'service_no_metadata_entity_1'", readResp.Id)
// 	}

// 	// Metadata should be empty
// 	if len(readResp.Metadata) > 0 {
// 		t.Errorf("Expected no metadata, but got %d metadata fields", len(readResp.Metadata))
// 	}

// 	log.Printf("Successfully created entity without metadata")
// }

// // TestServiceAddMetadataToExistingEntity tests adding metadata to an existing entity via update
// func TestServiceAddMetadataToExistingEntity(t *testing.T) {
// 	ctx := context.Background()

// 	// Create entity without metadata first
// 	entity := &pb.Entity{
// 		Id: "service_add_metadata_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Initially No Metadata"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}

// 	// Now add metadata via update
// 	metadata := make(map[string]*anypb.Any)
// 	metadata["department"], _ = anypb.New(&wrapperspb.StringValue{Value: "Sales"})
// 	metadata["location"], _ = anypb.New(&wrapperspb.StringValue{Value: "New York"})

// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_add_metadata_entity_1",
// 		Entity: &pb.Entity{
// 			Metadata: metadata,
// 		},
// 	}

// 	updateResp, err := server.UpdateEntity(ctx, updateReq)
// 	if err != nil {
// 		t.Fatalf("UpdateEntity() to add metadata error = %v", err)
// 	}
// 	if updateResp == nil {
// 		t.Fatal("UpdateEntity() returned nil response")
// 	}

// 	// Read entity back with metadata
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_add_metadata_entity_1"},
// 		Output: []string{"metadata"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	// Verify metadata was added
// 	if len(readResp.Metadata) != 2 {
// 		t.Errorf("Expected 2 metadata fields, got %d", len(readResp.Metadata))
// 	}

// 	if _, exists := readResp.Metadata["department"]; !exists {
// 		t.Error("Expected 'department' metadata field not found")
// 	}
// 	if _, exists := readResp.Metadata["location"]; !exists {
// 		t.Error("Expected 'location' metadata field not found")
// 	}

// 	log.Printf("Successfully added metadata to existing entity")
// }

// // TestServiceUpdateExistingMetadata tests updating metadata that already exists on an entity
// func TestServiceUpdateExistingMetadata(t *testing.T) {
// 	ctx := context.Background()

// 	// Create entity with initial metadata
// 	initialMetadata := make(map[string]*anypb.Any)
// 	initialMetadata["department"], _ = anypb.New(&wrapperspb.StringValue{Value: "Engineering"})
// 	initialMetadata["level"], _ = anypb.New(&wrapperspb.StringValue{Value: "Junior"})
// 	initialMetadata["location"], _ = anypb.New(&wrapperspb.StringValue{Value: "San Francisco"})

// 	entity := &pb.Entity{
// 		Id: "service_update_existing_metadata_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:     createNameValue("2025-04-01T00:00:00Z", "Metadata Update User"),
// 		Created:  "2025-04-01T00:00:00Z",
// 		Metadata: initialMetadata,
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}

// 	// Update the metadata with new values and add a new field
// 	updatedMetadata := make(map[string]*anypb.Any)
// 	updatedMetadata["department"], _ = anypb.New(&wrapperspb.StringValue{Value: "Product"}) // Changed
// 	updatedMetadata["level"], _ = anypb.New(&wrapperspb.StringValue{Value: "Senior"})       // Changed
// 	updatedMetadata["location"], _ = anypb.New(&wrapperspb.StringValue{Value: "New York"})  // Changed
// 	updatedMetadata["title"], _ = anypb.New(&wrapperspb.StringValue{Value: "Tech Lead"})    // New field

// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_update_existing_metadata_entity_1",
// 		Entity: &pb.Entity{
// 			Metadata: updatedMetadata,
// 		},
// 	}

// 	updateResp, err := server.UpdateEntity(ctx, updateReq)
// 	if err != nil {
// 		t.Fatalf("UpdateEntity() to update metadata error = %v", err)
// 	}
// 	if updateResp == nil {
// 		t.Fatal("UpdateEntity() returned nil response")
// 	}

// 	// Read entity back with metadata
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_update_existing_metadata_entity_1"},
// 		Output: []string{"metadata"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	// Verify metadata was updated with new values and new field added
// 	if len(readResp.Metadata) != 4 {
// 		t.Errorf("Expected 4 metadata fields, got %d", len(readResp.Metadata))
// 	}

// 	// Verify updated values
// 	if deptAny, exists := readResp.Metadata["department"]; exists {
// 		var deptValue wrapperspb.StringValue
// 		if err := deptAny.UnmarshalTo(&deptValue); err == nil {
// 			if deptValue.Value != "Product" {
// 				t.Errorf("department = %v, want 'Product'", deptValue.Value)
// 			}
// 		}
// 	} else {
// 		t.Error("Expected 'department' metadata field not found")
// 	}

// 	if levelAny, exists := readResp.Metadata["level"]; exists {
// 		var levelValue wrapperspb.StringValue
// 		if err := levelAny.UnmarshalTo(&levelValue); err == nil {
// 			if levelValue.Value != "Senior" {
// 				t.Errorf("level = %v, want 'Senior'", levelValue.Value)
// 			}
// 		}
// 	} else {
// 		t.Error("Expected 'level' metadata field not found")
// 	}

// 	if locAny, exists := readResp.Metadata["location"]; exists {
// 		var locValue wrapperspb.StringValue
// 		if err := locAny.UnmarshalTo(&locValue); err == nil {
// 			if locValue.Value != "New York" {
// 				t.Errorf("location = %v, want 'New York'", locValue.Value)
// 			}
// 		}
// 	} else {
// 		t.Error("Expected 'location' metadata field not found")
// 	}

// 	if titleAny, exists := readResp.Metadata["title"]; exists {
// 		var titleValue wrapperspb.StringValue
// 		if err := titleAny.UnmarshalTo(&titleValue); err == nil {
// 			if titleValue.Value != "Tech Lead" {
// 				t.Errorf("title = %v, want 'Tech Lead'", titleValue.Value)
// 			}
// 		}
// 	} else {
// 		t.Error("Expected 'title' metadata field not found")
// 	}

// 	log.Printf("Successfully updated existing metadata with new values and added new field")
// }

// // TestServiceUpdateEntityCoreAttributesMetadataAndRelationships tests updating core attributes, metadata, and relationships together
// func TestServiceUpdateEntityCoreAttributesMetadataAndRelationships(t *testing.T) {
// 	ctx := context.Background()

// 	// Create two entities
// 	entity1 := &pb.Entity{
// 		Id: "service_update_all_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Target Entity"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	metadata := make(map[string]*anypb.Any)
// 	metadata["department"], _ = anypb.New(&wrapperspb.StringValue{Value: "IT"})

// 	entity2 := &pb.Entity{
// 		Id: "service_update_all_entity_2",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Manager",
// 		},
// 		Name:     createNameValue("2025-04-01T00:00:00Z", "Manager User"),
// 		Created:  "2025-04-01T00:00:00Z",
// 		Metadata: metadata,
// 		Relationships: map[string]*pb.Relationship{
// 			"service_update_all_rel_1": {
// 				Id:              "service_update_all_rel_1",
// 				Name:            "MANAGES",
// 				RelatedEntityId: "service_update_all_entity_1",
// 				StartTime:       "2025-04-01T00:00:00Z",
// 			},
// 		},
// 	}

// 	_, err := server.CreateEntity(ctx, entity1)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity1) error = %v", err)
// 	}

// 	_, err = server.CreateEntity(ctx, entity2)
// 	if err != nil {
// 		t.Fatalf("CreateEntity(entity2) error = %v", err)
// 	}

// 	// Update core attributes, metadata, and relationships all together
// 	updatedMetadata := make(map[string]*anypb.Any)
// 	updatedMetadata["department"], _ = anypb.New(&wrapperspb.StringValue{Value: "HR"})
// 	updatedMetadata["title"], _ = anypb.New(&wrapperspb.StringValue{Value: "Director"})

// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_update_all_entity_2",
// 		Entity: &pb.Entity{
// 			Name:       createNameValue("2025-04-01T00:00:00Z", "Updated Manager"),
// 			Terminated: "2025-12-31T00:00:00Z",
// 			Metadata:   updatedMetadata,
// 			Relationships: map[string]*pb.Relationship{
// 				"service_update_all_rel_1": {
// 					Id:      "service_update_all_rel_1",
// 					EndTime: "2025-12-31T00:00:00Z",
// 				},
// 			},
// 		},
// 	}

// 	updateResp, err := server.UpdateEntity(ctx, updateReq)
// 	if err != nil {
// 		t.Fatalf("UpdateEntity() error = %v", err)
// 	}
// 	if updateResp == nil {
// 		t.Fatal("UpdateEntity() returned nil response")
// 	}

// 	// Read entity back with all fields
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_update_all_entity_2"},
// 		Output: []string{"metadata", "relationships"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	// Verify core attributes were updated
// 	var stringValue wrapperspb.StringValue
// 	err = readResp.Name.GetValue().UnmarshalTo(&stringValue)
// 	if err != nil {
// 		t.Fatalf("Error unpacking Name value: %v", err)
// 	}
// 	if stringValue.Value != "Updated Manager" {
// 		t.Errorf("Name = %v, want 'Updated Manager'", stringValue.Value)
// 	}
// 	if readResp.Terminated != "2025-12-31T00:00:00Z" {
// 		t.Errorf("Terminated = %v, want '2025-12-31T00:00:00Z'", readResp.Terminated)
// 	}

// 	// Verify metadata was updated
// 	if len(readResp.Metadata) != 2 {
// 		t.Errorf("Expected 2 metadata fields, got %d", len(readResp.Metadata))
// 	}
// 	if _, exists := readResp.Metadata["department"]; !exists {
// 		t.Error("Expected 'department' metadata field not found")
// 	}
// 	if _, exists := readResp.Metadata["title"]; !exists {
// 		t.Error("Expected 'title' metadata field not found")
// 	}

// 	// Verify relationship was updated
// 	rel := readResp.Relationships["service_update_all_rel_1"]
// 	if rel == nil {
// 		t.Fatal("Relationship 'service_update_all_rel_1' not found")
// 	}
// 	if rel.EndTime != "2025-12-31T00:00:00Z" {
// 		t.Errorf("Relationship EndTime = %v, want '2025-12-31T00:00:00Z'", rel.EndTime)
// 	}

// 	log.Printf("Successfully updated core attributes, metadata, and relationships together")
// }

// // TestServiceCreateEntityWithAttributes tests creating an entity with attributes
// func TestServiceCreateEntityWithAttributes(t *testing.T) {
// 	ctx := context.Background()

// 	// Create attributes with all three storage types: tabular, graph, and map
// 	attributes := make(map[string]*pb.TimeBasedValueList)

// 	// 1. Create a salary attribute with TABULAR data structure
// 	salaryData := map[string]interface{}{
// 		"columns": []interface{}{"amount", "currency"},
// 		"rows":    []interface{}{[]interface{}{"100000", "USD"}},
// 	}
// 	salaryStruct, _ := structpb.NewStruct(salaryData)
// 	salaryValue, _ := anypb.New(salaryStruct)

// 	attributes["salary"] = &pb.TimeBasedValueList{
// 		Values: []*pb.TimeBasedValue{
// 			{
// 				StartTime: "2025-04-01T00:00:00Z",
// 				EndTime:   "",
// 				Value:     salaryValue,
// 			},
// 		},
// 	}

// 	// 2. Create an org_chart attribute with GRAPH data structure
// 	orgChartData := map[string]interface{}{
// 		"nodes": []interface{}{
// 			map[string]interface{}{"id": "n1", "label": "Manager"},
// 			map[string]interface{}{"id": "n2", "label": "Employee"},
// 		},
// 		"edges": []interface{}{
// 			map[string]interface{}{"from": "n1", "to": "n2", "type": "manages"},
// 		},
// 	}
// 	orgChartStruct, _ := structpb.NewStruct(orgChartData)
// 	orgChartValue, _ := anypb.New(orgChartStruct)

// 	attributes["org_chart"] = &pb.TimeBasedValueList{
// 		Values: []*pb.TimeBasedValue{
// 			{
// 				StartTime: "2025-04-01T00:00:00Z",
// 				EndTime:   "",
// 				Value:     orgChartValue,
// 			},
// 		},
// 	}

// 	// 3. Create a profile attribute with MAP/DOCUMENT data structure
// 	profileData := map[string]interface{}{
// 		"skills":     []interface{}{"Go", "Python", "SQL"},
// 		"experience": "5 years",
// 		"education":  "Bachelor's",
// 	}
// 	profileStruct, _ := structpb.NewStruct(profileData)
// 	profileValue, _ := anypb.New(profileStruct)

// 	attributes["profile"] = &pb.TimeBasedValueList{
// 		Values: []*pb.TimeBasedValue{
// 			{
// 				StartTime: "2025-04-01T00:00:00Z",
// 				EndTime:   "",
// 				Value:     profileValue,
// 			},
// 		},
// 	}

// 	// Create entity with attributes
// 	entity := &pb.Entity{
// 		Id: "service_attributes_create_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:       createNameValue("2025-04-01T00:00:00Z", "Attributes User"),
// 		Created:    "2025-04-01T00:00:00Z",
// 		Attributes: attributes,
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() with attributes error = %v", err)
// 	}

// 	// Read entity back with attributes
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Id:         "service_attributes_create_entity_1",
// 			Attributes: attributes, // Pass attributes to indicate what to fetch
// 		},
// 		Output: []string{"attributes"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	// Verify all three attribute types were stored
// 	if len(readResp.Attributes) == 0 {
// 		t.Error("Expected attributes to be stored, but got empty attributes")
// 	}

// 	if _, exists := readResp.Attributes["salary"]; !exists {
// 		t.Error("Expected 'salary' (tabular) attribute not found")
// 	}

// 	if _, exists := readResp.Attributes["org_chart"]; !exists {
// 		t.Error("Expected 'org_chart' (graph) attribute not found")
// 	}

// 	if _, exists := readResp.Attributes["profile"]; !exists {
// 		t.Error("Expected 'profile' (map) attribute not found")
// 	}

// 	log.Printf("Successfully created entity with all three attribute types (tabular, graph, map)")
// }

// // TestServiceCreateEntityWithoutAttributes tests creating an entity without attributes
// func TestServiceCreateEntityWithoutAttributes(t *testing.T) {
// 	ctx := context.Background()

// 	// Create entity without attributes
// 	entity := &pb.Entity{
// 		Id: "service_no_attributes_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "No Attributes User"),
// 		Created: "2025-04-01T00:00:00Z",
// 		// No attributes field set
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() without attributes error = %v", err)
// 	}

// 	// Read entity back
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{Id: "service_no_attributes_entity_1"},
// 		Output: []string{"attributes"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	// Verify entity exists but has empty attributes
// 	if readResp.Id != "service_no_attributes_entity_1" {
// 		t.Errorf("Entity ID = %v, want 'service_no_attributes_entity_1'", readResp.Id)
// 	}

// 	// Attributes should be empty
// 	if len(readResp.Attributes) > 0 {
// 		t.Errorf("Expected no attributes, but got %d attribute fields", len(readResp.Attributes))
// 	}

// 	log.Printf("Successfully created entity without attributes")
// }

// // TestServiceAddAttributesToExistingEntity tests adding attributes to an existing entity via update
// func TestServiceAddAttributesToExistingEntity(t *testing.T) {
// 	ctx := context.Background()

// 	// Create entity without attributes first
// 	entity := &pb.Entity{
// 		Id: "service_add_attributes_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Initially No Attributes"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}

// 	// Now add attributes via update - all three types
// 	attributes := make(map[string]*pb.TimeBasedValueList)

// 	// Tabular type
// 	salaryData := map[string]interface{}{
// 		"columns": []interface{}{"amount", "currency"},
// 		"rows":    []interface{}{[]interface{}{"120000", "USD"}},
// 	}
// 	salaryStruct, _ := structpb.NewStruct(salaryData)
// 	salaryValue, _ := anypb.New(salaryStruct)

// 	attributes["salary"] = &pb.TimeBasedValueList{
// 		Values: []*pb.TimeBasedValue{
// 			{
// 				StartTime: "2025-04-01T00:00:00Z",
// 				EndTime:   "",
// 				Value:     salaryValue,
// 			},
// 		},
// 	}

// 	// Graph type
// 	teamData := map[string]interface{}{
// 		"nodes": []interface{}{
// 			map[string]interface{}{"id": "emp1", "name": "Employee 1"},
// 		},
// 		"edges": []interface{}{},
// 	}
// 	teamStruct, _ := structpb.NewStruct(teamData)
// 	teamValue, _ := anypb.New(teamStruct)

// 	attributes["team_structure"] = &pb.TimeBasedValueList{
// 		Values: []*pb.TimeBasedValue{
// 			{
// 				StartTime: "2025-04-01T00:00:00Z",
// 				EndTime:   "",
// 				Value:     teamValue,
// 			},
// 		},
// 	}

// 	// Map type
// 	contactData := map[string]interface{}{
// 		"email": "user@example.com",
// 		"phone": "555-1234",
// 	}
// 	contactStruct, _ := structpb.NewStruct(contactData)
// 	contactValue, _ := anypb.New(contactStruct)

// 	attributes["contact_info"] = &pb.TimeBasedValueList{
// 		Values: []*pb.TimeBasedValue{
// 			{
// 				StartTime: "2025-04-01T00:00:00Z",
// 				EndTime:   "",
// 				Value:     contactValue,
// 			},
// 		},
// 	}

// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_add_attributes_entity_1",
// 		Entity: &pb.Entity{
// 			Attributes: attributes,
// 		},
// 	}

// 	updateResp, err := server.UpdateEntity(ctx, updateReq)
// 	if err != nil {
// 		t.Fatalf("UpdateEntity() to add attributes error = %v", err)
// 	}
// 	if updateResp == nil {
// 		t.Fatal("UpdateEntity() returned nil response")
// 	}

// 	// Read entity back with attributes
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Id:         "service_add_attributes_entity_1",
// 			Attributes: attributes,
// 		},
// 		Output: []string{"attributes"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() error = %v", err)
// 	}

// 	// Verify all three attribute types were added
// 	if len(readResp.Attributes) == 0 {
// 		t.Error("Expected attributes to be added, but got empty attributes")
// 	}

// 	if _, exists := readResp.Attributes["salary"]; !exists {
// 		t.Error("Expected 'salary' (tabular) attribute not found")
// 	}

// 	if _, exists := readResp.Attributes["team_structure"]; !exists {
// 		t.Error("Expected 'team_structure' (graph) attribute not found")
// 	}

// 	if _, exists := readResp.Attributes["contact_info"]; !exists {
// 		t.Error("Expected 'contact_info' (map) attribute not found")
// 	}

// 	log.Printf("Successfully added all three attribute types to existing entity")
// }

// // TestServiceCreateEntityWithInvalidAttributeType tests that creating an entity with invalid attribute type fails
// func TestServiceCreateEntityWithInvalidAttributeType(t *testing.T) {
// 	ctx := context.Background()

// 	// Create an invalid attribute that doesn't match any storage type
// 	attributes := make(map[string]*pb.TimeBasedValueList)

// 	// Create an attribute with an invalid structure (not tabular, graph, or map)
// 	// Using a simple string value which is not a valid structured type
// 	invalidValue, _ := anypb.New(&wrapperspb.StringValue{Value: "just a string"})

// 	attributes["invalid_attr"] = &pb.TimeBasedValueList{
// 		Values: []*pb.TimeBasedValue{
// 			{
// 				StartTime: "2025-04-01T00:00:00Z",
// 				EndTime:   "",
// 				Value:     invalidValue,
// 			},
// 		},
// 	}

// 	entity := &pb.Entity{
// 		Id: "service_invalid_attribute_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:       createNameValue("2025-04-01T00:00:00Z", "Invalid Attr User"),
// 		Created:    "2025-04-01T00:00:00Z",
// 		Attributes: attributes,
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err == nil {
// 		t.Error("Expected error when creating entity with invalid attribute type, but got none")
// 	} else {
// 		log.Printf("CreateEntity correctly failed with invalid attribute type: %v", err)
// 	}

// 	log.Printf("Successfully verified that invalid attribute types are rejected")
// }

// // TestServiceUpdateEntityAddInvalidAttributeType tests that adding invalid attribute type via update fails
// func TestServiceUpdateEntityAddInvalidAttributeType(t *testing.T) {
// 	ctx := context.Background()

// 	// Create a valid entity first
// 	entity := &pb.Entity{
// 		Id: "service_invalid_attr_update_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:    createNameValue("2025-04-01T00:00:00Z", "Update Invalid Attr User"),
// 		Created: "2025-04-01T00:00:00Z",
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}

// 	// Try to add an invalid attribute via update
// 	attributes := make(map[string]*pb.TimeBasedValueList)

// 	// Create an invalid attribute (just a string, not a structured type)
// 	invalidValue, _ := anypb.New(&wrapperspb.StringValue{Value: "invalid data"})

// 	attributes["invalid_attr"] = &pb.TimeBasedValueList{
// 		Values: []*pb.TimeBasedValue{
// 			{
// 				StartTime: "2025-04-01T00:00:00Z",
// 				EndTime:   "",
// 				Value:     invalidValue,
// 			},
// 		},
// 	}

// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_invalid_attr_update_entity_1",
// 		Entity: &pb.Entity{
// 			Attributes: attributes,
// 		},
// 	}

// 	_, err = server.UpdateEntity(ctx, updateReq)
// 	if err == nil {
// 		t.Error("Expected error when adding invalid attribute type via update, but got none")
// 	} else {
// 		log.Printf("UpdateEntity correctly failed with invalid attribute type: %v", err)
// 	}

// 	log.Printf("Successfully verified that invalid attribute types are rejected during update")
// }

// // TestServiceAppendDataToTabularAttribute tests that updating a tabular attribute appends new rows
// func TestServiceAppendDataToTabularAttribute(t *testing.T) {
// 	ctx := context.Background()

// 	// Create entity with initial tabular attribute
// 	initialAttributes := make(map[string]*pb.TimeBasedValueList)

// 	// Create initial salary history with 2 entries
// 	salaryData := map[string]interface{}{
// 		"columns": []interface{}{"year", "amount", "currency"},
// 		"rows": []interface{}{
// 			[]interface{}{"2023", "90000", "USD"},
// 			[]interface{}{"2024", "100000", "USD"},
// 		},
// 	}
// 	salaryStruct, _ := structpb.NewStruct(salaryData)
// 	salaryValue, _ := anypb.New(salaryStruct)

// 	initialAttributes["salary_history"] = &pb.TimeBasedValueList{
// 		Values: []*pb.TimeBasedValue{
// 			{
// 				StartTime: "2025-04-01T00:00:00Z",
// 				EndTime:   "",
// 				Value:     salaryValue,
// 			},
// 		},
// 	}

// 	entity := &pb.Entity{
// 		Id: "service_append_tabular_entity_1",
// 		Kind: &pb.Kind{
// 			Major: "Person",
// 			Minor: "Employee",
// 		},
// 		Name:       createNameValue("2025-04-01T00:00:00Z", "Salary History User"),
// 		Created:    "2025-04-01T00:00:00Z",
// 		Attributes: initialAttributes,
// 	}

// 	_, err := server.CreateEntity(ctx, entity)
// 	if err != nil {
// 		t.Fatalf("CreateEntity() error = %v", err)
// 	}

// 	// Read initial attribute to verify 2 rows
// 	readReq := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Id:         "service_append_tabular_entity_1",
// 			Attributes: initialAttributes,
// 		},
// 		Output: []string{"attributes"},
// 	}

// 	readResp, err := server.ReadEntity(ctx, readReq)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() before update error = %v", err)
// 	}

// 	// Verify initial data has 2 rows
// 	if salaryHistoryList, exists := readResp.Attributes["salary_history"]; exists {
// 		if len(salaryHistoryList.Values) > 0 {
// 			var initialSalaryStruct structpb.Struct
// 			err := salaryHistoryList.Values[0].Value.UnmarshalTo(&initialSalaryStruct)
// 			if err != nil {
// 				t.Fatalf("Error unpacking initial salary_history: %v", err)
// 			}
// 			if rowsField, ok := initialSalaryStruct.Fields["rows"]; ok {
// 				if rowsList := rowsField.GetListValue(); rowsList != nil {
// 					if len(rowsList.Values) != 2 {
// 						t.Errorf("Expected 2 initial rows, got %d", len(rowsList.Values))
// 					}
// 				}
// 			}
// 		}
// 	} else {
// 		t.Fatal("Expected 'salary_history' attribute not found in initial read")
// 	}

// 	// Now update with same attribute name to append new row (2025 salary)
// 	appendAttributes := make(map[string]*pb.TimeBasedValueList)

// 	newSalaryData := map[string]interface{}{
// 		"columns": []interface{}{"year", "amount", "currency"},
// 		"rows": []interface{}{
// 			[]interface{}{"2025", "115000", "USD"}, // New row to append
// 		},
// 	}
// 	newSalaryStruct, _ := structpb.NewStruct(newSalaryData)
// 	newSalaryValue, _ := anypb.New(newSalaryStruct)

// 	appendAttributes["salary_history"] = &pb.TimeBasedValueList{
// 		Values: []*pb.TimeBasedValue{
// 			{
// 				StartTime: "2025-05-01T00:00:00Z",
// 				EndTime:   "",
// 				Value:     newSalaryValue,
// 			},
// 		},
// 	}

// 	updateReq := &pb.UpdateEntityRequest{
// 		Id: "service_append_tabular_entity_1",
// 		Entity: &pb.Entity{
// 			Attributes: appendAttributes,
// 		},
// 	}

// 	updateResp, err := server.UpdateEntity(ctx, updateReq)
// 	if err != nil {
// 		t.Fatalf("UpdateEntity() to append tabular data error = %v", err)
// 	}
// 	if updateResp == nil {
// 		t.Fatal("UpdateEntity() returned nil response")
// 	}

// 	// Read entity back and verify data was appended (should now have 3 rows total)
// 	readReq2 := &pb.ReadEntityRequest{
// 		Entity: &pb.Entity{
// 			Id:         "service_append_tabular_entity_1",
// 			Attributes: appendAttributes,
// 		},
// 		Output: []string{"attributes"},
// 	}

// 	readResp2, err := server.ReadEntity(ctx, readReq2)
// 	if err != nil {
// 		t.Fatalf("ReadEntity() after update error = %v", err)
// 	}

// 	// Verify attribute still exists
// 	if _, exists := readResp2.Attributes["salary_history"]; !exists {
// 		t.Fatal("Expected 'salary_history' attribute not found after update")
// 	}

// 	// Verify data was appended (should have 3 rows now: 2023, 2024, 2025)
// 	salaryHistoryList := readResp2.Attributes["salary_history"]
// 	if len(salaryHistoryList.Values) == 0 {
// 		t.Fatal("Expected salary_history to have values")
// 	}

// 	// Unpack the result - data comes back as a JSON string in a "data" field
// 	var resultStruct structpb.Struct
// 	err = salaryHistoryList.Values[0].Value.UnmarshalTo(&resultStruct)
// 	if err != nil {
// 		t.Fatalf("Error unpacking salary_history result: %v", err)
// 	}

// 	// Get the "data" field which contains the JSON string
// 	dataField, ok := resultStruct.Fields["data"]
// 	if !ok {
// 		t.Fatal("Expected 'data' field in salary_history result")
// 	}

// 	jsonDataStr := dataField.GetStringValue()
// 	if jsonDataStr == "" {
// 		t.Fatal("Expected non-empty JSON data string")
// 	}

// 	// Parse the JSON string to verify the rows
// 	// The JSON structure is: {"columns":["year","amount","currency"],"rows":[...]}
// 	// We'll do a simple check by counting occurrences of row data
// 	if !strings.Contains(jsonDataStr, `"2023"`) {
// 		t.Error("Expected to find 2023 row in data")
// 	}
// 	if !strings.Contains(jsonDataStr, `"2024"`) {
// 		t.Error("Expected to find 2024 row in data")
// 	}
// 	if !strings.Contains(jsonDataStr, `"2025"`) {
// 		t.Error("Expected to find 2025 row in data (newly appended)")
// 	}
// 	if !strings.Contains(jsonDataStr, `"115000"`) {
// 		t.Error("Expected to find new salary amount 115000 in data")
// 	}

// 	// Count rows by counting array elements in the JSON
// 	rowCount := strings.Count(jsonDataStr, `["2023"`) + strings.Count(jsonDataStr, `["2024"`) + strings.Count(jsonDataStr, `["2025"`)
// 	if rowCount != 3 {
// 		t.Errorf("Expected 3 rows in data, found %d", rowCount)
// 	}

// 	log.Printf("Successfully appended new row to tabular attribute (3 total rows now)")
// }
