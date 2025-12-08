#!/bin/bash
set -e

# Database cleanup script for CORE architecture
# Usage: ./cleanup.sh [pre|post]
#   pre  - Clean databases before starting services
#   post - Clean databases after all services complete

PHASE=${1:-pre}

if [[ "$PHASE" != "pre" && "$PHASE" != "post" ]]; then
    echo "Error: Invalid phase. Use 'pre' or 'post'"
    echo "Usage: $0 [pre|post]"
    exit 1
fi

echo "=== Database Cleanup: $PHASE Phase ==="
echo "Starting cleanup at: $(date)"

# Function to clean PostgreSQL tables
clean_postgresql() {
    echo "🧹 Cleaning PostgreSQL tables..."
    
    # Clean specific tables with DELETE CASCADE
    echo "  - Cleaning attribute_schemas and entity_attributes..."
    # Commented out: Keep attribute_schemas and entity_attributes tables intact
    # PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "DELETE FROM attribute_schemas CASCADE; DELETE FROM entity_attributes CASCADE;" || echo "    ⚠️  Some tables may not exist (this is normal)"
    
    # Clean attr_ prefix tables with DELETE CASCADE
    echo "  - Cleaning attr_ prefix tables..."
    PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "DO \$\$ DECLARE r RECORD; BEGIN FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE 'attr_%') LOOP EXECUTE 'DELETE FROM ' || quote_ident(r.tablename) || ' CASCADE'; RAISE NOTICE 'Cleaned attr_ table: %', r.tablename; END LOOP; END \$\$;" || echo "    ⚠️  No attr_ tables found (this is normal)"
    
    echo "  ✅ PostgreSQL cleanup completed"
}

# Function to clean MongoDB collections
clean_mongodb() {
    echo "🧹 Cleaning MongoDB collections..."
    
    # Clean metadata collections
    echo "  - Cleaning metadata collections..."
    mongosh --eval "db = db.getSiblingDB(\"$MONGO_DB_NAME\"); if (db.metadata) { db.metadata.drop(); print('Dropped metadata collection'); } if (db.metadata_test) { db.metadata_test.drop(); print('Dropped metadata_test collection'); }" $MONGO_URI || echo "    ⚠️  Some collections may not exist (this is normal)"
    
    echo "  ✅ MongoDB cleanup completed"
}

# Function to clean Neo4j database
clean_neo4j() {
    echo "🧹 Cleaning Neo4j database..."
    
    # Clean all nodes and relationships
    echo "  - Cleaning all nodes and relationships..."
    cypher-shell -u $NEO4J_USER -p $NEO4J_PASSWORD -a $NEO4J_URI "MATCH (n) DETACH DELETE n" || echo "    ⚠️  Neo4j cleanup completed"
    
    echo "  ✅ Neo4j cleanup completed"
}

# Main cleanup execution
echo "🚀 Starting database cleanup operations..."

# Clean all databases
clean_postgresql
clean_mongodb
clean_neo4j

echo ""
echo "🎉 Database cleanup $PHASE phase completed successfully!"
echo "Completed at: $(date)"
echo ""

# Phase-specific messages
if [[ "$PHASE" == "pre" ]]; then
    echo "✨ Databases are now clean and ready for services to start"
elif [[ "$PHASE" == "post" ]]; then
    echo "✨ All services have completed and databases have been cleaned"
fi

