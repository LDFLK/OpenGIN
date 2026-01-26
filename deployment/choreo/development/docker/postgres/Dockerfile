# Copyright 2025 Lanka Data Foundation
# SPDX-License-Identifier: Apache-2.0

# PostgreSQL Server Dockerfile
FROM postgres:17

# Install dependencies for downloading and extracting backups
RUN apt-get update && apt-get install -y \
    wget \
    unzip \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Set up Choreo user (required for Choreo platform)
RUN groupadd -g 10014 choreo && \
    useradd -u 10014 -g choreo -s /bin/bash -m choreouser && \
    mkdir -p /home/choreouser/.cache && \
    chown -R choreouser:choreo /home/choreouser && \
    chmod -R 755 /home/choreouser

# Build arguments for backup source
ARG GITHUB_BACKUP_REPO=LDFLK/data-backups
ARG BACKUP_VERSION=0.0.3
ARG BACKUP_ENVIRONMENT=development

# Create custom data directory to avoid base image VOLUME definition
# postgres:17 defines VOLUME ["/var/lib/postgresql/data"]
ENV PGDATA=/var/lib/postgresql/data_baked
# Set superuser password so the entrypoint never fails on uninitialized-DB check.
# Must match the password set during build-time data ingestion below.
ENV POSTGRES_PASSWORD=postgres

# Create directory and set permissions
RUN mkdir -p "$PGDATA" && chown -R postgres:postgres "$PGDATA"

# Build-time Data Ingestion
# We stay as root to manage users/permissions but switch to postgres for DB ops
USER postgres

RUN echo "Initializing database in $PGDATA..." && \
    initdb -D "$PGDATA" && \
    \
    # Configure networking and security for remote access
    # Allow all connections with md5 password auth
    echo "host all all 0.0.0.0/0 md5" >> "$PGDATA/pg_hba.conf" && \
    # Listen on all interfaces
    echo "listen_addresses='*'" >> "$PGDATA/postgresql.conf" && \
    \
    echo "Starting PostgreSQL..." && \
    pg_ctl -D "$PGDATA" -o "-c listen_addresses='localhost'" -w start && \
    \
    # Set default password to match docker-compose config
    psql -U postgres -c "ALTER USER postgres WITH PASSWORD 'postgres';" && \
    \
    # Create temp workspace
    temp_dir=$(mktemp -d) && \
    echo "Downloading backup from $GITHUB_BACKUP_REPO @ $BACKUP_VERSION..." && \
    wget -q "https://github.com/$GITHUB_BACKUP_REPO/archive/refs/tags/$BACKUP_VERSION.zip" -O "$temp_dir/archive.zip" && \
    \
    echo "Unzipping..." && \
    unzip -q "$temp_dir/archive.zip" -d "$temp_dir" && \
    \
    # Locate the dump file
    # Structure ex: data-backups-0.0.3/opengin/development/postgres/opengin.tar.gz
    dump_path="$temp_dir/data-backups-$BACKUP_VERSION/opengin/$BACKUP_ENVIRONMENT/postgres/opengin.tar.gz" && \
    if [ -f "$dump_path" ]; then \
    echo "Found backup archive: $dump_path"; \
    tar -xzf "$dump_path" -C "$temp_dir"; \
    # Find the .sql file extracted
    sql_file=$(find "$temp_dir" -name "*.sql" | head -n 1); \
    if [ -n "$sql_file" ]; then \
    echo "Restoring from $sql_file..."; \
    # Try creating text DB just in case, ignore error if exists
    createdb -h localhost -U postgres opengin || echo "Database opengin already exists"; \
    psql -h localhost -U postgres -d opengin -f "$sql_file"; \
    else \
    echo "Error: No SQL file found in extracted archive"; \
    exit 1; \
    fi; \
    else \
    echo "Error: Backup archive not found at $dump_path"; \
    ls -R "$temp_dir"; \
    exit 1; \
    fi && \
    \
    echo "Stopping PostgreSQL..." && \
    pg_ctl -D "$PGDATA" stop && \
    \
    # Cleanup
    rm -rf "$temp_dir"

# Finalize permissions for Choreo user
USER root
RUN chown -R 10014:10014 "$PGDATA" /var/run/postgresql

# Switch to choreo user
USER 10014

# Expose ports
EXPOSE 5432

# Health check
HEALTHCHECK --interval=10s --timeout=5s --start-period=30s --retries=5 \
    CMD pg_isready -U postgres || exit 1

# Start PostgreSQL
CMD ["postgres"]
