# ADR-002: Requisitos de Infraestrutura Terraform — OpenSearch e SQS

**Status:** Proposta  
**Data:** 2026-03-29  
**Autores:** Equipe Healing  
**Referência:** ADR-001 (Migração Kafka→SQS e Elasticsearch→OpenSearch)

---

## Contexto

Este documento detalha os requisitos de infraestrutura que devem ser implementados no projeto Terraform para suportar a migração descrita na ADR-001. Cobre a configuração do cluster OpenSearch com isolamento multi-tenant via FGAC, as IAM Roles e Policies necessárias para SQS e OpenSearch, e os recursos complementares.

O foco é em **o que criar** e **por que**, não em código Terraform (que será implementado no repositório de infraestrutura).

---

## Parte 1: AWS OpenSearch — Cluster e Isolamento Multi-Tenant

### 1.1 Cluster OpenSearch (um por ambiente)

Criar um domínio OpenSearch por ambiente com as seguintes características:

| Propriedade | Produção | Staging | Dev |
|-------------|----------|---------|-----|
| Engine version | OpenSearch 2.x (última estável) | Igual | Igual |
| Instance type | `r6g.large.search` (ou superior conforme carga) | `t3.medium.search` | `t3.small.search` |
| Instance count | 3 (multi-AZ) | 2 | 1 |
| Storage | GP3 EBS, 100GB+ | GP3 EBS, 50GB | GP3 EBS, 20GB |
| Multi-AZ | Sim (Zone Awareness) | Opcional | Não |
| Encryption at rest | Sim (KMS) | Sim | Sim |
| Node-to-node encryption | Sim | Sim | Sim |
| HTTPS enforcement | Sim | Sim | Sim |
| Fine-Grained Access Control | **Habilitado** | **Habilitado** | **Habilitado** |
| Master user type | IAM ARN (não usar internal database como master) | Igual | Igual |
| VPC | Sim (subnets privadas do EKS) | Igual | Igual |
| Automated snapshots | Sim (daily) | Sim | Opcional |

**Recurso Terraform**: `aws_opensearch_domain`

O FGAC (Fine-Grained Access Control) **deve estar habilitado em todos os ambientes** — inclusive dev — para garantir que o comportamento de autenticação e autorização seja testado desde o início.

### 1.2 Convenção de Nomes de Índices

Cada serviço/plataforma deve prefixar seus índices com o nome da plataforma:

| Plataforma | Serviço | Padrão de índice |
|-----------|---------|------------------|
| Healing | specialist | `healing-specialists` |
| Healing | appointment (futuro) | `healing-appointments` |
| Healing | patient (futuro) | `healing-patients` |
| MES (exemplo futuro) | workorder | `mes-workorders` |
| MES (exemplo futuro) | equipment | `mes-equipment` |

Esta convenção é **enforced** pelas OpenSearch Roles (seção 1.3) e pelas IAM Policies (seção 1.5).

### 1.3 OpenSearch Roles (FGAC — Camada 2 de isolamento)

Para cada plataforma/serviço que compartilha o cluster, criar uma OpenSearch Security Role que restringe o acesso a um index pattern específico.

**Role: `healing_readwrite`**

```json
{
  "cluster_permissions": [
    "cluster_composite_ops_ro"
  ],
  "index_permissions": [
    {
      "index_patterns": [
        "healing-*"
      ],
      "allowed_actions": [
        "crud",
        "create_index",
        "manage",
        "indices:data/read/search",
        "indices:data/write/index",
        "indices:data/write/bulk",
        "indices:admin/mapping/put",
        "indices:admin/create",
        "indices:admin/exists"
      ]
    }
  ],
  "tenant_permissions": []
}
```

**Role: `mes_readwrite`** (exemplo futuro)

```json
{
  "cluster_permissions": [
    "cluster_composite_ops_ro"
  ],
  "index_permissions": [
    {
      "index_patterns": [
        "mes-*"
      ],
      "allowed_actions": [
        "crud",
        "create_index",
        "manage",
        "indices:data/read/search",
        "indices:data/write/index",
        "indices:data/write/bulk",
        "indices:admin/mapping/put",
        "indices:admin/create",
        "indices:admin/exists"
      ]
    }
  ],
  "tenant_permissions": []
}
```

