.PHONY: run up down es-up es-down es-logs es-health

run:
	@APP_ENV=development go run ./cmd

# Run with Docker
up:
	@ENV_FILE=.env APP_ENV=development docker-compose up -d --build

down:
	@docker-compose down -v 2>/dev/null || true


