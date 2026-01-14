.PHONY: server-run server-build grpcserver-up grpcserver-down grpcserver-logs

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

grpcserver-down:
	@echo "Stopping gRPC server..."
	docker-compose down

grpcserver-logs:
	docker-compose logs -f healing-specialist
