# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. It works similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native. The platform allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

## 1. System Architecture Explanation

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database** (managed by GORM). Key tables include:

- `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
- `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
- `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
- `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g., CanCreate, CanRead). Contains row-level and field-level capabilities.
- `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- `backups`: Stores snapshot records or schema outputs.
- `users` and `roles`: General authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of structural and data-level modifications.

## 3. Go Project Structure

The project uses a clean, modular structure:

```
├── cmd
│   └── server          # Main entrypoint
├── internal
│   ├── config          # Application configuration
│   ├── database        # Database connection manager
│   ├── graphql         # GraphQL gateway
│   ├── handlers        # API endpoints (Presentation Layer)
│   ├── middleware      # Auth, rate limiting, metrics
│   ├── models          # Domain entities
│   ├── services        # Business logic (CMS, CRUD, Query, etc.)
│   └── tracing         # OpenTelemetry setup
├── pkg
│   ├── logger          # Zap structured logging
│   └── response        # Standardized API responses
├── tests
│   ├── integration     # Integration tests
│   ├── uat             # User Acceptance Testing
│   └── unit            # Unit tests
└── docker              # Docker and Docker Compose files
```

## 4. CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go` and `internal/handlers/dynamic_data_handler.go`.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations dynamically.
- Automatically generates operations: Create, Read, Update, Delete, List.
- Applies automated filtering via query params (e.g. `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go` and `internal/handlers/migration_handler.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM's `.Migrator()`.
- Logs migration status natively into `migrations` table.
- Supports rollback through snapshot retention logic.

## 6. GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.
- Generates generic optional schemas dynamically using `github.com/99designs/gqlgen/graphql`.

## 7. Backup System

Managed by `internal/services/backup_service.go` and `internal/handlers/backup_handler.go`.
- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.
- APIs exist for `POST /api/v1/cms/backups/{service_id}` etc.

## 8. Observability Integration

Implemented across the stack:
- **Logging**: Zap structured logging injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **Metrics**: Prometheus tracking request latency, status codes, and error rates via standard HTTP interceptors (`internal/middleware/prometheus.go`). Exported at `/metrics`.
- **Tracing**: OpenTelemetry/Jaeger initialized in `internal/tracing/` to wrap SQL commands and network logic.

## 9. Unit Tests

The system includes comprehensive tests:
- High minimum coverage targeted for Repository, Service, and API handlers.
- Standard test command: `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...` (also wrapped in `make test-coverage`).
- Tests utilize an in-memory SQLite database (`:memory:`) for quick validation of the domain models and services.

## 10. Docker Setup

Containerization is supported natively:
- `Dockerfile` (or `docker/Dockerfile`) wraps the Go compilation and execution.
- `docker-compose.yml` (and `docker/docker-compose.yml`) configures the main service alongside required infrastructure components (e.g. Prometheus).

## 11. Swagger Documentation

API Documentation via Swagger/OpenAPI:
- Built via swag (`github.com/swaggo/swag`).
- Generates docs using `make docs` or directly with `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
- Can be viewed when served on configured swagger paths.
