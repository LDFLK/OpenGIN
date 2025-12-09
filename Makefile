.PHONY: help env build build-core build-ingestion build-read test test-core test-ingestion test-read \
	run-core run-ingestion run-read stop-core stop-ingestion stop-read \
	e2e e2e-venv e2e-ingestion e2e-read e2e-docker \
	infra-up infra-down services-up services-down \
	up up-core up-ingestion up-read \
	down down-core down-ingestion down-read down-all \
	logs clean-pre clean-post dev \
	infra-up-choreo infra-down-choreo services-up-choreo services-down-choreo \
	up-choreo up-core-choreo up-ingestion-choreo up-read-choreo \
	down-choreo down-core-choreo down-ingestion-choreo down-read-choreo down-all-choreo \
	logs-choreo clean-pre-choreo clean-post-choreo \
	e2e-docker-choreo e2e-choreo e2e-ingestion-choreo e2e-read-choreo

# Select docker compose command. Override with: make COMPOSE="docker compose"
COMPOSE ?= docker-compose
COMPOSE_CHOREO ?= docker-compose -f docker-compose-choreo.yml

# Paths (updated after refactor)
CORE_DIR := opengin/core-api
INGESTION_DIR := opengin/ingestion-api
READ_DIR := opengin/read-api
E2E_DIR := opengin/tests/e2e
DEPLOY_DEV := deployment/development

.DEFAULT_GOAL := help

