# MongoDB Server Dockerfile
FROM mongo:4.4

# Set environment variables
ENV MONGO_INITDB_ROOT_USERNAME=admin
ENV MONGO_INITDB_ROOT_PASSWORD=admin123
ENV MONGO_INITDB_DATABASE=opengin

# Create additional directories for backups
RUN mkdir -p /data/backup && \
    chown -R mongodb:mongodb /data/backup && \
    chmod -R 755 /data/backup

# Define volumes for data persistence
VOLUME ["/data/db", "/data/configdb", "/data/backup"]

# Expose ports
EXPOSE 27017

# Health check
HEALTHCHECK --interval=10s --timeout=10s --start-period=60s --retries=10 \
    CMD mongo --eval "db.adminCommand('ping')" || exit 1

# Default command (inherited from base image)
