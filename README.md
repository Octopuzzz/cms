# Backend Platform (Dynamic CMS + API Builder)

A self-hosted, Go-native Backend-as-a-Service platform. This platform allows developers to dynamically create backend services, manage database connections, define data schemas, automatically generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability. It operates similarly to Hasura or Supabase but is fully self-hosted and written in Go.

## System Architecture Explanation

The platform follows **Clean Architecture and Domain-Driven Design (DDD)**.

```
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

The core internal configuration is stored in the **Metadata Database**. Tables defined in `internal/models/models.go` include:

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

```
.
├── cmd
│   └── server          # Main application entry point
├── docker              # Docker and containerization files
├── internal            # Internal application code
│   ├── config          # Configuration loading
│   ├── database        # Database connection and pooling management
│   ├── graphql         # GraphQL gateway and resolvers (gqlgen)
│   ├── handlers        # HTTP handlers (Gin controllers)
│   ├── middleware      # Gin middlewares (Auth, Prometheus, Tracing)
│   ├── models          # Domain models (GORM)
│   ├── services        # Application business logic (CRUD Engine, Migrations, etc.)
│   └── tracing         # OpenTelemetry tracing setup
├── pkg                 # Public packages
│   ├── logger          # Zap logging wrapper
│   └── response        # Standardized HTTP response formatters
└── tests               # Test suites
    ├── integration
    ├── uat
    └── unit
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
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
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

## Unit Tests

The system includes comprehensive tests covering repository, service, and API handler layers.
- Minimum coverage requirement: 80%
- Run full suite with coverage: `make test-coverage`
- Isolated tests use an in-memory SQLite database setup (`:memory:`).

## Docker Setup

The platform is containerized using Docker.
- `Dockerfile` for the Go application build.
- `docker-compose.yml` to orchestrate the backend, database (PostgreSQL), Redis, and observability stack (Prometheus, Jaeger).

## Swagger Documentation

API documentation is generated using Swagger/OpenAPI.
- Available at `/swagger/index.html` (when run with Swagger enabled).
- Generated via standard `swag init` or `~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
