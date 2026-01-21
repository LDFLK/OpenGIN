# Storage Types Documentation

This document describes the different storage types supported by the system and their JSON representations.

## Overview

The system supports five main storage types:
- Tabular Data
- Graph Data
- List Data
- Map Data
- Scalar Data

Each type has a specific JSON structure that helps in identifying and processing the data correctly.

## Storage Types

### 1. Tabular Data

Tabular data represents structured data in a table format with columns and rows.

```json
{
  "attributes": {
    "columns": ["id", "name", "age"],
    "rows": [
      [1, "John", 30],
      [2, "Jane", 25],
      [3, "Bob", 35]
    ]
  }
}
```

#### Features:
- Must have both `columns` and `rows` fields
- `columns` is an array of strings representing column names
- `rows` is an array of arrays, where each inner array represents a row
- Each row must have the same number of elements as there are columns
- Column types should be consistent within each column

### 2. Graph Data

Graph data represents a network of nodes and their relationships.

> **⚠️ Important Note**
> 
> Graph insertion as an attribute is not supported yet. However, the Entity level API can be used to create graphs. 

```json
{
  "attributes": {
    "nodes": [
      {
        "id": "node1",
        "type": "user",
        "properties": {
          "name": "John",
          "age": 30
        }
      }
    ],
    "edges": [
      {
        "source": "node1",
        "target": "node2",
        "type": "follows",
        "properties": {
          "since": "2024-01-01"
        }
      }
    ]
  }
}
```

#### Features:
- Must have both `nodes` and `edges` fields
- `nodes` is an array of objects with:
  - `id`: Unique identifier
  - `type`: Node type
  - `properties`: Additional node attributes
- `edges` is an array of objects with:
  - `source`: Source node ID
  - `target`: Target node ID
  - `type`: Relationship type
  - `properties`: Additional edge attributes

### 3. List Data

List data represents an ordered collection of items.

```json
{
  "attributes": {
    "items": [1, 2, 3, 4, 5]
  }
}
```

> **⚠️ Important Note**
> 
> List insertion as an attribute is not supported yet. However, the record can be saved as a table with one row or one column. 


#### Features:
- Must have an `items` field
- `items` is an array that can contain:
  - Simple values (numbers, strings, booleans)
  - Complex objects
  - Nested arrays
  - Mixed types

### 4. Map Data

Map data represents a collection of key-value pairs.

```json
{
  "attributes": {
    "key1": "value1",
    "key2": 42,
    "key3": {
      "nested": "value"
    }
  }
}
```

> **⚠️ Important Note**
> 
> Document insertion as an attribute is not supported yet. However, the Entity level API can be used to create metadata which is stored as a document. 


#### Features:
- Can have any number of key-value pairs
- Keys must be strings
- Values can be:
  - Simple values (numbers, strings, booleans)
  - Objects
  - Arrays
  - Nested structures

### 5. Scalar Data

Scalar data represents a single value.

```json
{
  "attributes": {
    "value": 42
  }
}
```

> **⚠️ Important Note**
> 
> Scalar insertion as an attribute is not supported yet. However, the Entity level API can be used to save as a one row one column table. 


#### Features:
- Must have exactly one field
- The value can be:
  - Number
  - String
  - Boolean
  - Null

## Type Inference Rules

The system follows these rules to determine the storage type:

1. If the structure has both `columns` and `rows` fields, it's classified as Tabular Data
2. If the structure has both `nodes` and `edges` fields, it's classified as Graph Data
3. If the structure has an `items` field containing an array, it's classified as List Data
4. If the structure has a single field with a scalar value, it's classified as Scalar Data
5. If none of the above conditions are met, it's classified as Map Data

## Best Practices

1. **Consistency**: Maintain consistent data types within columns for tabular data
2. **Unique Identifiers**: Use unique IDs for nodes in graph data
3. **Type Safety**: Ensure proper type checking when working with mixed-type lists
4. **Nesting**: Be mindful of nesting depth in map and graph structures
5. **Validation**: Validate data structures before processing

## Examples

### Complex Tabular Data
```json
{
  "attributes": {
    "columns": ["id", "name", "age", "address"],
    "rows": [
      [1, "John", 30, {"city": "NY", "zip": "10001"}],
      [2, "Jane", 25, {"city": "SF", "zip": "94105"}]
    ],
    "metadata": {
      "total_rows": 2,
      "last_updated": "2024-03-20"
    }
  }
}
```

### Complex Graph Data
```json
{
  "attributes": {
    "nodes": [
      {
        "id": "user1",
        "type": "user",
        "properties": {
          "name": "Alice",
          "age": 30,
          "location": "NY"
        }
      }
    ],
    "edges": [
      {
        "source": "user1",
        "target": "post1",
        "type": "created",
        "properties": {
          "timestamp": "2024-03-20T10:00:00Z"
        }
      }
    ]
  }
}
```

### Complex List Data
```json
{
  "attributes": {
    "items": [
      [1, 2, 3],
      {"id": 1, "name": "item1"},
      "string item",
      true
    ]
  }
}
```

### Complex Map Data
```json
{
  "attributes": {
    "user": {
      "profile": {
        "name": "John",
        "age": 30,
        "address": {
          "city": "New York",
          "zip": "10001"
        }
      },
      "settings": {
        "theme": "dark",
        "notifications": true
      }
    }
  }
}
```
