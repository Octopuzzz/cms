# Dynamic CMS + API Builder Backend Platform

A production-grade Backend-as-a-Service platform (Dynamic CMS + API Builder) written in Go. This system allows developers to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints. It is fully self-hosted, Go-native, modular, scalable, and cloud-ready.

## System Architecture Explanation

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

## Metadata Database Schema

The platform metadata is stored in a relational database (PostgreSQL recommended). Core tables include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
   - Fields: `id`, `name`, `database_connection_id`, `created_at`, `updated_at`, etc.
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean, modular structure:

```text
.
├── cmd/
│   └── server/             # Application entrypoint
├── docker/                 # Docker configuration and Prometheus configs
├── internal/
│   ├── config/             # Environment and configuration loading
│   ├── database/           # Database connection manager and pooling
│   ├── graphql/            # GraphQL gateway and schemas
│   ├── handlers/           # HTTP controllers / Presentation Layer
│   ├── middleware/         # Auth, Rate Limiting, Metrics, Tracing
│   ├── models/             # Domain entities and GORM models
│   ├── services/           # Business logic and Application Layer
│   └── tracing/            # OpenTelemetry setup
├── pkg/
│   ├── logger/             # Zap structured logger
│   └── response/           # Standard API response helpers
└── tests/
    ├── integration/
    ├── uat/
    └── unit/
```

## CRUD Engine Implementation

Managed by the `DynamicDataService` (`internal/services/dynamic_data_service.go`).
- Automatically generates CRUD endpoints when a service is created.
- Maps dynamic endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g., `?email=test@test.com`), dynamic joining via foreign key mappings, sorting, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## Schema Migration Engine

Managed by `MigrationService` (`internal/services/migration_service.go`).
- Tracks and applies safe schema diffs (Add, Drop, Rename Column, Change column type) using GORM’s `.Migrator()`.
- Logs migration status natively into the `migrations` table.
- Supports automatic backups before migrations and handles rollbacks through snapshot retention logic.

## GraphQL Gateway

Managed by `Gateway` (`internal/graphql/gateway.go`).
- Automatically generates generic optional GraphQL schemas from service schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill queries, mutations, and relations for defined data models dynamically.
- Provides a GraphQL Playground endpoint for easy querying.

## Backup System

Managed by `BackupService` (`internal/services/backup_service.go`).
- Generates data and schema snapshots. Currently implements generic service record backup features mapping dynamic table contents to JSON snapshots.
- Tracks backup metadata in the `backups` table.
- Supports restoring rows via JSON payload decoding (`POST /api/v1/cms/restore/:id`).

## Observability Integration

Implemented across the stack for high visibility and performance tracking:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context. Outputs structured JSON logs.
- **OpenTelemetry / Jaeger**: Initialized in `internal/tracing/` to trace requests, wrap SQL commands, and monitor network logic.
- **Prometheus Metrics**: The `internal/middleware/prometheus.go` middleware tracks latency, request counts, and status codes. Exported at the `/metrics` endpoint.

## Unit Tests

The system includes comprehensive tests with a minimum of 80% coverage.
- **Test Areas**: Repository, Service, API handlers.
- Uses an in-memory SQLite database setup (`:memory:`) via the `github.com/glebarez/sqlite` driver for fast execution.
- Command to run all tests with coverage: `make test-coverage` or `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`

## Docker Setup

Containerized for easy deployment and scaling.
- A `docker-compose.yml` is provided at the root to spin up the platform.
- Services included: `cms-backend`, `postgres` (metadata db), `redis` (caching), `jaeger` (tracing), and `prometheus` (metrics).
- **Run command**: `make docker-up` or `docker-compose up -d`.

## Swagger Documentation

API documentation is generated using `swaggo/swag`.
- Provides an interactive Swagger UI for the REST API endpoints.
- Access the docs via `GET /swagger/index.html` on the running server.
- To generate/update the documentation, run: `make swagger` or `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
