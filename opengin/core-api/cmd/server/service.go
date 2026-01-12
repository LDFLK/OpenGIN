package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"lk/datafoundation/core-api/db/config"
	pb "lk/datafoundation/core-api/lk/datafoundation/core-api"

	mongorepository "lk/datafoundation/core-api/db/repository/mongo"
	neo4jrepository "lk/datafoundation/core-api/db/repository/neo4j"
	postgres "lk/datafoundation/core-api/db/repository/postgres"
	engine "lk/datafoundation/core-api/engine"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

// Server implements the COREService
type Server struct {
	pb.UnimplementedCOREServiceServer
	mongoRepo    *mongorepository.MongoRepository
	neo4jRepo    *neo4jrepository.Neo4jRepository
	postgresRepo *postgres.PostgresRepository
}

// CreateEntity handles entity creation with relationships, metadata and attributes
func (s *Server) CreateEntity(ctx context.Context, req *pb.Entity) (*pb.Entity, error) {
	log.Printf("Creating Entity: %s", req.Id)

	// Validate required fields for Neo4j entity creation
	success, err := s.neo4jRepo.HandleGraphEntityCreation(ctx, req)
	if !success {
		log.Printf("[server.CreateEntity] Error saving entity in Neo4j: %v", err)
		return nil, err
	} else {
		log.Printf("[server.CreateEntity] Successfully saved entity in Neo4j for entity: %s", req.Id)
	}

	// Handle relationships
	err = s.neo4jRepo.HandleGraphRelationshipsCreate(ctx, req)
	if err != nil {
		log.Printf("[server.CreateEntity] Error saving relationships in Neo4j: %v", err)
		return nil, err
	} else {
		log.Printf("[server.CreateEntity] Successfully saved relationships in Neo4j for entity: %s", req.Id)
	}

	// The HandleMetadata function will only process it if it has metadata
	// If metadata is not provided, a document will not be created in MongoDB
	// FIXME: https://github.com/LDFLK/nexoan/issues/120
	err = s.mongoRepo.HandleMetadata(ctx, req.Id, req)
	if err != nil {
		log.Printf("[server.CreateEntity] Error saving metadata in MongoDB: %v", err)
		return nil, err
	} else {
		log.Printf("[server.CreateEntity] Successfully saved metadata in MongoDB for entity: %s", req.Id)
	}

	// Handle attributes
	processor := engine.NewEntityAttributeProcessor()
	attributeResults := processor.ProcessEntityAttributes(ctx, req, "create", nil)

	// Check if any attributes failed
	hasErrors := false
	for attrName, result := range attributeResults {
		if !result.Success || result.Error != nil {
			log.Printf("[server.CreateEntity] Error handling attribute %s: %v", attrName, result.Error)
			hasErrors = true
		} else {
			log.Printf("[server.CreateEntity] Successfully handled attribute %s for entity: %s", attrName, req.Id)
		}
	}

	if hasErrors {
		log.Printf("[server.CreateEntity] Some attributes failed to process")
		return nil, fmt.Errorf("some attributes failed to process")
	}

	return req, nil
}

