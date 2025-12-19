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

### Database Information

The Metadata Store is implemented using a document-oriented database to provide flexibility for storing unstructured or semi-structured data. This allows for dynamic metadata fields without rigid schema constraints.

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



#### 2. metadata_test

**Purpose**: Test collection for metadata (same schema as `metadata`)

Used during testing to isolate test data from production data.

---

## Neo4j Schema

### Database Information

The Graph Store is utilized to manage entities and their relationships. It supports both binary and HTTP protocols for interaction, enabling efficient graph traversals and complex relationship queries.

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

### Relationship Types

**Dynamic Relationship System**: OpenGIN uses a completely generic relationship model where relationship types are not predefined. Users can create any relationship type they need by simply providing a `name` field in the relationship data.

**How it works**:
1. User provides relationship with `name` field (e.g., "reports_to", "depends_on", "manages")
2. System dynamically creates Neo4j relationship with that type
3. Neo4j relationship type becomes the uppercased version or exact value of the `name` field
4. No schema validation or predefined list of relationship types

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

**Note**: The `direction` field is not stored in Neo4j - it's determined by the direction of the arrow in the graph (→ for outgoing, ← for incoming).

**Relationship Types**:
Relationship types are **completely dynamic and user-defined**. The system does not enforce any predefined relationship types. When creating a relationship, the `name` field from the `Relationship` protobuf message becomes the Neo4j relationship type.

Examples from tests and usage:
- `reports_to`: Organizational hierarchy (from E2E tests)
- `depends_on`: Package dependencies (from unit tests)
- Any other name: Users can define any relationship type they need

---

## PostgreSQL Schema

### Database Information

### Database Information

The Attribute Store is built on a relational database system to manage structured, time-series attribute data. It ensures data integrity and supports complex querying capabilities through defined schemas.

### Core Tables

#### 1. Attribute Schema

This table defines the structure of attributes for different entity kinds. It acts as a registry, specifying properties such as data types and storage strategies (e.g., scalar, list, map), ensuring consistent data handling for specific entity classifications.

#### 2. Entity Attributes

This table serves as a mapping layer, linking unique entity identifiers to the specific attribute definitions stored here. It allows the system to associate concrete data values with the defined structure for any given entity.

### Dynamic Attribute Tables

To optimize performance and organize data efficiently, the system automatically creates dedicated tables for attributes based on their entity classification. This approach ensures that attribute data is stored in a structured manner, allowing for faster retrieval and better data management compared to a single monolithic table.


### Data Integrity

**No Distributed Transactions**: Currently, OpenGIN doesn't use distributed transactions. Each database operation is independent.

**Eventual Consistency**: System relies on application-level consistency:
- Entity ID is the common key across all databases
- Core API orchestrates all operations
- Errors are logged but don't rollback previous successful operations

**Future Enhancement**: Implement distributed transactions.

---

## Backup and Restore

### Metadata Store Backup

```bash
mongodump --uri="mongodb://admin:<your_password>@mongodb:27017/opengin?authSource=admin" \
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
