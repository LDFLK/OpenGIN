# Copyright 2025 Lanka Data Foundation
# SPDX-License-Identifier: Apache-2.0

#!/bin/bash

# Exit on any error
set -e

# Function to cleanup on exit
cleanup() {
    echo "Cleaning up..."
    docker-compose down
}

# Set up trap to ensure cleanup happens
trap cleanup EXIT

# Build and run the services
echo "Starting services..."
docker-compose up --build -d

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 10

# Clean up databases
echo "Cleaning up databases..."
docker-compose exec mongo mongosh --eval 'db.getSiblingDB("testdb").dropDatabase()'
docker-compose exec neo4j cypher-shell -u neo4j -p test123456 'MATCH (n) DETACH DELETE n'

# Run the tests
echo "Running tests..."
if ! docker-compose run --rm test go test -v ./...; then
    echo "Tests failed!"
    exit 1
fi

echo "Tests passed!"