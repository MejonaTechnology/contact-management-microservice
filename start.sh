#!/bin/bash

echo "Starting Mejona Contact Service..."
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

# Build and run the service
echo "Building contact service..."
go build -o contact-service cmd/server/main.go
if [ $? -ne 0 ]; then
    echo "Error: Failed to build contact service"
    exit 1
fi

echo "Starting contact service on port 8081..."
echo "Dashboard endpoints will be available at:"
echo "  GET  http://localhost:8081/api/v1/dashboard/contacts"
echo "  GET  http://localhost:8081/api/v1/dashboard/contacts/stats"
echo "  POST http://localhost:8081/api/v1/dashboard/contact"
echo "  PUT  http://localhost:8081/api/v1/dashboard/contacts/{id}/status"
echo ""
echo "Press Ctrl+C to stop the service"
echo ""

./contact-service