# CMS Backend Platform

A production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. This system is fully self-hosted, scalable, cloud-ready, and acts as a backend infrastructure generator allowing users to connect databases, create data models dynamically, generate CRUD APIs and optional GraphQL endpoints, and more.

## System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

*   **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
*   **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
*   **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
*   **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

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

The core internal configuration is stored in the **Metadata Database**. Key tables include:

1.  `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2.  `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3.  `fields`: Defines attributes for each service (type: string, integer, float, uuid, JSON). Configures uniqueness, nullability, defaults.
4.  `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g., CanCreate, CanRead). Contains row-level and field-level capabilities.
5.  `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6.  `backups`: Stores snapshot records or schema outputs.
7.  `users` and `roles`: General authentication and authorization for the control plane.
8.  `audit_logs`: Detailed logging of structural and data-level modifications.
9.  `api_tokens`: API tokens for programmatic access.

## Go Project Structure

The project uses a clean modular structure:

```
.
├── ARCHITECTURE.md
├── Dockerfile
├── Makefile
├── cmd
│   └── server
│       └── main.go         # Application entrypoint
├── docker
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── prometheus.yml      # Observability configurations
├── docker-compose.yml      # Local dev environment
├── internal
│   ├── config            # Configuration structures
│   ├── database          # Database connection management
│   ├── graphql           # GraphQL gateway
│   ├── handlers          # REST API handlers (Presentation)
│   ├── middleware        # Auth, Rate limiting, Metrics
│   ├── models            # Core Domain Models
│   ├── services          # Business logic and Engine implementations
│   └── tracing           # OpenTelemetry setup
├── pkg
│   ├── logger            # Zap structured logging
│   └── response          # API response utilities
└── tests                 # Unit, Integration, and UAT testing
```

## Core Engine Implementations

### CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go` and `internal/handlers/dynamic_data_handler.go`.
*   Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
*   Applies automated filtering via query params (e.g. `?email=test@example.com`), dynamic joining via foreign key mappings, and pagination logic.
*   Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

### Schema Migration Engine
Managed by `internal/services/migration_service.go`.
*   Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM's `.Migrator()`.
*   Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.

### GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
*   Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
*   Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

### Backup System
Managed by `internal/services/backup_service.go`.
*   Generates data snapshots. Implements generic service record backup features mapping dynamic table contents to snapshots.
*   Supports restoring rows via JSON payload decoding.

## Observability Integration

The system implements state-of-the-art observability:
*   **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context. Request, error, and query logs.
*   **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
*   **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.

## Unit Tests

The system includes a comprehensive test suite (Unit, Integration, and UAT) aiming for >80% coverage.

Command to run tests:
```bash
make test
# or
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Docker Setup

The platform is containerized and ready for cloud deployment.
*   Use `docker-compose up -d` to launch the platform with necessary infrastructure (PostgreSQL, Redis, Jaeger, Prometheus).
*   The `Dockerfile` handles building the optimized Go binary (`CGO_ENABLED=1 GOOS=linux go build`).

## Swagger Documentation

Swagger/OpenAPI documentation is available and automatically generated from Go annotations.
Generate it with:
```bash
make swagger
# which runs: swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
Available at: `http://localhost:8080/swagger/index.html` (or your configured host/port).
