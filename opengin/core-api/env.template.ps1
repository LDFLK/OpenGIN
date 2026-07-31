$env:MONGO_URI=""
$env:MONGO_DB_NAME=""
$env:MONGO_COLLECTION=""

## Uncomment the following for development

# $env:NEO4J_URI=$env:NEO4J_DEVELOPMENT_DB_URI
# $env:NEO4J_USER=$env:NEO4J_DEVELOPMENT_USERNAME
# $env:NEO4J_PASSWORD=$env:NEO4J_DEVELOPMENT_PASSWORD

## If commented, uncomment the following will be used for testing

# $env:NEO4J_URI=$env:NEO4J_TESTING_DB_URI
# $env:NEO4J_USER=$env:NEO4J_TESTING_USERNAME
# $env:NEO4J_PASSWORD=$env:NEO4J_TESTING_PASSWORD

## PostgreSQL configuration

$env:POSTGRES_HOST="localhost"
$env:POSTGRES_PORT="5432"
$env:POSTGRES_USER="postgres"
$env:POSTGRES_PASSWORD="postgres"
$env:POSTGRES_DB="opengin"
$env:POSTGRES_SSL_MODE="disable"
$env:POSTGRES_TEST_DB_URI="postgresql://postgres:postgres@localhost:5432/opengin?sslmode=disable"

$env:CORE_SERVICE_HOST="localhost"
$env:CORE_SERVICE_PORT="50051"
