# Backend Platform (Dynamic CMS + API Builder)

Welcome to the fully self-hosted, Go-native Backend-as-a-Service (BaaS) Platform. This system allows developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

## System Architecture

The platform follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

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
PostgreSQL    MySQL       MongoDB
```

### Core Layers
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. The models include:

- `database_connections`: External DB configurations (id, name, type, host, port, credentials).
- `services`: User-created data models referring to `database_connection_id` and tracking dynamic schemas (`db_table_name`).
- `fields`: Attributes for each service (string, integer, float, uuid, JSON). Configures uniqueness, nullability, defaults.
- `service_permissions`: Connects `Role` to `Service` for fine-grained access checks.
- `migrations`: Tracks DDL executions with metadata required for rollbacks or history tracking.
- `backups`: Stores snapshot records or schema outputs.
- `users` and `roles`: General authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure:
```text
cmd/
  server/         # Entry point for the platform
internal/
  config/         # Configuration loading
  database/       # Connection pooling and management
  graphql/        # GraphQL schema generation and gateway
  handlers/       # HTTP handlers (Presentation Layer)
  middleware/     # Auth, Rate Limiter, Prometheus
  models/         # Domain models
  services/       # Business logic (Application Layer)
  tracing/        # OpenTelemetry tracing
pkg/
  logger/         # Zap structured logging
  response/       # Standardized Gin response helpers
tests/
  unit/           # Unit tests
  integration/    # Integration tests
  uat/            # UAT / E2E tests
docker/           # Docker setup and configurations
```

## CMS Control Plane & Service Builder

Managed under `/api/v1/cms/` routes and handled by `internal/services/service_service.go` and `internal/services/dbconn_service.go`.
- Manages database connections.
- Creates services, defines schemas and relations.
- Provides an API interface to store definitions in the metadata database.

## CRUD Engine & Query Engine

Managed by `internal/services/dynamic_data_service.go`.
- Dynamically generates CRUD REST endpoints (`GET`, `POST`, `PUT`, `DELETE`, `LIST`) based on service definitions.
- Supports advanced queries (Filtering, Sorting, Pagination, and Joins).
- Integrates with Role-Based Access Control and Row-Level filtering.

## Schema Migration Engine

Managed by `internal/services/migration_service.go`.
- Supports safe schema changes (Add, Drop, Rename Column, Change column type) via GORM.
- Logs migrations to the database and supports rollback capability and automatic backup before migration.

## GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- Automatically generates generic, optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

## Backup System

Managed by `internal/services/backup_service.go`.
- Implements generic service record backup features mapping dynamic table contents to JSON snapshots or SQL dumps.
- Supports table/schema backups, and snapshot restore via JSON payload decoding.

## Observability Integration

- **Logging**: Request, error, and query logs via Zap Logging, injected globally with trace IDs.
- **Metrics**: Request latency, error rate via Prometheus. Exported at `/metrics`.
- **Tracing**: OpenTelemetry/Jaeger initialized in `internal/tracing/` to wrap SQL commands and network logic.

## Unit Testing

Run unit tests directly:
```bash
go test -v -race -coverpkg=./... ./tests/...
```
The codebase requires an 80% test coverage minimum across Repository, Service, and API handlers. Tests are configured using in-memory SQLite instances.

## Docker Setup

Run `docker-compose up -d` to spin up necessary infrastructure (Prometheus, etc) along with the core application natively. Configuration files are maintained under the `docker/` directory.

## Swagger Documentation

API Documentation is auto-generated using Swagger. Run:
```bash
make swagger
```
or
```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

---

The final result is a **clean, modular, scalable, and production-ready** Backend-as-a-Service infrastructure.
