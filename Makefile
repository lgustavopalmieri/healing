.PHONY: server-run server-run-test server-build server-up server-up-test server-down server-logs stress-test stress-test-full help

help:
	@echo "📋 Available commands:"
	@echo ""
	@echo "  Local Development:"
	@echo "    make server-run              - Run server locally (development mode)"
	@echo "    make server-run-test         - Run server locally (test mode)"
	@echo "    make server-build            - Build server binary"
	@echo ""
	@echo "  Docker:"
	@echo "    make server-up               - Start server in Docker (development mode)"
	@echo "    make server-up-test          - Start server in Docker (test mode)"
	@echo "    make server-down             - Stop server"
	@echo "    make server-logs             - Show server logs"
	@echo ""
	@echo "  Stress Testing:"
	@echo "    make stress-test             - Run simple stress test (5 VUs, 30s)"
	@echo "    make stress-test-full        - Run full stress test (ramp up to 50 VUs)"
	@echo ""

# Run locally without Docker
server-run:
	@echo "🚀 Starting server locally in DEVELOPMENT mode..."
	@export APP_ENV=development && cd cmd/grpcserver && go run main.go

server-run-test:
	@echo "🧪 Starting server locally in TEST mode..."
	@export APP_ENV=test && cd cmd/grpcserver && go run main.go

server-build:
	@echo "🔨 Building server..."
	go build -o bin/server cmd/grpcserver/main.go

# Run with Docker
server-up:
	@echo "🐳 Starting server in Docker (DEVELOPMENT mode)..."
	ENV_FILE=.env APP_ENV=development docker-compose up -d --build
	@echo "✅ Server is running in DEVELOPMENT mode!"
	@echo "🚀 gRPC Server: localhost:50051"

server-up-test:
	@echo "🐳 Starting server in Docker (TEST mode)..."
	@docker-compose -f docker-compose.test.yml down 2>/dev/null || true
	docker-compose -f docker-compose.test.yml up -d --build
	@echo "✅ Server is running in TEST mode!"
	@echo "🚀 gRPC Server: localhost:50051"
	@echo "🧪 Connecting to test containers (postgres:5433, broker:9093)"

server-down:
	@echo "🛑 Stopping server..."
	docker-compose down -v 2>/dev/null || true
	docker-compose -f docker-compose.test.yml down -v 2>/dev/null || true
	@echo "✅ Server stopped"

server-logs:
	@echo "📋 Showing server logs..."
	docker-compose logs -f healing-specialist

# Stress Testing
stress-test:
	@echo "🔥 Running simple stress test (5 VUs, 30s)..."
	@echo "📊 Target: healing-specialist:50051"
	cd tests/stress && docker-compose -f docker-compose.k6.yml run --rm k6 run /scripts/simple-test.js
	@echo "✅ Stress test completed!"

stress-test-full:
	@echo "🔥 Running full stress test (ramp up to 50 VUs)..."
	@echo "📊 Target: healing-specialist:50051"
	cd tests/stress && docker-compose -f docker-compose.k6.yml run --rm k6 run /scripts/create-specialist.js
	@echo "✅ Full stress test completed!"
