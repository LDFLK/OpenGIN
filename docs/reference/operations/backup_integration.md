# Backup Integration Guide

This guide covers both local backup management and GitHub-based backup restoration using our `init.sh` script. There are two main workflows depending on your needs.

## Two Main Workflows

### 1. **Local Backup Workflow** 
- Create backups locally using `init.sh`
- Store backups in local directories
- Restore from local backup files
- **Use case**: Development, testing, local data management

### 2. **GitHub Release Workflow**
- Download and restore from [LDFLK/data-backups](https://github.com/LDFLK/data-backups) repository
- Use pre-built backup releases
- **Use case**: Production deployments, team collaboration, version management

---

## Local Backup Workflow

### Create Local Backups

```bash
# Create individual database backups
./deployment/development/init.sh backup_mongodb
./deployment/development/init.sh backup_postgres
./deployment/development/init.sh backup_neo4j

# List available local backups
./deployment/development/init.sh list_mongodb_backups
./deployment/development/init.sh list_postgres_backups
./deployment/development/init.sh list_neo4j_backups
```

### Restore from Local Backups

```bash
# Restore from local backup directories
./deployment/development/init.sh restore_mongodb
./deployment/development/init.sh restore_postgres
./deployment/development/init.sh restore_neo4j
```

### Local Backup Structure

```
backups/
├── mongodb/
│   └── opengin.tar.gz
├── postgres/
│   └── opengin.tar.gz
└── neo4j/
    └── neo4j.dump
```

### Environment Configuration

Set backup directories in `configs/backup.env`:

```bash
# MongoDB backup directory
MONGODB_BACKUP_DIR=/path/to/mongodb/backups

# PostgreSQL backup directory  
POSTGRES_BACKUP_DIR=/path/to/postgres/backups

# Neo4j backup directory
NEO4J_BACKUP_DIR=/path/to/neo4j/backups
```

---

## GitHub Release Workflow

### Restore from GitHub Releases

```bash
# Restore latest version from GitHub
./deployment/development/init.sh restore_from_github

# Restore specific version
./deployment/development/init.sh restore_from_github 0.0.1

# List available versions
./deployment/development/init.sh list_github_versions

# Get latest version info
./deployment/development/init.sh get_latest_github_version
```

### Docker Compose Integration

```bash
# Start all services (backup-manager auto-restores from GitHub)
docker-compose up -d

# Check backup-manager logs
docker logs backup-manager

# Run commands in backup-manager container
docker exec backup-manager /init.sh list_github_versions
docker exec backup-manager /init.sh restore_from_github 0.0.1
```

### GitHub Repository Structure

The system expects this structure in [LDFLK/data-backups](https://github.com/LDFLK/data-backups):

```
data-backups-0.0.1/
└── opengin
    └── development
        ├── mongodb
        │   └── mongodb.tar.gz
        ├── postgres
        │   └── postgres.tar.gz
        └── neo4j
            └── neo4j.dump
```

### GitHub Archive URLs

The system uses GitHub's built-in archive feature:

- **Version 0.0.1**: https://github.com/LDFLK/data-backups/archive/refs/tags/0.0.1.zip
- **Version 0.0.2**: https://github.com/LDFLK/data-backups/archive/refs/tags/0.0.2.zip
- **Any version**: https://github.com/LDFLK/data-backups/archive/refs/tags/{version}.zip

---

## Complete Workflow Examples

### Scenario 1: Development Setup

```bash
# 1. Start your services
docker-compose up -d

# 2. Create local backups
./deployment/development/init.sh backup_mongodb
./deployment/development/init.sh backup_postgres
./deployment/development/init.sh backup_neo4j

# 3. Test your application
# ... do development work ...

# 4. Restore from local backups if needed
./deployment/development/init.sh restore_mongodb
./deployment/development/init.sh restore_postgres
./deployment/development/init.sh restore_neo4j
```

### Scenario 2: Production Deployment

```bash
# 1. Deploy with GitHub backup restoration
docker-compose up -d

# 2. The backup-manager automatically restores from GitHub
# Check logs to confirm
docker logs backup-manager

# 3. Your application is ready with production data
```

### Scenario 3: Team Collaboration

```bash
# 1. Create backups locally
./deployment/development/init.sh backup_mongodb
./deployment/development/init.sh backup_postgres
./deployment/development/init.sh backup_neo4j

# 2. Upload to GitHub repository (manual process)
# - Create release in LDFLK/data-backups
# - Upload backup files to correct directory structure

# 3. Team members restore from GitHub
./deployment/development/init.sh restore_from_github 0.0.1
```

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ENVIRONMENT` | `development` | Environment (dev/staging/prod) |
| `MONGODB_BACKUP_DIR` | `./backups/mongodb` | Local MongoDB backup directory |
| `POSTGRES_BACKUP_DIR` | `./backups/postgres` | Local PostgreSQL backup directory |
| `NEO4J_BACKUP_DIR` | `./backups/neo4j` | Local Neo4j backup directory |

### Database Credentials

Configure in `configs/backup.env`:

```bash
# MongoDB
MONGODB_USERNAME=admin
MONGODB_PASSWORD=admin123
MONGODB_DATABASE=opengin

# PostgreSQL
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DATABASE=opengin
```

---

## Troubleshooting

### Local Backup Issues

```bash
# Check if backup directories exist
ls -la ./backups/mongodb/
ls -la ./backups/postgres/
ls -la ./backups/neo4j/

# Check database connectivity
docker exec mongodb mongo --eval "db.adminCommand('ping')"
docker exec postgres pg_isready -U postgres
docker exec neo4j cypher-shell -u neo4j -p neo4j123 "RETURN 1"
```

### GitHub Backup Issues

```bash
# Test GitHub archive download
wget -O test.zip "https://github.com/LDFLK/data-backups/archive/refs/tags/0.0.1.zip"
unzip -l test.zip

# Check archive structure
unzip -l test.zip | grep opengin

# Run with debug output
bash -x ./deployment/development/init.sh restore_from_github 0.0.1
```

### Container State Issues

```bash
# Check container status
docker-compose ps

# Check Neo4j container state (smart handling)
docker-compose ps neo4j

# View logs
docker logs mongodb
docker logs postgres
docker logs neo4j
docker logs backup-manager
```

---

## Key Features

### Smart Container Handling
- **Neo4j functions** intelligently detect container state
- Only stop/start containers when necessary
- Faster operations when containers are already in desired state

### No Authentication Required
- **GitHub workflow** uses public archive URLs
- No API tokens or rate limits
- Direct file downloads with `wget`

### Flexible Workflows
- **Local backups** for development and testing
- **GitHub releases** for production and team collaboration
- Both workflows use the same `init.sh` commands

### Environment Support
- Development, staging, production environments
- Configurable backup directories
- Environment-specific GitHub releases

---

## Command Reference

### Local Backup Commands
```bash
# Backup
./deployment/development/init.sh backup_mongodb
./deployment/development/init.sh backup_postgres
./deployment/development/init.sh backup_neo4j

# Restore
./deployment/development/init.sh restore_mongodb
./deployment/development/init.sh restore_postgres
./deployment/development/init.sh restore_neo4j

# List
./deployment/development/init.sh list_mongodb_backups
./deployment/development/init.sh list_postgres_backups
./deployment/development/init.sh list_neo4j_backups
```

### GitHub Commands
```bash
# Restore
./deployment/development/init.sh restore_from_github
./deployment/development/init.sh restore_from_github 0.0.1

# Version management
./deployment/development/init.sh list_github_versions
./deployment/development/init.sh get_latest_github_version
```

### Service Management
```bash
# Neo4j service
./deployment/development/init.sh setup_neo4j
./deployment/development/init.sh run_neo4j

# General
./deployment/development/init.sh setup
./deployment/development/init.sh help
```

This covers both local backup management and GitHub-based restoration workflows! 🎯
