#!/bin/bash
set -e

# Database cleanup script for CORE architecture - CHOREO ENVIRONMENT
# Usage: ./cleanup-choreo.sh [pre|post]
#   pre  - Clean databases before starting services
#   post - Clean databases after all services complete
#
# This script is specifically configured for the Choreo environment
# and uses choreo-specific database connections and configurations.

PHASE=${1:-pre}

if [[ "$PHASE" != "pre" && "$PHASE" != "post" ]]; then
    echo "Error: Invalid phase. Use 'pre' or 'post'"
    echo "Usage: $0 [pre|post]"
    exit 1
fi

echo "=== Database Cleanup: $PHASE Phase (Choreo Environment) ==="
echo "Starting cleanup at: $(date)"

# Function to clean PostgreSQL tables
clean_postgresql() {
    echo "🧹 Cleaning PostgreSQL tables (Choreo)..."
    
    # Clean specific tables with DELETE CASCADE
    echo "  - Cleaning attribute_schemas and entity_attributes..."
    PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "DELETE FROM attribute_schemas CASCADE; DELETE FROM entity_attributes CASCADE;" || echo "    ⚠️  Some tables may not exist (this is normal)"
    
    # Clean attr_ prefix tables with DELETE CASCADE
    echo "  - Cleaning attr_ prefix tables..."
    PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "DO \$\$ DECLARE r RECORD; BEGIN FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE 'attr_%') LOOP EXECUTE 'DELETE FROM ' || quote_ident(r.tablename) || ' CASCADE'; RAISE NOTICE 'Cleaned attr_ table: %', r.tablename; END LOOP; END \$\$;" || echo "    ⚠️  No attr_ tables found (this is normal)"
    
    echo "  ✅ PostgreSQL cleanup completed (Choreo)"
}

# Function to clean MongoDB collections
clean_mongodb() {
    echo "🧹 Cleaning MongoDB collections (Choreo)..."
    
    # Clean metadata collections
    echo "  - Cleaning metadata collections..."
    mongosh --eval "db = db.getSiblingDB(\"$MONGO_DB_NAME\"); if (db.metadata_choreo_test) { db.metadata_choreo_test.drop(); print('Dropped metadata_choreo_test collection'); }" $MONGO_URI || echo "    ⚠️  Some collections may not exist (this is normal)"
    
    echo "  ✅ MongoDB cleanup completed (Choreo)"
}

# Function to clean Neo4j database
clean_neo4j() {
    echo "🧹 Cleaning Neo4j database (Choreo)..."
    
    # Clean all nodes and relationships
    echo "  - Cleaning all nodes and relationships..."
    cypher-shell -u $NEO4J_USER -p $NEO4J_PASSWORD -a $NEO4J_URI "MATCH (n) DETACH DELETE n" || echo "    ⚠️  Neo4j cleanup completed"
    
    echo "  ✅ Neo4j cleanup completed (Choreo)"
}

# Main cleanup execution
echo "🚀 Starting database cleanup operations (Choreo Environment)..."

# Clean all databases
clean_postgresql
clean_mongodb
clean_neo4j

echo ""
echo "🎉 Database cleanup $PHASE phase completed successfully! (Choreo)"
echo "Completed at: $(date)"
echo ""

# Phase-specific messages
if [[ "$PHASE" == "pre" ]]; then
    echo "✨ Choreo databases are now clean and ready for services to start"
elif [[ "$PHASE" == "post" ]]; then
    echo "✨ All Choreo services have completed and databases have been cleaned"
fi

