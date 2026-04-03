// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

import ballerina/io;
import ballerina/test;
import ballerina/protobuf.types.'any as pbAny;
import ballerina/http;
import ballerina/os;

// Get environment variables without fallback values
string testIngestionHostname = os:getEnv("INGESTION_SERVICE_HOST");
string testIngestionPort = os:getEnv("INGESTION_SERVICE_PORT");

// Try to get complete CORE service URL from environment variable first
string? coreServiceUrlEnv = os:getEnv("CORE_SERVICE_URL");
string testCoreServiceUrl = coreServiceUrlEnv ?: "http://0.0.0.0:50051";
// Construct URLs using string concatenation
string testIngestionServiceUrl = "http://" + testIngestionHostname + ":" + testIngestionPort;

type JsonObject map<anydata>;

// Before Suite Function
@test:BeforeSuite
function beforeSuiteFunc() {
    io:println("I'm the before suite function!");
    io:println("CORE Service URL: " + testCoreServiceUrl);
    io:println("Ingestion Service URL: " + testIngestionServiceUrl);
}

// After Suite Function
@test:AfterSuite
function afterSuiteFunc() {
    io:println("I'm the after suite function!");
}

// Helper function to unpack Any values to strings
function unwrapAny(pbAny:Any anyValue) returns string|error {
    return pbAny:unpack(anyValue, string);
}

// Helper function to unpack Any values to JSON/Struct
function unwrapAnyToJson(pbAny:Any anyValue) returns json|error {
    // Try to unpack as string first
    string|error stringValue = pbAny:unpack(anyValue, string);
    if stringValue is string {
        return stringValue;
    }
    
    // If string unpacking fails, return the string representation as a fallback
    return anyValue.toString();
}

// Helper function to verify tabular data content
function verifyTabularData(json actual, json expected) {
    // Simple string comparison
    test:assertEquals(actual.toString(), expected.toString(), "Data JSON should match");
}

// Helper function to convert JSON to protobuf Any value
function jsonToAny(json data) returns pbAny:Any|error {
    if data is int {
        // For integer values
        map<json> structMap = {
            "value": data
        };
        return pbAny:pack(structMap);
    } else if data is float {
        // For float values
        map<json> structMap = {
            "value": data
        };
        return pbAny:pack(structMap);
    } else if data is string {
        // For string values
        map<json> structMap = {
            "value": data
        };
        return pbAny:pack(structMap);
    } else if data is boolean {
        // For boolean values
        map<json> structMap = {
            "value": data
        };
        return pbAny:pack(structMap);
    } else if data is () {
        // For null values
        map<json> structMap = {
            "null_value": ()
        };
        return pbAny:pack(structMap);
    } else if data is json[] {
        // For arrays, wrap in a list_value structure
        map<json> structMap = {
            "values": data
        };
        return pbAny:pack(structMap);
    } else if data is map<json> {
        // For objects, pack directly as structured data instead of converting to string
        return pbAny:pack(data);
    } else {
        return error("Unsupported data type: " + data.toString());
    }
}

@test:Config {}
function testMetadataHandling() returns error? {
    // Initialize the client
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Test data setup
    string testId = "test-entity-1";
    string expectedValue1 = "value1";
    string expectedValue2 = "value2";
    
    // Create the metadata array
    record {| string key; pbAny:Any value; |}[] metadataArray = [];

    // Pack string values into protobuf.Any directly
    pbAny:Any packedValue1 = check pbAny:pack(expectedValue1);
    pbAny:Any packedValue2 = check pbAny:pack(expectedValue2);

    // Add packed values to the metadata array
    metadataArray.push({key: "key1", value: packedValue1});
    metadataArray.push({key: "key2", value: packedValue2});

    // Create entity request
    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "metadata"
        },
        created: "2023-01-01",
        terminated: "",
        name: {
            startTime: "2023-01-01",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: metadataArray
    };

    // Create entity
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    io:println("Created entity metadata: ", createEntityResponse.metadata);
    
    // Read entity
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata"]
    };
    io:println("ReadEntityRequest: ", readEntityRequest);
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    io:println("Entity retrieved, verifying data...");
    io:println("Retrieved entity: ", readEntityResponse);
    io:println("Retrieved entity metadata: ", readEntityResponse.metadata);
    
    // Verify metadata values
    map<string> actualValues = {};
    foreach var item in readEntityResponse.metadata {
        string|error unwrapped = unwrapAny(item.value);
        if unwrapped is string {
            actualValues[item.key] = unwrapped.trim();
        } else {
            test:assertFail("Failed to unpack metadata value for key: " + item.key);
        }
    }
    
    // Assert the values match
    test:assertEquals(actualValues["key1"], expectedValue1, "Metadata value for key1 doesn't match");
    test:assertEquals(actualValues["key2"], expectedValue2, "Metadata value for key2 doesn't match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity deleted");
    
    return;
}

// TODO: Re-enable once the Result type response handling is added
// See: https://github.com/LDFLK/nexoan/issues/23
@test:Config {
    enable: false
}
function testMetadataUnpackError() returns error? {
    // Test case to verify handling of non-existent entities
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Try to read a non-existent entity
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: "non-existent-entity",
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata"]
    };
    Entity|error response = ep->ReadEntity(readEntityRequest);
    
    // Assert that we get an error for non-existent entity
    test:assertTrue(response is error, "Expected error for non-existent entity");
    
    return;
}

