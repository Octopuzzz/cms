# Dynamic CMS + API Builder (Go)

A self-hosted, production-grade Backend-as-a-Service (BaaS) platform written in Go. This system is a dynamic backend infrastructure generator that automatically creates data models, generates REST and GraphQL APIs, manages schema migrations, and provides integrated observability and database connections.

## Features

- **Connect Databases**: Native support for PostgreSQL, MySQL, and MongoDB.
- **Dynamic Data Models**: Create services and define schemas dynamically.
- **Automatic CRUD APIs**: Instant REST endpoints for data creation, reading, updating, and deleting.
- **Optional GraphQL**: Auto-generated GraphQL APIs using gqlgen.
- **Schema Migrations**: Safe DDL changes (add, drop, rename) with rollback support.
- **Observability Built-in**: Prometheus metrics, Jaeger/OpenTelemetry tracing, and Zap structured logging.
- **Automated Backups**: Table backups, schema snapshots, and data restores.
- **Security & RBAC**: Granular role-based access control and row-level security.

## Expected Output Documentation

As per the project specification, below is the comprehensive documentation of the platform's core components and features.

### 1. System Architecture Explanation
The CMS Backend is built on **Clean Architecture** and **Domain-Driven Design (DDD)** principles to guarantee scalability and modularity.

*   **Presentation Layer**: Located in `internal/handlers/` and `internal/graphql/`. It handles incoming REST/GraphQL requests, applies Gin middlewares, and routes them to business logic.
*   **Application Layer**: Located in `internal/services/`. It encapsulates business rules (e.g., dynamic schema generation, metadata management, migration coordination).
*   **Domain Layer**: Located in `internal/models/`. It defines core entities like `Service`, `Field`, `DatabaseConnection`, `Role`, etc.
*   **Infrastructure Layer**: Located in `internal/database/`, `pkg/logger/`, `internal/tracing/`, managing external concerns like database connections, telemetry, caching, and logs.

(For more details, see `ARCHITECTURE.md`).

### 2. Metadata Database Schema
The metadata database (default: PostgreSQL) acts as the control plane for the CMS. The core tables defined in `internal/models/models.go` include:

*   **`database_connections`**: Stores external DB configurations (`id`, `name`, `type`, `host`, `port`, `credentials`).
*   **`services`**: Represents dynamic user-created data models (`id`, `name`, `database_connection_id`, `db_table_name`).
*   **`fields`**: Configures attributes for services (`id`, `service_id`, `name`, `type`, `is_nullable`, `is_unique`).
*   **`service_permissions`**: Maps roles to services with field-level permissions.
*   **`migrations`**: Tracks DDL executions and snapshot history.
*   **`backups`**: Stores snapshot records or schema outputs.
*   **`users` / `roles`**: Manages authentication and RBAC.

### 3. Go Project Structure
The repository follows standard Go project conventions:

```text
├── cmd/
│   └── server/          # Main entrypoint (`main.go`) and application wiring
├── internal/
│   ├── api/             # API routing/definitions
│   ├── config/          # Configuration management
│   ├── database/        # Database Connection Manager
│   ├── graphql/         # GraphQL gateway (`gateway.go`)
│   ├── handlers/        # Gin HTTP controllers
│   ├── middleware/      # Rate limiters, prometheus, CORS, auth
│   ├── models/          # Domain metadata schema definition
│   ├── services/        # Core business logic (CRUD Engine, Service Builder, Migrations, etc.)
│   └── tracing/         # OpenTelemetry / Jaeger initialization
├── pkg/
│   ├── logger/          # Structured Zap logger implementation
│   └── response/        # Standardized HTTP response helpers
├── tests/               # Unit, Integration, and UAT test suites
├── docker/              # Docker setup config (Prometheus config)
├── docker-compose.yml   # Multi-container orchestration
├── Dockerfile           # App containerization
└── Makefile             # Standardization of build, test, and dev tasks
```

### 4. CRUD Engine Implementation
Located in `internal/services/dynamic_data_service.go`, the **Dynamic CRUD Engine** acts as the core interpreter for dynamic data.
*   Maps standard endpoints (e.g., `GET /api/v1/data/{slug}`) to on-the-fly GORM database operations.
*   Translates query parameters (e.g., `?sort=created_at:desc`, `?page=1&limit=20`) into complex SQL clauses natively.
*   Automatically resolves relational joins via `?join=users` using the Relation Engine based on metadata defined in `models.Field`.
*   Includes built-in validation rules (e.g., min, max, email) and role-based row-level filters.

### 5. Schema Migration Engine
Located in `internal/services/migration_service.go`, the **Schema Migration Engine** guarantees safe model schema changes.
*   Interprets user requests to mutate service schemas via `MigrationRequest`.
*   Applies operations (`add_column`, `drop_column`, `rename_column`, `change_type`) safely on the remote connected database.
*   Creates a `SchemaSnapshot` automatically before running the DDL, enabling status tracking (`pending`, `applied`, `failed`, `rolled_back`).

### 6. GraphQL Gateway
Located in `internal/graphql/gateway.go`, the platform generates optional GraphQL APIs automatically.
*   Utilizes `gqlgen` to present a unified gateway that interprets dynamic service definitions.
*   Exposes `POST /api/v1/graphql` allowing dynamic `queries` and `mutations` linked to the Dynamic CRUD engine.
*   Offers a GraphQL playground natively at `/api/v1/graphql/playground`.

### 7. Backup System
Located in `internal/services/backup_service.go`.
*   Provides granular snapshot generation mapping dynamic tables to structured backups.
*   Capable of taking Table Backups, Schema Snapshots, and full Service Snapshots.
*   Maintains status states and row counts for transparency and safety.
*   Supports restoring capabilities from previous metadata state snapshots.

### 8. Observability Integration
The platform offers full-scale observability required for modern distributed systems:
*   **Logging**: Handled by `pkg/logger/logger.go`, leveraging structured JSON logging (with trace and correlation IDs injected into context).
*   **Tracing**: Implemented via OpenTelemetry & Jaeger (`internal/tracing/`), wrapping network calls and GORM executions.
*   **Metrics**: Prometheus (`internal/middleware/prometheus.go`) exports critical usage metrics (latency, HTTP statuses) via `/metrics`.

### 9. Unit Tests
The project features a full suite of automated tests enforcing a minimum of **80% coverage** across Repository, Service, and API handler layers.
*   Uses in-memory SQLite (`:memory:`) to emulate and verify database behaviors safely.
*   Includes Unit Tests (`tests/unit/`), Integration Tests (`tests/integration/`), and User Acceptance Tests (`tests/uat/`).
*   Run tests and check coverage via:
    ```bash
    make test-coverage
    # or manually
    go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
    ```

### 10. Docker Setup
Fully containerized for simple distribution and deployment.
*   `Dockerfile`: Provides an optimized multi-stage build.
*   `docker-compose.yml`: Spins up the CMS Backend with all dependencies (`PostgreSQL`, `Redis`, `Jaeger`, `Prometheus`) natively linked via internal networks.
*   Deploy using: `make docker-up` or `docker-compose up -d`.

### 11. Swagger Documentation
Comprehensive API documentation for the generated REST APIs.
*   Annotations natively embedded in Go HTTP controllers (e.g. `cmd/server/main.go`).
*   To generate the latest Swagger configuration, run:
    ```bash
    make swagger
    # which executes: swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
    ```
*   Docs are accessible via the configured web server at `/swagger/index.html`.
