# PostgreSQL Test Helper

Este pacote fornece utilitários reutilizáveis para testes de integração com PostgreSQL usando TestContainers.

## Uso Básico

### 1. Configurar TestMain e setupTestDB

```go
package mypackage

import (
    "database/sql"
    "testing"
    
    "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/database/postgresql"
)

var testHelper = postgresql.NewTestHelper()

func TestMain(m *testing.M) {
    testHelper.RunTestMain(m)
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
    return testHelper.SetupTestDB(t)
}
```

### 2. Usar nos testes

```go
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Each test gets a clean database
			db, cleanup := setupTestDB(t)
			defer cleanup()

			repo := NewSpecialistCreateRepository(db)

            // ... id, email, licenseNumber := tt.setupMocks(repo)
            }
        )
    }
```

## Características

- **Container Compartilhado**: Um container PostgreSQL para todos os testes (mais rápido)
- **Bancos Isolados**: Cada teste recebe um banco de dados limpo
- **Migrations Automáticas**: Executa migrations automaticamente
- **Cleanup Automático**: Limpa recursos após os testes
- **PostgreSQL 16 Alpine**: Imagem leve e rápida

## Estrutura Interna

- `TestHelper`: Gerencia o ciclo de vida do container
- `PostgreSQLContainer`: Wrapper do container testcontainers
- `RunTestMain()`: Implementação reutilizável do TestMain
- `SetupTestDB()`: Cria banco limpo para cada teste
- `CreateCleanDatabase()`: Cria e configura novo banco
- `Terminate()`: Limpa recursos

## Vantagens

✅ **Reutilização**: Mesmo código para todos os testes de repository  
✅ **Performance**: Container compartilhado, bancos isolados  
✅ **Isolamento**: Cada teste tem seu próprio banco limpo  
✅ **Simplicidade**: Apenas 3 linhas para configurar  
✅ **Manutenibilidade**: Mudanças centralizadas em um lugar  