@test:Config {}
function testMetadataUpdating() returns error? {
    // Initialize the client
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Test data setup
    string testId = "test-entity-update";
    
    // Initial metadata values
    string initialValue1 = "initial-value1";
    string initialValue2 = "initial-value2";
    
    // Updated metadata values
    string updatedValue1 = "updated-value1";
    string updatedValue2 = "updated-value2";
    string newValue3 = "new-value3";
    
    // Create the initial metadata array
    record {| string key; pbAny:Any value; |}[] initialMetadataArray = [];
    pbAny:Any packedInitialValue1 = check pbAny:pack(initialValue1);
    pbAny:Any packedInitialValue2 = check pbAny:pack(initialValue2);
    initialMetadataArray.push({key: "key1", value: packedInitialValue1});
    initialMetadataArray.push({key: "key2", value: packedInitialValue2});

    // Create initial entity request
    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "update-metadata"
        },
        created: "2023-01-01",
        terminated: "",
        name: {
            startTime: "2023-01-01",
            endTime: "",
            value: check pbAny:pack("test-update-entity")
        },
        metadata: initialMetadataArray
    };

    // Create entity
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Verify initial metadata
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata"]
    };
    Entity initialReadResponse = check ep->ReadEntity(readEntityRequest);
    verifyMetadata(initialReadResponse.metadata, {"key1": initialValue1, "key2": initialValue2});
    io:println("Initial metadata verified");
    
    // Create updated metadata array
    record {| string key; pbAny:Any value; |}[] updatedMetadataArray = [];
    pbAny:Any packedUpdatedValue1 = check pbAny:pack(updatedValue1);
    pbAny:Any packedUpdatedValue2 = check pbAny:pack(updatedValue2);
    pbAny:Any packedNewValue3 = check pbAny:pack(newValue3);
    updatedMetadataArray.push({key: "key1", value: packedUpdatedValue1});
    updatedMetadataArray.push({key: "key2", value: packedUpdatedValue2});
    updatedMetadataArray.push({key: "key3", value: packedNewValue3});

    // Update entity request
    // Note: Don't include Kind, Created, or Name in update requests as they are immutable
    Entity updateEntityRequest = {
        id: testId,
        kind: {
            major: "",
            minor: ""
        },
        created: "2023-01-01",
        terminated: "",
        name: {
            startTime: "",
            endTime: "",
            value: check pbAny:pack("")
        },
        metadata: updatedMetadataArray,
        attributes: [],
        relationships: []
    };
    
    // Update entity
    UpdateEntityRequest updateRequest = {
        id: testId,
        entity: updateEntityRequest
    };
    Entity updateEntityResponse = check ep->UpdateEntity(updateRequest);
    io:println("Entity updated with ID: " + updateEntityResponse.id);
    
    // Verify updated metadata
    ReadEntityRequest updatedReadRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata"]
    };
    Entity updatedReadResponse = check ep->ReadEntity(updatedReadRequest);
    verifyMetadata(updatedReadResponse.metadata, {
        "key1": updatedValue1, 
        "key2": updatedValue2,
        "key3": newValue3
    });
    io:println("Updated metadata verified");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity deleted");
    
    return;
}

// Helper function to verify metadata
function verifyMetadata(record {| string key; pbAny:Any value; |}[] metadata, map<string> expected) {
    map<string> actual = {};
    foreach var item in metadata {
        string|error unwrapped = unwrapAny(item.value);
        if unwrapped is string {
            actual[item.key] = unwrapped.trim();
        }
    }
    
    // Verify all expected key-value pairs exist in the actual metadata
    foreach var [key, expectedValue] in expected.entries() {
        test:assertTrue(actual.hasKey(key), "Metadata key not found: " + key);
        test:assertEquals(actual[key] ?: "", expectedValue, 
            string `Metadata value for ${key} doesn't match: expected ${expectedValue}, got ${actual[key] ?: ""}`);
    }
}

@test:Config {}
function testEntityReading() returns error? {
    // Initialize the client
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Test data setup
    string testId = "test-entity-read";
    string metadataKey = "readTest";
    string metadataValue = "read-test-value";
    
    // Create a test entity first
    record {| string key; pbAny:Any value; |}[] metadataArray = [];
    pbAny:Any packedValue = check pbAny:pack(metadataValue);
    metadataArray.push({key: metadataKey, value: packedValue});
    
    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "read-entity"
        },
        created: "2023-01-01",
        terminated: "",
        name: {
            startTime: "2023-01-01",
            endTime: "",
            value: check pbAny:pack("test-read-entity")
        },
        metadata: metadataArray,
        attributes: [],
        relationships: []
    };
    
    // Create entity
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Test entity created with ID: " + createEntityResponse.id);
    io:println("Created entity metadata: ", createEntityResponse.metadata);
    
    // Read the entity
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata"]
    };
    io:println("ReadEntityRequest: ", readEntityRequest);
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    io:println("Entity retrieved, verifying data...");
    io:println("Retrieved entity: ", readEntityResponse);
    io:println("Retrieved entity metadata: ", readEntityResponse.metadata);
    
    // Verify entity fields
    test:assertEquals(readEntityResponse.id, testId, "Entity ID mismatch");
    
    // Verify metadata
    boolean foundMetadata = false;
    foreach var item in readEntityResponse.metadata {
        if item.key == metadataKey {
            string|error unwrapped = unwrapAny(item.value);
            if unwrapped is string {
                test:assertEquals(unwrapped.trim(), metadataValue, 
                    string `Metadata value mismatch: expected ${metadataValue}, got ${unwrapped}`);
                foundMetadata = true;
            }
        }
    }
    
    test:assertTrue(foundMetadata, "Expected metadata key not found");
    
    // Test reading non-existent entity
    string nonExistentId = "non-existent-entity-" + testId;
    ReadEntityRequest nonExistentRequest = {
        entity: {
            id: nonExistentId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata"]
    };
    Entity|error nonExistentResult = ep->ReadEntity(nonExistentRequest);
    
    // For now, expect an error for non-existent entities
    test:assertTrue(nonExistentResult is error, "Expected error for non-existent entity");
    if nonExistentResult is error {
        io:println("Non-existent entity correctly returned error: " + nonExistentResult.message());
    }
    
    // Assert that we get an error for non-existent entity
    // For non-existence entities, we send a response with an empty data
    // But once the Result API is integrated this can be tested. 
    // FIXME: https://github.com/LDFLK/nexoan/issues/23
    // test:assertTrue(nonExistentResponse is error, "Expected error for non-existent entity ID");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity deleted");
    
    return;
}

@test:Config {}
function testCreateMinimalGraphEntity() returns error? {
    // Initialize the client
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Test data setup - minimal entity with just required fields
    string testId = "test-minimal-entity";
    
    // Create entity request with only required fields - no metadata, attributes, or relationships
    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "minimal"
        },
        created: "2023-01-01",
        terminated: "",
        name: {
            startTime: "2023-01-01",
            endTime: "",
            value: check pbAny:pack("minimal-test-entity")
        },
        metadata: [],
        attributes: [],
        relationships: []
    };

    // Create entity
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Minimal entity created with ID: " + createEntityResponse.id);
    
    // Verify entity was created correctly
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Basic entity verification
    test:assertEquals(readEntityResponse.id, testId, "Entity ID doesn't match");
    test:assertEquals(readEntityResponse.kind.major, "test", "Entity kind.major doesn't match");
    test:assertEquals(readEntityResponse.kind.minor, "minimal", "Entity kind.minor doesn't match");
    
    // Verify empty collections
    test:assertEquals(readEntityResponse.metadata.length(), 0, "Metadata should be empty");
    test:assertEquals(readEntityResponse.attributes.length(), 0, "Attributes default value should be empty");
    test:assertEquals(readEntityResponse.relationships.length(), 0, "Relationships should be empty");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test minimal entity deleted");
    
    return;
}

