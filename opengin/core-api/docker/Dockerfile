# Copyright 2025 Lanka Data Foundation
# SPDX-License-Identifier: Apache-2.0

# Dockerfile
#
# Purpose:
# This Dockerfile builds and runs the CORE API service, which provides a gRPC interface
# for creating, reading, updating, and deleting entities in the system. The service
# connects to both Neo4j (for graph relationships), MongoDB (for metadata storage),
# and PostgreSQL (for tabular data).
#
# Usage:
# 1. Build the image:
#    docker build -t core-service -f opengin/core-api/docker/Dockerfile .
#
# 2. Run the container:
#    docker run -d \
#      --name core-service \
#      --network core-network \
#      -p 50051:50051 \
#      -e NEO4J_URI=bolt://neo4j-local:7687 \
#      -e NEO4J_USER=${NEO4J_USER} \
#      -e NEO4J_PASSWORD=${NEO4J_PASSWORD} \
#      -e MONGO_URI=${MONGO_URI} \
#      core-service
#
# Environment Variables:
# - NEO4J_URI: Connection URI for Neo4j database (default: bolt://neo4j:7687)
# - NEO4J_USER: Username for Neo4j authentication (default: neo4j)
# - NEO4J_PASSWORD: Password for Neo4j authentication (default: neo4j123)
# - MONGO_URI: Connection URI for MongoDB (default: mongodb://mongo:27017)
# - MONGO_DB_NAME: MongoDB database name (default: testdb)
# - MONGO_COLLECTION: MongoDB collection name (default: metadata)
# - MONGO_ADMIN_USER: MongoDB admin username (default: admin)
# - MONGO_ADMIN_PASSWORD: MongoDB admin password (default: test123456)
# - CORE_SERVICE_HOST: Host address to bind the service (default: 0.0.0.0)
# - CORE_SERVICE_PORT: Port to expose the gRPC service (default: 50051)
#
# Note: This service should be run on the same Docker network as Neo4j and MongoDB
# services for proper connectivity. Use the 'core-network' created with:
# docker network create core-network

# Build stage for CORE service
FROM golang:1.23 AS builder

# Set working directory
WORKDIR /app

# Copy the entire repository (build context is repo root: '.')
# The build context is set in docker-compose.yml as '.' (repo root)
# This allows access to all services in the monorepo if needed
COPY . .

# Download Go dependencies
# Note: We navigate to opengin/core-api because the build context is the repo root
RUN cd opengin/core-api && \
    go mod download

# Build the application
# Note: We navigate to opengin/core-api because the build context is the repo root
RUN cd opengin/core-api && \
    go build -o core-service cmd/server/service.go cmd/server/utils.go

## Create a new user with UID 10014
# RUN addgroup -g 10014 choreo && \
#     adduser  --disabled-password  --no-create-home --uid 10014 --ingroup choreo choreouser

# Final stage
FROM golang:1.24

# Install runtime dependencies and qemu for emulation
RUN apt-get update && \
    # Upgrade OpenSSL packages to fix CVE-2025-15467 (CRITICAL)
    apt-get upgrade -y libssl3t64 openssl openssl-provider-legacy && \
    apt-get install -y ca-certificates gnupg curl wget unzip default-jdk postgresql-client netcat-openbsd && \
    # Add MongoDB repository
    curl -fsSL https://pgp.mongodb.com/server-6.0.asc | \
    gpg -o /usr/share/keyrings/mongodb-server-6.0.gpg --dearmor && \
    echo "deb [ arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-6.0.gpg ] https://repo.mongodb.org/apt/debian bullseye/mongodb-org/6.0 main" | \
    tee /etc/apt/sources.list.d/mongodb-org-6.0.list && \
    apt-get update && \
    apt-get install -y mongodb-mongosh && \
    # Download and install Cypher Shell
    wget https://dist.neo4j.org/cypher-shell/cypher-shell-5.26.5.zip && \
    unzip cypher-shell-5.26.5.zip && \
    mkdir -p /usr/local/lib/cypher-shell && \
    mv cypher-shell-5.26.5/lib/* /usr/local/lib/cypher-shell/ && \
    # Create a wrapper script with proper classpath
    echo '#!/bin/sh\n\
    CLASSPATH=$(find /usr/local/lib/cypher-shell -name "*.jar" | tr "\\n" ":")\n\
    exec java -cp "$CLASSPATH" org.neo4j.shell.Main "$@"\n\
    ' > /usr/local/bin/cypher-shell && \
    chmod +x /usr/local/bin/cypher-shell && \
    rm -rf cypher-shell-5.26.5.zip cypher-shell-5.26.5 && \
    rm -rf /var/lib/apt/lists/*

# Set Java environment variables
ENV JAVA_HOME=/usr/lib/jvm/default-java
ENV PATH=$JAVA_HOME/bin:$PATH

# Copy the built binary from builder stage
COPY --from=builder /app/opengin/core-api/core-service /usr/local/bin/
COPY --from=builder /app/opengin/core-api /app/opengin/core-api
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy the startup script
COPY opengin/core-api/docker/core_startup.sh /app/start.sh
RUN chmod +x /app/start.sh

# Set environment variables with defaults for Neo4j connection
ENV NEO4J_URI=${NEO4J_URI:-bolt://neo4j:7687}
ENV NEO4J_USER=${NEO4J_USER:-neo4j}
ENV NEO4J_PASSWORD=${NEO4J_PASSWORD:-neo4j123}

# Set MongoDB environment variables with defaults
ENV MONGO_URI=${MONGO_URI:-mongodb://admin:admin123@mongodb:27017/admin?authSource=admin}
ENV MONGO_DB_NAME=${MONGO_DB_NAME:-testdb}
ENV MONGO_COLLECTION=${MONGO_COLLECTION:-metadata}
ENV MONGO_ADMIN_USER=${MONGO_ADMIN_USER:-admin}
ENV MONGO_ADMIN_PASSWORD=${MONGO_ADMIN_PASSWORD:-admin123}

# Set PostgreSQL environment variables with defaults
ENV POSTGRES_HOST=${POSTGRES_HOST:-postgres}
ENV POSTGRES_PORT=${POSTGRES_PORT:-5432}
ENV POSTGRES_USER=${POSTGRES_USER:-postgres}
ENV POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-postgres}
ENV POSTGRES_DB=${POSTGRES_DB:-opengin}

# Set CORE service environment variables
ENV CORE_SERVICE_HOST=${CORE_SERVICE_HOST:-0.0.0.0}
ENV CORE_SERVICE_PORT=${CORE_SERVICE_PORT:-50051}

# Add log configuration
ENV LOG_LEVEL=debug
ENV LOG_FORMAT=text

# USER 10014

# Expose ports
EXPOSE 50051

# The startup script is now copied from startup.sh

# Display environment variables and start the service
CMD ["/app/start.sh"]