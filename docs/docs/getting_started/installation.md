# Installation & Setup

## Quick Development Setup

### Prerequisites
- Docker and Docker Compose
- Go 1.19+ (for CORE service)
- Ballerina 2201.13.4 (for APIs)

### Start the System (Using Docker)

1. **Start databases**
   ```bash
   docker-compose up -d mongodb neo4j postgres
   ```

2. **Start CORE service**
   ```bash
   docker-compose up -d core
   ```
3. **Start APIs**
   ```bash
   docker-compose up -d ingestion read
   ```

### Start the System (Locally, without Docker)

1. **Start databases** — choose one:

   **Option A: Local databases (Docker)**
   ```bash
   docker-compose up -d mongodb neo4j postgres
   ```

   **Option B: Cloud databases** (e.g. [Neo4j Aura](https://neo4j.com/cloud/aura/), [MongoDB Atlas](https://www.mongodb.com/atlas), a managed PostgreSQL instance)

   Skip the Docker step above and instead point core-api at your cloud instances by setting the connection variables in `opengin/core-api/.env` (or `.env.ps1` on Windows) before starting CORE in step 3:
   ```bash
   export NEO4J_URI=neo4j+s://<your-instance-id>.databases.neo4j.io
   export NEO4J_USER=neo4j
   export NEO4J_PASSWORD=<your-password>
   export MONGO_URI=<your-mongodb-atlas-connection-string>
   export POSTGRES_HOST=<your-postgres-host>
   export POSTGRES_PORT=5432
   export POSTGRES_USER=<your-postgres-user>
   export POSTGRES_PASSWORD=<your-postgres-password>
   export POSTGRES_DB=<your-postgres-db>
   ```
   (Windows PowerShell: same variables using `$env:NEO4J_URI="..."` syntax, in `.env.ps1`.)

   Note Neo4j Aura uses the `neo4j+s://` scheme (encrypted), not `bolt://` — using `bolt://` against an Aura instance will fail to connect.

2. **Start CORE service**

   For LINUX & macOS
   ```bash
   cd opengin/core-api
   cp env.template .env
   source .env
   go build -o core-service cmd/server/service.go cmd/server/utils.go
   ./core-service
   ```

   For Windows (PowerShell)
   ```powershell
   cd opengin/core-api
   Copy-Item env.template.ps1 .env.ps1
   . .\.env.ps1
   go build -o core-service.exe cmd/server/service.go cmd/server/utils.go
   .\core-service.exe
   ```

   See [core-api/README.md](../../../opengin/core-api/README.md) for details.

3. **Start ingestion-api** (make sure CORE service is running)

   For LINUX & macOS
   ```bash
   cd opengin/ingestion-api
   cp env.template .env
   source .env
   bal run
   ```

   For Windows (PowerShell)
   ```powershell
   cd opengin/ingestion-api
   Copy-Item env.template.ps1 .env.ps1
   . .\.env.ps1
   bal run
   ```

   See [ingestion-api/README.md](../../../opengin/ingestion-api/README.md) for details.

4. **Start read-api** (make sure CORE service is running)

   For LINUX & macOS
   ```bash
   cd opengin/read-api
   cp env.template .env
   source .env
   bal run
   ```

   For Windows (PowerShell)
   ```powershell
   cd opengin/read-api
   Copy-Item env.template.ps1 .env.ps1
   . .\.env.ps1
   bal run
   ```

   See [read-api/README.md](../../../opengin/read-api/README.md) for details.

### Test the System

**Run E2E tests**
```bash
cd opengin/tests/e2e && ./run_e2e.sh
```