@test:Config {}
function testCreateMinimalGraphEntityViaRest() returns error? {
    // Initialize an HTTP client for the REST API with HTTP/2 support
    http:ClientConfiguration httpConfig = {
        httpVersion: "2.0" // Enable HTTP/2
    };
    http:Client restClient = check new (testIngestionServiceUrl, httpConfig);
    
    // Test data setup - minimal JSON entity
    string testId = "test-minimal-json-entity";
    
    // Minimal JSON payload with required fields matching the Entity structure
    json minimalEntityJson = {
        "id": testId,
        "kind": {
            "major": "test",
            "minor": "minimal-json"
        },
        "created": "2023-01-01",
        "terminated": "",
        "name": {
            "startTime": "2023-01-01",
            "endTime": "",
            "value": "minimal-json-test-entity"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    };

    // Create entity via REST API
    http:Response|error response = restClient->post("/entities", minimalEntityJson);
    
    // Verify HTTP request was successful
    if response is error {
        test:assertFail("Failed to create entity via REST API: " + response.message());
    }
    
    http:Response httpResponse = <http:Response>response;
    test:assertEquals(httpResponse.statusCode, 201, "Expected 201 OK status code");
    
    // Parse response JSON
    json responseJson = check httpResponse.getJsonPayload();
    test:assertEquals(check responseJson.id, testId, "Entity ID in response doesn't match");
    
    // Initialize the gRPC client to verify entity was properly created
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Verify entity data
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata","attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Basic entity verification
    test:assertEquals(readEntityResponse.id, testId, "Entity ID doesn't match");
    test:assertEquals(readEntityResponse.kind.major, "test", "Entity kind.major doesn't match");
    test:assertEquals(readEntityResponse.kind.minor, "minimal-json", "Entity kind.minor doesn't match");
    
    // Verify empty collections
    test:assertEquals(readEntityResponse.metadata.length(), 0, "Metadata should be empty");
    test:assertEquals(readEntityResponse.attributes.length(), 0, "Attributes default value should be empty");
    test:assertEquals(readEntityResponse.relationships.length(), 0, "Relationships should be empty");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test minimal JSON entity deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "relationship"]
}
function testEntityWithRelationship() returns error? {
    // Test IDs for entities
    string sourceEntityId = "test-entity-with-relationship-source";
    string targetEntityId = "test-entity-with-relationship-target";
    
    // Initialize REST client with HTTP/2 support
    http:ClientConfiguration httpConfig = {
        httpVersion: "2.0" // Enable HTTP/2
    };
    http:Client restClient = check new (testIngestionServiceUrl, httpConfig);
    
    // Create source entity
    json sourceEntityJson = {
        "id": sourceEntityId,
        "kind": {
            "major": "test",
            "minor": "relationship-source"
        },
        "created": "2023-01-01",
        "terminated": "",
        "name": {
            "startTime": "2023-01-01",
            "endTime": "",
            "value": "source-entity"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    };
    
    // Create target entity
    json targetEntityJson = {
        "id": targetEntityId,
        "kind": {
            "major": "test",
            "minor": "relationship-target"
        },
        "created": "2023-01-01", 
        "terminated": "",
        "name": {
            "startTime": "2023-01-01",
            "endTime": "",
            "value": "target-entity"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    };
    
    // Create both entities via REST API
    http:Response|error sourceResponse = restClient->post("/entities", sourceEntityJson);
    http:Response|error targetResponse = restClient->post("/entities", targetEntityJson);
    
    // Verify HTTP requests were successful
    if sourceResponse is error {
        test:assertFail("Failed to create source entity: " + sourceResponse.message());
    }
    if targetResponse is error {
        test:assertFail("Failed to create target entity: " + targetResponse.message());
    }
    
    http:Response sourceHttpResponse = <http:Response>sourceResponse;
    http:Response targetHttpResponse = <http:Response>targetResponse;
    test:assertEquals(sourceHttpResponse.statusCode, 201, "Expected 201 status code for source entity");
    test:assertEquals(targetHttpResponse.statusCode, 201, "Expected 201 status code for target entity");
    
    // Create relationship between entities - include full entity structure
    string relationshipId = "rel-" + sourceEntityId + "-" + targetEntityId;
    json relationshipJson = {
        "id": sourceEntityId,
        "kind": {
        },
        "created": "",
        "terminated": "",
        "name": {
        },
        "metadata": [],
        "attributes": [],
        "relationships": {
            relationshipId: {
                "relatedEntityId": targetEntityId,
                "startTime": "2023-01-01",
                "endTime": "",
                "id": relationshipId,
                "name": "CONNECTS_TO"
            }
        }
    };
    
    // Update source entity with relationship
    http:Response|error updateResponse = restClient->put("/entities/" + sourceEntityId, relationshipJson);
    
    // Verify update was successful
    if updateResponse is error {
        test:assertFail("Failed to update entity with relationship: " + updateResponse.message());
    }
    
    http:Response updateHttpResponse = <http:Response>updateResponse;
    test:assertEquals(updateHttpResponse.statusCode, 200, "Expected 200 status code for relationship update");
    
    // Initialize the gRPC client to verify relationship was properly created
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Read source entity to verify relationship
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: sourceEntityId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify relationship data
    test:assertEquals(readEntityResponse.relationships.length(), 1, "Entity should have one relationship");
    
    // Find the relationship by iterating through the array
    Relationship? targetRelationship = ();
    foreach var rel in readEntityResponse.relationships {
        if rel.key == relationshipId {
            targetRelationship = rel.value;
            break;
        }
    }
    
    io:println("Target relationship: " + targetRelationship.toJsonString());
    test:assertFalse(targetRelationship is (), "Relationship with key 'CONNECTS_TO' not found");
    Relationship relationship = <Relationship>targetRelationship;
    test:assertEquals(relationship.relatedEntityId, targetEntityId, "Related entity ID doesn't match");
    test:assertEquals(relationship.name, "CONNECTS_TO", "Relationship name doesn't match");
    test:assertEquals(relationship.startTime, "2023-01-01T00:00:00Z", "Relationship start time doesn't match");
    test:assertEquals(relationship.id, relationshipId, "Relationship ID doesn't match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: sourceEntityId});
    Empty _ = check ep->DeleteEntity({id: targetEntityId});
    io:println("Test entities with relationship deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "graph", "only_nodes"],
    enable: false
}
function testEntityWithSimpleOnlyNodesGraphAttributes() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-simple-only-nodes-graph";
    
    // Initialize the gRPC client to verify entity
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Create entity with tabular data in attributes
    json socialNetworkGraph = {
        "nodes": [
            {"id": "user1", "type": "user", "properties": {"name": "Alice", "age": 30, "location": "NY"}},
            {"id": "user2", "type": "user", "properties": {"name": "Bob", "age": 25, "location": "SF"}},
            {"id": "user3", "type": "user", "properties": {"name": "Charlie", "age": 35, "location": "LA"}},
            {"id": "post1", "type": "post", "properties": {"title": "Hello", "content": "World", "created": "2024-03-20"}},
            {"id": "post2", "type": "post", "properties": {"title": "Graph", "content": "DB", "created": "2024-03-21"}}
        ]
    };

    // Convert JSON to protobuf Any values
    pbAny:Any socialNetworkGraphAny = check jsonToAny(socialNetworkGraph);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "graph"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "employee_salary_history",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: socialNetworkGraphAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with graph attributes deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "graph", "simple"],
    enable: false
}
function testEntityWithSimpleGraphAttributes() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-simple-entity-graph";
    
    // Initialize the gRPC client to verify entity
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Create entity with tabular data in attributes
    json socialNetworkGraph = {
        "nodes": [
            {"id": "user1", "type": "user", "properties": {"name": "Alice", "age": 30, "location": "NY"}},
            {"id": "user2", "type": "user", "properties": {"name": "Bob", "age": 25, "location": "SF"}},
            {"id": "user3", "type": "user", "properties": {"name": "Charlie", "age": 35, "location": "LA"}},
            {"id": "post1", "type": "post", "properties": {"title": "Hello", "content": "World", "created": "2024-03-20"}},
            {"id": "post2", "type": "post", "properties": {"title": "Graph", "content": "DB", "created": "2024-03-21"}}
        ],
        "edges": [
            {"source": "user1", "target": "user2", "type": "follows", "properties": {"since": "2024-01-01"}},
            {"source": "user2", "target": "user3", "type": "follows", "properties": {"since": "2024-02-01"}},
            {"source": "user1", "target": "post1", "type": "created", "properties": {"timestamp": "2024-03-20T10:00:00Z"}},
            {"source": "user2", "target": "post1", "type": "likes", "properties": {"timestamp": "2024-03-20T11:00:00Z"}},
            {"source": "user3", "target": "post2", "type": "created", "properties": {"timestamp": "2024-03-21T09:00:00Z"}}
        ]
    };

    // Convert JSON to protobuf Any values
    pbAny:Any socialNetworkGraphAny = check jsonToAny(socialNetworkGraph);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "graph"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "employee_salary_history",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: socialNetworkGraphAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up

    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with graph attributes deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "graph", "multi"],
    enable: false
}
function testEntityWithMultiGraphAttributes() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-graph";
    
    // Initialize the gRPC client to verify entity
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Create entity with tabular data in attributes
    json salaryGraph = {
        "nodes": [
					{"id": "user1", "type": "user", "properties": {"name": "Alice", "age": 30, "location": "NY"}},
					{"id": "user2", "type": "user", "properties": {"name": "Bob", "age": 25, "location": "SF"}},
					{"id": "user3", "type": "user", "properties": {"name": "Charlie", "age": 35, "location": "LA"}},
					{"id": "post1", "type": "post", "properties": {"title": "Hello", "content": "World", "created": "2024-03-20"}},
					{"id": "post2", "type": "post", "properties": {"title": "Graph", "content": "DB", "created": "2024-03-21"}}
				],
				"edges": [
					{"source": "user1", "target": "user2", "type": "follows", "properties": {"since": "2024-01-01"}},
					{"source": "user2", "target": "user3", "type": "follows", "properties": {"since": "2024-02-01"}},
					{"source": "user1", "target": "post1", "type": "created", "properties": {"timestamp": "2024-03-20T10:00:00Z"}},
					{"source": "user2", "target": "post1", "type": "likes", "properties": {"timestamp": "2024-03-20T11:00:00Z"}},
					{"source": "user3", "target": "post2", "type": "created", "properties": {"timestamp": "2024-03-21T09:00:00Z"}}
				]
    };

    json projectGraph = {
        "nodes": [
            {"id": "proj_redesign", "type": "project", "properties": {"id": "P001", "name": "System Redesign", "status": "active"}},
            {"id": "proj_migration", "type": "project", "properties": {"id": "P002", "name": "API Migration", "status": "completed"}},
            {"id": "proj_audit", "type": "project", "properties": {"id": "P003", "name": "Security Audit", "status": "completed"}},
            {"id": "role_lead", "type": "role", "properties": {"title": "Lead Developer", "level": "senior"}},
            {"id": "role_dev", "type": "role", "properties": {"title": "Developer", "level": "mid"}}
        ],
        "edges": [
            {"source": "proj_redesign", "target": "role_lead", "type": "has_role", "properties": {"start_date": "2024-01-01", "end_date": ""}},
            {"source": "proj_migration", "target": "role_dev", "type": "has_role", "properties": {"start_date": "2023-06-01", "end_date": "2023-12-31"}},
            {"source": "proj_audit", "target": "role_dev", "type": "has_role", "properties": {"start_date": "2023-01-01", "end_date": "2023-05-31"}},
            {"source": "proj_redesign", "target": "proj_migration", "type": "follows", "properties": {"transition_date": "2023-12-31"}},
            {"source": "proj_migration", "target": "proj_audit", "type": "follows", "properties": {"transition_date": "2023-05-31"}}
        ]
    };

    // Convert JSON to protobuf Any values
    pbAny:Any salaryGraphAny = check jsonToAny(salaryGraph);
    pbAny:Any projectGraphAny = check jsonToAny(projectGraph);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "graph"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "employee_salary_history",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: salaryGraphAny
                        }
                    ]
                }
            },
            {
                key: "project_assignments",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: projectGraphAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with graph attributes deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "list"],
    enable: false
}
function testEntityWithSimpleListAttributes() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-list";
    
    // Initialize the gRPC client to verify entity
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Create entity with list data in attributes
    json salaryList = {
        "values": [
            1,
            2,
            3
        ]
    };

    // Convert JSON to protobuf Any values
    pbAny:Any salaryListAny = check jsonToAny(salaryList);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "list"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "salary_history",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: salaryListAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with list attributes deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "list", "mixed"],
    enable: false
}
function testEntityWithMixedTypeListAttributes() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-mixed-list";
    
    // Initialize the gRPC client to verify entity
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Create entity with mixed type list data in attributes
    json mixedTypeList = {
        "values": [
            "Hello",
            42,
            true,
            "2024-03-21",
            "2024-03-21T15:30:00Z",
            null
        ]
    };

    // Convert JSON to protobuf Any values
    pbAny:Any mixedTypeListAny = check jsonToAny(mixedTypeList);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "mixed-list"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "mixed_type_history",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: mixedTypeListAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with mixed type list attributes deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "list", "empty"],
    enable: false
}
function testEntityWithEmptyListAttributes() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-empty-list";
    
    // Initialize the gRPC client to verify entity
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Create entity with empty list data in attributes
    json emptyList = {
        "values": []
    };

    // Convert JSON to protobuf Any values
    pbAny:Any emptyListAny = check jsonToAny(emptyList);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "empty-list"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "empty_list_history",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: emptyListAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with empty list attributes deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "map"],
    enable: false
}
function testEntityWithMapAttributes() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-map";
    
    // Initialize the gRPC client to verify entity
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Create entity with map data in attributes
    json userProfileMap = {
       "properties": {
					"name": "John",
					"age": 30,
					"active": true
				}
    };

    // Convert JSON to protobuf Any values
    pbAny:Any userProfileMapAny = check jsonToAny(userProfileMap);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "map"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "user_profile",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: userProfileMapAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with map attributes deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "map", "nested"],
    enable: false
}
function testEntityWithNestedMapAttributes() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-nested-map";
    
    // Initialize the gRPC client to verify entity
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Create entity with nested map data in attributes
    json nestedMap = {
        "organization": {
            "name": "Tech Corp",
            "departments": {
                "engineering": {
                    "head": "Alice Smith",
                    "size": 50,
                    "projects": {
                        "project1": {
                            "name": "System Redesign",
                            "status": "active",
                            "team": ["John", "Jane", "Bob"]
                        },
                        "project2": {
                            "name": "API Migration",
                            "status": "completed",
                            "team": ["Alice", "Charlie"]
                        }
                    }
                },
                "marketing": {
                    "head": "Bob Johnson",
                    "size": 20,
                    "campaigns": {
                        "campaign1": {
                            "name": "Summer Sale",
                            "budget": 50000,
                            "channels": ["social", "email", "web"]
                        }
                    }
                }
            },
            "metadata": {
                "founded": "2020-01-01",
                "location": {
                    "country": "USA",
                    "offices": ["NY", "SF", "LA"]
                }
            }
        }
    };

    // Convert JSON to protobuf Any values
    pbAny:Any nestedMapAny = check jsonToAny(nestedMap);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "nested-map"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "org_structure",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: nestedMapAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with nested map attributes deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "map", "empty"],
    enable: false
}
function testEntityWithEmptyMapValues() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-empty-map-values";
    
    // Initialize the gRPC client to verify entity
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Create entity with map data containing empty values
    // FIXME: https://github.com/LDFLK/nexoan/issues/137
    json emptyValuesMap = {
        "properties": {
            "empty_str": "",
            "zero": 0,
            "null_val": null
        }
    };

    // Convert JSON to protobuf Any values
    pbAny:Any emptyValuesMapAny = check jsonToAny(emptyValuesMap);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "empty-map-values"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "empty_values",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: emptyValuesMapAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with empty map values deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "map", "nested"],
    enable: false
}
function testEntityWithNestedMapValues() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-nested-map-values";
    
    // Initialize the gRPC client to verify entity
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Create entity with deeply nested map data
    json nestedMap = {
        "properties": {
            "user": {
                "profile": {
                    "personal": {
                        "name": "John Doe",
                        "age": 30,
                        "address": {
                            "street": "123 Main St",
                            "city": "New York",
                            "zip": "10001"
                        }
                    },
                    "preferences": {
                        "theme": "dark",
                        "notifications": true,
                        "language": "en"
                    }
                },
                "settings": {
                    "account": {
                        "type": "premium",
                        "status": "active",
                        "subscription": {
                            "plan": "yearly",
                            "start_date": "2024-01-01",
                            "end_date": "2024-12-31"
                        }
                    },
                    "security": {
                        "two_factor": true,
                        "last_login": "2024-03-21T10:00:00Z",
                        "devices": {
                            "mobile": ["iPhone", "iPad"],
                            "desktop": ["MacBook", "Windows PC"]
                        }
                    }
                }
            },
            "metadata": {
                "created_at": "2024-01-01T00:00:00Z",
                "updated_at": "2024-03-21T15:30:00Z",
                "version": {
                    "major": 1,
                    "minor": 0,
                    "patch": 0
                }
            }
        }
    };

    // Convert JSON to protobuf Any values
    pbAny:Any nestedMapAny = check jsonToAny(nestedMap);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "nested-map-values"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "nested_data",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: nestedMapAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with nested map values deleted");
    
    return;
}


