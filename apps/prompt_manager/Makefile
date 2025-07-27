# Prompt Manager Service Management
.PHONY: help build start stop restart status test clean logs

# Default target
help:
	@echo "Prompt Manager Service Management"
	@echo "================================"
	@echo "Available commands:"
	@echo "  build      - Build all Go services"
	@echo "  start      - Start API server and queue processor"
	@echo "  stop       - Stop all services"
	@echo "  restart    - Restart all services"
	@echo "  status     - Show service status"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build artifacts and logs"
	@echo "  logs       - Show service logs"
	@echo ""

# Build services
build:
	@echo "Building prompt manager services..."
	@go build -o bin/api cmd/main.go
	@go build -o bin/processor cmd/processor/main.go
	@go build -o bin/test-db tools/test-db/main.go
	@echo "Build complete: bin/api, bin/processor, bin/test-db"

# Start services in background
start: build
	@echo "Starting prompt manager services..."
	@echo "Starting API server on port 8082..."
	@nohup ./bin/api > logs/api.log 2>&1 & echo $$! > logs/api.pid
	@sleep 2
	@echo "Services started. Use 'make status' to check health."

# Stop services
stop:
	@echo "Stopping prompt manager services..."
	@if [ -f logs/api.pid ]; then \
		kill `cat logs/api.pid` 2>/dev/null || true; \
		rm -f logs/api.pid; \
		echo "API server stopped"; \
	fi
	@if [ -f logs/processor.pid ]; then \
		kill `cat logs/processor.pid` 2>/dev/null || true; \
		rm -f logs/processor.pid; \
		echo "Queue processor stopped"; \
	fi
	@echo "All services stopped."

# Restart services
restart: stop start

# Check service status
status:
	@echo "Prompt Manager Service Status"
	@echo "============================"
	@if [ -f logs/api.pid ] && kill -0 `cat logs/api.pid` 2>/dev/null; then \
		echo "✓ API server (PID: `cat logs/api.pid`) - http://localhost:8082"; \
		curl -s http://localhost:8082/health | grep -q "success.*true" && echo "  Health check: PASS" || echo "  Health check: FAIL"; \
	else \
		echo "✗ API server - STOPPED"; \
	fi
	@if [ -f logs/processor.pid ] && kill -0 `cat logs/processor.pid` 2>/dev/null; then \
		echo "✓ Queue processor (PID: `cat logs/processor.pid`) - RUNNING"; \
	else \
		echo "✗ Queue processor - STOPPED"; \
	fi
	@echo ""
	@echo "Database status:"
	@if [ -f data/prompt_manager.db ]; then \
		echo "✓ Database exists: data/prompt_manager.db"; \
		./bin/test-db 2>/dev/null | head -3; \
	else \
		echo "✗ Database not found"; \
	fi

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Show logs
logs:
	@echo "=== API Server Logs ==="
	@tail -n 20 logs/api.log 2>/dev/null || echo "No API logs found"
	@echo ""
	@echo "=== Queue Processor Logs ==="
	@tail -n 20 logs/processor.log 2>/dev/null || echo "No processor logs found"

# Clean up
clean:
	@echo "Cleaning up..."
	@make stop 2>/dev/null || true
	@rm -rf bin/ logs/ data/queue/
	@echo "Cleanup complete"

# Development helpers
dev-api: build
	@echo "Starting API server in development mode..."
	@./bin/api

dev-processor: build
	@echo "Starting queue processor in development mode..."
	@./bin/processor

dev-test: build
	@echo "Testing database connection..."
	@./bin/test-db
