# Dynamic CMS & API Builder (Backend Platform)

A production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. This platform functions similarly to systems like Hasura or Supabase, acting as a backend infrastructure generator. It is fully self-hosted, cloud-ready, and follows Clean Architecture and Domain-Driven Design (DDD).

## Core Features

- Connect external databases (PostgreSQL, MySQL, MongoDB, SQLite).
- Create data models dynamically via a CMS Control Plane.
- Automatically generate CRUD REST APIs.
- Optional GraphQL APIs generated from service schemas.
- Advanced Query Engine (Filtering, Sorting, Pagination, Joins).
- Schema Migration Engine for safe data evolution.
- Backup Engine (Table, Schema, Service snapshots).
- State-of-the-art Observability (Prometheus, OpenTelemetry, Zap).
- Security integrated (RBAC, Rate Limiting, Validation).

## System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

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

The core internal configuration is stored in the **Metadata Database** (e.g. PostgreSQL or SQLite). Core tables include:

1. **`database_connections`**: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. **`services`**: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. **`fields`**: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. **`relations`**: Manages One-to-One, One-to-Many, Many-to-One, and Many-to-Many relationships.
5. **`service_permissions`**: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
6. **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
7. **`backups`**: Stores snapshot records or schema outputs.
8. **`users`** and **`roles`**: General authentication and authorization for the control plane.
9. **`audit_logs`**: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure:

```
├── cmd
│   └── server                # Entry point
├── internal
│   ├── config                # Configuration loading
│   ├── database              # Database connections & managers
│   ├── graphql               # GraphQL schemas & resolvers
│   ├── handlers              # Gin HTTP handlers
│   ├── middleware            # Auth, Telemetry, Prometheus
│   ├── models                # GORM domain entities
│   ├── services              # Business logic (CRUD, CMS, Builder, Migration)
│   └── tracing               # OpenTelemetry implementation
├── pkg
│   ├── logger                # Zap structured logger wrapper
│   └── response              # Standardized HTTP responses
├── tests
│   ├── integration           # Database/component tests
│   ├── uat                   # End-to-end API testing
│   └── unit                  # Unit tests (Mock DB)
├── docker
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── prometheus.yml
├── Makefile                  # Build, dev, and test scripts
└── ARCHITECTURE.md           # Internal architectural docs
```

## Component Implementation

### Dynamic CRUD & Query Engine

Managed by `internal/services/dynamic_data_service.go`.
- Automatically maps endpoints like `GET /api/v1/data/{slug}` to dynamic GORM database operations.
- Supported operations: Create, Read, Update, Delete, List.
- Applies automated filtering via query params (e.g. `?email=john@example.com`).
- Supports advanced features like pagination, sorting (`?sort=created_at:desc`), and dynamic joining via foreign key mappings.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

### Schema Migration Engine

Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column, Change Type) using GORM’s `.Migrator()`.
- Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.
- Built-in safety features like automatic backups before migration.

### GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.
- Generates generic optional schemas using `github.com/99designs/gqlgen`.
- Supports nested queries, mutations, and automatic relation resolutions.

### Backup System

Managed by `internal/services/backup_service.go`.
- Supports Table backup, Schema backup, and Service snapshots.
- Generates data snapshots (SQL dump / JSON snapshot) and supports restoring rows via JSON payload decoding.
- APIs for triggering backups: `POST /api/v1/cms/backup/service/{service_id}`.

### Observability Integration

Implemented across the stack:
- **Logging**: Request logs, Error logs, and Query logs managed globally via Zap (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **Metrics**: Track request latency, slow queries, and error rates using Prometheus (`internal/middleware/prometheus.go`). Exported at `/metrics`.
- **Tracing**: OpenTelemetry/Jaeger initialized in `internal/tracing/` to wrap SQL commands and network logic.

## Docker Setup

The platform is fully containerized. A `docker-compose.yml` is provided for running the complete stack including the application, database, Prometheus, Redis, etc.

Run the platform:
```bash
docker-compose up -d
```

## Unit Testing

The system includes comprehensive unit tests focusing on Repositories, Services, and API Handlers with a minimum 80% coverage requirement.

Run the test suite:
```bash
make test-unit
# or standard go test
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Swagger Documentation

API Documentation is auto-generated using standard OpenAPI / Swagger specifications.

Generate Swagger docs:
```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
(Docs are accessed via the platform's UI or standard Swagger endpoint)
