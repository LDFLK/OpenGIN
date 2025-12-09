# Dockerfile
#
# Purpose:
# This Dockerfile creates a lightweight container specifically for database cleanup operations.
# It contains only the necessary database clients and cleanup scripts, without any application code.
#
# Usage:
# 1. Build the image:
#    docker build -t cleanup-service -f deployment/development/docker/cleanup/Dockerfile .
#
# 2. Run cleanup:
#    docker run --rm cleanup-service /app/cleanup.sh [pre|post]

# Use minimal alpine image
FROM alpine:latest

# Install runtime dependencies for database operations
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    postgresql-client \
    mongodb-tools \
    cypher-shell \
    bash \
    curl \
    && rm -rf /var/cache/apk/*

# Create app directory
WORKDIR /app

# Copy the cleanup script
COPY deployment/development/docker/cleanup/cleanup.sh /app/cleanup.sh
RUN chmod +x /app/cleanup.sh

# Set environment variables with defaults
ENV POSTGRES_HOST=postgres
ENV POSTGRES_PORT=5432
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=postgres
ENV POSTGRES_DB=opengin

ENV MONGO_URI=mongodb://admin:admin123@mongodb:27017/admin?authSource=admin
ENV MONGO_DB_NAME=testdb

ENV NEO4J_URI=bolt://neo4j:7687
ENV NEO4J_USER=neo4j
ENV NEO4J_PASSWORD=neo4j123

# Default command
CMD ["/app/cleanup.sh"]
