# ADR-001: Migração Kafka → AWS SQS e Elasticsearch → AWS OpenSearch

**Status:** Proposta  
**Data:** 2026-03-29  
**Autores:** Equipe Healing  
**Serviço:** healing-specialist

---

## Contexto

O serviço healing-specialist utiliza Kafka (via franz-go) como message broker para eventos internos e Elasticsearch v8 (via go-elasticsearch/v8) como engine de busca. Ambos são gerenciados externamente ao cluster Kubernetes e exigem overhead operacional significativo.

A plataforma Healing roda inteiramente na AWS (EKS, RDS, ECR). Migrar para serviços gerenciados AWS (SQS e OpenSearch) reduz complexidade operacional, unifica o modelo de autenticação via IAM e alinha a stack com o ecossistema já utilizado.

---

## Decisão

1. Substituir **Apache Kafka** por **AWS SQS (FIFO queues)** para toda a comunicação assíncrona interna do serviço.
2. Substituir **Elasticsearch 8.x** por **AWS OpenSearch Service** para indexação e busca full-text.

---

## Parte 1: Kafka → AWS SQS

### 1.1 Estado Atual

O serviço utiliza Kafka com o client franz-go para:

- **1 Producer síncrono** — publica eventos `specialist.created` e `specialist.updated` após operações de escrita no PostgreSQL.
- **2 Consumer Groups**:
  - `specialist-validate-license-consumer-group` — consome `specialist.created`, valida licença via API externa e republica `specialist.updated`.
  - `specialist-update-data-repositories-consumer-group` — consome `specialist.updated`, indexa no Elasticsearch e atualiza repositórios de dados.
- **DLQ manual** — em caso de falha no Elasticsearch, o serviço publica para o tópico `specialist.updated.elasticsearch.dlq`.
- **Criação de tópicos no startup** — via `kadm.CreateTopics` (idempotente) em `cmd/server/bootstrap/kafka.go`.
- **Autenticação** — SASL/PLAIN + TLS configurável por ambiente.

Abstrações existentes que facilitam a migração:

- Interface `EventDispatcher` (`internal/commom/event/dipstacher.go`) — o `KafkaProducer` implementa esta interface.
- `Listener` + `ListenerManager` (`internal/commom/event/listener.go`) — abstraem o roteamento evento→handler.
- Interface `DataRepository` (`internal/modules/specialist/features/update/.../listener/interface.go`) — abstrai os repositórios de dados downstream.

### 1.2 Mapeamento Kafka → SQS

| Kafka Topic | SQS Queue | Tipo | Observação |
|-------------|-----------|------|------------|
| `specialist.created` | `specialist-created.fifo` | FIFO | MessageGroupId = specialist ID |
| `specialist.updated` | `specialist-updated.fifo` | FIFO | MessageGroupId = specialist ID |
| `specialist.updated.dlq` (manual) | Redrive Policy nativa na fila `specialist-updated.fifo` | FIFO DLQ | `maxReceiveCount=3` |
| `specialist.updated.elasticsearch.dlq` | Eliminada — coberta pela DLQ nativa acima | — | Simplificação |

### 1.3 Criação de Filas pela Aplicação (App-Level)

**Decisão**: as filas SQS serão criadas pela própria aplicação no startup, seguindo o mesmo pattern do `EnsureTopics` atual com Kafka.

**Justificativa**:

- A operação `CreateQueue` do SQS é **idempotente por design**: se a fila já existe com os mesmos atributos, retorna a Queue URL existente sem erro.
- Múltiplos pods chamando `CreateQueue` simultaneamente no Kubernetes é seguro — todos recebem a mesma Queue URL.
- Elimina a necessidade de gerenciar Queue URLs no Terraform e propagá-las via ConfigMap para cada novo tópico/fila.
- Os nomes das filas são derivados das mesmas constantes de eventos já existentes no código (`SpecialistCreatedEventName`, etc.).
- Mantém consistência com o pattern de auto-provisionamento que já existe no projeto.

**Fluxo de inicialização por pod**:

