## Rodar todos os testes do projeto
```bash
go test ./...
```

**Com output detalhado (recomendado sempre):**
```bash
go test -v ./...
```

**Sem cache (muito importante quando está depurando):**
```bash
go test -v -count=1 ./...
```


## Rodar todos os testes de um pacote / pasta específica

**Exemplo: internal/service**
```bash
go test ./internal/service
```

**Com verbose:**
```bash
go test -v ./internal/service
```

**Incluindo subpastas:**
```bash
go test -v ./internal/service/...
```


## Rodar um teste específico
```bash
go test -v ./internal/service -run ^TestCreateOrder$
```

## Rodar testes com race detector
```bash
go test -v -race ./...
```


##  Coverage do projeto todo (no terminal)
```bash
go test ./... -cover
```

**Com detalhes por pacote:**
```bash
go test ./... -coverprofile=coverage.out
```

**Abrindo HTML:**
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Coverage só de um pacote:**
```bash
go test ./internal/service -cover
go test ./internal/service -coverprofile=coverage.out

```