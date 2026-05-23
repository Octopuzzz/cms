# Dynamic CMS Backend Platform

## Overview

This is a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. The platform works similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native.

It acts as a backend infrastructure generator, allowing users to:
- Connect databases (PostgreSQL, MySQL, MongoDB, SQLite).
- Create data models dynamically via a metadata-driven approach.
- Generate CRUD APIs automatically.
- Generate optional GraphQL endpoints dynamically.
- Manage schema migrations.
- Manage backups.
- Monitor logs and performance via observability tools.
- Scale services.

## System Architecture Explanation

The system follows **Clean Architecture** and **Domain Driven Design (DDD)** principles, structured into the following layers:

1.  **Presentation Layer**: Implemented using the Gin HTTP framework (`internal/handlers`). Maps HTTP/REST requests (and GraphQL operations via `gqlgen` in `internal/graphql`) to the underlying services. Includes middleware for auth, rate limiting, and Prometheus metrics.
2.  **Application Layer**: Contains business logic (`internal/services`). Orchestrates the generation of schemas, database management, backups, and dynamic queries.
3.  **Domain Layer**: Data models defining the core entities like `DatabaseConnection`, `Service`, `Field`, `Permission`, `Backup`, etc. (`internal/models`).
4.  **Infrastructure Layer**: Cross-cutting concerns such as logging (Zap via `pkg/logger`), tracing (OpenTelemetry via `internal/tracing`), metrics (Prometheus), and database connection pooling/management (`internal/database`).

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

The CMS stores platform metadata in the metadata database (typically PostgreSQL or SQLite for local). Core tables include:

-   `database_connections`: Manages external database credentials, pooling, and configuration (id, name, type, host, port, etc.).
-   `services`: Represents dynamically created models/tables. Tracks `name`, `slug`, `db_table_name`, and references the `database_connection_id`.
-   `fields`: Defines attributes for services. Tracks `name`, `type`, `is_nullable`, `is_unique`, `default_value`, and options.
-   `relations`: Supported via `RelationConfig` inside `fields` for One-to-One, One-to-Many, Many-to-One, and Many-to-Many.
-   `migrations`: Records migration history, status (`pending`, `applied`, `failed`, `rolled_back`), and schema snapshots.
-   `backups`: Stores data/schema snapshots for services.
-   `users`, `roles`, `permissions`: Handles authentication, authorization, and RBAC.

## Go Project Structure

The project uses a clean and modular structure:
```text
.
├── cmd
│   └── server
│       └── main.go                  # Entry point
├── docker
│   ├── Dockerfile
│   ├── docker-compose.yml           # Local development infrastructure
│   └── prometheus.yml
├── internal
│   ├── config                       # Environment & settings
│   ├── database                     # Database connection manager & pooling
│   ├── graphql                      # Optional GraphQL Gateway (gqlgen)
│   ├── handlers                     # REST API controllers
│   ├── middleware                   # Auth, metrics, rate limiting
│   ├── models                       # Domain models and schema definitions
│   ├── services                     # Core business logic engines
│   └── tracing                      # OpenTelemetry integration
├── pkg
│   ├── logger                       # Zap structured logging
│   └── response                     # Standardized HTTP responses
└── tests                            # Unit, integration, and UAT tests
```

## Core Engines Implementation

### CRUD Engine & Query Engine
Managed primarily via `internal/services/dynamic_data_service.go` and exposed via `internal/handlers/dynamic_data_handler.go`.
-   **Dynamic Endpoints**: Auto-generates standard CRUD routes like `GET /api/v1/data/{slug}`, `POST /api/v1/data/{slug}`, `PUT /api/v1/data/{slug}/{id}`, `DELETE /api/v1/data/{slug}/{id}`.
-   **Advanced Queries**: Supports filtering (`?email=test@test.com`), sorting (`?sort=created_at:desc`), pagination (`?page=1&limit=20`), and relational joins.
-   **Mapping**: Uses GORM to dynamically map JSON payloads to the corresponding target databases and tables.

### Schema Migration Engine
Managed via `internal/services/migration_service.go`.
-   When a service or field definition changes, the system tracks and applies schema modifications (Add Column, Drop Column, Rename Column).
-   Utilizes GORM's `Migrator` for executing DDL safely.
-   Records status in the `migrations` table and handles failure logging and rollbacks via schema snapshots.

### GraphQL Gateway
Managed via `internal/graphql/gateway.go`.
-   Automatically inspects `services` and `fields` from the metadata DB.
-   Dynamically constructs a GraphQL schema and executable using `github.com/99designs/gqlgen`.
-   Resolves operations (queries, mutations) translating them into dynamic data service calls.

### Backup System
Managed via `internal/services/backup_service.go`.
-   Provides functionality to create snapshot backups of specific services (table data and schema).
-   Backup types: Data, Schema, Snapshot.
-   Backups are stored as JSON/SQL structures within the `backups` table for ease of restoration.

### Observability Integration
The platform incorporates deep observability suitable for production:
-   **Logging**: Structured logging via Uber's `zap` (wrapped in `pkg/logger`), including correlation IDs.
-   **Metrics**: Integrated Prometheus metrics via `internal/middleware/prometheus.go` to track request latency, rates, and errors. Exported on `/metrics`.
-   **Tracing**: OpenTelemetry (`internal/tracing`) integrated with Jaeger. Traces cross boundaries (HTTP Handlers -> Services -> DB).

## Unit Testing
The project ensures strict code quality through comprehensive unit, integration, and UAT testing, aiming for >80% coverage.
-   Tests are located in `tests/unit/`, `tests/integration/`, and `tests/uat/`.
-   Run tests via the Makefile: `make test-unit`, `make test-integration`, `make test-uat`, or full coverage via `make test-coverage`.

## Docker Setup
The platform is containerized for cloud deployment.
-   `Dockerfile` defines a multi-stage build to compile the Go binary compactly.
-   `docker-compose.yml` provides a local stack orchestrating the API backend, Postgres metadata database, Redis (for caching), and optional Jaeger/Prometheus tools.

## Swagger Documentation
REST API endpoints are documented using Swagger/OpenAPI.
-   Generates the docs via: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
-   (Note: generated `docs/` folder is ignored in version control per project guidelines).
