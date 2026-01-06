# Choreo 

## Local Development and Testing

### Prerequisites

- Docker and Docker Compose installed
- Git repository cloned
- Ports available:
  - 50051 (CORE service)
  - 8080 (Ingestion service)
  - 27017 (MongoDB choreo)
  - 7474/7687 (Neo4j choreo)
  - 5432 (PostgreSQL choreo)

### Docker Compose Setup (Recommended)

The easiest way to run the choreo services locally is using the dedicated docker-compose file that includes all required databases.

#### Quick Start

```bash
# Clone the repository and navigate to the root directory
cd /path/to/opengin

# Start all choreo services (includes databases)
docker-compose -f docker-compose-choreo.yml up --build

# Or start specific services
docker-compose -f docker-compose-choreo.yml up --build core-choreo ingestion-choreo

# Run in background
docker-compose -f docker-compose-choreo.yml up --build -d
```

#### Available Docker Compose Commands

```bash
# Build services with no cache (recommended after code changes)
docker-compose -f docker-compose-choreo.yml build --no-cache

# Start all services
docker-compose -f docker-compose-choreo.yml up

# Start with build
docker-compose -f docker-compose-choreo.yml up --build

# Start specific services
docker-compose -f docker-compose-choreo.yml up mongodb-choreo neo4j-choreo postgres-choreo
docker-compose -f docker-compose-choreo.yml up core-choreo ingestion-choreo

# View logs
docker-compose -f docker-compose-choreo.yml logs core-choreo
docker-compose -f docker-compose-choreo.yml logs ingestion-choreo

# Stop all services
docker-compose -f docker-compose-choreo.yml down

# Stop and remove volumes (clean slate)
docker-compose -f docker-compose-choreo.yml down -v

# Debug: Run interactive shell in a service
docker-compose -f docker-compose-choreo.yml run --entrypoint="" core-choreo sh
docker-compose -f docker-compose-choreo.yml run --entrypoint="" ingestion-choreo sh

# Run end-to-end tests
docker-compose -f docker-compose-choreo.yml up e2e-choreo
```

#### Service Architecture

The docker-compose setup includes:
- **mongodb-choreo**: MongoDB instance (port 27018)
- **neo4j-choreo**: Neo4j graph database (ports 7475, 7688)
- **postgres-choreo**: PostgreSQL database (port 5433)
- **core-choreo**: CORE API service (port 50051)
- **ingestion-choreo**: Ingestion API service (port 8080)
- **e2e-choreo**: End-to-end test runner

#### Database Access

- **MongoDB**: `mongodb://admin:admin123@localhost:27018/admin`
- **Neo4j**: `http://localhost:7475` (user: neo4j, password: neo4j123)
- **PostgreSQL**: `postgresql://postgres:postgres@localhost:5433/ldf_choreo_opengin`

### Manual Environment Setup (Alternative)

If you prefer to run services manually or against external databases, set up your environment variables:

```bash
# MongoDB Configuration
export MONGO_URI="mongodb+srv://username:password@your-cluster.mongodb.net/?retryWrites=true&w=majority"
export MONGO_DB_NAME="your-mongo-db-name"
export MONGO_COLLECTION="your-mongo-collection-name"
export MONGO_ADMIN_USER="your-mongo-admin-username"
export MONGO_ADMIN_PASSWORD="your-mongo-admin-password"

# Neo4j Configuration
export NEO4J_URI="neo4j+s://your-neo4j-instance.databases.neo4j.io"
export NEO4J_USER="your-neo4j-username"
export NEO4J_PASSWORD="your-neo4j-password"

# PostgreSQL Configuration
export POSTGRES_HOST="your-postgres-host"
export POSTGRES_PORT=5432
export POSTGRES_USER="your-postgres-username"
export POSTGRES_PASSWORD="your-postgres-password"
export POSTGRES_DB="your-postgres-database"
export POSTGRES_SSL_MODE="require"

# Service Configuration
export CORE_SERVICE_HOST="0.0.0.0"
export CORE_SERVICE_PORT="50051"
export INGESTION_SERVICE_HOST="0.0.0.0"
export INGESTION_SERVICE_PORT="8080"
export READ_SERVICE_HOST="0.0.0.0"
export READ_SERVICE_PORT="8081"
```

### Running Services Locally

1. Start the CORE Service:
```bash
# Build the CORE service image
# For ARM64 (Apple Silicon):
docker build --platform linux/arm64 -t ldf-choreo-core-service -f opengin/core-api/docker/Dockerfile.choreo .
# For AMD64:
docker build --platform linux/amd64 -t ldf-choreo-core-service -f opengin/core-api/docker/Dockerfile.choreo .

# Run the CORE service using environment variables
docker run -d \
  --name ldf-choreo-core-service \
  -p 50051:50051 \
  -e NEO4J_URI="$NEO4J_URI" \
  -e NEO4J_USER="$NEO4J_USER" \
  -e NEO4J_PASSWORD="$NEO4J_PASSWORD" \
  -e MONGO_URI="$MONGO_URI" \
  -e MONGO_DB_NAME="$MONGO_DB_NAME" \
  -e MONGO_COLLECTION="$MONGO_COLLECTION" \
  -e MONGO_ADMIN_USER="$MONGO_ADMIN_USER" \
  -e MONGO_ADMIN_PASSWORD="$MONGO_ADMIN_PASSWORD" \
  -e POSTGRES_HOST="$POSTGRES_HOST" \
  -e POSTGRES_PORT="$POSTGRES_PORT" \
  -e POSTGRES_USER="$POSTGRES_USER" \
  -e POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
  -e POSTGRES_DB="$POSTGRES_DB" \
  -e POSTGRES_SSL_MODE="$POSTGRES_SSL_MODE" \
  -e POSTGRES_TEST_DB_URI="$POSTGRES_TEST_DB_URI" \
  -e CORE_SERVICE_HOST="$CORE_SERVICE_HOST" \
  -e CORE_SERVICE_PORT="$CORE_SERVICE_PORT" \
  ldf-choreo-core-service
```