help:
	@echo "Nexoan — Make targets"
	@echo "------------------------------------------------------------"
	@echo "env                 Copy env.template to .env for all services"
	@echo "build               Build all APIs (Core, Ingestion, Read)"
	@echo "build-core          Build Core API"
	@echo "build-ingestion     Build Ingestion API"
	@echo "build-read          Build Read API"
	@echo "test                Run all tests (Core, Ingestion, Read)"
	@echo "test-core           Run tests for Core API"
	@echo "test-ingestion      Run tests for Ingestion API"
	@echo "test-read           Run tests for Read API"
	@echo "run-core            Run Core API service directly (without docker, sources .env, runs in background)"
	@echo "run-ingestion       Run Ingestion API service directly (without docker, sources .env, runs in background)"
	@echo "run-read            Run Read API service directly (without docker, sources .env, runs in background)"
	@echo "stop-core           Stop Core API service (started with make run-core)"
	@echo "stop-ingestion      Stop Ingestion API service (started with make run-ingestion)"
	@echo "stop-read           Stop Read API service (started with make run-read)"
	@# echo "coverage            Run coverage for all APIs"  # Disabled - requires testing
	@# echo "coverage-core       Run coverage for Core API"  # Disabled - requires testing
	@# echo "coverage-ingestion  Run coverage for Ingestion API"  # Disabled - requires testing
	@# echo "coverage-read       Run coverage for Read API"  # Disabled - requires testing
	@# echo "fmt                 Format Core API code (gofumpt + golines -m 120)"  # Disabled - requires testing
	@# echo "fmt-core            Same as 'fmt' (Core API only)"  # Disabled - requires testing
	@# echo "lint                Lint Core API code (golangci-lint)"  # Disabled - requires testing
	@# echo "lint-core           Same as 'lint' (Core API only)"  # Disabled - requires testing
	@# echo "tools-core          Install Go dev tools for Core API: gofumpt, golines, golangci-lint"  # Disabled - requires testing
	@# echo "hooks-install       Install git pre-commit hooks"  # Disabled - requires testing
	@echo "e2e-venv            Set up Python virtual environment for E2E tests"
	@echo "e2e                 Run all E2E tests locally (requires services running)"
	@echo "e2e-ingestion       Run E2E tests for Ingestion API (basic_core_tests.py)"
	@echo "e2e-read            Run E2E tests for Read API (basic_read_tests.py)"
	@echo "e2e-docker          Run E2E tests in docker-compose 'e2e' service"
	@echo "infra-up            Start databases (MongoDB, Neo4j, Postgres)"
	@echo "infra-down          Stop databases"
	@echo "services-up         Start services (core, ingestion, read)"
	@echo "services-down       Stop services"
	@echo "up                  Start full stack (infra + services)"
	@echo "up-core             Start Core API service (assumes infra is running)"
	@echo "up-ingestion        Start Ingestion API service (assumes infra is running)"
	@echo "up-read             Start Read API service (assumes infra is running)"
	@echo "down                Stop stack (keeps volumes)"
	@echo "down-core           Stop Core API service"
	@echo "down-ingestion      Stop Ingestion API service"
	@echo "down-read           Stop Read API service"
	@echo "down-all            Stop stack and remove volumes"
	@echo "logs                Tail logs for main services"
	@echo "clean-pre           Clean databases (pre) using cleanup profile"
	@echo "clean-post          Clean databases (post) using cleanup profile"
	@# echo "backup-<db>         Backup mongodb | postgres | neo4j"  # Disabled - requires testing
	@# echo "restore-<db>        Restore mongodb | postgres | neo4j"  # Disabled - requires testing
	@echo "dev                 One command: clean, build, start full stack, ready for development"
	@echo ""
	@echo "Choreo Environment Targets (docker-compose-choreo.yml):"
	@echo "------------------------------------------------------------"
	@echo "infra-up-choreo     Start databases (MongoDB, Neo4j, Postgres) - Choreo"
	@echo "infra-down-choreo   Stop databases - Choreo"
	@echo "services-up-choreo  Start services (core-choreo, ingestion-choreo, read-choreo)"
	@echo "services-down-choreo Stop services - Choreo"
	@echo "up-choreo           Start full stack (infra + services) - Choreo"
	@echo "up-core-choreo      Start Core API service - Choreo (assumes infra is running)"
	@echo "up-ingestion-choreo Start Ingestion API service - Choreo (assumes infra is running)"
	@echo "up-read-choreo      Start Read API service - Choreo (assumes infra is running)"
	@echo "down-choreo         Stop stack (keeps volumes) - Choreo"
	@echo "down-core-choreo    Stop Core API service - Choreo"
	@echo "down-ingestion-choreo Stop Ingestion API service - Choreo"
	@echo "down-read-choreo    Stop Read API service - Choreo"
	@echo "down-all-choreo     Stop stack and remove volumes - Choreo"
	@echo "logs-choreo         Tail logs for main services - Choreo"
	@echo "clean-pre-choreo    Clean databases (pre) using cleanup profile - Choreo"
	@echo "clean-post-choreo   Clean databases (post) using cleanup profile - Choreo"
	@echo "e2e-docker-choreo   Run E2E tests in docker-compose-choreo 'e2e-choreo' service"
	@echo "e2e-choreo          Run all E2E tests locally against Choreo services"
	@echo "e2e-ingestion-choreo Run E2E tests for Ingestion API against Choreo services"
	@echo "e2e-read-choreo     Run E2E tests for Read API against Choreo services"
	@echo "------------------------------------------------------------"
	@echo "Tip: override compose command with COMPOSE=\"docker compose\" if needed"

env:
	@echo "Setting up .env files from templates..."
	@if [ ! -f $(CORE_DIR)/.env ]; then \
		cp $(CORE_DIR)/env.template $(CORE_DIR)/.env && \
		echo "✓ Created $(CORE_DIR)/.env"; \
	else \
		echo "⚠ $(CORE_DIR)/.env already exists, skipping"; \
	fi
	@if [ ! -f $(INGESTION_DIR)/.env ]; then \
		cp $(INGESTION_DIR)/env.template $(INGESTION_DIR)/.env && \
		echo "✓ Created $(INGESTION_DIR)/.env"; \
	else \
		echo "⚠ $(INGESTION_DIR)/.env already exists, skipping"; \
	fi
	@if [ ! -f $(READ_DIR)/.env ]; then \
		cp $(READ_DIR)/env.template $(READ_DIR)/.env && \
		echo "✓ Created $(READ_DIR)/.env"; \
	else \
		echo "⚠ $(READ_DIR)/.env already exists, skipping"; \
	fi
	@echo ""
	@echo "✅ Environment files created. Please edit the .env files as required:"
	@echo "   - $(CORE_DIR)/.env"
	@echo "   - $(INGESTION_DIR)/.env"
	@echo "   - $(READ_DIR)/.env"

