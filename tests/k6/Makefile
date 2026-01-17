.PHONY: setup run-stress run-simple clean help

help:
	@echo "📋 K6 Stress Tests - Available commands:"
	@echo ""
	@echo "  make setup              - Setup proto files"
	@echo "  make run-stress         - Run stress test (CreateSpecialist)"
	@echo "  make run-simple         - Run simple test (5 VUs, 30s)"
	@echo "  make clean              - Clean proto files"
	@echo ""

# Executa teste de stress completo
run-stress: setup
	@echo "🔥 Running stress test (CreateSpecialist)..."
	@docker-compose -f docker-compose.k6.yml run --rm k6 run /scripts/modules/specialist/create/stress/create-specialist-test.js
	@echo "✅ Stress test completed!"

# Executa teste simples (para validação rápida)
run-simple: setup
	@echo "🧪 Running simple test..."
	@docker-compose -f docker-compose.k6.yml run --rm k6 run --vus 5 --duration 30s /scripts/modules/specialist/create/stress/create-specialist-test.js
	@echo "✅ Simple test completed!"

# Limpa arquivos temporários
clean:
	@echo "🧹 Cleaning up..."
	@rm -rf proto
	@echo "✅ Cleanup done"
