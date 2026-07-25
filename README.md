# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. It operates similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native. The platform enables users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints. It is modular, scalable, and cloud-ready.

## System Architecture Explanation

The CMS Backend adheres to **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

*   **Presentation Layer**: Powered by the Gin HTTP framework (`internal/handlers/`). It acts as the API gateway mapping requests to internal services, including REST and GraphQL endpoints (using `gqlgen`).
*   **Application Layer**: Contains business logic (`internal/services/`). Services handle data access, generate dynamic schemas, and perform operations requested by the handlers.
*   **Domain Layer**: Defines data models (`internal/models/`). This layer defines core entities such as `Service`, `Field`, `DatabaseConnection`, `User`, and `Role`.
*   **Infrastructure Layer**: Handles cross-cutting concerns like connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), logging (`pkg/logger/`), and external database connections.

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

The platform stores metadata representing system configurations in the **Metadata Database**. The following core tables are defined (as found in `internal/models/models.go`):

1.  **`database_connections`**: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2.  **`services`**: Represents user-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3.  **`fields`**: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4.  **`service_permissions`**: Connects `Role` to `Service`, providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5.  **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6.  **`backups`**: Stores snapshot records or schema outputs.
7.  **`users`** and **`roles`**: General authentication and authorization for the control plane.
8.  **`audit_logs`**: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project follows a clean, modular structure standard in modern Go applications:

```text
.
├── cmd/
│   └── server/               # Main application entry point
├── internal/
│   ├── api/                  # API routing definitions
│   ├── config/               # Configuration management
│   ├── database/             # Database connection management
│   ├── graphql/              # GraphQL gateway implementation
│   ├── handlers/             # HTTP request handlers (Presentation layer)
│   ├── middleware/           # HTTP middlewares (Auth, Tracing, Metrics)
│   ├── models/               # Domain models (Domain layer)
│   ├── services/             # Business logic (Application layer)
│   └── tracing/              # OpenTelemetry instrumentation
├── pkg/
│   ├── logger/               # Structured Zap logging
│   └── response/             # Standardized HTTP response helpers
├── tests/
│   ├── integration/          # Integration tests
│   ├── uat/                  # User Acceptance Tests
│   └── unit/                 # Unit tests
├── docker/                   # Docker environment configurations
└── ...
```

## CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`, the dynamic CRUD engine:

*   Maps dynamic API endpoints (e.g., `GET /api/v1/data/{slug}`) to standard GORM database operations on the fly.
*   Supports advanced Query Engine capabilities including:
    *   **Filtering**: Automated filtering via query params (e.g., `?email=test@example.com`).
    *   **Sorting**: Sorting logic based on schema fields.
    *   **Pagination**: Applying limits and offsets to queries.
    *   **Relations/Joins**: Dynamic joining via foreign key mappings.
*   Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## Schema Migration Engine

Managed by `internal/services/migration_service.go`, the system provides safe schema changes:

*   Tracks and applies schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
*   Logs migration status natively into the `migrations` table.
*   Handles rollbacks and tracks history through snapshot retention logic.

## GraphQL Gateway

Managed by `internal/graphql/gateway.go`, the GraphQL system provides optional APIs:

*   Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
*   Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models, supporting queries, mutations, and relations dynamically based on the service configurations.

## Backup System

Managed by `internal/services/backup_service.go`, the system supports:

*   Generating data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
*   Supports restoring rows via JSON payload decoding.
*   Endpoints to trigger and restore backups for a given service.

## Observability Integration

The system implements state-of-the-art observability:

*   **Logging**: Structured Zap logging (`pkg/logger/`) with correlation and trace IDs tied directly into the Context. Request, error, and query logs are supported.
*   **Tracing**: OpenTelemetry/Jaeger initialized in `internal/tracing/` to wrap SQL commands and network logic.
*   **Metrics**: Prometheus Metrics exported at `/metrics`. `internal/middleware/prometheus.go` tracks request latency, slow queries, error rates, and status codes via standard HTTP interceptors.

## Unit Tests

The system includes comprehensive unit testing.

*   Minimum coverage requirement: 80 percent across Repository, Service, and API handler layers.
*   Unit tests use an in-memory SQLite database setup (DSN: `:memory:`) via the `github.com/glebarez/sqlite` driver for fast and isolated execution.
*   Command to run full suite: `make test-coverage` or `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`.

## Docker Setup

The platform is containerized using Docker.

*   The project includes a `Dockerfile` and `docker-compose.yml` for simplified deployment and orchestrating external dependencies (like PostgreSQL, Redis, Jaeger, Prometheus).

## Swagger Documentation

API Documentation is auto-generated using Swagger/OpenAPI.

*   Generation command: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
*   Provides interactive documentation for the REST API interface.
