# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend-as-a-Service platform allowing developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

The system works similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native.

## System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

### High-Level Components
- **CMS Control Plane**: Management interface to control the system (DB connections, services, schemas, relations, migrations, backups).
- **Service Builder**: Dynamically create services (data models).
- **CRUD Engine**: Automatically generate CRUD endpoints for services.
- **Query Engine**: Advanced queries (filtering, sorting, pagination, joins).
- **Schema Migration Engine**: Safe schema changes.
- **GraphQL Gateway**: Auto-generated GraphQL endpoints.
- **Backup Engine**: System backup and restore functionalities.
- **Observability Engine**: Logging, metrics, tracing.

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. The models are defined in `internal/models/models.go`.

- **`database_connections`**: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
- **`services`**: Represents user-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas.
- **`fields`**: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
- **`service_permissions`**: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
- **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- **`backups`**: Stores snapshot records or schema outputs.
- **`users`** and **`roles`**: General authentication and authorization for the control plane.
- **`audit_logs`**: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project follows a clean, modular structure:

```
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection manager and helpers
│   ├── graphql/         # GraphQL gateway and schema generation
│   ├── handlers/        # HTTP handlers (Presentation Layer)
│   ├── middleware/      # Gin middlewares (Auth, Prometheus, Rate Limiter, etc.)
│   ├── models/          # Core domain models (Domain Layer)
│   ├── services/        # Business logic and engines (Application Layer)
│   └── tracing/         # OpenTelemetry tracing
├── pkg/
│   ├── logger/          # Zap structured logging
│   └── response/        # Standardized API response helpers
├── tests/
│   ├── integration/     # Integration tests
│   ├── uat/             # User Acceptance Testing
│   └── unit/            # Unit tests
├── docker/              # Docker and Prometheus configurations
├── Dockerfile           # App Dockerfile
├── docker-compose.yml   # Multi-container orchestration
├── Makefile             # Standardized tasks
└── go.mod               # Go dependencies
```

## Component Details

### CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go`.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g. `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

### Schema Migration Engine
Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
- Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.

### GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

### Backup System
Managed by `internal/services/backup_service.go`.
- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.

### Observability Integration
Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.

## Testing

The system includes unit tests targeting Repository, Service, and API handlers.
- **Unit Tests**: `make test-unit`
- **Integration Tests**: `make test-integration`
- **UAT Tests**: `make test-uat`
- **Full Coverage**: `make test-coverage`

To manually run tests with coverage reporting across all packages:
`go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`

## Docker Setup

A `docker-compose.yml` is provided to spin up the entire stack, including:
- **cms-backend**: The main Go application.
- **postgres**: The metadata database.
- **redis**: Caching and potential message queue backend.
- **jaeger**: Tracing.
- **prometheus**: Metrics.

Run the stack using:
`docker-compose up -d`

## Swagger Documentation

API Documentation is generated using `swag`. To re-generate the documentation, run:
`~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`

(If `swag` is in your PATH, you can simply run `swag init ...`).
