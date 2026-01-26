# Neo4j Docker for Choreo

This directory contains a custom Docker configuration for deploying Neo4j on the Choreo platform.

## Overview

Unlike standard Neo4j images, this image uses a **"Baked-In" Data Strategy**. Instead of downloading data when the container starts (which is slow and can time out), we download and pre-load the dataset during the **Docker build process**.

This means:
1.  **Fast Startup**: The database is ready immediately when the container launches.
2.  **Stateless Deployment**: You can destroy and recreate the container, and it always starts with fresh, correct data.
3.  **Choreo Compliance**: Runs as a non-root user (`choreo`, UID 10014) to meet security requirements.

## Dockerfile Flow

The logic proceeds in three main stages:

### 1. Setup & Dependencies
- Starts from `ubuntu:22.04`.
- Installs Java 17 (required for Neo4j 5) and Neo4j 5.12.0.
- Creates the mandatory `choreo` user (UID 10014) required by the platform.

### 2. Build-Time Data Ingestion
This is the core customization. During the `docker build` phase:
1.  Accepts build arguments (`ARG`) for the GitHub repository, version, and environment.
2.  Downloads the specified data backup (zip) from GitHub.
3.  Unzips and restores the database using `neo4j-admin database load`.
4.  Sets the correct owner (`choreo`) for all data files (`/neo4j_data`, `/neo4j_logs`, etc.).

**Build Arguments:**
You can customize the data source by passing these arguments during build:
- `GITHUB_BACKUP_REPO`: The repository containing backups (default: `LDFLK/data-backups`).
- `BACKUP_VERSION`: The release tag version (default: `0.0.4`).
- `BACKUP_ENVIRONMENT`: The environment folder name (default: `development`).

### 3. Runtime Configuration
- **Networking**: Configures Neo4j to listen on `0.0.0.0` so it's accessible externally.
- **Entrypoint (`docker-entrypoint.sh`)**: A custom startup script that:
    - Checks the `NEO4J_AUTH` environment variable.
    - Sets the initial database password using the Neo4j CLI if provided (e.g., `neo4j/my-password`).
    - Starts the Neo4j console.

## Verification

You can verify the data integrity using the provided Python script `verify_neo4j.py`.

### Prerequisites

1.  **Python 3** must be installed.
2.  **Create and activate a virtual environment**:
    ```bash
    python3 -m venv venv
    source venv/bin/activate
    ```
3.  **Install the Neo4j driver**:
    ```bash
    pip install neo4j
    ```
4.  The Neo4j container must be running (e.g., via `docker-compose up`).

### Running the Test

Run the script directly:

```bash
python3 verify_neo4j.py
```

You can customize connection details via environment variables if needed:
- `NEO4J_URI` (default: `bolt://localhost:7687`)
- `NEO4J_USER` (default: `neo4j`)
- `NEO4J_PASSWORD` (default: `neo4j123`)

