#!/bin/bash

# Opengin Database Backup and Restore Management Script
# 
# This script provides a comprehensive backup solution for MongoDB, PostgreSQL, and Neo4j databases.
# It includes setup, backup, restore, and list operations for each database type, plus GitHub integration
# for automated backup restoration from the LDFLK/data-backups repository.
#
# Usage:
#   ./init.sh {command} [options]
#
# Database Commands:
#   backup_mongodb        - Create MongoDB backup (creates opengin.tar.gz)
#   restore_mongodb       - Restore MongoDB from backup (uses MONGODB_BACKUP_DIR)
#   list_mongodb_backups  - List available MongoDB backups
#
#   backup_postgres       - Create PostgreSQL backup (creates opengin.tar.gz)
#   restore_postgres      - Restore PostgreSQL from backup (uses POSTGRES_BACKUP_DIR)
#   list_postgres_backups - List available PostgreSQL backups
#
#   backup_neo4j          - Create Neo4j backup (creates neo4j_YYYYMMDD_HHMMSS.dump)
#   restore_neo4j         - Restore Neo4j from backup file
#   list_neo4j_backups    - List available Neo4j backups
#
# Neo4j Service Management:
#   setup_neo4j           - Build Neo4j Docker image using docker-compose
#   run_neo4j             - Start Neo4j service using docker-compose
#
# GitHub Backup Integration:
#   restore_from_github   - Restore all databases from GitHub (version optional)
#   list_github_versions  - List available versions from GitHub repository
#   get_latest_github_version - Get latest version from GitHub
#
# Utility Commands:
#   setup                 - Load environment variables and setup backup environment
#   execute               - Execute backup operations
#   finalize              - Finalize backup process
#   help                  - Display usage information
#
# Examples:
#   # Individual database operations
#   ./init.sh backup_mongodb
#   ./init.sh restore_mongodb
#   ./init.sh backup_postgres
#   ./init.sh restore_postgres
#   ./init.sh backup_neo4j
#   ./init.sh restore_neo4j ./backups/neo4j/neo4j_20250115_143022.dump
#
#   # GitHub integration
#   ./init.sh restore_from_github 0.0.1  # Restore specific version
#   ./init.sh restore_from_github        # Restore latest version
#   ./init.sh list_github_versions       # List all available versions
#
#   # Service management
#   ./init.sh setup_neo4j
#   ./init.sh run_neo4j
#
# Environment Variables (from configs/backup.env):
#   MONGODB_BACKUP_DIR    - Directory for MongoDB backups
#   POSTGRES_BACKUP_DIR   - Directory for PostgreSQL backups
#   NEO4J_BACKUP_DIR      - Directory for Neo4j backups
#   MONGODB_USERNAME      - MongoDB username
#   MONGODB_PASSWORD      - MongoDB password
#   MONGODB_DATABASE      - MongoDB database name
#   POSTGRES_USER         - PostgreSQL username
#   POSTGRES_PASSWORD     - PostgreSQL password
#   POSTGRES_DATABASE     - PostgreSQL database name
#   ENVIRONMENT           - Environment (development/staging/production)

set -e

# Configuration
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    local level=$1
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case $level in
        "INFO")
            echo -e "${BLUE}[INFO]${NC} $message"
            ;;
        "SUCCESS")
            echo -e "${GREEN}[SUCCESS]${NC} $message"
            ;;
        "WARNING")
            echo -e "${YELLOW}[WARNING]${NC} $message"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} $message"
            ;;
    esac
}

# Setup function
setup() {
    log "INFO" "Setting up backup environment..."
    source ../../configs/backup.env
    echo "NEO4J_BACKUP_DIR: $NEO4J_BACKUP_DIR"
    echo "POSTGRES_BACKUP_DIR: $POSTGRES_BACKUP_DIR"
    echo "MONGODB_BACKUP_DIR: $MONGODB_BACKUP_DIR"
    # Add your setup logic here
    log "SUCCESS" "Setup completed"
}

# Setup Neo4j - Build the Docker image
setup_neo4j() {
    log "INFO" "Building Neo4j-backup service using docker-compose..."
    docker-compose -f ../../docker-compose.yml build neo4j
    log "SUCCESS" "Neo4j-backup service built successfully"
}

# Run Neo4j container
run_neo4j() {
    log "INFO" "Loading environment variables..."
    source ../../configs/backup.env
    
    log "INFO" "Starting Neo4j-backup service using docker-compose..."
    docker-compose -f ../../docker-compose.yml up -d neo4j
    log "SUCCESS" "Neo4j-backup service started successfully"
}

