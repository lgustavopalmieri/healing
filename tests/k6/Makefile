.PHONY: setup run-stress run-simple run-create-specialist clean help

help:
	@echo "📋 K6 Stress Tests - Available commands:"
	@echo ""
	@echo "  make setup              - Setup proto files"
	@echo "  make run-stress         - Run stress test (CreateSpecialist)"
	@echo "  make run-simple         - Run simple test (5 VUs, 30s)"
	@echo "  make run-create-specialist - Run create specialist test (as configured)"
	@echo "  make run-create-specialist-http - Run create specialist HTTP test"
	@echo "  make clean              - Clean proto files"
	@echo ""

# Executa teste create specialist (configuração interna do arquivo)
run-create-specialist: setup
	@echo "🚀 Running create specialist test..."
	@docker-compose -f docker-compose.k6.yml run --rm k6 run /scripts/modules/specialist/create/stress/create-specialist-test.js
	@echo "✅ Create specialist test completed!"

run-create-specialist-http:
	@echo "🚀 Running create specialist HTTP test..."
	@docker-compose -f docker-compose.k6.yml run --rm k6 run /scripts/modules/specialist/create/stress-http/create-specialist-test.js
	@echo "✅ Create specialist HTTP test completed!"

# Limpa arquivos temporários
clean:
	@echo "🧹 Cleaning up..."
	@rm -rf proto
	@echo "✅ Cleanup done"
