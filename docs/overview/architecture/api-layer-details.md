# API Layer - Detailed Architecture

This document provides comprehensive details about the Ingestion API and Read API layers of the OpenGIN system.

---

## Overview

The API Layer consists of two Ballerina-based REST services that provide external access to the OpenGIN system:
- **Ingestion API**: Handles entity mutations (CREATE, UPDATE, DELETE)
- **Read API**: Handles entity queries and retrieval

Both APIs act as translation layers between external HTTP/JSON clients and the internal gRPC/Protobuf CORE service.

---

## Ingestion API

### Overview

**Location**: `opengin/ingestion-api/`  
**Language**: Ballerina  
**Protocol**: HTTP/REST + JSON  
**Port**: 8080  
**Contract**: `opengin/contracts/rest/ingestion_api.yaml`

### Request/Response Flow

#### CREATE Entity

**Request**:
```bash
POST /entities
Content-Type: application/json

{
  "id": "entity123",
  "kind": {
    "major": "Person",
    "minor": "Employee"
  },
  "created": "2024-01-01T00:00:00Z",
  "name": {
    "startTime": "2024-01-01T00:00:00Z",
    "endTime": "",
    "value": "John Doe"
  },
  "metadata": [
    {"key": "department", "value": "Engineering"},
    {"key": "role", "value": "Engineer"}
  ],
  "attributes": [
    {
      "key": "salary",
      "value": {
        "values": [
          {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": 100000
          }
        ]
      }
    }
  ],
  "relationships": [
    {
      "key": "reports_to",
      "value": {
        "id": "rel123",
        "relatedEntityId": "manager456",
        "name": "reports_to",
        "startTime": "2024-01-01T00:00:00Z",
        "endTime": "",
        "direction": "outgoing"
      }
    }
  ]
}
```

**Response**:
```json
{
  "id": "entity123",
  "kind": {
    "major": "Person",
    "minor": "Employee"
  },
  "created": "2024-01-01T00:00:00Z",
  "name": {
    "startTime": "2024-01-01T00:00:00Z",
    "value": "John Doe"
  }
}
```

**Status Codes**:
- `201 Created`: Entity successfully created
- `400 Bad Request`: Invalid JSON or missing required fields
- `409 Conflict`: Entity with ID already exists
- `500 Internal Server Error`: CORE service error

#### READ Entity

**Request**:
```bash
GET /entities/entity123
```

**Response**:
```json
{
  "id": "entity123",
  "kind": {
    "major": "Person",
    "minor": "Employee"
  },
  "created": "2024-01-01T00:00:00Z",
  "name": {
    "startTime": "2024-01-01T00:00:00Z",
    "value": "John Doe"
  },
  "metadata": [
    {"key": "department", "value": "Engineering"}
  ],
  "attributes": [],
  "relationships": []
}
```

**Status Codes**:
- `200 OK`: Entity found and returned
- `404 Not Found`: Entity doesn't exist
- `500 Internal Server Error`: CORE service error

#### UPDATE Entity

**Request**:
```bash
PUT /entities/entity123
Content-Type: application/json

{
  "id": "entity123",
  "kind": {
    "major": "Person",
    "minor": "Employee"
  },
  "metadata": [
    {"key": "department", "value": "Sales"}
  ]
}
```

**Response**:
```json
{
  "id": "entity123",
  "kind": {
    "major": "Person",
    "minor": "Employee"
  },
  "metadata": [
    {"key": "department", "value": "Sales"}
  ]
}
```

**Status Codes**:
- `200 OK`: Entity successfully updated
- `404 Not Found`: Entity doesn't exist
- `400 Bad Request`: Invalid update data
- `500 Internal Server Error`: Core API error

#### DELETE Entity

**Request**:
```bash
DELETE /entities/entity123
```

**Response**:
```
204 No Content
```

**Status Codes**:
- `204 No Content`: Entity successfully deleted
- `404 Not Found`: Entity doesn't exist
- `500 Internal Server Error`: CORE service error

---

## Read API

### Overview

**Location**: `opengin/read-api/`  
**Language**: Ballerina  
**Protocol**: HTTP/REST + JSON  
**Port**: 8081  
**Contract**: `opengin/contracts/rest/read_api.yaml`

### Read Operations

#### Get Metadata

**Request**:
```bash
GET /v1/entities/entity123/metadata
```

**Response**:
```json
{
  "metadata": {
    "department": "Engineering",
    "role": "Engineer",
    "employeeId": "EMP-123"
  }
}
```

**Use Case**: Retrieve flexible key-value metadata for an entity

#### Get Relationships

**Request**:
```bash
GET /v1/entities/entity123/relationships?name=reports_to&direction=outgoing
```

