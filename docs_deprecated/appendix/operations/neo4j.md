# Neo4j Migration Guide: Docker Container to Aura

This guide provides step-by-step instructions for migrating a Neo4j database from a Docker container to Neo4j Aura.

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
- Access to your Neo4j Docker container
- Neo4j Aura account and database instance
- Neo4j Aura connection URI, username, and password

## Migration Steps

### 1. Stop the Neo4j Container

Make sure the neo4j container is stopped before creating the dump.

### 2. Identify where the data is stored in the container

Find the directory path that's mapped to /data in your neo4j container, this is where the data is stored. 

Run `docker inspect <container_name>` and look for the `Source` path in the Mounts section.

This will look something like the following: 

If you are using an standalone Docker container built on top of an exisitng neo4j image with volumes named
`/var/lib/docker/volumes/neo4j_data/_data`

Or if you are using the `docker-compose.yml` to build you will see the volume as `ldfarchitecture_neo4j_data`.

Set this as `NEO4J_CONTAINER_DATA_VOLUME`

```bash
export NEO4J_CONTAINER_DATA_VOLUME=<path-to-source-path-in-docker>
```

Also we need to figure out the path where this volume is mounted in the container.
That is defined in the `Dockerfile` as `NEO4J_dbms_directories_data`.

```bash
export NEO4J_DBMS_DIRECTORIES_DATA=<path-configured-in-docker>
```

So you can actually see if the data has been written to this volume via a docker container since this file system is not accessible outside a docker environment and you may have to mount it manually to an image and check.

```bash
docker run --rm -it \
  --volume ${NEO4J_CONTAINER_DATA_VOLUME}:/${NEO4J_DBMS_DIRECTORIES_DATA} \
  alpine:latest \
  sh
```

Within this you can check for the data

### 3. Create a Local Dump Folder

Create a folder on your local machine to store the dump file.

### 4. Create a Database Dump

Run the following command to create a dump from your Neo4j Docker container:  

What happen below is we are going to run a temporary container (that's why there is `rm`)
and mount the volume which has neo4j data which is `NEO4J_CONTAINER_DATA_VOLUME` volume and it
is mounted to the `/data` folder of this temporary container.

And the following command will create a dump on the `/backups` folder. 

```bash
neo4j-admin database dump neo4j --to-path=/backups
```

And since we `--volume=/Users/your_username/Documents/neo4j_dump:/backups ` since that the 
host system can also access the backup file.

```bash
docker run --rm \
--volume=${NEO4J_CONTAINER_DATA_VOLUME}:/data \
--volume=${NEO4J_BACKUP_DIR}:/backups \
neo4j/neo4j-admin:latest \
neo4j-admin database dump neo4j --to-path=/backups
```

### 4. Verify the Database Dump was created

Navigate to the local folder specified previously and check that a dump file has been created inside.

### 5. Upload the Dump to the Neo4j Docker (Local)

Stop the container

```bash
docker compose down neo4j
```

```bash
docker run --interactive --tty --rm \
    --volume=${NEO4J_CONTAINER_DATA_VOLUME}:/data \
    --volume=${NEO4J_BACKUP_DIR}:/backups \
    neo4j/neo4j-admin \
neo4j-admin database load neo4j --from-path=/backups --overwrite-destination=true
```

### 6. Upload the Dump to Neo4j Aura

Run the following command to upload the dump to your Neo4j Aura instance:

```bash
docker run --rm \
--volume=/Users/your_username/Documents/neo4j_dump:/dump \
neo4j/neo4j-admin:latest \
neo4j-admin database upload neo4j \
--from-path=/dump \
--to-uri=<neo4j_uri> \
--to-user=<neo4j_user> \
--to-password=<neo4j_password> \
--overwrite-destination=true
```

**Important:** Replace the following placeholders with your actual values:
- `/Users/your_username/Documents/neo4j_dump` with the correct path to your local folder with the dump file.
- `<neo4j_uri>`,`<neo4j_user>`,`<neo4j_password>` with your actual aura db credentials

### 5. Verify the Migration

Connect to your Neo4j Aura instance and verify that the new data has been inserted.
