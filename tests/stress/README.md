# Stress Tests - K6

Testes de carga e stress para o serviço Healing Specialist usando K6.

## 📋 Pré-requisitos

1. Docker rodando
2. Aplicação `healing-specialist` rodando na rede `healing-network`
3. Containers de infraestrutura (Postgres, Kafka) rodando

## 🚀 Como Executar

### Teste Simples (Recomendado para começar)

```bash
make stress-test
```

**Configuração:**
- 5 usuários virtuais (VUs)
- Duração: 30 segundos
- Threshold: 95% das requisições < 1s

### Teste Completo (Stress Test)

```bash
make stress-test-full
```

**Configuração:**
- Ramp up: 10 → 50 usuários virtuais
- Duração total: ~4 minutos
- Threshold: 95% das requisições < 500ms

## 📊 Estrutura dos Testes

```
tests/stress/
├── simple-test.js           # Teste básico (5 VUs, 30s)
├── create-specialist.js     # Teste completo com ramp up
├── proto/
│   └── specialist.proto     # Definição do serviço gRPC
├── docker-compose.k6.yml    # Configuração do K6
└── README.md
```

## 🎯 O que é testado

- **CreateSpecialist gRPC endpoint**
- Criação de specialists com dados únicos
- Validação de resposta (status, ID, email)
- Performance sob carga

## 📈 Métricas Coletadas

- `grpc_req_duration`: Duração das requisições gRPC
- `checks`: Taxa de sucesso das verificações
- `http_reqs`: Total de requisições
- `vus`: Usuários virtuais ativos

## 🔍 Analisando Resultados

O K6 mostra no terminal:

```
✓ status is OK
✓ has specialist
✓ has ID

checks.........................: 100.00% ✓ 150  ✗ 0
grpc_req_duration..............: avg=245ms min=120ms med=230ms max=450ms p(95)=380ms
http_reqs......................: 150     5/s
vus............................: 5       min=5 max=5
```

**Interpretação:**
- ✅ `checks 100%` = Todas as verificações passaram
- ✅ `p(95)=380ms` = 95% das requisições < 380ms
- ✅ `http_reqs 5/s` = 5 requisições por segundo

## 🛠️ Customização

### Ajustar número de usuários

Edite `simple-test.js`:

```javascript
export const options = {
  vus: 10,              // Aumentar para 10 VUs
  duration: '1m',       // Aumentar para 1 minuto
};
```

### Ajustar thresholds

```javascript
thresholds: {
  'grpc_req_duration': ['p(95)<500'],  // Mais rigoroso
  'checks': ['rate>0.99'],              // 99% de sucesso
},
```

## 🐛 Troubleshooting

### Erro: "connection refused"

```bash
# Verifique se a aplicação está rodando
docker ps | grep healing-specialist

# Verifique se está na rede correta
docker network inspect healing-network
```

### Erro: "proto file not found"

```bash
# Verifique se o arquivo proto existe
ls tests/stress/proto/specialist.proto
```

### Muitos erros no teste

1. Verifique os logs da aplicação: `make server-logs`
2. Reduza o número de VUs
3. Aumente o delay entre requisições (`sleep(1)`)

## 📝 Próximos Passos

- [ ] Adicionar testes de spike (carga súbita)
- [ ] Adicionar testes de soak (longa duração)
- [ ] Integrar com Grafana para visualização
- [ ] Adicionar testes para outros endpoints
- [ ] Configurar CI/CD para rodar testes automaticamente
