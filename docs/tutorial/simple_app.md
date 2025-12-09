# Simple Application using OpenGIN

In this tutorial, we will walk through how to model and create a simple "Employee" entity using OpenGIN. We will define the entity, add some attributes, and link it to an organization.

## Scenario

We want to store information about an employee named **"Alice Smith"**.
- **Role**: Software Engineer
- **Department**: Engineering
- **Joined**: Jan 15, 2024
- **Salary**: $120,000 (starting value)

## Step 1: Define the Entity JSON

We need to construct a JSON payload that matches the OpenGIN Entity specification.

### 1.1 Core Identity
We'll assign an ID `emp-001` and define the Kind.
```json
"id": "emp-001",
"kind": {
    "major": "Person",
    "minor": "Employee"
}
```

### 1.2 Name
The name is a `TimeBasedValue`.
```json
"name": {
    "startTime": "2024-01-15T00:00:00Z",
    "value": "Alice Smith"
}
```

### 1.3 Metadata
Static info like department goes here.
```json
"metadata": [
    {"key": "department", "value": "Engineering"},
    {"key": "role", "value": "Software Engineer"}
]
```

### 1.4 Attributes
Time-sensitive data like Salary.
```json
"attributes": {
    "salary": {
        "values": [
            {
                "startTime": "2024-01-15T00:00:00Z",
                "value": "120000"
            }
        ]
    }
}
```

## Step 2: Create the Entity

Send a POST request to the **Ingestion API** (`http://localhost:8080/entities`).

```bash
curl -X POST http://localhost:8080/entities \
-H "Content-Type: application/json" \
-d '{
    "id": "emp-001",
    "kind": {
        "major": "Person",
        "minor": "Employee"
    },
    "created": "2024-01-15T00:00:00Z",
    "name": {
        "startTime": "2024-01-15T00:00:00Z",
        "value": "Alice Smith"
    },
    "metadata": [
        {"key": "department", "value": "Engineering"},
        {"key": "role", "value": "Software Engineer"}
    ],
    "attributes": {
        "salary": {
            "values": [
                {
                    "startTime": "2024-01-15T00:00:00Z",
                    "value": "120000"
                }
            ]
        }
    },
    "relationships": {}
}'
```

## Step 3: Verify Creation

You can retrieve the entity using the Read API or Ingestion API (for simple lookups).

```bash
curl -X GET http://localhost:8080/entities/emp-001
```

You should see the JSON response mimicking what you sent, confirming it's stored in the underlying databases (MongoDB, Neo4j, PostgreSQL).
