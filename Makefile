.PHONY: server-run server-run-test server-build server-up server-up-test server-down server-logs k6-setup k6-stress k6-simple k6-clean help

# Run locally without Docker
server-run:
	@echo "🚀 Starting server locally in DEVELOPMENT mode..."
	@export APP_ENV=development && cd cmd/grpcserver && go run main.go

# Run with Docker
server-up:
	@echo "🐳 Starting server in Docker (DEVELOPMENT mode)..."
	ENV_FILE=.env APP_ENV=development docker-compose up -d --build
	@echo "✅ Server is running in DEVELOPMENT mode!"
	@echo "🚀 gRPC Server: localhost:50051"

server-down:
	@echo "🛑 Stopping server..."
	docker-compose down -v 2>/dev/null || true
	docker-compose -f docker-compose.test.yml down -v 2>/dev/null || true
	@echo "✅ Server stopped"
