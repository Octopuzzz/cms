# Dynamic CMS & API Builder Backend Platform

A production-grade, self-hosted Backend-as-a-Service platform written in Go. This platform acts as a backend infrastructure generator, allowing developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

## Table of Contents

- [System Architecture Explanation](#system-architecture-explanation)
- [Metadata Database Schema](#metadata-database-schema)
- [Go Project Structure](#go-project-structure)
- [Dynamic CRUD Engine Implementation](#dynamic-crud-engine-implementation)
- [Schema Migration Engine](#schema-migration-engine)
- [GraphQL Gateway](#graphql-gateway)
- [Backup System](#backup-system)
- [Observability Integration](#observability-integration)
- [Security](#security)
- [Development Setup & Instructions](#development-setup--instructions)

---

## System Architecture Explanation

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles to ensure scalability, testability, and modularity.

### High-Level Architecture

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

### Core Layers

1.  **Presentation Layer**: Located in `internal/handlers/` and `internal/graphql/`. It exposes a Gin-based REST API and an optional gqlgen-based GraphQL API. It handles incoming requests, authentication mapping, and delegating to services.
2.  **Application Layer**: Located in `internal/services/`. It contains the core business logic, orchestrating metadata updates, dynamic schema generation, and query execution.
3.  **Domain Layer**: Located in `internal/models/`. It defines the core data structures and metadata schemas used to manage the platform (e.g., `Service`, `Field`, `DatabaseConnection`, `Migration`).
4.  **Infrastructure Layer**: Cross-cutting concerns like database connection caching (`internal/database/`), observability (`internal/tracing/`), middleware (`internal/middleware/`), and logging (`pkg/logger/`).

---

## Metadata Database Schema

The platform stores configuration and definitions in a centralized metadata database (default: SQLite or PostgreSQL depending on environment).

### Core Tables

*   **`database_connections`**: Stores external database connection settings (PostgreSQL, MySQL, MongoDB, etc.). Fields include `id`, `name`, `type`, `host`, `port`, `username`, `password`, `database_name`.
*   **`services`**: Represents the dynamic data models created by users. Each service maps to an actual table in the registered database.
    *   Fields: `id`, `name`, `database_connection_id`, `db_table_name`, `created_at`, `updated_at`.
*   **`fields`**: Defines the schema for each `Service`. Maps attributes like strings, integers, UUIDs, and relations.
    *   Fields: `id`, `service_id`, `name`, `type` (string, integer, float, boolean, uuid, json, timestamp), `is_nullable`, `is_unique`, `default_value`, `is_index`.
*   **`service_permissions`**: Maps RBAC (Role-Based Access Control) between `roles` and `services`. It controls `CanCreate`, `CanRead`, `CanUpdate`, `CanDelete`, and holds field-level permissions and row-level filters.
*   **`migrations`**: Tracks all schema migrations applied to dynamic services, holding migration descriptions, status, schema snapshots, and applied timestamps.
*   **`backups`**: Records backup jobs (snapshots, schemas) tied to services, tracking status and snapshot data.
*   **`users`**, **`roles`**, **`user_roles`**, **`permissions`**: Standard Identity and Access Management for the control plane.
*   **`audit_logs`**: Captures platform mutations and activities.

---

## Go Project Structure

The codebase is organized modularly based on Go standard layout practices:

```text
├── cmd
│   └── server                # Application entrypoint (main.go)
├── internal                  # Private application code
│   ├── config                # Configuration parsing and environment management
│   ├── database              # Database connection management and connection pooling
│   ├── graphql               # GraphQL gateway implementation using gqlgen
│   ├── handlers              # Gin HTTP handlers for API endpoints
│   ├── middleware            # HTTP middlewares (Auth, Prometheus, Rate Limiter)
│   ├── models                # Core Domain Models (GORM schemas)
│   ├── services              # Application business logic (CRUD Engine, CMS, etc.)
│   └── tracing               # OpenTelemetry integration
├── pkg                       # Public library code
│   ├── logger                # Zap structured logging setup
│   ├── pagination            # Standardized pagination utilities
│   ├── response              # Standardized HTTP JSON responses
│   └── validation            # Input validation utilities
├── tests                     # Automated test suites
│   ├── integration           # Integration tests
│   ├── uat                   # End-to-End User Acceptance Tests
│   └── unit                  # Unit tests (mocks, logic)
├── docker                    # Docker configurations and scripts
├── Makefile                  # Build, test, and run task runner
├── go.mod                    # Go dependencies
└── Dockerfile                # Docker build definition
```

---

## Dynamic CRUD Engine Implementation

The Dynamic CRUD Engine (`internal/services/dynamic_data_service.go`) automatically fulfills data operations for dynamically created services.

### Operations Supported
When a `Service` is configured, it dynamically maps generic endpoints to underlying GORM operations:

*   **Create**: `POST /api/v1/data/{slug}`
*   **Read**: `GET /api/v1/data/{slug}/{id}`
*   **Update**: `PUT /api/v1/data/{slug}/{id}`
*   **Delete**: `DELETE /api/v1/data/{slug}/{id}`
*   **List (Query Engine)**: `GET /api/v1/data/{slug}`

### Query Engine Features
The List endpoint acts as an advanced Query Engine:
*   **Filtering**: Standard query parameters map to SQL WHERE clauses (e.g., `?email=test@example.com`).
*   **Sorting**: Supports field sorting (e.g., `?sort=created_at:desc`).
*   **Pagination**: Handled via `page` and `limit` query parameters.
*   **Security & Joining**: Incorporates dynamic joining for foreign keys based on defined relations and applies row-level security based on `service_permissions`.

---

## Schema Migration Engine

The Schema Migration Engine (`internal/services/migration_service.go`) safely handles database schema changes.

*   **GORM Integration**: Utilizes GORM's `AutoMigrate` or `Migrator` interface to apply non-destructive schema changes based on the dynamic `fields` definitions.
*   **Supported Operations**: Adding columns, dropping columns (managed carefully depending on adapter), renaming columns, and changing types.
*   **Safety**: Captures the state before migration. Records the operation in the `migrations` table with timestamps and status (pending, applied, failed).
*   **Rollback**: The architecture supports reading previous schema states from the `migrations` metadata for rollback capabilities.

---

## GraphQL Gateway

An optional GraphQL Gateway (`internal/graphql/gateway.go`) dynamically handles GraphQL operations.

*   **Implementation**: Powered by `github.com/99designs/gqlgen`.
*   **Endpoint**: Exposes `POST /api/v1/graphql`.
*   **Capabilities**: Maps GraphQL queries directly to the Dynamic CRUD Engine, allowing clients to query nested relations and specific fields without over-fetching.

---

## Backup System

The Backup Engine (`internal/services/backup_service.go`) allows snapshotting of dynamic tables.

*   **Snapshot Logic**: Extracts all rows from a dynamically created service table and encodes them as JSON.
*   **Storage**: Stores the snapshot payload in the metadata database's `backups` table alongside metadata about the operation (row count, status).
*   **Endpoints**: Handled by `BackupHandler` (`POST /api/v1/cms/backups`), enabling full backup jobs and potential restore endpoints.

---

## Observability Integration

The platform is designed to be cloud-native and highly observable.

*   **Logging**: Uses `go.uber.org/zap` for high-performance structured logging. Integrated into handlers and services. Includes request ID correlation.
*   **Tracing**: Implements OpenTelemetry (`internal/tracing/tracing.go`). Distributed traces capture HTTP requests, database queries, and service-to-service operations, exportable to Jaeger or other OTLP collectors.
*   **Metrics**: Prometheus metrics are natively integrated via middleware (`internal/middleware/prometheus.go`). It tracks request latencies, status codes, and error rates, exposing them on the `/metrics` endpoint.

---

## Security

*   **Authentication**: JWT-based authentication via `internal/middleware/auth.go`.
*   **Authorization**: Granular RBAC (Role-Based Access Control) defined in the metadata DB and verified before dynamic data access.
*   **Rate Limiting**: Integrated HTTP rate limiting (`internal/middleware/rate_limiter.go`).

---

## Development Setup & Instructions

### Prerequisites
*   Go (>=1.24)
*   Make
*   Docker & Docker Compose (optional for local infrastructure)

### Run Tasks via Makefile
The project heavily utilizes a `Makefile` for standardized task execution.

*   **Download Dependencies**:
    ```bash
    make deps
    ```
*   **Build the Application**:
    ```bash
    make build
    ```
    Outputs binary to `bin/cms-backend`.
*   **Run Development Server**:
    ```bash
    make dev
    ```
    Runs the application with hot-reloading (requires `air`). Alternatively, `make run`.

### Docker Setup
To run the platform and its required infrastructure using Docker Compose:

```bash
# Start all services in the background
make docker-up

# View logs
make docker-logs

# Shut down
make docker-down
```

### Unit Tests
The project maintains an 80%+ coverage requirement across core layers.

*   **Run All Tests**:
    ```bash
    make test
    ```
*   **Run Unit Tests**:
    ```bash
    make test-unit
    ```
*   **Generate Coverage Report**:
    ```bash
    make test-coverage
    ```
    *Note: When running tests in network-restricted environments, the system defaults to local go toolchains and SQLite memory databases (`:memory:`).*

### Swagger Documentation
The REST API is documented using Swagger. To generate or update the documentation:

```bash
make swagger
```
*Note: Requires `swag` CLI installed globally.* The generated docs are placed in the ignored `docs/` folder.
