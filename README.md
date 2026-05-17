# Platform Architecture Overview (CMS Backend)

## Overview
This platform is a Backend-as-a-Service system written in Go that allows developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability. The platform acts as a backend infrastructure generator similar to platforms like Hasura or Supabase, but is fully self-hosted and Go-native.

## System Architecture Principles

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## High-Level Platform Architecture

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

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database** (Default: PostgreSQL). The platform relies on the following core entities as defined in `internal/models/models.go`:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project relies on a modular layout:
```text
cmd/
└── server/          # Entry point for the application (main.go)
internal/
├── config/          # Environment configuration loading
├── database/        # Database Connection Manager (PostgreSQL, MySQL, MongoDB, SQLite)
├── graphql/         # GraphQL Gateway and schema resolution
├── handlers/        # Gin HTTP route handlers
├── middleware/      # Auth, Logging, Metrics, and Rate Limiting
├── models/          # Core Domain and Metadata structures
├── services/        # Application Business Logic
└── tracing/         # OpenTelemetry / Jaeger Setup
pkg/
├── logger/          # Zap structured logging
└── response/        # Standardized Gin API response builders
tests/               # Unit, Integration, and UAT test suites
docker/              # Container deployment artifacts (Dockerfile, Compose, Prom YAML)
```

## CMS Control Plane & Service Builder
Managed under `/api/v1/cms/` routes and handled by `internal/services/service_service.go` and `internal/services/dbconn_service.go`. These services:
- Provide an API interface to store service and field definitions in the metadata database.
- Utilize the `database_connections` engine to validate foreign database connectivity.
- Trigger migrations on service model updates to reflect schema changes.

## Dynamic CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go` handling standard operations (Create, Read, Update, Delete, List).
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g. `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## Schema Migration Engine
Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
- Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.

## GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models mapping directly to dynamic services.

## Backup System
Managed by `internal/services/backup_service.go`.
- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to JSON snapshots.
- Supports restoring rows via JSON payload decoding.

## Observability Integration
Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.

## Unit Testing
Unit, Integration, and UAT tests are configured under `/tests/`.
The system strives for a minimum of 80% test coverage.
Tests are invoked via `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...` or using the internal Makefile `make test-unit`.

## Docker Setup
The platform is fully containerized.
- The `Dockerfile` provides a multi-stage build system resulting in a minimal footprint executing `bin/cms-backend`.
- `docker-compose.yml` orchestrates the system, bringing up the Core Go App alongside auxiliary services like Redis (Caching), Prometheus (Metrics), and Jaeger (Tracing).

## Swagger / OpenAPI Documentation
The system provides fully generated REST documentation.
Managed via `swaggo/swag`, documentation can be built via `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs` and accessed at the `/swagger/index.html` endpoint in the running application.
