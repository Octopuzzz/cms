# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade, self-hosted Backend-as-a-Service platform written in Go. It enables developers to dynamically create backend services, define schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

The system is modular, scalable, and cloud-ready, functioning similarly to Hasura or Supabase but is fully self-hosted and Go-native.

## System Architecture Explanation

The platform follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles to ensure separation of concerns and maintainability.

*   **Presentation Layer**: The API Gateway maps requests to internal services via a Gin HTTP router (`internal/handlers/`). This includes the GraphQL endpoint utilizing `gqlgen`.
*   **Application Layer**: Core business logic (`internal/services/`). The Services manage data access, execute dynamic schema construction, and orchestrate the platform's various capabilities (migrations, CRUD engines).
*   **Domain Layer**: Data models defining core platform entities like `Service`, `Field`, `DatabaseConnection`, `User`, and `Role` (`internal/models/`).
*   **Infrastructure Layer**: Cross-cutting tools spanning multiple layers, such as connection management/pooling (`internal/database/`), observability components (`pkg/logger/`, `internal/tracing/`, `internal/middleware/`), and external adapters.

### High Level Architecture Flow

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

The CMS manages platform configuration in a core relational Metadata Database (SQLite default for dev, easily substitutable).

Core tables include:

1.  **`database_connections`**: Stores configurations for external databases (PostgreSQL, MySQL, MongoDB), including host, port, credentials, and connection pool properties.
2.  **`services`**: Represents dynamic data models created by users. Each service tracks its physical `db_table_name` and points to a `database_connection_id`.
3.  **`fields`**: Defines the attributes for each service (string, integer, float, uuid, JSON). It also tracks validation rules, nullability, uniqueness, and defaults.
4.  **`service_permissions`**: Links a `Role` to a `Service`, providing granular access control rules (e.g., CanCreate, CanRead, row-level filters).
5.  **`migrations`**: Logs DDL operations for history tracking and rollback functionality.
6.  **`backups`**: Stores references and metadata for database or snapshot backups.
7.  **`users`** and **`roles`**: RBAC entities used by the control plane.
8.  **`audit_logs`**: Detailed logging of control plane and schema alterations.

## Go Project Structure

The codebase is highly modularized, adopting a standard Go project layout:

```text
.
├── cmd
│   └── server          # Entry point for the application
├── internal
│   ├── config          # Environment configuration loading
│   ├── database        # Database Connection Manager & Connection pooling
│   ├── graphql         # GraphQL Gateway schema and resolver generation
│   ├── handlers        # Gin presentation layer (Presentation Layer)
│   ├── middleware      # Gin middlewares (Auth, Rate Limiting, Metrics)
│   ├── models          # Domain models (Domain Layer)
│   ├── services        # Application core logic (Application Layer)
│   └── tracing         # OpenTelemetry setup
├── pkg
│   ├── logger          # Structured zap logger with context support
│   └── response        # Standardized HTTP response structures
├── tests
│   ├── integration     # Service integration tests
│   ├── uat             # User Acceptance Testing for API surface
│   └── unit            # Unit tests for domain and application layers
└── docker              # Containerization assets
```

## CRUD Engine & Query Engine Implementation

The **Dynamic CRUD Engine** and **Query Engine** are managed by `internal/services/dynamic_data_service.go`.

*   **CRUD Operations**: Handlers map dynamically generated service endpoints (e.g., `GET /api/v1/data/{service_slug}`) to underlying standard GORM operations, dynamically mapping payloads to the service's defined table schema.
*   **Query Engine**: Applies advanced queries based on HTTP request parameters. It automates filtering (e.g., `?email=user@domain.com`), joining data based on defined entity relationships, sorting (`?sort=created_at:desc`), and pagination (`?page=1&limit=20`).
*   **Security Integration**: Enforces RBAC checks and appends row-level filters seamlessly into the database query builder prior to execution.

## Schema Migration Engine

Implemented in `internal/services/migration_service.go`, the **Schema Migration Engine** provides safe database schema updates for dynamically created models.

*   **Diff Tracking**: Tracks schema differences (adding, dropping, or renaming columns) and safely executes DDL statements using GORM's built-in `Migrator()`.
*   **Safety Features**: Integrates with the metadata schema to log execution statuses and supports snapshot retention logic, paving the way for safe rollback mechanisms.

## GraphQL Gateway

The **GraphQL Gateway** is built using `github.com/99designs/gqlgen` and resides in `internal/graphql/gateway.go`.

*   **Dynamic Generation**: Provides a dynamic schema integration point, allowing defined REST metadata services to be queried over a unified generic GraphQL endpoint (`POST /api/v1/graphql`).
*   **Capabilities**: Structured to support dynamic generation of schemas containing generic models to expose underlying service data.

## Backup System

The **Backup Engine** (`internal/services/backup_service.go`) enables taking snapshots of defined services.

*   **Snapshot Format**: Leverages JSON exports and schema definitions to capture the current state of a dynamic service.
*   **Restore Support**: Provides functionality for decoding saved JSON payloads and restoring tabular states in a generic manner.

## Observability Integration

The platform provides complete telemetry covering Logs, Metrics, and Tracing out-of-the-box.

*   **Logging**: Powered by Zap and located in `pkg/logger/logger.go`. Integrates correlation and trace IDs globally via Go contexts to tie logic chains across domains.
*   **Metrics**: Handled by Prometheus in `internal/middleware/prometheus.go`. Instruments the Gin router to automatically emit standard metrics like HTTP request latency, error rates, and request counts (`/metrics`).
*   **Tracing**: Handled via OpenTelemetry/Jaeger (`internal/tracing/tracing.go`). Instrumentations span incoming HTTP requests and can wrap downstream GORM commands and external calls.

## Unit Tests

The platform mandates strict testing rules for Domain, Service, and API Layers.

*   Tests exist across multiple directories: `tests/unit/`, `tests/integration/`, and `tests/uat/`.
*   Unit tests leverage an in-memory SQLite database setup (`github.com/glebarez/sqlite`) to rapidly evaluate core Service logic without requiring external dependencies.
*   The tests aim to exceed an **80% minimum coverage** requirement.
*   Run tests using the coverage target: `make test-coverage` or `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`.

## Docker Setup

The platform is designed to be cloud-ready and containerized.

*   **Dockerfiles**: Standalone multi-stage build instructions exist in `Dockerfile` and `docker/Dockerfile`.
*   **Docker Compose**: A full local stack configuration can be spun up using `docker-compose.yml`, spinning up the API along with optional observability systems (Prometheus config in `docker/prometheus.yml`).

## Swagger Documentation

The platform includes OpenAPI / Swagger capabilities to document the core control plane and dynamic capabilities.
*   Generate the Swagger docs by running: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
*   Documentation is ignored from source control by default, allowing it to be compiled dynamically per build.