1. Init SQS Client (AWS SDK v2, credenciais via IRSA)
2. `EnsureQueues` — para cada fila definida no código:
   a. Criar DLQ primeiro (se aplicável)
   b. Obter ARN da DLQ via `GetQueueAttributes`
   c. Criar fila principal com `RedrivePolicy` apontando para a DLQ
   d. Retornar mapa `eventName → queueURL`
3. Init Producer com o mapa de Queue URLs
4. Init Consumers (um por fila)

**Cuidados**:

- Os atributos da fila (`VisibilityTimeout`, `MessageRetentionPeriod`, `FifoQueue`, etc.) devem ser **constantes no código**, não configuráveis por env. Se dois pods tentarem criar a mesma fila com atributos diferentes, o SQS retorna erro `QueueNameExists`.
- A IAM Policy do pod deve incluir `sqs:CreateQueue` além das permissões de leitura/escrita (ver seção de Terraform no documento complementar).
- Para ambientes de desenvolvimento local, usar **LocalStack** ou **ElasticMQ** como substituto do SQS (configurável via `Endpoint` na config).

**Configuração simplificada** — ao invés de gerenciar Queue URLs:

```
SQS_REGION=us-east-1
SQS_QUEUE_PREFIX=specialist
SQS_ENDPOINT=          # opcional, para LocalStack
```

Os nomes completos das filas são derivados: `{prefix}-{event-suffix}.fifo`.

### 1.4 Tradeoffs Técnicos

| Aspecto | Kafka | AWS SQS | Veredicto |
|---------|-------|---------|-----------|
| **Modelo de entrega** | Log imutável com offsets | Fila com delete após processamento | SQS é suficiente para event-driven interno |
| **Ordenação** | Garantida por partição | FIFO: garantida por MessageGroupId | Equivalente usando specialist ID como group ID |
| **Throughput máximo** | Milhões msg/s por cluster | FIFO: 3.000 msg/s (70.000 com high throughput mode) | Mais que suficiente para volume de eventos internos |
| **Replay/Reprocessamento** | Possível (reset de offset) | Impossível após delete | **Tradeoff aceito**: reindexação deve ser feita a partir do PostgreSQL (source of truth), que é uma prática mais robusta |
| **DLQ** | Manual (tópico separado + lógica de publish) | Nativa (Redrive Policy + maxReceiveCount) | Simplifica código significativamente |
| **Latência** | Sub-milissegundo (intra-cluster) | ~10-50ms (Long Polling com 20s wait) | Aceitável para processamento assíncrono |
| **Consumer Groups** | Nativos com rebalanceamento automático | Não existe; cada fila é independente | Para 2 consumers com tópicos distintos, 1 fila por consumer é equivalente |
| **Custo** | MSK: ~$300-500/mês (mínimo 2 brokers) ou self-managed | ~$0.40 por milhão de mensagens | Redução drástica de custo para este volume |
| **Operacional** | Alto (brokers, rebalance, partitions, retention) | Zero (fully managed, sem servidores) | Eliminação total de ops de broker |
| **Retenção** | Configurável (dias a semanas) | Máximo 14 dias | Suficiente; histórico longo pertence ao PostgreSQL |

### 1.5 Mudanças de Código por Camada

**Platform** (`internal/platform/`):

- Criar pacote `internal/platform/sqs/` com: `client.go`, `producer.go`, `consumer.go`, `ensure.go`, `health.go`
- O `SQSProducer` implementa `EventDispatcher` (mesma interface).
- O `SQSConsumer` substitui o loop `PollFetches` por `ReceiveMessage` com Long Polling + `DeleteMessage` após handler.Handle() bem-sucedido.
- Remover pacote `internal/platform/kafka/` inteiramente.

**Config** (`cmd/server/config/`):

- Substituir `KafkaConfig` por `SQSConfig` em `config.go`, `load.go`, `validate.go`.
- Novas env vars: `SQS_REGION`, `SQS_QUEUE_PREFIX`, `SQS_ENDPOINT`.
- Remover env vars: `KAFKA_BOOTSTRAP_SERVERS`, `KAFKA_AUTO_OFFSET_RESET`, `KAFKA_SASL_*`, `KAFKA_USE_TLS`.

**Bootstrap** (`cmd/server/bootstrap/`):

