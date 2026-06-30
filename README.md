# Dynamic CMS & API Builder Platform

Welcome to the self-hosted, Go-native Backend-as-a-Service (BaaS) platform. This project dynamically creates backend services, manages database connections, defines data schemas, auto-generates CRUD REST APIs and optional GraphQL endpoints, and provides state-of-the-art observability.

This platform operates as a powerful backend infrastructure generator, similar to Hasura or Supabase.

---

## 1. System Architecture Explanation

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles to ensure maximum modularity, scalability, and maintainability.

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

### Core Layers:
- **Presentation Layer (`internal/handlers/`, `internal/graphql/`)**: Provides the Gin HTTP Router and Handlers mapping requests to internal services, including REST default APIs and optional GraphQL endpoints (using `gqlgen`).
- **Application Layer (`internal/services/`)**: Orchestrates the business logic. Controls data access, schema generation, and service operations requested by the handlers.
- **Domain Layer (`internal/models/`)**: Contains the core data models/entities (e.g., `Service`, `Field`, `DatabaseConnection`, `User`, `Role`).
- **Infrastructure Layer (`internal/tracing/`, `internal/middleware/`, `pkg/logger/`)**: Manages cross-cutting concerns like telemetry (OpenTelemetry), metrics (Prometheus), structured logging (Zap), and connection pooling.

---

## 2. Metadata Database Schema

The core configuration for the platform is stored in a relational Metadata Database (typically PostgreSQL).

### Core Tables (`internal/models/models.go`)
- **`database_connections`**: Stores external DB configurations (`id`, `name`, `type`, `host`, `port`, `username`, `password`, `database_name`).
- **`services`**: Represents dynamic user-created data models. Includes `id`, `name`, `database_connection_id`, `db_table_name`, `created_at`, `updated_at`.
- **`fields`**: Defines attributes for each service. Tracks `name`, `type` (string, int, float, boolean, uuid, json, timestamp, array), `nullable`, `unique`, `default_value`, and `index`.
- **`service_permissions`**: Maps `Role` to `Service` for RBAC (e.g., CanCreate, CanRead).
- **`migrations`**: Tracks DDL executions and metadata for schema rollbacks.
- **`backups`**: Stores records of generated backups and snapshot data.
- **`users` / `roles`**: Manages authentication and authorization for the CMS Control Plane.
- **`audit_logs`**: Detailed logs of structural and data-level modifications.

---

## 3. Go Project Structure

The project implements a clean and modular structure:

```text
├── cmd/
│   └── server/          # Main application entry point (main.go)
├── docker/              # Docker configuration files
├── internal/
│   ├── config/          # Environment configuration mapping
│   ├── database/        # DB connection managers
│   ├── graphql/         # GraphQL gateway generation and execution
│   ├── handlers/        # Gin HTTP route handlers
│   ├── middleware/      # Auth, Prometheus, OpenTelemetry, Rate Limiting
│   ├── models/          # Domain layer and metadata schemas
│   ├── services/        # Application logic (CMS, CRUD, Backup, Migration)
│   └── tracing/         # OpenTelemetry setup
├── pkg/
│   ├── errors/          # Custom error definitions
│   ├── logger/          # Zap structured logger wrapper
│   ├── pagination/      # Standardized pagination utilities
│   ├── response/        # Standard Gin JSON response helpers
│   └── validation/      # Input validation utilities
├── tests/               # Unit, Integration, and UAT test suites
├── Makefile             # Automation for build, run, and test steps
├── Dockerfile           # App containerization
└── docker-compose.yml   # Dev/Prod infrastructure setup
```

---

## 4. CRUD Engine Implementation

Handled by `internal/services/dynamic_data_service.go`, the **Dynamic CRUD Engine** acts as the core API builder:

