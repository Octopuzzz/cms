# CMS Backend Platform

A production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. This system is fully self-hosted and Go-native, allowing users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

## High Level Platform Architecture

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

## System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. Tables include:

- `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
- `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
- `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
- `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
- `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- `backups`: Stores snapshot records or schema outputs.
- `users` and `roles`: General authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

```text
.
├── ARCHITECTURE.md       # Architectural overview
├── Dockerfile            # Multi-stage Docker build
├── Makefile              # Build, test, and run scripts
├── cmd/
│   └── server/           # Main application entrypoint
├── docker/               # Docker Compose and configs
├── internal/
│   ├── config/           # Environment configuration
│   ├── database/         # DB Connection Manager
│   ├── graphql/          # Auto-generated GraphQL Gateway
│   ├── handlers/         # Presentation Layer (Gin REST API)
│   ├── middleware/       # Auth, Rate Limiter, Prometheus
│   ├── models/           # Domain Layer
│   ├── services/         # Application Layer (CRUD, Query, Backup, etc.)
│   └── tracing/          # Observability (OpenTelemetry)
├── pkg/
│   ├── logger/           # Zap structured logging
│   └── response/         # Standardized API responses
└── tests/                # Unit, Integration, and UAT tests
```

## CRUD Engine Implementation & Query Engine
Managed by `internal/services/dynamic_data_service.go` and `internal/handlers/dynamic_data_handler.go`.
- Maps dynamic REST endpoints (`GET /api/v1/data/{slug}`) to GORM database operations.
- Translates URL query parameters to filtering (`?email=test@example.com`), sorting (`?sort=created_at:desc`), pagination (`?page=1&limit=20`), and relation joining.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## Schema Migration Engine
Managed by `internal/services/migration_service.go` and `internal/handlers/migration_handler.go`.
- Handles applying schema modifications safely using GORM’s `.Migrator()`.
- Logs and retains historical states inside the `migrations` table natively and handles rollbacks.

## GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

## Backup System
Managed by `internal/services/backup_service.go` and `internal/handlers/backup_handler.go`.
- Triggers data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding (`POST /cms/restore/service/{service_id}`).

## Observability Integration
Implemented across the stack:
- **Logging**: Zap structured logger with context propagation, tracing IDs tied directly into Context. Request logs, error logs, query logs.
- **Metrics**: Prometheus instrumentation injected as Gin middleware. Exported at `/metrics`. Request latency, slow queries, error rate.
- **Tracing**: OpenTelemetry/Jaeger initialized in `internal/tracing/` to wrap SQL commands and network logic.

## Unit Tests
Tests are placed under the `tests/` directory with sections for `unit`, `integration`, and `uat`.
- Use `make test` or `go test -v -race ./tests/...` to run the suite.
- The project requires a minimum of 80% unit test coverage across Repository, Service, and API handler layers.

## Docker Setup
The project utilizes `docker-compose.yml` and `Dockerfile` to establish containerized environments seamlessly. It orchestrates the Go API along with required infrastructure.

## Swagger Documentation
Swagger documentation is generated automatically based on code comments.
Run `make swagger` to generate the `docs/` folder, which serves the API documentation.