- `kafka.go` → `sqs.go` (init client + EnsureQueues)
- `kafka_consumers.go` → `sqs_consumers.go` (init consumers)
- `shutdown.go` — adaptar para fechar SQS consumers gracefully.

**Features** (`internal/modules/specialist/features/`):

- `create/event_listeners/validate_license/adapters/inbound/kafka/` → `.../sqs/`
- `update/event_listeners/update_data_repositories/adapters/inbound/kafka/` → `.../sqs/`
- Eliminar `publish_dlq.go` e `errors.go` do adapter Elasticsearch de update — a DLQ é gerenciada pelo SQS automaticamente (basta não deletar a mensagem).

**Testes**:

- `internal/commom/tests/event/kafka/setup.go` → `.../sqs/setup.go` com testcontainers **LocalStack**.
- Testes de integração dos handlers adaptam o setup de infraestrutura.

**Dependências Go**:

- Remover: `github.com/twmb/franz-go`, `github.com/twmb/franz-go/pkg/kadm`, testcontainers-go/modules/kafka
- Adicionar: `github.com/aws/aws-sdk-go-v2/service/sqs`, `github.com/aws/aws-sdk-go-v2/config`, `github.com/aws/aws-sdk-go-v2/credentials`

---

## Parte 2: Elasticsearch → AWS OpenSearch

### 2.1 Estado Atual

O serviço utiliza Elasticsearch 8.x via `go-elasticsearch/v8` para:

- **1 Índice**: `specialists` — criado no startup se não existir, com analyzers customizados (`name_analyzer`, `standard_analyzer`) e 12 campos mapeados.
- **Busca full-text**: queries `bool` + `multi_match` (fuzziness AUTO) + `wildcard` + `term`/`terms`, com paginação `search_after` e sort customizável.
- **Indexação event-driven**: o listener `update_data_repositories` consome eventos e indexa via `client.Index()` (upsert por document ID).
- **DLQ manual**: em caso de falha na indexação, publica evento DLQ no Kafka (será simplificado com SQS).

Abstrações existentes:

- `SpecialistSearchRepositoryInterface` — interface de busca no domínio.
- `DataRepository` — interface para repositórios de dados (Elasticsearch é uma implementação).
- Factory + IndexRegistry — pattern de inicialização desacoplado.

### 2.2 Compatibilidade OpenSearch com o Uso Atual

O AWS OpenSearch é um fork do Elasticsearch 7.10.2. Todas as features utilizadas neste serviço são **100% compatíveis**:

| Feature utilizada | Compatível? | Observação |
|-------------------|-------------|------------|
| `bool` query | Sim | Idêntico |
| `multi_match` com fuzziness | Sim | Idêntico |
| `wildcard` query | Sim | Idêntico |
| `term` / `terms` query | Sim | Idêntico |
| `search_after` pagination | Sim | Idêntico |
| Custom analyzers (`standard`, `custom` com `asciifolding`) | Sim | Idêntico |
| `keyword`, `text`, `float`, `boolean`, `date` field types | Sim | Idêntico |
| Index creation com settings/mappings | Sim | Idêntico |
| Document indexing (upsert by ID) | Sim | Idêntico |

Features do ES 8.x **não utilizadas** neste serviço (e ausentes no OpenSearch): kNN nativo, ES|QL, runtime fields v2. Impacto: nenhum.

### 2.3 Isolamento Multi-Tenant por Cluster

**Decisão**: utilizar um único cluster OpenSearch por ambiente (produção, staging, dev) compartilhado entre todos os serviços e plataformas, com isolamento lógico total via 4 camadas de segurança.

**Justificativa**:

- Reduz custo de infraestrutura (um cluster vs N clusters).
- Simplifica operação (upgrades, backups, monitoring centralizados).
- O OpenSearch possui mecanismos robustos de isolamento que eliminam risco de cruzamento de dados.

**As 4 camadas de isolamento são detalhadas no documento complementar `002-opensearch-sqs-terraform-requirements.md`.**

Resumo:

