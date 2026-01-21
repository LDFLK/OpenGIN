# MongoDB Backup and Restore Guide

This guide provides comprehensive instructions for backing up and restoring MongoDB databases in Docker containers.

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
- MongoDB container running (using docker-compose)
- Access to MongoDB container
- MongoDB credentials (replace `<your_username>` and `<your_password>` in commands below)

## Backup and Restore Commands

### Method 1: Direct Docker Commands (Recommended)

#### 1.1 Create MongoDB Backup

```bash
# Create backup directory
mkdir -p ./backups/mongodb

# Create MongoDB dump
docker exec mongodb mongodump \
    --host=localhost:27017 \
    --username=<your_username> \
    --password=<your_password> \
    --authenticationDatabase=admin \
    --db=opengin \
    --out=/data/backup/mongodb_backup_$(date +%Y%m%d_%H%M%S)

# Copy backup from container to host
docker cp mongodb:/data/backup/mongodb_backup_$(date +%Y%m%d_%H%M%S) ./backups/mongodb/

# Create compressed archive
cd ./backups/mongodb
tar -czf mongodb_backup_$(date +%Y%m%d_%H%M%S).tar.gz mongodb_backup_*
rm -rf mongodb_backup_*

# Clean up container backup
docker exec mongodb rm -rf /data/backup/mongodb_backup_*
```

#### 1.2 Restore MongoDB from Backup

```bash
# Extract backup file
tar -xzf mongodb_backup_20241215_143022.tar.gz

# Copy backup to container
docker cp mongodb_backup_20241215_143022 mongodb:/data/backup/

# Restore database
docker exec mongodb mongorestore \
    --host=localhost:27017 \
    --username=<your_username> \
    --password=<your_password> \
    --authenticationDatabase=admin \
    --db=opengin \
    --drop \
    /data/backup/mongodb_backup_20241215_143022/opengin

# Clean up
docker exec mongodb rm -rf /data/backup/mongodb_backup_20241215_143022
rm -rf mongodb_backup_20241215_143022
```

### Method 2: Using Docker Volume Mounts

#### 2.1 Create Backup with Volume Mount

```bash
# Create backup directory
mkdir -p ./backups/mongodb

# Run mongodump with volume mount
docker run --rm \
    --network=ldf-network \
    --volume=mongodb_data:/data/db \
    --volume=$(pwd)/backups/mongodb:/backups \
    mongo:4.4 \
    mongodump \
    --host=mongodb:27017 \
    --username=<your_username> \
    --password=<your_password> \
    --authenticationDatabase=admin \
    --db=opengin \
    --out=/backups/mongodb_backup_$(date +%Y%m%d_%H%M%S)

# Create compressed archive
cd ./backups/mongodb
tar -czf mongodb_backup_$(date +%Y%m%d_%H%M%S).tar.gz mongodb_backup_*
rm -rf mongodb_backup_*
```

#### 2.2 Restore with Volume Mount

```bash
# Extract backup file
tar -xzf mongodb_backup_20241215_143022.tar.gz

# Run mongorestore with volume mount
docker run --rm \
    --network=ldf-network \
    --volume=mongodb_data:/data/db \
    --volume=$(pwd)/backups/mongodb:/backups \
    mongo:4.4 \
    mongorestore \
    --host=mongodb:27017 \
    --username=<your_username> \
    --password=<your_password> \
    --authenticationDatabase=admin \
    --db=opengin \
    --drop \
    /backups/mongodb_backup_20241215_143022/opengin

# Clean up
rm -rf mongodb_backup_20241215_143022
```

## Configuration

### Environment Variables

Configure your backup settings in `configs/backup.env`:

```bash
# MongoDB Backup Configuration
MONGODB_BACKUP_DIR=/path/to/your/backups/mongodb

# MongoDB Credentials
MONGO_USER=admin
MONGO_PASSWORD=admin123
MONGO_DATABASE=opengin
```

### Docker Compose Volumes

The MongoDB service uses the following volumes:

```yaml
volumes:
  - mongodb_data:/data/db          # Database data
  - mongodb_config:/data/configdb  # Configuration
  - mongodb_backup:/data/backup    # Backup storage
```

## Backup Strategies

### 1. Full Database Backup

