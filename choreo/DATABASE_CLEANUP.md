# Database Cleanup Best Practices

## 🔧 Cleanup Commands by Database

### **MongoDB Cleanup**
```bash
# Simple cleanup - drop specific collections
mongosh --eval "
db = db.getSiblingDB('$MONGO_DB_NAME'); 
if (db.metadata) { 
  db.metadata.drop(); 
  print('Dropped metadata collection'); 
} 
if (db.metadata_test) { 
  db.metadata_test.drop(); 
  print('Dropped metadata_test collection'); 
}" mongodb://admin:admin123@mongodb:27017/admin?authSource=admin
```

**Features**:
- ✅ Targets only the specific database
- ✅ Drops specific collections: `metadata` and `metadata_test`
- ✅ Simple and fast operation
- ✅ No authorization issues
- ✅ Clear logging of what was dropped

### **Neo4j Cleanup**
```bash
# Clean all nodes and relationships
cypher-shell -u neo4j -p neo4j123 -a bolt://neo4j:7687 "MATCH (n) DETACH DELETE n"
```

**Features**:
- ✅ Removes all nodes and relationships
- ✅ Uses DETACH DELETE to handle relationships properly
- ✅ Simple and reliable

### **PostgreSQL Cleanup**
```sql
-- Safe cleanup with conditional table existence checks
DO $$ 
BEGIN 
  IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'metadata') THEN 
    TRUNCATE TABLE metadata CASCADE; 
  END IF; 
  IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'entities') THEN 
    TRUNCATE TABLE entities CASCADE; 
  END IF; 
  IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'relationships') THEN 
    TRUNCATE TABLE relationships CASCADE; 
  END IF; 
END $$;

-- Comprehensive cleanup - clean specific tables and attr_ prefix tables
DO $$ 
DECLARE r RECORD; 
BEGIN 
  FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND (tablename LIKE 'attr_%' OR tablename IN ('attribute_schemas', 'entity_attributes', 'metadata', 'entities', 'relationships'))) 
  LOOP 
    EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE'; 
    RAISE NOTICE 'Cleaned table: %', r.tablename; 
  END LOOP; 
END $$;

-- Drop test tables if they exist (optional cleanup)
DO $$ 
BEGIN 
  IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'test_metadata') THEN 
    DROP TABLE test_metadata CASCADE; 
  END IF; 
  IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'test_entities') THEN 
    DROP TABLE test_entities CASCADE; 
  END IF; 
END $$;
```

**Features**:
- ✅ Checks table existence before truncating
- ✅ Uses CASCADE to handle foreign key constraints
- ✅ Safe execution with DO block
- ✅ Targets specific tables: `attribute_schemas`, `entity_attributes`, `metadata`, `entities`, `relationships`
- ✅ Automatically cleans all tables with `attr_` prefix
- ✅ Skips system tables (pg_*, sql_*)

## 🚀 Implementation in opengin/core-api/docker/Dockerfile

The cleanup is implemented in two phases:

### **Phase 1: Pre-Test Cleanup**
```bash
# Clean databases before running tests
echo "=== Cleaning Databases Before Tests ==="
# MongoDB, Neo4j, PostgreSQL cleanup
```

### **Phase 2: Post-Test Cleanup**
```bash
# Clean databases after tests complete
echo "=== Cleaning Databases After Tests ==="
# MongoDB, Neo4j, PostgreSQL cleanup
```

## 🧪 Testing the Cleanup

### **Verify MongoDB Cleanup**
```bash
# Check collections in specific database
docker exec mongodb mongosh --eval "db = db.getSiblingDB('testdb'); db.getCollectionNames()"

# Check document counts
docker exec mongodb mongosh --eval "db = db.getSiblingDB('testdb'); db.metadata.countDocuments()"
```

### **Verify Neo4j Cleanup**
```bash
# Check node count
docker exec neo4j cypher-shell -u neo4j -p neo4j123 "MATCH (n) RETURN count(n) as nodeCount"

# Check relationship count
docker exec neo4j cypher-shell -u neo4j -p neo4j123 "MATCH ()-[r]->() RETURN count(r) as relCount"
```

### **Verify PostgreSQL Cleanup**
```bash
# Check table row counts
docker exec postgres psql -U postgres -d opengin -c "SELECT 'metadata' as table_name, COUNT(*) as row_count FROM metadata UNION ALL SELECT 'entities', COUNT(*) FROM entities UNION ALL SELECT 'relationships', COUNT(*) FROM relationships;"
```

