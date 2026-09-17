# Backend Platform (Dynamic CMS + API Builder)

A production-grade Backend-as-a-Service platform (Dynamic CMS + API Builder) written in Go. This system acts as a backend infrastructure generator, allowing users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles.

### High-Level Platform Architecture

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

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. It tracks system configuration, dynamically created models, connections, and system state.

Core tables include:

- **`database_connections`**: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
- **`services`**: Represents user-created data models (services). Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`). Fields: `id`, `name`, `database_id`, `created_at`, `updated_at`.
- **`fields`**: Defines attributes for each service (e.g., string, integer, float, uuid, JSON). Configures uniqueness, nullability, defaults, indexing.
- **`service_permissions`**: Connects `Role` to `Service` providing fine-grained access checks (e.g., CanCreate, CanRead).
- **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- **`backups`**: Stores snapshot records or schema outputs.
- **`users`** and **`roles`**: Authentication and authorization (RBAC).

## 3. Go Project Structure

The project uses a clean, modular structure:

```text
.
├── cmd/
│   └── server/             # Application entrypoint
├── internal/
│   ├── config/             # Configuration management
│   ├── database/           # Database connection & pooling
│   ├── graphql/            # GraphQL gateway implementation
│   ├── handlers/           # HTTP handlers (Presentation)
│   ├── middleware/         # Auth, Prometheus, OpenTelemetry, Rate limiting
│   ├── models/             # Domain layer & schema definitions
│   ├── services/           # Application business logic
│   └── tracing/            # OpenTelemetry integration
├── pkg/
│   ├── logger/             # Zap structured logging
│   └── response/           # Standardized API responses
├── tests/
│   ├── integration/        # Integration tests
│   ├── uat/                # User Acceptance tests
│   └── unit/               # Unit tests
├── docker/                 # Dockerfile and compose configurations
├── Dockerfile              # App Dockerfile
└── Makefile                # Standardized tasks
```

## 4. CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`, the **Dynamic CRUD Engine** provides out-of-the-box endpoints for user-defined services without requiring manual handler creation.

- Operations: Create, Read, Update, Delete, List.
- Endpoint mapping via standard REST conventions:
  - `POST /api/v1/data/{slug}`
  - `GET /api/v1/data/{slug}/{id}`
  - `PUT /api/v1/data/{slug}/{id}`
  - `DELETE /api/v1/data/{slug}/{id}`
  - `GET /api/v1/data/{slug}`
- Automatically applies filtering, pagination, and joins based on HTTP query parameters (Query Engine capabilities).
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`, the Migration Engine supports safe changes to user-defined databases.

- Tracks and applies safe schema diffs (Add, Drop, Rename Column, Change column type) using GORM’s `.Migrator()`.
- Logs migration status natively into the `migrations` metadata table.
- Supports automatic backups and rollbacks before migrations to ensure data safety.

## 6. GraphQL Gateway

A dynamically generated GraphQL layer complements the REST APIs, handled in `internal/graphql/gateway.go`.

- Uses `gqlgen` for generation.
- Generates optional generic GraphQL schemas based on service definitions.
- Exposes `POST /api/v1/graphql` for user querying.
- Supports `queries`, `mutations`, and deeply nested `relations` reflecting dynamic schemas.

## 7. Backup System

Managed by `internal/services/backup_service.go`, this engine provides comprehensive data preservation capabilities.

- Generates data snapshots (table, schema, or service snapshot).
- Output formats include JSON snapshots and SQL dumps.
- APIs exist to trigger backups (`POST /cms/backup/service/{service_id}`) and initiate restores (`POST /cms/restore/service/{service_id}`) directly from the control plane.

## 8. Observability Integration

Comprehensive tracing, logging, and metrics are built into the platform:

- **Logging**: Global Zap structured logging (`pkg/logger/`) includes request details, error logs, and query logs. Integrates trace/correlation IDs.
- **Metrics**: Prometheus middleware (`internal/middleware/prometheus.go`) tracks request latency, status codes, error rates, and slow queries. Exported at `/metrics`.
- **Tracing**: OpenTelemetry/Jaeger (`internal/tracing/`) initialization wraps network logic and database queries for end-to-end tracing.

## 9. Unit Tests

Unit testing is fundamental, ensuring stability across Repository, Service, and API handler layers.

- Minimum coverage target: **80%**.
- Testing frameworks use the standard Go `testing` library, combined with in-memory SQLite (`:memory:`) via the `github.com/glebarez/sqlite` driver for DB mocking.
- Execute unit tests with `make test-unit` or `go test -v -race ./tests/unit/...`.
- Execute full test suite coverage with `make test-coverage`.

## 10. Docker Setup

Containerized for scalability and deployment ease:

- Application relies on a multi-stage `Dockerfile` to build minimal Go binaries.
- Local development is orchestrated with `docker-compose.yml`, spinning up the application, database dependencies (e.g., PostgreSQL), Prometheus for metrics collection, and other required infrastructure.
- Commands: `make docker-up`, `make docker-down`, `make docker-build`.

## 11. Swagger Documentation

API Documentation is auto-generated using standard OpenAPI specifications.

- Implemented via `swaggo/swag`.
- The generation builds out the standard `docs/` folder.
- Execute the `make swagger` command to build and parse dependencies. (Internally runs: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`).
