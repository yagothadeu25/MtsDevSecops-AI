# Serasa Cyber Shield — Guia de Manutenção

## Índice

- [Origem do Projeto](#origem-do-projeto)
- [Mapa de Nomes e Identificadores](#mapa-de-nomes-e-identificadores)
- [Estrutura do Repositório](#estrutura-do-repositório)
- [Dependências Internas (libs/)](#dependências-internas-libs)
- [Imagens Docker](#imagens-docker)
- [Variáveis de Ambiente](#variáveis-de-ambiente)
- [Build e Deploy](#build-e-deploy)
- [Testes](#testes)
- [Kubernetes (EKS)](#kubernetes-eks)
- [Atualizando o Upstream](#atualizando-o-upstream)
- [Problemas Conhecidos](#problemas-conhecidos)
- [Referência Rápida de Comandos](#referência-rápida-de-comandos)

---

## Origem do Projeto

Este projeto é um fork do [PentAGI](https://github.com/vxcontrol/pentagi) (vxcontrol/pentagi), uma ferramenta de testes de penetração autônomos com IA. O fork foi renomeado para **Serasa Cyber Shield** e adaptado para uso interno da Serasa Experian.

### O que foi alterado em relação ao upstream

| Área | Mudança |
|---|---|
| Branding | "PentAGI" / "Monkey.G" → "Serasa Cyber Shield" em toda UI, README, favicons, logo |
| Dependências Go | 3 libs vxcontrol copiadas para `backend/libs/` com `replace` directives no `go.mod` |
| Imagens Docker | `vxcontrol/kali-linux` → `kalilinux/kali-rolling`, `vxcontrol/pgvector` → `pgvector/pgvector:pg16` |
| Dockerfile | Reescrito (~55 linhas), paths `/opt/serasacyber`, user `appuser`, labels Serasa |
| Certificados SSL | Subject atualizado para "Serasa Experian/CyberShield" |
| ECR | Script `scripts/mirror-to-ecr.sh` para espelhar imagens externas |

---

## Mapa de Nomes e Identificadores

O projeto tem 3 camadas de nomenclatura. Entender isso é essencial para manutenção:

### 1. Nome público (visível ao usuário)
**"Serasa Cyber Shield"** — usado em toda a UI, README, documentação.

### 2. Identificadores técnicos de infraestrutura (NÃO renomear)
Estes nomes são usados em Docker Compose, volumes, redes e banco de dados. Renomeá-los quebraria deployments existentes:

| Identificador | Onde é usado |
|---|---|
| `monkeyg-network` | Docker Compose (todas as stacks) |
| `monkeyg-data`, `monkeyg-ssl`, `monkeyg-ollama` | Docker Compose volumes |
| `monkeyg-postgres-data` | Docker Compose volume do PostgreSQL |
| `monkeygdb` | Nome do banco PostgreSQL |
| `MONKEYG_IMAGE` | Variável de env para imagem Docker |

### 3. Identificadores internos do código Go (NÃO renomear)
O código Go usa `pentagi` como module name e `PENTAGI_*` como prefixo de variáveis de ambiente. Renomear quebraria centenas de arquivos:

| Identificador | Onde é usado |
|---|---|
| `pentagi` | Go module name (`go.mod`) |
| `PENTAGI_LISTEN_IP`, `PENTAGI_LISTEN_PORT` | Variáveis de env do servidor |
| `PENTAGI_POSTGRES_USER`, `PENTAGI_POSTGRES_PASSWORD`, `PENTAGI_POSTGRES_DB` | Variáveis de env do PostgreSQL |
| `PENTAGI_DATA_DIR`, `PENTAGI_SSL_DIR`, `PENTAGI_DOCKER_*` | Variáveis de env de paths |
| `pentagi-*` | Nomes de recursos Kubernetes |

### Regra de ouro
> **UI e documentação** → "Serasa Cyber Shield"
> **Infraestrutura Docker/K8s** → manter nomes existentes (`monkeyg-*`, `pentagi-*`)
> **Código Go** → manter `pentagi` como module name e `PENTAGI_*` como env vars

---

## Estrutura do Repositório

```
monkeyg/
├── backend/                    # Backend Go
│   ├── cmd/                    # Binários (pentagi, ctester, etester, ftester, installer)
│   ├── libs/                   # ⚠️ Libs vxcontrol internalizadas (ver seção abaixo)
│   │   ├── langchaingo/        # Fork do langchaingo (~32MB)
│   │   ├── cloud/              # Fork do cloud (~776KB)
│   │   └── graphiti-go-client/ # Fork do graphiti-go-client (~40KB)
│   ├── pkg/                    # Pacotes principais
│   ├── docs/                   # Documentação técnica interna (ainda referencia "PentAGI")
│   ├── go.mod                  # Module pentagi + replace directives para libs/
│   └── go.sum
├── frontend/                   # Frontend React + TypeScript
│   ├── src/
│   ├── scripts/generate-ssl.ts # Geração de certificados SSL dev
│   └── package.json            # name: "serasa-cyber-shield"
├── k8s/                        # Manifests Kubernetes (EKS)
├── scripts/                    # Scripts auxiliares + cópia original do upstream
│   ├── entrypoint.sh           # Entrypoint do container
│   ├── mirror-to-ecr.sh        # Script para espelhar imagens no ECR
│   └── Dockerfile              # Dockerfile original do upstream (referência)
├── docs/                       # Documentação do projeto Serasa
│   └── MANUTENCAO.md           # Este arquivo
├── reports/                    # Relatórios de pentest
├── Dockerfile                  # Dockerfile otimizado para Serasa
├── docker-compose.yml          # Stack principal
├── docker-compose-langfuse.yml # Stack Langfuse
├── docker-compose-graphiti.yml # Stack Graphiti + Neo4j
├── docker-compose-observability.yml # Stack de monitoramento
├── .env.example                # Template de variáveis de ambiente
└── README.md                   # Documentação principal (pt-BR)
```

---

## Dependências Internas (libs/)

As 3 bibliotecas Go do vxcontrol foram copiadas para `backend/libs/` para eliminar dependência externa:

| Lib | Módulo original | Tamanho | Função |
|---|---|---|---|
| `langchaingo` | `github.com/vxcontrol/langchaingo` | ~32MB | Framework LLM (fork do tmc/langchaingo) |
| `cloud` | `github.com/vxcontrol/cloud` | ~776KB | Utilitários cloud |
| `graphiti-go-client` | `github.com/vxcontrol/graphiti-go-client` | ~40KB | Cliente Go para Graphiti API |

### Como funciona

No `backend/go.mod`, há `replace` directives que redirecionam os imports para os paths locais:

```go
replace (
    github.com/vxcontrol/langchaingo => ./libs/langchaingo
    github.com/vxcontrol/cloud => ./libs/cloud
    github.com/vxcontrol/graphiti-go-client => ./libs/graphiti-go-client
)
```

### Atualizando uma lib

1. Baixe a nova versão: `go get github.com/vxcontrol/langchaingo@vX.Y.Z`
2. Copie do cache: `cp -r $(go env GOMODCACHE)/github.com/vxcontrol/langchaingo@vX.Y.Z backend/libs/langchaingo`
3. Torne editável: `chmod -R u+w backend/libs/langchaingo`
4. Atualize a versão no `go.mod` se necessário
5. Rode: `go mod tidy && go build ./...`

---

## Imagens Docker

### Imagens utilizadas

| Imagem | Uso | Configurável via |
|---|---|---|
| `kalilinux/kali-rolling` | Container de pentest (padrão) | `DOCKER_DEFAULT_IMAGE_FOR_PENTEST` |
| `debian:latest` | Container de tarefas gerais | `DOCKER_DEFAULT_IMAGE` |
| `pgvector/pgvector:pg16` | PostgreSQL com pgvector | Hardcoded no docker-compose |
| `docker:27-dind` | Docker-in-Docker (K8s) | Hardcoded no K8s manifest |
| `serasacyber:latest` | Aplicação principal | `MONKEYG_IMAGE` |

### Espelhamento para ECR

Para ambientes sem acesso direto ao Docker Hub:

```bash
# Configurar e rodar
./scripts/mirror-to-ecr.sh
```

ECR: `110380501820.dkr.ecr.us-east-1.amazonaws.com/pentagi/monkeyg/hml`

---

## Variáveis de Ambiente

### Variáveis obrigatórias

| Variável | Descrição |
|---|---|
| `OPEN_AI_KEY` ou `ANTHROPIC_API_KEY` ou `GEMINI_API_KEY` | Pelo menos 1 provedor LLM |
| `PENTAGI_POSTGRES_PASSWORD` | Senha do PostgreSQL |
| `COOKIE_SIGNING_SALT` | Salt para cookies de sessão |

### Variáveis de infraestrutura

| Variável | Padrão | Descrição |
|---|---|---|
| `PENTAGI_LISTEN_IP` | `127.0.0.1` | IP de escuta |
| `PENTAGI_LISTEN_PORT` | `8443` | Porta HTTPS |
| `PENTAGI_POSTGRES_USER` | `postgres` | Usuário PostgreSQL |
| `PENTAGI_POSTGRES_PASSWORD` | `postgres` | Senha PostgreSQL |
| `PENTAGI_POSTGRES_DB` | `monkeygdb` | Nome do banco |
| `MONKEYG_IMAGE` | `serasacyber:latest` | Imagem Docker da aplicação |

### Variáveis opcionais

| Variável | Descrição |
|---|---|
| `BEDROCK_REGION` | Região AWS para Bedrock |
| `TAVILY_API_KEY` | API key do Tavily (busca) |
| `GRAPHITI_ENABLED` | Habilitar grafo de conhecimento |
| `LANGFUSE_BASE_URL` | URL do Langfuse |

> Consulte `.env.example` para a lista completa.

---

## Build e Deploy

### Build local

```bash
# Backend
cd backend && go build ./...

# Frontend
cd frontend && npm install && npm run build

# Docker image
source ./scripts/version.sh
docker build \
  --build-arg PACKAGE_VER=$PACKAGE_VER \
  --build-arg PACKAGE_REV=$PACKAGE_REV \
  --platform linux/amd64 \
  -t serasacyber:$PACKAGE_VER .
```

### Deploy com Docker Compose

```bash
cp .env.example .env
# Editar .env com as chaves necessárias
docker compose up -d
```

### Deploy no EKS

```bash
# Build e push para ECR
docker build --platform linux/amd64 -t 110380501820.dkr.ecr.us-east-1.amazonaws.com/pentagi/monkeyg/hml:latest .
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 110380501820.dkr.ecr.us-east-1.amazonaws.com
docker push 110380501820.dkr.ecr.us-east-1.amazonaws.com/pentagi/monkeyg/hml:latest

# Aplicar manifests
kubectl apply -f k8s/ --insecure-skip-tls-verify
kubectl rollout restart deployment/pentagi-app -n pentagi --insecure-skip-tls-verify
```

> **Nota:** `--insecure-skip-tls-verify` é necessário por causa do proxy McAfee Web Gateway que intercepta TLS.

---

## Testes

### Backend

```bash
cd backend && go test ./...
```

### Falhas conhecidas (pré-existentes do upstream)

| Pacote | Falha | Causa |
|---|---|---|
| `pkg/controller` | build failed | Mocks desatualizados (faltam `APITokenCreated`, `GetAssistantExecutor`) |
| `cmd/installer` | TestValidateEnvPath, TestCreateEmptyEnvFile | Embedded provider não inicializado |
| `cmd/installer/files` | TestList | Arquivos de observability não encontrados |

Estas falhas existem no código original do upstream e não foram introduzidas pelas nossas mudanças.

### Frontend

```bash
cd frontend && npm run build   # Validação de compilação
cd frontend && npm run test    # Testes unitários (cobertura baixa)
```

### Testando agentes LLM

```bash
cd backend && go run cmd/ctester/*.go -verbose
```

---

## Kubernetes (EKS)

### Cluster

- **Conta AWS:** 110380501820
- **Região:** us-east-1
- **Domínio:** `pentagi-hml.pagueveloz.com.br`
- **Namespace:** `pentagi`

### Manifests (k8s/)

| Arquivo | Recurso |
|---|---|
| `01-namespace.yaml` | Namespace `pentagi` |
| `02-configmap.yaml` | ConfigMap com variáveis de ambiente |
| `03-secrets.yaml` | Secrets (API keys, senhas) |
| `04-pgvector.yaml` | StatefulSet do PostgreSQL + pgvector |
| `05-pgvector-pvc.yaml` | PersistentVolumeClaim |
| `06-pentagi-app.yaml` | Deployment (app + DinD sidecar) |
| `07-ingress.yaml` | ALB Ingress com sticky sessions |
| `08-provider-config.yaml` | ConfigMap de provedores LLM |

### Particularidades do ALB

O Ingress usa ALB com anotações específicas:

```yaml
alb.ingress.kubernetes.io/healthcheck-success-codes: "200-399"  # ALB retorna 301
alb.ingress.kubernetes.io/load-balancer-attributes: idle_timeout.timeout_seconds=3600
alb.ingress.kubernetes.io/target-group-attributes: stickiness.enabled=true,stickiness.lb_cookie.duration_seconds=86400
```

### Limitações conhecidas

- **WebSocket via ALB:** GraphQL subscriptions (WebSocket) não funcionam perfeitamente via ALB. A aplicação funciona mas sem atualizações em tempo real.
- **DinD privilegiado:** O sidecar Docker-in-Docker roda como `privileged: true` — risco de segurança aceito para HML.

---

## Atualizando o Upstream

Para incorporar mudanças do upstream (vxcontrol/pentagi):

1. Adicione o remote:
   ```bash
   git remote add upstream https://github.com/vxcontrol/pentagi.git
   ```

2. Fetch e merge:
   ```bash
   git fetch upstream
   git merge upstream/master --no-commit
   ```

3. Resolva conflitos priorizando:
   - **Nossos:** README, Dockerfile, branding, `.env.example`, `go.mod` (replace directives)
   - **Upstream:** Código Go em `pkg/`, `cmd/`, queries SQL, frontend components

4. Verifique:
   ```bash
   cd backend && go mod tidy && go build ./... && go test ./...
   cd ../frontend && npm install && npm run build
   ```

5. Atenção especial:
   - Se o upstream atualizar as libs vxcontrol, atualize também em `backend/libs/`
   - Se o upstream adicionar novas variáveis `PENTAGI_*`, adicione no `.env.example`
   - Se o upstream mudar interfaces Go, os testes de `pkg/controller` podem precisar de atualização nos mocks

---

## Problemas Conhecidos

### Proxy McAfee Web Gateway
O proxy corporativo intercepta tráfego HTTPS, causando erros de TLS com kubectl e Docker. Workarounds:
- kubectl: `--insecure-skip-tls-verify`
- Docker: Adicionar `McAfee.pem` ao trust store

### Credenciais padrão
A aplicação vem com `admin@serasacyber.com` / `admin`. **Alterar imediatamente em produção.**

### Sem CI/CD
Não há pipeline de CI/CD configurado. Build e deploy são manuais.

### Documentação interna do backend
Os arquivos em `backend/docs/` ainda referenciam "PentAGI" extensivamente. São documentação técnica interna do upstream e não são visíveis ao usuário final.

---

## Referência Rápida de Comandos

```bash
# Build backend
cd backend && go build ./...

# Build frontend
cd frontend && npm run build

# Build Docker image (amd64)
docker build --platform linux/amd64 -t serasacyber:dev .

# Subir stack local
docker compose up -d

# Subir com Langfuse + Observability
docker compose -f docker-compose.yml -f docker-compose-langfuse.yml -f docker-compose-observability.yml up -d

# Testes backend
cd backend && go test ./...

# Push para ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 110380501820.dkr.ecr.us-east-1.amazonaws.com
docker tag serasacyber:dev 110380501820.dkr.ecr.us-east-1.amazonaws.com/pentagi/monkeyg/hml:latest
docker push 110380501820.dkr.ecr.us-east-1.amazonaws.com/pentagi/monkeyg/hml:latest

# Deploy K8s
kubectl apply -f k8s/ --insecure-skip-tls-verify
kubectl rollout restart deployment/pentagi-app -n pentagi --insecure-skip-tls-verify

# Logs do pod
kubectl logs -f deployment/pentagi-app -c pentagi-app -n pentagi --insecure-skip-tls-verify

# Espelhar imagens para ECR
./scripts/mirror-to-ecr.sh
```