## 🎯 Best Practices

### **1. Database-Specific Cleanup**
- Use appropriate commands for each database type
- Respect database-specific syntax and limitations
- Handle errors gracefully

### **2. System Collection Protection**
- Never attempt to clean system collections
- Use whitelist approach for collections to clean
- Add safety checks and error handling

### **3. Transaction Safety**
- Use appropriate transaction boundaries
- Handle foreign key constraints properly
- Ensure cleanup is atomic where possible

### **4. Logging and Monitoring**
- Log all cleanup operations
- Track successful and failed cleanups
- Provide clear error messages

### **5. Performance Considerations**
- Use efficient cleanup commands (TRUNCATE vs DELETE)
- Batch operations where possible
- Consider cleanup timing (before/after tests)

## 🔍 Troubleshooting

### **Common Issues**

1. **MongoDB Authorization Errors**
   - Ensure using correct database
   - Skip system collections
   - Use proper authentication

2. **PostgreSQL Syntax Errors**
   - Use DO blocks for conditional operations
   - Check table existence before operations
   - Use proper SQL syntax

3. **Neo4j Connection Issues**
   - Verify connection parameters
   - Check Neo4j service health
   - Use simple, reliable commands

### **Debug Commands**
```bash
# Check service health
docker compose ps

# View service logs
docker compose logs mongodb neo4j postgres core

# Interactive debugging
docker compose exec mongodb mongosh
docker compose exec neo4j cypher-shell -u neo4j -p neo4j123
docker compose exec postgres psql -U postgres -d opengin
```

## 📚 References

