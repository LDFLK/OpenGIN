# Frequently Asked Questions

## General

### What is OpenGIN?
OpenGIN is a platform for building time-aware digital twins using a polyglot database architecture.

### Why does it use three databases?
OpenGIN leverages the strengths of multiple databases:
- **MongoDB**: For flexible, schema-less metadata.
- **Neo4j**: For handling complex relationships and graph traversals.
- **PostgreSQL**: For storing time-series attributes and ensuring ACID compliance for structured data.

## Technical

### How do I reset the databases?
You can stop the containers and remove the volumes:
```bash
docker-compose down -v
```
Then start them up again:
```bash
docker-compose up -d
```

### Can I add my own Entity types?
Yes, Entities are flexible. You define the `Kind` (Major/Minor) and the system adapts. You don't need to pre-define schemas in the database.
