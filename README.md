# CMS Backend Platform

A production-grade Backend-as-a-Service platform allowing developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

## System Architecture

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

*   **Presentation Layer**: Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
*   **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
*   **Domain Layer**: Data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
*   **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

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

The core internal configuration is stored in the **Metadata Database**.

*   `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
*   `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
*   `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
*   `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks.
*   `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
*   `backups`: Stores snapshot records or schema outputs.
*   `users` and `roles`: General authentication and authorization for the control plane.
*   `audit_logs`: Detailed logging of structural and data-level modifications.

## Project Structure

The project uses a clean modular structure.

```text
cmd/
└── server/
    └── main.go              # Entry point of the application
internal/
├── config/                  # Configuration management
├── database/                # Database connection manager
├── graphql/                 # GraphQL Gateway using gqlgen
├── handlers/                # HTTP route handlers (Presentation Layer)
├── middleware/              # Gin middlewares (Auth, Rate Limiting, Metrics)
├── models/                  # Domain entities (Domain Layer)
├── services/                # Business logic (Application Layer)
└── tracing/                 # OpenTelemetry/Jaeger tracing
pkg/
├── logger/                  # Zap logger implementation
└── response/                # Standardized HTTP responses
tests/                       # Unit, Integration, and UAT tests
docker/                      # Dockerfile and configurations
```

## Component Implementations

### CRUD Engine & Query Engine

Managed by `internal/services/dynamic_data_service.go`.
When a service is created, the system automatically maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
It applies automated filtering via query params (e.g., `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination logic. Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

Supported operations: Create, Read, Update, Delete, List.

### Schema Migration Engine

Managed by `internal/services/migration_service.go`.
Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
Logs migration status natively into the `migrations` table and handles rollbacks through snapshot retention logic.

### GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
Automatically generates generic optional schemas using `github.com/99designs/gqlgen/graphql` from service definitions.
Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

### Backup System

Managed by `internal/services/backup_service.go`.
Generates data snapshots. Supports table backup, schema backup, and service snapshots.
Format supported includes SQL dump and JSON snapshots. Can restore rows via JSON payload decoding.

### Observability Integration

Implemented across the stack:
*   **Logging**: Request, Error, and Query logs managed by Zap (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
*   **Tracing**: OpenTelemetry/Jaeger initialized in `internal/tracing/` to wrap SQL commands and network logic.
*   **Metrics**: Prometheus metrics tracked by `internal/middleware/prometheus.go` via standard HTTP interceptors. Exported at `/metrics`.

## Unit Tests

The system includes a comprehensive test suite targeting Repository, Service, and API handler layers.
We require a minimum coverage of 80%.

To run the unit tests:
```bash
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```
A Makefile is also provided: `make test-unit`, `make test-integration`, `make test-uat`, `make test-coverage`.

## Docker Setup

The platform is containerized using Docker.
You can run the full stack (API, PostgreSQL, Redis, Jaeger, Prometheus) using Docker Compose:

```bash
docker-compose up -d
```
The Dockerfile is located at the root of the project.

## Swagger Documentation

API documentation is generated using Swaggo.
To generate documentation:
```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
It is served at the endpoint `/api/v1/swagger/index.html`.
