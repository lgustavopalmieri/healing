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

down:
	@echo "🛑 Stopping server..."
	docker-compose down -v 2>/dev/null || true
	docker-compose -f docker-compose.test.yml down -v 2>/dev/null || true
	@echo "✅ Server stopped"

# Elasticsearch commands
es-up:
	@echo "🔍 Starting Elasticsearch..."
	docker-compose -f docker-compose.elasticsearch.yml up -d
	@echo "✅ Elasticsearch is running!"
	@echo "🔍 Elasticsearch: http://localhost:9200"

es-down:
	@echo "🛑 Stopping Elasticsearch..."
	docker-compose -f docker-compose.elasticsearch.yml down -v
	@echo "✅ Elasticsearch stopped"

es-logs:
	@echo "📋 Elasticsearch logs..."
	docker-compose -f docker-compose.elasticsearch.yml logs -f elasticsearch

es-health:
	@echo "🏥 Checking Elasticsearch health..."
	@curl -s http://localhost:9200/_cluster/health?pretty
