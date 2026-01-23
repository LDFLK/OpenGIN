# DEPLOYMENT

## CORE SERVICE

We deploy the CORE service as a gRPC service and we have to add a few environmental variables 
and a few file mounts to make things work. 

### Environmental Variables 

1. MONGO_URI
2. MONGO_DB_NAME
3. MONGO_COLLECTION
4. NEO4J_URI
5. NEO4J_USER
6. NEO4J_PASSWORD
7. CORE_SERVICE_HOST
8. CORE_SERVICE_PORT
9. POSTGRES_USER
10. POSTGRES_HOST
11. POSTGRES_PORT
12. POSTGRES_DB
13. POSTGRES_SSL_MODE
14. POSTGRES_PASSWORD

### File Mounts

| # | Mount Name | Type | Mount Path | Description |
|---|------------|------|------------|-------------|
| 1 | core-go-build-mnt | Empty Directory (In-Memory) | /home/choreouser/.cache | Go build cache directory |
| 2 | default-tmp-emptydir | Empty Directory (In-Memory) | /tmp | Temporary files directory |
| 3 | mnt-go-core-dir | Empty Directory (In-Memory) | /go | Go core directory |

### Choreo Configs

When deploying the CORE service one thing to note is that GRPC services are not exposed through the Gateway in Choreo. So we have to choose the `PROJECT_URL` from `Manage`->`Overview` tabs in Choreo
console. Make sure to extract that URL and use it as the `coreServiceURL` config in both `Ingestion` API and
`Read` API services.

## Choreo Configurations

### Configuration Groups

Choreo provides a feature called `Configuration Groups` to manage configurations for different environments. This feature allows you to add a list of environment variables and file mounts for each environment. During deployment, you simply link the configuration group based on the environment.

### Components

Choreo provides a feature called `Components` to manage different services. This feature allows you to add a list of services for each environment. During deployment, you simply link the component based on the environment.

#### Core API

For `Core API` we use the following component of type service:

```yaml
schemaVersion: 1.2
endpoints:
  - name: core-api
    displayName: CORE API
    service:
      basePath: /
      port: 50051
    type: GRPC
    networkVisibilities: 
      - Project
```

In Choreo/OpenChoreo there is a Docker-based build pack.
Ensure the following build configurations:
- Dockerfile: `opengin/core-api/docker/Dockerfile.choreo`
- Build Context: `opengin/core-api`

#### Ingestion API

For `Ingestion API` we use the following component of type service:

```yaml
schemaVersion: 1.2
endpoints:
  - name: ingestion-ep
    displayName: OpenGIN Ingestion API
    service:
      basePath: /
      port: 8080
    type: REST
    networkVisibilities: 
      - Public
```

In Choreo/OpenChoreo there is a Ballerina-based build pack. 
Also Ballerina supports configurations which is more or less like environment variables, 
but not exactly. We use them to provide parameters to the Ballerina service. 

It requires the Core-API URL. For this, use the `Project URL` from the `Manage` -> `Overview` tab in the Choreo console. This URL refers to the internal URL of the service, which communicates directly with the pod.

And the `Build Context` we provide is the `opengin/ingestion-api` directory.

We don't host this service in Choreo as this is not going to be a publicly exposed API 
at the moment, but we need to come up with a token-based authentication and provide 
additional security via Choreo/OpenChoreo. 

#### Read API

For `Read API` we use the following component of type service:

```yaml
schemaVersion: 1.2
endpoints:
  - name: read-ep
    displayName: OpenGIN Read API
    service:
      basePath: /
      port: 8081
    type: REST
    networkVisibilities: 
      - Public
```

In Choreo/OpenChoreo there is a Ballerina-based build pack. 
Also Ballerina supports configurations which is more or less like environment variables, 
but not exactly. We use them to provide parameters to the Ballerina service. 

It requires the Core-API URL. For this, use the `Project URL` from the `Manage` -> `Overview` tab in the Choreo console. This URL refers to the internal URL of the service, which communicates directly with the pod.

And the `Build Context` we provide is the `opengin/read-api` directory.

When deploying, edit the `Configurables` as follows:
- coreServiceUrl: `the project url of your core service`
- readServiceHost: `0.0.0.0`
- readServicePort: `8081`

> **Warning:** Make sure to disable the auth and security features for the moment as we are keeping APIs open. This configuration is not suitable for production environments.
