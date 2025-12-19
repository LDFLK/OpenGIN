# Quick Start Guide

This guide will help you interact with OpenGIN using cURL commands. Ensure that the system is running locally before proceeding (see [Installation](./installation.md)).

## APIs

*   **Ingestion API**: `http://localhost:8080` (Write operations)
*   **Read API**: `http://localhost:8081` (Read operations)

## Run a sample query with CURL

### Ingestion API (Write)

**Create**

```bash
curl -X POST http://localhost:8080/entities \
-H "Content-Type: application/json" \
-d '{
  "id": "12345",
  "kind": {
    "major": "example",
    "minor": "test"
  },
  "created": "2024-03-17T10:00:00Z",
  "terminated": "",
  "name": {
    "startTime": "2024-03-17T10:00:00Z",
    "endTime": "",
    "value": {
      "typeUrl": "type.googleapis.com/google.protobuf.StringValue",
      "value": "entity-name"
    }
  },
  "metadata": [
    {"key": "owner", "value": "test-user"},
    {"key": "version", "value": "1.0"},
    {"key": "developer", "value": "V8A"}
  ],
  "attributes": [],
  "relationships": []
}'
```

**Read (via Ingestion API - usually for verification)**

```bash
curl -X GET http://localhost:8080/entities/12345
```

**Update**

> **Note**: The update functionality is currently being refined to ensure it updates existing entities correctly.

```bash
curl -X PUT http://localhost:8080/entities/12345 \
  -H "Content-Type: application/json" \
  -d '{
    "id": "12345",
    "created": "2024-03-18T00:00:00Z",
    "name": {
      "startTime": "2024-03-18T00:00:00Z",
      "value": "entity-name"
    },
    "metadata": [
      {"key": "version", "value": "5.0"}
    ]
  }'
```

**Delete**

```bash
curl -X DELETE http://localhost:8080/entities/12345
```

### Read API (Read-Only)

**Retrieve Metadata**

```bash
curl -X GET "http://localhost:8081/v1/entities/12345/metadata"
```
