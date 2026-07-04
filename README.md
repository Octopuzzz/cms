# Dynamic CMS + API Builder Platform

## Overview
This platform is a production-grade Backend-as-a-Service (BaaS) and Dynamic CMS written in Go. It enables developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

## System Architecture

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

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
PostgreSQL    MySQL       MongoDB
```

### Core Layers
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. The tables include:

1. `database_connections`: Stores external DB configurations (id, name, type, host, port, credentials).
2. `services`: Represents user-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas.
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects roles to services providing fine-grained access checks (e.g., CanCreate, CanRead).
5. `migrations`: Tracks DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` & `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project is structured into clear, modular domains:

```text
├── cmd
│   └── server          # Application entry point
├── internal
│   ├── config          # Configuration loading and environment variables
│   ├── database        # Database connection managers
│   ├── graphql         # GraphQL schema generation and gateway
│   ├── handlers        # HTTP and REST API handlers (Presentation layer)
│   ├── middleware      # Gin middlewares (Auth, Prometheus, Tracing)
│   ├── models          # Domain models and metadata structures
│   ├── services        # Core business logic engines (Service Builder, Migrations)
│   └── tracing         # OpenTelemetry tracing setup
├── pkg
│   ├── logger          # Zap structured logging wrapper
│   └── response        # Standardized API response helpers
├── docker              # Docker container configurations
└── tests
    ├── integration     # Cross-layer integration tests
    ├── uat             # User Acceptance Testing
    └── unit            # Isolated unit tests for internal packages
```

## CMS Control Plane & Service Builder
Managed under `/api/v1/cms/` routes (`internal/services/service_service.go`, `internal/services/dbconn_service.go`).
- Provides an API to store service and field definitions.
- Uses `database_connections` engine to validate external database connectivity.
- Dynamically creates internal schemas for defined services.

## Dynamic CRUD & Query Engine
Managed by `internal/services/dynamic_data_service.go`.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g. `?email=test@example.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control before querying.

## Schema Migration Engine
Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s Migrator.
- Logs migration status natively into the `migrations` table and supports basic rollbacks.

## GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL queries.
- Dynamically generates schemas for user-defined models via `gqlgen`.

## Backup Engine
Managed by `internal/services/backup_service.go`.
- Generates data snapshots of dynamic table contents.
- Supports generic service record backup features and payload restoration.

## Observability Integration
Implemented across the stack:
- **Logging**: Zap structured logging (`pkg/logger/`) with correlation and trace IDs tied to Context.
- **Tracing**: OpenTelemetry (`internal/tracing/`) wrapping SQL commands and network logic.
- **Metrics**: Prometheus metrics via `internal/middleware/` tracking latency and status codes exported at `/metrics`.

## Security & Performance
- JWT Authentication and Role-Based Access Control.
- Database connection pooling implemented in the `database` manager.
- N+1 query performance optimizations using bulk batched operations.

## Docker Setup
The platform is fully containerized. A `Dockerfile` is provided for the core application, and `docker-compose.yml` can spin up the application alongside PostgreSQL, Redis, Prometheus, Jaeger, and other dependent services.

## Swagger Documentation
API Documentation is available via Swagger/OpenAPI.
Generate using: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`

## Unit Testing
The platform maintains strict quality controls with comprehensive tests.
- Uses in-memory SQLite (`:memory:`) via `github.com/glebarez/sqlite` for database mocking in unit tests.
- Targets >80% code coverage across Repository, Service, and API Handler layers.
Run tests: `make test-unit` or `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`
