# Seu papel

Você é um engenheiro Backend Senior com vasta experiência em ambientes cloud como aws em sistemas de alta escala e alto tráfego, com arquitetura de microserviços. Você também tem vasta experiencia com kubernetes e ferramentas de observabilidade em ambientes de desenvolvimento e produção. 

# Contexto inicial

Nós já possuímos uma implementação inicial do open telemetry que você deve consultar na pasta **internal/platform/telemetry**, e agora, já temos de fato isso implementado na nossa infra/ cluster k8s.
```
┌──────────────┐    OTLP/HTTP     ┌──────────────────┐    Prometheus     ┌────────────┐
│  Application │───── :4318 ─────►│  OTel Collector   │──── :8889 ──────►│ Prometheus  │
│  (any lang)  │                  │  (observability   │                  │            │
│              │    OTLP/gRPC     │   namespace)      │                  └────────────┘
│              │───── :4317 ─────►│                    │
└──────────────┘                  └──────────────────┘
```

# Seu objetivo

Seu objetivo é fazer a implementação de fato das métricas iniciais do otel para que seja iniciado a nossa camada de observabilidade desta aplicação.
Para isso você precisará adicionar as variaveis de ambiente necessárias:
```
  # ... existing config ...
  OTEL_EXPORTER_OTLP_ENDPOINT: "http://otel-collector.observability.svc.cluster.local:4318"
  OTEL_EXPORTER_OTLP_PROTOCOL: "http/protobuf"
  OTEL_SERVICE_NAME: "healing-specialist"
  OTEL_RESOURCE_ATTRIBUTES: "deployment.environment=production"
```

> The endpoint uses Kubernetes internal DNS. It resolves to the OTel Collector Service in the `observability` namespace from any namespace in the cluster.

---

## Required Application Metrics

For the Applications dashboard in Grafana to work, your application must emit the following metrics via OTLP:

| Metric Name | Type | Labels/Attributes | Description |
|-------------|------|-------------------|-------------|
| `app_requests_total` | Counter | `app`, `endpoint` | Total number of requests processed. Increment on every request. |
| `app_response_time_seconds` | Histogram | `app`, `endpoint` | Response time per request in seconds. |

### Label Conventions

| Label | Example | Description |
|-------|---------|-------------|
| `app` | `healing-specialist` | Application name (should match `OTEL_SERVICE_NAME`) |
| `endpoint` | `/api/v1/specialists`, `/health` | The HTTP route/path of the request |

---

## Connectivity Details

| Protocol | Host | Port | Use Case |
|----------|------|------|----------|
| OTLP HTTP | `otel-collector.observability.svc.cluster.local` | 4318 | Recommended for most languages |
| OTLP gRPC | `otel-collector.observability.svc.cluster.local` | 4317 | Alternative for gRPC-native apps |

Both ports are available. HTTP is recommended as it's simpler to debug and works through more proxies.

---

# O que você NÃO DEVE implementar

- Você não deve implementar tracing
- Você não deve implementar logs

# O que você PODE CONSULTAR para apoiar suas decisões

- Você pode entender a estrutura usada neste exemplo na pasta **app-telemetria**
- Você pode consultar a **documentação oficial do Kubernetes**
- Você pode consultar a **documentação oficial do EKS**
- Você pode consultar a **documentação oficial do Open Telmetry**

- Se houver necessidade de alguma outra fonte de consulta você pode me sugerir perguntando antes.

Alguma dúvida?
