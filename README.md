# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. It works similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native. The platform allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

The system is modular, scalable, and cloud-ready, acting as a backend infrastructure generator.

## Core Technology Stack
- **Language:** Go (1.24)
- **API Layer:** REST (default), GraphQL (optional via `gqlgen`)
- **Web Framework:** Gin
- **ORM:** GORM
- **Logging:** Zap structured logging
- **Metrics:** Prometheus
- **Tracing:** OpenTelemetry
- **Cache:** Redis
- **Message Queue:** NATS or Kafka (optional)
- **Containerization:** Docker
- **API Documentation:** Swagger / OpenAPI

## System Architecture Explanation

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

- **Presentation Layer:** The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer:** Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer:** The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer:** Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**, primarily PostgreSQL. Tables include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure:

```text
.
├── cmd/
│   └── server/          # Main application entry point (main.go)
├── docker/              # Docker and Docker Compose configuration
├── internal/
│   ├── config/          # Configuration loading and management
│   ├── database/        # Database connection management
│   ├── graphql/         # GraphQL gateway generation and endpoints
│   ├── handlers/        # HTTP presentation layer
│   ├── middleware/      # Gin middleware (Auth, Metrics, etc.)
│   ├── models/          # Domain layer and metadata schema
│   ├── services/        # Application business logic layer
│   └── tracing/         # OpenTelemetry setup
├── pkg/
│   ├── logger/          # Zap logger wrapper
│   └── response/        # Standardized HTTP response helpers
└── tests/
    ├── integration/     # Integration tests
    ├── uat/             # User Acceptance Tests
    └── unit/            # Unit tests
```

## CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`, the dynamic CRUD engine automatically generates CRUD endpoints when a service is created.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations dynamically.
- Applies automated filtering via query params (e.g. `?email=test@example.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically before running queries.

## Schema Migration Engine

Managed by `internal/services/migration_service.go`, this component handles safe schema changes:
- Tracks and applies safe schema diffs (Add, Drop, Rename Column, Change column type) using GORM’s `.Migrator()`.
- Logs migration status natively into the `migrations` table.
- Handles rollbacks through snapshot retention logic to ensure database consistency.

## GraphQL Gateway

Managed by `internal/graphql/gateway.go`:
- Automatically generates generic optional GraphQL schemas from service schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes queries, mutations, and relations dynamically through `POST /api/v1/graphql`.
- A Playground is available at `GET /api/v1/graphql/playground`.

## Backup System

Managed by `internal/services/backup_service.go`:
- Generates data snapshots (table backup, schema backup, service snapshot).
- Maps dynamic table contents to snapshots and allows restoring rows via JSON payload decoding.
- APIs exist for `POST /api/v1/cms/backup/service/{service_id}` and `POST /api/v1/cms/restore/{id}`.

## Observability Integration

Implemented across the stack to ensure comprehensive monitoring:
- **Logging:** Zap structured logging is injected globally (`pkg/logger/`) capturing Request logs, Error logs, and Query logs. Correlates with Trace IDs directly from the Context.
- **Metrics:** `internal/middleware/prometheus.go` tracks latency, error rates, and status codes. Exported via standard HTTP interceptors at `/metrics`.
- **Tracing:** OpenTelemetry/Jaeger is initialized in `internal/tracing/` to wrap SQL commands and network logic.

## Unit Tests

The system maintains a robust test suite covering Repository, Service, and API handler layers.
- **Minimum Coverage:** >= 80%
- Uses standard Go testing techniques and in-memory SQLite database (`:memory:`) via the `github.com/glebarez/sqlite` driver for database mocking.
- **Execution:** `make test-coverage` or `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`

## Docker Setup

Containerization is fully supported.
- The `Dockerfile` compiles a minimal, production-ready image.
- `docker-compose.yml` provides a unified environment containing the CMS backend, PostgreSQL, Jaeger, and other dependent services out of the box.

## Swagger Documentation

API Documentation is auto-generated using Swagger / OpenAPI.
- Generate docs locally with: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`
- Exposed at `GET /api/v1/swagger/*any`