@test:Config {}
function testEntityWithTabularAttributes() returns error? {
    // Initialize the client
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Test data setup
    string testId = "ID-MIN-A";
    
    // Create tabular data structure
    // TODO: https://github.com/LDFLK/nexoan/issues/284
    json tabularData = {
        "columns": ["id", "name", "age", "department", "salary"],
        "rows": [
            [1, "John Doe", 30, "Engineering", 75000.50],
            [2, "Jane Smith", 25, "Marketing", 65000],
            [3, "Bob Wilson", 35, "Sales", 85000.75],
            [4, "Alice Brown", 28, "Engineering", 70000.25],
            [5, "Charlie Davis", 32, "Finance", 80000]
        ]
    };

    // Convert JSON to protobuf Any values
    pbAny:Any tabularDataAny = check jsonToAny(tabularData);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "Organization",
            minor: "Minister"
        },
        created: "2025-02-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2025-02-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("Minister of Health")
        },
        metadata: [
            {
                key: "minister_of_health_metadata",
                value: check pbAny:pack("tabular_minister_of_health_test_value")
            }
        ],
        attributes: [
            {
                key: "employee_data",
                value: {
                    values: [
                        {
                            startTime: "2025-04-01T00:00:00Z",
                            endTime: "",
                            value: tabularDataAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    test:assertTrue(createEntityResponse.id != "", "Entity should be created successfully");
    test:assertEquals(createEntityResponse.id, testId, "Entity ID should match");

    json tabularDataFilter = {
        "columns": ["id", "name", "age", "department", "salary"],
        "rows": [[]]
    };

    pbAny:Any tabularDataFilterAny = check jsonToAny(tabularDataFilter);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [
                {
                    key: "employee_data",
                    value: {
                        values: [
                            {
                                startTime: "",
                                endTime: "",
                                value: tabularDataFilterAny
                            }
                        ]
                    }
                }
            ],
            relationships: []
        },
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");

    // Verify attributes are present
    test:assertTrue(readEntityResponse.attributes.length() > 0, "Entity should have attributes");

    var attributeValue = readEntityResponse.attributes[0].value.values[0].value;

    JsonObject attributeValueJson = check pbAny:unpack(attributeValue);
    io:println("Attribute value JSON: " + attributeValueJson.toString());

    // Extract the nested data field
    string dataJsonString = <string>attributeValueJson["data"];
    io:println("Data JSON string: " + dataJsonString);

    // Parse the nested JSON string to get the actual tabular data
    json dataJson = check dataJsonString.fromJsonString();
    io:println("Data JSON: " + dataJson.toString());

    // Compare the actual data content instead of exact JSON structure
    verifyTabularData(dataJson, tabularData);

    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with tabular attributes deleted");

    return;
}


@test:Config {}
function testEntityWithTabularAttributesMultiRels() returns error? {
    // Initialize the client
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Test data setup
    string testId = "ID-MIN-A-MULTI-RELS";
    
    // Create tabular data structure
    // TODO: https://github.com/LDFLK/nexoan/issues/284
    json employeeData = {
        "columns": ["id", "name", "age", "department", "salary"],
        "rows": [
            [1, "John Doe", 30, "Engineering", 75000.50],
            [2, "Jane Smith", 25, "Marketing", 65000],
            [3, "Bob Wilson", 35, "Sales", 85000.75],
            [4, "Alice Brown", 28, "Engineering", 70000.25],
            [5, "Charlie Davis", 32, "Finance", 80000]
        ]
    };
    
    json budgetData = {
        "columns": ["id", "category", "amount", "quarter", "status"],
        "rows": [
            [1, "Infrastructure", 150000.50, "Q1", "Approved"],
            [2, "Marketing", 75000, "Q1", "Pending"],
            [3, "Research", 200000.75, "Q2", "Approved"],
            [4, "Operations", 120000.25, "Q1", "Approved"],
            [5, "Training", 45000, "Q2", "Pending"]
        ]
    };

    // Convert JSON to protobuf Any values
    pbAny:Any employeeDataAny = check jsonToAny(employeeData);
    pbAny:Any budgetDataAny = check jsonToAny(budgetData);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "Organization",
            minor: "Minister"
        },
        created: "2025-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2025-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("Minister of Education and Skills")
        },
        metadata: [
            {
                key: "minister_of_education_and_skills_metadata",
                value: check pbAny:pack("tabular_minister_of_education_and_skills_test_value")
            }
        ],
        attributes: [
            {
                key: "employee_data",
                value: {
                    values: [
                        {
                            startTime: "2025-04-01T00:00:00Z",
                            endTime: "",
                            value: employeeDataAny
                        }
                    ]
                }
            },
            {
                key: "budget_data",
                value: {
                    values: [
                        {
                            startTime: "2025-06-01T00:00:00Z",
                            endTime: "",
                            value: budgetDataAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    test:assertTrue(createEntityResponse.id != "", "Entity should be created successfully");
    test:assertEquals(createEntityResponse.id, testId, "Entity ID should match");

    json employeeDataFilter = {
        "columns": ["id", "name", "age", "department", "salary"],
        "rows": [[]]
    };

    json budgetDataFilter = {
        "columns": ["id", "category", "amount", "quarter", "status"],
        "rows": [[]]
    };

    pbAny:Any employeeDataFilterAny = check jsonToAny(employeeDataFilter);
    pbAny:Any budgetDataFilterAny = check jsonToAny(budgetDataFilter);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityWithBudgetDataRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [
                {
                    key: "budget_data",
                    value: {
                        values: [
                            {
                                startTime: "",
                                endTime: "",
                                value: budgetDataFilterAny
                            }
                        ]
                    }
                }
            ],
            relationships: []
        },
        output: ["metadata", "attributes", "relationships"]
    };

    ReadEntityRequest readEntityWithEmployeeDataRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [
                {
                    key: "employee_data",
                    value: {
                        values: [
                            {
                                startTime: "",
                                endTime: "",
                                value: employeeDataFilterAny
                            }
                        ]
                    }
                }
            ],
            relationships: []
        },
        output: ["metadata", "attributes", "relationships"]
    };

    Entity readEntityWithBudgetDataResponse = check ep->ReadEntity(readEntityWithBudgetDataRequest);

    Entity readEntityWithEmployeeDataResponse = check ep->ReadEntity(readEntityWithEmployeeDataRequest);

    // Verify the response
    test:assertTrue(readEntityWithEmployeeDataResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityWithEmployeeDataResponse.id, testId, "Entity ID should match");

    // Verify the response
    test:assertTrue(readEntityWithBudgetDataResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityWithBudgetDataResponse.id, testId, "Entity ID should match");

    // Verify attributes are present
    test:assertTrue(readEntityWithBudgetDataResponse.attributes.length() > 0, "Entity should have attributes");
    test:assertTrue(readEntityWithEmployeeDataResponse.attributes.length() > 0, "Entity should have attributes");

    var budgetAttributeValue = readEntityWithBudgetDataResponse.attributes[0].value.values[0].value;
    var employeeAttributeValue = readEntityWithEmployeeDataResponse.attributes[0].value.values[0].value;

    JsonObject budgetAttributeValueJson = check pbAny:unpack(budgetAttributeValue);
    JsonObject employeeAttributeValueJson = check pbAny:unpack(employeeAttributeValue);
    io:println("Attribute value JSON: " + budgetAttributeValueJson.toString());
    io:println("Attribute value JSON: " + employeeAttributeValueJson.toString());

    // Extract the nested data field
    string dataJsonString = <string>budgetAttributeValueJson["data"];
    string employeeDataJsonString = <string>employeeAttributeValueJson["data"];
    io:println("Budget data JSON string: " + dataJsonString);
    io:println("Employee data JSON string: " + employeeDataJsonString);

    // Parse the nested JSON string to get the actual tabular data
    json dataJson = check dataJsonString.fromJsonString();
    json employeeDataJson = check employeeDataJsonString.fromJsonString();
    io:println("Budget data JSON: " + dataJson.toString());
    io:println("Employee data JSON: " + employeeDataJson.toString());

    // Compare the actual data content instead of exact JSON structure
    verifyTabularData(dataJson, budgetData);
    verifyTabularData(employeeDataJson, employeeData);

    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with tabular attributes deleted");

    return;
}


@test:Config {}
function testEntityWithTabularAttributesUpdate() returns error? {
    // Initialize the client
    COREServiceClient ep = check new (testCoreServiceUrl);
    
    // Test data setup
    string testId = "ID-MIN-A-UPDATE";
    
    // Create tabular data structure
    json tabularData = {
        "columns": ["id", "name", "age", "department", "salary"],
        "rows": [
            [1, "John Doe", 30, "Engineering", 75000.50],
            [2, "Jane Smith", 25, "Marketing", 65000],
            [3, "Bob Wilson", 35, "Sales", 85000.75],
            [4, "Alice Brown", 28, "Engineering", 70000.25],
            [5, "Charlie Davis", 32, "Finance", 80000]
        ]
    };

    // Convert JSON to protobuf Any values
    pbAny:Any tabularDataAny = check jsonToAny(tabularData);

    // First: Create entity with just core values (no attributes, relationships, or metadata)
    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "Organization",
            minor: "Minister"
        },
        created: "2025-11-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2025-11-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("Minister of Finance and Economy")
        },
        metadata: [],
        attributes: [],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    test:assertTrue(createEntityResponse.id != "", "Entity should be created successfully");
    test:assertEquals(createEntityResponse.id, testId, "Entity ID should match");

    // Second: Update entity with attributes
    // Note: Don't include Kind, Created, or Name in update requests as they are immutable
    Entity updateEntityRequest = {
        id: testId,
        kind: {
            major: "",
            minor: ""
        },
        created: "2025-11-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "",
            endTime: "",
            value: check pbAny:pack("")
        },
        metadata: [],
        attributes: [
            {
                key: "employee_data",
                value: {
                    values: [
                        {
                            startTime: "2025-12-01T00:00:00Z",
                            endTime: "",
                            value: tabularDataAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Update entity with attributes
    UpdateEntityRequest updateRequest = {
        id: testId,
        entity: updateEntityRequest
    };
    Entity updateEntityResponse = check ep->UpdateEntity(updateRequest);
    io:println("Entity updated with attributes: " + updateEntityResponse.id);
    test:assertTrue(updateEntityResponse.id != "", "Entity should be updated successfully");

    // put a filter on the attributes
    json tabularDataFilter = {
        "columns": ["id", "name", "age", "department", "salary"],
        "rows": [[]]
    };
    pbAny:Any tabularDataFilterAny = check jsonToAny(tabularDataFilter);

    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: {
                startTime: "",
                endTime: "",
                value: check pbAny:pack("")
            },
            metadata: [],
            attributes: [
                {
                    key: "employee_data",
                    value: {
                        values: [
                            {
                                startTime: "",
                                endTime: "",
                                value: tabularDataFilterAny
                            }
                        ]
                    }
                }
            ],
            relationships: []
        },
        output: ["metadata", "attributes", "relationships"]
    };

    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);

    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");

    // Verify attributes are present
    test:assertTrue(readEntityResponse.attributes.length() > 0, "Entity should have attributes");

    var attributeValue = readEntityResponse.attributes[0].value.values[0].value;

    JsonObject attributeValueJson = check pbAny:unpack(attributeValue);
    io:println("Attribute value JSON: " + attributeValueJson.toString());

    // Extract the nested data field
    string dataJsonString = <string>attributeValueJson["data"];
    io:println("Data JSON string: " + dataJsonString);

    // Parse the nested JSON string to get the actual tabular data
    json dataJson = check dataJsonString.fromJsonString();
    io:println("Data JSON: " + dataJson.toString());

    // Compare the actual data content instead of exact JSON structure
    verifyTabularData(dataJson, tabularData);

    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with tabular attributes deleted");

    return;
}

// Tabular Attribute Integrity Tests

// Test 1 — Idempotency: submitting the same (entityId, attrName) pair a second
// time must append rows to the existing Postgres table, not create a new table.
// No new IS_ATTRIBUTE relationship should appear in the graph either.
@test:Config {
    groups: ["entity", "attributes", "tabular", "integrity"]
}
function testTabularAttributeIdempotency() returns error? {
    COREServiceClient ep = check new (testCoreServiceUrl);
    string testId = "test-tabular-idempotency";
    string attrName = "employee_data";

    json firstBatch = {
        "columns": ["id", "name", "department"],
        "rows": [
            [1, "Alice", "Engineering"],
            [2, "Bob", "Marketing"]
        ]
    };
    json secondBatch = {
        "columns": ["id", "name", "department"],
        "rows": [
            [3, "Charlie", "Finance"],
            [4, "Diana", "Sales"]
        ]
    };

    pbAny:Any firstAny  = check jsonToAny(firstBatch);
    pbAny:Any secondAny = check jsonToAny(secondBatch);

    // Create entity with the first batch of data
    Entity createReq = {
        id: testId,
        kind: { major: "test", minor: "tabular-integrity" },
        created: "2025-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2025-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("Tabular Idempotency Test Entity")
        },
        metadata: [],
        attributes: [
            {
                key: attrName,
                value: {
                    values: [
                        { startTime: "2025-01-01T00:00:00Z", endTime: "", value: firstAny }
                    ]
                }
            }
        ],
        relationships: []
    };
    Entity createResp = check ep->CreateEntity(createReq);
    test:assertEquals(createResp.id, testId, "Entity should be created successfully");
    io:println("[testTabularAttributeIdempotency] First create OK, ID=" + createResp.id);

    // Read back to record the relationship count after first ingest
    ReadEntityRequest readReq1 = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: { startTime: "", endTime: "", value: check pbAny:pack("") },
            metadata: [],
            attributes: [],
            relationships: []
        },
        output: ["relationships"]
    };
    Entity readResp1 = check ep->ReadEntity(readReq1);
    int relCountAfterFirst = readResp1.relationships.length();
    io:println("[testTabularAttributeIdempotency] Relationship count after first ingest: " + relCountAfterFirst.toString());

    // Update entity with a SECOND batch for the SAME attribute
    UpdateEntityRequest updateReq = {
        id: testId,
        entity: {
            id: testId,
            kind: { major: "", minor: "" },
            created: "",
            terminated: "",
            name: { startTime: "", endTime: "", value: check pbAny:pack("") },
            metadata: [],
            attributes: [
                {
                    key: attrName,
                    value: {
                        values: [
                            { startTime: "2025-02-01T00:00:00Z", endTime: "", value: secondAny }
                        ]
                    }
                }
            ],
            relationships: []
        }
    };
    Entity updateResp = check ep->UpdateEntity(updateReq);
    test:assertEquals(updateResp.id, testId, "Entity should be updated successfully");
    io:println("[testTabularAttributeIdempotency] Second update OK");

    // Read back to verify relationship count has NOT grown
    Entity readResp2 = check ep->ReadEntity(readReq1);
    int relCountAfterSecond = readResp2.relationships.length();
    io:println("[testTabularAttributeIdempotency] Relationship count after second ingest: " + relCountAfterSecond.toString());

    test:assertEquals(relCountAfterSecond, relCountAfterFirst,
        "IS_ATTRIBUTE relationship count must NOT grow when appending to an existing attribute");

    // Verify the table now contains ALL rows (first + second batch)
    json allRowsFilter = {
        "columns": ["id", "name", "department"],
        "rows": [[]]
    };
    pbAny:Any allRowsFilterAny = check jsonToAny(allRowsFilter);

    ReadEntityRequest readAttrReq = {
        entity: {
            id: testId,
            kind: {},
            created: "",
            terminated: "",
            name: { startTime: "", endTime: "", value: check pbAny:pack("") },
            metadata: [],
            attributes: [
                {
                    key: attrName,
                    value: {
                        values: [
                            { startTime: "", endTime: "", value: allRowsFilterAny }
                        ]
                    }
                }
            ],
            relationships: []
        },
        output: ["attributes"]
    };
    Entity readAttrResp = check ep->ReadEntity(readAttrReq);
    test:assertTrue(readAttrResp.attributes.length() > 0, "Should return the attribute");

    var attrValue = readAttrResp.attributes[0].value.values[0].value;
    JsonObject attrJson = check pbAny:unpack(attrValue);
    string dataJsonStr = <string>attrJson["data"];
    json dataJson = check dataJsonStr.fromJsonString();

    json rows = check dataJson.rows;
    json[] rowArr = <json[]>rows;
    io:println("[testTabularAttributeIdempotency] Total rows after both batches: " + rowArr.length().toString());

    test:assertEquals(rowArr.length(), 4,
        "After two appends the table must contain exactly 4 rows (2 + 2), not a fresh table");

    // Clean up
    Empty _ = check ep->DeleteEntity({ id: testId });
    io:println("[testTabularAttributeIdempotency] Cleaned up test entity");
    return;
}

