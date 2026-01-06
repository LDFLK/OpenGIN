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
    BACKUP_VERSION=${OPENGIN_DB_BACKUP_VERSION:-0.0.1} \
    BACKUP_ENVIRONMENT=${OPENGIN_CHOREO_ENVIRONMENT:-development} \
    RESTORE_FROM_GITHUB=true

# Create data directories with proper permissions for choreo user
RUN mkdir -p /neo4j_data /neo4j_logs /neo4j_plugins /var/log/neo4j /var/lib/neo4j/data /var/lib/neo4j/run \
    && chown -R 10014:10014 /neo4j_data /neo4j_logs /neo4j_plugins /var/log/neo4j /var/lib/neo4j /var/lib/neo4j/run \
    && chmod -R 755 /neo4j_data /neo4j_logs /neo4j_plugins /var/log/neo4j /var/lib/neo4j /var/lib/neo4j/run

# Configure Neo4j to listen on all interfaces
RUN echo "server.default_listen_address=0.0.0.0" >> /etc/neo4j/neo4j.conf \
    && echo "server.bolt.listen_address=0.0.0.0:7687" >> /etc/neo4j/neo4j.conf \
    && echo "server.http.listen_address=0.0.0.0:7474" >> /etc/neo4j/neo4j.conf \
    && echo "server.https.listen_address=0.0.0.0:7473" >> /etc/neo4j/neo4j.conf

# Entrypoint script with GitHub backup restore functionality
RUN echo '#!/bin/bash\n\
set -e\n\
\n\
# Logging function\n\
log() {\n\
    echo "[$(date +%Y-%m-%d\ %H:%M:%S)] $1: $2"\n\
}\n\
\n\
# Function to restore from GitHub backup\n\
restore_from_github() {\n\
    local github_repo="${GITHUB_BACKUP_REPO:-LDFLK/data-backups}"\n\
    local version="${BACKUP_VERSION:-0.0.1}"\n\
    local environment="${BACKUP_ENVIRONMENT:-development}"\n\
    \n\
    log "INFO" "Starting GitHub backup restore..."\n\
    log "INFO" "Repository: $github_repo"\n\
    log "INFO" "Version: $version"\n\
    log "INFO" "Environment: $environment"\n\
    \n\
    # Create temporary directory for download\n\
    local temp_dir=$(mktemp -d)\n\
    local archive_url="https://github.com/$github_repo/archive/refs/tags/$version.zip"\n\
    local archive_file="$temp_dir/archive.zip"\n\
    \n\
    log "INFO" "Downloading backup from: $archive_url"\n\
    if ! wget -q "$archive_url" -O "$archive_file"; then\n\
        log "ERROR" "Failed to download backup archive"\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
    \n\
    # Extract the archive\n\
    log "INFO" "Extracting backup archive..."\n\
    if ! unzip -q "$archive_file" -d "$temp_dir"; then\n\
        log "ERROR" "Failed to extract backup archive"\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
    \n\
    # Find the Neo4j backup file\n\
    local archive_dir="$temp_dir/data-backups-$version"\n\
    local neo4j_backup="$archive_dir/opengin/version/$version/$environment/neo4j/neo4j.dump"\n\
    \n\
    if [ ! -f "$neo4j_backup" ]; then\n\
        log "ERROR" "Neo4j backup file not found: $neo4j_backup"\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
    \n\
    log "INFO" "Found Neo4j backup: $(basename "$neo4j_backup")"\n\
    \n\
    # Check if database already exists\n\
    if [ -d "${NEO4J_server_directories_data}/databases/neo4j" ]; then\n\
        log "INFO" "Database already exists, skipping restore"\n\
        rm -rf "$temp_dir"\n\
        return 0\n\
    fi\n\
    \n\
    # Restore the database\n\
    log "INFO" "Restoring Neo4j database..."\n\
    if neo4j-admin database load neo4j --from-path="$(dirname "$neo4j_backup")" --overwrite-destination=true; then\n\
        log "SUCCESS" "Neo4j database restored successfully"\n\
        rm -rf "$temp_dir"\n\
        return 0\n\
    else\n\
        log "ERROR" "Failed to restore Neo4j database"\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
}\n\
\n\
# Handle NEO4J_AUTH like the official image\n\
if [ -n "${NEO4J_AUTH:-}" ]; then\n\
  if [ "${NEO4J_AUTH}" = "none" ]; then\n\
    log "INFO" "Disabling authentication..."\n\
    echo "dbms.security.auth_enabled=false" >> /etc/neo4j/neo4j.conf\n\
  elif [[ "${NEO4J_AUTH}" == neo4j/* ]]; then\n\
    password="${NEO4J_AUTH#neo4j/}"\n\
    if [ ! -d "${NEO4J_server_directories_data}/databases" ]; then\n\
      log "INFO" "Setting initial password..."\n\
      neo4j-admin dbms set-initial-password "$password" || true\n\
    fi\n\
  fi\n\
fi\n\
\n\
# Restore from GitHub if enabled and database doesn'\''t exist\n\
if [ "${RESTORE_FROM_GITHUB:-false}" = "true" ] && [ ! -d "${NEO4J_server_directories_data}/databases/neo4j" ]; then\n\
    restore_from_github || log "WARNING" "GitHub restore failed, continuing with empty database"\n\
fi\n\
\n\
log "INFO" "Starting Neo4j..."\n\
exec neo4j console' > /docker-entrypoint.sh \
    && chmod +x /docker-entrypoint.sh

# Switch to choreo user (required for Choreo platform)
USER 10014

# Expose ports
EXPOSE 7474 7687

# Volumes for persistence
VOLUME ["/neo4j_data", "/neo4j_logs", "/var/lib/neo4j/import", "/neo4j_plugins"]

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:7474 || exit 1

# Use entrypoint
CMD ["/docker-entrypoint.sh"]