# MongoDB Backup Functions
backup_mongodb() {
    log "INFO" "Starting MongoDB backup..."
    source ../../configs/backup.env
    
    local backup_dir="${MONGODB_BACKUP_DIR:-./backups/mongodb}"
    local backup_file="opengin"
    
    # Create backup directory
    mkdir -p "$backup_dir"
    
    log "INFO" "Creating MongoDB dump..."
    echo "MONGODB_USERNAME: $MONGODB_USERNAME"
    echo "MONGODB_PASSWORD: $MONGODB_PASSWORD"
    echo "MONGODB_DATABASE: $MONGODB_DATABASE"
    echo "backup_dir: $backup_dir"
    echo "backup_file: $backup_file"
    
    # Ensure backup directory exists in container
    log "INFO" "Creating backup directory in container..."
    docker exec mongodb mkdir -p /data/backup
    
    # Test MongoDB connection first
    log "INFO" "Testing MongoDB connection..."
    if docker exec mongodb mongo --host=localhost:27017 --username=${MONGODB_USERNAME} --password=${MONGODB_PASSWORD} --authenticationDatabase=admin --eval "db.adminCommand('ping')" > /dev/null 2>&1; then
        log "SUCCESS" "MongoDB connection successful"
    else
        log "ERROR" "MongoDB connection failed"
        return 1
    fi
    
    # Check what databases exist
    log "INFO" "Checking available databases..."
    docker exec mongodb mongo --host=localhost:27017 --username=${MONGODB_USERNAME} --password=${MONGODB_PASSWORD} --authenticationDatabase=admin --eval "db.adminCommand('listDatabases')"
    
    # Check if target database exists and has collections
    log "INFO" "Checking target database: ${MONGODB_DATABASE}"
    docker exec mongodb mongo --host=localhost:27017 --username=${MONGODB_USERNAME} --password=${MONGODB_PASSWORD} --authenticationDatabase=admin --eval "db = db.getSiblingDB('${MONGODB_DATABASE}'); db.getCollectionNames()"
    
    # Run mongodump and capture output
    log "INFO" "Running mongodump command..."
    mongodump_output=$(docker exec mongodb mongodump --host=localhost:27017 \
        --username=${MONGODB_USERNAME} --password=${MONGODB_PASSWORD} \
        --authenticationDatabase=admin \
        --db=${MONGODB_DATABASE} \
        --out="/data/backup/${backup_file}" 2>&1)
    mongodump_exit_code=$?
    
    log "INFO" "Mongodump output: $mongodump_output"
    log "INFO" "Mongodump exit code: $mongodump_exit_code"
    
    if [ $mongodump_exit_code -eq 0 ]; then
        
        log "SUCCESS" "MongoDB dump command completed"
        
        # Check what was actually created
        log "INFO" "Checking backup directory contents..."
        docker exec mongodb ls -la "/data/backup/"
        
        # Verify backup was created
        log "INFO" "Verifying backup files..."
        if docker exec mongodb test -d "/data/backup/${backup_file}"; then
            docker exec mongodb ls -la "/data/backup/${backup_file}"
        else
            log "WARNING" "Backup directory not found, trying alternative approach..."
            # Try creating the directory first and running mongodump again
            docker exec mongodb mkdir -p "/data/backup/${backup_file}"
            log "INFO" "Retrying mongodump with pre-created directory..."
            if docker exec mongodb mongodump --host=localhost:27017 \
                --username=${MONGODB_USERNAME} --password=${MONGODB_PASSWORD} \
                --authenticationDatabase=admin \
                --db=${MONGODB_DATABASE} \
                --out="/data/backup/${backup_file}" 2>&1; then
                log "SUCCESS" "Retry successful"
                docker exec mongodb ls -la "/data/backup/${backup_file}"
            else
                log "ERROR" "Backup failed even with retry"
                return 1
            fi
        fi
        
        # Copy backup from container to host
        log "INFO" "Copying backup to host..."
        docker cp "mongodb:/data/backup/${backup_file}" "$backup_dir/"
        
        # Create compressed archive
        log "INFO" "Creating compressed archive..."
        cd "$backup_dir"
        tar -czf "opengin.tar.gz" "$backup_file"
        rm -rf "$backup_file"
        
        # Clean up container backup
        docker exec mongodb rm -rf "/data/backup/${backup_file}"
        
        log "SUCCESS" "MongoDB backup completed: opengin.tar.gz"
    else
        log "ERROR" "MongoDB backup failed"
        return 1
    fi
}