// Test 2 — Schema Mismatch: after a table is created with a numeric column,
// pushing a string value into that column must cause an error.
@test:Config {
    groups: ["entity", "attributes", "tabular", "integrity"]
}
function testTabularSchemaMismatch() returns error? {
    COREServiceClient ep = check new (testCoreServiceUrl);
    string testId = "test-tabular-schema-mismatch";
    string attrName = "numeric_data";

    json validBatch = {
        "columns": ["id", "score"],
        "rows": [
            [1, 99],
            [2, 85]
        ]
    };
    // ❌ score column is now a string — incompatible with the integer schema above
    json invalidBatch = {
        "columns": ["id", "score"],
        "rows": [
            [3, "not-a-number"],
            [4, "also-wrong"]
        ]
    };

    pbAny:Any validAny   = check jsonToAny(validBatch);
    pbAny:Any invalidAny = check jsonToAny(invalidBatch);

    // Create entity with a valid numeric schema
    Entity createReq = {
        id: testId,
        kind: { major: "test", minor: "tabular-integrity" },
        created: "2025-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2025-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("Schema Mismatch Test Entity")
        },
        metadata: [],
        attributes: [
            {
                key: attrName,
                value: {
                    values: [
                        { startTime: "2025-01-01T00:00:00Z", endTime: "", value: validAny }
                    ]
                }
            }
        ],
        relationships: []
    };
    Entity createResp = check ep->CreateEntity(createReq);
    test:assertEquals(createResp.id, testId, "Initial entity must be created successfully");
    io:println("[testTabularSchemaMismatch] Initial entity created OK with numeric schema");

    // Attempt to append type-incompatible data — MUST fail
    UpdateEntityRequest updateReq = {
        id: testId,
        entity: {
            id: testId,
            kind: { major: "", minor: "" },
            created: "",
            terminated: "",
            name: { startTime: "", endTime: "", value: check pbAny:pack("") },
            metadata: [],
            attributes: [
                {
                    key: attrName,
                    value: {
                        values: [
                            { startTime: "2025-02-01T00:00:00Z", endTime: "", value: invalidAny }
                        ]
                    }
                }
            ],
            relationships: []
        }
    };
    Entity|error updateResult = ep->UpdateEntity(updateReq);

    if updateResult is error {
        io:println("[testTabularSchemaMismatch] ✅ Schema mismatch correctly rejected: " + updateResult.message());
        // The error message should be informative — not a generic message
        test:assertTrue(updateResult.message().length() > 0, "Error message should not be empty");
    } else {
        test:assertFail("Expected an error when appending schema-incompatible data, but got success");
    }

    // Clean up
    Empty _ = check ep->DeleteEntity({ id: testId });
    io:println("[testTabularSchemaMismatch] Cleaned up test entity");
    return;
}

