# Copyright 2025 Lanka Data Foundation
# SPDX-License-Identifier: Apache-2.0

echo "Building all packages..."
go build -v ./...
echo "Building core-service binary..."
go build -v -o core-service cmd/server/service.go cmd/server/utils.go
echo "Build complete!"