# Dynamic CMS + API Builder (Go-native Backend Platform)

## Overview
A production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. This system acts as a backend infrastructure generator, similar to platforms like Hasura or Supabase, but is fully self-hosted and Go-native. It empowers users to connect databases, dynamically create data models, generate CRUD APIs, generate optional GraphQL APIs, manage schema migrations, monitor systems, manage backups, and scale services.

## 1. System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

```text
       Client Applications
                │
   ┌────────────┴─────────────┐
   ▼                          ▼
REST API                 GraphQL API
   │                          │
   └────────────┬─────────────┘
                ▼
           API Gateway
                │
                ▼
      Backend Platform Core
   ┌──────────────────────────┐
   │ ─ CMS Control Plane      │
   │ ─ Service Builder        │
   │ ─ CRUD Engine            │
   │ ─ Query Engine           │
   │ ─ Schema Migration Engine│
   │ ─ Backup Engine          │
   │ ─ Observability Engine   │
   └────────────┬─────────────┘
                ▼
       Database Connectors
   ┌────────────┼─────────────┐
   ▼            ▼             ▼
PostgreSQL    MySQL       MongoDB / SQLite
```

## 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. Tables defined in `internal/models/models.go` include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## 3. Go Project Structure

The project uses a clean modular structure:

```
.
├── cmd/server/main.go            # Application entrypoint
├── internal/
│   ├── config/                   # Configuration management
│   ├── database/                 # Connection manager & DB utilities
│   ├── graphql/                  # GraphQL Gateway
│   ├── handlers/                 # REST API Handlers (Presentation)
│   ├── middleware/               # Auth, Prometheus, Rate Limiter
│   ├── models/                   # Domain Models & Metadata Schema
│   ├── services/                 # Application Services (CRUD, Builder, Migration, etc.)
│   └── tracing/                  # OpenTelemetry logic
├── pkg/
│   ├── logger/                   # Zap Logger wrapper
│   └── response/                 # Gin HTTP response standardization
├── tests/
│   ├── integration/              # Integration tests
│   ├── uat/                      # End-to-End User Acceptance tests
│   └── unit/                     # Unit tests
├── docker/                       # Dockerfile and Docker compose resources
└── docs/                         # Swagger / OpenAPI documentation
```

## 4. CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go` and `internal/handlers/dynamic_data_handler.go`:
- Maps dynamically created endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Features an advanced **Query Engine** that automatically applies:
  - Filtering via query params (e.g. `?email=test@example.com`).
  - Sorting via query params (e.g. `?sort=created_at:desc`).
  - Pagination (e.g. `?page=1&limit=20`).
  - Dynamic joining via foreign key mappings.
- Enforces Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`:
- Provides safe schema changes mapping dynamic schema definitions in the metadata database to underlying DDL operations.
- Tracks and applies safe schema diffs (Add Column, Drop Column, Rename Column, Change Type) using GORM’s `.Migrator()`.
- Logs migration status natively into the `migrations` metadata table and supports automated backup before migration and rollback capabilities.

## 6. GraphQL Gateway

Managed by `internal/graphql/gateway.go`:
- Generates generic optional GraphQL schemas from service schemas dynamically using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.
- Supports GraphQL queries, mutations, and resolutions of relationships.

## 7. Backup System

Managed by `internal/services/backup_service.go`:
- Provides comprehensive automated table backups, schema backups, and service snapshots.
- Generates data snapshots via generic service record backup features mapping dynamic table contents to snapshots.
- Supports SQL dump and JSON snapshot formats, along with data restoration endpoints.

## 8. Observability Integration

Comprehensive observability capabilities implemented across the stack:
- **Logging**: Zap structured logging injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into the context. Outputs Request, Error, and Query logs.
- **Tracing**: OpenTelemetry/Jaeger initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Metrics**: Prometheus Metrics (`internal/middleware/prometheus.go`) tracks latency, slow queries, and HTTP status codes via standard HTTP interceptors. Exposed at `/metrics`.

## 9. Unit Testing

The platform enforces rigorous test quality and includes unit tests covering:
- Repositories and Data access layer
- Service business logic
- API Handlers (Presentation)
- End-to-end / UAT suites

Run tests with `make test-coverage` or `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`. Aimed to maintain a minimum of 80% test coverage.

## 10. Docker Setup

Containerized for scalability and smooth deployment:
- Provided `Dockerfile` builds a lightweight, production-ready Go binary.
- `docker-compose.yml` spins up the Go backend, PostgreSQL (metadata database), Redis (caching and rate limiting), Prometheus (metrics server), and Jaeger (tracing collector).

Run the cluster with `make docker-up` or `docker-compose up -d`.

## 11. Swagger Documentation

API documentation is fully generated using Swagger / OpenAPI. Swagger endpoints describe all core control plane actions dynamically.
Re-generate docs easily via:
```bash
make swagger
# or
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