// Test 3 — Duplicate Primary Key: inserting a row whose primary key already
// exists in the table must be rejected with an error.
@test:Config {
    groups: ["entity", "attributes", "tabular", "integrity"]
}
function testTabularDuplicatePrimaryKey() returns error? {
    COREServiceClient ep = check new (testCoreServiceUrl);
    string testId = "test-tabular-duplicate-pk";
    string attrName = "pk_data";

    json firstBatch = {
        "columns": ["id", "value"],
        "rows": [
            [1, "first"],
            [2, "second"]
        ]
    };
    // ❌ id=1 already exists — this is a duplicate PK
    json duplicateBatch = {
        "columns": ["id", "value"],
        "rows": [
            [1, "duplicate!"]
        ]
    };

    pbAny:Any firstAny     = check jsonToAny(firstBatch);
    pbAny:Any duplicateAny = check jsonToAny(duplicateBatch);

    // Create entity with initial rows
    Entity createReq = {
        id: testId,
        kind: { major: "test", minor: "tabular-integrity" },
        created: "2025-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2025-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("Duplicate PK Test Entity")
        },
        metadata: [],
        attributes: [
            {
                key: attrName,
                value: {
                    values: [
                        { startTime: "2025-01-01T00:00:00Z", endTime: "", value: firstAny }
                    ]
                }
            }
        ],
        relationships: []
    };
    Entity createResp = check ep->CreateEntity(createReq);
    test:assertEquals(createResp.id, testId, "Initial entity must be created successfully");
    io:println("[testTabularDuplicatePrimaryKey] Initial rows created with ids 1 and 2");

    // Attempt to insert a row with a duplicate PK — MUST fail
    UpdateEntityRequest updateReq = {
        id: testId,
        entity: {
            id: testId,
            kind: { major: "", minor: "" },
            created: "",
            terminated: "",
            name: { startTime: "", endTime: "", value: check pbAny:pack("") },
            metadata: [],
            attributes: [
                {
                    key: attrName,
                    value: {
                        values: [
                            { startTime: "2025-02-01T00:00:00Z", endTime: "", value: duplicateAny }
                        ]
                    }
                }
            ],
            relationships: []
        }
    };
    Entity|error updateResult = ep->UpdateEntity(updateReq);

    if updateResult is error {
        io:println("[testTabularDuplicatePrimaryKey] ✅ Duplicate PK correctly rejected: " + updateResult.message());
        test:assertTrue(updateResult.message().length() > 0, "Error message should not be empty");
    } else {
        test:assertFail("Expected an error when inserting a duplicate primary key, but got success");
    }

    // Clean up 
    Empty _ = check ep->DeleteEntity({ id: testId });
    io:println("[testTabularDuplicatePrimaryKey] Cleaned up test entity");
    return;
}