build: build-core build-ingestion build-read

build-core:
	@echo "Building Core API"
	@cd $(CORE_DIR) && \
		if [ -f .env ]; then \
			set -a && source .env && set +a; \
		fi && \
		go build ./... && go build -o core-service ./cmd/server

build-ingestion:
	@echo "Building Ingestion API"
	@cd $(INGESTION_DIR) && \
		if [ -f .env ]; then \
			set -a && source .env && set +a; \
		fi && \
		bal build

build-read:
	@echo "Building Read API"
	@cd $(READ_DIR) && \
		if [ -f .env ]; then \
			set -a && source .env && set +a; \
		fi && \
		bal build

test: test-core test-ingestion test-read

test-core:
	@echo "Running tests for Core API"
	@cd $(CORE_DIR) && \
		if [ -f .env ]; then \
			set -a && source .env && set +a; \
		fi && \
		go test -v -count=1 ./...

test-ingestion:
	@echo "Running tests for Ingestion API"
	@cd $(INGESTION_DIR) && \
		if [ -f .env ]; then \
			set -a && source .env && set +a; \
		fi && \
		bal test

test-read:
	@echo "Running tests for Read API"
	@cd $(READ_DIR) && \
		if [ -f .env ]; then \
			set -a && source .env && set +a; \
		fi && \
		bal test

run-core:
	@echo "Running Core API service directly (without docker, in background)"
	@cd $(CORE_DIR) && \
		if [ -f .env ]; then \
			set -a && source .env && set +a; \
		fi && \
		(go run ./cmd/server > /tmp/core-api.log 2>&1 & echo $$! > /tmp/core-api.pid) && \
		echo "Core API started in background (PID: $$(cat /tmp/core-api.pid))"
	@echo "Logs: tail -f /tmp/core-api.log"
	@echo "Stop: make stop-core"

run-ingestion:
	@echo "Running Ingestion API service directly (without docker, in background)"
	@cd $(INGESTION_DIR) && \
		if [ -f .env ]; then \
			set -a && source .env && set +a; \
		fi && \
		(bal run > /tmp/ingestion-api.log 2>&1 & echo $$! > /tmp/ingestion-api.pid) && \
		echo "Ingestion API started in background (PID: $$(cat /tmp/ingestion-api.pid))"
	@echo "Logs: tail -f /tmp/ingestion-api.log"
	@echo "Stop: make stop-ingestion"

run-read:
	@echo "Running Read API service directly (without docker, in background)"
	@cd $(READ_DIR) && \
		if [ -f .env ]; then \
			set -a && source .env && set +a; \
		fi && \
		(bal run > /tmp/read-api.log 2>&1 & echo $$! > /tmp/read-api.pid) && \
		echo "Read API started in background (PID: $$(cat /tmp/read-api.pid))"
	@echo "Logs: tail -f /tmp/read-api.log"
	@echo "Stop: make stop-read"

stop-core:
	@if [ -f /tmp/core-api.pid ]; then \
		PID=$$(cat /tmp/core-api.pid); \
		if ps -p $$PID > /dev/null 2>&1; then \
			kill $$PID && echo "Stopped Core API (PID: $$PID)"; \
			rm -f /tmp/core-api.pid; \
		else \
			echo "Core API process (PID: $$PID) not running"; \
			rm -f /tmp/core-api.pid; \
		fi; \
	else \
		echo "Core API PID file not found. Service may not be running."; \
	fi

stop-ingestion:
	@if [ -f /tmp/ingestion-api.pid ]; then \
		PID=$$(cat /tmp/ingestion-api.pid); \
		if ps -p $$PID > /dev/null 2>&1; then \
			kill $$PID && echo "Stopped Ingestion API (PID: $$PID)"; \
			rm -f /tmp/ingestion-api.pid; \
		else \
			echo "Ingestion API process (PID: $$PID) not running"; \
			rm -f /tmp/ingestion-api.pid; \
		fi; \
	else \
		echo "Ingestion API PID file not found. Service may not be running."; \
	fi

