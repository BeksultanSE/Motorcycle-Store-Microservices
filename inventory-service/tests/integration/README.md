# Inventory Service Integration Tests

This directory contains integration tests for the Inventory Service.

## Requirements

- MongoDB running locally or accessible via environment variables

## Environment Variables

- `MONGO_URI`: Connection string for MongoDB (default: "mongodb://localhost:27017")

## Running the Tests

```bash
go test -v ./...
``` 