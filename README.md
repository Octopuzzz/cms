# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade, self-hosted Backend-as-a-Service (BaaS) platform written in Go. It acts as a backend infrastructure generator, allowing users to dynamically create backend services, define schemas, auto-generate REST and GraphQL APIs, manage database connections, handle migrations, and monitor the platform with advanced observability tools.

---

## 1. System Architecture Explanation

The platform is designed following **Clean Architecture** and **Domain-Driven Design (DDD)**.

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

### Core Layers
- **Presentation Layer (`internal/handlers/` & `internal/graphql/`)**: Handles REST and GraphQL requests. Maps incoming requests to the relevant application layer services.
- **Application Layer (`internal/services/`)**: Orchestrates business logic, controls dynamic data access, manages schema builds, and triggers migrations.
- **Domain Layer (`internal/models/`)**: Contains the core entities defining the platform state (e.g., `Service`, `Field`, `DatabaseConnection`, `User`).
- **Infrastructure Layer (`internal/database/`, `internal/tracing/`, `pkg/logger/`)**: Manages external database connections, logging (Zap), metrics (Prometheus), and distributed tracing (OpenTelemetry).

---

## 2. Metadata Database Schema

The CMS state is stored in a metadata database (PostgreSQL/SQLite) that controls the dynamic platform features. Core tables include:

- `database_connections`: Stores configurations for registered external databases (id, name, type, host, port, credentials).
- `services`: Represents dynamically created data models/services. Stores the dynamic table name and relates to a `database_connection_id`.
- `fields`: Defines attributes for a service. Includes field name, type (string, integer, float, boolean, uuid, json, timestamp, array), constraints (nullable, unique, default_value, index).
- `service_permissions`: Role-Based Access Control connecting `roles` and `services`.
- `migrations`: Tracks executed schema changes (DDL) and enables rollbacks.
- `backups`: Records backup snapshots and exports.
- `users` / `roles`: Manages access to the control plane.
- `audit_logs`: Logs platform configuration modifications.

---

## 3. Go Project Structure

The project relies on a clean, modular layout standard to modern Go applications:

```text
.
├── cmd
│   └── server                # Application entrypoint
├── docker                    # Docker and Prometheus configurations
├── internal
│   ├── config                # Application and environment configuration
│   ├── database              # Database connection manager and pooling
│   ├── graphql               # Auto-generated GraphQL Gateway (gqlgen)
│   ├── handlers              # Presentation layer: REST handlers
│   ├── middleware            # Auth, Prometheus, Rate Limiter
│   ├── models                # Domain layer: Platform Entities
│   ├── services              # Application layer: CMS, CRUD, Migration logic
│   └── tracing               # Observability: OpenTelemetry initialization
├── pkg
│   ├── logger                # Structured logging using Zap
│   └── response              # Standardized API response utilities
└── tests
    ├── integration           # Integration tests
    ├── uat                   # User Acceptance Testing
    └── unit                  # Isolated unit tests
```

---

## 4. CRUD Engine Implementation

The Dynamic CRUD Engine automatically generates robust endpoints for user-created services.

- **Endpoints Provided**:
  - `POST /api/v1/data/{slug}` (Create)
  - `GET /api/v1/data/{slug}/{id}` (Read)
  - `PUT /api/v1/data/{slug}/{id}` (Update)
  - `DELETE /api/v1/data/{slug}/{id}` (Delete)
  - `GET /api/v1/data/{slug}` (List & Query)

- **Query Engine Features**:
  - **Filtering**: e.g., `?email=john@example.com`
  - **Sorting**: e.g., `?sort=created_at:desc`
  - **Pagination**: e.g., `?page=1&limit=20`
  - **Relationships/Joins**: Implicit relational loading depending on foreign keys.

The engine uses standard GORM database operations mapped dynamically based on the metadata defined in the `services` and `fields` tables.

---

## 5. Schema Migration Engine

The Schema Migration Engine handles safe schema changes to dynamically generated tables whenever a service or its fields are updated.

- **Capabilities**: Add column, Drop column, Rename column, Change column type.
- **Implementation**: Utilizes GORM's `Migrator()` interface. Migrations are recorded natively in the `migrations` table, providing transparency and rollback capabilities through historical schema tracking.

---

## 6. GraphQL Gateway

The platform auto-generates optional GraphQL schemas representing the created services, offering a flexible query interface alternative to REST.

- **Implementation**: Powered by `github.com/99designs/gqlgen`.
- **Endpoint**: Exposed at `POST /api/v1/graphql`.
- **Capabilities**: Supports robust `query` and `mutation` operations for dynamic tables, alongside nested relation querying.

---

## 7. Backup System

A comprehensive backup mechanism supports taking snapshots of dynamic table states or configurations.

- **Endpoints**:
  - `POST /api/v1/backup/service/{service_id}` (Trigger Backup)
  - `POST /api/v1/restore/service/{service_id}` (Restore Backup)
- **Features**: Generates data snapshots (JSON) and records backup metadata. Enables smooth point-in-time restoration.

---

## 8. Observability Integration

The platform integrates deep observability ensuring production readiness.

- **Logging**: Configured via `go.uber.org/zap` for high-performance structured JSON logging. Request IDs and correlation IDs flow through the Go `context`.
- **Tracing**: Fully integrated with OpenTelemetry (`go.opentelemetry.io/otel`). Traces SQL commands, logic blocks, and external requests to tools like Jaeger.
- **Metrics**: Exposes Prometheus metrics at `/metrics`. `internal/middleware/prometheus.go` tracks HTTP request latencies, sizes, and status codes.

---

## 9. Unit Tests

Testing strategy requires high test confidence, with a required minimum of **80% coverage** across Repository, Service, and API handler layers.

- **Setup**: Tests utilize an in-memory SQLite database setup (`:memory:`) via `github.com/glebarez/sqlite` ensuring hermetic and rapid execution.
- **Commands**:
  - Standard test run: `go test -v ./...`
  - Full coverage check: `make test-coverage` or `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`

---

## 10. Docker Setup

The platform is designed to be fully containerized.

- **Dockerfile**: Located in `docker/Dockerfile`, implements multi-stage builds ensuring minimal production images.
- **Docker Compose**: The `docker-compose.yml` sets up the core CMS API alongside critical dependencies like PostgreSQL (metadata database), Redis (caching), Jaeger (tracing), and Prometheus (metrics).

---

## 11. Swagger Documentation

REST endpoints and core control-plane APIs are thoroughly documented using Swagger/OpenAPI.

- **Implementation**: Handled via standard Swaggo CLI annotations throughout the codebase.
- **Generation**: To regenerate docs, use `go install github.com/swaggo/swag/cmd/swag@latest` followed by `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
- **Note**: The generated `docs/` folder is typically ignored from version control to prevent artifact pollution.
