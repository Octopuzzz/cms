# CMS Backend Platform

This platform is a production-grade Backend-as-a-Service (Dynamic CMS + API Builder) written in Go, functioning similarly to Hasura or Supabase. It allows users to dynamically create backend services, schemas, APIs (REST and optional GraphQL), manage database connections, schema migrations, and more.

## System Architecture Explanation

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles.

*   **Presentation Layer**: The Gin HTTP Router and Handlers mapping requests to internal services, including GraphQL endpoints via `gqlgen`.
*   **Application Layer**: Business logic via Services controlling data access, dynamic schemas, and operations requested by handlers.
*   **Domain Layer**: Data models defining core entities like `Service`, `Field`, `DatabaseConnection`, `User`, and `Role`.
*   **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (OpenTelemetry), metrics (Prometheus), and logging (Zap).

### High Level Platform Architecture

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

The CMS stores platform metadata, primarily in PostgreSQL (or SQLite/MySQL depending on configuration). Core tables include:

*   **database_connections**: Stores external DB configurations with id, name, type, host, port, credentials, and pooling settings.
*   **services**: User-created data models referencing `database_connection_id` and tracking dynamic schemas (`db_table_name`).
*   **fields**: Defines attributes for each service (type, nullable, unique, default_value).
*   **relations**: Connects services for relationships (e.g., One-to-Many, Many-to-Many).
*   **service_permissions**: Role-Based Access Control and Row-Level security mappings.
*   **migrations**: Tracks DDL executions and rollback metadata.
*   **backups**: Stores snapshot records for service backup and restore operations.
*   **users & roles**: System authentication and authorization.

## Go Project Structure

The project uses a clean, modular structure:

```
.
├── cmd
│   └── server                # Application entrypoint
├── docker                    # Docker configurations
├── internal
│   ├── config                # Configuration loading
│   ├── database              # DB connection manager and routing
│   ├── graphql               # GraphQL gateway and resolvers
│   ├── handlers              # Presentation layer (REST and API)
│   ├── middleware            # HTTP middlewares (Auth, Tracing, Metrics)
│   ├── models                # Domain entities and structs
│   ├── services              # Application business logic
│   └── tracing               # Observability (OpenTelemetry)
├── pkg
│   ├── logger                # Zap structured logging
│   └── response              # Standardized API responses
└── tests
    ├── integration
    ├── uat
    └── unit
```

## CRUD Engine Implementation

The Dynamic CRUD Engine automatically exposes endpoints (Create, Read, Update, Delete, List) for registered services via `DynamicDataService`.

*   Endpoints map standard operations natively to GORM database queries.
*   **Query Engine Integration**: Supports advanced filtering (`?email=test@test.com`), sorting (`?sort=created_at:desc`), pagination (`?page=1&limit=20`), and join queries.
*   **RBAC & Row-Level Security**: The engine dynamically checks user role constraints and applies data filtering directly into the generated SQL query.

## Schema Migration Engine

The Schema Migration Engine provides automated, safe schema changes natively managed through GORM's `Migrator()`.

*   **Operations**: Add column, drop column, rename column, change type.
*   **Safety**: Triggers backups before execution and persists statuses in the `migrations` table for reliable history tracking and rollback functionality.

## GraphQL Gateway

The system optionally exposes a generated GraphQL Gateway utilizing `github.com/99designs/gqlgen/graphql`.

*   **Endpoint**: `POST /api/v1/graphql`
*   Automatically resolves queries and mutations mapped directly from registered Service models, allowing users to query related objects seamlessly.

## Backup System

The Backup Engine supports robust data persistence operations.

*   **Capabilities**: Allows for snapshotting full service datasets mapping dynamic table contents to JSON format backups.
*   **Endpoints**: Examples include `POST /api/v1/cms/backup/service/{service_id}` for creation and corresponding restore capabilities via JSON payload injection.

## Observability Integration

Comprehensive, out-of-the-box observability is built directly into the system layer.

*   **Logging**: Handled globally by structured **Zap** logging with embedded correlation and trace IDs.
*   **Tracing**: Handled via **OpenTelemetry (Jaeger)**, wrapping HTTP handlers, SQL database drivers, and service logic natively.
*   **Metrics**: Prometheus intercepts HTTP activity through a dedicated middleware (`internal/middleware/prometheus.go`), exporting latency, error rates, and request counts at `/metrics`.

## Unit Tests

The system maintains comprehensive automated testing, striving for a minimum of 80% coverage across Repository, Service, and API handler layers.

*   Tests utilize an in-memory SQLite database (`:memory:`) via `github.com/glebarez/sqlite`.
*   Run tests explicitly setting the `-coverpkg` flag to capture accurate total codebase coverage: `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`
*   Available Make commands: `make test-unit`, `make test-integration`, `make test-uat`, `make test-coverage`.

## Docker Setup

Containerization allows for quick production deployment scaling.

*   **Dockerfile**: Multi-stage build process optimized for minimal binary size (`go build -ldflags "-s -w"`).
*   **Docker Compose**: Easily orchestrates the main server along with supporting containers (e.g., PostgreSQL, Redis, Jaeger) using `make docker-up` / `make docker-down`.

## Swagger Documentation

REST endpoints and core control plane APIs are automatically documented via Swagger/OpenAPI.

*   **Generation**: Handled by the `swaggo/swag` CLI. Execute `make swagger` to parse internal comments and regenerate `docs/`.
*   The generated output provides standard OpenAPI 2.0 schemas for seamless client generation and UI exploration.
