# FinTrack API

> API REST de finanças pessoais construída com Go — Clean Architecture, autenticação JWT com rotação de refresh token, PostgreSQL e logging estruturado.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)

---

## Visão geral

FinTrack é uma API REST pronta para produção voltada ao gerenciamento de finanças pessoais. Usuários podem acompanhar contas, categorizar transações e obter resumos financeiros com análise de tendência mensal.

Construído como projeto de portfólio para demonstrar desenvolvimento backend em Go em nível profissional — com ênfase em clean architecture, testabilidade e preocupações operacionais como graceful shutdown e logging estruturado.

---

## Funcionalidades

- **Autenticação** — Registro, login e rotação de refresh token (JWT, access 15min / refresh 7 dias)
- **Rate limiting** — 5 tentativas de login falhas ativam bloqueio de 15 minutos
- **Contas** — Criação e gerenciamento de contas financeiras com atualização atômica de saldo
- **Categorias** — Categorias personalizadas por usuário para organizar transações
- **Transações** — CRUD completo com paginação e filtragem por categoria e intervalo de datas
- **Relatórios** — Resumo financeiro com totais de receita/despesa, detalhamento por categoria e análise de tendência mensal
- **Swagger UI** — Documentação interativa da API em `/swagger/`

---

## Stack tecnológica

| Camada | Tecnologia | Motivo |
|---|---|---|
| Linguagem | Go 1.25 | Performance, primitivas de concorrência, tipagem forte |
| Router | chi | Leve, idiomático, comprovado em produção |
| Banco de dados | PostgreSQL 16 | Confiabilidade, suporte a UUID, indexação avançada |
| Queries | sqlc | SQL type-safe sem magia de ORM |
| Migrações | golang-migrate | Alterações de schema versionadas e reproduzíveis |
| Autenticação | JWT + bcrypt | Padrão da indústria, access tokens stateless |
| Configuração | viper | 12-factor app, suporte a `.env` + variáveis de ambiente |
| Logging | zerolog | Logs JSON estruturados, zero allocation |
| Documentação | swaggo | Swagger gerado automaticamente a partir de anotações |
| Validação | go-playground/validator | Validação de input baseada em structs com mensagens customizadas |
| Testes | pgregory.net/rapid | Property-based testing para garantias de corretude |
| Container | Docker multi-stage | Imagem final mínima, builds rápidos |

---

## Arquitetura

Este projeto segue **Clean Architecture** com separação estrita de camadas e inversão de dependência — camadas externas dependem das internas, nunca o inverso.

```
cmd/
└── api/
    └── main.go              # Entrypoint, injeção de dependência, graceful shutdown

internal/
├── domain/                  # Entidades, interfaces, erros tipados (sem deps externas)
├── repository/              # Implementações PostgreSQL (gerado pelo sqlc)
├── service/                 # Regras de negócio, orquestração
├── handler/                 # Handlers HTTP, mapeamento request/response
├── middleware/              # RequestID, Logger, Auth, Recovery
├── router/                  # Configuração do router chi
├── config/                  # Carregamento de config via viper
└── validator/               # Validação de input + conversão monetária

migrations/                  # Arquivos de migração SQL (golang-migrate)
docs/                        # Documentação Swagger gerada
```

**Ciclo de vida de uma requisição:**

```
Requisição HTTP
    → Middleware Recovery
    → Middleware RequestID
    → Middleware Logger
    → Middleware Auth (rotas protegidas)
    → Handler (validação de input)
    → Service (regras de negócio)
    → Repository (banco de dados)
    → Handler (mapeamento de resposta)
Resposta HTTP
```

---

## Endpoints da API

