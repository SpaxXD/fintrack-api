# FinTrack API

> Personal finance REST API built with Go — Clean Architecture, JWT auth with refresh token rotation, PostgreSQL, and structured logging.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)

---

## Overview

FinTrack is a production-ready REST API for personal finance management. Users can track accounts, categorize transactions, and get financial summaries with monthly trend analysis.

Built as a portfolio project to demonstrate Go backend development at a professional level — with emphasis on clean architecture, testability, and operational concerns like graceful shutdown and structured logging.

---

## Features

- **Auth** — Register, login, and refresh token rotation (JWT, 15min access / 7 days refresh)
- **Rate limiting** — 5 failed login attempts trigger a 15-minute lockout
- **Accounts** — Create and manage financial accounts with atomic balance updates
- **Categories** — Custom categories per user for organizing transactions
- **Transactions** — Full CRUD with pagination and filtering by category and date range
- **Reports** — Financial summary with income/expense totals, per-category breakdown, and monthly trend analysis
- **Swagger UI** — Interactive API docs at `/swagger/`

---

## Tech stack

| Layer | Technology | Reason |
|---|---|---|
| Language | Go 1.25 | Performance, concurrency primitives, strong typing |
| Router | chi | Lightweight, idiomatic, production-proven |
| Database | PostgreSQL 16 | Reliability, UUID support, rich indexing |
| Queries | sqlc | Type-safe SQL without ORM magic |
| Migrations | golang-migrate | Versioned, reproducible schema changes |
| Auth | JWT + bcrypt | Industry-standard, stateless access tokens |
| Config | viper | 12-factor app, `.env` + env var support |
| Logging | zerolog | Structured JSON logs, zero allocation |
| Docs | swaggo | Auto-generated Swagger from annotations |
| Validation | go-playground/validator | Struct-based input validation with custom messages |
| Testing | pgregory.net/rapid | Property-based testing for correctness guarantees |
| Container | Docker multi-stage | Minimal final image, fast builds |

---

## Architecture

This project follows **Clean Architecture** with strict layer separation and dependency inversion — outer layers depend on inner layers, never the reverse.

```
cmd/
└── api/
    └── main.go              # Entrypoint, DI wiring, graceful shutdown

internal/
├── domain/                  # Entities, interfaces, typed errors (no external deps)
├── repository/              # PostgreSQL implementations (sqlc generated)
├── service/                 # Business rules, orchestration
├── handler/                 # HTTP handlers, request/response mapping
├── middleware/              # RequestID, Logger, Auth, Recovery
├── router/                  # chi router setup
└── config/                  # viper config loading

├── validator/               # Input validation + monetary conversion

migrations/                  # SQL migration files (golang-migrate)
docs/                        # Generated Swagger documentation
```

**Request lifecycle:**

```
HTTP Request
    → Recovery middleware
    → RequestID middleware
    → Logger middleware
    → Auth middleware (protected routes)
    → Handler (input validation)
    → Service (business rules)
    → Repository (database)
    → Handler (response mapping)
HTTP Response
```

---

## API endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/auth/register` | No | Register a new user |
| POST | `/api/v1/auth/login` | No | Login, returns token pair |
| POST | `/api/v1/auth/refresh` | No | Refresh access token |
| GET | `/api/v1/accounts` | Yes | List user accounts |
| POST | `/api/v1/accounts` | Yes | Create account |
| PUT | `/api/v1/accounts/{id}` | Yes | Update account |
| DELETE | `/api/v1/accounts/{id}` | Yes | Delete account |
| GET | `/api/v1/categories` | Yes | List categories |
| POST | `/api/v1/categories` | Yes | Create category |
| PUT | `/api/v1/categories/{id}` | Yes | Update category |
| DELETE | `/api/v1/categories/{id}` | Yes | Delete category |
| GET | `/api/v1/transactions` | Yes | List transactions (paginated) |
| POST | `/api/v1/transactions` | Yes | Create transaction |
| PUT | `/api/v1/transactions/{id}` | Yes | Update transaction |
| DELETE | `/api/v1/transactions/{id}` | Yes | Delete transaction |
| GET | `/api/v1/summary` | Yes | Financial summary |
| GET | `/api/v1/summary/categories` | Yes | Per-category breakdown |
| GET | `/api/v1/summary/trend` | Yes | Monthly trend |

Full interactive docs: `http://localhost:8080/swagger/`

---

## Getting started

### Requirements

- [Docker](https://www.docker.com/) and Docker Compose

### Run with Docker (recommended)

```bash
git clone https://github.com/muriloabranches/fintrack-api.git
cd fintrack-api

docker-compose up --build
```

The API will be available at `http://localhost:8080`.  
Migrations run automatically on startup.

### Run locally

```bash
# Copy and configure environment
cp .env.example .env

# Start only the database
docker-compose up -d postgres

# Apply migrations
make migrate-up

# Run the API
make run
```

### Makefile commands

```bash
make run            # Run the API locally
make build          # Build the binary
make test           # Run all tests
make lint           # Run golangci-lint
make migrate-up     # Apply pending migrations
make migrate-down   # Rollback last migration
make migrate-create # Create new migration (make migrate-create name=xxx)
make swagger        # Regenerate Swagger docs
```

---

## Quick test

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"senha12345"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"senha12345"}'

# Use the returned access_token
curl http://localhost:8080/api/v1/accounts \
  -H "Authorization: Bearer <access_token>"
```

---

## Architecture decisions

**sqlc over GORM** — Raw SQL with generated type-safe code. Avoids ORM magic, makes queries explicit and reviewable, and eliminates N+1 surprises.

**Typed domain errors** — Instead of `errors.New("something went wrong")`, every failure is a typed error that maps to a specific HTTP status code. Consistent error responses across the entire API.

**Refresh token rotation** — On each refresh, the old token is revoked and a new pair is issued. Prevents token reuse after theft.

**Monetary values as integers** — All amounts stored as cents (int64) to avoid floating point precision issues. Conversion happens only at the API boundary.

**Graceful shutdown** — The server listens for `SIGINT`/`SIGTERM` and waits for in-flight requests to complete before exiting, preventing data corruption.

**Soft deletes** — Transactions and accounts use `deleted_at` instead of hard deletes, preserving financial history and enabling audit trails.

---

## Database schema

```
users
  id (uuid, pk)
  email (unique), password_hash
  failed_attempts, locked_until
  created_at, updated_at

accounts
  id (uuid, pk), user_id (fk)
  name, type (checking/savings/credit_card/cash/investment)
  balance (bigint, stored in cents)
  created_at, updated_at, deleted_at

categories
  id (uuid, pk), user_id (fk)
  name, type (income/expense)
  created_at, updated_at, deleted_at

transactions
  id (uuid, pk), user_id (fk), account_id (fk), category_id (fk, nullable)
  type (income/expense), amount (bigint, cents)
  description, date
  created_at, updated_at, deleted_at

refresh_tokens
  id (uuid, pk), user_id (fk)
  token (unique), expires_at, revoked (boolean)
  created_at
```

---

## License

MIT
