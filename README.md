# CMS Backend Platform

A production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. This platform acts as a backend infrastructure generator, allowing users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints. It works similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native.

## Features

- Connect external databases (PostgreSQL, MySQL, MongoDB, SQLite).
- Create data models dynamically.
- Generate CRUD APIs automatically.
- Generate optional GraphQL APIs.
- Manage schema migrations.
- Monitor logs and performance.
- Manage backups.

## System Architecture Explanation

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

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

The CMS stores platform metadata using the following core tables:

- `database_connections`: External DB configurations (id, name, type, host, port, username, password, database_name). Includes connection pooling settings.
- `services`: User-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
- `fields`: Defines attributes for each service (name, type, nullable, unique, default_value, index). Supported types: string, integer, float, boolean, uuid, json, array, timestamp.
- `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
- `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- `backups`: Stores snapshot records or schema outputs.
- `users` and `roles`: General authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean, modular structure:

```text
cmd/
└── server/               # Application entry point
internal/
├── api/                  # API routes setup
├── config/               # Environment config
├── database/             # Database connection manager
├── graphql/              # GraphQL gateway implementation
├── handlers/             # HTTP controllers
├── middleware/           # Auth, Prometheus, Rate Limiting
├── models/               # Domain models (Metadata schema)
├── services/             # Business logic (CMS, CRUD, Migrations)
└── tracing/              # OpenTelemetry integration
pkg/
├── errors/               # Custom error definitions
├── logger/               # Zap structured logging
├── pagination/           # Pagination logic
└── response/             # Standardized HTTP responses
tests/                    # Unit, Integration, UAT tests
docker/                   # Dockerfile and Docker Compose configs
```

## CRUD Engine Implementation

The Dynamic CRUD Engine automatically generates endpoints when a service is created.

**Endpoints:**
- `POST /api/v1/data/{slug}` (Create)
- `GET /api/v1/data/{slug}/{id}` (Read)
- `PUT /api/v1/data/{slug}/{id}` (Update)
- `DELETE /api/v1/data/{slug}/{id}` (Delete)
- `GET /api/v1/data/{slug}` (List)

**Query Engine:**
Supports advanced queries directly via GET parameters:
- Filtering: `?email=john@example.com`
- Sorting: `?sort=created_at:desc`
- Pagination: `?page=1&limit=20`

Role-Based Access Control and Row-Level filtering are dynamically verified prior to executing operations.

## Schema Migration Engine

Supports safe schema changes tracked and applied using GORM's `.Migrator()`.
- Add column, Drop column, Rename column, Change column type.
- Automatic backup before migration.
- Rollback capability via snapshot retention logic.
- Managed by `internal/services/migration_service.go`.

## GraphQL Gateway

A GraphQL gateway is automatically generated from service schemas using `github.com/99designs/gqlgen`.
- Exposes `POST /api/v1/graphql`.
- Supports dynamic queries, mutations, and relations for defined data models.

## Backup System

Supports generating and restoring data snapshots.
- Table backup, Schema backup, Service snapshot.
- Available formats: JSON snapshot and generic service record backups mapping dynamic table contents.
- REST endpoints: `POST /api/v1/backups/service/{service_id}` for creating and `POST /api/v1/backups/restore/{backup_id}` for restoring.

## Observability Integration

The system implements state-of-the-art observability:
- **Logging**: Zap structured logging injected globally (`pkg/logger/`) with request tracking and correlation IDs.
- **Metrics**: Prometheus metrics (`internal/middleware/prometheus.go`) track request latency, error rates, and statuses. Exported at `/metrics`.
- **Tracing**: OpenTelemetry (Jaeger) initialized in `internal/tracing/` to wrap SQL commands and network logic.

## Unit Tests

The system includes comprehensive tests with a minimum of 80% coverage.
- Test areas: Repository, Service, API handlers.
- Run tests: `make test` or `go test ./...`
- Run coverage: `make test-coverage`

Note: Unit tests use an in-memory SQLite setup via `github.com/glebarez/sqlite`.

## Docker Setup

The platform is containerized using Docker.
- `docker/Dockerfile`: Contains instructions for building the Go application.
- `docker/docker-compose.yml`: Defines the application stack including the CMS backend, Prometheus, Redis, and Jaeger for local development and deployment.
- Commands: `make docker-up` and `make docker-down`.

## Swagger Documentation

API documentation is generated using Swaggo (Swagger/OpenAPI).
- Includes parse dependency and internal endpoints.
- Generate docs using: `make swagger` or `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
