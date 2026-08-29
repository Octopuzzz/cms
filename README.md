# CMS Backend Platform

This is a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go.
It works similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native.
The platform allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.
It is modular, scalable, and cloud-ready.

## System Architecture Explanation

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

### High-Level Platform Architecture

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

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database** (typically PostgreSQL). Key tables include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials.
2. `services`: Represents user-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service`, providing fine-grained access checks.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.
9. `custom_validations`: Reusable validation rules for fields.
10. `api_tokens`: API tokens for access.

## Go Project Structure

The project uses a clean, modular structure:

```
.
├── cmd/
│   └── server/          # Main application entrypoint
├── docker/              # Docker configurations (Prometheus, etc.)
├── internal/
│   ├── config/          # Configuration loading
│   ├── database/        # Database connection management
│   ├── graphql/         # GraphQL schema gateway
│   ├── handlers/        # HTTP handlers (Presentation Layer)
│   ├── middleware/      # Gin middlewares (Auth, Rate Limiting, Metrics)
│   ├── models/          # Domain models
│   ├── services/        # Business logic (Application Layer)
│   └── tracing/         # OpenTelemetry setup
├── pkg/
│   ├── logger/          # Structured logger
│   └── response/        # Standardized API responses
├── tests/               # Unit, integration, and UAT tests
│   ├── integration/
│   ├── uat/
│   └── unit/
├── .env.example
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`, the **Dynamic CRUD Engine**:

- Automatically generates CRUD endpoints when a service is created.
- Maps dynamic endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations.
- Supports advanced queries (filtering via query params `?email=...`, dynamic joining, sorting, pagination).
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## Schema Migration Engine

Managed by `internal/services/migration_service.go`, the **Schema Migration Engine**:

- Supports safe schema changes (Add column, Drop column, Rename column, Change column type).
- Tracks and applies safe schema diffs using GORM’s `.Migrator()`.
- Logs migration status natively into the `migrations` table.
- Supports rollback capabilities.

## GraphQL Gateway

Managed by `internal/graphql/gateway.go`, the **Optional GraphQL Gateway**:

- Implements an executable schema utilizing `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.
- Provides a GraphQL Playground at `GET /api/v1/graphql/playground`.

## Backup System

Managed by `internal/services/backup_service.go`, the **Backup Engine**:

- Creates snapshot backups (schema + data) for dynamic services.
- Stores backup metadata in the `backups` table.
- Enables restoring service data from saved snapshots.

## Observability Integration

The platform includes a robust **Observability Engine**:

- **Logging**: Zap structured logging injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into context.
- **Tracing**: OpenTelemetry (`internal/tracing/`) wrapped over SQL commands and network logic. Exported to Jaeger.
- **Metrics**: Prometheus metrics exported at `/metrics`. `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors.

## Unit Tests

The system includes comprehensive tests covering Repository, Service, and API handler layers.

**Minimum Coverage Requirement:** 80%

To run all tests and generate a coverage report, use:
```sh
make test-coverage
```
Or use: `make test-unit`, `make test-integration`, `make test-uat` for specific test suites.

## Docker Setup

The platform is containerized using Docker and Docker Compose.
The setup includes services for the Backend, PostgreSQL (metadata DB), Redis (cache), Jaeger (tracing), and Prometheus (metrics).

- **Dockerfile**: Located in the root directory. Uses a multi-stage build starting with `golang:1.23-alpine`.
- **docker-compose.yml**: Brings up all dependencies in a single `cms-network`.

**Commands:**
- Start: `make docker-up` or `docker-compose up -d`
- Stop: `make docker-down` or `docker-compose down`

## Swagger Documentation

The API Documentation is generated using Swagger/OpenAPI.

To generate the documentation, run:
```sh
make swagger
```
This uses `swag init` to parse comments in `cmd/server/main.go` and handlers, generating docs in the `docs/` folder.
The API documentation is then accessible via `GET /api/v1/swagger/*any` on the running server.
