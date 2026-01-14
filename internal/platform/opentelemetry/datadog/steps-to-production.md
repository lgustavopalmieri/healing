# Guia Completo: OpenTelemetry + Datadog em Produção

## 📋 Índice
1. [Análise da Implementação Atual](#análise-da-implementação-atual)
2. [Criação da Conta Datadog](#criação-da-conta-datadog)
3. [Configuração do Datadog](#configuração-do-datadog)
4. [Ajustes Necessários no Código](#ajustes-necessários-no-código)
5. [Variáveis de Ambiente](#variáveis-de-ambiente)
6. [Deploy da Aplicação](#deploy-da-aplicação)
7. [Verificação e Troubleshooting](#verificação-e-troubleshooting)
8. [Boas Práticas de Produção](#boas-práticas-de-produção)

---

## ✅ Análise da Implementação Atual

### O que está pronto:
- ✅ Integração completa com OpenTelemetry SDK
- ✅ Exporters OTLP para Datadog (traces, metrics, logs)
- ✅ Configuração de TLS para comunicação segura
- ✅ Graceful shutdown dos providers
- ✅ Abstrações para Logger, Tracer e Metrics
- ✅ Propagação de contexto distribuído

### ⚠️ O que precisa de atenção:
- ⚠️ **Falta tratamento de erro robusto** em caso de falha de conexão
- ⚠️ **Falta configuração de retry** para exporters
- ⚠️ **Falta configuração de batching otimizada** para produção
- ⚠️ **Falta health check** para validar conectividade com Datadog
- ⚠️ **Falta configuração de sampling** para ambientes de alto volume

**Conclusão**: A implementação está **80% pronta para produção**, mas precisa de ajustes para ser production-grade.

---

## 🚀 Criação da Conta Datadog

### Passo 1: Criar Conta
1. Acesse: https://www.datadoghq.com/
2. Clique em **"Get Started Free"** ou **"Start Free Trial"**
3. Preencha os dados:
   - Email corporativo
   - Nome da empresa
   - Região (escolha baseado na localização):
     - **US1** (datadoghq.com) - Padrão, Virginia
     - **US3** (us3.datadoghq.com) - Oregon
     - **US5** (us5.datadoghq.com) - Califórnia
     - **EU** (datadoghq.eu) - Frankfurt
     - **AP1** (ap1.datadoghq.com) - Tóquio

⚠️ **IMPORTANTE**: A região não pode ser alterada depois! Escolha a mais próxima dos seus servidores.

### Passo 2: Verificar Email
- Confirme o email de verificação
- Complete o onboarding inicial

### Passo 3: Obter API Key
1. Após login, vá para: **Organization Settings** → **API Keys**
2. Ou acesse diretamente: `https://app.datadoghq.com/organization-settings/api-keys`
3. Clique em **"New Key"**
4. Dê um nome descritivo: `healing-specialist-production`
5. **Copie e guarde a API Key em local seguro** (não será mostrada novamente)

---

## ⚙️ Configuração do Datadog

### Passo 1: Configurar APM (Application Performance Monitoring)
1. No menu lateral, vá para **APM** → **Setup & Configuration**
2. Selecione **"OpenTelemetry"** como método de instrumentação
3. Anote o endpoint OTLP da sua região:
   - US1: `api.datadoghq.com:443`
   - US3: `api.us3.datadoghq.com:443`
   - EU: `api.datadoghq.eu:443`

### Passo 2: Habilitar Log Management
1. Vá para **Logs** → **Configuration**
2. Clique em **"Get Started"**
3. Habilite **"OTLP Ingestion"**

### Passo 3: Configurar Métricas Customizadas
1. Vá para **Metrics** → **Summary**
2. Verifique se **"Custom Metrics"** está habilitado no seu plano

### Passo 4: Criar Service no Datadog
1. Vá para **APM** → **Services**
2. Aguarde a primeira telemetria chegar (após deploy)
3. O serviço `healing-specialist` aparecerá automaticamente

---

## 🔧 Ajustes Necessários no Código

### 1. Melhorar Configuração de Batching e Retry

Crie um novo arquivo: `internal/platform/opentelemetry/config.go`

```go
package opentelemetry

import (
	"time"
)

// ProductionConfig returns optimized settings for production
type ExporterConfig struct {
	// Batching
	MaxQueueSize       int
	BatchTimeout       time.Duration
	ExportTimeout      time.Duration
	MaxExportBatchSize int

	// Retry
	RetryEnabled      bool
	InitialInterval   time.Duration
	MaxInterval       time.Duration
	MaxElapsedTime    time.Duration
}

func DefaultProductionConfig() ExporterConfig {
	return ExporterConfig{
		// Batching - otimizado para produção
		MaxQueueSize:       2048,
		BatchTimeout:       5 * time.Second,
		ExportTimeout:      30 * time.Second,
		MaxExportBatchSize: 512,

		// Retry - configuração agressiva
		RetryEnabled:    true,
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
		MaxElapsedTime:  5 * time.Minute,
	}
}

func DefaultDevelopmentConfig() ExporterConfig {
	return ExporterConfig{
		MaxQueueSize:       512,
		BatchTimeout:       1 * time.Second,
		ExportTimeout:      10 * time.Second,
		MaxExportBatchSize: 128,
		RetryEnabled:       true,
		InitialInterval:    500 * time.Millisecond,
		MaxInterval:        5 * time.Second,
		MaxElapsedTime:     1 * time.Minute,
	}
}
```

### 2. Adicionar Health Check

Adicione ao arquivo `datadog.go`:

```go
// HealthCheck verifica se a conexão com Datadog está funcionando
func (p *DatadogProvider) HealthCheck(ctx context.Context) error {
	// Cria um span de teste
	tracer := otel.Tracer("health-check")
	_, span := tracer.Start(ctx, "datadog.health_check")
	span.SetAttributes(attribute.Bool("health_check", true))
	span.End()

	// Force flush para garantir envio imediato
	if err := p.tracerProvider.ForceFlush(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}
```

### 3. Adicionar Sampling Configurável

Modifique `DatadogConfig`:

```go
type DatadogConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	DatadogSite    string
	APIKey         string
	
	// Sampling
	SamplingRate float64 // 0.0 a 1.0 (1.0 = 100%)
}
```

E ajuste `initTraceProvider`:

```go
func (p *DatadogProvider) initTraceProvider(ctx context.Context, cfg DatadogConfig) error {
	// ... código existente ...

	// Configurar sampler baseado no ambiente
	var sampler sdktrace.Sampler
	if cfg.SamplingRate > 0 && cfg.SamplingRate < 1.0 {
		sampler = sdktrace.TraceIDRatioBased(cfg.SamplingRate)
	} else {
		sampler = sdktrace.AlwaysSample()
	}

	p.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(p.resource),
		sdktrace.WithSampler(sampler), // Usar sampler configurável
	)

	// ... resto do código ...
}
```

---

## 🔐 Variáveis de Ambiente

### Desenvolvimento (`.env.dev`)
```bash
# Datadog Configuration
DD_API_KEY=your_dev_api_key_here
DD_SITE=datadoghq.com
DD_ENV=development
DD_SERVICE_NAME=healing-specialist
DD_SERVICE_VERSION=1.0.0
DD_SAMPLING_RATE=1.0

# Application
PORT=8080
```

### Staging (`.env.staging`)
```bash
# Datadog Configuration
DD_API_KEY=your_staging_api_key_here
DD_SITE=datadoghq.com
DD_ENV=staging
DD_SERVICE_NAME=healing-specialist
DD_SERVICE_VERSION=1.0.0
DD_SAMPLING_RATE=0.5  # 50% sampling

# Application
PORT=8080
```

### Produção (`.env.prod` ou secrets manager)
```bash
# Datadog Configuration
DD_API_KEY=your_production_api_key_here
DD_SITE=datadoghq.com
DD_ENV=production
DD_SERVICE_NAME=healing-specialist
DD_SERVICE_VERSION=1.0.0
DD_SAMPLING_RATE=0.1  # 10% sampling para alto volume

# Application
PORT=8080
```

⚠️ **NUNCA commite arquivos .env no Git!**

---

## 🚢 Deploy da Aplicação

### Opção 1: Docker

#### Dockerfile
```dockerfile
FROM golang:1.25.0-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /healing-specialist ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /healing-specialist .

EXPOSE 8080

CMD ["./healing-specialist"]
```

#### docker-compose.yml (para testes locais)
```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DD_API_KEY=${DD_API_KEY}
      - DD_SITE=${DD_SITE:-datadoghq.com}
      - DD_ENV=${DD_ENV:-development}
      - DD_SERVICE_NAME=healing-specialist
      - DD_SERVICE_VERSION=1.0.0
      - DD_SAMPLING_RATE=${DD_SAMPLING_RATE:-1.0}
    env_file:
      - .env.dev
```

### Opção 2: Kubernetes

#### deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: healing-specialist
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: healing-specialist
  template:
    metadata:
      labels:
        app: healing-specialist
        version: "1.0.0"
    spec:
      containers:
      - name: app
        image: your-registry/healing-specialist:1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: DD_API_KEY
          valueFrom:
            secretKeyRef:
              name: datadog-secrets
              key: api-key
        - name: DD_SITE
          value: "datadoghq.com"
        - name: DD_ENV
          value: "production"
        - name: DD_SERVICE_NAME
          value: "healing-specialist"
        - name: DD_SERVICE_VERSION
          value: "1.0.0"
        - name: DD_SAMPLING_RATE
          value: "0.1"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Secret
metadata:
  name: datadog-secrets
  namespace: production
type: Opaque
data:
  api-key: <base64-encoded-api-key>
```

### Opção 3: AWS ECS/Fargate

#### task-definition.json
```json
{
  "family": "healing-specialist",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "containerDefinitions": [
    {
      "name": "app",
      "image": "your-registry/healing-specialist:1.0.0",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "DD_SITE",
          "value": "datadoghq.com"
        },
        {
          "name": "DD_ENV",
          "value": "production"
        },
        {
          "name": "DD_SERVICE_NAME",
          "value": "healing-specialist"
        },
        {
          "name": "DD_SERVICE_VERSION",
          "value": "1.0.0"
        },
        {
          "name": "DD_SAMPLING_RATE",
          "value": "0.1"
        }
      ],
      "secrets": [
        {
          "name": "DD_API_KEY",
          "valueFrom": "arn:aws:secretsmanager:region:account:secret:datadog-api-key"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/healing-specialist",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

---

## 🔍 Verificação e Troubleshooting

### 1. Verificar Conectividade Local

Antes de fazer deploy, teste localmente:

```bash
# Exportar variáveis
export DD_API_KEY="your_api_key"
export DD_SITE="datadoghq.com"
export DD_ENV="development"

# Rodar aplicação
go run cmd/server/main.go

# Em outro terminal, fazer requisições
curl http://localhost:8080/health
```

### 2. Verificar no Datadog

Após 1-2 minutos do deploy:

1. **APM → Services**: Procure por `healing-specialist`
2. **APM → Traces**: Veja traces individuais
3. **Logs → Live Tail**: Filtre por `service:healing-specialist`
4. **Metrics → Explorer**: Procure por métricas customizadas

### 3. Problemas Comuns

#### ❌ Nenhum dado aparece no Datadog

**Possíveis causas**:
- API Key incorreta
- Site incorreto (US1 vs EU vs US3)
- Firewall bloqueando porta 443
- Aplicação não está enviando telemetria

**Solução**:
```bash
# Testar conectividade
curl -v https://api.datadoghq.com

# Verificar logs da aplicação
docker logs <container-id>

# Adicionar debug logging
export OTEL_LOG_LEVEL=debug
```

#### ❌ Traces aparecem mas logs não

**Causa**: Log provider pode não estar inicializando corretamente

**Solução**: Verificar se `global.SetLoggerProvider` está sendo chamado

#### ❌ Métricas não aparecem

**Causa**: Métricas customizadas podem levar até 5 minutos

**Solução**: Aguardar ou forçar flush:
```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
meterProvider.ForceFlush(ctx)
```

#### ❌ Alto custo de ingestão

**Causa**: Sampling rate muito alto ou muitos spans

**Solução**:
- Reduzir `DD_SAMPLING_RATE` para 0.1 (10%)
- Implementar sampling inteligente (apenas erros em 100%)

---

## 🎯 Boas Práticas de Produção

### 1. Configuração por Ambiente

```go
func GetDatadogConfig() DatadogConfig {
	env := os.Getenv("DD_ENV")
	
	cfg := DatadogConfig{
		ServiceName:    os.Getenv("DD_SERVICE_NAME"),
		ServiceVersion: os.Getenv("DD_SERVICE_VERSION"),
		Environment:    env,
		DatadogSite:    os.Getenv("DD_SITE"),
		APIKey:         os.Getenv("DD_API_KEY"),
	}

	// Sampling baseado no ambiente
	switch env {
	case "production":
		cfg.SamplingRate = 0.1 // 10%
	case "staging":
		cfg.SamplingRate = 0.5 // 50%
	default:
		cfg.SamplingRate = 1.0 // 100%
	}

	return cfg
}
```

### 2. Graceful Shutdown Robusto

```go
func main() {
	ctx := context.Background()
	
	// Inicializar Datadog
	ddProvider, err := opentelemetry.NewDatadogProvider(ctx, GetDatadogConfig())
	if err != nil {
		log.Fatalf("Failed to initialize Datadog: %v", err)
	}

	// Health check inicial
	if err := ddProvider.HealthCheck(ctx); err != nil {
		log.Printf("Warning: Datadog health check failed: %v", err)
		// Não falhar a aplicação, apenas logar
	}

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutting down gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := ddProvider.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}

		os.Exit(0)
	}()

	// Sua aplicação aqui
	startServer()
}
```

### 3. Monitoramento de Custos

No Datadog:
1. Vá para **Organization Settings** → **Usage & Cost**
2. Configure alertas para:
   - Spans ingeridos por hora
   - Logs ingeridos por dia
   - Custom metrics count

### 4. Alertas Importantes

Configure no Datadog:

```yaml
# Exemplo de monitor para alta latência
name: "High Latency - healing-specialist"
type: metric alert
query: "avg(last_5m):avg:trace.grpc.server.duration{service:healing-specialist} > 1000"
message: |
  Latência alta detectada no serviço healing-specialist
  Valor: {{value}}ms
  @slack-alerts @pagerduty
```

### 5. Dashboards Recomendados

Crie um dashboard com:
- **Golden Signals**:
  - Latency (p50, p95, p99)
  - Traffic (requests/sec)
  - Errors (error rate %)
  - Saturation (CPU, Memory)

- **Business Metrics**:
  - Specialists criados por hora
  - Taxa de validação de licença
  - Tempo de resposta do gRPC

---

## 📊 Checklist Final de Produção

- [ ] Conta Datadog criada e verificada
- [ ] API Key gerada e armazenada em secrets manager
- [ ] Região correta configurada
- [ ] Código atualizado com melhorias de retry e batching
- [ ] Variáveis de ambiente configuradas
- [ ] Sampling rate ajustado para produção (0.1)
- [ ] Health check implementado
- [ ] Graceful shutdown testado
- [ ] Deploy realizado com sucesso
- [ ] Dados aparecendo no Datadog (traces, logs, metrics)
- [ ] Dashboards criados
- [ ] Alertas configurados
- [ ] Documentação atualizada
- [ ] Time treinado para usar Datadog

---

## 🆘 Suporte

- **Datadog Docs**: https://docs.datadoghq.com/tracing/setup_overview/open_standards/otlp_ingest_in_the_agent/
- **OpenTelemetry Go**: https://opentelemetry.io/docs/instrumentation/go/
- **Datadog Support**: https://help.datadoghq.com/

---

**Última atualização**: Janeiro 2026
**Versão**: 1.0.0