stop-read:
	@if [ -f /tmp/read-api.pid ]; then \
		PID=$$(cat /tmp/read-api.pid); \
		if ps -p $$PID > /dev/null 2>&1; then \
			kill $$PID && echo "Stopped Read API (PID: $$PID)"; \
			rm -f /tmp/read-api.pid; \
		else \
			echo "Read API process (PID: $$PID) not running"; \
			rm -f /tmp/read-api.pid; \
		fi; \
	else \
		echo "Read API PID file not found. Service may not be running."; \
	fi

# Coverage targets disabled - requires testing in separate PR
# coverage: coverage-core coverage-ingestion coverage-read
#
# coverage-core:
# 	@echo "Running coverage for Core API"
# 	@cd $(CORE_DIR) && go test -coverprofile=coverage.out ./...
# 	@cd $(CORE_DIR) && go tool cover -func=coverage.out | tail -n 1 || true
# 	@cd $(CORE_DIR) && go tool cover -html=coverage.out -o coverage.html
# 	@echo "Core API coverage HTML report: $(CORE_DIR)/coverage.html"
#
# coverage-ingestion:
# 	@echo "Running coverage for Ingestion API"
# 	@cd $(INGESTION_DIR) && bal test --code-coverage
#
# coverage-read:
# 	@echo "Running coverage for Read API"
# 	@cd $(READ_DIR) && bal test --code-coverage

e2e-venv:
	@echo "Setting up Python virtual environment for E2E tests..."
	@cd $(E2E_DIR) && \
		if [ ! -d .venv ]; then \
			python3 -m venv .venv && \
			echo "✓ Created virtual environment"; \
		else \
			echo "✓ Virtual environment already exists"; \
		fi
	@cd $(E2E_DIR) && \
		.venv/bin/pip install --quiet --upgrade pip && \
		.venv/bin/pip install --quiet requests && \
		.venv/bin/pip install --quiet protobuf && \
		.venv/bin/pip install --quiet pandas && \
		echo "✓ Installed requests library"

e2e: e2e-venv e2e-ingestion e2e-read

e2e-ingestion: e2e-venv
	@echo "Running E2E tests for Ingestion API (ensure services are up: make up)"
	@cd $(E2E_DIR) && .venv/bin/python basic_core_tests.py

e2e-read: e2e-venv
	@echo "Running E2E tests for Read API (ensure services are up: make up)"
	@cd $(E2E_DIR) && .venv/bin/python basic_read_tests.py

e2e-docker:
	@echo "Running E2E tests via docker-compose (will build and run dependent services if needed)"
	@$(COMPOSE) up --build -d mongodb neo4j postgres core ingestion read
	@$(COMPOSE) up --build e2e
	@$(COMPOSE) rm -f e2e || true

infra-up:
	@echo "Starting databases (MongoDB, Neo4j, Postgres)"
	@$(COMPOSE) up -d --build mongodb neo4j postgres

infra-down:
	@echo "Stopping databases"
	@$(COMPOSE) stop mongodb neo4j postgres || true

services-up:
	@echo "Starting services (core, ingestion, read)"
	@$(COMPOSE) up -d --build core ingestion read

services-down:
	@echo "Stopping services (core, ingestion, read)"
	@$(COMPOSE) stop core ingestion read || true

up: infra-up services-up
	@echo "Full stack started."
	@echo "- Core (gRPC): localhost:50051"
	@echo "- Ingestion API:  http://localhost:8080"
	@echo "- Read API:   http://localhost:8081"

up-core:
	@echo "Starting Core API service (ensure infra is running: make infra-up)"
	@$(COMPOSE) up -d --build core
	@echo "Core API started."
	@echo "- Core (gRPC): localhost:50051"

up-ingestion:
	@echo "Starting Ingestion API service (ensure infra is running: make infra-up)"
	@$(COMPOSE) up -d --build ingestion
	@echo "Ingestion API started."
	@echo "- Ingestion API:  http://localhost:8080"

