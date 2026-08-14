# Backend Platform (Dynamic CMS + API Builder)

This is a **production-grade Backend Platform (Dynamic CMS + API Builder) written in Go**. It is a fully self-hosted, Go-native Backend-as-a-Service (BaaS) platform that dynamically generates REST and GraphQL APIs.

It allows developers to:
- Connect databases
- Create data models dynamically
- Generate CRUD APIs automatically
- Generate optional GraphQL APIs
- Manage schema migrations
- Monitor logs and performance
- Manage backups
- Scale services

---

## 1. System Architecture Explanation

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

---

## 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database** (using PostgreSQL/SQLite). Tables defined in `internal/models/models.go` include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

---

## 3. Go Project Structure

The project uses a clean modular structure:

```text
.
├── cmd
│   └── server          # Main entrypoint of the application
├── internal
│   ├── handlers        # API Handlers (REST & GraphQL)
│   ├── services        # Core business logic engines
│   ├── database        # DB connections and migrations
│   ├── models          # Domain models
│   ├── config          # Application configuration
│   ├── graphql         # GraphQL schema & gateway
│   ├── tracing         # OpenTelemetry setup
│   └── middleware      # Gin middlewares (Auth, metrics)
├── pkg
│   ├── logger          # Zap structured logging
│   └── response        # Standardized HTTP response helpers
├── tests               # Unit, integration, and UAT tests
├── docker              # Dockerfile and compose configurations
└── docs                # Generated Swagger documentation
```

---

## 4. CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g. `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

---

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
- Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.

---

## 6. GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models, supporting queries, mutations, and relations.

---

## 7. Backup System

Managed by `internal/services/backup_service.go`.
- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.

---

## 8. Observability Integration

Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.

---

## 9. Unit Tests

Unit tests cover the Repository, Service, and API handler layers, striving for minimum 80% coverage.
Unit tests use an in-memory SQLite database setup (`:memory:`) via the `github.com/glebarez/sqlite` driver.
To run all tests:
```bash
make test-coverage
# or
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

---

## 10. Docker Setup

The system includes a fully configured `Dockerfile` and `docker-compose.yml` to set up the Go application, databases, Redis, and observability stack.
Run using:
```bash
docker-compose up -d
```

---

## 11. Swagger Documentation

API documentation is generated automatically using `swag`. The API docs expose REST endpoints for CMS management and CRUD access.
Generate docs using:
```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
(Generated `docs/` are in `.gitignore` and not checked into source control)
