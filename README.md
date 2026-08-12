# Go Dynamic CMS + API Builder Platform

A production-grade, self-hosted Backend-as-a-Service (BaaS) platform written natively in Go. This system dynamically generates REST and GraphQL APIs, manages database connections, defines schemas, and provides automated CRUD endpoints on the fly. It follows Clean Architecture and Domain-Driven Design (DDD).

## 1. System Architecture Explanation

The platform architecture acts as a backend infrastructure generator, sitting between client applications and backend databases:

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

**Core Layers:**
- **Presentation Layer (`internal/handlers/`):** Gin HTTP Router acting as the API Gateway. Includes REST endpoints and GraphQL endpoints using `gqlgen`.
- **Application Layer (`internal/services/`):** Business logic layer managing data access, dynamic schemas, and orchestrating operations.
- **Domain Layer (`internal/models/`):** Core data entities such as `Service`, `Field`, `DatabaseConnection`, `User`, and `Role`.
- **Infrastructure Layer:** Handles database connections, caching (Redis), logging (`pkg/logger/` using Zap), metrics (Prometheus), and tracing (`internal/tracing/` using OpenTelemetry).

## 2. Metadata Database Schema

The CMS manages its own state using a Metadata Database (typically PostgreSQL, but SQLite is supported for testing). The primary tables include:

1. **`database_connections`**: Stores external database configurations (id, name, type, host, port, credentials).
2. **`services`**: Represents user-created data models, linking to `database_connections` and holding the physical `db_table_name`.
3. **`fields`**: Configures attributes for each service, defining the type (string, integer, float, boolean, uuid, json), nullability, defaults, and uniqueness.
4. **`relations`** / **`service_permissions`**: Connects entities for relation engines (One-to-One, One-to-Many, etc.) and binds `Role` access to `Service`.
5. **`migrations`**: Tracks DDL executions (add/drop/rename columns) for rollback and history.
6. **`backups`**: Records generated snapshots and their paths/formats.
7. **`users`** & **`roles`**: Authentication, authorization, and RBAC implementation.
8. **`audit_logs`**: Logs for structural and data-level modifications.

## 3. Go Project Structure

The repository follows a clean, modular layout standard to robust Go applications:

```text
├── cmd/
│   └── server/
│       └── main.go              # Main application entrypoint
├── internal/
│   ├── config/                  # Configuration loaders (.env, etc.)
│   ├── database/                # Connection Managers for multi-tenant databases
│   ├── graphql/                 # gqlgen generated optional Gateway
│   ├── handlers/                # Gin REST API Controllers
│   ├── middleware/              # Auth, Prometheus, Rate Limiter, OpenTelemetry
│   ├── models/                  # GORM Definitions and Domain Logic
│   ├── services/                # Business logic (Service Builder, Migration Engine, etc.)
│   └── tracing/                 # OpenTelemetry setup
├── pkg/
│   ├── logger/                  # Zap structured logger
│   └── response/                # Standardized Gin API response handlers
├── tests/                       # Test suites
│   ├── integration/
│   ├── uat/
│   └── unit/
├── docker/                      # Prometheus and deployment configurations
├── Dockerfile                   # Service containerization
├── Makefile                     # Task runner
└── docker-compose.yml           # Local multi-container environment
```

## 4. CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`, the **Dynamic CRUD Engine** automatically resolves REST operations for generated services:

- **Create**: Maps `POST /api/v1/data/{slug}` to dynamic GORM `Create()` / `CreateInBatches()`.
- **Read**: Resolves `GET /api/v1/data/{slug}/{id}`.
- **Update**: Performs structural validation and updates `PUT /api/v1/data/{slug}/{id}`.
- **Delete**: Supports soft/hard deletes `DELETE /api/v1/data/{slug}/{id}`.
- **List / Query Engine**: Handled via `GET /api/v1/data/{slug}`. Includes pagination (`?page=1&limit=20`), advanced filtering (`?email=user@example.com`), and sorting dynamically passed into GORM.

## 5. Schema Migration Engine

Implemented in `internal/services/migration_service.go`, the migration engine safely applies structural updates when users alter their models in the Service Builder.

- **Capabilities**: Add column, Drop column, Rename column, Change column type.
- **Execution**: Uses GORM's `Migrator()` interface on the target dynamically connected database.
- **Safety**: Modifies the `migrations` metadata table to track operations, making state rollbacks and historical auditing natively possible.

## 6. GraphQL Gateway

Located at `internal/graphql/gateway.go`, the system provides an optional GraphQL interface layer on top of the dynamic REST core.

- Powered by `github.com/99designs/gqlgen/graphql`.
- Generates queries and mutations dynamically inferred from the `services` and `fields` metadata.
- Mounted centrally (e.g., `POST /api/v1/graphql`) acting as a flexible Gateway for complex nested querying.

## 7. Backup System

The Backup Engine (`internal/services/backup_service.go`) ensures data durability for defined services:

- **Data Snapshots**: Extracts raw records from dynamic tables.
- **Formats**: Dumps into standardized JSON payload structures.
- **Restore**: Can re-import data rows dynamically for failure recovery.
- Accessible via the Control Plane APIs (e.g., `POST /api/v1/cms/backup/service/{id}`).

## 8. Observability Integration

Production-ready telemetry is built into the infrastructure:

- **Logging**: Configured via `pkg/logger/` using **Uber's Zap**. Provides JSON structured logging, injecting Trace and Correlation IDs dynamically from the context.
- **Metrics**: Standard HTTP performance (latency, requests) tracked by **Prometheus** via `internal/middleware/prometheus.go` and exposed on `/metrics`.
- **Tracing**: Deep request lifecycle observation using **OpenTelemetry** (`internal/tracing/tracing.go`), forwarding spans to **Jaeger** for query latency analysis and distributed monitoring.

## 9. Unit Tests

Robust testing ensures stability across core logic, targeting 80%+ coverage across handlers, services, and repositories.

- Structured by scope: `tests/unit/`, `tests/integration/`, and `tests/uat/`.
- **Run command**:
  ```bash
  go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
  ```
- Heavily utilizes the `github.com/glebarez/sqlite` memory driver (`:memory:`) to spin up isolated, fast iterations for testing GORM schemas and database logic.

## 10. Docker Setup

Containerization makes the system cloud-ready:
- **Dockerfile**: Minimal multi-stage build targeting scratch/alpine containers.
- **docker-compose.yml**: Orchestrates the entire ecosystem containing the main application alongside dependency infrastructure like Prometheus, Jaeger, and PostgreSQL (or other connected DBs).

## 11. Swagger Documentation

API Documentation is auto-generated using `swaggo/swag` annotations integrated into the Gin handlers.

- **Generator Command**:
  ```bash
  swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
  ```
- Exposes standard OpenAPI definitions (`swagger.json`, `swagger.yaml`) that define Control Plane interactions (Databases, Services, Relations).
