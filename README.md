# Go Dynamic CMS & API Builder

A production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. This system acts as a fully self-hosted and Go-native backend infrastructure generator, similar to Hasura or Supabase.

It allows users to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST APIs, generate optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

## Platform Features

- **Connect Databases:** PostgreSQL, MySQL, MongoDB, SQLite.
- **Dynamic Data Models:** Dynamically create and manage services/schemas.
- **Auto-generated APIs:** Automatically generate REST CRUD endpoints.
- **GraphQL APIs:** Automatically generate GraphQL queries and mutations.
- **Schema Migrations:** Manage and track DDL schema changes.
- **Observability:** Built-in logging (Zap), metrics (Prometheus), and tracing (OpenTelemetry).
- **Backups:** Generic service snapshot backup and restoration capabilities.

## System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

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

## Metadata Database Schema

The platform metadata is stored internally (usually in PostgreSQL or SQLite for local dev). Tables include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service (string, integer, float, uuid, JSON). Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Project Structure

```text
.
├── cmd
│   └── server          # Application entrypoint
├── internal
│   ├── config          # Application configuration
│   ├── database        # Database connection management
│   ├── graphql         # GraphQL Gateway implementation
│   ├── handlers        # Presentation Layer: Gin HTTP handlers
│   ├── middleware      # Auth, tracing, metrics middlewares
│   ├── models          # Domain Layer: Data models and entities
│   ├── services        # Application Layer: Business logic (CMS, CRUD, Migration, Backup)
│   └── tracing         # Infrastructure Layer: OpenTelemetry tracing
├── pkg
│   ├── logger          # Structured Zap logging
│   ├── pagination      # Pagination utilities
│   ├── response        # Standardized API responses
│   └── validation      # Input validation utilities
├── tests               # Unit and Integration tests
├── docker              # Docker configurations
├── .env.example        # Example environment variables
├── Dockerfile          # Multi-stage Docker build
├── Makefile            # Standardized task runner
└── ARCHITECTURE.md     # Primary source of truth for platform architecture
```

## Core Components Implementation

### 1. CMS Control Plane & Service Builder
Managed under `/api/v1/cms/` routes and handled by `internal/services/service_service.go` and `internal/services/dbconn_service.go`.
- Provides an API interface to store service and field definitions in the metadata database.
- Utilizes the `database_connections` engine to validate foreign database connectivity.
- Triggers migrations on service model updates to reflect schema changes.

### 2. CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go`.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g. `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

### 3. Schema Migration Engine
Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
- Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.

### 4. GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

### 5. Backup System
Managed by `internal/services/backup_service.go`.
- Generates data snapshots.
- Implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.

### 6. Observability Integration
Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.

## Unit Testing

The system includes comprehensive unit tests with a minimum coverage of 80% across Repository, Service, and API handler layers.

To run the full test suite and calculate coverage:

```bash
make test-coverage
# or manually
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

Unit tests use an in-memory SQLite database setup (`:memory:`) via `github.com/glebarez/sqlite` driver for fast execution and database mocking.

## Swagger Documentation

Swagger / OpenAPI documentation is generated using `swaggo`.

To generate docs:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

*(Note: The generated `docs/` folder is gitignored.)*

## Docker Setup

The platform includes a `Dockerfile` for building a containerized Go binary, and a `docker-compose.yml` file to quickly spin up the backend platform alongside necessary infrastructure (e.g., PostgreSQL for metadata, Redis for caching, Jaeger for tracing, Prometheus for metrics).

```bash
docker-compose up -d
```
