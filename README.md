# Dynamic CMS & API Builder

A production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. This system operates similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native. The platform allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

The system is modular, scalable, cloud-ready, and follows Clean Architecture and Domain-Driven Design (DDD).

## System Architecture

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer**: The Gin HTTP Router and Handlers. Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (Services). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models. Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry, metrics (Prometheus), and logging.

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

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. Tables include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean, modular structure:

```text
├── cmd
│   └── server          # Entry point for the application
├── docker              # Docker configurations and scripts
├── internal
│   ├── api             # API routes and configuration
│   ├── config          # Application configuration
│   ├── database        # Database connection manager
│   ├── graphql         # GraphQL gateway and schemas
│   ├── handlers        # HTTP handlers (REST)
│   ├── middleware      # Gin middleware (Auth, Rate Limiting, Prometheus)
│   ├── models          # Domain models (Metadata DB schema)
│   ├── services        # Core business logic (CMS, CRUD, Migration, Backup)
│   └── tracing         # OpenTelemetry tracing
├── pkg
│   ├── logger          # Zap structured logging
│   └── response        # Standardized API responses
├── tests
│   ├── integration     # Integration tests
│   ├── uat             # User Acceptance Testing
│   └── unit            # Unit tests
├── Dockerfile
├── Makefile
├── docker-compose.yml
└── go.mod
```

## Core Components Implementation

### CRUD Engine Implementation

Managed by the `DynamicDataService` (`internal/services/dynamic_data_service.go`).
- Automatically generates CRUD endpoints for dynamically created services.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g., `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to executing queries.

### Schema Migration Engine

Managed by `MigrationService` (`internal/services/migration_service.go`).
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
- Logs migration status natively into the `migrations` table and handles rollbacks through snapshot retention logic.
- Automatically triggers migrations on service model updates to reflect schema changes.

### GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.
- Supports querying dynamic relations and mutations based on configured service models.

### Backup System

Managed by `BackupService` (`internal/services/backup_service.go`).
- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.
- Can create comprehensive system backups and schema backups to safeguard data models.

## Observability Integration

Implemented across the stack to ensure system health and monitor performance:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context. Captures request logs, error logs, and query logs.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic. Provides deep tracing for request latency and system interactions.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks request latency, error rates, and status codes via standard HTTP interceptors. Exported at the `/metrics` endpoint.

## Security and Performance

- **Security**: Implements JWT authentication, rate limiting middleware, SQL injection protection through parameterized queries (GORM), and rigorous input validation.
- **Performance**: Includes database connection pooling, pagination optimization, and efficient context handling.

## Unit Tests

The system includes comprehensive test coverage for Repository, Service, and API handlers.
- **Requirement**: Minimum of 80% unit test coverage.
- **Run Tests**: Use `make test-coverage` to execute the full test suite. To accurately capture coverage across all packages (like `internal/...`) when running tests isolated in `tests/...`, the system uses `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`.

## Docker Setup

The platform is containerized for easy deployment.
- **Build Image**: `docker build -t cms-backend .`
- **Run with Compose**: Utilize `docker-compose.yml` to spin up the CMS Backend alongside its dependencies (PostgreSQL, Redis, Prometheus, Jaeger).
- `docker-compose up -d`

## Swagger Documentation

API documentation is generated using `swaggo/swag`.
- **Generate Docs**: Run `~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs` (or use the system PATH equivalent).
- **Access Docs**: Once running, the Swagger UI is available at `/swagger/index.html`. (Note: The `docs/` directory is ignored in version control).
