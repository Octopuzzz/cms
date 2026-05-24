# Dynamic CMS & API Builder - Backend Platform

This repository is a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. It operates similarly to platforms like Hasura or Supabase, acting as a fully self-hosted and Go-native backend infrastructure generator.

The system allows developers to:
- Connect databases
- Create data models dynamically
- Generate CRUD APIs automatically
- Generate optional GraphQL APIs
- Manage schema migrations
- Monitor logs and performance
- Manage backups
- Scale services

---

## 1. System Architecture Explanation

The Backend Platform follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles, separating concerns into distinct layers:

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

- **Presentation Layer**: Handled by the Gin HTTP Router (`internal/handlers/`), acting as an API Gateway. Includes standard REST endpoints and an optional GraphQL gateway via `gqlgen`.
- **Application Layer**: Contains business logic (`internal/services/`). Responsibilities include managing services, defining schemas dynamically, processing data, and triggering operations (CRUD/GraphQL).
- **Domain Layer**: The data models (`internal/models/`). It encapsulates standard platform entities (`Service`, `Field`, `DatabaseConnection`, `User`, `Role`, etc.).
- **Infrastructure Layer**: Cross-cutting capabilities like telemetry (`internal/tracing/`), database connections (`internal/database/`), logging (`pkg/logger/`), and Prometheus metrics (`internal/middleware/`).

---

## 2. Metadata Database Schema

The CMS manages platform state using an internal Metadata Database (usually PostgreSQL or SQLite). Core entities (defined in `internal/models/models.go`) include:

- `database_connections`: External database configurations (id, type, host, credentials, connection pool settings).
- `services`: User-created data models. They map to specific `database_connection_id` and maintain underlying DB representations via `db_table_name`.
- `fields`: Service fields mapping type (string, integer, UUID, json, etc.), nullability, constraints, defaults, and relationships (One-to-One, One-to-Many, Many-to-Many).
- `service_permissions`: Granular RBAC permissions for services and field-level permissions.
- `migrations`: Track dynamic table and column schema operations, enabling audit trails and rollbacks.
- `backups`: Records of snapshots (both schema definitions and table data).
- `users`, `roles`, `permissions`: Control plane authentication.
- `audit_logs`: Telemetry to monitor platform data modifications.
- `api_tokens`: Managing access scopes and expiration times for API access.

---

## 3. Go Project Structure

The project maintains a clean, modular layout typical for scalable Go applications:

```text
.
├── cmd/
│   └── server/          # Main entrypoint, dependency injection, runtime init
├── internal/
│   ├── api/             # High-level API integrations
│   ├── config/          # Application configuration loaders
│   ├── database/        # External database connection managers & pool configuration
│   ├── graphql/         # GraphQL optional gateway schema/resolver engine
│   ├── handlers/        # Gin HTTP request/response handlers
│   ├── middleware/      # Auth, observability (Prometheus), rate limiting
│   ├── models/          # GORM definitions for the Metadata Schema
│   ├── services/        # Application business logic (CRUD Engine, CMS Plane, etc.)
│   └── tracing/         # OpenTelemetry Jaeger tracing setup
├── pkg/                 # Sharable utilities (e.g., logger, standard responses, pagination)
├── tests/               # Unit, integration, and UAT (end-to-end) testing
├── docker/              # Deployment scripts and container resources
├── Dockerfile           # Standardized multi-stage Go build image
├── docker-compose.yml   # Multi-service local environment (Postgres, Redis, App)
└── Makefile             # Task runner automation
```

---

## 4. CRUD Engine & Query Engine Implementation

When a developer dynamically defines a `Service` and its `Fields`, the **Dynamic CRUD Engine** (`internal/services/dynamic_data_service.go`) instantly maps HTTP requests to the target SQL database dynamically using GORM on the fly.

