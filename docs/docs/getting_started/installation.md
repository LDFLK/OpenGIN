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

### Windows notes

The same stack can be started from PowerShell when Docker Desktop is running.
Use the current Compose subcommand spelling shown below:

```powershell
docker compose up -d mongodb neo4j postgres
docker compose up -d core
docker compose up -d ingestion read
```

Before starting the services, confirm that Docker Desktop's Linux engine is
running. `docker compose config --quiet` can validate the Compose file without
starting containers. The repository's `run_e2e.sh` script is POSIX shell
syntax, so run it from WSL or Git Bash rather than native PowerShell:

```bash
cd opengin/tests/e2e
bash ./run_e2e.sh
```

Go 1.19+ and Ballerina are also required by the CORE and API services. They
must be installed separately on Windows and available on `PATH` before using
the non-container development commands.

### Test the System

**Run E2E tests**
```bash
cd opengin/tests/e2e && ./run_e2e.sh
```
