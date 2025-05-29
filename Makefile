# OnlyFans Event Publisher Makefile

.PHONY: help build up down logs clean restart status topics

# Default target
help:
	@echo "OnlyFans Event Publisher - Available commands:"
	@echo ""
	@echo "  build     - Build the Docker image"
	@echo "  up        - Start all services (Redpanda cluster + Publisher)"
	@echo "  down      - Stop all services"
	@echo "  restart   - Restart all services"
	@echo "  logs      - Show logs from all services"
	@echo "  logs-app  - Show logs from the publisher app only"
	@echo "  logs-rp   - Show logs from Redpanda services"
	@echo "  status    - Show status of all services"
	@echo "  topics    - List all Kafka topics"
	@echo "  clean     - Stop services and remove volumes"
	@echo "  shell     - Open shell in running app container"
	@echo ""

# Build the Docker image
build:
	@echo "Building OnlyFans Event Publisher..."
	docker-compose build onlyfans-publisher

# Start all services
up:
	@echo "Starting Redpanda cluster and OnlyFans Event Publisher..."
	docker-compose up -d
	@echo ""
	@echo "Services are starting up..."
	@echo "Redpanda Console: http://localhost:8080"
	@echo "Use 'make logs' to see application logs"

# Stop all services
down:
	@echo "Stopping all services..."
	docker-compose down

# Restart all services
restart: down up

# Show logs from all services
logs:
	docker-compose logs -f

# Show logs from publisher app only
logs-app:
	docker-compose logs -f onlyfans-publisher

# Show logs from Redpanda services
logs-rp:
	docker-compose logs -f redpanda-1 redpanda-2

# Show status of all services
status:
	@echo "Service Status:"
	@docker-compose ps
	@echo ""
	@echo "Redpanda Cluster Health:"
	@docker-compose exec redpanda-1 rpk cluster health 2>/dev/null || echo "Redpanda not ready yet"

# List Kafka topics
topics:
	@echo "Kafka Topics:"
	@docker-compose exec redpanda-1 rpk topic list 2>/dev/null || echo "Redpanda not ready yet"

# Clean up everything (stop services and remove volumes)
clean:
	@echo "Cleaning up all services and volumes..."
	docker-compose down -v
	docker system prune -f

# Open shell in running app container
shell:
	docker-compose exec onlyfans-publisher sh

# Development targets
dev-build:
	@echo "Building for development..."
	go build -o main ./cmd/publisher

dev-run:
	@echo "Running locally (requires local Redpanda)..."
	./main

# Quick development cycle
dev: dev-build dev-run

# Show application environment
env:
	@echo "Current environment variables:"
	@docker-compose exec onlyfans-publisher env | grep -E "(REDPANDA|TOPIC|NUM_|INTERVAL|ABNORMAL)" | sort