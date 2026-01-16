.PHONY: server-run server-run-test server-build server-up server-up-test server-down server-logs help

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
	@docker-compose down 2>/dev/null || true
	ENV_FILE=.env.test APP_ENV=test docker-compose up -d --build
	@echo "✅ Server is running in TEST mode!"
	@echo "🚀 gRPC Server: localhost:50051"
	@echo "🧪 Connecting to test containers (postgres:5433, broker:9093)"

server-down:
	@echo "🛑 Stopping server..."
	docker-compose down
	@echo "✅ Server stopped"

server-logs:
	@echo "📋 Showing server logs..."
	docker-compose logs -f healing-specialist
