# Simple Neo4j Server Dockerfile
FROM neo4j:5.12.0

# Set environment variables
ENV NEO4J_AUTH=neo4j/neo4j123
ENV NEO4J_dbms_memory_pagecache_size=512M
ENV NEO4J_dbms_memory_heap_initial__size=512M
ENV NEO4J_dbms_memory_heap_max__size=512M
ENV NEO4J_dbms_security_procedures_unrestricted=gds.*,apoc.*

# Set data directory paths
ENV NEO4J_dbms_directories_data=/neo4j_data
ENV NEO4J_dbms_directories_logs=/neo4j_logs
ENV NEO4J_dbms_directories_import=/var/lib/neo4j/import
ENV NEO4J_dbms_directories_plugins=/neo4j_plugins

# Expose ports
EXPOSE 7474 7687

# Create data directories with proper ownership
RUN mkdir -p /neo4j_data && \
    mkdir -p /neo4j_logs && \
    mkdir -p /neo4j_plugins && \
    chown -R neo4j:neo4j /neo4j_data /neo4j_logs /neo4j_plugins && \
    chmod -R 755 /neo4j_data /neo4j_logs /neo4j_plugins

# Set working directory
WORKDIR /var/lib/neo4j

# Switch to neo4j user
USER neo4j

# Define volumes for data persistence
VOLUME ["/neo4j_data", "/neo4j_logs", "/var/lib/neo4j/import", "/neo4j_plugins"]

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:7474 || exit 1

# Default command
CMD ["neo4j"]
