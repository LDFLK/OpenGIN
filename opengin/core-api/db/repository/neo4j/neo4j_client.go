package neo4jrepository

import (
	"context"
	"fmt"
	"lk/datafoundation/core-api/db/config"
	pb "lk/datafoundation/core-api/lk/datafoundation/core-api"
	"log"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jRepository struct {
	client neo4j.DriverWithContext
	config *config.Neo4jConfig
}

// NewNeo4jRepository initializes a Neo4j driver
func NewNeo4jRepository(ctx context.Context, config *config.Neo4jConfig) (*Neo4jRepository, error) {
	client, err := neo4j.NewDriverWithContext(config.URI, neo4j.BasicAuth(config.Username, config.Password, ""))
	if err != nil {
		log.Printf("[neo4j_client.NewNeo4jRepository] failed to create Neo4j driver: %v", err)
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	// Verify connectivity
	if err := client.VerifyConnectivity(ctx); err != nil {
		client.Close(ctx) // Close if connectivity check fails
		log.Printf("[neo4j_client.NewNeo4jRepository] failed to connect to Neo4j: %v", err)
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	log.Println("[neo4j_client.NewNeo4jRepository] Connected to Neo4j successfully!")

	return &Neo4jRepository{
		client: client,
		config: config,
	}, nil
}

// Close properly closes the Neo4j driver
func (r *Neo4jRepository) Close(ctx context.Context) {
	if r.client != nil {
		r.client.Close(ctx)
		log.Println("Neo4j connection closed")
	}
}

// getSession creates a new session
func (r *Neo4jRepository) getSession(ctx context.Context) neo4j.SessionWithContext {
	return r.client.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
}

// CreateGraphEntity checks if an entity exists and creates it if it doesn't
func (r *Neo4jRepository) CreateGraphEntity(ctx context.Context, kind *pb.Kind, entityMap map[string]interface{}) (map[string]interface{}, error) {
	// Validate the kind parameter
	if kind == nil || kind.Major == "" {
		log.Printf("[neo4j_client.CreateGraphEntity] missing or invalid 'Kind.Major' field")
		return nil, fmt.Errorf("[neo4j_client.CreateGraphEntity] missing or invalid 'Kind.Major' field")
	} else {
		log.Printf("[neo4j_client.CreateGraphEntity] Kind.Major: %v", kind.Major)
	}

	// Extract the required fields from the entityMap
	id, ok := entityMap["Id"].(string)
	if !ok {
		log.Printf("[neo4j_client.CreateGraphEntity] missing or invalid 'Id' field")
		return nil, fmt.Errorf("[neo4j_client.CreateGraphEntity] missing or invalid 'Id' field")
	} else {
		log.Printf("[neo4j_client.CreateGraphEntity] Id: %v", id)
	}

	name, ok := entityMap["Name"].(string)
	if !ok {
		log.Printf("[neo4j_client.CreateGraphEntity] missing or invalid 'Name' field")
		return nil, fmt.Errorf("[neo4j_client.CreateGraphEntity] missing or invalid 'Name' field")
	} else {
		log.Printf("[neo4j_client.CreateGraphEntity] Name: %v", name)
	}

	created, ok := entityMap["Created"].(string)
	if !ok {
		log.Printf("[neo4j_client.CreateGraphEntity] missing or invalid 'Created' field")
		return nil, fmt.Errorf("[neo4j_client.CreateGraphEntity] missing or invalid 'Created' field")
	} else {
		log.Printf("[neo4j_client.CreateGraphEntity] Created: %v", created)
	}

	// Optional field
	var terminated *string
	if term, ok := entityMap["Terminated"].(string); ok {
		terminated = &term
	} else {
		log.Printf("[neo4j_client.CreateGraphEntity] Terminated: %v", terminated)
	}

	// Open a session
	session := r.getSession(ctx)
	defer session.Close(ctx)

	// Check if the node already exists
	existsQuery := `MATCH (e:` + kind.Major + ` {Id: $Id}) RETURN e`
	result, err := session.Run(ctx, existsQuery, map[string]interface{}{"Id": id})
	if err != nil {
		log.Printf("[neo4j_client.CreateGraphEntity] error checking if entity exists: %v", err)
		return nil, fmt.Errorf("[neo4j_client.CreateGraphEntity] error checking if entity exists: %v", err)
	} else {
		log.Printf("[neo4j_client.CreateGraphEntity] existsQuery: %v", existsQuery)
	}

	// If entity exists, return an error
	if result.Next(ctx) {
		log.Printf("[neo4j_client.CreateGraphEntity] entity with Id %s already exists", id)
		return nil, fmt.Errorf("[neo4j_client.CreateGraphEntity] entity with Id %s already exists", id)
	} else {
		log.Printf("[neo4j_client.CreateGraphEntity] entity with Id %s does not exist", id)
	}

	// Create the node
	createQuery := `CREATE (e:` + kind.Major + ` {Id: $Id, Name: $Name, Created: datetime($Created), MinorKind: $MinorKind`
	if terminated != nil {
		createQuery += `, Terminated: datetime($Terminated)`
	}
	createQuery += `}) RETURN e`

	// Set parameters for the query
	params := map[string]interface{}{
		"Id":        id,
		"Name":      name,
		"Created":   created,
		"MinorKind": kind.Minor,
	}
	if terminated != nil {
		params["Terminated"] = *terminated
	}

	// Run the query to create the entity and return it
	result, err = session.Run(ctx, createQuery, params)
	if err != nil {
		log.Printf("[neo4j_client.CreateGraphEntity] error creating entity: %v", err)
		return nil, fmt.Errorf("[neo4j_client.CreateGraphEntity] error creating entity: %v", err)
	} else {
		log.Printf("[neo4j_client.CreateGraphEntity] created entity(run query): %v", params)
	}

	// Retrieve the created entity
	if result.Next(ctx) {
		createdEntity, _ := result.Record().Get("e")
		node, ok := createdEntity.(neo4j.Node)
		if !ok {
			log.Printf("[neo4j_client.CreateGraphEntity] failed to cast created entity to neo4j.Node")
			return nil, fmt.Errorf("[neo4j_client.CreateGraphEntity] failed to cast created entity to neo4j.Node")
		} else {
			log.Printf("[neo4j_client.CreateGraphEntity] created entity(retrieved-initial): %v", createdEntity)
		}

		// Convert the node properties to a map
		createdEntityMap := map[string]interface{}{
			"Id":        fmt.Sprintf("%v", node.Props["Id"]),
			"Name":      fmt.Sprintf("%v", node.Props["Name"]),
			"MinorKind": fmt.Sprintf("%v", node.Props["MinorKind"]),
		}

		// Handle date fields with proper formatting
		if created, ok := node.Props["Created"].(time.Time); ok {
			createdEntityMap["Created"] = created.Format(time.RFC3339)
		} else {
			createdEntityMap["Created"] = fmt.Sprintf("%v", node.Props["Created"])
		}

		if terminated != nil {
			if term, ok := node.Props["Terminated"].(time.Time); ok {
				createdEntityMap["Terminated"] = term.Format(time.RFC3339)
			} else {
				createdEntityMap["Terminated"] = fmt.Sprintf("%v", *terminated)
			}
		} else {
			log.Printf("[neo4j_client.CreateGraphEntity] Terminated: %v", terminated)
		}
		log.Printf("[neo4j_client.CreateGraphEntity] created entity(retrieved-final): %v", createdEntityMap)
		return createdEntityMap, nil
	}

	log.Printf("[neo4j_client.CreateGraphEntity] failed to create entity")
	return nil, fmt.Errorf("[neo4j_client.CreateGraphEntity] failed to create entity")
}

// CreateRelationship creates a relationship between two entities
func (r *Neo4jRepository) CreateRelationship(ctx context.Context, entityID string, rel *pb.Relationship) (map[string]interface{}, error) {
	session := r.getSession(ctx)
	defer session.Close(ctx)

	// Check if relationship with this ID already exists
	relExistsQuery := `MATCH ()-[r {Id: $relationshipID}]->() RETURN r`
	relResult, err := session.Run(ctx, relExistsQuery, map[string]interface{}{
		"relationshipID": rel.Id,
	})
	if err != nil {
		log.Printf("[neo4j_client.CreateRelationship] error checking if relationship exists: %v", err)
		return nil, fmt.Errorf("error checking if relationship exists: %v", err)
	}
	if relResult.Next(ctx) {
		log.Printf("[neo4j_client.CreateRelationship] relationship with Id %s already exists", rel.Id)
		return nil, fmt.Errorf("relationship with Id %s already exists", rel.Id)
	}

	existsQuery := `MATCH (p {Id: $parentID}), (c {Id: $childID}) RETURN p, c`
	result, err := session.Run(ctx, existsQuery, map[string]interface{}{
		"parentID": entityID,
		"childID":  rel.RelatedEntityId,
	})
	if err != nil {
		log.Printf("[neo4j_client.CreateRelationship] error checking entities: %v", err)
		return nil, fmt.Errorf("error checking entities: %v", err)
	} else {
		log.Printf("[neo4j_client.CreateRelationship] existsQuery: %v", existsQuery)
	}
	if !result.Next(ctx) {
		log.Printf("[neo4j_client.CreateRelationship] either parent or child entity does not exist")
		return nil, fmt.Errorf("either parent or child entity does not exist")
	} else {
		log.Printf("[neo4j_client.CreateRelationship] either parent or child entity exist")
	}

	params := map[string]interface{}{
		"parentID":       entityID,
		"childID":        rel.RelatedEntityId,
		"relationshipID": rel.Id,
		"startDate":      rel.StartTime,
	}

	createQuery := `MATCH (p {Id: $parentID}), (c {Id: $childID})
					CREATE (p)-[r:` + rel.Name + ` {Id: $relationshipID, Created: datetime($startDate)`

	if rel.EndTime != "" {
		createQuery += `, Terminated: datetime($endDate)`
		params["endDate"] = rel.EndTime
	}

	createQuery += `}]->(c) RETURN r`

	result, err = session.Run(ctx, createQuery, params)
	if err != nil {
		log.Printf("[neo4j_client.CreateRelationship] error creating relationship: %v", err)
		return nil, fmt.Errorf("error creating relationship: %v", err)
	} else {
		log.Printf("[neo4j_client.CreateRelationship] createQuery: %v", createQuery)
		log.Printf("[neo4j_client.CreateRelationship] params: %v", params)
	}

	if result.Next(ctx) {
		createdRel, _ := result.Record().Get("r")
		relationship, ok := createdRel.(neo4j.Relationship)
		if !ok {
			log.Printf("[neo4j_client.CreateRelationship] failed to cast created relationship to neo4j.Relationship")
			return nil, fmt.Errorf("failed to cast created relationship to neo4j.Relationship")
		} else {
			log.Printf("[neo4j_client.CreateRelationship] created relationship: %v", createdRel)
		}

		relationshipMap := map[string]interface{}{
			"Id":               fmt.Sprintf("%v", relationship.Props["Id"]),
			"relationshipType": rel.Name,
		}

		// Handle date fields with proper formatting
		if created, ok := relationship.Props["Created"].(time.Time); ok {
			relationshipMap["Created"] = created.Format(time.RFC3339)
		} else {
			relationshipMap["Created"] = fmt.Sprintf("%v", relationship.Props["Created"])
		}

		if rel.EndTime != "" {
			if terminated, ok := relationship.Props["Terminated"].(time.Time); ok {
				relationshipMap["Terminated"] = terminated.Format(time.RFC3339)
			} else {
				relationshipMap["Terminated"] = fmt.Sprintf("%v", relationship.Props["Terminated"])
			}
		}

		log.Printf("[neo4j_client.CreateRelationship] created relationship: %v", relationshipMap)
		return relationshipMap, nil
	} else {
		log.Printf("[neo4j_client.CreateRelationship] failed to retrieve created relationship: %v", result)
	}

	return nil, fmt.Errorf("failed to retrieve created relationship")
}

// ReadGraphEntity retrieves an entity by its ID from the Neo4j database and returns it as a map.
func (r *Neo4jRepository) ReadGraphEntity(ctx context.Context, entityID string) (map[string]interface{}, error) {
	if entityID == "" {
		return nil, fmt.Errorf("entity Id cannot be empty")
	}

	// Open a session
	session := r.getSession(ctx)
	defer session.Close(ctx)

	// Cypher query to retrieve the entity with both Major and Minor kinds
	query := `
        MATCH (e {Id: $Id})
        RETURN labels(e)[0] AS MajorKind, e.MinorKind AS MinorKind, e.Id AS Id, e.Name AS Name, 
               toString(e.Created) AS Created, 
               CASE WHEN e.Terminated IS NOT NULL THEN toString(e.Terminated) ELSE NULL END AS Terminated
    `

	// Run the query
	result, err := session.Run(ctx, query, map[string]interface{}{"Id": entityID})
	if err != nil {
		log.Printf("[neo4j_client.ReadGraphEntity] error querying entity: %v", err)
		return nil, fmt.Errorf("error querying entity: %v", err)
	}

	// Process the result
	if result.Next(ctx) {
		record := result.Record()

		// Map the entity properties
		entity := map[string]interface{}{
			"Id":        fmt.Sprintf("%v", record.Values[2]), // e.Id
			"Name":      fmt.Sprintf("%v", record.Values[3]), // e.Name
			"Created":   fmt.Sprintf("%v", record.Values[4]), // e.Created
			"MajorKind": fmt.Sprintf("%v", record.Values[0]), // labels(e)[0]
			"MinorKind": fmt.Sprintf("%v", record.Values[1]), // e.MinorKind
		}

		// Add Terminated if it exists
		if terminatedVal, exists := record.Get("Terminated"); exists && terminatedVal != nil {
			entity["Terminated"] = fmt.Sprintf("%v", terminatedVal)
		}

		return entity, nil
	}

	// If no entity is found
	return nil, fmt.Errorf("entity with Id %s not found", entityID)
}

// ReadRelatedGraphEntityIds retrieves related relationships based on a given relationship type and timestamp
func (r *Neo4jRepository) ReadRelatedGraphEntityIds(ctx context.Context, entityID string, relationship string, ts string) ([]map[string]interface{}, error) {
	if entityID == "" {
		return nil, fmt.Errorf("entity Id cannot be empty")
	}

	session := r.getSession(ctx)
	defer session.Close(ctx)

	query := fmt.Sprintf(`
        MATCH (e {Id: $entityID})-[r:%s]->(related)
        WHERE r.Created <= datetime($ts) AND (r.Terminated IS NULL OR r.Terminated > datetime($ts))
        RETURN r.Id AS relationshipID, r.Created AS startTime, r.Terminated AS endTime, type(r) AS name, related.Id AS relatedEntityId
    `, relationship)

	result, err := session.Run(ctx, query, map[string]interface{}{
		"entityID": entityID,
		"ts":       ts,
	})
	if err != nil {
		log.Printf("[neo4j_client.ReadRelatedGraphEntityIds] error querying related entities: %v", err)
		return nil, fmt.Errorf("error querying related entities: %v", err)
	}

	var relationships []map[string]interface{}
	for result.Next(ctx) {
		record := result.Record()

		// Extract fields from the query result
		relationshipID, _ := record.Get("relationshipID")
		startTime, _ := record.Get("startTime")
		endTime, _ := record.Get("endTime")
		name, _ := record.Get("name")
		relatedEntityID, _ := record.Get("relatedEntityId")

		// Ensure the relationship ID exists
		if relationshipID == nil {
			continue
		}

		// Format datetime fields
		var formattedStartTime, formattedEndTime string
		if startTime != nil {
			if t, ok := startTime.(time.Time); ok {
				formattedStartTime = t.Format(time.RFC3339) // Format as ISO 8601
			} else {
				formattedStartTime = fmt.Sprintf("%v", startTime)
			}
		}
		if endTime != nil {
			if t, ok := endTime.(time.Time); ok {
				formattedEndTime = t.Format(time.RFC3339) // Format as ISO 8601
			} else {
				formattedEndTime = fmt.Sprintf("%v", endTime)
			}
		}

		// Populate the relationship map
		relationship := map[string]interface{}{
			"Id":              fmt.Sprintf("%v", relationshipID),
			"StartTime":       formattedStartTime,
			"EndTime":         formattedEndTime,
			"Name":            fmt.Sprintf("%v", name),
			"RelatedEntityId": fmt.Sprintf("%v", relatedEntityID),
		}

		relationships = append(relationships, relationship)
	}

	if err := result.Err(); err != nil {
		log.Printf("[neo4j_client.ReadRelatedGraphEntityIds] error iterating over query result: %v", err)
		return nil, fmt.Errorf("error iterating over query result: %v", err)
	}

	return relationships, nil
}

func (r *Neo4jRepository) ReadRelationships(ctx context.Context, entityID string) ([]map[string]interface{}, error) {

	if entityID == "" {
		return nil, fmt.Errorf("entity Id cannot be empty")
	}

	// Open session
	session := r.getSession(ctx)
	defer session.Close(ctx)

	// Cypher query to get all relationships (incoming and outgoing)
	query := `
        MATCH (e {Id: $entityID})-[r]->(related)
        RETURN type(r) AS type, related.Id AS relatedID, "OUTGOING" AS direction, 
               toString(r.Created) AS Created, 
               CASE WHEN r.Terminated IS NOT NULL THEN toString(r.Terminated) ELSE NULL END AS Terminated,
               r.Id AS relationshipID
        UNION
        MATCH (e {Id: $entityID})<-[r]-(related)
        RETURN type(r) AS type, related.Id AS relatedID, "INCOMING" AS direction, 
               toString(r.Created) AS Created, 
               CASE WHEN r.Terminated IS NOT NULL THEN toString(r.Terminated) ELSE NULL END AS Terminated,
               r.Id AS relationshipID
    `

	// Run the query
	result, err := session.Run(ctx, query, map[string]interface{}{
		"entityID": entityID,
	})
	if err != nil {
		log.Printf("[neo4j_client.ReadRelationships] error querying relationships: %v", err)
		return nil, fmt.Errorf("error querying relationships: %v", err)
	}

	// Process results
	var relationships []map[string]interface{}
	for result.Next(ctx) {
		record := result.Record()
		values := record.Values

		// Ensure expected values exist
		if len(values) < 6 {
			continue
		}

		// Relationship structure
		rel := map[string]interface{}{
			"type":           fmt.Sprintf("%v", values[0]), // Relationship type
			"relatedID":      fmt.Sprintf("%v", values[1]),
			"direction":      fmt.Sprintf("%v", values[2]), // "INCOMING" or "OUTGOING"
			"Created":        fmt.Sprintf("%v", values[3]),
			"relationshipID": fmt.Sprintf("%v", values[5]), // Relationship ID
		}

		// Optional Terminated
		if values[4] != nil {
			rel["Terminated"] = fmt.Sprintf("%v", values[4])
		}

		relationships = append(relationships, rel)
	}

	// Return relationships as a map
	return relationships, nil
}

func (r *Neo4jRepository) ReadRelationship(ctx context.Context, relationshipID string) (map[string]interface{}, error) {

	if relationshipID == "" {
		return nil, fmt.Errorf("relationship Id cannot be empty")
	}

	session := r.getSession(ctx)
	defer session.Close(ctx)

	// Cypher query to find the relationship by its ID
	query := `
        MATCH ()-[r]->()
        WHERE r.Id = $relationshipID
        RETURN type(r) AS type, startNode(r).Id AS startEntityID, endNode(r).Id AS endEntityID, 
               toString(r.Created) AS Created, 
               CASE WHEN r.Terminated IS NOT NULL THEN toString(r.Terminated) ELSE NULL END AS Terminated, 
               r.Id AS relationshipID
    `

	// Run the query to fetch the relationship
	result, err := session.Run(ctx, query, map[string]interface{}{
		"relationshipID": relationshipID,
	})
	if err != nil {
		log.Printf("[neo4j_client.ReadRelationship] error querying relationship: %v", err)
		return nil, fmt.Errorf("error querying relationship: %v", err)
	}

	// Process results
	if result.Next(ctx) {
		record := result.Record()
		values := record.Values

		// Ensure expected values exist
		if len(values) < 6 {
			log.Printf("[neo4j_client.ReadRelationship] unexpected data format for relationship")
			return nil, fmt.Errorf("unexpected data format for relationship")
		}

		// Map to hold the relationship data
		relationship := map[string]interface{}{
			"type":           fmt.Sprintf("%v", values[0]), // Relationship type
			"startEntityID":  fmt.Sprintf("%v", values[1]), // ID of the start entity
			"endEntityID":    fmt.Sprintf("%v", values[2]), // ID of the end entity
			"Created":        fmt.Sprintf("%v", values[3]),
			"relationshipID": fmt.Sprintf("%v", values[5]), // Relationship ID
		}

		// Optional Terminated
		if values[4] != nil {
			relationship["Terminated"] = fmt.Sprintf("%v", values[4])
		}

		// Return the relationship data as a map
		return relationship, nil
	}

	// If no relationship was found
	return nil, fmt.Errorf("relationship with Id %s not found", relationshipID)
}

// UpdateGraphEntity updates the properties of an existing entity
func (r *Neo4jRepository) UpdateGraphEntity(ctx context.Context, id string, updateData map[string]interface{}) (map[string]interface{}, error) {
	if id == "" {
		return nil, fmt.Errorf("entity Id cannot be empty")
	}

	// Prepare update parameters
	params := map[string]interface{}{
		"Id": id,
	}

	// Open session
	session := r.getSession(ctx)
	defer session.Close(ctx)

	// Check if the entity exists
	existsQuery := `MATCH (e {Id: $Id}) RETURN e`
	result, err := session.Run(ctx, existsQuery, params)
	if err != nil {
		log.Printf("[neo4j_client.UpdateGraphEntity] error checking if entity exists: %v", err)
		return nil, fmt.Errorf("error checking if entity exists: %v", err)
	}

	if !result.Next(ctx) {
		log.Printf("[neo4j_client.UpdateGraphEntity] entity with Id %s does not exist", id)
		return nil, fmt.Errorf("entity with Id %s does not exist", id)
	}

	// Build Cypher query for updating entity
	query := `
        MATCH (e {Id: $Id})
    `

	// Add `Name` if provided
	if name, exists := updateData["Name"]; exists {
		params["Name"] = name
		query += `SET e.Name = $Name `
	}

	// Add `Terminated` if provided
	if terminated, exists := updateData["Terminated"]; exists {
		params["Terminated"] = terminated
		query += `SET e.Terminated = datetime($Terminated) `
	}

	// Execute update query and return updated entity
	query += ` RETURN e`

	result, err = session.Run(ctx, query, params)
	if err != nil {
		log.Printf("[neo4j_client.UpdateGraphEntity] error updating entity: %v", err)
		return nil, fmt.Errorf("error updating entity: %v", err)
	}

	// Retrieve updated entity
	if result.Next(ctx) {
		node, ok := result.Record().Get("e")
		if !ok {
			log.Printf("[neo4j_client.UpdateGraphEntity] unexpected error retrieving entity")
			return nil, fmt.Errorf("unexpected error retrieving entity")
		}

		// Convert node properties to map
		entityNode := node.(neo4j.Node)
		updatedEntity := make(map[string]interface{})
		for key, value := range entityNode.Props {
			if key == "Created" || key == "Terminated" {
				if timeValue, ok := value.(time.Time); ok {
					updatedEntity[key] = timeValue.Format(time.RFC3339)
				} else {
					updatedEntity[key] = fmt.Sprintf("%v", value)
				}
			} else {
				updatedEntity[key] = fmt.Sprintf("%v", value)
			}
		}

		return updatedEntity, nil
	}

	return nil, fmt.Errorf("failed to retrieve updated entity")
}

func (r *Neo4jRepository) UpdateRelationship(ctx context.Context, relationshipID string, updateData map[string]interface{}) (map[string]interface{}, error) {
	log.Printf("[neo4j_client.UpdateRelationship] Updating relationship %s with data: %+v", relationshipID, updateData)

	if relationshipID == "" {
		log.Printf("[neo4j_client.UpdateRelationship] relationship Id cannot be empty")
		return nil, fmt.Errorf("relationship Id cannot be empty")
	}

	// Prepare update parameters
	params := map[string]interface{}{
		"relationshipID": relationshipID,
	}

	// Open session
	session := r.getSession(ctx)
	defer session.Close(ctx)

	// Check if the relationship exists
	existsQuery := `MATCH ()-[r {Id: $relationshipID}]->() RETURN r`
	result, err := session.Run(ctx, existsQuery, params)
	if err != nil {
		log.Printf("[neo4j_client.UpdateRelationship] error checking if relationship exists: %v", err)
		return nil, fmt.Errorf("error checking if relationship exists: %v", err)
	}

	if !result.Next(ctx) {
		log.Printf("[neo4j_client.UpdateRelationship] relationship with Id %s does not exist", relationshipID)
		return nil, fmt.Errorf("relationship with Id %s does not exist", relationshipID)
	}

	// Build Cypher query for updating relationship
	query := `
        MATCH ()-[r {Id: $relationshipID}]->()
    `

	// Track if we have any fields to update
	hasUpdates := false

	// Add `Created` if provided
	if created, exists := updateData["Created"]; exists {
		params["Created"] = created
		if hasUpdates {
			query += `, r.Created = datetime($Created) `
		} else {
			query += `SET r.Created = datetime($Created) `
			hasUpdates = true
		}
	}

	// Add `Terminated` if provided
	if terminated, exists := updateData["Terminated"]; exists {
		params["Terminated"] = terminated
		if hasUpdates {
			query += `, r.Terminated = datetime($Terminated) `
		} else {
			query += `SET r.Terminated = datetime($Terminated) `
			hasUpdates = true
		}
	}

	// Check for any unsupported fields
	for key := range updateData {
		if key != "Created" && key != "Terminated" {
			log.Printf("[neo4j_client.UpdateRelationship] unsupported field '%s' provided for update", key)
			return nil, fmt.Errorf("unsupported field '%s' for relationship update. Only 'Created' and 'Terminated' are allowed", key)
		}
	}

	// If no fields to update, return error
	if !hasUpdates {
		log.Printf("[neo4j_client.UpdateRelationship] no valid fields provided for update")
		return nil, fmt.Errorf("no valid fields provided for update")
	}

	query += `RETURN r`

	// Execute update query and return updated relationship
	result, err = session.Run(ctx, query, params)
	if err != nil {
		log.Printf("[neo4j_client.UpdateRelationship] error updating relationship: %v", err)
		return nil, fmt.Errorf("error updating relationship: %v", err)
	}

	// Retrieve updated relationship
	if result.Next(ctx) {
		rel, ok := result.Record().Get("r")
		if !ok {
			log.Printf("[neo4j_client.UpdateRelationship] unexpected error retrieving relationship")
			return nil, fmt.Errorf("unexpected error retrieving relationship")
		}

		// Convert relationship properties to map with string values
		relationship := rel.(neo4j.Relationship)
		updatedRelationship := make(map[string]interface{})
		for key, value := range relationship.Props {
			if key == "Created" || key == "Terminated" {
				if timeValue, ok := value.(time.Time); ok {
					updatedRelationship[key] = timeValue.Format(time.RFC3339)
				} else {
					updatedRelationship[key] = fmt.Sprintf("%v", value)
				}
			} else {
				updatedRelationship[key] = fmt.Sprintf("%v", value)
			}
		}

		return updatedRelationship, nil
	}

	return nil, fmt.Errorf("failed to retrieve updated relationship")
}

func (r *Neo4jRepository) DeleteRelationship(ctx context.Context, relationshipID string) error {
	if relationshipID == "" {
		return fmt.Errorf("entity Id cannot be empty")
	}

	// Prepare query parameters
	params := map[string]interface{}{
		"relationshipID": relationshipID,
	}

	// Open session
	session := r.getSession(ctx)
	defer session.Close(ctx)

	// Check if the relationship exists
	query := `MATCH ()-[r {Id: $relationshipID}]->() RETURN r`
	result, err := session.Run(ctx, query, params)
	if err != nil {
		log.Printf("[neo4j_client.DeleteRelationship] error checking if relationship exists: %v", err)
		return fmt.Errorf("error checking if relationship exists: %v", err)
	}

	// If no relationship is found, return an error
	if !result.Next(ctx) {
		log.Printf("[neo4j_client.DeleteRelationship] relationship with Id %s does not exist", relationshipID)
		return fmt.Errorf("relationship with Id %s does not exist", relationshipID)
	}

	// Delete the relationship
	deleteQuery := `MATCH ()-[r {Id: $relationshipID}]->() DELETE r`
	_, err = session.Run(ctx, deleteQuery, params)
	if err != nil {
		log.Printf("[neo4j_client.DeleteRelationship] error deleting relationship: %v", err)
		return fmt.Errorf("error deleting relationship: %v", err)
	}

	return nil
}

// DeleteGraphEntity deletes an entity by its ID
func (r *Neo4jRepository) DeleteGraphEntity(ctx context.Context, entityID string) error {
	if entityID == "" {
		log.Printf("[neo4j_client.DeleteGraphEntity] entity Id cannot be empty")
		return fmt.Errorf("entity Id cannot be empty")
	}

	session := r.getSession(ctx)
	defer session.Close(ctx)

	query := `MATCH (e {Id: $entityID}) RETURN e`
	params := map[string]interface{}{
		"entityID": entityID,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		log.Printf("[neo4j_client.DeleteGraphEntity] error checking if entity exists: %v", err)
		return fmt.Errorf("error checking if entity exists: %v", err)
	}

	if !result.Next(ctx) {
		log.Printf("[neo4j_client.DeleteGraphEntity] entity with Id %s does not exist", entityID)
		return fmt.Errorf("entity with Id %s does not exist", entityID)
	}

	// Get the relationships of the entity
	relationships, err := r.ReadRelationships(ctx, entityID)
	if err != nil {
		log.Printf("[neo4j_client.DeleteGraphEntity] error getting relationships: %v", err)
		return fmt.Errorf("error getting relationships: %v", err)
	}

	// If there are relationships, return an error with relationship details
	if len(relationships) > 0 {
		log.Printf("[neo4j_client.DeleteGraphEntity] entity has relationships and cannot be deleted. Relationships: %v", relationships)
		return fmt.Errorf("entity has relationships and cannot be deleted. Relationships: %v", relationships)
	}

	// Delete the entity (node) with the given Id
	deleteQuery := `MATCH (e {Id: $entityID}) DELETE e`
	_, err = session.Run(ctx, deleteQuery, params)
	if err != nil {
		log.Printf("[neo4j_client.DeleteGraphEntity] error deleting entity: %v", err)
		return fmt.Errorf("error deleting entity: %v", err)
	}

	return nil
}

func (r *Neo4jRepository) FilterEntities(ctx context.Context, kind *pb.Kind, filters map[string]interface{}) ([]map[string]interface{}, error) {
	// Open a session
	session := r.getSession(ctx)
	defer session.Close(ctx)

	var query string
	var params map[string]interface{}

	// If we have an ID filter, use a simpler query
	if id, ok := filters["id"].(string); ok && id != "" {
		query = `
			MATCH (e {Id: $id})
			RETURN e.Id AS id, labels(e)[0] AS kind, 
				   toString(e.Created) AS created, 
				   CASE WHEN e.Terminated IS NOT NULL THEN toString(e.Terminated) ELSE NULL END AS terminated, 
				   e.Name AS name, 
				   e.MinorKind AS minorKind
		`
		params = map[string]interface{}{
			"id": id,
		}
	} else {
		// Original query for other filters
		if kind == nil || kind.Major == "" {
			return nil, fmt.Errorf("kind.Major is required")
		}

		// Start building the Cypher query
		query = `MATCH (e:` + kind.Major + `) WHERE 1=1 ` // Use kind.Major as the label
		params = map[string]interface{}{}

		// Add MinorKind filter if provided
		if kind.Minor != "" {
			query += `AND e.MinorKind = $minorKind `
			params["minorKind"] = kind.Minor
		}

		// Add optional filters
		if created, ok := filters["created"].(string); ok && created != "" {
			query += `AND e.Created = datetime($created) `
			params["created"] = created
		}
		if terminated, ok := filters["terminated"].(string); ok && terminated != "" {
			query += `AND e.Terminated = datetime($terminated) `
			params["terminated"] = terminated
		}

		if name, ok := filters["name"].(string); ok && name != "" {
			// Build regex pattern for case-insensitive partial match
			// Escape special regex characters in the name to prevent regex injection
			namePattern := "(?i).*" + name + ".*"
			query += `AND e.Name =~ $namePattern `
			params["namePattern"] = namePattern
		}

		// Return the matched entities
		query += `
			RETURN e.Id AS id, labels(e)[0] AS kind, 
				   toString(e.Created) AS created, 
				   CASE WHEN e.Terminated IS NOT NULL THEN toString(e.Terminated) ELSE NULL END AS terminated, 
				   e.Name AS name, 
				   e.MinorKind AS minorKind
		`
	}

	// Run the query
	result, err := session.Run(ctx, query, params)
	if err != nil {
		log.Printf("[neo4j_client.FilterEntities] error querying entities: %v", err)
		return nil, fmt.Errorf("error querying entities: %v", err)
	}

	// Process the results
	var entities []map[string]interface{}
	for result.Next(ctx) {
		record := result.Record()

		entity := map[string]interface{}{
			"id":         record.Values[0], // e.Id
			"kind":       record.Values[1], // labels(e)[0]
			"created":    record.Values[2], // e.Created
			"terminated": record.Values[3], // e.Terminated
			"name":       record.Values[4], // e.Name
			"minorKind":  record.Values[5], // e.MinorKind
		}

		entities = append(entities, entity)
	}

	// Check for errors during iteration
	if err := result.Err(); err != nil {
		log.Printf("[neo4j_client.FilterEntities] error iterating over query results: %v", err)
		return nil, fmt.Errorf("error iterating over query results: %v", err)
	}

	return entities, nil
}

// ReadFilteredRelationships retrieves relationships for an entity based on provided filters
func (r *Neo4jRepository) ReadFilteredRelationships(ctx context.Context, entityID string, relationshipFilters map[string]interface{}, activeAt string) ([]map[string]interface{}, error) {
	if entityID == "" {
		return nil, fmt.Errorf("entity Id cannot be empty")
	}

	session := r.getSession(ctx)
	defer session.Close(ctx)

	// Build the base query parts
	var outgoingQuery, incomingQuery string
	var params map[string]interface{}

	// Initialize parameters with entityID
	params = map[string]interface{}{
		"entityID": entityID,
	}

	// Build outgoing relationships query
	outgoingQuery = `
		MATCH (e {Id: $entityID})-[r]->(related)
		WHERE 1=1
	`

	// Build incoming relationships query
	incomingQuery = `
		MATCH (e {Id: $entityID})<-[r]-(related)
		WHERE 1=1
	`

	// Add relationship filters
	paramIndex := 1
	if id, ok := relationshipFilters["id"].(string); ok && id != "" {
		paramName := fmt.Sprintf("id%d", paramIndex)
		params[paramName] = id
		outgoingQuery += fmt.Sprintf(` AND r.Id = $%s`, paramName)
		incomingQuery += fmt.Sprintf(` AND r.Id = $%s`, paramName)
		paramIndex++
	}
	if relatedEntityId, ok := relationshipFilters["relatedEntityId"].(string); ok && relatedEntityId != "" {
		paramName := fmt.Sprintf("relatedEntityId%d", paramIndex)
		params[paramName] = relatedEntityId
		outgoingQuery += fmt.Sprintf(` AND related.Id = $%s`, paramName)
		incomingQuery += fmt.Sprintf(` AND related.Id = $%s`, paramName)
		paramIndex++
	}

	if name, ok := relationshipFilters["name"].(string); ok && name != "" {
		paramName := fmt.Sprintf("relationshipName%d", paramIndex)
		params[paramName] = name
		outgoingQuery += fmt.Sprintf(` AND type(r) = $%s`, paramName)
		incomingQuery += fmt.Sprintf(` AND type(r) = $%s`, paramName)
		paramIndex++
	}

	if startTime, ok := relationshipFilters["startTime"].(string); ok && startTime != "" {
		paramName := fmt.Sprintf("startTime%d", paramIndex)
		params[paramName] = startTime
		outgoingQuery += fmt.Sprintf(` AND r.Created = datetime($%s)`, paramName)
		incomingQuery += fmt.Sprintf(` AND r.Created = datetime($%s)`, paramName)
		paramIndex++
	}

	if endTime, ok := relationshipFilters["endTime"].(string); ok && endTime != "" {
		paramName := fmt.Sprintf("endTime%d", paramIndex)
		params[paramName] = endTime
		outgoingQuery += fmt.Sprintf(` AND r.Terminated = datetime($%s)`, paramName)
		incomingQuery += fmt.Sprintf(` AND r.Terminated = datetime($%s)`, paramName)
		paramIndex++
	}

	// Add activeAt filter if provided
	if activeAt != "" {
		paramName := fmt.Sprintf("activeAt%d", paramIndex)
		params[paramName] = activeAt
		activeAtCondition := fmt.Sprintf(` AND r.Created <= datetime($%s) AND (r.Terminated IS NULL OR r.Terminated > datetime($%s))`, paramName, paramName)
		outgoingQuery += activeAtCondition
		incomingQuery += activeAtCondition
		paramIndex++
	}

	// Add return clause
	outgoingQuery += `
		RETURN r.Id AS relationshipID, type(r) AS name, related.Id AS relatedEntityId, 
		       toString(r.Created) AS startTime, 
		       CASE WHEN r.Terminated IS NOT NULL THEN toString(r.Terminated) ELSE NULL END AS endTime,
		       "OUTGOING" AS direction
	`

	incomingQuery += `
		RETURN r.Id AS relationshipID, type(r) AS name, related.Id AS relatedEntityId, 
		       toString(r.Created) AS startTime, 
		       CASE WHEN r.Terminated IS NOT NULL THEN toString(r.Terminated) ELSE NULL END AS endTime,
		       "INCOMING" AS direction
	`

	// Determine which queries to run based on direction filter
	direction, hasDirection := relationshipFilters["direction"].(string)
	var finalQuery string

	if hasDirection {
		switch direction {
		case "OUTGOING":
			finalQuery = outgoingQuery
		case "INCOMING":
			finalQuery = incomingQuery
		default:
			// If direction is not recognized, return both
			finalQuery = outgoingQuery + " UNION " + incomingQuery
		}
	} else {
		// No direction specified, get both outgoing and incoming
		finalQuery = outgoingQuery + " UNION " + incomingQuery
	}

	// Execute the query
	result, err := session.Run(ctx, finalQuery, params)
	if err != nil {
		log.Printf("[neo4j_client.ReadFilteredRelationships] error querying relationships: %v", err)
		return nil, fmt.Errorf("error querying relationships: %v", err)
	}

	// Process results
	var relationships []map[string]interface{}
	for result.Next(ctx) {
		record := result.Record()

		// Extract fields from the query result
		relationshipID, _ := record.Get("relationshipID")
		name, _ := record.Get("name")
		relatedEntityID, _ := record.Get("relatedEntityId")
		startTime, _ := record.Get("startTime")
		endTime, _ := record.Get("endTime")
		direction, _ := record.Get("direction")

		// Ensure the relationship ID exists
		if relationshipID == nil {
			continue
		}

		// Format datetime fields
		var formattedStartTime, formattedEndTime string
		if startTime != nil {
			if t, ok := startTime.(time.Time); ok {
				formattedStartTime = t.Format(time.RFC3339)
			} else {
				formattedStartTime = fmt.Sprintf("%v", startTime)
			}
		}
		if endTime != nil {
			if t, ok := endTime.(time.Time); ok {
				formattedEndTime = t.Format(time.RFC3339)
			} else {
				formattedEndTime = fmt.Sprintf("%v", endTime)
			}
		}

		// Populate the relationship map
		relationship := map[string]interface{}{
			"id":              fmt.Sprintf("%v", relationshipID),
			"name":            fmt.Sprintf("%v", name),
			"relatedEntityId": fmt.Sprintf("%v", relatedEntityID),
			"startTime":       formattedStartTime,
			"endTime":         formattedEndTime,
			"direction":       fmt.Sprintf("%v", direction),
		}

		relationships = append(relationships, relationship)
	}

	if err := result.Err(); err != nil {
		log.Printf("[neo4j_client.ReadFilteredRelationships] error iterating over query result: %v", err)
		return nil, fmt.Errorf("error iterating over query result: %v", err)
	}

	return relationships, nil
}
