# Core API Documentation

The Core API is the central orchestration service in the OpenGIN platform, responsible for coordinating data operations across multiple databases and providing a unified interface for entity management.

## Overview

The Core API serves as the business logic layer that:
- Receives protobuf messages from the Ingestion and Read APIs
- Orchestrates data operations across MongoDB, Neo4j, and PostgreSQL
- Implements type and storage inference algorithms
- Manages temporal data and relationships
- Returns processed data back to the API layer

---

## Architecture

### Service Components

```
┌─────────────────────────────────────────────────────────────────┐
│                        Core API (Go)                            │
│                         Port: 50051                             │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                gRPC Server                              │    │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐        │    │
│  │  │CreateEntity │ │ ReadEntity  │ │UpdateEntity │        │    │
│  │  │             │ │             │ │             │        │    │
│  │  │DeleteEntity │ │ QueryEntity │ │             │        │    │
│  │  └─────────────┘ └─────────────┘ └─────────────┘        │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                  Engine Layer                           │    │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐        │    │
│  │  │AttributeProc│ │TypeInference│ │StorageInfer │        │    │
│  │  │essor        │ │             │ │ence         │        │    │
│  │  └─────────────┘ └─────────────┘ └─────────────┘        │    │
│  │  ┌─────────────────────────────────────────────────┐    │    │
│  │  │         GraphMetadataManager                    │    │    │
│  │  └─────────────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                Repository Layer                         │    │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐        │    │
│  │  │   MongoDB   │ │    Neo4j    │ │ PostgreSQL  │        │    │
│  │  │ Repository  │ │ Repository  │ │ Repository  │        │    │
│  │  └─────────────┘ └─────────────┘ └─────────────┘        │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

---

## gRPC Service Methods

### 1. CreateEntity

Creates a new entity with metadata, attributes, and relationships.

**Request Flow:**
1. Validate entity data and structure
2. Process metadata → MongoDB
3. Create entity node → Neo4j
4. Process attributes → PostgreSQL (with type/storage inference)
5. Create relationships → Neo4j
6. Return created entity

**Key Features:**
- Automatic type inference for attributes
- Dynamic storage strategy determination
- Temporal relationship support
- Atomic operations across databases

### 2. ReadEntity

Retrieves entity data from multiple databases based on requested output parameters.

**Request Flow:**
1. Always fetch core entity info from Neo4j
2. Conditionally fetch metadata from MongoDB
3. Conditionally fetch attributes from PostgreSQL
4. Conditionally fetch relationships from Neo4j
5. Assemble complete entity response

**Output Parameters:**
- `metadata` - Include entity metadata
- `attributes` - Include entity attributes
- `relationships` - Include entity relationships
- `all` - Include everything

### 3. UpdateEntity

Updates existing entity data while maintaining temporal consistency.

**Request Flow:**
1. Validate entity exists
2. Update metadata in MongoDB (if provided)
3. Update entity properties in Neo4j (if provided)
4. Update attributes in PostgreSQL (if provided)
5. Update relationships in Neo4j (if provided)
6. Return updated entity

### 4. DeleteEntity

Removes entity and all associated data from all databases.

**Request Flow:**
1. Delete metadata from MongoDB
2. Delete entity node and relationships from Neo4j
3. Delete attributes from PostgreSQL
4. Return deletion confirmation

### 5. QueryEntity

Performs complex queries across multiple databases.

**Capabilities:**
- Cross-database joins
- Temporal queries
- Relationship traversal
- Attribute filtering
- Metadata search

---

## Engine Layer Components

### AttributeProcessor

**Purpose:** Processes and validates entity attributes before storage.

**Key Functions:**
- Validates attribute data types
- Handles time-based attribute values
- Manages attribute schema evolution
- Coordinates with type and storage inference

**Process Flow:**
```
Input Attribute → Type Validation → Storage Determination → Database Storage
```

### TypeInference

**Purpose:** Automatically determines data types for attribute values.

**Supported Types:**
- `int` - Integer numbers
- `float` - Decimal numbers
- `string` - Text data
- `bool` - Boolean values
- `date` - Date values
- `time` - Time values
- `datetime` - Date and time values

**Inference Rules:**
1. **Integer Detection:** Whole numbers without decimals
2. **Float Detection:** Numbers with decimal points or scientific notation
3. **Boolean Detection:** `true`, `false`, `1`, `0`
4. **Date Detection:** ISO date format patterns
5. **Time Detection:** Time format patterns
6. **String Fallback:** Any non-matching data

### StorageInference

**Purpose:** Determines optimal storage strategy for attributes.

**Storage Types:**
- `SCALAR` - Single values (numbers, strings, booleans)
- `LIST` - Arrays of values
- `MAP` - Key-value pairs
- `TABULAR` - Table-like data with rows and columns
- `GRAPH` - Network data (not fully supported)

**Inference Logic:**
```
Data Structure Analysis → Storage Type Determination → Database Assignment
```

### GraphMetadataManager

**Purpose:** Manages graph-specific metadata and relationships. Acts as a lookup
graph for entity and its data. 

**Key Functions:**
- Entity node creation and updates
- Relationship management
- Graph traversal optimization
- Temporal relationship handling

## Related Documentation

- [Architecture Overview](./index.md) - System architecture
- [Service APIs](./api-layer-details.md) - API documentation
- [Database Schemas](./database-schemas.md) - Database structures
- [How It Works](data_flow.md) - End-to-end data flow
- [Storage Types](../../reference/storage.md) - Storage inference details
