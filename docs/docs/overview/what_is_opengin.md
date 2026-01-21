# What is OpenGIN?

Open General Information Network, hereafter referred to as **OpenGIN**, is an open-source platform designed to build a time-aware digital twin of an ecosystem by defining its entities, relationships, and data according to a specification. OpenGIN core supports a wide variety of data formats to provide efficient querying to simulate the digital twin. Underneath, OpenGIN uses a polyglot database definition which supports representing the changes of an ecosystem through time-traveling.

## Data Model

* **Temporal Data Model**: The use of `TimeBasedValue` enables temporal data management, allowing the system to track both when data is valid (business time) and when it was recorded (system time).
* **Flexible Schema**: The `metadata` field uses `google.protobuf.Any` to support arbitrary key-value pairs without schema constraints.
* **Immutable Core Fields**: Fields like `id`, `kind`, and `created` are read-only, ensuring data integrity.
* **Graph-Ready Structure**: The `relationships` field enables graph-based data modeling and traversal.
* **Storage Awareness**: The `attributes` of each entity can be stored in the most suitable storage format via a Polyglot database.

## Entity Definition

An **Entity** is defined by a set of core parameters: **Metadata**, **Relationships**, and **Attributes**. The core parameters are defined such that they are common to any data stored through the system.

* The **Metadata** contains any unstructured data associated with an Entity.
* The **Relationships** refer to how an Entity is connected with other entities.
* The **Attributes** can be data in any format such as tabular, unstructured, graph, or blob. These can also be interpreted as the datasets owned by an entity.

The format of the **Entity** is as follows:

* `Id`: Unique Read-only identifier for the entity
* `Kind`: Read-only entity type classification
* `Created Time`: Read-only timestamp indicating entity creation time
* `Terminated Time`: Nullable timestamp indicating when the entity was terminated
* `Name` (**TimeBasedValue**): TimeBased value representing the entity's name
* `Metadata` (`map<string, Any>`): Flexible key-value map to store arbitrary metadata
* `Attributes` (`map<string, List<TimeBasedValue>>`): TimeBased attributes stored as lists
* `Relationships` (`map<string, Relationship>`): Relationships to other entities

The core attributes of an **Entity** are *Id*, *Kind*, *Created Time*, *Terminated Time*, and *Name*.

**Metadata** is very useful when we have to store unstructured values that can be subject to change from one entity to the other. **Attributes** are defined in such a way that they have the generic capability to store data of any storage type.

**Relationships** are defined as a map in which the key is a unique identity and the value is a relationship. Note that the relationship contains the type (referred to as name) where one entity can have many relationships of the same type. This resembles the connections an entity has with other entities. The general idea of this specification is to provide a generic model to represent a workflow, an event, or static content in the real world.

## Kind

**OpenGIN** introduces a type system to interpret and represent the entities in a generalized manner. The type system defined in OpenGIN follows the definition of MIME (Multipurpose Internet Mail Extensions) and is named as **Kind**.

**Kind** refers to a classification of various entities based on the nature of existence. It is defined by following the MIME type definition, where a major and a minor component together define a Kind.

* `Major`: Base category of the type
* `Minor`: Sub-category of the type

For instance, when we define an entity like a "Department of Education," the major of the Kind could be *Organization*, and the minor of the *Kind* could be *Department*. This information needs to be determined before creating a dataset for insertion. Once the major and the minor are selected for an entity, they cannot be changed once it is inserted into the system.

## TimeBasedValue

Any value except for immutable types is defined as a `TimeBasedValue`. This value has a start and an end time. One of the major purposes of OpenGIN is to record with time sensitivity. Also, the value of a record is defined as *Any* (protobuf) in order to support all data types and custom data types as decided by the user.

It enables temporal versioning of values with the following fields:

* `Start Time`: Timestamp when the value becomes active
* `End Time`: Timestamp when the value becomes inactive (nullable)
* `Value` (`Any`): Value to be stored of any type

> **Example:** A `TimeBasedValue` would be: *Start Time=2025-01-10*, *End Time=N/A*, *Value=Facebook Handle of a user*.

## Metadata

**Metadata** in OpenGIN provides a flexible mechanism to store unstructured, key-value data associated with entities. Unlike **Attributes** which are time-based and stored in PostgreSQL, metadata is schema-less and stored in MongoDB, making it ideal for storing arbitrary information that doesn't require temporal tracking or complex querying.

Metadata is defined as a `map<string, Any>` where:
* **Key**: A string identifier for the metadata field
* **Value**: Any protobuf `Any` type, allowing for maximum flexibility

## Relationship

**Relationship** defines the connection between two entities. Any parameter that changes with time is easy to process with OpenGIN. Likewise, a Relationship also contains the temporal values in the definition along with a direction.

A Relationship can be defined from one entity to another only in one direction, but it could be queried as an incoming or an outgoing relationship. "Incoming" refers to a relationship originated from another entity towards the referred entity, and "outgoing" refers to that of the opposite.

* `Id`: Unique identifier for the relationship
* `Related Entity Id`: ID of the related entity
* `Name`: Name or type of the relationship
* `Start Time`: Timestamp when the relationship begins
* `End Time`: Timestamp when the relationship ends (nullable)
* `Direction`: Direction of the relationship (Incoming or Outgoing)

**Example:** A Relationship definition could include a scenario where we need to define a relationship between an organization and its employees. The entities here are *Employee* and *Organization*. The Relationship is *HIRED_AS* at a given time. The start time refers to the moment this employee gets into the organization. When that employee no longer continues to work, this relationship comes to an end, and the end time is updated.

## Attribute

OpenGIN considers that an **Entity** has a sense of belonging to data that originated through it or which are part of its core definition. To represent this, OpenGIN supports various storage types since various data can take various formats. Thus, one of the main objectives of OpenGIN is to provide a variety of storage formats.

This is motivated by two main reasons:
1.  **Storage Representation:** Representing entities and their connections in a traditional primary-key and foreign-key approach through a tabular data storage format may not be practical when those connections get denser (Vicknair et al., 2010). At scale, this implies the usage of a high-performance graph database.
2.  **Efficient Data Ownership:** There is a necessity to efficiently store various datasets owned by each entity. These attributes can be in various forms, such as structured, unstructured, graph, or blob data.

From the aforementioned cases, the necessity of a **polyglot database** is justified.

OpenGIN automatically detects and classifies the core storage types when attributes are entered into the system. The system uses a hierarchical detection approach with the following precedence order for the four core storage types:

1.  Graph
2.  Tabular
3.  Document
4.  Blob[^1]

[^1]: *Blob storage format has not yet been released.*

## Key Features

- **Temporal Data**: Native support for time-based values (startTime, endTime) for attributes and relationships.
- **Graph Capabilities**: Powerful relationship traversal and querying.
- **Scalability**: Microservices architecture allows independent scaling of components.
- **Strict Contracts**: Uses Protobuf for internal communication and OpenAPI for external REST APIs.
