# Database Schemas - Detailed Documentation

This document provides comprehensive details about the database schemas used in OpenGIN across MongoDB, Neo4j, and PostgreSQL.

---

## Overview

OpenGIN uses a multi-database architecture where each database is optimized for specific data types:

| Database | Purpose | Data Stored |
|----------|---------|-------------|
| MongoDB | Flexible metadata | Key-value metadata pairs |
| Neo4j | Graph relationships | Entity nodes and relationship edges |
| PostgreSQL | Structured attributes | Time-series attribute data with schemas |

---

## MongoDB Schema

### Database Information

**Database Name**: `opengin`  
**Connection**: `mongodb://admin:admin123@mongodb:27017/opengin?authSource=admin`

### Collections

#### 1. metadata

**Purpose**: Store entity metadata as flexible key-value pairs

**Schema** (document structure):
```javascript
{
    "_id": "entity123",                    // Entity ID (Primary Key)
    "metadata": {                          // Metadata object
        "key1": "value1",
        "key2": "value2",
        "key3": 123,
        "key4": true,
        "nested": {
            "subkey": "subvalue"
        }
    },
    "created_at": ISODate("2024-01-01T00:00:00Z"),  // Optional timestamp
    "updated_at": ISODate("2024-01-01T00:00:00Z")   // Optional timestamp
}
```

**Example Document**:
```javascript
{
    "_id": "employee_001",
    "metadata": {
        "department": "Engineering",
        "role": "Software Engineer",
        "manager": "manager_123",
        "employeeId": "EMP-001",
        "hireDate": "2024-01-01",
        "location": "New York",
        "active": true,
        "skills": ["Go", "Ballerina", "Neo4j"],
        "performance": {
            "rating": 4.5,
            "lastReview": "2024-06-01"
        }
    }
}
```

**Indexes**:
```javascript
// Primary index (automatic)
{ "_id": 1 }

// Optional: Text search on metadata values
{ "metadata.$**": "text" }

// Optional: Query specific metadata fields
{ "metadata.department": 1 }
{ "metadata.active": 1 }
```

**Operations**:

**Insert/Update**:
```javascript
db.metadata.updateOne(
    { "_id": "entity123" },
    { 
        "$set": { 
            "metadata": {
                "department": "Engineering",
                "role": "Engineer"
            }
        }
    },
    { "upsert": true }
)
```

**Query**:
```javascript
// Get metadata by entity ID
db.metadata.findOne({ "_id": "entity123" })

// Find entities by metadata value
db.metadata.find({ "metadata.department": "Engineering" })

// Search metadata (if text index exists)
db.metadata.find({ "$text": { "$search": "engineer" } })
```

**Delete**:
```javascript
db.metadata.deleteOne({ "_id": "entity123" })
```

#### 2. metadata_test

**Purpose**: Test collection for metadata (same schema as `metadata`)

Used during testing to isolate test data from production data.

---

## Neo4j Schema

### Database Information

**Connection**: 
- Bolt: `bolt://neo4j:7687`
- HTTP: `http://neo4j:7474`
- Credentials: `neo4j/neo4j123`

### Node Types

#### Entity Node

**Label**: `:Entity`

**Properties**:
```cypher
{
    id: String,              // Unique entity identifier (REQUIRED)
    kind_major: String,      // Major entity classification (REQUIRED)
    kind_minor: String,      // Minor entity classification (optional)
    name: String,            // Entity name (REQUIRED)
    created: String,         // ISO 8601 timestamp (REQUIRED)
    terminated: String       // ISO 8601 timestamp (optional, null = active)
}
```

**Example**:
```cypher
(:Entity {
    id: "employee_001",
    kind_major: "Person",
    kind_minor: "Employee",
    name: "John Doe",
    created: "2024-01-01T00:00:00Z",
    terminated: null
})
```

