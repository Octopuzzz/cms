# Dynamic CMS + API Builder Backend Platform

## Overview
This platform is a **Backend-as-a-Service (BaaS)** that allows developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability. It works similarly to platforms like Hasura or Supabase but is **fully self-hosted and Go-native**.

The system is modular, scalable, cloud-ready, and follows Clean Architecture and Domain-Driven Design (DDD).

---

## 1. System Architecture Explanation

The architecture is composed of core layers handling distinct responsibilities:
- **Presentation Layer**: The Gin HTTP Router and Handlers (e.g., `internal/handlers/`). Acts as the API gateway mapping requests to internal services and GraphQL endpoints via `gqlgen`.
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

---

## 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**, leveraging standard RDBMS setups (e.g. PostgreSQL, SQLite).

Key tables include:
1. `database_connections`: Stores external DB configurations (`id`, `name`, `type`, `host`, `port`, credentials).
2. `services`: Represents user-created data models, maps to a `database_connection_id`, and tracks the dynamic `db_table_name`.
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, json. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g., CanCreate, CanRead).
5. `migrations`: Tracks DDL executions with metadata for rollbacks/history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` & `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

---

## 3. Go Project Structure

The project uses a clean, modular structure:

```
.
├── cmd/
│   └── server/          # Entry point for the Go application
├── internal/
│   ├── api/             # API routing configurations
│   ├── config/          # Environment and application configurations
│   ├── database/        # Database connection manager and pooling
│   ├── graphql/         # GraphQL gateway setup and schemas
│   ├── handlers/        # HTTP handlers for REST API (Presentation Layer)
│   ├── middleware/      # Auth, Rate Limiter, Prometheus middleware
│   ├── models/          # Core Domain models
│   ├── services/        # Business logic (Application Layer)
│   └── tracing/         # OpenTelemetry tracing setup
├── pkg/
│   ├── logger/          # Zap structured logging wrapper
│   ├── response/        # Standardized API responses
│   └── ...              # Other shared packages (pagination, validation)
├── tests/               # Unit and integration tests
├── docker/              # Docker configuration files
├── docs/                # Generated Swagger/OpenAPI documentation (untracked)
├── Makefile             # Standardized build/test commands
├── Dockerfile           # App containerization
└── docker-compose.yml   # Multi-container orchestration
```

---

## 4. CRUD Engine & Query Engine Implementation

Managed dynamically via `internal/services/dynamic_data_service.go` and `internal/handlers/dynamic_data_handler.go`.
- Auto-generates CRUD endpoints for each created service (`POST /api/v1/data/{slug}`, `GET /api/v1/data/{slug}/{id}`, etc.).
- The Query Engine parses query parameters to apply filtering (`?email=test@test.com`), sorting (`?sort=created_at:desc`), and pagination (`?page=1&limit=20`).
- Supports automatic join queries based on defined relationships and validates order-by constraints.

---

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema changes using GORM's `Migrator()`.
- Supports adding, dropping, and renaming columns safely.
- Logs migration statuses in the `migrations` metadata table.
- Supports rollbacks through snapshot retention capabilities.

---

## 6. GraphQL Gateway

Located in `internal/graphql/gateway.go`.
- Implemented using `gqlgen`.
- Dynamically exposes schemas based on the configurations stored in the metadata database.
- Allows clients to run complex nested queries, mutations, and relationship retrievals over a single `/api/v1/graphql` endpoint.

---

## 7. Backup System

Managed by `internal/services/backup_service.go`.
- Supports generating table and schema backups as data snapshots (e.g. JSON dumps).
- Implements endpoints like `POST /api/v1/cms/backup/service/{service_id}`.
- Supports restoring records via JSON payload decoding into the dynamic services.

---

## 8. Observability Integration

Integrated deeply across the stack:
- **Logging**: Zap structured logging is injected globally with correlation and trace IDs tied to context requests.
- **Metrics**: Prometheus tracks request latency, status codes, and error rates (exported at `/metrics`) via standard HTTP interceptors (`internal/middleware/prometheus.go`).
- **Tracing**: OpenTelemetry (integrated with Jaeger) wraps SQL commands and network logic in `internal/tracing/`.

---

## 9. Unit Tests

The system enforces a minimum of 80% test coverage across Repository, Service, and API handler layers.
- Uses an in-memory SQLite database setup (`:memory:`) via `github.com/glebarez/sqlite` to mock DB connections during testing.
- Test suites can be run via:
  ```bash
  make test-unit
  make test-coverage  # Full coverage profile creation
  go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
  ```
- Includes UATs testing superadmin promotion and complex access flows.

---

## 10. Docker Setup

Containerized for cloud readiness:
- The standard `Dockerfile` handles building the Go binary and running it.
- `docker-compose.yml` orchestrates the backend platform along with necessary infrastructure containers like PostgreSQL (Metadata DB), Redis (Caching), Prometheus (Metrics), and Jaeger (Tracing).

---

## 11. Swagger Documentation

REST endpoints are self-documented via Swagger/OpenAPI.
- Generated via the `swag` CLI.
- Run the generator with:
  ```bash
  ~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
  ```
- The documentation is accessible at the Swagger UI endpoint when running the server, enabling developers to interactive test the API Builder endpoints.
