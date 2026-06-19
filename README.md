# CMS Backend Platform

A production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. This system acts as a fully self-hosted, Go-native backend infrastructure generator similar to platforms like Hasura or Supabase. It allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

## Features

- **Connect Databases:** Register external databases (PostgreSQL, MySQL, MongoDB).
- **Dynamic Data Models:** Dynamically create services and define schemas and fields.
- **Auto-generated APIs:** Automatically generate CRUD REST APIs and optional GraphQL endpoints for defined data models.
- **Relation Management:** Support for 1:1, 1:N, N:1, and N:M relationships with automatic join tables.
- **Schema Migrations:** Track and apply safe schema changes with rollback capabilities.
- **Backup System:** Snapshot service data and restore records.
- **Observability:** Built-in logging, metrics (Prometheus), and distributed tracing (OpenTelemetry).
- **Security:** Role-based access control (RBAC), JWT authentication, and rate limiting.

## System Architecture

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

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

- **Presentation Layer:** Gin HTTP router and handlers (`internal/handlers/`), and GraphQL endpoints (`internal/graphql/`).
- **Application Layer:** Business logic and service orchestration (`internal/services/`).
- **Domain Layer:** Core entity models representing services, fields, users, etc. (`internal/models/`).
- **Infrastructure Layer:** Observability, database connection pools, telemetry (`internal/tracing/`), metrics, and logging (`pkg/logger/`).

## Metadata Database Schema

The platform stores configuration and metadata in its primary database. Key tables include:

- `database_connections`: Stores external database configurations (id, name, type, host, port, credentials).
- `services`: Represents dynamically created user services/models (id, name, database_connection_id, db_table_name).
- `fields`: Attributes for each service (name, type, nullable, unique, default_value).
- `service_relations`: Defines relationships between services (one-to-many, many-to-many, etc.).
- `migrations`: Tracks database schema migrations for rollback/audit capabilities.
- `backups`: Stores backup snapshots or references.
- `users` & `roles`: General platform authentication, RBAC, and access control.

## Project Structure

```
.
├── cmd/
│   └── server/               # Main application entrypoint
├── docker/                   # Docker Compose and Prometheus configuration
├── internal/
│   ├── config/               # Application configuration
│   ├── database/             # Connection manager & external DB handlers
│   ├── graphql/              # Dynamic GraphQL gateway implementation
│   ├── handlers/             # REST API controllers
│   ├── middleware/           # HTTP middlewares (Auth, CORS, Metrics)
│   ├── models/               # Domain data models & metadata definitions
│   ├── services/             # Core business logic (CRUD, CMS, Builder, Auth)
│   └── tracing/              # OpenTelemetry and Jaeger setup
├── pkg/
│   ├── logger/               # Structured Zap logging wrapper
│   └── response/             # Standard API response formats
└── tests/
    ├── integration/          # Integration tests
    ├── uat/                  # User Acceptance Tests (E2E)
    └── unit/                 # Unit tests (Repository, Service, Handlers)
```

## CRUD & Query Engine Implementation

The dynamic CRUD engine maps dynamic endpoints (e.g., `/api/v1/data/{service_name}`) to GORM database operations on the fly:
- **Filtering:** `GET /api/v1/data/users?email=test@test.com`
- **Sorting:** `GET /api/v1/data/users?sort=created_at:desc`
- **Pagination:** `GET /api/v1/data/users?page=1&limit=20`
- **RBAC:** Enforces role-based permissions before executing queries.

## Schema Migration Engine

Track and apply safe schema diffs dynamically via `MigrationService`.
- Uses GORM's auto-migrator natively for ADD, DROP, and RENAME operations.
- Keeps track of execution history in the `migrations` table.
- Supports reverting modifications via snapshot rollback.

## GraphQL Gateway

Provides a GraphQL playground and endpoint at `/api/v1/graphql`.
- Generates GraphQL types and queries based on the dynamic services schemas.
- Built using `github.com/99designs/gqlgen/graphql`.

## Backup System

Creates snapshots of dynamic service data (`BackupService`).
- Snapshot-based logical backups of table rows into JSON formats.
- Exposes endpoints to backup (`POST /api/v1/cms/backup/service/{id}`) and restore data.

## Observability

- **Logging:** Structured JSON logging across the stack using `go.uber.org/zap`.
- **Metrics:** Prometheus metrics exposed at `/metrics` tracking request latency, response codes, and HTTP traffic.
- **Tracing:** Distributed tracing initialized with OpenTelemetry (Jaeger integration) wrapping external database logic and network paths.

## Local Development & Docker

The system is containerized and includes a `docker-compose.yml` file to spin up dependencies easily (Prometheus, Jaeger, Redis, etc., if needed, and the app itself).

Start the platform via Docker:
```bash
make docker-up
```

Alternatively, use `make dev` for local development with hot-reloading (via `air`).

## Unit Testing

The repository uses Go's standard testing toolkit with over 80% coverage requirements across the Repository, Service, and Handler layers. Database mocks use in-memory SQLite instances.

Run tests via:
```bash
make test-unit
make test-integration
make test-coverage
```

## Swagger API Documentation

Interactive Swagger (OpenAPI) documentation is auto-generated for the core platform.

1. Ensure the swag CLI is installed: `go install github.com/swaggo/swag/cmd/swag@latest`
2. Generate docs: `make swagger`
3. Access docs at `http://localhost:8080/swagger/index.html`
