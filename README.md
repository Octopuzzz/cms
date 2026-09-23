# Backend Platform (Dynamic CMS + API Builder)

A self-hosted, Go-native Backend-as-a-Service (BaaS) platform that dynamically generates REST and GraphQL APIs. It allows users to dynamically create backend services, schemas, manage database connections, and auto-generates CRUD APIs.

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

The core internal configuration is stored in the **Metadata Database**. The metadata database tracks platform configuration:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean, modular structure:

```text
├── cmd
│   └── server                # Application entrypoint
├── docker                    # Docker configurations
├── internal
│   ├── config                # Configuration management
│   ├── database              # Database connection and pooling
│   ├── graphql               # GraphQL gateway implementation
│   ├── handlers              # HTTP API Handlers (Gin)
│   ├── middleware            # Authentication, rate limiting, metrics
│   ├── models                # Core domain models
│   ├── services              # Application business logic
│   └── tracing               # OpenTelemetry integration
├── pkg
│   ├── logger                # Structured Zap logger
│   └── response              # API response helpers
└── tests
    ├── integration           # Integration tests
    ├── uat                   # User Acceptance Testing
    └── unit                  # Unit tests
```

## Core Engines

### CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go`.
- Automatically generates CRUD endpoints when a service is created.
- Maps endpoints (e.g., `POST /api/v1/data/{slug}`, `GET /api/v1/data/{slug}/{id}`) to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g. `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

### Schema Migration Engine
Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
- Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.
- Ensures safe schema changes to dynamic services.

### GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL queries, mutations, and relations for defined data models.

### Backup System
Managed by `internal/services/backup_service.go`.
- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.
- Facilitates backup creation via `POST /cms/backup/service/{service_id}`.

## Observability Integration

The platform provides built-in observability features:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context. Logs requests, errors, and queries.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.

## Unit Tests

The system is rigorously tested with over 80% test coverage across Repository, Service, and API handler layers.
- Uses standard Go testing tools and an in-memory SQLite database (`:memory:`) via the `github.com/glebarez/sqlite` driver for fast execution.
- Command to run all tests: `make test-coverage` or `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`

## Docker Setup

Containerized for scalability and ease of deployment:
- `Dockerfile`: Multi-stage build for the Go application.
- `docker-compose.yml`: Spins up the CMS Backend, PostgreSQL metadata database, Redis (caching), Prometheus (metrics), and Jaeger (tracing).
- Use `docker-compose up -d` to launch the entire stack.

## Swagger Documentation

API documentation is automatically generated using `swag`.
- Provides a UI to interact with the CMS Control Plane and dynamically generated endpoints.
- Generate documentation via: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`