// ReadEntity retrieves an entity
func (s *Server) ReadEntity(ctx context.Context, req *pb.ReadEntityRequest) (*pb.Entity, error) {
	log.Printf("Reading Entity: %s with output fields: %v", req.Entity.Id, req.Output)

	// Initialize a complete response entity with empty fields
	response := &pb.Entity{
		Id:            req.Entity.Id,
		Kind:          &pb.Kind{},
		Name:          &pb.TimeBasedValue{},
		Created:       "",
		Terminated:    "",
		Metadata:      make(map[string]*anypb.Any),
		Attributes:    make(map[string]*pb.TimeBasedValueList),
		Relationships: make(map[string]*pb.Relationship),
	}

	// Always fetch basic entity info from Neo4j
	kind, name, created, terminated, err := s.neo4jRepo.GetGraphEntity(ctx, req.Entity.Id)
	if err != nil {
		log.Printf("Error fetching entity info: %v", err)
		return nil, fmt.Errorf("error fetching entity info: %v", err)
	} else {
		response.Kind = kind
		response.Name = name
		response.Created = created
		response.Terminated = terminated
	}

	// If no output fields specified, return the entity with basic info
	if len(req.Output) == 0 {
		log.Printf("Returning entity from ReadEntity: %+v", response)
		return response, nil
	}

	// Process each requested output field
	for _, field := range req.Output {
		log.Printf("[DEBUG] Entering switch statement for entity ID: %s", req.Entity.Id)
		switch field {
		case "metadata":
			log.Printf("[DEBUG] Processing metadata field for entity ID: %s", req.Entity.Id)
			// Get metadata from MongoDB
			metadata, err := s.mongoRepo.GetMetadata(ctx, req.Entity.Id)
			if err != nil {
				log.Printf("Error fetching metadata: %v", err)
				return nil, fmt.Errorf("error fetching metadata: %v", err)
			} else {
				log.Printf("[DEBUG] Retrieved metadata: %+v", metadata)
				response.Metadata = metadata
			}

		case "relationships":
			// Handle relationships based on the input entity
			if req.Entity != nil {
				if len(req.Entity.Relationships) == 0 {
					// No filters provided, fetch all relationships for the entity
					filteredRels, err := s.neo4jRepo.GetFilteredRelationships(ctx, req.Entity.Id, "", "", "", "", "", "", req.ActiveAt)
					if err != nil {
						log.Printf("Error fetching related entity IDs for entity %s: %v", req.Entity.Id, err)
						return nil, fmt.Errorf("error fetching related entity IDs: %v", err)
					} else {
						for id, relationship := range filteredRels {
							response.Relationships[id] = relationship
						}
					}
				} else {
					// Call GetFilteredRelationships for each relationship
					for _, rel := range req.Entity.Relationships {
						log.Printf("Fetching related entity IDs for entity %s with relationship %s and start time %s", req.Entity.Id, rel.Name, rel.StartTime)
						filteredRels, err := s.neo4jRepo.GetFilteredRelationships(ctx, req.Entity.Id, rel.Id, rel.Name, rel.RelatedEntityId, rel.StartTime, rel.EndTime, rel.Direction, req.ActiveAt)
						if err != nil {
							log.Printf("Error fetching related entity IDs for entity %s: %v", req.Entity.Id, err)
							return nil, fmt.Errorf("error fetching related entity IDs: %v", err)
						}

						// Add the relationships to the response
						for id, relationship := range filteredRels {
							response.Relationships[id] = relationship
						}
					}
				}
			} else {
				return nil, fmt.Errorf("entity is required to fetch relationships")
			}

		case "attributes":
			log.Printf("Processing attributes for entity: %s", req.Entity.Id)

			// For now, create a minimal entity with test attributes to demonstrate the conversion

			log.Printf("[server.ReadEntity] Processing attributes for entity: %s, attributes: %+v", req.Entity.Id, req.Entity.Attributes)

			// Use the EntityAttributeProcessor to read and process attributes
			processor := engine.NewEntityAttributeProcessor()

			// Extract fields from the request attributes based on storage type
			fields := extractFieldsFromAttributes(req.Entity.Attributes)
			log.Printf("Extracted fields from attributes: %v", fields)

			readOptions := engine.NewReadOptions(make(map[string]interface{}), fields...)

			// Process the entity with attributes to get the results map
			attributeResults := processor.ProcessEntityAttributes(ctx, req.Entity, "read", readOptions)

			log.Printf("[server.ReadEntity] Successfully processed attributes for entity: %s, results: %+v", req.Entity.Id, attributeResults)

			// Convert the results map back to TimeBasedValueList and attach to response.Attributes
			for attrName, result := range attributeResults {
				log.Printf("[server.ReadEntity] Successfully processed attribute %s for entity: %s, result: %+v", attrName, req.Entity.Id, result)
				if result.Success && result.Data != nil {
					// Convert the result data back to TimeBasedValue format
					if timeBasedValue, ok := result.Data.(*pb.TimeBasedValue); ok {
						// If the data is already in TimeBasedValue format, use it directly
						log.Printf("[server.ReadEntity] Successfully processed attribute %s for entity: %s", attrName, req.Entity.Id)
						response.Attributes[attrName] = &pb.TimeBasedValueList{
							Values: []*pb.TimeBasedValue{timeBasedValue},
						}
					} else {
						// Convert other data types to TimeBasedValue format
						log.Printf("[server.ReadEntity] Successfully processed attribute %s for entity: %s", attrName, req.Entity.Id)
						response.Attributes[attrName] = &pb.TimeBasedValueList{
							Values: []*pb.TimeBasedValue{
								{
									StartTime: "",
									EndTime:   "",
									Value: &anypb.Any{
										TypeUrl: "type.googleapis.com/google.protobuf.StringValue",
										Value:   []byte(fmt.Sprintf("%v", result.Data)),
									},
								},
							},
						}
					}
				}
			}

		default:
			log.Printf("Unknown output field requested: %s", field)
			return nil, fmt.Errorf("unknown output field requested: %s", field)
		}
	}
	return response, nil
}

