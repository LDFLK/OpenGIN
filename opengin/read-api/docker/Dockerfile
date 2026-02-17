# Copyright 2025 Lanka Data Foundation
# SPDX-License-Identifier: Apache-2.0

## TODO: Please complete the Dockerfile.read file
## WORK IN PROGRESS


FROM ballerina/ballerina:2201.11.0

# Install system packages
USER root
RUN apk update && apk add --no-cache \
    curl gnupg wget net-tools nano \
    unzip netcat-openbsd iputils bind-tools \
    dos2unix \
    && rm -rf /var/cache/apk/*

WORKDIR /app

# Copy the entire repository (build context is repo root: '.')
# The build context is set in docker-compose.yml as '.' (repo root)
# This allows access to all services in the monorepo if needed
COPY . .

# Print Ballerina version with clear formatting
RUN echo "=========================================" \
    && echo "Checking Ballerina version:" \
    && bal version \
    && echo "========================================="

# Set proper permissions for the entire project
RUN mkdir -p /app/opengin/read-api/target \
    && chmod -R 755 /app

# Set environment variables for configurable strings
ENV BAL_CONFIG_VAR_CORESERVICEURL=${BAL_CONFIG_VAR_CORESERVICEURL:-http://core:50051}
ENV BAL_CONFIG_VAR_READSERVICEHOST=${BAL_CONFIG_VAR_READSERVICEHOST:-0.0.0.0}
ENV BAL_CONFIG_VAR_READSERVICEPORT=${BAL_CONFIG_VAR_READSERVICEPORT:-8081}
ENV JAVA_OPTS="-Xmx512m"

# Build the Ballerina service
# Note: We navigate to opengin/read-api because the build context is the repo root
RUN cd opengin/read-api && \
    bal build

EXPOSE 8081

# Create a startup script that checks connectivity before running the service
RUN printf '#!/bin/sh\n\
    echo "Waiting for CORE service to be ready..."\n\
    # Extract host and port from CORE service URL for connectivity check\n\
    CORE_HOST=$(echo ${BAL_CONFIG_VAR_CORESERVICEURL} | sed "s|http://||" | cut -d: -f1)\n\
    CORE_PORT=$(echo ${BAL_CONFIG_VAR_CORESERVICEURL} | cut -d: -f3 | cut -d/ -f1)\n\
    until nc -z -w 5 ${CORE_HOST} ${CORE_PORT}; do\n\
    echo "Waiting for CORE service at ${CORE_HOST}:${CORE_PORT}..."\n\
    sleep 2\n\
    done\n\
    echo "CORE service is ready!"\n\
    \n\
    echo "Run ballerina test..."\n\
    bal test opengin/read-api\n\
    \n\
    echo "Starting Read service..."\n\
    exec bal run opengin/read-api\n' > /app/start.sh && \
    chmod +x /app/start.sh && \
    dos2unix /app/start.sh

CMD ["/app/start.sh"]

#  CMD ["tail", "-f", "/dev/null"]