**O que cada permissão significa**:

| Permissão | Necessidade |
|-----------|-------------|
| `cluster_composite_ops_ro` | Permite operações read-only de cluster (necessário para health checks, cat APIs) |
| `crud` | Read + write em documentos dos índices permitidos |
| `create_index` | Permite que a app crie índices no startup (com o prefixo correto) |
| `manage` | Permite gerenciar settings/mappings dos índices próprios |
| `indices:admin/exists` | Permite verificar se um índice existe (usado no startup) |

**O que está explicitamente negado** (por omissão):

- Acesso a índices de outras plataformas (ex: `mes-*` para a role `healing_readwrite`)
- Acesso a índices de sistema (`.opendistro_security`, `.kibana`, etc.)
- Criação de roles, role mappings, ou configurações de segurança
- Operações de cluster destrutivas (delete index de outro prefixo, snapshot/restore)

**Recurso Terraform**: `opensearch_role` (provider `opensearch`)

### 1.4 OpenSearch Role Mappings (FGAC — Camada 3 de isolamento)

Cada IAM Role de serviço é mapeada para a OpenSearch Role correspondente. Isso conecta a identidade AWS (quem o pod é) com as permissões OpenSearch (o que o pod pode fazer).

**Role Mapping: `healing_readwrite`**

```json
{
  "backend_roles": [
    "arn:aws:iam::ACCOUNT_ID:role/healing-specialist-pod-role"
  ],
  "hosts": [],
  "users": []
}
```

**Role Mapping: `mes_readwrite`** (exemplo futuro)

```json
{
  "backend_roles": [
    "arn:aws:iam::ACCOUNT_ID:role/mes-workorder-pod-role"
  ],
  "hosts": [],
  "users": []
}
```

**Como funciona o fluxo de autenticação**:

1. Pod no EKS assume IAM Role via IRSA (ServiceAccount anotado com role ARN).
2. App assina requests HTTP com SigV4 usando as credenciais IAM temporárias.
3. OpenSearch recebe o request, extrai o IAM Role ARN do header de autenticação.
4. OpenSearch consulta os Role Mappings, encontra que aquele ARN pertence à role `healing_readwrite`.
5. OpenSearch aplica as permissões da role: qualquer operação em `healing-*` é permitida; qualquer outra coisa retorna **403 Forbidden**.

**Adicionar novos serviços ao cluster** requer apenas:

1. Criar uma nova OpenSearch Role com o index pattern correto (ex: `billing-*`).
2. Criar o Role Mapping vinculando a IAM Role do novo serviço à OpenSearch Role.
3. O novo serviço já pode usar o cluster com isolamento completo.

**Recurso Terraform**: `opensearch_roles_mapping` (provider `opensearch`)

### 1.5 IAM Resource Policy no Domínio OpenSearch (Camada 4 de isolamento)

