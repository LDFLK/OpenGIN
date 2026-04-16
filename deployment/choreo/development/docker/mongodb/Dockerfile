# Copyright 2025 Lanka Data Foundation
# SPDX-License-Identifier: Apache-2.0

# Upgraded to Ubuntu 22.04 (Jammy) - LTS supported until 2027
FROM ubuntu:22.04

# Prevent interactive prompts during package install
ENV DEBIAN_FRONTEND=noninteractive

# ✅ FIX: Switch to a reliable mirror BEFORE apt-get install
# Using kernel.org mirror which is globally stable
# Also adding --retry logic and --fix-missing as fallback
RUN sed -i 's|http://archive.ubuntu.com/ubuntu|http://mirrors.edge.kernel.org/ubuntu|g' /etc/apt/sources.list && \
    sed -i 's|http://security.ubuntu.com/ubuntu|http://mirrors.edge.kernel.org/ubuntu|g' /etc/apt/sources.list && \
    apt-get update -o Acquire::Retries=3 && \
    apt-get install -y --fix-missing \
        wget gnupg2 curl apt-transport-https sudo \
    && rm -rf /var/lib/apt/lists/*

# Add MongoDB GPG key and repository
# ✅ FIX: Upgraded to MongoDB 7.0 (actively supported until 2027)
RUN wget -qO - https://www.mongodb.org/static/pgp/server-7.0.asc | \
    gpg --dearmor -o /usr/share/keyrings/mongodb-server-7.0.gpg && \
    echo "deb [ arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-7.0.gpg ] https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/7.0 multiverse" \
    > /etc/apt/sources.list.d/mongodb-org-7.0.list

# Install MongoDB 7.0 and tools
RUN apt-get update -o Acquire::Retries=3 && \
    apt-get install -y --fix-missing \
        mongodb-org \
        wget unzip \
    && rm -rf /var/lib/apt/lists/*

# Create choreo user and group (required for Choreo platform)
RUN groupadd -g 10014 choreo && \
    useradd -u 10014 -g choreo -s /bin/bash -m choreouser && \
    mkdir -p /home/choreouser/.cache && \
    chown -R choreouser:choreo /home/choreouser && \
    chmod -R 755 /home/choreouser && \
    echo "choreouser ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

# Create directories with proper permissions
RUN mkdir -p /var/lib/mongodb /var/log/mongodb /data/db /data/configdb /data/backup /var/run/mongodb \
    && chown -R 10014:10014 /var/lib/mongodb /var/log/mongodb /data/db /data/configdb /data/backup /var/run/mongodb \
    && chmod -R 755 /var/lib/mongodb /var/log/mongodb /data/db /data/configdb /data/backup /var/run/mongodb

# Set environment variables
ENV MONGO_INITDB_ROOT_USERNAME=admin
ENV MONGO_INITDB_ROOT_PASSWORD=admin123
ENV MONGO_INITDB_DATABASE=opengin

# GitHub backup restore configuration
ENV GITHUB_BACKUP_REPO=${OPENGIN_GITHUB_BACKUP_REPO:-LDFLK/data-backups} \
    BACKUP_VERSION=${OPENGIN_DB_BACKUP_VERSION:-0.0.1} \
    BACKUP_ENVIRONMENT=${OPENGIN_CHOREO_ENVIRONMENT:-development} \
    RESTORE_FROM_GITHUB=true

# Create MongoDB configuration
RUN printf "net:\n\
  port: 27017\n\
  bindIp: 0.0.0.0\n\
storage:\n\
  dbPath: /data/db\n\
systemLog:\n\
  destination: file\n\
  logAppend: true\n\
  path: /var/log/mongodb/mongod.log\n\
processManagement:\n\
  fork: false\n" > /etc/mongod.conf

# ✅ FIX: Use printf instead of echo for entrypoint (avoids \n literal issues)
RUN printf '#!/bin/bash\n\
set -e\n\
\n\
log() {\n\
    echo "[$(date +%%Y-%%m-%%d\\ %%H:%%M:%%S)] $1: $2"\n\
}\n\
\n\
log "INFO" "Setting up permissions for choreo user..."\n\
sudo chown -R 10014:10014 /var/lib/mongodb /var/log/mongodb /data/db /data/configdb /data/backup /var/run/mongodb\n\
sudo chmod -R 755 /var/lib/mongodb /var/log/mongodb /data/db /data/configdb /data/backup /var/run/mongodb\n\
\n\
restore_from_github() {\n\
    local github_repo="${GITHUB_BACKUP_REPO:-LDFLK/data-backups}"\n\
    local version="${BACKUP_VERSION:-0.0.1}"\n\
    local environment="${BACKUP_ENVIRONMENT:-development}"\n\
    log "INFO" "Starting MongoDB GitHub backup restore..."\n\
    local temp_dir=$(mktemp -d)\n\
    local archive_url="https://github.com/$github_repo/archive/refs/tags/$version.zip"\n\
    local archive_file="$temp_dir/archive.zip"\n\
    log "INFO" "Downloading backup from: $archive_url"\n\
    if ! wget -q "$archive_url" -O "$archive_file"; then\n\
        log "ERROR" "Failed to download backup archive"\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
    if ! unzip -q "$archive_file" -d "$temp_dir"; then\n\
        log "ERROR" "Failed to extract backup archive"\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
    local archive_dir="$temp_dir/data-backups-$version"\n\
    local mongodb_backup="$archive_dir/opengin/$environment/mongodb/opengin.tar.gz"\n\
    if [ ! -f "$mongodb_backup" ]; then\n\
        log "ERROR" "MongoDB backup file not found: $mongodb_backup"\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
    local backup_extract_dir="$temp_dir/mongodb_restore"\n\
    mkdir -p "$backup_extract_dir"\n\
    if ! tar -xzf "$mongodb_backup" -C "$backup_extract_dir"; then\n\
        log "ERROR" "Failed to extract MongoDB backup"\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
    if [ ! -d "$backup_extract_dir/opengin" ]; then\n\
        log "ERROR" "opengin database not found in backup"\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
    local backup_source="$backup_extract_dir/opengin"\n\
    if [ -d "$backup_extract_dir/opengin/opengin" ]; then\n\
        backup_source="$backup_extract_dir/opengin/opengin"\n\
    fi\n\
    cp -r "$backup_source" /data/backup/opengin\n\
    find /data/backup/opengin -name "._*" -delete 2>/dev/null || true\n\
    log "INFO" "Waiting for MongoDB to be ready..."\n\
    for i in {1..30}; do\n\
        if mongosh --eval "db.adminCommand({ping:1})" > /dev/null 2>&1; then\n\
            log "INFO" "MongoDB is ready!"\n\
            break\n\
        fi\n\
        log "INFO" "Waiting for MongoDB... attempt $i/30"\n\
        sleep 2\n\
    done\n\
    log "INFO" "Restoring MongoDB database..."\n\
    if mongorestore --host=localhost:27017 --db=opengin --drop /data/backup/opengin; then\n\
        log "SUCCESS" "MongoDB restored successfully"\n\
        rm -rf /data/backup/opengin\n\
    else\n\
        log "ERROR" "Failed to restore MongoDB"\n\
        rm -rf /data/backup/opengin "$temp_dir"\n\
        return 1\n\
    fi\n\
    rm -rf "$temp_dir"\n\
    return 0\n\
}\n\
\n\
log "INFO" "Starting MongoDB in background..."\n\
mongod --dbpath /data/db --logpath /var/log/mongodb/mongod.log --bind_ip_all &\n\
MONGODB_PID=$!\n\
\n\
for i in {1..30}; do\n\
    if mongosh --eval "db.adminCommand({ping:1})" > /dev/null 2>&1; then\n\
        log "INFO" "MongoDB is ready!"\n\
        break\n\
    fi\n\
    log "INFO" "Waiting for MongoDB... attempt $i/30"\n\
    sleep 2\n\
done\n\
\n\
if [ -n "${MONGO_INITDB_ROOT_USERNAME}" ] && [ -n "${MONGO_INITDB_ROOT_PASSWORD}" ]; then\n\
    log "INFO" "Checking for admin user..."\n\
    if ! mongosh admin --quiet --eval "db.getUser('"'"'${MONGO_INITDB_ROOT_USERNAME}'"'"')" 2>/dev/null | grep -q "user"; then\n\
        log "INFO" "Creating admin user..."\n\
        mongosh admin --quiet --eval "db.createUser({user:'"'"'${MONGO_INITDB_ROOT_USERNAME}'"'"',pwd:'"'"'${MONGO_INITDB_ROOT_PASSWORD}'"'"',roles:[{role:'"'"'root'"'"',db:'"'"'admin'"'"'}]})"\n\
        log "SUCCESS" "Admin user created"\n\
    fi\n\
fi\n\
\n\
if [ "${RESTORE_FROM_GITHUB:-false}" = "true" ]; then\n\
    if ! mongosh --eval "db.adminCommand({listDatabases:1})" 2>/dev/null | grep -q "opengin"; then\n\
        log "INFO" "opengin database not found, starting GitHub restore..."\n\
        restore_from_github || log "WARNING" "GitHub restore failed, continuing with empty database"\n\
    else\n\
        log "INFO" "opengin database already exists, skipping restore"\n\
    fi\n\
fi\n\
\n\
log "INFO" "Restarting MongoDB in foreground..."\n\
kill $MONGODB_PID 2>/dev/null || true\n\
wait $MONGODB_PID 2>/dev/null || true\n\
sleep 3\n\
\n\
log "INFO" "Starting MongoDB in foreground mode..."\n\
exec mongod --dbpath /data/db --logpath /var/log/mongodb/mongod.log --bind_ip_all\n\
' > /custom-entrypoint.sh && chmod +x /custom-entrypoint.sh

USER 10014

VOLUME ["/data/db", "/data/configdb", "/data/backup"]

EXPOSE 27017

# ✅ FIX: Use mongosh instead of mongo (mongo shell removed in MongoDB 6+)
HEALTHCHECK --interval=10s --timeout=10s --start-period=60s --retries=10 \
    CMD mongosh --eval "db.adminCommand('ping')" || exit 1

CMD ["/custom-entrypoint.sh"]