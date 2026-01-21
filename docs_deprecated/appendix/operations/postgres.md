# PostgreSQL Backup and Restore Guide

This guide provides comprehensive instructions for backing up and restoring PostgreSQL databases in Docker containers.

## Data Backups Repository Structure

Backups are stored in a structured repository following this hierarchy:

```
data-backups/
├── README.md
└── opengin
    ├── development
    │   ├── mongodb
    │   │   └── opengin.tar.gz
    │   ├── neo4j
    │   │   └── neo4j.dump
    │   └── postgres
    │       └── opengin.tar.gz
    ├── production
    │   ├── mongodb
    │   └── neo4j
    └── staging
        ├── mongodb
        └── neo4j
```

This structure allows for:
- **Environment separation**: development, staging, production
- **Database type organization**: mongodb, neo4j, postgres
- **Version management**: 0.0.1, 0.0.2, etc.
- **Consistent naming**: All backups follow the same pattern

## Prerequisites

- Docker installed and running
- PostgreSQL container running (using docker-compose)
- Access to PostgreSQL container
- PostgreSQL credentials (replace `<your_username>` and `<your_password>` in commands below)

## Backup and Restore Commands

### Method 1: Direct Docker Commands (Recommended)

#### 1.1 Create PostgreSQL Backup

```bash
# Create backup directory
mkdir -p ./backups/postgres

# Create PostgreSQL dump
docker exec postgres pg_dump -U <your_username> -h localhost -d <your_database> -f /var/lib/postgresql/backup/opengin.sql

# Copy backup from container to host
docker cp postgres:/var/lib/postgresql/backup/opengin.sql ./backups/postgres/

# Create compressed archive
cd ./backups/postgres
tar -czf opengin.tar.gz opengin.sql
rm -rf opengin.sql

# Clean up container backup
docker exec postgres rm -rf /var/lib/postgresql/backup/opengin.sql
```

#### 1.2 Restore PostgreSQL from Backup

```bash
# Extract backup file
tar -xzf opengin.tar.gz

# Copy backup to container
docker cp opengin.sql postgres:/var/lib/postgresql/backup/

# Restore database
docker exec postgres psql -U <your_username> -d <your_database> -f /var/lib/postgresql/backup/opengin.sql

# Clean up
docker exec postgres rm -rf /var/lib/postgresql/backup/opengin.sql
rm -rf opengin.sql
```

### Method 2: Using Docker Volume Mounts

#### 2.1 Create Backup with Volume Mount

```bash
# Create backup directory
mkdir -p ./backups/postgres

# Run pg_dump with volume mount
docker run --rm \
    --network=ldf-network \
    --volume=postgres_data:/var/lib/postgresql/data \
    --volume=$(pwd)/backups/postgres:/backups \
    postgres:16 \
    pg_dump -U <your_username> -h postgres -d <your_database> -f /backups/opengin.sql

# Create compressed archive
cd ./backups/postgres
tar -czf opengin.tar.gz opengin.sql
rm -rf opengin.sql
```

#### 2.2 Restore with Volume Mount

```bash
# Extract backup file
tar -xzf opengin.tar.gz

# Run psql with volume mount
docker run --rm \
    --network=ldf-network \
    --volume=postgres_data:/var/lib/postgresql/data \
    --volume=$(pwd)/backups/postgres:/backups \
    postgres:16 \
    psql -U <your_username> -h postgres -d <your_database> -f /backups/opengin.sql

# Clean up
rm -rf opengin.sql
```

## Configuration

### Environment Variables

The backup process uses the following environment variables from `configs/backup.env`:

```bash
# PostgreSQL Backup Configuration
POSTGRES_BACKUP_DIR=/path/to/backup/directory
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DATABASE=opengin
```

### Docker Compose Volumes

The PostgreSQL service uses the following volumes:

```yaml
volumes:
  - postgres_data:/var/lib/postgresql/data    # Database data
  - postgres_backup:/var/lib/postgresql/backup # Backup storage
```

## Backup Strategies

### 1. Full Database Backup

```bash
# Backup entire database
docker exec postgres pg_dump -U <your_username> -h localhost -d <your_database> -f /var/lib/postgresql/backup/opengin.sql
```

### 2. Specific Schema Backup

```bash
# Backup specific schema
docker exec postgres pg_dump -U <your_username> -h localhost -d <your_database> -n <schema_name> -f /var/lib/postgresql/backup/schema_backup.sql
```

### 3. Compressed Backup

```bash
# Create compressed backup
docker exec postgres pg_dump -U <your_username> -h localhost -d <your_database> -Z 9 -f /var/lib/postgresql/backup/opengin.sql.gz
```

### 4. Custom Format Backup

```bash
# Create custom format backup (binary format)
docker exec postgres pg_dump -U <your_username> -h localhost -d <your_database> -Fc -f /var/lib/postgresql/backup/opengin.dump
```

## Restore Strategies

### 1. Full Database Restore

```bash
# Restore entire database
docker exec postgres psql -U <your_username> -d <your_database> -f /var/lib/postgresql/backup/opengin.sql
```

### 2. Custom Format Restore

```bash
# Restore from custom format backup
docker exec postgres pg_restore -U <your_username> -h localhost -d <your_database> /var/lib/postgresql/backup/opengin.dump
```

### 3. Schema-only Restore

```bash
# Restore schema only (no data)
docker exec postgres psql -U <your_username> -d <your_database> -f /var/lib/postgresql/backup/schema_backup.sql
```

## Troubleshooting

### Common Issues

#### 1. Connection Refused

```bash
# Check if PostgreSQL container is running
docker ps | grep postgres

# Check container logs
docker logs postgres

# Test connection
docker exec postgres pg_isready -U postgres
```

#### 2. Permission Denied

```bash
# Check file permissions in container
docker exec postgres ls -la /var/lib/postgresql/backup/

# Fix permissions if needed
docker exec postgres chown postgres:postgres /var/lib/postgresql/backup/
```

#### 3. Database Not Found

```bash
# List available databases
docker exec postgres psql -U postgres -c "\l"

# Create database if needed
docker exec postgres createdb -U postgres <database_name>
```

#### 4. Backup File Not Found

```bash
# Check if backup directory exists
docker exec postgres ls -la /var/lib/postgresql/

# Create backup directory
docker exec postgres mkdir -p /var/lib/postgresql/backup
```

### Advanced Troubleshooting

#### 1. Check PostgreSQL Version

```bash
docker exec postgres psql -U postgres -c "SELECT version();"
```

#### 2. Check Database Size

```bash
docker exec postgres psql -U postgres -c "SELECT pg_size_pretty(pg_database_size('<database_name>'));"
```

#### 3. Check Active Connections

```bash
docker exec postgres psql -U postgres -c "SELECT * FROM pg_stat_activity;"
```

## Restoring in Neon 

```bash
psql "<connection-string-from-neon>" -f <path-to-backup-folder>/opengin.sql
```

## Best Practices

### 1. Regular Backups

- Schedule automated backups using cron jobs
- Keep multiple backup versions
- Test restore procedures regularly

### 2. Backup Storage

- Store backups in multiple locations
- Use compression to save space
- Encrypt sensitive backup data

### 3. Monitoring

- Monitor backup success/failure
- Set up alerts for backup failures
- Log backup activities

### 4. Security

- Use strong passwords
- Limit backup file permissions
- Secure backup storage locations