Além do FGAC interno, aplicar uma Access Policy no próprio domínio OpenSearch que restringe quais IAM principals podem sequer alcançar o cluster.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::ACCOUNT_ID:role/healing-specialist-pod-role",
          "arn:aws:iam::ACCOUNT_ID:role/mes-workorder-pod-role",
          "arn:aws:iam::ACCOUNT_ID:role/opensearch-admin-role"
        ]
      },
      "Action": "es:ESHttp*",
      "Resource": "arn:aws:es:REGION:ACCOUNT_ID:domain/shared-cluster/*"
    }
  ]
}
```

**Observações**:

- Cada novo serviço que precisar acessar o cluster deve ser adicionado como principal nesta policy.
- A `opensearch-admin-role` é usada para operações administrativas (criar roles, mappings, etc. via Terraform ou CI/CD).
- Esta policy atua como defesa em profundidade: mesmo que o FGAC tivesse uma falha, IAM bloqueia principals não autorizados.

**Recurso Terraform**: `aws_opensearch_domain_policy`

### 1.6 VPC e Security Groups

O domínio OpenSearch deve estar em subnets privadas dentro da mesma VPC do cluster EKS.

**Security Group do OpenSearch**:

| Regra | Tipo | Protocolo | Porta | Origem |
|-------|------|-----------|-------|--------|
| Inbound | HTTPS | TCP | 443 | Security Group dos worker nodes do EKS |
| Outbound | All | All | All | 0.0.0.0/0 |

Isso garante que apenas pods rodando no EKS conseguem alcançar o cluster OpenSearch na rede.

**Recursos Terraform**: `aws_security_group`, `aws_security_group_rule`

### 1.7 OpenSearch Admin Role (para Terraform)

O Terraform precisa de uma IAM Role com permissões administrativas no OpenSearch para criar roles, role mappings e configurar o domínio. Esta role é usada **apenas pelo Terraform**, nunca por aplicações.

**Permissões**:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "es:ESHttp*",
        "es:Describe*",
        "es:List*",
        "es:Create*",
        "es:Update*",
        "es:Delete*"
      ],
      "Resource": "arn:aws:es:REGION:ACCOUNT_ID:domain/shared-cluster/*"
    }
  ]
}
```

No OpenSearch FGAC, esta role deve ser configurada como **master user** ou mapeada para a role `all_access` interna.

### 1.8 Resumo Visual do Isolamento

```
┌─────────────────────────────────────────────────────────┐
│                    AWS Account                          │
│                                                         │
│  ┌──────────────────────────────────────────────────┐   │
│  │  VPC (subnets privadas)                          │   │
│  │                                                  │   │
│  │  ┌────────────┐       ┌──────────────────────┐   │   │
│  │  │  EKS       │       │  OpenSearch Cluster   │   │   │
│  │  │            │       │                      │   │   │
│  │  │ healing-   │──443──│ FGAC enforced:       │   │   │
│  │  │ specialist │  SigV4│  healing-* ✓         │   │   │
│  │  │ pod        │       │  mes-*     ✗ (403)   │   │   │
│  │  │            │       │                      │   │   │
│  │  │ mes-       │──443──│ FGAC enforced:       │   │   │
│  │  │ workorder  │  SigV4│  mes-*     ✓         │   │   │
│  │  │ pod        │       │  healing-* ✗ (403)   │   │   │
│  │  │            │       │                      │   │   │
│  │  └────────────┘       └──────────────────────┘   │   │
│  │       │                        │                 │   │
│  │       │ IRSA                   │ Access Policy   │   │
│  │       ▼                        ▼                 │   │
│  │  ┌────────────┐       ┌──────────────────────┐   │   │
│  │  │  IAM Roles │       │  Role Mappings       │   │   │
│  │  │            │       │                      │   │   │
│  │  │ healing-   │◄──────│ → healing_readwrite  │   │   │
│  │  │ specialist │       │   (healing-*)        │   │   │
│  │  │ -pod-role  │       │                      │   │   │
│  │  │            │       │                      │   │   │
│  │  │ mes-       │◄──────│ → mes_readwrite      │   │   │
│  │  │ workorder  │       │   (mes-*)            │   │   │
│  │  │ -pod-role  │       │                      │   │   │
│  │  └────────────┘       └──────────────────────┘   │   │
│  └──────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

---

## Parte 2: AWS SQS — IAM Policies para App-Level Queue Management

### 2.1 Modelo de Responsabilidade

Conforme decidido na ADR-001, as filas SQS são criadas pela própria aplicação no startup (pattern idempotente). O Terraform é responsável apenas pela IAM Role e Policy que autoriza o pod a gerenciar suas filas.

### 2.2 IAM Role por Serviço (IRSA)

Cada serviço no EKS recebe sua própria IAM Role via IRSA (IAM Roles for Service Accounts). Essa role é anotada no ServiceAccount do Kubernetes.

**Recurso Terraform**: `aws_iam_role` com trust policy para o OIDC provider do EKS.

**Trust Policy (padrão IRSA)**:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::ACCOUNT_ID:oidc-provider/oidc.eks.REGION.amazonaws.com/id/EKS_CLUSTER_ID"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "oidc.eks.REGION.amazonaws.com/id/EKS_CLUSTER_ID:sub": "system:serviceaccount:healing:healing-specialist",
          "oidc.eks.REGION.amazonaws.com/id/EKS_CLUSTER_ID:aud": "sts.amazonaws.com"
        }
      }
    }
  ]
}
```