**Indexes and Constraints**:
```cypher
// Unique constraint on id
CREATE CONSTRAINT entity_id_unique IF NOT EXISTS
FOR (e:Entity) REQUIRE e.id IS UNIQUE;

// Index on id for fast lookups
CREATE INDEX entity_id_index IF NOT EXISTS
FOR (e:Entity) ON (e.id);

// Index on kind for filtering
CREATE INDEX entity_kind_major_index IF NOT EXISTS
FOR (e:Entity) ON (e.kind_major);

// Composite index
CREATE INDEX entity_kind_composite_index IF NOT EXISTS
FOR (e:Entity) ON (e.kind_major, e.kind_minor);
```

### Relationship Types

**Dynamic Relationship System**: OpenGIN uses a completely generic relationship model where relationship types are not predefined. Users can create any relationship type they need by simply providing a `name` field in the relationship data.

**How it works**:
1. User provides relationship with `name` field (e.g., "reports_to", "depends_on", "manages")
2. System dynamically creates Neo4j relationship with that type
3. Neo4j relationship type becomes the uppercased version or exact value of the `name` field
4. No schema validation or predefined list of relationship types

**Implementation** (from `neo4j_client.go`):
```go
// Relationship type is dynamically injected into Cypher query
createQuery := `MATCH (p {Id: $parentID}), (c {Id: $childID})
                MERGE (p)-[r:` + rel.Name + ` {Id: $relationshipID}]->(c)
                SET r.Created = datetime($startDate)`
```

#### Relationship Structure

All relationships in Neo4j store the following properties:

