# Dynamic CMS & API Builder

A self-hosted, Go-native Backend-as-a-Service (BaaS) platform that dynamically generates REST and GraphQL APIs, managing schema migrations, scaling services, and providing extensive observability.

## System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. Tables include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks. Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

```text
.
├── cmd
│   └── server          # Entry point for the application
├── docker              # Docker and containerization setup
├── internal
│   ├── config          # Configuration loading
│   ├── database        # DB connection management
│   ├── graphql         # GraphQL gateway and resolvers
│   ├── handlers        # HTTP/REST handlers (Presentation Layer)
│   ├── middleware      # Gin middleware (Auth, Metrics, etc.)
│   ├── models          # Domain models (Domain Layer)
│   ├── services        # Business logic (Application Layer)
│   └── tracing         # OpenTelemetry setup
├── pkg
│   ├── logger          # Zap structured logging
│   └── response        # HTTP response standardization
└── tests               # Unit, integration, and UAT tests
```

## CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g. `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## Schema Migration Engine

Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
- Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.

## GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

## Backup System

Managed by `internal/services/backup_service.go`.
- Generates data snapshots mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.

## Observability Integration

Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.

## Unit Tests

Run unit tests using the standard Go testing tool:
```bash
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Docker Setup

The platform includes a `Dockerfile` and `docker-compose.yml` for easy deployment. Run the platform via:
```bash
docker-compose up -d
```

## Swagger Documentation

Swagger API documentation can be generated using:
```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
