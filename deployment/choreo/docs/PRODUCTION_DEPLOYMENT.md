# Production Deployment Guide

This guide describes how to deploy the OpenGIN services in a production-like environment using the Choreo-optimized Docker configuration (`docker-compose-prod.yml`).

## Overview

The production setup differs from development in the following ways:
*   **Databases**: Connects to remote managed databases (Supabase for PostgreSQL, MongoDB Atlas, Neo4j Aura) instead of valid local containers.
*   **Images**: Uses `.choreo` Dockerfiles which are optimized for cloud deployment (e.g., using specific build contexts and startup scripts).
*   **Configuration**: All secrets and connection details are managed via environment variables.

## Prerequisites

*   Docker Desktop (or Docker Engine + Compose plugin) installed.
*   Access to the remote database instances.

## Configuration (.env)

Create a `.env` file in the project root to store your database credentials. 

> **Important**: This file contains secrets. Ensure it is added to your `.gitignore`.

### Required Variables

Add the following variables to your `.env` file, replacing the values with your actual credentials.

#### PostgreSQL (Supabase)
```bash
POSTGRES_HOST=your-db-host.supabase.co
POSTGRES_PORT=5432
POSTGRES_USER=your-db-user
POSTGRES_PASSWORD=your-db-password
POSTGRES_DB=postgres
# SSL Mode is usually 'require' for managed databases
POSTGRES_SSL_MODE=require
```
![Supabase Connection URL Settings](supabase-url.png)
*Find your connection details in Project Settings > Database > Connection Parameters*

#### MongoDB (Atlas)
```bash
MONGO_URI=mongodb+srv://user:pass@cluster.mongodb.net/
MONGO_DB_NAME=opengin
MONGO_ADMIN_USER=your-admin-user
MONGO_ADMIN_PASSWORD=your-admin-pass
```

#### Neo4j (Aura)
```bash
NEO4J_URI=neo4j+s://your-instance.databases.neo4j.io
NEO4J_USER=neo4j
NEO4J_PASSWORD=your-neo4j-password
NEO4J_DATABASE=neo4j
```

## Running the Application

To start the application using the production configuration:

```bash
docker compose -f docker-compose-prod.yml --env-file .env up --build
```

### Detached Mode
To run in the background:
```bash
docker compose -f docker-compose-prod.yml --env-file .env up -d --build
```

### Stopping the Application
```bash
docker compose -f docker-compose-prod.yml down
```

## Troubleshooting

### Environment Variable Collisions (Auth Failures)
If you encounter authentication failures (e.g., `password authentication failed for user "postgres"`), check if you have local shell environment variables set that conflict with your `.env` file.

Docker Compose prioritizes shell variables over the `.env` file. Common conflicts include:
*   `POSTGRES_USER`
*   `POSTGRES_PASSWORD`

**Solution**: Unset the conflicting variables in your shell before running docker-compose:
```bash
unset POSTGRES_USER POSTGRES_PASSWORD
docker compose -f docker-compose-prod.yml --env-file .env up --build
```

### Hostname Resolution
If services cannot communicate (e.g., `Could not resolve host: core-choreo`), ensure that `docker-compose-prod.yml` explicitly overrides the `BAL_CONFIG_VAR_CORESERVICEURL` environment variable for dependent services (ingestion, read).

```yaml
environment:
  - BAL_CONFIG_VAR_CORESERVICEURL=http://core:50051
```
This is already configured in the provided `docker-compose-prod.yml`.
