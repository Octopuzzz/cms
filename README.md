# CMS Backend Platform

This repository contains the Go-native Backend-as-a-Service platform (Dynamic CMS + API Builder). It follows a self-hosted model similar to Hasura or Supabase. It allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

The system is designed to be modular, scalable, and cloud-ready.

## System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

### High-Level Platform Architecture

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

The core internal configuration is stored in the **Metadata Database**. The models include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Project Structure

The project structure is organized as follows:

```
cmd
└── server
    └── main.go           # Application entrypoint

internal
├── config                # Configuration setup
├── database              # Database Connection Manager
├── graphql               # GraphQL Gateway implementation
├── handlers              # HTTP API Handlers (Presentation Layer)
├── middleware            # Middlewares (Auth, Prometheus, Rate Limiting)
├── models                # Domain Models
├── services              # Business Logic (Application Layer)
└── tracing               # OpenTelemetry Integration

pkg
├── logger                # Structured Logging (Zap)
└── response              # API Response Formatters

tests                     # Test suites (unit, integration, UAT)
docker                    # Docker setups (Dockerfile, Compose, Prom)
```

## Core Components

### CMS Control Plane & Service Builder
Managed under `/api/v1/cms/` routes and handled by `internal/services/service_service.go` and `internal/services/dbconn_service.go`. These services:
- Provide an API interface to store service and field definitions in the metadata database.
- Utilize the database_connections engine to validate foreign database connectivity.
- Trigger migrations on service model updates to reflect schema changes.

### CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go`.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g. `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

### Schema Migration Engine
Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM's `.Migrator()`.
- Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.

### GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

### Backup System
Managed by `internal/services/backup_service.go`.
- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.

### Observability Integration
Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.

## Unit Testing

The system includes comprehensive test suites spanning unit, integration, and user acceptance testing (UAT). Minimum test coverage requirement is 80%.

Commands to run tests:
- `make test` : Runs all tests
- `make test-unit` : Runs unit tests
- `make test-integration` : Runs integration tests
- `make test-uat` : Runs UAT/E2E tests
- `make test-coverage` : Generates coverage report

## Docker Setup

The application is containerized and includes a `docker-compose.yml` for local multi-service orchestration (DB, Redis, Prometheus, Jaeger).

Commands:
- `make docker-up` : Start Docker services
- `make docker-down` : Stop Docker services
- `make docker-build` : Build the Docker image

## Swagger Documentation

API Documentation is auto-generated using standard Go Swagger annotations.

Command to regenerate documentation:
- `make swagger` (which runs: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`)

The Swagger UI is accessible at `/api/v1/swagger/index.html` when the application is running.
