#!/bin/bash

# Demo script for go-smart-queue
# This script demonstrates various features of the job scheduler

echo "======================================"
echo "  Go Smart Queue Demo"
echo "======================================"
echo ""

# Build the application
echo "Building go-smart-queue..."
go build -o go-smart-queue
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi
echo "Build successful!"
echo ""

# Add some jobs with different priorities
echo "Adding jobs to the queue..."
echo ""

echo "1. Adding high priority job..."
./go-smart-queue add --name "high-priority-job" --priority 20 --max-cpu 90 --sleep 3

echo ""
echo "2. Adding medium priority job..."
./go-smart-queue add --name "medium-priority-job" --priority 10 --max-cpu 80 --sleep 4

echo ""
echo "3. Adding low priority job..."
./go-smart-queue add --name "low-priority-job" --priority 5 --max-cpu 70 --sleep 2

echo ""
echo "4. Adding another high priority job..."
./go-smart-queue add --name "urgent-task" --priority 25 --max-cpu 85 --sleep 3

echo ""
echo "5. Adding background task..."
./go-smart-queue add --name "background-task" --priority 1 --max-cpu 60 --max-memory 512 --sleep 5

echo ""
echo "======================================"
echo "Waiting 2 seconds before checking status..."
sleep 2

echo ""
echo "======================================"
echo "Listing all jobs:"
echo "======================================"
./go-smart-queue list

echo ""
echo "======================================"
echo "System Status:"
echo "======================================"
./go-smart-queue status

echo ""
echo "======================================"
echo "Demo completed!"
echo "Try running: ./go-smart-queue dashboard"
echo "======================================"
