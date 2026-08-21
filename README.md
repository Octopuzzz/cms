# Backend Platform (Dynamic CMS + API Builder)

## Platform Goal
This is a self-hosted, Go-native Backend-as-a-Service (BaaS) platform. It allows developers to dynamically connect databases, create data models, automatically generate CRUD REST APIs, generate optional GraphQL APIs, manage schema migrations, monitor logs/performance, and manage backups.

This platform acts as a complete backend infrastructure generator, similar to Hasura or Supabase.

## System Architecture

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles.

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
1. **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
2. **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
3. **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
4. **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## Metadata Database Schema

The CMS stores platform metadata in a PostgreSQL (or SQLite/MySQL) database.
Core tables include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Component Implementation

### CMS Control Plane & Service Builder
Managed under `/api/v1/cms/` routes and handled by `internal/services/service_service.go` and `internal/services/dbconn_service.go`. These services:
- Provide an API interface to store service and field definitions in the metadata database.
- Utilize the `database_connections` engine to validate foreign database connectivity.
- Trigger migrations on service model updates to reflect schema changes.

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

## Project Structure

```text
.
├── cmd
│   └── server                # Main application entrypoint
├── internal
│   ├── config                # Configuration loading
│   ├── database              # DB connection manager and GORM setup
│   ├── graphql               # GraphQL gateway and schema generation
│   ├── handlers              # API HTTP handlers (Presentation Layer)
│   ├── middleware            # Auth, Logging, Metrics, Tracing
│   ├── models                # Domain models (Metadata DB schema)
│   ├── services              # Business logic (Service Builder, CRUD engine, etc.)
│   └── tracing               # OpenTelemetry integration
├── pkg
│   ├── logger                # Zap logger wrapper
│   └── response              # Standardized API responses
├── tests
│   ├── integration           # Integration tests
│   ├── uat                   # User Acceptance Tests
│   └── unit                  # Unit tests
├── docker                    # Docker configurations and Compose files
```

## Unit Testing

The system includes comprehensive unit testing.
Minimum coverage requirement is 80%.

To run tests with coverage:
```bash
make test-coverage
# or
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Setup & Deployment

1. **Docker Setup**: Utilize `docker-compose.yml` to spin up the application with its required backing services (Database, Prometheus, Jaeger, Redis, etc.).
2. **Swagger Documentation**: Accessible and generated using `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
