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
   docker-compose up -d core
   ```
3. **Start APIs**
   ```bash
   docker-compose up -d ingestion read
   ```

### Test the System

**Run E2E tests**
```bash
cd opengin/tests/e2e && ./run_e2e.sh
```
