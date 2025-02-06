# Simple Makefile for a Go project

# Build the application
all: build test

build:
	@echo "Building..."
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go
# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v
# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

# Migration Commands
migrate-up:
	@echo "Running migrations..."
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found"; \
		exit 1; \
	fi
	@source .env && migrate -path internal/database/migrations -database "postgresql://$${USERNAME}:$${PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DATABASE}?sslmode=disable" up

migrate-down:
	@echo "Rolling back migrations..."
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found"; \
		exit 1; \
	fi
	@source .env && migrate -path internal/database/migrations -database "postgresql://$${USERNAME}:$${PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DATABASE}?sslmode=disable" down

migrate-create:
	@if [ -z "$(name)" ]; then \
		read -p "Enter migration name: " migration_name; \
		if [ -z "$$migration_name" ]; then \
			echo "Migration name cannot be empty"; \
			exit 1; \
		fi; \
		migrate create -ext sql -dir internal/database/migrations -seq $$migration_name; \
	else \
		migrate create -ext sql -dir internal/database/migrations -seq $(name); \
	fi
	@echo "Migration files created successfully!"

.PHONY: all build run test clean watch docker-run docker-down itest migrate-up migrate-down migrate-create