```bash
# Backup entire database
docker exec mongodb mongodump \
    --host=localhost:27017 \
    --username=<your_username> \
    --password=<your_password> \
    --authenticationDatabase=admin \
    --db=opengin \
    --out=/data/backup/mongodb_backup_$(date +%Y%m%d_%H%M%S)
```

### 2. Specific Collection Backup

```bash
# Backup specific collection
docker exec mongodb mongodump \
    --host=localhost:27017 \
    --username=<your_username> \
    --password=<your_password> \
    --authenticationDatabase=admin \
    --db=opengin \
    --collection=your_collection \
    --out=/data/backup/collection_backup
```

### 3. Compressed Backup

```bash
# Create compressed backup
docker exec mongodb mongodump \
    --host=localhost:27017 \
    --username=<your_username> \
    --password=<your_password> \
    --authenticationDatabase=admin \
    --db=opengin \
    --archive=/data/backup/mongodb_backup.gz \
    --gzip
```

## Restore Strategies

### 1. Full Database Restore

```bash
# Restore entire database
docker exec mongodb mongorestore \
    --host=localhost:27017 \
    --username=<your_username> \
    --password=<your_password> \
    --authenticationDatabase=admin \
    --db=opengin \
    --drop \
    /data/backup/mongodb_backup_20241215_143022/opengin
```

### 2. Specific Collection Restore

```bash
# Restore specific collection
docker exec mongodb mongorestore \
    --host=localhost:27017 \
    --username=<your_username> \
    --password=<your_password> \
    --authenticationDatabase=admin \
    --db=opengin \
    --collection=your_collection \
    /data/backup/collection_backup/opengin/your_collection.bson
```

### 3. Compressed Restore

```bash
# Restore from compressed backup
docker exec mongodb mongorestore \
    --host=localhost:27017 \
    --username=<your_username> \
    --password=<your_password> \
    --authenticationDatabase=admin \
    --db=opengin \
    --archive=/data/backup/mongodb_backup.gz \
    --gzip
```

## Troubleshooting

### Common Issues

#### 1. Permission Denied

```bash
# Check container permissions
docker exec mongodb ls -la /data/backup

# Fix permissions if needed
docker exec mongodb chown -R mongodb:mongodb /data/backup
```

#### 2. Authentication Failed

```bash
# Verify MongoDB is running
docker exec mongodb mongo --eval "db.adminCommand('ping')"

# Check credentials
docker exec mongodb mongo -u admin -p admin123 --authenticationDatabase=admin
```

#### 3. Backup Directory Not Found

```bash
# Create backup directory
docker exec mongodb mkdir -p /data/backup

# Set proper permissions
docker exec mongodb chown -R mongodb:mongodb /data/backup
```

### Verification Commands

#### Check Backup Integrity

```bash
# List backup contents
tar -tzf mongodb_backup_20241215_143022.tar.gz

# Verify database structure
docker exec mongodb mongo --eval "db.adminCommand('listCollections')"
```

#### Check Restore Success

```bash
# Verify data was restored
docker exec mongodb mongo --eval "db.stats()"

# Check specific collections
docker exec mongodb mongo --eval "db.your_collection.count()"
```

## Restore to Mongodb Atlas

Note that this dump must be taken from OpenGIN dump program. 

```bash
mongorestore --uri="<mongodb-service-uri>" --db=opengin --drop "<path-to-the-dump-folder>"
```

## Best Practices

### 1. Regular Backups

- Schedule daily backups using cron
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

- Use strong authentication credentials
- Limit backup access permissions
- Encrypt backup files

### Backup Script Example

```bash
#!/bin/bash
# Daily MongoDB backup script

BACKUP_DIR="/path/to/backups/mongodb"
LOG_FILE="/var/log/mongodb_backup.log"

# Create backup
/path/to/deployment/development/init.sh backup_mongodb >> $LOG_FILE 2>&1

# Clean up old backups (keep last 7 days)
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete

# Send notification
echo "MongoDB backup completed on $(date)" | mail -s "MongoDB Backup" admin@example.com
```

## Support

For issues or questions regarding MongoDB backup and restore:

1. Check the troubleshooting section above
2. Review MongoDB documentation
3. Check container logs: `docker logs mongodb`
4. Verify network connectivity: `docker network ls`
