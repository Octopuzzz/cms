# CMS Backend Platform

This repository contains the implementation of a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. The platform acts as a backend infrastructure generator, allowing developers to connect databases, create data models dynamically, generate CRUD APIs, configure optionally GraphQL APIs, manage schemas, and monitor logs, metrics, and performance.

The platform follows Clean Architecture and Domain Driven Design, resulting in a system that is modular, scalable, and cloud-ready.

## Goal

Build a Backend-as-a-Service platform allowing developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

## System Architecture Explanation

The backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
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

The core internal configuration is stored in the **Metadata Database**. Tables defined in `internal/models/models.go` include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `relations`: Contains relational definitions mapping `One-to-One`, `One-to-Many`, `Many-to-One` inside `RelationConfig`.
5. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
6. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
7. `backups`: Stores snapshot records or schema outputs.
8. `users` and `roles`: General authentication and authorization for the control plane.
9. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure.

```
cmd/
  server/                 # Application entry point
internal/
  api/                    # REST / Routing definitions
  config/                 # Configuration and Env loading
  database/               # Connection pooling & logic
  graphql/                # GraphQL gateway & schemas
  handlers/               # API Controllers
  middleware/             # HTTP Interceptors (Auth, Logging)
  models/                 # Domain objects
  services/               # Business logic (CRUD, Migrations, CMS)
  tracing/                # OpenTelemetry
pkg/
  logger/                 # Structured logging
  response/               # Standardized JSON formats
tests/
  integration/            # Integration Tests
  uat/                    # User Acceptance Tests
  unit/                   # Unit Tests
docker/
  Dockerfile              # Container build definition
  docker-compose.yml      # Local dev environment with PostgreSQL / Redis
```

## CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go` and `internal/handlers/dynamic_data_handler.go`.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g. `?search=term`), dynamic joining via foreign key mappings (`?joins=`), and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

Operations available:
- **Create**: `POST /api/v1/data/{slug}`
- **Read**: `GET /api/v1/data/{slug}/{id}`
- **Update**: `PUT /api/v1/data/{slug}/{id}`
- **Delete**: `DELETE /api/v1/data/{slug}/{id}`
- **List**: `GET /api/v1/data/{slug}`

## Schema Migration Engine

Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using custom dialector specific DDL construction.
- Safety measures include validate SQL identifiers to prevent injection and pre-checking allowed column types.
- Logs migration status natively into `migrations` table and handles rollbacks through status updating and tracking.

## GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.
- Provides a GraphQL Playground at `GET /api/v1/graphql/playground`.

## Backup System

Managed by `internal/services/backup_service.go`.
- Generates data snapshots. It implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports generic `Backup` definitions to save table schema alongside rows `Type: "snapshot"`.
- Supports restoring rows via JSON payload decoding.

## Observability Integration

Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic. Configurable via `.env` metrics.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency, request count, and status codes via standard HTTP interceptors. Exported at `GET /metrics`.

## Unit Tests

The system includes comprehensive tests covering Handlers, Services, and Utilities.
The project targets an 80% minimum coverage.

Commands to run tests:
- All tests: `make test`
- Unit tests: `make test-unit`
- Integration tests: `make test-integration`
- Coverage report: `make test-coverage`

To manually run tests:
```bash
go test ./...
```

## Docker Setup

The application contains full Docker integration.
- `Dockerfile`: Multi-stage build process generating a minimal container runtime based on Alpine.
- `docker-compose.yml`: Spins up the CMS Backend with associated metrics platforms.

Commands:
- `make docker-up`: Start Docker services
- `make docker-down`: Stop Docker services
- `make docker-build`: Build Docker images

## Swagger Documentation

Swagger API documentation defines the available endpoints using annotations inside `cmd/server/main.go` and the `handlers` package.

To generate the documentation:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

The Swagger UI is exposed on the running server at `/api/v1/swagger/index.html`.
