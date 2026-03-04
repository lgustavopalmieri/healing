.PHONY: run up down es-up es-down es-logs es-health

# Run locally without Docker
run:
	@echo "🚀 Starting server locally in DEVELOPMENT mode..."
	@export APP_ENV=development && cd cmd/grpcserver && go run main.go

# Run with Docker
up:
	@echo "🐳 Starting server in Docker (DEVELOPMENT mode)..."
	ENV_FILE=.env APP_ENV=development docker-compose up -d --build
	@echo "✅ Server is running in DEVELOPMENT mode!"
	@echo "🚀 gRPC Server: localhost:50051"
	@echo "🌐 HTTP Server: localhost:4000"
	@echo "📊 Metrics:     localhost:4001"

down:
	@echo "🛑 Stopping server..."
	docker-compose down -v 2>/dev/null || true
	docker-compose -f docker-compose.test.yml down -v 2>/dev/null || true
	@echo "✅ Server stopped"


