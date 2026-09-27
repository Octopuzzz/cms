# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend-as-a-Service (BaaS) platform written in Go, acting as a dynamic CMS and API builder. It is designed to be fully self-hosted, modular, scalable, and cloud-ready, similar to platforms like Hasura or Supabase.

## System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

### High Level Architecture

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

The project uses a clean modular structure:

```
├── cmd
│   └── server
│       └── main.go           # Entry point for the application
├── docker
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── prometheus.yml
├── internal
│   ├── config              # Configuration management
│   ├── database            # Database connection manager
│   ├── graphql             # GraphQL gateway
│   ├── handlers            # Presentation layer (HTTP handlers)
│   ├── middleware          # Gin middlewares (Auth, Rate Limiter, Prometheus)
│   ├── models              # Domain models (Metadata schema)
│   ├── services            # Application logic (Auth, Backup, Dynamic Data, Migrations, etc.)
│   └── tracing             # OpenTelemetry setup
├── pkg
│   ├── logger              # Zap structured logging
│   └── response            # Standardized API responses
├── tests                   # Unit, Integration, and UAT tests
│   ├── integration
│   ├── uat
│   └── unit
└── docs                    # Swagger documentation
```

## Component Implementations

### CMS Control Plane & Service Builder
Managed under `/api/v1/cms/` routes and handled by `internal/services/service_service.go` and `internal/services/user_service.go` (`DBConnService`).
- Provide an API interface to store service and field definitions in the metadata database.
- Utilize the `database_connections` engine to validate foreign database connectivity.
- Trigger migrations on service model updates to reflect schema changes.

### CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go`.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g. `?email=test@example.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

### Schema Migration Engine
Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
- Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.

### GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

### Backup System
Managed by `internal/services/backup_service.go`.
- Generates data snapshots. Implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.

### Observability Integration
Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.

## Unit Tests

The system includes unit tests covering the Repository, Service, and API handler layers. To run tests and generate coverage:

```bash
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Docker Setup

The project includes Docker support for easy deployment and containerization.
Run using docker-compose:

```bash
docker-compose -f docker/docker-compose.yml up -d
```

This sets up:
- CMS Backend application
- PostgreSQL (Metadata database)
- Redis (Caching / Connection tracking)
- Prometheus (Metrics)
- Jaeger (OpenTelemetry tracing)

## Swagger Documentation

Swagger / OpenAPI documentation is automatically generated. Access the documentation by generating it with:

```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

The documentation is then available via the API endpoint at `/swagger/index.html`.
