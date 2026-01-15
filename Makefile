.PHONY: server-run server-build grpcserver-up grpcserver-down grpcserver-logs grpcserver-rebuild

server-run:
	@echo "Starting server..."
	cd cmd/grpcserver && go run main.go

server-build:
	@echo "Building server..."
	go build -o bin/server cmd/grpcserver/main.go

grpcserver-up:
	@echo "Starting gRPC server with Docker Compose..."
	docker-compose up -d
	@echo "✅ gRPC Server is running!"
	@echo "🚀 gRPC Server: localhost:50051"

grpcserver-rebuild:
	@echo "Rebuilding and starting gRPC server..."
	docker-compose down
	docker-compose up --build -d
	@echo "✅ gRPC Server rebuilt and running!"
	@echo "🚀 gRPC Server: localhost:50051"

grpcserver-down:
	@echo "Stopping gRPC server..."
	docker-compose down -v

grpcserver-logs:
	docker-compose logs -f healing-specialist