up-read:
	@echo "Starting Read API service (ensure infra is running: make infra-up)"
	@$(COMPOSE) up -d --build read
	@echo "Read API started."
	@echo "- Read API:   http://localhost:8081"

logs:
	@$(COMPOSE) logs -f core ingestion read

down:
	@echo "Stopping stack (keeping volumes)"
	@$(COMPOSE) down

down-core:
	@echo "Stopping Core API service"
	@$(COMPOSE) stop core || true

down-ingestion:
	@echo "Stopping Ingestion API service"
	@$(COMPOSE) stop ingestion || true

down-read:
	@echo "Stopping Read API service"
	@$(COMPOSE) stop read || true

down-all:
	@echo "Stopping stack and removing volumes"
	@$(COMPOSE) down -v

clean-pre:
	@echo "Cleaning databases (pre) via cleanup profile"
	@$(COMPOSE) --profile cleanup run --rm cleanup /app/cleanup.sh pre

clean-post:
	@echo "Cleaning databases (post) via cleanup profile"
	@$(COMPOSE) --profile cleanup run --rm cleanup /app/cleanup.sh post

# Backup and restore targets disabled - requires testing in separate PR
# backup-mongodb:
# 	@cd $(DEPLOY_DEV) && ./init.sh backup_mongodb
#
# backup-postgres:
# 	@cd $(DEPLOY_DEV) && ./init.sh backup_postgres
#
# backup-neo4j:
# 	@cd $(DEPLOY_DEV) && ./init.sh backup_neo4j
#
# restore-mongodb:
# 	@cd $(DEPLOY_DEV) && ./init.sh restore_mongodb
#
# restore-postgres:
# 	@cd $(DEPLOY_DEV) && ./init.sh restore_postgres
#
# restore-neo4j:
# 	@cd $(DEPLOY_DEV) && ./init.sh restore_neo4j

# Formatting & linting
# Formatting targets disabled - requires testing in separate PR
# fmt: fmt-core
#
# fmt-core:
# 	@echo "Formatting Core API code (gofumpt + golines -m 120)"
# 	@cd $(CORE_DIR) && gofumpt -w .
# 	@cd $(CORE_DIR) && golines -m 120 -w .

# Linting targets disabled - requires testing in separate PR
# lint: lint-core
#
# lint-core:
# 	@echo "Linting Core API code (golangci-lint)"
# 	@cd $(CORE_DIR) && golangci-lint run ./...

# Install dev tools locally (ensure GOPATH/bin is on your PATH)
# Tools target disabled - requires testing in separate PR
# tools-core:
# 	@echo "Installing Go dev tools for Core API: gofumpt, golines, golangci-lint"
# 	@go install mvdan.cc/gofumpt@latest
# 	@go install github.com/segmentio/golines@latest
# 	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install and activate git pre-commit hooks
# hooks-install target disabled - requires testing in separate PR
# hooks-install:
# 	@echo "Installing pre-commit and setting up git hooks"
# 	@python3 -m pip install --user pre-commit
# 	@pre-commit install

# One-shot developer bootstrap: clean -> build -> full stack up -> tail logs hint
# Note: Feel free to interrupt logs with Ctrl+C; services keep running.
dev: clean-pre build up
	@echo "\n✅ Dev environment is up and ready!"
	@echo "- Core (gRPC): localhost:50051"
	@echo "- Ingestion API:  http://localhost:8080"
	@echo "- Read API:   http://localhost:8081"
	@echo "- Tail logs:   make logs"
	@echo "- Run E2E:     make e2e or make e2e-docker"

# Choreo Environment Targets
infra-up-choreo:
	@echo "Starting databases (MongoDB, Neo4j, Postgres) - Choreo"
	@$(COMPOSE_CHOREO) up -d --build mongodb-choreo neo4j-choreo postgres-choreo

infra-down-choreo:
	@echo "Stopping databases - Choreo"
	@$(COMPOSE_CHOREO) stop mongodb-choreo neo4j-choreo postgres-choreo || true

