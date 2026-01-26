# Copyright 2025 Lanka Data Foundation
# SPDX-License-Identifier: Apache-2.0

# Start from Ubuntu base image
FROM ubuntu:22.04

# Install prerequisites
RUN apt-get update && apt-get install -y \
    wget gnupg2 curl apt-transport-https openjdk-17-jdk \
    && rm -rf /var/lib/apt/lists/*

# Add Neo4j GPG key and repository
RUN wget -O - https://debian.neo4j.com/neotechnology.gpg.key | apt-key add - \
    && echo "deb https://debian.neo4j.com stable 5" > /etc/apt/sources.list.d/neo4j.list

# Install Neo4j and additional tools for backup restore
RUN apt-get update && apt-get install -y neo4j=1:5.12.0 \
    wget unzip curl \
    && rm -rf /var/lib/apt/lists/*

# Create choreo user and group (required for Choreo platform)
RUN groupadd -g 10014 choreo && \
    useradd -u 10014 -g choreo -s /bin/bash -m choreouser && \
    mkdir -p /home/choreouser/.cache && \
    chown -R choreouser:choreo /home/choreouser && \
    chmod -R 755 /home/choreouser

# Environment defaults (can be overridden at runtime)
ENV NEO4J_AUTH=neo4j/neo4j123 \
    NEO4J_server_memory_pagecache_size=512M \
    NEO4J_server_memory_heap_initial__size=512M \
    NEO4J_server_memory_heap_max__size=512M \
    NEO4J_server_security_procedures_unrestricted=gds.*,apoc.* \
    NEO4J_server_directories_data=/neo4j_data \
    NEO4J_server_directories_logs=/neo4j_logs \
    NEO4J_server_directories_import=/var/lib/neo4j/import \
    NEO4J_server_directories_plugins=/neo4j_plugins \
    NEO4J_server_default__listen__address=0.0.0.0 \
    NEO4J_server_bolt_listen__address=0.0.0.0:7687 \
    NEO4J_server_http_listen__address=0.0.0.0:7474 \
    GITHUB_BACKUP_REPO=${OPENGIN_GITHUB_BACKUP_REPO:-LDFLK/data-backups} \
    BACKUP_VERSION=${OPENGIN_DB_BACKUP_VERSION:-0.0.4} \
    BACKUP_ENVIRONMENT=${OPENGIN_CHOREO_ENVIRONMENT:-development} \
    RESTORE_FROM_GITHUB=true


# ----------------------------------------------------------------------
# NEW STEP: Build-time Data Ingestion
# ----------------------------------------------------------------------
# We switch to root to perform operations, will switch back to choreo later
USER root

# 1. Create all necessary directories
RUN mkdir -p /neo4j_data /neo4j_logs /neo4j_plugins /var/log/neo4j /var/lib/neo4j/data /var/lib/neo4j/run

# 2. Download and Restore NOW (During Build)
# FIXME: https://github.com/LDFLK/OpenGIN/issues/425
RUN apt-get update && apt-get install -y wget unzip neo4j=1:5.12.0 && \
    # Define variables here for the build scope
    GITHUB_REPO="${GITHUB_BACKUP_REPO}" && \
    VERSION="${BACKUP_VERSION}" && \
    ENV_NAME="${BACKUP_ENVIRONMENT}" && \
    \
    # Create temp workspace
    temp_dir=$(mktemp -d) && \
    echo "Downloading backup..." && \
    wget -q "https://github.com/$GITHUB_REPO/archive/refs/tags/$VERSION.zip" -O "$temp_dir/archive.zip" && \
    \
    echo "Unzipping..." && \
    unzip -q "$temp_dir/archive.zip" -d "$temp_dir" && \
    \
    echo "Restoring Database..." && \
    neo4j-admin database load neo4j \
    --from-path="$temp_dir/data-backups-$VERSION/opengin/$ENV_NAME/neo4j" \
    --overwrite-destination=true && \
    \
    # Cleanup to keep image small
    # Cleanup to keep image small
    rm -rf "$temp_dir" && \
    # Fix permissions so 'choreo' user can own the data we just wrote
    chown -R 10014:10014 /neo4j_data /neo4j_logs /neo4j_plugins /var/log/neo4j /var/lib/neo4j /etc/neo4j

# ----------------------------------------------------------------------

# Configure Neo4j to listen on all interfaces
RUN echo "server.default_listen_address=0.0.0.0" >> /etc/neo4j/neo4j.conf \
    && echo "server.bolt.listen_address=0.0.0.0:7687" >> /etc/neo4j/neo4j.conf \
    && echo "server.http.listen_address=0.0.0.0:7474" >> /etc/neo4j/neo4j.conf \
    && echo "server.https.listen_address=0.0.0.0:7473" >> /etc/neo4j/neo4j.conf

# Switch to choreo user
USER 10014

# Create entrypoint script to handle AUTH
RUN echo '#!/bin/bash\n\
    set -e\n\
    \n\
    # Logging function\n\
    log() {\n\
    echo "[$(date +%Y-%m-%d\ %H:%M:%S)] $1: $2"\n\
    }\n\
    \n\
    # Handle NEO4J_AUTH\n\
    if [ -n "${NEO4J_AUTH:-}" ]; then\n\
    if [ "${NEO4J_AUTH}" = "none" ]; then\n\
    log "INFO" "Disabling authentication..."\n\
    echo "dbms.security.auth_enabled=false" >> /etc/neo4j/neo4j.conf\n\
    elif [[ "${NEO4J_AUTH}" == neo4j/* ]]; then\n\
    password="${NEO4J_AUTH#neo4j/}"\n\
    log "INFO" "Setting initial password..."\n\
    # Try to set initial password (works if system db is fresh)\n\
    neo4j-admin dbms set-initial-password "$password" || log "WARNING" "Failed to set initial password (user might already exist)"\n\
    fi\n\
    fi\n\
    \n\
    log "INFO" "Starting Neo4j..."\n\
    exec neo4j console' > /home/choreouser/docker-entrypoint.sh \
    && chmod +x /home/choreouser/docker-entrypoint.sh

# Now declare volumes (Data is already inside /neo4j_data)
# WARNING: Any data written to /neo4j_data BEFORE this line persists. 
# Any data written AFTER this line in a RUN instruction would be lost.
VOLUME ["/neo4j_data", "/neo4j_logs", "/var/lib/neo4j/import", "/neo4j_plugins"]

# Expose ports
EXPOSE 7474 7687

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:7474 || exit 1

# Use entrypoint
CMD ["/home/choreouser/docker-entrypoint.sh"]