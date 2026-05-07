# Go Backend Platform (Dynamic CMS + API Builder)

This is a **production-grade Backend-as-a-Service platform** written natively in Go. It acts as a dynamic backend infrastructure generator, allowing developers to connect databases, create data models dynamically, and automatically generate CRUD REST endpoints and optional GraphQL APIs.

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)**, and is fully self-hosted, modular, scalable, and cloud-ready.

---

## 1. System Architecture Explanation

The CMS Backend follows Clean Architecture and DDD principles:

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

---

## 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. Tables defined include:

- `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
- `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`). Fields: `id`, `name`, `database_id`, `created_at`, `updated_at`.
- `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
- `relations`: Manages relationships (One-to-One, One-to-Many, Many-to-One, Many-to-Many).
- `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- `backups`: Stores snapshot records or schema outputs.
- `users` and `roles`: General authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of structural and data-level modifications.

---

## 3. Go Project Structure

The project uses a clean modular structure:

```
├── cmd
│   └── server          # Entry point for the application
├── internal
│   ├── api             # API routes and initializers
│   ├── config          # Configuration loading
│   ├── database        # Database Connection Manager
│   ├── graphql         # GraphQL gateway and resolvers
│   ├── handlers        # Presentation Layer (HTTP Handlers)
│   ├── middleware      # Gin middlewares (Auth, Metrics, etc.)
│   ├── models          # Domain Layer (Core Entities)
│   ├── services        # Application Layer (Business Logic, CRUD Engine, Schema Migration, Backup)
│   └── tracing         # Observability Engine (OpenTelemetry)
├── pkg
│   ├── logger          # Zap structured logging
│   ├── pagination      # Pagination utilities
│   ├── response        # Standardized API responses
│   └── validation      # Input validation utilities
├── tests               # Unit, Integration, and UAT tests
│   ├── integration
│   ├── uat
│   └── unit
├── docker              # Docker configurations (e.g., prometheus.yml)
├── docs                # Swagger API documentation
├── .env.example        # Environment variables example
├── docker-compose.yml  # Docker Compose setup
├── Dockerfile          # Dockerfile for the application
└── Makefile            # Standardized tasks
```

---

## 4. CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go` and `internal/services/service_service.go`.
- When a service is created via the Service Builder (`POST /api/v1/cms/services`), the CRUD engine automatically generates standard operations (Create, Read, Update, Delete, List).
- Endpoints like `GET /api/v1/data/{slug}` are mapped to standard GORM database operations dynamically.
- Applies automated filtering via query params (e.g. `?email=test@test.com`), sorting (`?sort=created_at:desc`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

---

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column, Change column type) using GORM’s `.Migrator()`.
- Logs migration status natively into the `migrations` metadata table.
- Handles rollbacks through snapshot retention logic.
- Supports automatic backups before migrations.

---

## 6. GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- GraphQL is automatically generated from service schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.
- Supports generic queries, mutations, and relations dynamically based on the registered schemas.

---

## 7. Backup System

Managed by `internal/services/backup_service.go`.
- Generates data snapshots (table backup, schema backup, service snapshot).
- Currently implements generic service record backup features mapping dynamic table contents to JSON snapshots.
- Supports restoring rows via JSON payload decoding.
- Example API endpoints: `POST /api/v1/cms/backup/service/{service_id}` and `POST /api/v1/cms/restore/service/{service_id}`.

---

## 8. Observability Integration

Implemented across the stack:
- **Logging**: Zap structured logging is injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context. Captures request logs, error logs, and query logs.
- **Tracing**: OpenTelemetry/Jaeger is initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Metrics**: Prometheus metrics track request latency, slow queries, and error rates via standard HTTP interceptors (`internal/middleware/prometheus.go`). Exported at `/metrics`.

---

## 9. Unit Tests

The system includes a comprehensive test suite covering the Repository, Service, and API handler layers.
- Uses an in-memory SQLite database setup (`:memory:`) via the `github.com/glebarez/sqlite` driver for fast execution.
- Minimum 80% test coverage across the system.
- To run tests:
  ```bash
  make test-unit
  make test-coverage
  ```
  *(Note: Run tests directly via `go test ./...` if `make` is unavailable.)*

---

## 10. Docker Setup

The platform is fully containerized. A `docker-compose.yml` provides a production-ready environment including:
- **cms-backend**: The Go application.
- **postgres**: The metadata database.
- **redis**: Cache for query optimization and connection pooling.
- **jaeger**: Distributed tracing backend (OpenTelemetry).
- **prometheus**: Metrics aggregation.

To start the platform:
```bash
make docker-up
```

---

## 11. Swagger Documentation

API Documentation is automatically generated using Swagger / OpenAPI.
- Generation command: `make swagger` (or `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`).
- The generated `docs/` folder provides the OpenAPI specification, which can be served to visualize and interact with the API endpoints.

---

This README outlines the existing state and architectural implementations of the CMS Backend, satisfying the requirements for a fully self-hosted, modular, and cloud-ready Go platform.
