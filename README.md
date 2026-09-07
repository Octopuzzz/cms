# Dynamic CMS & API Builder

A production-grade Backend-as-a-Service (BaaS) platform written in Go. This platform acts as a backend infrastructure generator, allowing users to dynamically create backend services, manage schemas, automatically generate REST and GraphQL APIs, handle migrations, and monitor the system.

## System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

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

The metadata database stores the platform's core configuration. It is managed by GORM. Key tables include:

- `database_connections`: External DB configurations (id, name, type, host, port, credentials).
- `services`: Represents user-created data models, tracks schemas (`db_table_name`), and references `database_connection_id`.
- `fields`: Attributes for each service (type, uniqueness, nullability, defaults).
- `service_permissions`: Connects `Role` to `Service` for fine-grained access control (CanCreate, CanRead).
- `migrations`: DDL execution history and metadata for rollbacks.
- `backups`: Snapshot records or schema outputs.
- `users` and `roles`: Platform authentication and authorization.
- `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure:

```
├── cmd/
│   └── server/               # Main application entrypoint
├── internal/
│   ├── api/                  # API routes and initialization
│   ├── config/               # Application configuration
│   ├── database/             # Database connection manager
│   ├── graphql/              # GraphQL gateway and schemas
│   ├── handlers/             # HTTP request handlers (Presentation Layer)
│   ├── middleware/           # Gin middleware (Auth, Prometheus, Tracing)
│   ├── models/               # Core data models (Domain Layer)
│   ├── services/             # Business logic (Application Layer)
│   └── tracing/              # OpenTelemetry integration
├── pkg/
│   ├── logger/               # Zap logging wrappers
│   ├── pagination/           # Pagination utilities
│   └── response/             # Standardized HTTP responses
├── tests/                    # Unit, Integration, and UAT tests
├── docker/                   # Docker configurations
├── Dockerfile                # Production Docker image build instructions
└── docker-compose.yml        # Multi-container local orchestration
```

## CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`, the dynamic CRUD engine handles the standard data operations.
- Automatically maps standard endpoints (e.g., `GET /api/v1/data/{slug}`) to GORM database queries based on metadata configurations.
- Handles filtering via query params (e.g., `?email=test@example.com`), sorting, dynamic joins via foreign keys, and pagination.
- Role-Based Access Control and Row-Level policies are automatically applied before execution.

## Schema Migration Engine

Managed by `internal/services/migration_service.go`, ensuring safe schema evolution.
- Applies safe DDL diffs (Add, Drop, Rename Column) using GORM’s Migrator.
- Retains migration history in the `migrations` table natively, supporting history tracking and snapshot retention for rollbacks.
- Supports automatic backups before schema modifications are executed.

## GraphQL Gateway

Managed by `internal/graphql/gateway.go`, utilizing `gqlgen`.
- Automatically generates an optional generic GraphQL schema for dynamically defined data models.
- Exposed under `POST /api/v1/graphql` to satisfy queries, mutations, and dynamic relationship resolutions without code changes.

## Backup System

Managed by `internal/services/backup_service.go`.
- Captures state through data snapshots.
- Implements generic record backups that dump dynamic table state into retrievable JSON snapshots.
- Supports data restoration operations by decoding these structured snapshots back into the target database.

## Observability Integration

Full-stack observability is embedded into the platform:
- **Logging**: Zap structured logging integrated in `pkg/logger/`, tied with correlation IDs in the request context.
- **Metrics**: Prometheus HTTP interceptors (`internal/middleware/prometheus.go`) record request latency, status codes, and error rates. Exposed on `/metrics`.
- **Tracing**: OpenTelemetry (integrated via `internal/tracing/`) traces API handlers, wrapping SQL commands and network logic.

## Unit Tests

The system maintains comprehensive unit testing aiming for at least 80% coverage across Repository, Service, and Handler layers.
- Tests utilize an in-memory SQLite backend to mock the database without external dependencies.
- Standard test command: `go test ./...`
- Coverage execution: `make test-coverage`

## Docker Setup

The repository is containerized for seamless scaling and deployment.
- `Dockerfile` provides an optimized, multi-stage build generating a minimal executable.
- `docker-compose.yml` orchestrates the backend along with necessary infrastructure like PostgreSQL, Redis, and observability platforms (Prometheus/Jaeger) for local environments.

## Swagger Documentation

Auto-generated API documentation is supported via `swaggo/swag`.
- Generation is facilitated by running `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
- Fully integrated to provide interactive standard REST documentation conforming to OpenAPI specifications.