**Query Parameters**:
- `name`: Filter by relationship type
- `direction`: `outgoing` or `incoming`
- `relatedEntityId`: Filter by related entity
- `activeAt`: Temporal query (ISO 8601 timestamp)

**Response**:
```json
{
  "relationships": [
    {
      "id": "rel123",
      "name": "reports_to",
      "relatedEntityId": "manager456",
      "startTime": "2024-01-01T00:00:00Z",
      "endTime": null,
      "direction": "outgoing"
    }
  ]
}
```

**Use Case**: Graph traversal, finding related entities

#### Get Attributes

**Request**:
```bash
GET /v1/entities/entity123/attributes?name=salary&activeAt=2024-06-01T00:00:00Z
```

**Query Parameters**:
- `name`: Filter by attribute name
- `activeAt`: Get attribute value at specific time

**Response**:
```json
{
  "attributes": {
    "salary": {
      "values": [
        {
          "startTime": "2024-01-01T00:00:00Z",
          "endTime": "2024-06-30T23:59:59Z",
          "value": 100000
        },
        {
          "startTime": "2024-07-01T00:00:00Z",
          "endTime": null,
          "value": 110000
        }
      ]
    }
  }
}
```

**Use Case**: Time-series data, historical attribute values

#### Get Entity with Selective Fields

**Request**:
```bash
GET /v1/entities/entity123?output=metadata,relationships
```

**Query Parameters**:
- `output`: Comma-separated list of fields to retrieve
  - Options: `metadata`, `relationships`, `attributes`
  - If omitted: returns only basic entity info

**Response**:
```json
{
  "id": "entity123",
  "kind": {
    "major": "Person",
    "minor": "Employee"
  },
  "name": {
    "value": "John Doe",
    "startTime": "2024-01-01T00:00:00Z"
  },
  "created": "2024-01-01T00:00:00Z",
  "metadata": {
    "department": "Engineering"
  },
  "relationships": [
    {
      "name": "reports_to",
      "relatedEntityId": "manager456"
    }
  ]
}
```

**Use Case**: Optimized queries, reduce payload size

### Temporal Queries

The Read API supports temporal queries using the `activeAt` parameter:

**Example**: Get employee's salary on specific date
```bash
GET /v1/entities/entity123/attributes?name=salary&activeAt=2024-03-15T00:00:00Z
```

**Backend Filter**:
```sql
WHERE start_time <= '2024-03-15T00:00:00Z'
  AND (end_time IS NULL OR end_time >= '2024-03-15T00:00:00Z')
```

This returns only the attribute value that was active on March 15, 2024.

### Performance Optimization

**Selective Field Retrieval**:
- Only fetch requested fields from CORE service
- Reduces database load
- Reduces network bandwidth
- Faster response times

**Example**:
```
Query: GET /v1/entities/entity123?output=metadata

Instead of:
  ├─ MongoDB (metadata)         ✓ Retrieved
  ├─ Neo4j (entity info)        ✓ Retrieved
  ├─ Neo4j (relationships)      ✗ Skipped
  └─ PostgreSQL (attributes)    ✗ Skipped
```

---

## OpenAPI Contracts

### Ingestion API Contract

**File**: `opengin/contracts/rest/update_api.yaml`

Defines:
- Entity schema
- Request/response structures
- HTTP methods and paths
- Status codes
- Error responses

**Code Generation**:
```bash
bal openapi -i ../contracts/rest/ingestion_api.yaml --mode service
```

### Read API Contract

**File**: `opengin/contracts/rest/read_api.yaml`

Defines:
- Query parameters
- Filter options
- Response structures
- Pagination (if implemented)

**Code Generation**:
```bash
bal openapi -i ../contracts/rest/read_api.yaml --mode service
```

---



## Best Practices

### 1. Use Selective Retrieval

❌ **Don't**:
```bash
GET /entities/entity123
# Returns everything (slow, large payload)
```

✅ **Do**:
```bash
GET /v1/entities/entity123?output=metadata
# Returns only what you need (fast, small payload)
```

### 2. Use Specific Read Endpoints

❌ **Don't**:
```bash
GET /entities/entity123  # Get full entity, then filter relationships in client
```

✅ **Do**:
```bash
GET /v1/entities/entity123/relationships?name=reports_to
# Server-side filtering, better performance
```

### 3. Use Temporal Queries

❌ **Don't**:
```bash
GET /v1/entities/entity123/attributes?name=salary
# Returns all values, filter in client
```

✅ **Do**:
```bash
GET /v1/entities/entity123/attributes?name=salary&activeAt=2024-03-15T00:00:00Z
# Returns only relevant value
```

---

## Related Documentation

- [Main Architecture Overview](./index.md)
- [Ingestion API](../../opengin/ingestion-api/README.md)
- [Read API](../../opengin/read-api/README.md)

---