// UpdateEntity modifies existing metadata
func (s *Server) UpdateEntity(ctx context.Context, req *pb.UpdateEntityRequest) (*pb.Entity, error) {
	// Extract ID from request parameter and entity data
	updateEntityID := req.Id
	updateEntity := req.Entity

	// Ensure the entity ID matches the URL parameter - since the id is already passed in the url param, the user does not need to pass it again in the payload
	if updateEntity.Id == "" || updateEntity.Id != updateEntityID {
		updateEntity.Id = updateEntityID
	}

	// Pass the ID and metadata to HandleMetadata- if no metadata was provided this will rerturn nil
	err := s.mongoRepo.HandleMetadata(ctx, updateEntityID, updateEntity)
	if err != nil {
		log.Printf("[server.UpdateEntity] Error updating metadata for entity %s: %v", updateEntityID, err)
		return nil, fmt.Errorf("error updating metadata for entity %s: %v", updateEntityID, err)
	}

	// Handle Graph Entity update if entity has required fields
	success, err := s.neo4jRepo.HandleGraphEntityUpdate(ctx, updateEntity)
	if !success {
		log.Printf("[server.UpdateEntity] Error updating graph entity for %s: %v", updateEntityID, err)
		return nil, fmt.Errorf("error updating graph entity for entity %s: %v", updateEntityID, err)
	}

	// Handle Relationships update
	err = s.neo4jRepo.HandleGraphRelationshipsUpdate(ctx, updateEntity)
	if err != nil {
		log.Printf("[server.UpdateEntity] Error updating relationships for entity %s: %v", updateEntityID, err)
		return nil, fmt.Errorf("error updating relationships for entity %s: %v", updateEntityID, err)
	}

	// Handle attributes
	processor := engine.NewEntityAttributeProcessor()
	// Note that in the perspective of the attribute this is a creation operation
	// The entity is already there but here the attribute is set later.
	// There is no alignment of update operation with the attribute.
	// TODO: https://github.com/LDFLK/nexoan/issues/286
	attributeResults := processor.ProcessEntityAttributes(ctx, req.Entity, "create", nil)

	// Check if any attributes failed
	hasErrors := false
	for attrName, result := range attributeResults {
		if !result.Success || result.Error != nil {
			log.Printf("[server.CreateEntity] Error handling attribute %s: %v", attrName, result.Error)
			hasErrors = true
		} else {
			log.Printf("[server.CreateEntity] Successfully handled attribute %s for entity: %s", attrName, req.Id)
		}
	}

	if hasErrors {
		log.Printf("[server.CreateEntity] Some attributes failed to process")
		return nil, fmt.Errorf("some attributes failed to process")
	}

	// Prepare the Update Response

	// Read entity data from Neo4j to include in response
	kind, name, created, terminated, _ := s.neo4jRepo.GetGraphEntity(ctx, updateEntityID)

	// Get relationships from Neo4j
	relationships, _ := s.neo4jRepo.GetGraphRelationships(ctx, updateEntityID)

	// Get metadata from MongoDB
	metadata, _ := s.mongoRepo.GetMetadata(ctx, updateEntityID)

	// Return updated entity with all available information
	return &pb.Entity{
		Id:            updateEntity.Id,
		Kind:          kind,
		Name:          name,
		Created:       created,
		Terminated:    terminated,
		Metadata:      metadata,
		Attributes:    make(map[string]*pb.TimeBasedValueList), // Empty attributes
		Relationships: relationships,
	}, nil
}

