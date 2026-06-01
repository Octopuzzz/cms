# Dynamic CMS + API Builder (Go-native BaaS)

A production-grade Backend-as-a-Service platform allowing developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability. Built fully in Go.

## 1. System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles.

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

Core layers:
*   **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
*   **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
*   **Domain Layer**: The data models (`internal/models/`). Defines core entities.
*   **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry, metrics, and logging.

## 2. Metadata Database Schema

The system stores platform metadata in the Metadata Database. Key tables include:

*   **`database_connections`**: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
*   **`services`**: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
*   **`fields`**: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
*   **`service_permissions`**: Connects `Role` to `Service` providing fine-grained access checks. Contains row-level and field-level capabilities.
*   **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
*   **`backups`**: Stores snapshot records or schema outputs.
*   **`users` / `roles`**: General authentication and authorization for the control plane.
*   **`audit_logs`**: Detailed logging of structural and data-level modifications.

## 3. Go Project Structure

The project uses a clean modular structure:

```text
.
├── cmd/
│   └── server/             # Application entrypoint
├── docker/                 # Docker configuration and prometheus setup
├── internal/
│   ├── config/             # Configuration management
│   ├── database/           # DB Connectors & Pooling
│   ├── graphql/            # GraphQL Gateway
│   ├── handlers/           # Presentation Layer (Gin)
│   ├── middleware/         # Gin middlewares (Auth, metrics, tracing)
│   ├── models/             # Domain Layer models
│   ├── services/           # Application Layer business logic
│   └── tracing/            # OpenTelemetry setup
├── pkg/
│   ├── logger/             # Zap structured logging
│   └── response/           # Standardized API responses
└── tests/
    ├── integration/
    ├── uat/
    └── unit/
```

## 4. CRUD Engine Implementation

When a service is created, the system maps endpoints to standard database operations on the fly via `DynamicDataService`.

*   Operations: Create, Read, Update, Delete, List.
*   Endpoint format: `GET /api/v1/data/{slug}`, `POST /api/v1/data/{slug}`, etc.
*   Includes automated filtering via query params (e.g. `?email=test@example.com`), dynamic joining via foreign key mappings, and pagination logic.
*   Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## 5. Schema Migration Engine

Managed by `MigrationService`.
*   Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM's `.Migrator()`.
*   Logs migration status natively into the `migrations` table and handles rollbacks through snapshot retention logic.
*   Safety features: Tracks `applied_at`, `status`, and retains a `schema_snapshot`.

## 6. GraphQL Gateway (Optional)

Automatically generated from service schemas.
*   Managed by `internal/graphql/gateway.go`.
*   Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
*   Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models, supporting queries, mutations, and relations.

## 7. Backup System

Managed by `BackupService`.
*   Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
*   Supports restoring rows via JSON payload decoding.
*   Tracks backups in the `backups` table with statuses and schema/table data storage.

## 8. Observability Integration

*   **Logging**: Zap structured logging injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
*   **Metrics**: Prometheus tracking request latency, status codes, and error rates via standard HTTP interceptors. Exported at `/metrics`.
*   **Tracing**: OpenTelemetry/Jaeger initialized in `internal/tracing/` to wrap SQL commands and network logic.

## 9. Unit Tests

The system includes unit, integration, and UAT testing with a minimum of 80% coverage requirement.
*   Command to run unit tests: `make test-unit` or `go test -v -race ./tests/unit/...`
*   Command for full coverage report: `make test-coverage`

## 10. Docker Setup

*   Start services: `docker-compose up -d` or `make docker-up`
*   Stop services: `docker-compose down` or `make docker-down`
*   The `docker-compose.yml` includes the backend service, postgres DB, and observability tools.

## 11. Swagger Documentation

Generate API documentation using the provided Make target:
*   `make swagger`
*   This uses `swaggo/swag` to generate OpenAPI docs into the `docs/` folder based on annotations in the handler files.
