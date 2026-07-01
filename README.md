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

The core internal configuration is stored in the **Metadata Database**. Tables defined in `internal/models/models.go` include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Component Implementation

### CMS Control Plane & Service Builder
Managed under `/api/v1/cms/` routes and handled by `internal/services/service_service.go` and `internal/services/dbconn_service.go`. These services:
- Provide an API interface to store service and field definitions in the metadata database.
- Utilize the `database_connections` engine to validate foreign database connectivity.
- Trigger migrations on service model updates to reflect schema changes.

### CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go`.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g. `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination logic.
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
- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.

### Observability Integration
Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.

### Go Project Structure
The project follows a clean, modular structure standard for Go projects:
- **`cmd/server/`**: The main application entry point (`main.go`).
- **`internal/`**: Private application code.
  - **`api/`**, **`handlers/`**: Presentation layer (Gin routes).
  - **`services/`**: Application layer with business logic (CRUD, CMS Control Plane, Schema Migration).
  - **`models/`**: Domain layer defining data structures.
  - **`database/`**: Database connection managers.
  - **`graphql/`**: GraphQL gateway and generation (`gqlgen`).
  - **`middleware/`**, **`tracing/`**: Security, observability, and interceptors.
- **`pkg/`**: Public libraries used across the project (e.g., `logger/`, `response/`).
- **`tests/`**: Contains Unit, Integration, and UAT (End-to-End) tests.
- **`docker/`**: Docker and Prometheus configurations.

### Unit Tests
The system includes comprehensive test coverage for Repository, Service, and API handlers.
- **Minimum Coverage Target**: 80%
- **Execution Command**: `go test -v -race ./tests/...`
- **Coverage Command**: `make test-coverage`
- Tests use an in-memory SQLite database setup (`:memory:`) via the `github.com/glebarez/sqlite` driver for fast and isolated execution without external dependencies.

### Docker Setup
The platform is fully containerized using Docker and Docker Compose.
- **`Dockerfile`**: A multi-stage build starting with `golang:1.24-alpine` for the builder and `alpine:latest` for the lightweight runtime image.
- **`docker-compose.yml`**: Defines the services needed for local development and testing:
  - `cms-backend`: The Go platform application.
  - `postgres`: The metadata database.
  - `redis`: Caching layer.
  - `jaeger`: OpenTelemetry tracing backend.
  - `prometheus`: Metrics collection.
- **Run**: `make docker-up` or `docker-compose up -d`.

### Swagger Documentation
The REST APIs are fully documented using Swagger/OpenAPI.
- **Generation Tool**: `swaggo/swag`
- **Command**: `make swagger` (which runs `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`).
- **Endpoint**: Once generated and running, the Swagger UI is available at `http://localhost:8080/api/v1/swagger/index.html`. Note: the generated `docs/` directory is deliberately excluded from version control.