| Método | Path | Auth | Descrição |
|---|---|---|---|
| POST | `/api/v1/auth/register` | Não | Registrar novo usuário |
| POST | `/api/v1/auth/login` | Não | Login, retorna par de tokens |
| POST | `/api/v1/auth/refresh` | Não | Renovar access token |
| GET | `/api/v1/accounts` | Sim | Listar contas do usuário |
| POST | `/api/v1/accounts` | Sim | Criar conta |
| PUT | `/api/v1/accounts/{id}` | Sim | Atualizar conta |
| DELETE | `/api/v1/accounts/{id}` | Sim | Excluir conta |
| GET | `/api/v1/categories` | Sim | Listar categorias |
| POST | `/api/v1/categories` | Sim | Criar categoria |
| PUT | `/api/v1/categories/{id}` | Sim | Atualizar categoria |
| DELETE | `/api/v1/categories/{id}` | Sim | Excluir categoria |
| GET | `/api/v1/transactions` | Sim | Listar transações (paginado) |
| POST | `/api/v1/transactions` | Sim | Criar transação |
| PUT | `/api/v1/transactions/{id}` | Sim | Atualizar transação |
| DELETE | `/api/v1/transactions/{id}` | Sim | Excluir transação |
| GET | `/api/v1/summary` | Sim | Resumo financeiro |
| GET | `/api/v1/summary/categories` | Sim | Detalhamento por categoria |
| GET | `/api/v1/summary/trend` | Sim | Tendência mensal |

Documentação interativa completa: `http://localhost:8080/swagger/`

---

## Como começar

### Requisitos

- [Docker](https://www.docker.com/) e Docker Compose

### Rodar com Docker (recomendado)

```bash
git clone https://github.com/muriloabranches/fintrack-api.git
cd fintrack-api

docker-compose up --build
```

A API estará disponível em `http://localhost:8080`.  
As migrações são aplicadas automaticamente na inicialização.

### Rodar localmente

```bash
# Copiar e configurar o ambiente
cp .env.example .env

# Subir apenas o banco de dados
docker-compose up -d postgres

# Aplicar migrações
make migrate-up

# Rodar a API
make run
```

### Comandos do Makefile

```bash
make run            # Rodar a API localmente
make build          # Compilar o binário
make test           # Executar todos os testes
make lint           # Executar golangci-lint
make migrate-up     # Aplicar migrações pendentes
make migrate-down   # Reverter última migração
make migrate-create # Criar nova migração (make migrate-create name=xxx)
make swagger        # Regenerar documentação Swagger
```

---

## Teste rápido

```bash
# Registrar
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"senha12345"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"senha12345"}'

# Usar o access_token retornado
curl http://localhost:8080/api/v1/accounts \
  -H "Authorization: Bearer <access_token>"
```

---

## Decisões de arquitetura

**sqlc ao invés de GORM** — SQL puro com código type-safe gerado. Evita magia de ORM, torna queries explícitas e revisáveis, e elimina surpresas de N+1.

**Erros de domínio tipados** — Ao invés de `errors.New("algo deu errado")`, toda falha é um erro tipado que mapeia para um código HTTP específico. Respostas de erro consistentes em toda a API.

**Rotação de refresh token** — A cada refresh, o token antigo é revogado e um novo par é emitido. Previne reuso de token após roubo.

**Valores monetários como inteiros** — Todos os valores armazenados em centavos (int64) para evitar problemas de precisão de ponto flutuante. A conversão acontece apenas na fronteira da API.

**Graceful shutdown** — O servidor escuta `SIGINT`/`SIGTERM` e aguarda requisições em andamento serem completadas antes de encerrar, prevenindo corrupção de dados.

**Soft deletes** — Transações e contas usam `deleted_at` ao invés de exclusão permanente, preservando histórico financeiro e possibilitando trilhas de auditoria.

---

## Schema do banco de dados

```
users
  id (uuid, pk)
  email (unique), password_hash
  failed_attempts, locked_until
  created_at, updated_at

accounts
  id (uuid, pk), user_id (fk)
  name, type (checking/savings/credit_card/cash/investment)
  balance (bigint, armazenado em centavos)
  created_at, updated_at, deleted_at

categories
  id (uuid, pk), user_id (fk)
  name, type (income/expense)
  created_at, updated_at, deleted_at

transactions
  id (uuid, pk), user_id (fk), account_id (fk), category_id (fk, nullable)
  type (income/expense), amount (bigint, centavos)
  description, date
  created_at, updated_at, deleted_at

refresh_tokens
  id (uuid, pk), user_id (fk)
  token (unique), expires_at, revoked (boolean)
  created_at
```

---

## Licença

MIT
