# Backend Platform (Dynamic CMS + API Builder)

This project is a **production-grade Backend-as-a-Service (BaaS) platform** written in Go. It operates similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native. The platform enables developers to dynamically create backend services, schemas, and APIs (both REST and optional GraphQL), and manages schema migrations, database connections, observability, and backups.

## System Architecture Explanation

The CMS Backend is built on **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`), serving as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen` (`internal/graphql/`).
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (`internal/middleware/prometheus.go`), and logging (`pkg/logger/`).

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

The platform metadata is stored in a relational database (e.g., PostgreSQL). Key tables include:

1. `database_connections`: Stores external DB configurations (id, name, type, host, port, credentials).
2. `services`: Represents user-created data models, linking to `database_connections` and tracking dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service (type, uniqueness, nullability, defaults).
4. `service_permissions`: Connects `Role` to `Service` for fine-grained access checks.
5. `migrations`: Tracks DDL executions for rollbacks and history.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` & `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure:

```
├── cmd/
│   └── server/             # Main entry point for the application
├── docker/                 # Docker configuration (Prometheus, etc.)
├── internal/
│   ├── config/             # Configuration management
│   ├── database/           # Database connection manager
│   ├── graphql/            # GraphQL gateway implementation
│   ├── handlers/           # Gin HTTP handlers
│   ├── middleware/         # Auth, rate limiting, metrics middlewares
│   ├── models/             # Domain data models
│   ├── services/           # Application business logic
│   └── tracing/            # OpenTelemetry integration
├── pkg/
│   ├── logger/             # Zap structured logging
│   └── response/           # Standardized HTTP responses
├── tests/                  # Unit, Integration, and UAT tests
├── docker-compose.yml      # Multi-container setup
├── Dockerfile              # Containerization instructions
├── Makefile                # Build, test, run commands
└── README.md
```

## CMS Control Plane & Service Builder
Managed under `/api/v1/cms/` routes and handled by `service_service.go` and `dbconn_service.go`. These services provide an API interface to store service and field definitions, validate external database connectivity, and trigger migrations on service model updates to reflect schema changes.

## CRUD Engine & Query Engine

Managed by `internal/services/dynamic_data_service.go`, this engine dynamically handles standard GORM database operations based on defined services.
- Auto-generates standard CRUD operations: Create, Read, Update, Delete, List.
- Applies automated filtering via query params (e.g., `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination.
- Enforces Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## Schema Migration Engine

Managed by `internal/services/migration_service.go`, the migration engine supports safe schema changes dynamically.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM's `.Migrator()`.
- Logs migration status natively into the `migrations` table and handles rollbacks through snapshot retention logic.

## GraphQL Gateway

Managed by `internal/graphql/gateway.go`, the GraphQL integration dynamically generates generic optional schemas based on the defined service schemas.
- Uses `github.com/99designs/gqlgen` for execution.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for data models, supporting queries, mutations, and relations.

## Backup System

Managed by `internal/services/backup_service.go`, the system provides reliable snapshot mechanisms.
- Generates data snapshots mapping dynamic table contents to JSON or SQL dumps.
- Supports restoring rows via JSON payload decoding.

## Observability Integration

The platform includes state-of-the-art observability:
- **Logging**: Zap structured logging (`pkg/logger/`) with correlation and trace IDs injected into context.
- **Metrics**: Prometheus metrics exported at `/metrics`, tracking request latency, slow queries, and error rates via `internal/middleware/prometheus.go`.
- **Tracing**: OpenTelemetry/Jaeger (`internal/tracing/`) initialization to wrap SQL commands and network logic.

## Unit Tests

The system requires a minimum of 80% unit test coverage across Repository, Service, and API handler layers.
- Tests use an in-memory SQLite database setup (`:memory:`) via `github.com/glebarez/sqlite`.
- Commands:
  - `make test-unit`: Run unit tests
  - `make test-coverage`: Run the full test suite with coverage reporting. To ensure accuracy, tests use `-coverpkg=./...`.

## Docker Setup

The system is fully containerized using Docker and Docker Compose.
- **Dockerfile**: Multi-stage build process compiling the Go application statically for minimal footprint.
- **docker-compose.yml**: Orchestrates the entire stack, including the Go Backend, PostgreSQL (metadata), Redis (caching), Jaeger (tracing), and Prometheus (metrics).

Commands:
- `make docker-up` to start the cluster.
- `make docker-down` to stop it.

## Swagger Documentation

The project utilizes Swagger/OpenAPI for API documentation.
- Built via `swaggo/swag`.
- Generated using the `make swagger` command, which compiles the annotations in the source code to the `docs/` folder (ignored in version control).
- Provides a UI interface natively accessible via the running application to easily explore dynamic APIs.