- **Dynamic REST APIs**: Generates `GET /api/v1/data/{slug}`, `POST /api/v1/data/{slug}`, `PUT /api/v1/data/{slug}/{id}`, and `DELETE` endpoints automatically.
- **Advanced Query Engine**:
  - Allows `GET /api/v1/data/users?email=test@example.com` for direct filtering.
  - Supports ordering: `?sort=created_at:desc`.
  - Implements pagination using standard logic: `?page=1&limit=20`.
  - Facilitates dynamic joins dynamically extracting relations from the `fields` metadata.

---

## 5. Schema Migration Engine

The **Schema Migration Engine** (`internal/services/migration_service.go`) manages zero-downtime alterations to data tables connected to `Services`.

- **Safe Alterations**: Automatically translates metadata `Field` modifications (e.g. Add, Drop, Rename column, Type Change) into native SQL DDL using GORM's `Migrator()`.
- **Safety**: Automatically tracks statuses (pending, applied, failed) in the `migrations` table, supporting snapshot retention to allow rollbacks on failure.

---

## 6. GraphQL Gateway

The platform optionally hosts an auto-generated GraphQL Gateway via `/api/v1/graphql` (`internal/graphql/gateway.go`).

- Powered by `gqlgen`, the system maps the active `services` into GraphQL types dynamically.
- Supported features: Root Queries for listing and fetching entities, Root Mutations for CRUD operations, and nested associative resolving for relationships defined in the metadata layer.

---

## 7. Backup System

Managed by `internal/services/backup_service.go`, ensuring platform durability.

- Allows ad-hoc manual or programmatic backups of both table schema and row data payloads (`POST /api/v1/cms/backups`).
- Dumps data as serialized JSON payloads natively within the internal metadata or to files, and implements snapshot restorations for fast operational recovery.

---

## 8. Observability Integration

Production-ready observability ensures full system state awareness:

- **Logging**: Structured JSON logging powered by `zap` (`pkg/logger/`), tracking request/response lifecycles, queries, and errors with context correlation.
- **Metrics**: Standardized metrics through Prometheus (`internal/middleware/prometheus.go`), exporting request latency, error rates, and query durations at `/metrics`.
- **Tracing**: Fully integrated OpenTelemetry and Jaeger (`internal/tracing/`). Wrap network boundaries and database queries to visualize distributed system traces.

---

## 9. Unit Testing

The platform enforces high-quality stability with comprehensive automated testing (`tests/`).

- **Coverage Goal**: At least 80% coverage across `Repository`, `Service`, and `API handlers`.
- Use the standard `testing` library to execute fast, isolated unit tests using an in-memory SQLite backend for mock connections.
- Command shortcuts in `Makefile`:
  - Run all: `make test`
  - Unit tests: `make test-unit`
  - Integration: `make test-integration`
  - Coverage: `make test-coverage` (exports an HTML report for CI systems)

---

## 10. Docker Setup

Platform deployment is container-native, defined in `Dockerfile` and `docker-compose.yml`.

- **Dockerfile**: Multi-stage lightweight `alpine` build mapping the binary executable.
- **Docker Compose**: Orchestrates dependent services like the backend binary, external user PostgreSQL nodes, Redis for cache/rate limiting, and observability tools (Jaeger, Prometheus) into an integrated stack.
- To run: `make docker-up`

---

## 11. Swagger Documentation

API documentation is generated directly from source code using standard Swagger / OpenAPI tools.

- The system scans controller/handler signatures using `swag` to output interactive docs.
- The compiled definitions are generated at `docs/` and normally accessible via the `/swagger/index.html` UI when hosted.
- Generate via command: `make swagger` (requires `swaggo/swag/cmd/swag`).

---

## Quick Start

1. Start databases via `make docker-up`.
2. Install dependencies: `make deps`.
3. Auto-generate swagger specs: `make swagger`.
4. Run locally with hot reload: `make dev`.
5. Access the API at `http://localhost:8080`.
