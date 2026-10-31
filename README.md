# Backend Platform (Dynamic CMS + API Builder)

Welcome to the **Self-Hosted, Go-Native Backend Platform**! This platform acts as a backend infrastructure generator, allowing you to dynamically create backend services, schemas, REST APIs, and optional GraphQL endpoints on the fly. It is fully self-hosted, cloud-ready, and scalable.

## Table of Contents

- [Platform Goal](#platform-goal)
- [System Architecture Explanation](#system-architecture-explanation)
- [Metadata Database Schema](#metadata-database-schema)
- [Go Project Structure](#go-project-structure)
- [Core Components](#core-components)
  - [CRUD Engine Implementation](#crud-engine-implementation)
  - [Schema Migration Engine](#schema-migration-engine)
  - [GraphQL Gateway](#graphql-gateway)
  - [Backup System](#backup-system)
- [Observability Integration](#observability-integration)
- [Unit Tests](#unit-tests)
- [Docker Setup](#docker-setup)
- [Swagger Documentation](#swagger-documentation)

---

## Platform Goal

Build a **Backend-as-a-Service platform** that allows developers to:
- connect external databases (PostgreSQL, MySQL, MongoDB, SQLite)
- create data models dynamically (Service Builder)
- generate CRUD APIs automatically (REST & GraphQL)
- manage schema migrations safely
- monitor logs, metrics, and tracing
- manage database backups

---

## System Architecture Explanation

This platform follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles to guarantee scalability, testability, and separation of concerns.

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

### Layers:
1. **Presentation Layer (`internal/handlers/`, `internal/graphql/`)**: Maps HTTP and GraphQL requests to the application layer.
2. **Application Layer (`internal/services/`)**: Core business logic, acting on requests to execute database migrations, dynamic SQL building, and authorization checks.
3. **Domain Layer (`internal/models/`)**: Platform metadata schemas (e.g. `Service`, `Field`, `DatabaseConnection`, `User`, `Role`).
4. **Infrastructure Layer**: Cross-cutting capabilities like global context telemetry (`internal/tracing/`), Prometheus metrics (`internal/middleware/`), and structured logging (`pkg/logger/`).

---

## Metadata Database Schema

The core control plane relies on an internal database to store the state of dynamically generated services, schemas, and configurations.

- **`database_connections`**: Stores connections to remote user databases. Tracks `id`, `name`, `type`, `host`, `port`, `username`, `password`, `database_name`.
- **`services`**: Tracks dynamically created tables. Columns include `id`, `name`, `database_connection_id`, `db_table_name`, `created_at`, `updated_at`.
- **`fields`**: Maps columns to a specific `service`. Configurations for `name`, `type` (string, int, float, json, etc.), `nullable`, `unique`, `default_value`.
- **`migrations`**: Audit log of schema changes (Add, Drop, Rename) and status.
- **`backups`**: Records of backups and snapshots taken for services.
- **`service_permissions`**: Joins `roles` to `services` to control who can perform CRUD actions.
- **`users`** / **`roles`**: RBAC system to manage control plane access.

---

## Go Project Structure

The codebase is highly modular, separating routing, business logic, configuration, and infrastructure components:

```
.
├── cmd
│   └── server          # Entry point for the application
├── docker
│   └── docker-compose.yml # Example localized Docker deployments
├── internal
│   ├── config          # Application configuration (Env vars)
│   ├── database        # Database Connection Manager & GORM configuration
│   ├── graphql         # GraphQL schema generation and gateway logic (gqlgen)
│   ├── handlers        # REST API endpoints (Gin)
│   ├── middleware      # Gin middleware (Auth, Prometheus, Rate Limiter)
│   ├── models          # Domain entities & schemas
│   ├── services        # Core business logic (CRUD Engine, Service Builder)
│   └── tracing         # OpenTelemetry setup
├── pkg
│   ├── logger          # Zap structured logging wrapper
│   ├── pagination      # Generic pagination logic
│   ├── response        # Standard API responses
│   └── validation      # Input validation utilities
├── tests               # Unit, Integration, and UAT tests
├── Dockerfile          # Multi-stage Docker build for the backend
├── Makefile            # Build, run, and test scripts
└── go.mod              # Module dependencies (Go 1.24)
```

---

## Core Components

### CRUD Engine Implementation
Located in `internal/services/dynamic_data_service.go`, the **Dynamic CRUD Engine** auto-generates functionality when a service is registered.

- **Endpoints**: Operations like `POST /api/v1/data/{slug}` or `GET /api/v1/data/{slug}/{id}` map to database queries built dynamically based on the target `Service` definition.
- **Query Engine**: URL query parameters automatically map to filtering (`?email=john@example.com`), sorting (`?sort=created_at:desc`), nested JOINs, and pagination (`?page=1&limit=20`).
- **Authorization**: Row-level access control and RBAC checks are verified seamlessly within these generated requests.

### Schema Migration Engine
The system natively handles zero-downtime database updates through `internal/services/migration_service.go`.
- Tracks and applies data definition language (DDL) changes like creating a table, adding a column, or updating a column type via GORM's `Migrator()`.
- Supports automatic backups before migrations and features a rollback capability using snapshots.

### GraphQL Gateway
Automatically exposes a single GraphQL endpoint at `POST /api/v1/graphql` via `internal/graphql/gateway.go`.
- Built on `github.com/99designs/gqlgen`.
- Reads `services` and `fields` from the metadata database and generates runtime GraphQL Schemas capable of querying and mutating dynamic data tables seamlessly alongside the REST API.

### Backup System
Handled within `internal/services/backup_service.go`.
- Exposes endpoints to take a full snapshot (e.g. `POST /api/v1/cms/backup/service/{service_id}`) and restore them.
- Extracts current schema states and table payloads into a format (like a JSON snapshot or SQL dump format) for safe recovery in the event of an outage or bad migration.

---

## Observability Integration
The application ships with a world-class observability stack:
- **Logging**: Configured in `pkg/logger/` using Uber's **Zap**. All requests are tagged with `trace_id` for distributed system debugging.
- **Tracing**: Fully integrated with **OpenTelemetry**. Traces cover HTTP middleware (`internal/tracing/`) and internal GORM operations to inspect slow SQL queries in tools like Jaeger.
- **Metrics**: Standard **Prometheus** metrics track request latency, failure rates, and database pool sizing. Exported at `GET /metrics` on the API.

---

## Unit Tests
The repository demands at least **80% test coverage** for the Repository, Service, and API handler layers. Testing is written using Go's standard library combined with in-memory SQLite instances to verify database logic efficiently.

To run the unit tests:
```bash
make test-unit
```

To run coverage metrics across the platform (Requires go 1.24):
```bash
make test-coverage
```

---

## Docker Setup
The repository provides a multi-stage Docker build optimized for production (`Dockerfile`) and an orchestrated environment via `docker-compose.yml`.

- **Build Image**: `docker build -t cms-backend .`
- **Run Locally (Dev)**: `docker-compose up --build -d`

Docker seamlessly networks the Go application with a PostgreSQL database, Prometheus (for metrics scraping), and Jaeger (for OpenTelemetry tracing).

---

## Swagger Documentation
Fully integrated with standard OpenAPI 3.0 via `swaggo/swag`.

1. Generate documentation by installing the CLI: `go install github.com/swaggo/swag/cmd/swag@latest`
2. Run generator: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`
3. Swagger UI will be hosted at `/swagger/index.html` on the server runtime.

*Note: The generated `docs/` folder should not be checked into version control.*