services-up-choreo:
	@echo "Starting services (core-choreo, ingestion-choreo, read-choreo)"
	@$(COMPOSE_CHOREO) up -d --build core-choreo ingestion-choreo read-choreo

services-down-choreo:
	@echo "Stopping services - Choreo"
	@$(COMPOSE_CHOREO) stop core-choreo ingestion-choreo read-choreo || true

up-choreo: infra-up-choreo services-up-choreo
	@echo "Full stack started (Choreo)."
	@echo "- Core (gRPC): localhost:50051"
	@echo "- Ingestion API:  http://localhost:8080"
	@echo "- Read API:   http://localhost:8081"

up-core-choreo:
	@echo "Starting Core API service - Choreo (ensure infra is running: make infra-up-choreo)"
	@$(COMPOSE_CHOREO) up -d --build core-choreo
	@echo "Core API started (Choreo)."
	@echo "- Core (gRPC): localhost:50051"

up-ingestion-choreo:
	@echo "Starting Ingestion API service - Choreo (ensure infra is running: make infra-up-choreo)"
	@$(COMPOSE_CHOREO) up -d --build ingestion-choreo
	@echo "Ingestion API started (Choreo)."
	@echo "- Ingestion API:  http://localhost:8080"

up-read-choreo:
	@echo "Starting Read API service - Choreo (ensure infra is running: make infra-up-choreo)"
	@$(COMPOSE_CHOREO) up -d --build read-choreo
	@echo "Read API started (Choreo)."
	@echo "- Read API:   http://localhost:8081"

logs-choreo:
	@$(COMPOSE_CHOREO) logs -f core-choreo ingestion-choreo read-choreo

down-choreo:
	@echo "Stopping stack (keeping volumes) - Choreo"
	@$(COMPOSE_CHOREO) down

down-core-choreo:
	@echo "Stopping Core API service - Choreo"
	@$(COMPOSE_CHOREO) stop core-choreo || true

down-ingestion-choreo:
	@echo "Stopping Ingestion API service - Choreo"
	@$(COMPOSE_CHOREO) stop ingestion-choreo || true

down-read-choreo:
	@echo "Stopping Read API service - Choreo"
	@$(COMPOSE_CHOREO) stop read-choreo || true

down-all-choreo:
	@echo "Stopping stack and removing volumes - Choreo"
	@$(COMPOSE_CHOREO) down -v

clean-pre-choreo:
	@echo "Cleaning databases (pre) via cleanup profile - Choreo"
	@$(COMPOSE_CHOREO) --profile cleanup run --rm cleanup-choreo /app/cleanup-choreo.sh pre

clean-post-choreo:
	@echo "Cleaning databases (post) via cleanup profile - Choreo"
	@$(COMPOSE_CHOREO) --profile cleanup run --rm cleanup-choreo /app/cleanup-choreo.sh post

e2e-docker-choreo:
	@echo "Running E2E tests via docker-compose-choreo (will build and run dependent services if needed)"
	@$(COMPOSE_CHOREO) up --build -d mongodb-choreo neo4j-choreo postgres-choreo core-choreo ingestion-choreo read-choreo
	@$(COMPOSE_CHOREO) up --build e2e-choreo
	@$(COMPOSE_CHOREO) rm -f e2e-choreo || true

e2e-choreo: e2e-venv e2e-ingestion-choreo e2e-read-choreo

e2e-ingestion-choreo: e2e-venv
	@echo "Running E2E tests for Ingestion API against Choreo services (ensure services are up: make up-choreo)"
	@cd $(E2E_DIR) && \
		INGESTION_SERVICE_URL=http://localhost:8080 \
		.venv/bin/python basic_core_tests.py

e2e-read-choreo: e2e-venv
	@echo "Running E2E tests for Read API against Choreo services (ensure services are up: make up-choreo)"
	@cd $(E2E_DIR) && \
		INGESTION_SERVICE_URL=http://localhost:8080 \
		READ_SERVICE_URL=http://localhost:8081 \
		.venv/bin/python basic_read_tests.py