O `Condition` garante que apenas o ServiceAccount `healing-specialist` no namespace `healing` pode assumir esta role.

### 2.3 SQS IAM Policy

A policy de SQS permite que o serviço crie e gerencie apenas filas com seu próprio prefixo.

**Policy: `healing-specialist-sqs-policy`**

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowQueueManagement",
      "Effect": "Allow",
      "Action": [
        "sqs:CreateQueue",
        "sqs:GetQueueUrl",
        "sqs:GetQueueAttributes",
        "sqs:SetQueueAttributes",
        "sqs:TagQueue"
      ],
      "Resource": "arn:aws:sqs:REGION:ACCOUNT_ID:specialist-*"
    },
    {
      "Sid": "AllowMessageOperations",
      "Effect": "Allow",
      "Action": [
        "sqs:SendMessage",
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage",
        "sqs:ChangeMessageVisibility"
      ],
      "Resource": "arn:aws:sqs:REGION:ACCOUNT_ID:specialist-*"
    },
    {
      "Sid": "AllowListQueues",
      "Effect": "Allow",
      "Action": [
        "sqs:ListQueues"
      ],
      "Resource": "*"
    }
  ]
}
```

**Observações**:

- O Resource `specialist-*` garante que o serviço só pode criar e operar filas com este prefixo. Ele **não pode** criar ou acessar filas de outros serviços.
- `sqs:ListQueues` requer `Resource: *` por limitação da API AWS, mas só retorna filas que o principal pode acessar.
- `sqs:SetQueueAttributes` é necessário caso o app precise atualizar atributos (ex: RedrivePolicy) em filas já existentes.
- `sqs:ChangeMessageVisibility` é útil para estender o visibility timeout em processamentos longos.

**Recurso Terraform**: `aws_iam_policy` + `aws_iam_role_policy_attachment`

### 2.4 OpenSearch IAM Policy (para o mesmo serviço)

O mesmo IAM Role do pod também recebe permissão para acessar o OpenSearch.

**Policy: `healing-specialist-opensearch-policy`**

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowOpenSearchAccess",
      "Effect": "Allow",
      "Action": [
        "es:ESHttpGet",
        "es:ESHttpPost",
        "es:ESHttpPut",
        "es:ESHttpHead",
        "es:ESHttpDelete"
      ],
      "Resource": "arn:aws:es:REGION:ACCOUNT_ID:domain/shared-cluster/*"
    }
  ]
}
```

Nota: esta policy permite acesso HTTP ao cluster, mas o FGAC interno do OpenSearch (roles + role mappings) é quem restringe os índices acessíveis. A combinação das duas camadas fornece defesa em profundidade.

**Recurso Terraform**: `aws_iam_policy` + `aws_iam_role_policy_attachment`

### 2.5 Kubernetes ServiceAccount

