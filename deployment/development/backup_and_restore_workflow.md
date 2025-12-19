# Data Backup and Restore Workflow

This guide explains how to perform data backup and restore operations using the `init.sh` script for MongoDB, PostgreSQL, and Neo4j databases.

## Prerequisites

- Docker and Docker Compose installed
- Git access to the LDFLK/data-backups repository
- Access to the OpenGIN project

## 1. Setting Up Configuration

### Step 1.1: Fork and Clone the Data-Backups Repository

1. **Fork the repository:**
   - Go to [LDFLK/data-backups](https://github.com/LDFLK/data-backups)
   - Click "Fork" to create your own copy

2. **Clone your fork:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/data-backups.git
   cd data-backups
   ```

3. **Create version folder structure:**
   ```bash
   # Create the version folder (replace 0.0.1 with your version)
   mkdir -p data-backups/opengin/development/{mongodb,postgres,neo4j}
   
   # The structure should look like:
    opengin
        └── development
            │   ├── mongodb
            │   │   └── opengin.tar.gz
            │   ├── neo4j
            │   │   └── neo4j.dump
            │   │   └── postgres
            │   │       └── opengin.tar.gz
            │   ├── production
            │   │   ├── mongodb
            │   │   └── neo4j
            │   └── staging
            │       ├── mongodb
            │       └── neo4j
            └── README.md

### Step 1.2: Configure Backup Environment

1. **Navigate to the Opengin project:**
   ```bash
   cd /path/to/opengin
   ```

2. **Update the backup configuration file:**
   Edit `configs/backup.env` to point to your data-backups repository:
   ```bash
   # Set the paths to your local data-backups repository
   MONGODB_BACKUP_DIR=/path/to/your/data-backups/opengin/version/0.0.1/development/mongodb
   POSTGRES_BACKUP_DIR=/path/to/your/data-backups/opengin/version/0.0.1/development/postgres
   NEO4J_BACKUP_DIR=/path/to/your/data-backups/opengin/version/0.0.1/development/neo4j
   
   # Database credentials (update as needed)
   MONGODB_USERNAME=your_mongodb_user
   MONGODB_PASSWORD=your_mongodb_password
   MONGODB_DATABASE=your_mongodb_database
   
   POSTGRES_USER=your_postgres_user
   POSTGRES_PASSWORD=your_postgres_password
   POSTGRES_DATABASE=your_postgres_database
   
   ENVIRONMENT=development
   ```

## 2. Creating Database Backups

### Step 2.1: Start Required Services

Ensure your database containers are running:

```bash
# Start all services
docker-compose up -d

# Or start individual services
docker-compose up -d mongodb postgres neo4j
```

### Step 2.2: Create Backups

1. **Backup MongoDB:**
   ```bash
   cd deployment/development
   ./init.sh backup_mongodb
   ```
   This creates `opengin.tar.gz` in your MongoDB backup directory.

2. **Backup PostgreSQL:**
   ```bash
   ./init.sh backup_postgres
   ```
   This creates `opengin.tar.gz` in your PostgreSQL backup directory.

3. **Backup Neo4j:**
   ```bash
   ./init.sh backup_neo4j
   ```
   This creates `neo4j.dump` in your Neo4j backup directory.

## 3. Creating GitHub Releases

### Step 3.1: Commit and Push Changes

1. **Navigate to your data-backups repository:**
   ```bash
   cd /path/to/your/data-backups
   ```

2. **Add and commit your backup files:**
   ```bash
   git add opengin/version/0.0.1/development/
   git commit -m "Add database backups for version 0.0.1"
   ```

3. **Create a release branch:**
   ```bash
   git checkout -b release-0.0.1
   git push origin release-0.0.1
   ```

### Step 3.2: Create a Pull Request

1. Go to your forked repository on GitHub
2. Click "Compare & pull request" for the `release-0.0.1` branch
3. Create a PR to merge into the main branch of the original repository
4. Wait for review and approval

### Step 3.3: Create a GitHub Release

1. **After PR is merged, create a release:**
   - Go to the original [LDFLK/data-backups](https://github.com/LDFLK/data-backups) repository
   - Click "Releases" → "Create a new release"
   - Tag version: `0.0.1`
   - Release title: `Version 0.0.1 - Database Backups`
   - Description: Brief description of the backup contents
   - Click "Publish release"

2. **Update the latest release (optional):**
   - If this is the most recent version, also create/update a `latest` tag
   - This allows users to restore the latest version without specifying a version number

## 4. Restoring from GitHub

### Step 4.1: List Available Versions

```bash
cd deployment/development
./init.sh list_github_versions
```

### Step 4.2: Restore from Specific Version

```bash
# Restore from a specific version
./init.sh restore_from_github 0.0.1
```

### Step 4.3: Verify Restore

After restoration, verify that your databases contain the expected data.
```