# MongoDB Restore Functions
restore_mongodb() {
    log "INFO" "Starting MongoDB restore..."
    source ../../configs/backup.env
    
    # Accept optional backup file path parameter
    local backup_file="$1"
    
    # If no parameter provided, use default backup directory
    if [ -z "$backup_file" ]; then
        local backup_dir="${MONGODB_BACKUP_DIR:-./backups/mongodb}"

        if [ ! -d "$backup_dir" ]; then
            log "ERROR" "Backup directory not found: $backup_dir"
            return 1
        fi

        log "INFO" "Using backup directory: $backup_dir"

        # Look for opengin.tar.gz file
        backup_file="$backup_dir/opengin.tar.gz"
    fi
    
    
    if [ ! -f "$backup_file" ]; then
        log "ERROR" "Backup file not found: $backup_file"
        return 1
    fi
    
    log "INFO" "Using backup file: $(basename "$backup_file")"
    
    # Extract backup file
    local temp_dir=$(mktemp -d)
    local backup_name="opengin"
    
    log "INFO" "Extracting backup file..."
    tar -xzf "$backup_file" -C "$temp_dir"
    
    # Copy backup to container
    log "INFO" "Copying backup to container..."
    docker cp "$temp_dir/$backup_name" "mongodb:/data/backup/"
    
    # Check what was actually created
    log "INFO" "Checking backup structure in container..."
    docker exec mongodb find "/data/backup/$backup_name" -type d -name "*" 2>/dev/null || true
    docker exec mongodb ls -la "/data/backup/$backup_name" 2>/dev/null || true
    
    # Use the backup directory directly (mongorestore will handle the database structure)
    local db_path="/data/backup/$backup_name"
    
    log "INFO" "Using backup path: $db_path"
    
    # Restore database
    log "INFO" "Restoring MongoDB database..."
    if docker exec mongodb mongorestore --host=localhost:27017 \
        --username=${MONGODB_USERNAME} --password=${MONGODB_PASSWORD} \
        --authenticationDatabase=admin \
        --drop \
        "$db_path"; then
        
        log "SUCCESS" "MongoDB restore completed successfully"
        
        # Verify what was restored
        log "INFO" "Verifying restored databases..."
        docker exec mongodb mongo --host=localhost:27017 --username=${MONGODB_USERNAME} --password=${MONGODB_PASSWORD} --authenticationDatabase=admin --eval "db.adminCommand('listDatabases')"
        
        # Clean up
        docker exec mongodb rm -rf "/data/backup/$backup_name"
        rm -rf "$temp_dir"
        
    else
        log "ERROR" "MongoDB restore failed"
        rm -rf "$temp_dir"
        return 1
    fi
}

