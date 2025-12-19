#!/bin/bash

# Build and run the example program

set -e

echo "Building example..."
cd "$(dirname "$0")"

# Create temporary directory for build
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Copy necessary files
cp example.go ../job.go ../monitor.go ../scheduler.go ../go.mod ../go.sum "$TEMP_DIR/" 2>/dev/null || true

# Build
cd "$TEMP_DIR"
go mod download 2>/dev/null || true
go build -o example-run .

echo ""
echo "Running example..."
echo ""
./example-run