// DeleteEntity removes metadata
func (s *Server) DeleteEntity(ctx context.Context, req *pb.EntityId) (*pb.Empty, error) {
	log.Printf("[server.DeleteEntity] Deleting Entity metadata: %s", req.Id)

	// Check if entity exists before deleting
	_, err := s.mongoRepo.ReadEntity(ctx, req.Id)
	if err != nil {
		// NOTE: Not returning an error here because we want to delete the
		// entity even if it does not contain metadata
		log.Printf("[server.DeleteEntity] Entity %s does not contain metadata: %v", req.Id, err)
	} else {
		log.Printf("[server.DeleteEntity] Entity %s metadata exists.", req.Id)
		_, err = s.mongoRepo.DeleteEntity(ctx, req.Id)
		if err != nil {
			// Log error
			log.Printf("[server.DeleteEntity] Error deleting metadata for entity %s: %v", req.Id, err)
			return nil, fmt.Errorf("error deleting metadata for entity %s: %v", req.Id, err)
		} else {
			log.Printf("[server.DeleteEntity] Entity %s metadata deleted.", req.Id)
		}
	}
	// TODO: Implement Relationship Deletion in Neo4j
	// TODO: Implement Entity Deletion in Neo4j
	// TODO: Implement Attribute Deletion in Neo4j
	return &pb.Empty{}, nil
}

// ReadEntities retrieves a list of entities filtered by base attributes
func (s *Server) ReadEntities(ctx context.Context, req *pb.ReadEntityRequest) (*pb.EntityList, error) {
	if req.Entity == nil {
		return nil, fmt.Errorf("entity is required for filtering entities")
	}

	// Check if we have either an ID or Kind.Major
	if req.Entity.Id == "" && (req.Entity.Kind == nil || req.Entity.Kind.Major == "") {
		return nil, fmt.Errorf("either Entity.Id or Entity.Kind.Major is required for filtering entities")
	}

	// If we have an ID, add it to the filters
	if req.Entity.Id != "" {
		log.Printf("Filtering entities by ID: %s", req.Entity.Id)
	} else {
		log.Printf("Filtering entities by Kind.Major: %s", req.Entity.Kind.Major)
	}

	// Use HandleGraphEntityFilter to get filtered entities
	filteredEntities, err := s.neo4jRepo.HandleGraphEntityFilter(ctx, req)
	if err != nil {
		log.Printf("Error filtering entities: %v", err)
		return nil, err
	}

	// Convert filtered entities to pb.Entity format
	var entities []*pb.Entity
	for _, entity := range filteredEntities {
		pbEntity := &pb.Entity{
			Id: entity["id"].(string),
			Kind: &pb.Kind{
				Major: entity["kind"].(string),
				Minor: entity["minorKind"].(string),
			},
			Created: entity["created"].(string),
			Name: &pb.TimeBasedValue{ // How to represent time based value name?
				StartTime: entity["created"].(string),
				Value: &anypb.Any{
					TypeUrl: "type.googleapis.com/google.protobuf.StringValue",
					Value:   []byte(entity["name"].(string)),
				},
			},
		}

		// Add terminated if present
		if terminated, ok := entity["terminated"].(string); ok && terminated != "" {
			pbEntity.Terminated = terminated
			pbEntity.Name.EndTime = terminated
		}

		entities = append(entities, pbEntity)
	}

	return &pb.EntityList{
		Entities: entities,
	}, nil
}

// extractFieldsFromAttributes extracts field names from entity attributes based on storage type
// TODO: Limitation in multi-value attribute reads.
// FIXME: https://github.com/LDFLK/nexoan/issues/285
func extractFieldsFromAttributes(attributes map[string]*pb.TimeBasedValueList) []string {
	var fields []string

	for attrName, attrValueList := range attributes {
		if attrValueList == nil || len(attrValueList.Values) == 0 {
			continue
		}

		// Get the first value to determine storage type
		value := attrValueList.Values[0]
		if value == nil || value.Value == nil {
			continue
		}

		// Determine storage type and extract fields accordingly
		storageType, err := determineStorageTypeFromValue(value.Value)
		if err != nil {
			log.Printf("Warning: could not determine storage type for attribute %s: %v", attrName, err)
			continue
		}

		switch storageType {
		case "tabular":
			// For tabular data, extract columns from the attribute value
			if columns, err := extractColumnsFromTabularAttribute(value.Value); err == nil {
				fields = append(fields, columns...)
			} else {
				log.Printf("Warning: could not extract columns from tabular attribute %s: %v", attrName, err)
			}
		case "graph":
			// TODO: Handle graph data fields
			log.Printf("Graph data fields extraction not implemented yet for attribute %s", attrName)
		case "map":
			// TODO: Handle document/map data fields
			log.Printf("Document data fields extraction not implemented yet for attribute %s", attrName)
		default:
			log.Printf("Unknown storage type %s for attribute %s", storageType, attrName)
		}
	}

	return fields
}