- [MongoDB Collection Operations](https://docs.mongodb.com/manual/reference/method/db.collection.deleteMany/)
- [PostgreSQL DO Blocks](https://www.postgresql.org/docs/current/sql-do.html)
- [Neo4j Cypher DELETE](https://neo4j.com/docs/cypher-manual/current/clauses/delete/)
- [Docker Compose Health Checks](https://docs.docker.com/compose/compose-file/compose-file-v3/#healthcheck)

## 🆕 Dedicated Cleanup Service

### **Overview**
The system now includes a dedicated cleanup service that provides a centralized, reusable way to clean all databases. This service is separate from individual application containers and can be run independently.

### **Service Architecture**

```yaml
# docker-compose.yml
cleanup:
  build:
    context: .
    dockerfile: deployment/development/docker/cleanup/Dockerfile
  container_name: cleanup
  networks:
    - ldf-network
  environment:
    - POSTGRES_HOST=postgres
    - POSTGRES_PORT=5432
    - POSTGRES_USER=postgres
    - POSTGRES_PASSWORD=postgres
    - POSTGRES_DB=opengin
    - MONGO_URI=mongodb://admin:admin123@mongodb:27017/admin?authSource=admin
    - MONGO_DB_NAME=testdb
    - NEO4J_URI=bolt://neo4j:7687
    - NEO4J_USER=neo4j
    - NEO4J_PASSWORD=neo4j123
  depends_on:
    - postgres
    - mongodb
    - neo4j
  profiles:
    - cleanup  # Only starts when explicitly requested
```

### **Cleanup Script (cleanup.sh)**

The service uses a dedicated cleanup script that handles all database types:

```bash
#!/bin/bash
set -e

PHASE=${1:-pre}  # "pre" or "post"

echo "=== Database Cleanup: $PHASE Phase ==="

# PostgreSQL cleanup
clean_postgresql() {
    echo "🧹 Cleaning PostgreSQL tables..."
    PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "DELETE FROM attribute_schemas CASCADE; DELETE FROM entity_attributes CASCADE;"
    
    # Clean attr_ prefix tables
    PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "DO \$\$ DECLARE r RECORD; BEGIN FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE 'attr_%') LOOP EXECUTE 'DELETE FROM ' || quote_ident(r.tablename) || ' CASCADE'; END LOOP; END \$\$;"
}

# MongoDB cleanup
clean_mongodb() {
    echo "🧹 Cleaning MongoDB collections..."
    mongosh --eval "db = db.getSiblingDB(\"$MONGO_DB_NAME\"); if (db.metadata) { db.metadata.drop(); } if (db.metadata_test) { db.metadata_test.drop(); }" mongodb://admin:admin123@mongodb:27017/admin?authSource=admin
}

# Neo4j cleanup
clean_neo4j() {
    echo "🧹 Cleaning Neo4j database..."
    cypher-shell -u neo4j -p neo4j123 -a bolt://neo4j:7687 "MATCH (n) DETACH DELETE n"
}

# Execute cleanup
clean_postgresql
clean_mongodb
clean_neo4j
echo "🎉 Database cleanup $PHASE phase completed!"
```

### **Usage Patterns**

#### **1. Pre-Service Cleanup**
```bash
# Clean databases before starting services
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh pre
```

#### **2. Post-Service Cleanup**
```bash
# Clean databases after services complete
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh post
```

#### **3. On-Demand Cleanup**
```bash
# Clean databases anytime
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh pre
```

### **Benefits of Dedicated Cleanup Service**

1. **Separation of Concerns**
   - Cleanup logic is separate from application services
   - No cleanup code in individual service containers
   - Focused, single-purpose containers

2. **Flexibility**
   - Run cleanup independently of service lifecycle
   - Can clean databases without starting services
   - Easy to integrate into CI/CD pipelines

3. **Reusability**
   - Same cleanup logic for different scenarios
   - Consistent cleanup across environments
   - Easy to maintain and update

4. **Non-Intrusive**
   - Doesn't interfere with normal service operation
   - Uses cleanup profile to avoid automatic startup
   - Clean, controlled execution

5. **Better Testing**
   - Clean databases before running tests
   - Clean databases after test completion
   - Isolated cleanup testing

### **Integration with Existing Workflow**

#### **Before (Embedded Cleanup)**
```bash
# Cleanup was embedded in each service
# Hard to maintain, duplicate code
# Tightly coupled with service lifecycle
```

#### **After (Dedicated Service)**
```bash
# Cleanup is a separate, focused service
# Easy to maintain, single source of truth
# Loosely coupled, run when needed
```

### **Migration from Embedded Cleanup**

If you're migrating from the old embedded cleanup approach:

1. **Remove cleanup logic** from individual service Dockerfiles
2. **Ingestion service startup scripts** to remove cleanup calls
3. **Use the dedicated cleanup service** for all database cleaning needs
4. **Update CI/CD pipelines** to call cleanup service at appropriate times

### **Advanced Usage**

#### **Custom Cleanup Phases**
```bash
# Add custom cleanup phases
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh custom

# Modify cleanup.sh to handle custom phases
case $PHASE in
    "pre") echo "Pre-service cleanup" ;;
    "post") echo "Post-service cleanup" ;;
    "custom") echo "Custom cleanup phase" ;;
    *) echo "Unknown phase: $PHASE" ;;
esac
```

#### **Selective Database Cleanup**
```bash
# Clean only specific databases
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh postgres-only
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh mongo-only
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh neo4j-only
```

### **Monitoring and Logging**

The cleanup service provides comprehensive logging:

```bash
# View cleanup logs
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh pre

# Expected output:
# === Database Cleanup: pre Phase ===
# 🧹 Cleaning PostgreSQL tables...
# 🧹 Cleaning MongoDB collections...
# 🧹 Cleaning Neo4j database...
# 🎉 Database cleanup pre phase completed!
```

### **Troubleshooting the Cleanup Service**

#### **Common Issues**

1. **Service Not Found**
   ```bash
   # Ensure you're using the cleanup profile
   docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh pre
   ```

2. **Database Connection Errors**
   ```bash
   # Check if databases are running
   docker-compose ps postgres mongodb neo4j
   
   # Check database connectivity
   docker-compose exec postgres pg_isready -U postgres
   docker-compose exec mongodb mongosh --eval "db.adminCommand('ping')"
   docker-compose exec neo4j cypher-shell -u neo4j -p neo4j123 "CALL dbms.components()"
   ```

3. **Permission Issues**
   ```bash
   # Ensure cleanup service has proper credentials
   # Check environment variables in docker-compose.yml
   ```

#### **Debug Commands**
```bash
# Run cleanup with verbose output
docker-compose --profile cleanup run --rm cleanup /bin/bash

# Inside container, run cleanup manually
/app/cleanup.sh pre

# Check environment variables
env | grep -E "(POSTGRES|MONGO|NEO4J)"
```