2. Start the Ingestion Service:

```bash
# Build the ingestion service image
docker build -t ldf-choreo-ingestion-service -f opengin/ingestion-api/docker/Dockerfile.choreo .

# Run the ingestion service using environment variables
docker run -d -p 8080:8080 \
  --name ldf-choreo-ingestion-service \
  -e CORE_SERVICE_URL="http://host.docker.internal:$CORE_SERVICE_PORT" \
  -e INGESTION_SERVICE_HOST="$INGESTION_SERVICE_HOST" \
  -e INGESTION_SERVICE_PORT="$INGESTION_SERVICE_PORT" \
  ldf-choreo-ingestion-service
```

## Testing Locally 

### Running Database services

```bash
docker-compose -f docker-compose-choreo.yml up -d mongodb-choreo neo4j-choreo postgres-choreo
```

### Using docker-compose.override.yml for Fresh Testing

The `docker-compose.override.yml` file provides a way to override default Docker Compose settings for testing and CI environments. This is particularly useful for ensuring fresh, clean databases for each test run.

#### What docker-compose.override.yml Does

The override file modifies the base `docker-compose.yml` to:
- **Use tmpfs volumes** for databases (in-memory, no persistence)
- **Ensure fresh databases** for each test run
- **Prevent data contamination** between test runs
- **Optimize CI/CD pipelines** with clean environments

#### How to Use docker-compose.override.yml

```bash
# Start services with override (recommended for testing)
docker-compose -f docker-compose.yml -f docker-compose.override.yml up -d

# Start specific services with override
docker-compose -f docker-compose.yml -f docker-compose.override.yml up -d mongodb neo4j postgres core

# Build with override
docker-compose -f docker-compose.yml -f docker-compose.override.yml build --no-cache

# Stop services with override
docker-compose -f docker-compose.yml -f docker-compose.override.yml down
```

#### Override File Structure

```yaml
# docker-compose.override.yml
version: '3.8'

services:
  mongodb:
    volumes:
      - type: tmpfs
        target: /data/db
        tmpfs:
          size: 1000000000  # 1GB in-memory storage
  
  neo4j:
    volumes:
      - type: tmpfs
        target: /data
        tmpfs:
          size: 1000000000  # 1GB in-memory storage
  
  postgres:
    volumes:
      - type: tmpfs
        target: /var/lib/postgresql/data
        tmpfs:
          size: 1000000000  # 1GB in-memory storage
```

#### Benefits of Using Override

1. **🧹 Fresh Start Every Time**
   - No leftover data from previous test runs
   - Consistent test environment
   - Predictable test results

2. **⚡ Faster CI/CD**
   - No need to wait for database cleanup
   - Immediate test execution
   - Reduced flaky test issues

3. **🔒 Test Isolation**
   - Each test run is completely independent
   - No data leakage between runs
   - Clean slate for debugging

4. **💾 Memory Efficiency**
   - Databases run in memory (tmpfs)
   - Faster database operations
   - No disk I/O overhead

#### When to Use Override vs Standard

| Use Case | Command | Reason |
|----------|---------|---------|
| **Development** | `docker-compose up` | Persistent data, faster development |
| **Testing** | `docker-compose -f docker-compose.yml -f docker-compose.override.yml up` | Fresh databases, test isolation |
| **CI/CD** | `docker-compose -f docker-compose.yml -f docker-compose.override.yml up` | Clean environment, reliable builds |
| **Production** | `docker-compose up` | Data persistence, production stability |

#### Troubleshooting Override Issues

```bash
# Check if override is being applied
docker-compose -f docker-compose.yml -f docker-compose.override.yml config

# Verify tmpfs volumes are created
docker volume ls | grep tmpfs

# Check database freshness
docker exec mongodb mongosh --eval "db.getMongo().getDBNames()"
docker exec neo4j cypher-shell -u neo4j -p neo4j123 "MATCH (n) RETURN count(n)"
docker exec postgres psql -U postgres -d opengin -c "SELECT COUNT(*) FROM information_schema.tables;"
```

#### Integration with GitHub Actions

The override file is automatically used in CI workflows:

```yaml
# .github/workflows/ingestion-api-test.yml
- name: Start services with fresh databases
  run: |
    docker-compose -f docker-compose.yml -f docker-compose.override.yml up -d mongodb neo4j postgres core
```

This ensures that every CI run starts with completely fresh databases, preventing test failures due to leftover data from previous runs.

## 📖 Additional Documentation

For detailed information about database cleanup strategies and troubleshooting, see:
- [Database Cleanup Best Practices](./DATABASE_CLEANUP.md) - Comprehensive guide to database cleanup issues and solutions



## References

1. https://wso2.com/choreo/docs/develop-components/develop-components-with-git/
2. https://wso2.com/choreo/docs/devops-and-ci-cd/manage-configurations-and-secrets/
3. https://wso2.com/choreo/docs/choreo-concepts/ci-cd/
4. [Test Runner](https://wso2.com/choreo/docs/testing/test-components-with-test-runner/)
5. [Choreo Examples](https://github.com/wso2/choreo-samples)
6. [Expose TCP Server via a Service](https://wso2.com/choreo/docs/develop-components/develop-services/expose-a-tcp-server-via-a-service/#step-2-build-and-deploy)