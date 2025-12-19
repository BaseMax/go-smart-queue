#!/bin/bash

# Build and run the example program

set -e

echo "Building example..."
cd "$(dirname "$0")"

# Create temporary directory for build
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Copy necessary files
if [ ! -f ../job.go ] || [ ! -f ../monitor.go ] || [ ! -f ../scheduler.go ]; then
    echo "Error: Required source files not found in parent directory"
    exit 1
fi

cp example.go ../job.go ../monitor.go ../scheduler.go "$TEMP_DIR/"
if [ -f ../go.mod ]; then
    cp ../go.mod "$TEMP_DIR/"
fi
if [ -f ../go.sum ]; then
    cp ../go.sum "$TEMP_DIR/"
fi

# Build
cd "$TEMP_DIR"
if [ -f go.mod ]; then
    go mod download
fi
go build -o example-run .

echo ""
echo "Running example..."
echo ""
./example-run