| Camada | Mecanismo | Protege contra |
|--------|-----------|----------------|
| 1. Convenção de nomes | Prefixo `{plataforma}-{índice}` | Conflito acidental de nomes |
| 2. OpenSearch FGAC | Roles com index patterns (`healing-*`) | Acesso indevido entre serviços |
| 3. IAM Role Mapping | IRSA → OpenSearch Role | Falsificação de identidade |
| 4. IAM Resource Policy | `es:ESHttp*` com Resource restrito | Defesa em profundidade |

### 2.4 Tradeoffs Técnicos

| Aspecto | Elasticsearch 8.x | AWS OpenSearch | Veredicto |
|---------|-------------------|----------------|-----------|
| **Query DSL** | Completo | Idêntico (baseado em ES 7.10 + extensões) | Sem impacto para o uso atual |
| **Client Go** | `go-elasticsearch/v8` | `opensearch-go/v4` | API similar, requer adaptação de imports e estilo de chamada |
| **Autenticação** | Basic Auth, API Keys | IAM SigV4 (recomendado), Basic Auth | IAM SigV4 é mais seguro e se integra com IRSA |
| **Analyzers customizados** | Todos disponíveis | Mesmos analyzers | Sem impacto |
| **Performance** | Depende do sizing manual | Idêntico para provisionado; Serverless auto-escala | Equivalente ou melhor |
| **Custo** | Self-managed (EC2, EBS, ops) | Managed ($0.10–$3.40/h por instância) | Troca ops por custo previsível |
| **Segurança** | X-Pack (licença paga) | FGAC incluído gratuitamente | Vantagem OpenSearch |
| **Multi-tenancy** | Requer X-Pack ou proxy | FGAC nativo + IAM | Vantagem significativa |
| **Operacional** | Patches, upgrades, snapshots manuais | Automatizados pela AWS | Redução drástica de ops |
| **Features exclusivas ES 8.x** | kNN, ES\|QL, runtime fields v2 | Ausentes | Sem impacto (não utilizadas) |
| **Vendor lock-in** | Neutro | AWS-specific | Tradeoff aceito em troca de ops reduzido |

### 2.5 Mudanças de Código por Camada

**Platform** (`internal/platform/elasticsearch/` → renomear para `internal/platform/opensearch/`):

- `client.go` — trocar import de `go-elasticsearch/v8` para `opensearch-go/v4`. Implementar autenticação via AWS SigV4 (usando `requestsigner` do opensearch-go).
- `factory.go` — adaptar tipo do client.
- `indexes/registry.go` — adaptar tipo do client. Adicionar suporte a `IndexPrefix` para multi-tenancy.
- `indexes/specialists.go` — adaptar tipo do client (mapping e analyzers permanecem idênticos).

**Features** — adaptar o tipo do client nos seguintes arquivos:

- `search/adapters/outbound/elasticsearch/repository.go` — adaptar chamada `client.Search()` para o estilo do opensearch-go (request objects ao invés de functional options).
- `search/adapters/outbound/elasticsearch/new.go` — trocar tipo do client.
- `update/.../adapters/outbound/elasticsearch/repository.go` — adaptar chamada `client.Index()`.
- `update/.../adapters/outbound/elasticsearch/new.go` — trocar tipo do client.
- `search/adapters/inbound/grpc_service/di.go` e `http_handler/di.go` — trocar tipo do client na struct de Dependencies.

**O que NÃO muda**:

- `builders.go` — Query DSL construído como `map[string]any` é idêntico.
- `mappers.go` — Response JSON structure é a mesma.
- `dto.go` — Structs de decode permanecem iguais.
- Interfaces de domínio (`SpecialistSearchRepositoryInterface`, `DataRepository`) — inalteradas.

**Config** (`cmd/server/config/`):

- Renomear `ElasticsearchConfig` para `OpenSearchConfig`.
- Substituir `Addresses`/`CloudID` por `Endpoint`/`Region`.
- Remover `Username`/`Password` se usar exclusivamente IAM SigV4.

**Bootstrap** (`cmd/server/bootstrap/`):

- `elasticsearch.go` → `opensearch.go`.
- Adaptar tipo de retorno da factory.

**Testes**:

