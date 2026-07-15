# Dynamic CMS & API Builder

A production-grade Backend Platform written in Go. This system acts as a backend infrastructure generator allowing users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

It is modular, scalable, cloud-ready, and follows Clean Architecture and Domain-Driven Design (DDD).

---

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles.

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

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

---

## 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. Tables include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks. Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

---

## 3. Go Project Structure

The project uses a clean modular structure:

```
cmd/
  server/         # Main application entry point
internal/
  config/         # Configuration management
  database/       # Database connection manager & pooling
  graphql/        # GraphQL gateway and schemas
  handlers/       # Gin HTTP handlers (Presentation Layer)
  middleware/     # Auth, Prometheus, Rate Limiter, OpenTelemetry
  models/         # Domain models
  services/       # Business logic & Core Engines (CRUD, Migration, Service Builder)
  tracing/        # OpenTelemetry tracing setup
pkg/
  logger/         # Zap structured logging
  response/       # Standardized API responses
tests/
  integration/    # Integration tests
  uat/            # User Acceptance tests
  unit/           # Unit tests
docker/           # Docker setup files
```

---

## 4. CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`.
- Automatically generates CRUD endpoints for dynamically created services.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params, dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

---

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema changes (Add column, Drop column, Rename column, Change column type) using GORM’s `.Migrator()`.
- Logs migration status natively into the `migrations` table.
- Supports rollback capabilities through snapshot retention logic and automatic backups before migration.

---

## 6. GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.
- Supports queries, mutations, and relations.

---

## 7. Backup System

Managed by `internal/services/backup_service.go`.
- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.
- Provides table backup, schema backup, and service snapshot capabilities.

---

## 8. Observability Integration

Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context. Request, error, and query logs.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency, request rate, and status codes via standard HTTP interceptors. Exported at `/metrics`.

---

## 9. Unit Tests

The system includes comprehensive unit tests targeting a minimum of 80% coverage.
- **Test areas**: Repository, Service, and API handlers.
- **Mocking**: In-memory SQLite database setup via `github.com/glebarez/sqlite`.
- **Command**: Run tests using `make test-unit` or `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`.

---

## 10. Docker Setup

The platform is fully containerized.
- Contains `Dockerfile` for the Go backend.
- `docker-compose.yml` sets up the backend along with external databases (PostgreSQL, MySQL, MongoDB), Redis for caching, and observability tools (Prometheus, Jaeger).

---

## 11. Swagger Documentation

API Documentation is generated using `swaggo/swag`.
- Provides OpenAPI/Swagger specs for the CMS Control Plane and dynamically generated REST endpoints.
- Generate docs with: `~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
