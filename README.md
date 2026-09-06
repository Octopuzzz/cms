# Dynamic CMS & API Builder Platform

A production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. This system works similarly to platforms like Hasura or Supabase, acting as a fully self-hosted, Go-native backend infrastructure generator.

## System Architecture

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)**.

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

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## Metadata Database Schema

The CMS stores platform metadata in a relational database.

- **`database_connections`**: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
- **`services`**: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
- **`fields`**: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
- **`service_permissions`**: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
- **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- **`backups`**: Stores snapshot records or schema outputs.
- **`users` and `roles`**: General authentication and authorization for the control plane.
- **`audit_logs`**: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure.

```
.
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection manager
│   ├── graphql/         # GraphQL gateway and schema generation
│   ├── handlers/        # API handlers (Presentation Layer)
│   ├── middleware/      # Gin middlewares (Auth, Metrics, Rate Limiting)
│   ├── models/          # Domain data models (Domain Layer)
│   ├── services/        # Business logic (Application Layer)
│   └── tracing/         # OpenTelemetry tracing setup
├── pkg/
│   ├── logger/          # Zap structured logging wrapper
│   └── response/        # Standardized API responses
├── tests/               # Unit, integration, and UAT tests
├── docker/              # Docker configuration files
├── Dockerfile           # Docker container configuration
└── Makefile             # Task runner
```

## CMS Control Plane & Service Builder

Managed under `/api/v1/cms/` routes and handled by `internal/services/service_service.go` and `internal/services/dbconn_service.go`. These services:
- Provide an API interface to store service and field definitions in the metadata database.
- Utilize the `database_connections` engine to validate foreign database connectivity.
- Trigger migrations on service model updates to reflect schema changes.

## CRUD Engine & Query Engine

Managed by `internal/services/dynamic_data_service.go`.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g. `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## Schema Migration Engine

Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM's `.Migrator()`.
- Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.

## GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

## Backup System

Managed by `internal/services/backup_service.go`.
- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.

## Observability Integration

Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.

## Unit Testing Setup

The system includes comprehensive tests covering the Repository, Service, and API handlers. Tests use an in-memory SQLite database (`:memory:`) via the `github.com/glebarez/sqlite` driver to isolate testing environments.
To run tests and see the coverage:

```bash
make test-coverage
# or
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Docker Setup

The platform is containerized.
- Run via Docker Compose: `docker-compose up -d`
- See `docker/docker-compose.yml` and `Dockerfile` for full setups of the Go backend, PostgreSQL, Redis, Jaeger, and Prometheus.

## Swagger Documentation

API documentation is generated using `swag` and integrated into the application.
- To generate documentation: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`
- Swagger UI will be available on the configured route (e.g. `/swagger/index.html`).
