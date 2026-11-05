# Read API

## Prerequisites

⚠️ **Warning**  
You should have the Ballerina version of **2201.11.0** in your device. You can download it by following link
```bash
https://ballerina.io/downloads/archived/#swan-lake-archived-versions
```

## Generate OpenAPI Service

This will generate the endpoints for the Read API server using the OpenAPI specification.
The OpenAPI specification is the base for public API for Read API.

> 💡 Note: Always make sure the contract has the expected endpoints and request params
> before working on the code. The generated endpoints should not be editable at all.
> Maybe the only changes that can be done is adding error handlers, but request and response
> must be defined in the contract.
> Note that this will create the read_api_service.bal file and types.bal file

```bash
bal openapi -i ../contracts/rest/read_api.yaml --mode service
```

## Generate gRPC Stubs

The client stub generated here will be sending and receiving values via Grpc.
This will send requests to the corresponding CORE server endpoint.
This command will generate the types_v1_pb.bal file

```bash
bal grpc --mode client --input ../core-api/protos/types_v1.proto --output .
```

> 💡 **Note**  
> At the generation make sure to remove any sample code generated to show how to use the API. Because that might add an unnecessary main file.

## Environment Variables

For LINUX & macOS
```bash
cd opengin/read-api
cp env.template .env
# update the required fields to set the environment variables
source .env
```

For Windows (PowerShell)
```powershell
cd opengin/read-api
Copy-Item env.template.ps1 .env.ps1
# after updating the required fields to be added to the environment
. .\.env.ps1
```

Example .env configuration
```bash
# For LINUX & macOS
export CORE_SERVICE_HOST=localhost
export CORE_SERVICE_PORT=50051
export READ_SERVICE_HOST=localhost
export READ_SERVICE_PORT=8081

# For Windows (paste on Powershell CLI)
$env:CORE_SERVICE_HOST="localhost"
$env:CORE_SERVICE_PORT="50051"
$env:READ_SERVICE_HOST="localhost"
$env:READ_SERVICE_PORT="8081"
```

## Run Tests

Make sure the CORE server is running.
(LINUX & macOS: `cd opengin/core-api && ./core-service`; Windows: `cd opengin/core-api; .\core-service.exe`)

```bash
# Run all tests in the current package
bal test

# Run tests with verbose output
bal test --test-report

# Run a specific test file
bal test tests/read_api_service_test.bal

# Run a specific test function
bal test --tests testEntityAttributeRetrieval

# Run tests and generate a coverage report
bal test --code-coverage
```

## Run Service

```bash
cd opengin/read-api
bal run
```
