# Copyright 2025 Lanka Data Foundation
# SPDX-License-Identifier: Apache-2.0

# MongoDB Server Dockerfile with GitHub backup restore
FROM ubuntu:20.04

# Install prerequisites
RUN apt-get update && apt-get install -y \
    wget gnupg2 curl apt-transport-https sudo \
    && rm -rf /var/lib/apt/lists/*

# Add MongoDB GPG key and repository for 4.4
RUN wget -qO - https://www.mongodb.org/static/pgp/server-4.4.asc | apt-key add - \
    && echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu focal/mongodb-org/4.4 multiverse" > /etc/apt/sources.list.d/mongodb-org-4.4.list

# Install MongoDB 4.4 and additional tools
RUN apt-get update && apt-get install -y \
    mongodb-org=4.4.28 \
    mongodb-org-server=4.4.28 \
    mongodb-org-shell=4.4.28 \
    mongodb-org-tools=4.4.28 \
    wget unzip \
    && rm -rf /var/lib/apt/lists/*

# Create choreo user and group (required for Choreo platform)
RUN groupadd -g 10014 choreo && \
    useradd -u 10014 -g choreo -s /bin/bash -m choreouser && \
    mkdir -p /home/choreouser/.cache && \
    chown -R choreouser:choreo /home/choreouser && \
    chmod -R 755 /home/choreouser && \
    echo "choreouser ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

# Create directories with proper permissions for choreo user
RUN mkdir -p /var/lib/mongodb /var/log/mongodb /data/db /data/configdb /data/backup /var/run/mongodb /tmp/db \
    && chown -R 10014:10014 /var/lib/mongodb /var/log/mongodb /data/db /data/configdb /data/backup /var/run/mongodb /tmp/db \
    && chmod -R 755 /var/lib/mongodb /var/log/mongodb /data/db /data/configdb /data/backup /var/run/mongodb /tmp/db

# Set environment variables
ENV MONGO_INITDB_ROOT_USERNAME=admin
ENV MONGO_INITDB_ROOT_PASSWORD=admin123
ENV MONGO_INITDB_DATABASE=opengin

# GitHub backup restore configuration
ENV GITHUB_BACKUP_REPO=${OPENGIN_GITHUB_BACKUP_REPO:-LDFLK/data-backups} \
    BACKUP_VERSION=${OPENGIN_DB_BACKUP_VERSION:-0.0.4} \
    BACKUP_ENVIRONMENT=${OPENGIN_CHOREO_ENVIRONMENT:-development} \
    RESTORE_FROM_GITHUB=true

# Create MongoDB configuration (without fork for choreo user)
RUN echo "net:\n\
  port: 27017\n\
  bindIp: 0.0.0.0\n\
storage:\n\
  dbPath: /tmp/db\n\
systemLog:\n\
  destination: file\n\
  logAppend: true\n\
  path: /var/log/mongodb/mongod.log\n\
processManagement:\n\
  fork: false" > /etc/mongod.conf

# Additional directories are already created above with proper permissions

# Create entrypoint script with GitHub backup restore
RUN echo '#!/bin/bash\n\
set -e\n\
\n\
# Logging function\n\
log() {\n\
    echo "[$(date +%Y-%m-%d\ %H:%M:%S)] $1: $2"\n\
}\n\
\n\
# Ensure choreo user has proper permissions (volumes may reset ownership)\n\
log "INFO" "Setting up permissions for choreo user..."\n\
sudo chown -R 10014:10014 /var/lib/mongodb /var/log/mongodb /data/db /data/configdb /data/backup /var/run/mongodb /tmp/db\n\
sudo chmod -R 755 /var/lib/mongodb /var/log/mongodb /data/db /data/configdb /data/backup /var/run/mongodb /tmp/db\n\
\n\
# Function to restore from GitHub backup\n\
restore_from_github() {\n\
    local github_repo="${GITHUB_BACKUP_REPO:-LDFLK/data-backups}"\n\
    local version="${BACKUP_VERSION:-0.0.4}"\n\
    local environment="${BACKUP_ENVIRONMENT:-development}"\n\
    \n\
    log "INFO" "Starting MongoDB GitHub backup restore..."\n\
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
    if ! unzip -q "$archive_file" -d "$temp_dir"; then\n\
        log "ERROR" "Failed to extract backup archive"\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
    \n\
    # Find the MongoDB backup file\n\
    local archive_dir="$temp_dir/data-backups-$version"\n\
    local mongodb_backup="$archive_dir/opengin/$environment/mongodb/opengin.tar.gz"\n\
    \n\
    if [ ! -f "$mongodb_backup" ]; then\n\
        log "ERROR" "MongoDB backup file not found: $mongodb_backup"\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
    \n\
    log "INFO" "Found MongoDB backup: $(basename "$mongodb_backup")"\n\
    \n\
    # Extract the backup file\n\
    local backup_extract_dir="$temp_dir/mongodb_restore"\n\
    mkdir -p "$backup_extract_dir"\n\
    \n\
    if ! tar -xzf "$mongodb_backup" -C "$backup_extract_dir"; then\n\
        log "ERROR" "Failed to extract MongoDB backup"\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
    \n\
    # Check if opengin database exists in the backup\n\
    if [ ! -d "$backup_extract_dir/opengin" ]; then\n\
        log "ERROR" "opengin database not found in backup"\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
    \n\
    # Handle nested backup structure\n\
    local backup_source="$backup_extract_dir/opengin"\n\
    if [ -d "$backup_extract_dir/opengin/opengin" ]; then\n\
        log "INFO" "Found nested backup structure, using inner directory"\n\
        backup_source="$backup_extract_dir/opengin/opengin"\n\
    fi\n\
    \n\
    # Copy backup to MongoDB backup directory\n\
    log "INFO" "Preparing backup for mongorestore..."\n\
    cp -r "$backup_source" /data/backup/opengin\n\
    \n\
    # Clean up any macOS metadata files\n\
    find /data/backup/opengin -name "._*" -delete 2>/dev/null || true\n\
    \n\
    # Wait for MongoDB to be ready\n\
    log "INFO" "Waiting for MongoDB to be ready..."\n\
    for i in {1..30}; do\n\
        if mongo --eval "db.adminCommand('\''ping'\'')" > /dev/null 2>&1; then\n\
            log "INFO" "MongoDB is ready!"\n\
            break\n\
        fi\n\
        log "INFO" "Waiting for MongoDB... attempt $i/30"\n\
        sleep 2\n\
    done\n\
    \n\
    # Restore using mongorestore (following the documentation approach)\n\
    log "INFO" "Restoring MongoDB database using mongorestore..."\n\
    if mongorestore --host=localhost:27017 \\\n\
        --db=opengin \\\n\
        --drop \\\n\
        /data/backup/opengin; then\n\
        \n\
        log "SUCCESS" "MongoDB database restored successfully using mongorestore"\n\
        # Clean up backup files\n\
        rm -rf /data/backup/opengin\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    else\n\
        log "ERROR" "Failed to restore MongoDB database using mongorestore"\n\
        rm -rf /data/backup/opengin\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
    \n\
    rm -rf "$temp_dir"\n\
    return 0\n\
}\n\
\n\
# Start MongoDB in background first\n\
log "INFO" "Starting MongoDB in background..."\n\
mongod --dbpath /tmp/db --logpath /var/log/mongodb/mongod.log --bind_ip_all &\n\
MONGODB_PID=$!\n\
\n\
# Wait for MongoDB to start\n\
log "INFO" "Waiting for MongoDB to start..."\n\
for i in {1..30}; do\n\
    if mongo --eval "db.adminCommand('\''ping'\'')" > /dev/null 2>&1; then\n\
        log "INFO" "MongoDB is ready!"\n\
        break\n\
    fi\n\
    log "INFO" "Waiting for MongoDB... attempt $i/30"\n\
    sleep 2\n\
done\n\
\n\
# Create admin user if it doesn'\''t exist (similar to official mongo image behavior)\n\
if [ -n "${MONGO_INITDB_ROOT_USERNAME}" ] && [ -n "${MONGO_INITDB_ROOT_PASSWORD}" ]; then\n\
    log "INFO" "Checking for admin user..."\n\
    if ! mongo admin --quiet --eval "db.getUser('\''${MONGO_INITDB_ROOT_USERNAME}'\'')" 2>/dev/null | grep -q "user"; then\n\
        log "INFO" "Creating admin user..."\n\
        mongo admin --quiet > /dev/null 2>&1 <<< "db.createUser({user: '\''${MONGO_INITDB_ROOT_USERNAME}'\'', pwd: '\''${MONGO_INITDB_ROOT_PASSWORD}'\'', roles: [{role: '\''root'\'', db: '\''admin'\''}]})"\n\
        log "SUCCESS" "Admin user created"\n\
        # Create user in opengin db as well if needed, but not strictly required if using admin auth\n\
    fi\n\
fi\n\
\n\
# Restore from GitHub if enabled and database doesn'\''t exist\n\
if [ "${RESTORE_FROM_GITHUB:-false}" = "true" ]; then\n\
    # Check if opengin database exists\n\
    if ! mongo --eval "db.adminCommand('\''listDatabases'\'')" 2>/dev/null | grep -q "opengin"; then\n\
        log "INFO" "opengin database not found, starting GitHub restore..."\n\
        restore_from_github || log "WARNING" "GitHub restore failed, continuing with empty database"\n\
    else\n\
        log "INFO" "opengin database already exists, skipping restore"\n\
        # List collections for verification\n\
        mongo opengin --quiet --eval "db.getCollectionNames()" || true\n\
    fi\n\
fi\n\
\n\
# Stop the background MongoDB and start in foreground\n\
log "INFO" "Stopping background MongoDB and starting in foreground..."\n\
kill $MONGODB_PID 2>/dev/null || true\n\
wait $MONGODB_PID 2>/dev/null || true\n\
sleep 3\n\
\n\
# Start MongoDB in foreground\n\
log "INFO" "Starting MongoDB in foreground mode..."\n\
exec mongod --dbpath /tmp/db --logpath /var/log/mongodb/mongod.log --bind_ip_all' > /custom-entrypoint.sh \
    && chmod +x /custom-entrypoint.sh

# Switch to choreo user (required for Choreo platform)
# ----------------------------------------------------------------------
# NEW STEP: Build-time Data Ingestion
# ----------------------------------------------------------------------
USER root

# Run the build-time restore
RUN mkdir -p /tmp/db && chown -R 10014:10014 /tmp/db && chmod -R 755 /tmp/db && \
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
    echo "Starting temporary MongoDB..." && \
    mongod --fork --logpath /var/log/mongodb/mongod_build.log --dbpath /tmp/db --bind_ip 127.0.0.1 && \
    \
    # Wait for MongoDB
    until mongo --eval "db.adminCommand('ping')" > /dev/null 2>&1; do sleep 1; done && \
    \
    # Setup Admin User (so we can restore)
    mongo admin --quiet --eval "db.createUser({user: '${MONGO_INITDB_ROOT_USERNAME}', pwd: '${MONGO_INITDB_ROOT_PASSWORD}', roles: [{role: 'root', db: 'admin'}]})" && \
    \
    echo "Restoring Database..." && \
    # The backup extraction path logic needs to match entrypoint logic roughly
    archive_dir="$temp_dir/data-backups-$VERSION" && \
    mongodb_backup="$archive_dir/opengin/$ENV_NAME/mongodb/opengin.tar.gz" && \
    \
    # Extract the tar.gz payload inside the zip
    mkdir -p "$temp_dir/restore_source" && \
    tar -xzf "$mongodb_backup" -C "$temp_dir/restore_source" && \
    \
    restore_path="$temp_dir/restore_source" && \
    if [ -d "$temp_dir/restore_source/opengin" ]; then restore_path="$temp_dir/restore_source/opengin"; fi && \
    if [ -d "$temp_dir/restore_source/opengin/opengin" ]; then restore_path="$temp_dir/restore_source/opengin/opengin"; fi && \
    \
    echo "Running mongorestore from $restore_path..." && \
    mongorestore --host=localhost:27017 --username=${MONGO_INITDB_ROOT_USERNAME} --password=${MONGO_INITDB_ROOT_PASSWORD} --authenticationDatabase=admin --drop "$restore_path" && \
    \
    echo "Shutting down MongoDB..." && \
    mongo admin --username=${MONGO_INITDB_ROOT_USERNAME} --password=${MONGO_INITDB_ROOT_PASSWORD} --authenticationDatabase=admin --eval "db.shutdownServer()" && \
    \
    echo "Cleanup..." && \
    rm -rf "$temp_dir" && \
    # Ensure permissions are correct after all root operations
    chown -R 10014:10014 /tmp/db /var/log/mongodb

USER 10014

# Define volumes for data persistence
VOLUME ["/tmp/db", "/data/configdb", "/data/backup"]

# Expose ports
EXPOSE 27017

# Health check
HEALTHCHECK --interval=10s --timeout=10s --start-period=60s --retries=10 \
    CMD mongo --eval "db.adminCommand('ping')" || exit 1

# Use custom entrypoint
CMD ["/custom-entrypoint.sh"]
