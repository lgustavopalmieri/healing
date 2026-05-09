.PHONY: run up down es-up es-down es-logs es-health auth-keys

run:
	@APP_ENV=development go run ./cmd

# Run with Docker
up:
	@ENV_FILE=.env APP_ENV=development docker-compose up -d --build

down:
	@docker-compose down -v 2>/dev/null || true

auth-keys:
	@mkdir -p keys
	@echo "*.pem" > keys/.gitignore
	@echo "!.gitignore" >> keys/.gitignore
	@openssl genpkey -algorithm RSA -out keys/auth-private.pem -pkeyopt rsa_keygen_bits:2048
	@openssl rsa -pubout -in keys/auth-private.pem -out keys/auth-public.pem
	@chmod 600 keys/auth-private.pem
	@chmod 644 keys/auth-public.pem
	@echo "Auth keys generated at keys/"