// determineStorageTypeFromValue determines the storage type from a protobuf Any value
func determineStorageTypeFromValue(anyValue *anypb.Any) (string, error) {
	// Unpack the Any value to get the underlying message
	message, err := anyValue.UnmarshalNew()
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal Any value: %v", err)
	}

	// Check if it's a struct
	if structValue, ok := message.(*structpb.Struct); ok {
		// Check for tabular structure (has both "columns" and "rows" fields)
		if _, hasColumns := structValue.Fields["columns"]; hasColumns {
			if _, hasRows := structValue.Fields["rows"]; hasRows {
				return "tabular", nil
			}
		}

		// Check for graph structure (has both "nodes" and "edges" fields)
		if _, hasNodes := structValue.Fields["nodes"]; hasNodes {
			if _, hasEdges := structValue.Fields["edges"]; hasEdges {
				return "graph", nil
			}
		}

		// If it has fields but not the specific structures above, treat as map/document
		if len(structValue.Fields) > 0 {
			return "map", nil
		}
	}

	return "unknown", nil
}

// extractColumnsFromTabularAttribute extracts column names from a tabular attribute value
func extractColumnsFromTabularAttribute(anyValue *anypb.Any) ([]string, error) {
	// Unpack the Any value to get the underlying message
	message, err := anyValue.UnmarshalNew()
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Any value: %v", err)
	}

	// Check if it's a struct
	structValue, ok := message.(*structpb.Struct)
	if !ok {
		return nil, fmt.Errorf("expected struct value")
	}

	// Get the columns field
	columnsField, exists := structValue.Fields["columns"]
	if !exists {
		return nil, fmt.Errorf("no columns field found")
	}

	// Extract column names from the list value
	listValue := columnsField.GetListValue()
	if listValue == nil {
		return nil, fmt.Errorf("columns field is not a list")
	}

	columns := make([]string, len(listValue.Values))
	for i, col := range listValue.Values {
		columns[i] = col.GetStringValue()
	}

	return columns, nil
}

// Start the gRPC server
func main() {
	// Initialize MongoDB config
	mongoConfig := &config.MongoConfig{
		URI:        os.Getenv("MONGO_URI"),
		DBName:     os.Getenv("MONGO_DB_NAME"),
		Collection: os.Getenv("MONGO_COLLECTION"),
	}

	// Initialize Neo4j config
	neo4jConfig := &config.Neo4jConfig{
		URI:      os.Getenv("NEO4J_URI"),
		Username: os.Getenv("NEO4J_USER"),
		Password: os.Getenv("NEO4J_PASSWORD"),
	}

	// Initialize PostgreSQL config
	postgresConfig := &postgres.Config{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  os.Getenv("POSTGRES_SSL_MODE"),
	}

	// Get host and port from environment variables with defaults
	host := os.Getenv("CORE_SERVICE_HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	port := os.Getenv("CORE_SERVICE_PORT")
	if port == "" {
		port = "50051"
	}

	// Create MongoDB repository
	ctx := context.Background()
	mongoRepo := mongorepository.NewMongoRepository(ctx, mongoConfig)

	// Create Neo4j repository
	neo4jRepo, err := neo4jrepository.NewNeo4jRepository(ctx, neo4jConfig)
	if err != nil {
		log.Fatalf("[service.main] Failed to create Neo4j repository: %v", err)
	}
	defer neo4jRepo.Close(ctx)

	// Create PostgreSQL repository
	postgresRepo, err := postgres.NewPostgresRepository(*postgresConfig)
	if err != nil {
		log.Fatalf("[service.main] Failed to create PostgreSQL repository: %v", err)
	}
	defer postgresRepo.Close()

	listener, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		log.Fatalf("[service.main] Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	server := &Server{
		mongoRepo:    mongoRepo,
		neo4jRepo:    neo4jRepo,
		postgresRepo: postgresRepo,
	}

	pb.RegisterCOREServiceServer(grpcServer, server)

	// Register reflection service
	reflection.Register(grpcServer)

	log.Printf("[service.main] CORE Service is running on %s:%s...", host, port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("[service.main] Failed to serve: %v", err)
	}
}
