# Go Dynamic CMS & API Builder

A self-hosted, Go-native Backend-as-a-Service (BaaS) platform. This system acts as a backend infrastructure generator, allowing users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints. It is designed to be modular, scalable, and cloud-ready, functioning similarly to platforms like Hasura or Supabase.

## Table of Contents
- [System Architecture Explanation](#system-architecture-explanation)
- [Metadata Database Schema](#metadata-database-schema)
- [Go Project Structure](#go-project-structure)
- [CRUD Engine Implementation](#crud-engine-implementation)
- [Schema Migration Engine](#schema-migration-engine)
- [GraphQL Gateway](#graphql-gateway)
- [Backup System](#backup-system)
- [Observability Integration](#observability-integration)
- [Unit Tests](#unit-tests)
- [Docker Setup](#docker-setup)
- [Swagger Documentation](#swagger-documentation)

## System Architecture Explanation

The system follows **Clean Architecture and Domain-Driven Design (DDD)** principles:

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

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes optional GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database** (e.g., PostgreSQL).

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `relations`: Manages relationships (One-to-One, One-to-Many, Many-to-One, Many-to-Many).
5. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
6. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
7. `backups`: Stores snapshot records or schema outputs.
8. `users` and `roles`: General authentication and authorization for the control plane.
9. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean, modular structure:

```text
cmd/
└── server/             # Application entrypoint
internal/
├── api/                # API routes
├── config/             # Configuration management
├── database/           # DB Connection management
├── graphql/            # GraphQL gateway & schemas
├── handlers/           # HTTP handlers
├── middleware/         # Middleware (Auth, Prom, Tracing)
├── models/             # Domain data models
├── services/           # Application logic (CRUD, Migrations, etc.)
└── tracing/            # OpenTelemetry setup
pkg/
├── errors/             # Custom error definitions
├── logger/             # Zap logger wrapper
├── pagination/         # Pagination helpers
└── validation/         # Input validation
tests/                  # Unit, Integration, UAT tests
docker/                 # Dockerfile and Docker Compose
```

## CRUD Engine Implementation

When a service is created, the system automatically generates REST endpoints powered by the dynamic CRUD Engine (`internal/services/dynamic_data_service.go`).

Operations supported:
- **Create**: `POST /api/v1/data/{service}`
- **Read**: `GET /api/v1/data/{service}/{id}`
- **Update**: `PUT /api/v1/data/{service}/{id}`
- **Delete**: `DELETE /api/v1/data/{service}/{id}`
- **List/Query**: `GET /api/v1/data/{service}`

The engine maps these endpoints to standard GORM operations on the fly, applying automated filtering via query params (e.g., `?email=test@test.com`), dynamic joining via foreign key mappings, sorting (`?sort=created_at:desc`), and pagination (`?page=1&limit=20`).

## Schema Migration Engine

Managed by `internal/services/migration_service.go`, the Schema Migration Engine supports safe schema changes to dynamically constructed models.

Supported Operations:
- Add column
- Drop column
- Rename column
- Change column type

It logs migration status natively into the `migrations` metadata table and tracks and applies safe schema diffs using GORM’s `.Migrator()`. Safety features include rollback capability through snapshot retention logic.

## GraphQL Gateway

An optional GraphQL API can be exposed and is managed by `internal/graphql/gateway.go`.
It utilizes `github.com/99designs/gqlgen` to generate generic, optional schemas.
The system automatically generates GraphQL schemas from service definitions, exposing a `POST /api/v1/graphql` endpoint for clients. This gateway supports advanced querying, mutations, and dynamic relationship resolutions.

## Backup System

The Backup Engine (`internal/services/backup_service.go`) handles snapshotting and restoring data models.

Features:
- **Table / Service Snapshotting**: Generates JSON/SQL snapshots of the dynamic table contents.
- **Restore Capability**: Decodes JSON payloads and restores rows back into the respective service table.

API Examples:
- Backup: `POST /cms/backup/service/{service_id}`
- Restore: `POST /cms/restore/service/{service_id}`

## Observability Integration

The platform includes comprehensive observability out-of-the-box:

- **Logging**: Zap structured logging is injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into the Context.
- **Metrics**: Prometheus metrics track request latency, status codes, and error rates via standard HTTP middleware (`internal/middleware/prometheus.go`). Exported at `/metrics`.
- **Tracing**: OpenTelemetry (integrated with Jaeger) is initialized in `internal/tracing/` to wrap SQL commands and network logic.

## Unit Tests

The system maintains high test coverage (minimum 80%).
Unit testing utilizes standard `testing` and in-memory SQLite (`:memory:`) via the `github.com/glebarez/sqlite` driver for fast execution and database mocking.

To run the test suite:
```bash
make test-coverage
# or
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Docker Setup

The platform is containerized for easy deployment. The `docker/` directory provides everything needed for local or production deployment.

To start the platform with its dependencies:
```bash
make docker-up
# or
docker-compose -f docker/docker-compose.yml up -d
```

## Swagger Documentation

API Documentation is generated using `swaggo/swag`.

To generate or update Swagger docs:
```bash
make swagger
# or
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
