# Backend Platform (Dynamic CMS + API Builder)

A production-grade, self-hosted, Go-native Backend-as-a-Service (BaaS) platform that dynamically generates REST and GraphQL APIs.

This platform allows developers to:
- Connect databases
- Create data models dynamically
- Generate CRUD APIs automatically
- Generate optional GraphQL APIs
- Manage schema migrations
- Monitor logs and performance
- Manage backups
- Scale services

## System Architecture

The platform follows **Clean Architecture** and **Domain-Driven Design (DDD)**.

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

The core internal configuration is stored in the **Metadata Database** (e.g., PostgreSQL). Tables include:

1. `database_connections`: External DB configurations (id, name, type, host, port, credentials).
2. `services`: User-created data models referencing `database_connection_id` and tracking dynamic schemas.
3. `fields`: Defines attributes for each service (type, uniqueness, nullability, defaults).
4. `relations`: Defines one-to-one, one-to-many, many-to-one, and many-to-many relations between services.
5. `service_permissions`: Connects `Role` to `Service` for fine-grained access control.
6. `migrations`: Tracks DDL executions with metadata required for rollbacks or history tracking.
7. `backups`: Stores snapshot records or schema outputs.
8. `users` and `roles`: General authentication and authorization for the control plane.
9. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

A clean modular structure is used:

```text
.
├── cmd
│   └── server          # Application entrypoint
├── docker              # Docker and Docker Compose configuration
├── internal
│   ├── config          # Application configuration
│   ├── database        # Database connection manager
│   ├── graphql         # GraphQL gateway (gqlgen)
│   ├── handlers        # HTTP handlers (Presentation Layer)
│   ├── middleware      # Gin middleware (Auth, Prometheus, Rate Limiter)
│   ├── models          # Domain models (Domain Layer)
│   ├── services        # Business logic (Application Layer)
│   └── tracing         # OpenTelemetry setup
├── pkg
│   ├── logger          # Zap structured logger
│   └── response        # Standardized HTTP responses
└── tests
    ├── integration     # Integration tests
    ├── uat             # User Acceptance Tests
    └── unit            # Unit tests
```

## CRUD Engine & Query Engine

Managed by `internal/services/dynamic_data_service.go`.
- Automatically generates generic CRUD endpoints when a service is created.
- Maps endpoints (e.g., `GET /api/v1/data/{slug}`) to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g. `?email=test@test.com`), sorting (`?sort=created_at:desc`), pagination (`?page=1&limit=20`), and dynamic joining via foreign key mappings.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## Schema Migration Engine

Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s auto-migrator.
- Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.

## GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- Automatically generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models, supporting queries, mutations, and relations.

## Backup System

Managed by `internal/services/backup_service.go`.
- Generates data snapshots. Implements generic service record backup features mapping dynamic table contents to snapshots (JSON payloads).
- Supports restoring rows via JSON payload decoding.

## Observability Integration

Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.

## Unit Tests

The system includes unit tests with a minimum coverage requirement of 80% across the Repository, Service, and API handlers.
- Uses standard Go testing library with an in-memory SQLite setup via `github.com/glebarez/sqlite`.
- Commands:
  - Unit tests: `make test-unit`
  - Integration tests: `make test-integration`
  - UAT tests: `make test-uat`
  - Full coverage: `make test-coverage`

## Docker Setup

The platform is containerized using Docker and Docker Compose.
- `Dockerfile` provides a multistage build for the Go binary.
- `docker-compose.yml` sets up the CMS Backend, a PostgreSQL metadata database, Redis for caching, Jaeger for tracing, and Prometheus for metrics.

## API Documentation (Swagger/OpenAPI)

Swagger documentation is automatically generated for the REST APIs.
- Generate with: `~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`
- Exposed via the Gin router to provide an interactive API explorer for the CMS Control Plane and dynamically generated endpoints.
