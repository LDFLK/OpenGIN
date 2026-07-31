# config format 1 (default for development)
$env:CORE_SERVICE_HOST="0.0.0.0"
$env:CORE_SERVICE_PORT="50051"
$env:INGESTION_SERVICE_HOST="0.0.0.0"
$env:INGESTION_SERVICE_PORT="8080"
$env:CORE_SERVICE_URL="http://localhost:50051"

# config format 2 (default for deployment)

$env:CORE_SERVICE_URL="http://0.0.0.0:50051"
$env:INGESTION_SERVICE_URL="http://0.0.0.0:8080"