- **Generation:** Maps standard endpoints (e.g., `GET /api/v1/data/{service_slug}`) to GORM database operations based on the dynamic metadata of the `Service`.
- **Operations Supported:** Create, Read (by ID), Update, Delete, List.
- **Query Engine Features:**
  - *Filtering:* E.g., `GET /api/v1/data/users?email=test@example.com`
  - *Sorting:* E.g., `GET /api/v1/data/users?sort=created_at:desc`
  - *Pagination:* E.g., `GET /api/v1/data/users?page=1&limit=20`
  - *Dynamic Joins:* Supported via foreign key metadata mapping across tables.
- **Security:** Verifies Role-Based Access Control and Row-Level security dynamically on every request. Performance optimizations include caching and batched inserts (N+1 query resolution).

---

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`. This engine enables safe alterations to user-defined backend structures:

- **Supported Operations:** Add Column, Drop Column, Rename Column, Change Column Type.
- **Mechanism:** Leverages GORM's `.Migrator()` alongside custom logic to dynamically reflect `Service` and `Field` updates in the physical database schema.
- **Safety:** Logs the complete state and status of executed operations into the `migrations` metadata table. The system tracks diffs, facilitating snapshot retention and rollback capabilities.

---

## 6. GraphQL Gateway

The platform leverages `gqlgen` to provide an optional GraphQL interface alongside the default REST API, built inside `internal/graphql/gateway.go`.

- **Automatic Generation:** Transforms `Service` and `Field` metadata schemas into GraphQL types.
- **Capabilities:**
  - Generic queries for reading single or multiple records.
  - Mutations for dynamic entity creation and updates.
  - Resolving nested relational data directly via GraphQL schema associations.
- **Endpoint:** `POST /api/v1/graphql` handles all dynamic GraphQL requests.

---

## 7. Backup System

Managed by `internal/services/backup_service.go`. The system provides infrastructure safety features for managed databases.

- **Supported Features:**
  - Table backup
  - Service snapshots (mapping dynamic table contents to snapshot states)
  - Schema backups
- **Format:** Supports JSON snapshot payloads.
- **Endpoints:**
  - `POST /api/v1/cms/backup/service/{service_id}` generates data snapshots.
  - Restoring capabilities via JSON payload decoding.

---

## 8. Observability Integration

Comprehensive tracing, metrics, and logging are implemented to provide a production-ready observability suite.

- **Logging (Zap):** Implemented in `pkg/logger/`. Structured JSON logs capture request paths, error data, and SQL query context, automatically binding correlation/trace IDs from the context.
- **Metrics (Prometheus):** Intercepted via `internal/middleware/prometheus.go`. Exposes standard metrics like request latency, error rates, and HTTP status code counts on the `/metrics` endpoint.
- **Tracing (OpenTelemetry):** Initialized in `internal/tracing/` to wrap complex SQL query commands and HTTP network logic, enabling distributed tracking across the core CMS and dynamic API requests (e.g., exporting to Jaeger).

---

## 9. Unit Tests

Unit tests are written using standard Go testing practices, heavily leveraging in-memory SQLite (`:memory:`) via `github.com/glebarez/sqlite` to mock DB connections, thus isolating logic.

- **Coverage:** The platform aims for high coverage (>80%) across the core `internal/services/`, `internal/handlers/`, and `internal/models/` layers.
- **Commands:**
  - Standard run: `go test ./...`
  - Run all with coverage output: `make test-coverage` (which internally executes `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`).

---

## 10. Docker Setup

Containerization provides an easy path to deploy both the platform and its required infrastructure elements.

- **Dockerfile:** An optimized, multi-stage Docker build to compile the Go application into a lightweight runtime image.
- **docker-compose.yml:** Provisions the main application, a PostgreSQL instance (for metadata), Redis (for caching/sessions), Prometheus, and Jaeger.
  - *Run locally with:* `docker-compose up -d`

---

## 11. Swagger Documentation

The platform automatically provides robust OpenAPI 3.0 documentation for all system-level REST endpoints using Swaggo.

- **Location:** The generated documentation lives in the `docs/` folder (note: ignored by version control).
- **Generation Command:**
  ```bash
  swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
  ```
- Exposes standard routes defining expected request/response bodies for the CMS Control plane and Dynamic CRUD engine.
