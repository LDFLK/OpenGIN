---

title: Why OpenGIN?
sidebar_label: Why OpenGIN?
---

# Why OpenGIN?

OpenGIN (Open General Information Network) is designed to solve the complexity of modeling and simulating valid digital twins for large-scale ecosystems. Traditional data platforms often struggle with the duality of time (business validity vs. system record) and the diversity of data structures required to represent a real-world entity completely. OpenGIN introduces a specification to model an institute, organization, business, government, or any ecosystem with precision. It provides a framework to build a digital twin of the ecosystem with all its complexities and nuances.

## The Core Problems

### 1. The Time Dimension
Most systems record **what** happened, but fail to accurately model **when** it was valid.
*   *When did this employee actually join the department?*
*   *When was this data entered into the system?*

OpenGIN built its core around a **Temporal Data Model**. Every value can track both its **validity time** (real-world applicability) and **transaction time** (system record), enabling precise time-travel queries and simulation.

### 2. Data Diversity (The Polyglot Challenge)
Real-world entities are complex.
*   **Relationships** are best represented as a Graph.
*   **Metadata** is often unstructured and schema-less.
*   **Time-Series** attributes require rigid, efficient storage.
*   **Unstructured Data** (documents, images) needs blob storage.

Forcing all this into a single database type (like just SQL or just NoSQL) leads to compromises. OpenGIN embraces **Polyglot Storage**, automatically persisting data in its most suitable format while providing a unified API to access it.

### 3. The Connectivity and Tracing

Once an eco-system is built, it is important to understand the connectivity and tracing of the data. OpenGIN provides a framework to understand the connectivity and tracing of the data. It provides a way to query the data in a way that is easy to understand and use. 
