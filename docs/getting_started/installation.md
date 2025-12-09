# Installation & Setup

## Quick Development Setup

### Prerequisites
- Docker and Docker Compose
- Go 1.19+ (for CORE service)
- Ballerina (for APIs)

### Start the System

1. **Start databases**
   ```bash
   docker-compose up -d mongodb neo4j postgres
   ```

2. **Start CORE service**
   ```bash
   cd opengin/core-api && ./core-service
   ```

3. **Start APIs**
   - **Ingestion API**: `http://localhost:8080`
   - **Read API**: `http://localhost:8081`

   *(Ensure you run the Ballerina services as per their specific instructions in `opengin/ingestion-api` and `opengin/read-api`)*

### Test the System

**Run E2E tests**
```bash
cd opengin/tests/e2e && ./run_e2e.sh
```

**Run performance tests**
```bash
cd perf && python performance_test.py
```