O ServiceAccount do pod deve ser anotado com o ARN da IAM Role.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: healing-specialist
  namespace: healing
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT_ID:role/healing-specialist-pod-role
```

Este ServiceAccount deve ser referenciado no Deployment do healing-specialist.

**Recurso Terraform**: pode ser gerenciado via Terraform (`kubernetes_service_account`) ou via manifesto K8s no repositório da aplicação. Recomendação: gerenciar no Terraform junto com a IAM Role para manter a associação explícita.

---

## Parte 3: Checklist de Recursos Terraform

### OpenSearch

| # | Recurso Terraform | Descrição |
|---|-------------------|-----------|
| 1 | `aws_opensearch_domain` | Cluster OpenSearch com FGAC habilitado, VPC, encryption |
| 2 | `aws_opensearch_domain_policy` | Access Policy restringindo principals autorizados |
| 3 | `aws_security_group` | SG para o domínio OpenSearch (inbound 443 do EKS) |
| 4 | `opensearch_role` (provider opensearch) | Role `healing_readwrite` com index pattern `healing-*` |
| 5 | `opensearch_roles_mapping` (provider opensearch) | Mapping IAM Role → OpenSearch Role |
| 6 | `aws_iam_role` | Admin role para Terraform gerenciar FGAC |

### SQS / IAM

| # | Recurso Terraform | Descrição |
|---|-------------------|-----------|
| 7 | `aws_iam_role` (IRSA) | Role do pod healing-specialist com trust policy OIDC |
| 8 | `aws_iam_policy` (SQS) | Permissões SQS restritas ao prefixo `specialist-*` |
| 9 | `aws_iam_policy` (OpenSearch) | Permissões `es:ESHttp*` no domínio OpenSearch |
| 10 | `aws_iam_role_policy_attachment` | Associar policies à IAM Role |
| 11 | `kubernetes_service_account` | ServiceAccount anotado com IAM Role ARN |

### Para cada novo serviço/plataforma, repetir

| # | Recurso | Observação |
|---|---------|------------|
| A | `aws_iam_role` (IRSA) | Trust policy com namespace e SA do novo serviço |
| B | `aws_iam_policy` (SQS) | Resource com o prefixo do novo serviço |
| C | `aws_iam_policy` (OpenSearch) | Mesma policy (FGAC restringe internamente) |
| D | `opensearch_role` | Index pattern do novo serviço (ex: `mes-*`) |
| E | `opensearch_roles_mapping` | Mapping da IAM Role do novo serviço |
| F | Atualizar `aws_opensearch_domain_policy` | Adicionar o novo principal |

---

## Parte 4: Provider Terraform para OpenSearch FGAC

Para gerenciar roles e role mappings dentro do OpenSearch (não apenas o domínio AWS), é necessário usar o **provider OpenSearch para Terraform**:

**Provider**: `opensearch-project/opensearch` (registry.terraform.io)

```hcl
terraform {
  required_providers {
    opensearch = {
      source  = "opensearch-project/opensearch"
      version = "~> 2.0"
    }
  }
}

provider "opensearch" {
  url         = aws_opensearch_domain.shared.endpoint
  aws_region  = var.region
  sign_aws_requests = true
}
```

Este provider permite gerenciar:

- `opensearch_role` — roles de segurança com index patterns
- `opensearch_roles_mapping` — mapeamentos IAM Role → OpenSearch Role
- `opensearch_index_template` — templates de índice (opcional, para enforçar defaults)
- `opensearch_ism_policy` — políticas de Index State Management (rotação, delete, etc.)

**Nota sobre ordem de criação**: o domínio OpenSearch deve existir antes de aplicar os recursos do provider `opensearch`. Isso pode requerer `depends_on` explícito ou aplicação em duas fases (`terraform apply -target=aws_opensearch_domain.shared` primeiro).

---

## Parte 5: Considerações de Ambiente

### Dev

- Cluster OpenSearch menor (1 nó `t3.small.search`).
- FGAC habilitado (para testar o fluxo completo de autenticação).
- Alternativamente, usar **LocalStack** para SQS e **container OpenSearch local** para desenvolvimento totalmente offline.
- As mesmas roles e mappings de produção devem existir em dev para evitar surpresas.

### Staging

- Configuração idêntica à produção, mas com instâncias menores.
- Mesmo FGAC, mesmas policies, mesmas roles.
- Serve como validação completa antes de deploy em produção.

### Produção

- Multi-AZ obrigatório (3 nós em 3 AZs).
- Dedicated master nodes recomendados se o cluster crescer além de 3 plataformas.
- Automated snapshots habilitados.
- CloudWatch alarms para cluster health (yellow/red), JVM memory pressure, storage.
- Slow log e audit log habilitados e enviados para CloudWatch Logs.
