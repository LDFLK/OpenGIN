# core-api

## TODOs

## Note

⚠️ **Warning**  
Please do not commit generated protobuf files, please generate them at build time..

## Pre-requisites

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go get google.golang.org/grpc
go get go.mongodb.org/mongo-driver/mongo
go get github.com/lib/pq  # PostgreSQL driver
```

## Development



### Generate Go Code (Developer)

For LINUX & macOS
```bash
cd core-api
cp env.template .env
# after updating the required fields to be added to the environment
# (you can find the example env configurations here)
source .env
go test -v ./...
./core-service


## Go Module Setup

Initialize a Go module. (For LINUX, macOS & Windows)

```bash
go mod init github.com/LDFLK/nexoan/opengin/core-api
```

Generating the protobuf stubs for Go-lang
This will be used to write the data pipeline from CORE server to the API server
via Grpc. 

```bash
protoc --go_out=. --go-grpc_out=. --proto_path=protos protos/types_v1.proto
```

### Build

For LINUX & macOS
```bash
go build ./...
go build -o core-service cmd/server/service.go cmd/server/utils.go
```

For windows (make sure you open the **Powershell CLI**)
```bash
go build ./...
go build -o core-service.exe cmd/server/service.go cmd/server/utils.go
```

## Usage

### Run Service

For LINUX & macOS
```bash
./core-service
```

For windows (make sure you open the **Powershell CLI**)
```bash
./core-service.exe
```

The service will be running in port `50051` and it is hard coded. This needs to be configurable. 

#### Run with Docker

`opengin/core-api/docker/Dockerfile` refers to just running the

Make sure to create a network for this work since we need every other service to be accessible hence
we place them in the same network. 

```bash
docker network create core-network
```

```bash
docker build -t core-service -f opengin/core-api/docker/Dockerfile .
```

```bash
docker run -d \
  --name core-service \
  --network core-network \
  -p 50051:50051 \
  -e NEO4J_URI=bolt://neo4j-local:7687 \
  -e NEO4J_USER=${NEO4J_USER} \
  -e NEO4J_PASSWORD=${NEO4J_PASSWORD} \
  -e MONGO_URI=${MONGO_URI} \
  -e POSTGRES_HOST=${POSTGRES_HOST} \
  -e POSTGRES_PORT=${POSTGRES_PORT} \
  -e POSTGRES_USER=${POSTGRES_USER} \
  -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} \
  -e POSTGRES_DB=${POSTGRES_DB} \
  core-service
```

#### Validate 

```bash
brew install grpcurl
```

```bash
grpcurl -plaintext localhost:50051 list
```

Output

```bash
core.core
grpc.reflection.v1.ServerReflection
grpc.reflection.v1alpha.ServerReflection
```

### Run Tests: Mode 1 (Independent Environments and Services)

We assume the Mongodb, Neo4j, and PostgreSQL are provided as services or they exist in the same network. 

> **Note**: For detailed PostgreSQL testing instructions including coverage reports and specific test selection, see [DEVELOPMENT.md](DEVELOPMENT.md).

```bash
# Build the test image
docker build -t core-service-test-v1 -f Dockerfile.test.v1 .

# Run the tests
docker run --rm \
  --network core-network \
  -e NEO4J_URI=bolt://neo4j-local:7687 \
  -e NEO4J_USER=${NEO4J_USER} \
  -e NEO4J_PASSWORD=${NEO4J_PASSWORD} \
  -e MONGO_URI=${MONGO_URI} \
  -e POSTGRES_TEST_DB_URI=${POSTGRES_TEST_DB_URI} \
  core-service-test-v1
```

### Run Tests: Mode 2 (Choreo and CI/CD)

MongoDB, Neo4j, and PostgreSQL are running in the same container. 

```bash
docker build -t core-service-test-standalone -f Dockerfile.test .


# Run the tests

docker run --rm core-service-test-standalone
```

