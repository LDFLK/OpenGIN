#!/bin/bash
set -e

# Configuration
GITHUB_REPO="LDFLK/data-backups"
ENVIRONMENT="${ENVIRONMENT:-development}"
VERSION="${1:-latest}"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    local level=$1
    shift
    local message="$*"
    case $level in
        "INFO") echo -e "${BLUE}[INFO]${NC} $message" ;;
        "SUCCESS") echo -e "${GREEN}[SUCCESS]${NC} $message" ;;
        "ERROR") echo -e "${RED}[ERROR]${NC} $message" ;;
    esac
}

# Prerequisites Checks
if [ -z "$SUPABASE_DB_URL" ]; then
    log "ERROR" "SUPABASE_DB_URL environment variable is required."
    echo -e "Example: export SUPABASE_DB_URL='postgresql://user:pass@host:5432/db'"
    exit 1
fi

if ! command -v psql &> /dev/null; then
    log "ERROR" "psql could not be found. Please install PostgreSQL client tools."
    exit 1
fi

if ! command -v wget &> /dev/null; then
    log "ERROR" "wget could not be found. Please install wget."
    exit 1
fi

if ! command -v unzip &> /dev/null; then
    log "ERROR" "unzip could not be found. Please install unzip."
    exit 1
fi

# Create temporary directory
TEMP_DIR=$(mktemp -d)
log "INFO" "Created temporary directory: $TEMP_DIR"

cleanup() {
    log "INFO" "Cleaning up temporary files..."
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# 1. Download Release
ARCHIVE_URL="https://github.com/$GITHUB_REPO/archive/refs/tags/$VERSION.zip"
ARCHIVE_FILE="$TEMP_DIR/archive.zip"

log "INFO" "Downloading GitHub archive for version $VERSION from $ARCHIVE_URL..."

if wget -q "$ARCHIVE_URL" -O "$ARCHIVE_FILE"; then
    log "SUCCESS" "Download complete."
else
    log "ERROR" "Failed to download archive. Please check the version/tag."
    exit 1
fi

# 2. Extract Archive
log "INFO" "Extracting archive..."
unzip -q "$ARCHIVE_FILE" -d "$TEMP_DIR"
EXTRACTED_DIR="$TEMP_DIR/data-backups-$VERSION"

# Note: If version is 'latest', the folder name inside might be based on the commit hash or tag name
# We should try to find the directory if the predictable name fails, or just assume the single directory extracted
if [ ! -d "$EXTRACTED_DIR" ]; then
    EXTRACTED_DIR=$(find "$TEMP_DIR" -maxdepth 1 -type d -name "data-backups-*" | head -n 1)
fi

log "INFO" "Extracted to: $EXTRACTED_DIR"

# 3. Locate Postgres Backup
# Structure: opengin/$environment/postgres/opengin.tar.gz
BACKUP_ARCHIVE="$EXTRACTED_DIR/opengin/$ENVIRONMENT/postgres/opengin.tar.gz"

if [ ! -f "$BACKUP_ARCHIVE" ]; then
    log "ERROR" "PostgreSQL backup not found at: $BACKUP_ARCHIVE"
    exit 1
fi

log "INFO" "Found backup archive: $BACKUP_ARCHIVE"

# 4. Extract SQL Dump
log "INFO" "Extracting SQL dump from archive..."
tar -xzf "$BACKUP_ARCHIVE" -C "$TEMP_DIR"

# The dump file name inside the tar usually matches the tar name base or is specifically named.
# Based on init.sh, backup_postgres creates `opengin.tar.gz` containing `${backup_file}.sql` which defaults to `opengin.sql`
DUMP_FILE="$TEMP_DIR/opengin.sql"

if [ ! -f "$DUMP_FILE" ]; then
    # Fallback: find any .sql file in temp dir
    DUMP_FILE=$(find "$TEMP_DIR" -maxdepth 1 -name "*.sql" | head -n 1)
fi

if [ ! -f "$DUMP_FILE" ]; then
    log "ERROR" "Could not find extracted SQL file."
    exit 1
fi

log "INFO" "SQL Dump ready: $DUMP_FILE"

# 5. Restore to Supabase
log "INFO" "Using local psql: $(psql --version)"
log "INFO" "Restoring to Supabase..."

if psql -d "$SUPABASE_DB_URL" -f "$DUMP_FILE"; then
    log "SUCCESS" "Migration completed successfully!"
else
    log "ERROR" "Migration failed."
    exit 1
fi
