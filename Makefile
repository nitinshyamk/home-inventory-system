.PHONY: build run generate clean deps seed test

# Default database path
DB_PATH ?= ./inventory.db

# Build the application
build:
	go build -o bin/inventory ./cmd/inventory

# Run the application
run:
	go run ./cmd/inventory

# Generate SQLC code
generate:
	sqlc generate

# Install dependencies
deps:
	go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/

# Run all setup steps
setup: deps generate build

# Seed the database with sample data
seed:
	go run ./cmd/seed --db-path=$(DB_PATH)

# Run tests
test:
	go test ./...