**Neo4j Properties** (what's actually stored in the graph):
```cypher
{
    Id: String,              // Relationship identifier (uppercase I)
    Created: DateTime,       // When relationship started (Neo4j datetime type)
    Terminated: DateTime     // When relationship ended (Neo4j datetime type, null = active)
}
```

**Important**: The `name` field from the API/Protobuf becomes the **relationship TYPE** in Neo4j, not a property. It appears in the Cypher syntax as `[:relationshipType]`.

**Example in Neo4j**:
```cypher
(:Entity {Id: "employee_001"})-[
    :reports_to {
        Id: "rel_001",
        Created: datetime("2024-01-01T00:00:00Z"),
        Terminated: null
    }
]->(:Entity {Id: "manager_123"})
```

**Note**: The `direction` field is not stored in Neo4j - it's determined by the direction of the arrow in the graph (→ for outgoing, ← for incoming).

**Relationship Types**:
Relationship types are **completely dynamic and user-defined**. The system does not enforce any predefined relationship types. When creating a relationship, the `name` field from the `Relationship` protobuf message becomes the Neo4j relationship type.

Examples from tests and usage:
- `reports_to`: Organizational hierarchy (from E2E tests)
- `depends_on`: Package dependencies (from unit tests)
- Any other name: Users can define any relationship type they need

### Cypher Queries

#### Create Entity Node

```cypher
CREATE (e:Entity {
    id: $id,
    kind_major: $kind_major,
    kind_minor: $kind_minor,
    name: $name,
    created: $created,
    terminated: $terminated
})
RETURN e
```

#### Create Relationship

**Note**: The relationship type in the Cypher query is dynamically generated from the `name` field. The example below shows the pattern, but `[relationshipType]` is replaced with the actual value from `relationship.Name` at runtime.

```cypher
MATCH (source:Entity {id: $sourceId})
MATCH (target:Entity {id: $targetId})
MERGE (source)-[r:[relationshipType] {
    Id: $relationshipId
}]->(target)
SET r.Created = datetime($startTime)
SET r.Terminated = datetime($endTime)  -- if endTime is provided
RETURN r
```

**Actual Go code** (from `neo4j_client.go`):
```go
createQuery := `MATCH (p {Id: $parentID}), (c {Id: $childID})
                MERGE (p)-[r:` + rel.Name + ` {Id: $relationshipID}]->(c)
                SET r.Created = datetime($startDate)`
```

#### Query Entity with Relationships

```cypher
MATCH (e:Entity {id: $entityId})
OPTIONAL MATCH (e)-[r]->(related:Entity)
RETURN e, collect(r) as relationships, collect(related) as relatedEntities
```

#### Temporal Relationship Query

**Note**: Replace `[relationshipType]` with the actual relationship type (e.g., `reports_to`, `depends_on`, etc.) or use a variable relationship pattern `[r]` to match any type.

```cypher
-- Query specific relationship type at a point in time
MATCH (e:Entity {id: $entityId})-[r:reports_to]->(m:Entity)
WHERE r.Created <= $activeAt
  AND (r.Terminated IS NULL OR r.Terminated >= $activeAt)
RETURN e, r, m

-- Query ANY relationship type at a point in time
MATCH (e:Entity {id: $entityId})-[r]->(m:Entity)
WHERE r.Created <= $activeAt
  AND (r.Terminated IS NULL OR r.Terminated >= $activeAt)
RETURN e, r, m, type(r) as relationshipType
```

#### Find Entities by Kind

```cypher
MATCH (e:Entity)
WHERE e.kind_major = $kindMajor 
  AND e.kind_minor = $kindMinor
  AND e.terminated IS NULL
RETURN e
ORDER BY e.created DESC
LIMIT 100
```

#### Graph Traversal Examples

**Example 1: Find All Entities with Specific Relationship (e.g., organizational hierarchy)**

```cypher
-- Find all entities that have a specific relationship TO this entity
MATCH (entity:Entity)-[r:reports_to]->(central:Entity {Id: $entityId})
WHERE entity.Terminated IS NULL
RETURN entity
```

**Example 2: Multi-level Traversal (e.g., organizational tree)**

```cypher
-- Find all entities connected via a relationship path (1 to n levels deep)
MATCH path = (entity:Entity)-[:reports_to*1..5]->(central:Entity {Id: $entityId})
WHERE entity.Terminated IS NULL
RETURN entity, length(path) as depth
ORDER BY depth
```

**Example 3: Find All Related Entities (any relationship type)**

```cypher
-- Find all entities related to this entity (outgoing relationships)
MATCH (source:Entity {Id: $entityId})-[r]->(target:Entity)
RETURN source, type(r) as relationshipType, target
```

#### Delete Entity and Relationships

```cypher
MATCH (e:Entity {id: $entityId})
DETACH DELETE e
```
---

## PostgreSQL Schema

### Database Information

**Database Name**: `opengin`  
**Connection**: `postgresql://postgres:postgres@postgres:5432/opengin`

### Core Tables

#### 1. attribute_schemas

**Purpose**: Define attribute schemas for different entity kinds

**Schema**:
```sql
CREATE TABLE attribute_schemas (
    id SERIAL PRIMARY KEY,
    kind_major VARCHAR(255) NOT NULL,
    kind_minor VARCHAR(255),
    attr_name VARCHAR(255) NOT NULL,
    data_type VARCHAR(50) NOT NULL,
    storage_type VARCHAR(50) NOT NULL,
    is_nullable BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(kind_major, kind_minor, attr_name)
);
```

**Columns**:
- `id`: Auto-incrementing primary key
- `kind_major`: Entity major classification (e.g., "Person", "Organization")
- `kind_minor`: Entity minor classification (e.g., "Employee", "Contractor")
- `attr_name`: Attribute name (e.g., "salary", "address")
- `data_type`: Inferred data type (`int`, `float`, `string`, `bool`, `date`, `time`, `datetime`)
- `storage_type`: Storage strategy (`SCALAR`, `LIST`, `MAP`, `TABULAR`, `GRAPH`)
- `is_nullable`: Whether null values are allowed
- `created_at`: Schema creation timestamp
- `updated_at`: Schema last update timestamp

**Example Data**:
```sql
INSERT INTO attribute_schemas 
    (kind_major, kind_minor, attr_name, data_type, storage_type)
VALUES
    ('Person', 'Employee', 'salary', 'int', 'SCALAR'),
    ('Person', 'Employee', 'skills', 'string', 'LIST'),
    ('Person', 'Employee', 'address', 'string', 'MAP'),
    ('Project', NULL, 'budget', 'float', 'SCALAR'),
    ('Project', NULL, 'timeline', 'datetime', 'TABULAR');
```

**Indexes**:
```sql
CREATE INDEX idx_attr_schemas_kind ON attribute_schemas(kind_major, kind_minor);
CREATE INDEX idx_attr_schemas_name ON attribute_schemas(attr_name);
```

#### 2. entity_attributes

**Purpose**: Link entities to their attributes

**Schema**:
```sql
CREATE TABLE entity_attributes (
    id SERIAL PRIMARY KEY,
    entity_id VARCHAR(255) NOT NULL,
    attr_name VARCHAR(255) NOT NULL,
    schema_id INTEGER REFERENCES attribute_schemas(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(entity_id, attr_name)
);
```

**Columns**:
- `id`: Auto-incrementing primary key
- `entity_id`: Reference to entity (matches Entity.id from Neo4j)
- `attr_name`: Attribute name
- `schema_id`: Foreign key to `attribute_schemas`
- `created_at`: Link creation timestamp

**Example Data**:
```sql
INSERT INTO entity_attributes (entity_id, attr_name, schema_id)
VALUES
    ('employee_001', 'salary', 1),
    ('employee_001', 'skills', 2),
    ('employee_001', 'address', 3);
```

**Indexes**:
```sql
CREATE INDEX idx_entity_attrs_entity_id ON entity_attributes(entity_id);
CREATE INDEX idx_entity_attrs_attr_name ON entity_attributes(attr_name);
```

### Dynamic Attribute Tables

#### Naming Convention

Dynamic tables follow the pattern: `attr_{kind_major}_{attr_name}`

Examples:
- `attr_Person_salary`
- `attr_Person_skills`
- `attr_Project_budget`
- `attr_Organization_revenue`

### Type Mapping

| Inferred Type | PostgreSQL Type | Example |
|---------------|----------------|---------|
| int | INTEGER | 42, -100, 0 |
| float | DOUBLE PRECISION | 3.14, -0.001, 1.5e10 |
| string | TEXT | "Hello", "12345" |
| bool | BOOLEAN | true, false |
| date | DATE | 2024-01-01 |
| time | TIME | 14:30:00 |
| datetime | TIMESTAMP | 2024-01-01 14:30:00Z |
| array | TEXT[] or INTEGER[] | ["a", "b"] or [1, 2, 3] |
| object/map | JSONB | {"key": "value"} |


### Data Integrity

**No Distributed Transactions**: Currently, OpenGIN doesn't use distributed transactions. Each database operation is independent.

**Eventual Consistency**: System relies on application-level consistency:
- Entity ID is the common key across all databases
- Core API orchestrates all operations
- Errors are logged but don't rollback previous successful operations

**Future Enhancement**: Implement two-phase commit or saga pattern for true distributed transactions.

---

## Backup and Restore

### MongoDB Backup

```bash
mongodump --uri="mongodb://admin:admin123@mongodb:27017/opengin?authSource=admin" \
    --out=/backup/mongodb/
```

### Neo4j Backup

```bash
neo4j-admin dump --database=neo4j --to=/backup/neo4j/neo4j.dump
```

### PostgreSQL Backup

```bash
pg_dump -h postgres -U postgres -d opengin -F tar -f /backup/postgres/opengin.tar
```

See [Backup Integration Guide](../../reference/operations/backup_integration.md) for complete backup/restore workflow.

---

## Related Documentation

- [Main Architecture Overview](./index.md)
- [How It Works](data_flow.md)
- [Data Types](../../reference/datatype.md)
- [Storage Types](../../reference/storage.md)

---