# List MongoDB backups
list_mongodb_backups() {
    source ../../configs/backup.env
    local backup_dir="${MONGODB_BACKUP_DIR:-./backups/mongodb}"
    
    log "INFO" "MongoDB backups in: $backup_dir"
    
    if [ -d "$backup_dir" ]; then
        ls -la "$backup_dir"/*.tar.gz 2>/dev/null || log "WARNING" "No backup files found"
    else
        log "WARNING" "Backup directory does not exist: $backup_dir"
    fi
}

# PostgreSQL Backup Functions
backup_postgres() {
    log "INFO" "Starting PostgreSQL backup..."
    source ../../configs/backup.env
    
    local backup_dir="${POSTGRES_BACKUP_DIR:-./backups/postgres}"
    local backup_file="opengin"
    
    # Create backup directory
    mkdir -p "$backup_dir"
    
    log "INFO" "Creating PostgreSQL dump..."
    echo "POSTGRES_USER: $POSTGRES_USER"
    echo "POSTGRES_PASSWORD: $POSTGRES_PASSWORD"
    echo "POSTGRES_DATABASE: $POSTGRES_DATABASE"
    echo "backup_dir: $backup_dir"
    echo "backup_file: $backup_file"
    
    # Ensure backup directory exists in container
    log "INFO" "Creating backup directory in container..."
    docker exec postgres mkdir -p /var/lib/postgresql/backup
    
    # Test PostgreSQL connection first
    log "INFO" "Testing PostgreSQL connection..."
    if docker exec postgres pg_isready -U postgres > /dev/null 2>&1; then
        log "SUCCESS" "PostgreSQL connection successful"
    else
        log "ERROR" "PostgreSQL connection failed"
        return 1
    fi
    
    # Check what databases exist
    log "INFO" "Checking available databases..."
    docker exec postgres psql -U postgres -c "\l"
    
    # Run pg_dump and capture output
    log "INFO" "Running pg_dump command..."
    pg_dump_output=$(docker exec postgres pg_dump -U postgres -h localhost -d ${POSTGRES_DATABASE} -f "/var/lib/postgresql/backup/${backup_file}.sql" 2>&1)
    pg_dump_exit_code=$?
    
    log "INFO" "Pg_dump output: $pg_dump_output"
    log "INFO" "Pg_dump exit code: $pg_dump_exit_code"
    
    if [ $pg_dump_exit_code -eq 0 ]; then
        log "SUCCESS" "PostgreSQL dump command completed"
        
        # Check what was actually created
        log "INFO" "Checking backup directory contents..."
        docker exec postgres ls -la "/var/lib/postgresql/backup/"
        
        # Copy backup from container to host
        log "INFO" "Copying backup to host..."
        docker cp "postgres:/var/lib/postgresql/backup/${backup_file}.sql" "$backup_dir/"
        
        # Create compressed archive
        log "INFO" "Creating compressed archive..."
        cd "$backup_dir"
        tar -czf "opengin.tar.gz" "${backup_file}.sql"
        rm -rf "${backup_file}.sql"
        
        # Clean up container backup
        docker exec postgres rm -rf "/var/lib/postgresql/backup/${backup_file}.sql"
        
        log "SUCCESS" "PostgreSQL backup completed: opengin.tar.gz"
    else
        log "ERROR" "PostgreSQL backup failed"
        return 1
    fi
}

# PostgreSQL Restore Functions
restore_postgres() {
    log "INFO" "Starting PostgreSQL restore..."
    source ../../configs/backup.env
    
    # Accept optional backup file path parameter
    local backup_file="$1"
    
    # If no parameter provided, use default backup directory
    if [ -z "$backup_file" ]; then
        local backup_dir="${POSTGRES_BACKUP_DIR:-./backups/postgres}"

        if [ ! -d "$backup_dir" ]; then
            log "ERROR" "Backup directory not found: $backup_dir"
            return 1
        fi

        log "INFO" "Using backup directory: $backup_dir"

        # Look for opengin.tar.gz file
        backup_file="$backup_dir/opengin.tar.gz"
    fi
    
    
    if [ ! -f "$backup_file" ]; then
        log "ERROR" "Backup file not found: $backup_file"
        return 1
    fi
    
    log "INFO" "Using backup file: $(basename "$backup_file")"
    
    # Extract backup file
    local temp_dir=$(mktemp -d)
    local backup_name="opengin.sql"
    
    log "INFO" "Extracting backup file..."
    tar -xzf "$backup_file" -C "$temp_dir"
    
    # Copy backup to container
    log "INFO" "Copying backup to container..."
    docker cp "$temp_dir/$backup_name" "postgres:/var/lib/postgresql/backup/"
    
    # Check what was actually created
    log "INFO" "Checking backup structure in container..."
    docker exec postgres ls -la "/var/lib/postgresql/backup/"
    
    # Restore database
    log "INFO" "Restoring PostgreSQL database..."
    if docker exec postgres psql -U postgres -d ${POSTGRES_DATABASE} -f "/var/lib/postgresql/backup/$backup_name"; then
        log "SUCCESS" "PostgreSQL restore completed successfully"
        
        # Verify what was restored
        log "INFO" "Verifying restored database..."
        docker exec postgres psql -U postgres -d ${POSTGRES_DATABASE} -c "\dt"
        
        # Clean up
        docker exec postgres rm -rf "/var/lib/postgresql/backup/$backup_name"
        rm -rf "$temp_dir"
        
    else
        log "ERROR" "PostgreSQL restore failed"
        rm -rf "$temp_dir"
        return 1
    fi
}

# List PostgreSQL backups
list_postgres_backups() {
    source ../../configs/backup.env
    local backup_dir="${POSTGRES_BACKUP_DIR:-./backups/postgres}"
    
    log "INFO" "PostgreSQL backups in: $backup_dir"
    
    if [ -d "$backup_dir" ]; then
        ls -la "$backup_dir"/*.tar.gz 2>/dev/null || log "WARNING" "No backup files found"
    else
        log "WARNING" "Backup directory does not exist: $backup_dir"
    fi
}

# Neo4j Restore Function
restore_neo4j() {
    log "INFO" "Starting Neo4j restore..."
    source ../../configs/backup.env
    
    # Accept optional backup file path parameter
    local backup_file="$1"
    
    # If no parameter provided, use default backup directory
    if [ -z "$backup_file" ]; then
        local backup_dir="${NEO4J_BACKUP_DIR:-./backups/neo4j}"

        if [ ! -d "$backup_dir" ]; then
            log "ERROR" "Backup directory not found: $backup_dir"
            return 1
        fi

        log "INFO" "Using backup directory: $backup_dir"

        # Look for neo4j.dump file
        backup_file="$backup_dir/neo4j.dump"
    fi
    
    if [ ! -f "$backup_file" ]; then
        log "ERROR" "Backup file not found: $backup_file"
        return 1
    fi
    
    log "INFO" "Using backup file: $(basename "$backup_file")"
    
    # Check if Neo4j container is running
    local neo4j_running=false
    if docker-compose -f ../../docker-compose.yml ps neo4j | grep -q "Up"; then
        neo4j_running=true
        log "INFO" "Neo4j container is running, will stop it for restore"
        
        # Wait for Neo4j to be ready before stopping
        if ! wait_for_service "neo4j" 7687; then
            log "WARNING" "Neo4j service not responding, proceeding anyway"
        fi
        
        # Stop Neo4j container before restore
        log "INFO" "Stopping Neo4j container for restore..."
        docker-compose -f ../../docker-compose.yml stop neo4j
    else
        log "INFO" "Neo4j container is already stopped, proceeding with restore"
    fi
    
    # Get Neo4j data volume path
    local neo4j_volume=$(docker volume inspect opengin_neo4j_data --format '{{ .Mountpoint }}' 2>/dev/null || echo "")
    if [ -z "$neo4j_volume" ]; then
        log "ERROR" "Could not find Neo4j data volume"
        return 1
    fi
    
    log "INFO" "Neo4j data volume: $neo4j_volume"
    
    # Create temporary directory for backup
    local temp_backup_dir=$(mktemp -d)
    log "INFO" "Copying backup file to temporary directory: $temp_backup_dir"
    cp "$backup_file" "$temp_backup_dir/"
    
    # Verify the file was copied
    local backup_filename=$(basename "$backup_file")
    if [ ! -f "$temp_backup_dir/$backup_filename" ]; then
        log "ERROR" "Failed to copy backup file to temporary directory"
        rm -rf "$temp_backup_dir"
        return 1
    fi
    
    # Restore using neo4j-admin with proper volume mounting
    log "INFO" "Restoring Neo4j database using neo4j-admin..."
    if docker run --rm \
        --volume="$neo4j_volume:/data" \
        --volume="$temp_backup_dir:/backups" \
        neo4j/neo4j-admin:latest \
        neo4j-admin database load neo4j --from-path=/backups --overwrite-destination=true; then
        
        log "SUCCESS" "Neo4j restore completed"
        
        # Clean up temporary directory
        rm -rf "$temp_backup_dir"
        
        # Only start Neo4j container if it was running before
        if [ "$neo4j_running" = true ]; then
            log "INFO" "Starting Neo4j container..."
            docker-compose -f ../../docker-compose.yml start neo4j
            
            # Wait for Neo4j to be ready again
            if wait_for_service "neo4j" 7687; then
                log "SUCCESS" "Neo4j container started successfully"
                return 0
            else
                log "ERROR" "Neo4j container failed to start after restore"
                return 1
            fi
        else
            log "INFO" "Neo4j container was already stopped, leaving it stopped"
            return 0
        fi
    else
        log "ERROR" "Neo4j restore failed"
        rm -rf "$temp_backup_dir"
        return 1
    fi
}

# Wait for service to be ready
wait_for_service() {
    local service=$1
    local port=$2
    local timeout=${3:-300}
    
    log "INFO" "Waiting for $service to be ready on port $port..."
    
    # Determine the correct hostname to use
    local hostname="localhost"
    if [ "$service" = "neo4j" ]; then
        # For Neo4j, we can use localhost since ports are exposed
        hostname="localhost"
    elif [ "$service" = "mongodb" ]; then
        # For MongoDB, we can use localhost since ports are exposed
        hostname="localhost"
    elif [ "$service" = "postgres" ]; then
        # For PostgreSQL, we can use localhost since ports are exposed
        hostname="localhost"
    else
        # For other services, try the service name first, then localhost
        hostname="$service"
    fi
    
    local count=0
    while [ $count -lt $timeout ]; do
        # Use different health check methods based on service
        local is_ready=false
        
        if [ "$service" = "neo4j" ]; then
            # For Neo4j, check both HTTP (7474) and Bolt (7687) ports
            if nc -z $hostname 7474 2>/dev/null && nc -z $hostname 7687 2>/dev/null; then
                is_ready=true
            fi
        elif [ "$service" = "mongodb" ]; then
            # For MongoDB, check if we can connect to the database
            if docker exec mongodb mongo --eval "db.adminCommand('ping')" > /dev/null 2>&1; then
                is_ready=true
            fi
        elif [ "$service" = "postgres" ]; then
            # For PostgreSQL, check if pg_isready works
            if docker exec postgres pg_isready -U postgres > /dev/null 2>&1; then
                is_ready=true
            fi
        else
            # For other services, use basic port check
            if nc -z $hostname $port 2>/dev/null; then
                is_ready=true
            fi
        fi
        
        if [ "$is_ready" = true ]; then
            log "SUCCESS" "$service is ready on $hostname:$port"
            return 0
        fi
        
        sleep 5
        count=$((count + 5))
    done
    
    log "ERROR" "Timeout waiting for $service on $hostname:$port"
    return 1
}

# GitHub Backup Manager Functions
restore_from_github() {
    # TODO: make sure there is a release tag called latest in data-backups repository.
    #  Make sure to always update this release with what is in the main branch which is
    #  the latest version ready for releasing.
    local version="${1:-latest}"
    log "INFO" "Restoring from GitHub version: $version"
    
    # Set GitHub repository
    local github_repo="LDFLK/data-backups"
    local environment="${ENVIRONMENT:-development}"
    
    # Create temporary directory for downloads
    local temp_dir=$(mktemp -d)
    local results=""
    
    # Function to download and extract files from GitHub archive
    download_github_archive() {
        local version="$1"
        local extract_dir="$2"
        
        log "INFO" "Downloading GitHub archive for version $version..."
        
        # Download the archive
        local archive_url="https://github.com/$github_repo/archive/refs/tags/$version.zip"
        local archive_file="$extract_dir/archive.zip"
        
        if wget -q "$archive_url" -O "$archive_file"; then
            log "SUCCESS" "Downloaded archive for version $version"
            
            # Extract the archive
            if unzip -q "$archive_file" -d "$extract_dir"; then
                log "SUCCESS" "Extracted archive"
                rm -f "$archive_file"  # Clean up archive file
                return 0
            else
                log "ERROR" "Failed to extract archive"
                return 1
            fi
        else
            log "ERROR" "Failed to download archive for version $version"
            return 1
        fi
    }
    
    # Download the entire archive
    if ! download_github_archive "$version" "$temp_dir"; then
        log "ERROR" "Failed to download GitHub archive for version $version"
        rm -rf "$temp_dir"
        return 1
    fi
    
    # Set the extracted directory path
    local archive_dir="$temp_dir/data-backups-$version"
    
    # Download and restore MongoDB
    log "INFO" "Processing MongoDB backup..."
    local mongodb_file="$archive_dir/opengin/version/$version/$environment/mongodb/opengin.tar.gz"
    if [ -f "$mongodb_file" ]; then
        if restore_mongodb "$mongodb_file"; then
            results="${results}mongodb:true,"
            log "SUCCESS" "MongoDB restored successfully"
        else
            results="${results}mongodb:false,"
            log "ERROR" "MongoDB restore failed"
        fi
    else
        results="${results}mongodb:false,"
        log "ERROR" "MongoDB backup not found: $mongodb_file"
    fi
    
    # Download and restore PostgreSQL
    log "INFO" "Processing PostgreSQL backup..."
    local postgres_file="$archive_dir/opengin/version/$version/$environment/postgres/opengin.tar.gz"
    if [ -f "$postgres_file" ]; then
        if restore_postgres "$postgres_file"; then
            results="${results}postgres:true,"
            log "SUCCESS" "PostgreSQL restored successfully"
        else
            results="${results}postgres:false,"
            log "ERROR" "PostgreSQL restore failed"
        fi
    else
        results="${results}postgres:false,"
        log "ERROR" "PostgreSQL backup not found: $postgres_file"
    fi
    
    # Download and restore Neo4j
    log "INFO" "Processing Neo4j backup..."
    local neo4j_file="$archive_dir/opengin/version/$version/$environment/neo4j/neo4j.dump"
    if [ -f "$neo4j_file" ]; then
        if restore_neo4j "$neo4j_file"; then
            results="${results}neo4j:true,"
            log "SUCCESS" "Neo4j restored successfully"
        else
            results="${results}neo4j:false,"
            log "ERROR" "Neo4j restore failed"
        fi
    else
        results="${results}neo4j:false,"
        log "ERROR" "Neo4j backup not found: $neo4j_file"
    fi
    
    # Clean up
    rm -rf "$temp_dir"
    
    # Display results
    log "INFO" "GitHub restore completed for version $version"
    log "INFO" "Results: ${results%,}"
    
    # Check if all succeeded
    if [[ "$results" == *"mongodb:true"* ]] && [[ "$results" == *"postgres:true"* ]] && [[ "$results" == *"neo4j:true"* ]]; then
        log "SUCCESS" "All databases restored successfully from GitHub"
        return 0
    else
        log "WARNING" "Some databases may not have been restored successfully"
        return 1
    fi
}

list_github_versions() {
    log "INFO" "Fetching available versions from GitHub..."
    
    local github_repo="LDFLK/data-backups"
    
    # Get releases
    log "INFO" "Fetching releases..."
    local releases=$(curl -s "https://api.github.com/repos/$github_repo/releases")
    
    if [ $? -eq 0 ]; then
        log "SUCCESS" "Available versions:"
        echo "$releases" | jq -r '.[] | select(.tag_name | startswith("0.0.")) | "release: \(.tag_name) - \(.name) (\(.created_at))"'
    else
        log "ERROR" "Failed to fetch releases from GitHub"
        return 1
    fi
    
    # Get branches
    log "INFO" "Fetching branches..."
    local branches=$(curl -s "https://api.github.com/repos/$github_repo/branches")
    
    if [ $? -eq 0 ]; then
        echo "$branches" | jq -r '.[] | select(.name | startswith("release-0.0.")) | "branch: \(.name | sub("release-"; "")) - Branch \(.name) (\(.commit.commit.author.date))"'
    else
        log "ERROR" "Failed to fetch branches from GitHub"
        return 1
    fi
}

get_latest_github_version() {
    log "INFO" "Getting latest version from GitHub..."
    
    local github_repo="LDFLK/data-backups"
    
    # Get releases and find the latest
    local releases=$(curl -s "https://api.github.com/repos/$github_repo/releases")
    
    if [ $? -eq 0 ]; then
        local latest=$(echo "$releases" | jq -r '.[] | select(.tag_name | startswith("0.0.")) | .tag_name' | head -1)
        if [ -n "$latest" ]; then
            log "SUCCESS" "Latest version: $latest"
            echo "$latest"
            return 0
        fi
    fi
    
    # If no releases, check branches
    local branches=$(curl -s "https://api.github.com/repos/$github_repo/branches")
    
    if [ $? -eq 0 ]; then
        local latest=$(echo "$branches" | jq -r '.[] | select(.name | startswith("release-0.0.")) | .name | sub("release-"; "")' | head -1)
        if [ -n "$latest" ]; then
            log "SUCCESS" "Latest version: $latest"
            echo "$latest"
            return 0
        fi
    fi
    
    log "ERROR" "Failed to get latest version from GitHub"
    return 1
}

# Note: backup-manager service is now just a simple Alpine container that runs init.sh
# No need for separate management functions

backup_neo4j() {
    log "INFO" "Starting Neo4j backup..."
    source ../../configs/backup.env
    
    local backup_dir="${NEO4J_BACKUP_DIR:-./backups/neo4j}"
    local backup_file="neo4j.dump"
    
    # Create backup directory
    mkdir -p "$backup_dir"
    
    log "INFO" "Creating Neo4j dump..."
    echo "NEO4J_BACKUP_DIR: $backup_dir"
    echo "backup_file: $backup_file"
    
    # Check if Neo4j container is running
    local neo4j_running=false
    if docker-compose -f ../../docker-compose.yml ps neo4j | grep -q "Up"; then
        neo4j_running=true
        log "INFO" "Neo4j container is running, will stop it for backup"
        
        # Wait for Neo4j to be ready before stopping
        if ! wait_for_service "neo4j" 7687; then
            log "WARNING" "Neo4j service not responding, proceeding anyway"
        fi
        
        # Stop Neo4j container before backup
        log "INFO" "Stopping Neo4j container for backup..."
        docker-compose -f ../../docker-compose.yml stop neo4j
    else
        log "INFO" "Neo4j container is already stopped, proceeding with backup"
    fi
    
    # Get Neo4j data volume path
    local neo4j_volume=$(docker volume inspect opengin_neo4j_data --format '{{ .Mountpoint }}' 2>/dev/null || echo "")
    if [ -z "$neo4j_volume" ]; then
        log "ERROR" "Could not find Neo4j data volume"
        return 1
    fi
    
    log "INFO" "Neo4j data volume: $neo4j_volume"
    
    # Create backup using neo4j-admin with proper volume mounting
    log "INFO" "Creating Neo4j database dump..."
    if docker run --rm \
        --volume="$neo4j_volume:/data" \
        --volume="$backup_dir:/backups" \
        neo4j/neo4j-admin:latest \
        neo4j-admin database dump neo4j --to-path=/backups; then
        
        log "SUCCESS" "Neo4j dump created successfully"
        
        # Rename the dump file to include timestamp
        if [ -f "$backup_dir/neo4j.dump" ]; then
            mv "$backup_dir/neo4j.dump" "$backup_dir/$backup_file"
            log "SUCCESS" "Neo4j backup completed: $backup_file"
        else
            log "WARNING" "Neo4j dump file not found, but command succeeded"
        fi
        
        # Only start Neo4j container if it was running before
        if [ "$neo4j_running" = true ]; then
            log "INFO" "Starting Neo4j container..."
            docker-compose -f ../../docker-compose.yml start neo4j
            
            # Wait for Neo4j to be ready
            if wait_for_service "neo4j" 7687; then
                log "SUCCESS" "Neo4j container started successfully"
                return 0
            else
                log "ERROR" "Neo4j container failed to start after backup"
                return 1
            fi
        else
            log "INFO" "Neo4j container was already stopped, leaving it stopped"
            return 0
        fi
    else
        log "ERROR" "Neo4j backup failed"
        return 1
    fi
}

# List Neo4j backups
list_neo4j_backups() {
    source ../../configs/backup.env
    local backup_dir="${NEO4J_BACKUP_DIR:-./backups/neo4j}"
    
    log "INFO" "Neo4j backups in: $backup_dir"
    
    if [ -d "$backup_dir" ]; then
        ls -la "$backup_dir"/*.dump 2>/dev/null || log "WARNING" "No backup files found"
    else
        log "WARNING" "Backup directory does not exist: $backup_dir"
    fi
}

# Execute function
execute() {
    log "INFO" "Executing backup operations..."
    # Add your execute logic here
    log "SUCCESS" "Execute completed"
}

# Finalize function
finalize() {
    log "INFO" "Finalizing backup process..."
    # Add your finalize logic here
    log "SUCCESS" "Finalize completed"
}

# Main function
main() {
    local command="${1:-help}"
    
    case $command in
        "setup")
            setup
            ;;
        "setup_neo4j")
            setup_neo4j
            ;;
        "run_neo4j")
            run_neo4j
            ;;
        "backup_mongodb")
            backup_mongodb
            ;;
        "restore_mongodb")
            restore_mongodb "$2"
            ;;
        "list_mongodb_backups")
            list_mongodb_backups
            ;;
        "backup_postgres")
            backup_postgres
            ;;
        "restore_postgres")
            restore_postgres
            ;;
        "list_postgres_backups")
            list_postgres_backups
            ;;
        "backup_neo4j")
            backup_neo4j
            ;;
        "restore_neo4j")
            restore_neo4j
            ;;
        "list_neo4j_backups")
            list_neo4j_backups
            ;;
        "execute")
            execute
            ;;
        "finalize")
            finalize
            ;;
        "restore_from_github")
            restore_from_github "$2"
            ;;
        "list_github_versions")
            list_github_versions
            ;;
        "get_latest_github_version")
            get_latest_github_version
            ;;
        "help"|*)
            echo "Opengin Database Backup and Restore Management Script"
            echo ""
            echo "Usage: $0 {command} [options]"
            echo ""
            echo "Database Commands:"
            echo "  backup_mongodb        - Create MongoDB backup (creates opengin.tar.gz)"
            echo "  restore_mongodb       - Restore MongoDB from backup (uses MONGODB_BACKUP_DIR)"
            echo "  list_mongodb_backups  - List available MongoDB backups"
            echo ""
            echo "  backup_postgres       - Create PostgreSQL backup (creates opengin.tar.gz)"
            echo "  restore_postgres      - Restore PostgreSQL from backup (uses POSTGRES_BACKUP_DIR)"
            echo "  list_postgres_backups - List available PostgreSQL backups"
            echo ""
            echo "  backup_neo4j          - Create Neo4j backup (creates neo4j.dump)"
            echo "  restore_neo4j         - Restore Neo4j from backup (uses NEO4J_BACKUP_DIR)"
            echo "  list_neo4j_backups    - List available Neo4j backups"
            echo ""
            echo "Neo4j Service Management:"
            echo "  setup_neo4j           - Build Neo4j Docker image using docker-compose"
            echo "  run_neo4j             - Start Neo4j service using docker-compose"
            echo ""
            echo "GitHub Backup Integration:"
            echo "  restore_from_github   - Restore all databases from GitHub (version optional)"
            echo "  list_github_versions  - List available versions from GitHub repository"
            echo "  get_latest_github_version - Get latest version from GitHub"
            echo ""
            echo "Utility Commands:"
            echo "  setup                 - Load environment variables and setup backup environment"
            echo "  execute               - Execute backup operations"
            echo "  finalize              - Finalize backup process"
            echo "  help                  - Display this help information"
            echo ""
            echo "Examples:"
            echo "  # Individual database operations"
            echo "  $0 backup_mongodb"
            echo "  $0 restore_mongodb"
            echo "  $0 backup_postgres"
            echo "  $0 restore_postgres"
            echo "  $0 backup_neo4j"
            echo "  $0 restore_neo4j"
            echo ""
            echo "  # GitHub integration"
            echo "  $0 restore_from_github 0.0.1  # Restore specific version"
            echo "  $0 restore_from_github        # Restore latest version"
            echo "  $0 list_github_versions       # List all available versions"
            echo ""
            echo "  # Service management"
            echo "  $0 setup_neo4j"
            echo "  $0 run_neo4j"
            echo ""
            echo "Environment Variables (from configs/backup.env):"
            echo "  MONGODB_BACKUP_DIR, POSTGRES_BACKUP_DIR, NEO4J_BACKUP_DIR"
            echo "  MONGODB_USERNAME, MONGODB_PASSWORD, MONGODB_DATABASE"
            echo "  POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DATABASE"
            echo "  ENVIRONMENT (development/staging/production)"
            ;;
    esac
}

# Run main function with all arguments
main "$@"