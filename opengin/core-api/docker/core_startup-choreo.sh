# Copyright 2025 Lanka Data Foundation
# SPDX-License-Identifier: Apache-2.0

#!/bin/sh
set -e

# CORE Service Startup Script - CHOREO ENVIRONMENT
# This script is specifically configured for the Choreo environment
# and includes database cleanup, testing, and service startup for Choreo.

echo "=== CORE Service Startup (Choreo Environment) ==="
echo "Running tests with environment:"
echo "NEO4J_URI: $NEO4J_URI"
echo "MONGO_URI: $MONGO_URI"
echo "POSTGRES_HOST: $POSTGRES_HOST"


echo "=== Starting CORE Service (Choreo Environment) ==="
exec core-service 2>&1 | tee /app/core-service-choreo.log

