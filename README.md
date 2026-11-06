# Platform Architecture Overview

## Goal
Build a Backend-as-a-Service platform allowing developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

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

## System Architecture Principles

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database** (using PostgreSQL as recommended, though SQLite is used for development/testing). Tables defined in `internal/models/models.go` include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`), id, created_at, updated_at, etc.
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Component Implementation

### Go Project Structure

The project uses a clean, modular structure:

```
├── cmd
│   └── server                # Application entrypoint
├── docker                    # Docker and containerization files
├── internal
│   ├── config                # Configuration and environment setup
│   ├── database              # Database Connection Manager
│   ├── graphql               # GraphQL Gateway implementation
│   ├── handlers              # API Route Handlers (Presentation Layer)
│   ├── middleware            # HTTP Middleware (Auth, Metrics, etc.)
│   ├── models                # Domain Models (Database Schema)
│   ├── services              # Application Business Logic (CMS, Builders, Engines)
│   └── tracing               # OpenTelemetry tracing setup
├── pkg
│   ├── logger                # Zap structured logger wrapper
│   └── response              # Standardized API response helpers
└── tests                     # Test suites
    ├── integration
    ├── uat
    └── unit
```

### CMS Control Plane & Service Builder
Managed under `/api/v1/cms/` routes and handled by `internal/services/service_service.go` and `internal/services/dbconn_service.go`. These services:
- Provide an API interface to store service and field definitions in the metadata database.
- Utilize the `database_connections` engine to validate foreign database connectivity.
- Trigger migrations on service model updates to reflect schema changes.

### Dynamic CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go`.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Generates automatic CRUD operations (Create, Read, Update, Delete, List) upon service creation.
- Applies automated filtering via query params (e.g. `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

### Schema Migration Engine
Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
- Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.
- Implements automated safety features such as automatic backups before migration operations.

### GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.
- Supports generic queries and mutations based on dynamic service models.

### Backup System
Managed by `internal/services/backup_service.go`.
- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports formats like JSON snapshots and SQL dumps.
- Supports restoring rows via JSON payload decoding, allowing for seamless state recovery.

### Observability Integration
Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context. Captures request logs, error logs, and query logs.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic for full distributed tracing capabilities.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency, request volume, error rates, and status codes via standard HTTP interceptors. Exported at `/metrics`.

## Unit Testing

The system includes comprehensive tests covering Repository, Service, and API handler layers.
- A minimum 80% coverage is maintained across these key components.
- Unit tests utilize an in-memory SQLite database (`:memory:`) via `github.com/glebarez/sqlite` to prevent external dependencies.
- You can run tests using `make test` or generate coverage reports via `make test-coverage`.
  - Alternatively: `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`

## Docker Setup

The platform is containerized and cloud-ready out of the box.
- Contains a `Dockerfile` for multi-stage building of the Go binary.
- A `docker-compose.yml` file is provided for spinning up the Backend alongside requisite databases (PostgreSQL/MySQL), Redis for caching, Jaeger for tracing, and Prometheus/Grafana for monitoring metrics.
- Utilize `make docker-up` and `make docker-down` for easy orchestration.

## Swagger Documentation

Auto-generated API Documentation is configured via `swaggo/swag`.
- Accessible at `/swagger/index.html`.
- To regenerate the documentation after modifying handler annotations, simply run `make swagger` (which maps to `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`).
