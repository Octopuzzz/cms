# Dynamic CMS & API Builder - Go Backend Platform

A production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. This platform acts as a backend infrastructure generator, similar to Hasura or Supabase, but is fully self-hosted and Go-native. It allows users to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles to ensure modularity, scalability, and maintainability.

### Core Layers:
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

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

## 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database** (typically PostgreSQL or SQLite). Tables include:

- `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, username, password, database_name. Includes connection pooling settings.
- `services`: Represents user-created data models. Fields include `id`, `name`, `database_id`, `db_table_name`, `created_at`, `updated_at`.
- `fields`: Defines attributes for each service (name, type, nullable, unique, default_value, index). Supported types: string, integer, float, boolean, uuid, json, array, timestamp.
- `relations`: Maps relationships between services (One-to-One, One-to-Many, Many-to-One, Many-to-Many).
- `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead).
- `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- `backups`: Stores snapshot records or schema outputs.
- `users` and `roles`: General authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of structural and data-level modifications.

## 3. Go Project Structure

The project uses a clean modular structure, ensuring separation of concerns:

```
.
├── cmd/
│   └── server/             # Application entrypoint (main.go)
├── internal/
│   ├── api/                # API routing
│   ├── config/             # Environment and config loading
│   ├── database/           # DB connection manager and pool caching
│   ├── graphql/            # gqlgen integration and optional GraphQL Gateway
│   ├── handlers/           # Gin HTTP handlers mapping to services
│   ├── middleware/         # Auth, rate limiting, Prometheus, tracing
│   ├── models/             # Domain layer models for the Metadata DB
│   ├── services/           # Application business logic (CRUD Engine, CMS, etc.)
│   └── tracing/            # OpenTelemetry initialization
├── pkg/
│   ├── errors/             # Standardized error handling
│   ├── logger/             # Structured Zap logger wrapper
│   ├── pagination/         # Common pagination utilities
│   ├── response/           # Standardized Gin HTTP response helpers
│   └── validation/         # Validation logic
├── tests/                  # Test suites
│   ├── integration/
│   ├── uat/
│   └── unit/
├── docker/                 # Container configs (e.g. Redis, Jaeger)
├── Dockerfile              # Docker build instructions
├── docker-compose.yml      # Local dev services
├── go.mod                  # Go module dependencies
└── Makefile                # Task definitions (build, test, dev, etc.)
```

## 4. CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`, the **Dynamic CRUD Engine** and **Query Engine** automatically generate operations when a service is created.
- Maps dynamic endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g. `?email=test@test.com`), dynamic joining via foreign key mappings, sorting (`?sort=created_at:desc`), and pagination (`?page=1&limit=20`).
- Validates field constraints and relation definitions at runtime.

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
- Logs migration status natively into the `migrations` table and handles rollbacks through snapshot retention logic.
- Supports automatic backups before migrations to ensure data safety.

## 6. GraphQL Gateway (Optional)

Managed by `internal/graphql/gateway.go`.
- Dynamically generates GraphQL schemas using `github.com/99designs/gqlgen`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL queries, mutations, and relations for defined data models.

## 7. Backup System

Managed by `internal/services/backup_service.go`.
- Implements table backup, schema backup, and service snapshots.
- Supports backup formats like JSON snapshot and SQL dump.
- Exposed via APIs like `POST /cms/backup/service/{service_id}` and `POST /cms/restore/service/{service_id}`.

## 8. Observability Integration

Implemented comprehensively across the platform:
- **Logging (Zap)**: Structured logs injected globally (`pkg/logger/`) with request, error, and query logging.
- **Tracing (OpenTelemetry)**: Initialized in `internal/tracing/` to track request latency and wrap SQL commands for distributed tracing (viewable in Jaeger).
- **Metrics (Prometheus)**: Configured in `internal/middleware/prometheus.go` to expose request counts, latency, and error rates at the `/metrics` endpoint.

## 9. Unit Tests

The system maintains a minimum 80% code coverage focusing on Repositories, Services, and API Handlers.
- Unit tests use an in-memory SQLite database (`:memory:`) via `github.com/glebarez/sqlite`.
- **Running tests:**
  ```bash
  go test ./...
  ```
  Or for coverage:
  ```bash
  make test-coverage
  ```

## 10. Docker Setup

The platform is fully containerized.
- **Dockerfile**: Builds the Go application as a lightweight container.
- **docker-compose.yml**: Brings up the backend API along with necessary infrastructure like PostgreSQL, Redis (for caching), Prometheus, and Jaeger (OpenTelemetry).

## 11. Swagger Documentation

The project includes API documentation mapped via Swagger/OpenAPI.
- Generated using the standard `swag` CLI.
- Standard generation command:
  ```bash
  ~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
  ```
- Accessible typically at `/swagger/index.html`.