- `internal/commom/tests/elasticsearch/setup.go` — trocar imagem Docker de `elasticsearch:8.17.0` para `opensearchproject/opensearch:2.x`.
- Adaptar configuração do container (env vars de segurança diferem).

**Dependências Go**:

- Remover: `github.com/elastic/go-elasticsearch/v8`, testcontainers-go/modules/elasticsearch
- Adicionar: `github.com/opensearch-project/opensearch-go/v4`, `github.com/aws/aws-sdk-go-v2/config`

---

## Parte 3: Ordem de Execução e Estimativa

### Complexidade

| Migração | Dificuldade | Estimativa | Risco |
|----------|------------|------------|-------|
| Elasticsearch → OpenSearch | Baixa | 1–2 dias | Baixo (DSL idêntico, só troca de client) |
| Kafka → SQS | Média | 3–5 dias | Médio (semântica diferente, DLQ nativa, polling model) |

### Ordem Recomendada

1. **OpenSearch primeiro** — menor risco, ganho rápido de confiança, valida o padrão IAM/IRSA que será reusado no SQS.
2. **SQS depois** — mudança mais profunda mas bem encapsulada pelas interfaces existentes (`EventDispatcher`, `Listener`).

### Checklist de Execução

**Fase 1: OpenSearch (1–2 dias)**

- [ ] Adicionar dependência `opensearch-go/v4` e remover `go-elasticsearch/v8`
- [ ] Adaptar `internal/platform/elasticsearch/` → `internal/platform/opensearch/`
- [ ] Implementar autenticação SigV4 no client
- [ ] Adaptar chamadas de API nos repositórios (search e update)
- [ ] Adicionar suporte a `IndexPrefix` no registry
- [ ] Adaptar testes de integração (imagem OpenSearch + setup)
- [ ] Atualizar config, bootstrap, DI
- [ ] Atualizar env vars, ConfigMap/Secret k8s
- [ ] Testar end-to-end localmente

**Fase 2: SQS (3–5 dias)**

- [ ] Adicionar dependência AWS SDK v2 SQS e remover franz-go
- [ ] Criar pacote `internal/platform/sqs/`
- [ ] Implementar `EnsureQueues` com criação idempotente + DLQ nativa
- [ ] Implementar `SQSProducer` (interface `EventDispatcher`)
- [ ] Implementar `SQSConsumer` (Long Polling + delete após sucesso)
- [ ] Adaptar event_listeners (kafka/ → sqs/)
- [ ] Remover lógica manual de DLQ do adapter Elasticsearch
- [ ] Adaptar testes de integração (LocalStack)
- [ ] Atualizar config, bootstrap, shutdown
- [ ] Atualizar env vars, ConfigMap/Secret k8s, IAM policies
- [ ] Testar end-to-end localmente

**Fase 3: Cleanup**

- [ ] Remover `internal/platform/kafka/` por completo
- [ ] Remover testes e setup antigos de Kafka e Elasticsearch
- [ ] Atualizar README, docs, Swagger description
- [ ] Atualizar regras `.cursor/rules/` e `.kiro/steering/`
- [ ] Executar suíte completa de testes
- [ ] Deploy em staging e validar

---

## Consequências

### Positivas

- Eliminação total de overhead operacional de Kafka e Elasticsearch.
- Unificação do modelo de autenticação via IAM (IRSA) para todos os serviços AWS.
- DLQ nativa do SQS simplifica código e aumenta confiabilidade.
- FGAC do OpenSearch permite multi-tenancy seguro em cluster compartilhado.
- Custo operacional significativamente menor.
- Alinhamento completo com o ecossistema AWS já utilizado (EKS, RDS, ECR).

### Negativas / Riscos Aceitos

- **Sem replay de eventos no SQS** — aceito; reindexação será feita a partir do PostgreSQL (source of truth).
- **Latência ligeiramente maior no SQS** (~20ms vs sub-ms) — aceitável para processamento assíncrono.
- **Vendor lock-in AWS** — aceito em troca de redução operacional; as interfaces de domínio permanecem agnósticas.
- **OpenSearch baseado em ES 7.10** — sem impacto para o DSL utilizado; features ES 8.x exclusivas não são necessárias.
