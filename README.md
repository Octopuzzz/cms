# Dynamic CMS + API Builder Backend Platform

This repository contains a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. It allows developers to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints. The system is modular, scalable, and cloud-ready, following Clean Architecture and Domain Driven Design (DDD) principles.

## Core Technologies

- **Language:** Go 1.24+
- **API Layers:** REST (default using Gin), GraphQL (optional using gqlgen)
- **ORM:** GORM
- **Database Support:** PostgreSQL, MySQL, MongoDB, SQLite (for testing)
- **Observability:** Zap (Structured Logging), Prometheus (Metrics), OpenTelemetry/Jaeger (Tracing)
- **Caching & Message Queue Readiness:** Architecture designed to easily integrate Redis and NATS/Kafka.
- **Containerization:** Docker & Docker Compose
- **Documentation:** Swagger / OpenAPI

---

## High Level Platform Architecture

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

## Metadata Database Schema

The core platform configuration is stored in the Metadata Database. These tables track services, fields, and connections:

1.  **`database_connections`**: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2.  **`services`**: Represents user-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3.  **`fields`**: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults, etc.
4.  **`service_permissions`**: Connects `Role` to `Service` providing fine-grained access checks (e.g., CanCreate, CanRead). Contains row-level and field-level capabilities.
5.  **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6.  **`backups`**: Stores snapshot records or schema outputs.
7.  **`users`** and **`roles`**: Manage authentication and authorization for the control plane.
8.  **`audit_logs`**: Detailed logging of structural and data-level modifications.

---

## Go Project Structure

The project structure adheres to standard Go layout practices:

-   `cmd/server/`: The main entry point for the application. Initializes the server, database, and routes.
-   `internal/api/`, `internal/handlers/`: Presentation layer. Contains HTTP handlers and routing logic mapping to services. Includes Swagger annotations.
-   `internal/services/`: Application layer. Holds the core business logic (CMS Control Plane, CRUD Engine, Schema Migration, Backup System, etc.).
-   `internal/models/`: Domain layer. Defines the core entities (Service, Field, User, DatabaseConnection) representing the metadata schema.
-   `internal/database/`: Infrastructure layer. Database connection management and connection pooling logic.
-   `internal/middleware/`: HTTP middlewares (Authentication, Rate Limiting, Prometheus Metrics).
-   `internal/tracing/`: OpenTelemetry and Jaeger tracing setup.
-   `internal/graphql/`: GraphQL gateway implementation using `gqlgen`.
-   `pkg/`: Reusable packages such as `logger` (Zap) and `response` formatting.
-   `tests/`: Comprehensive test suites (`unit`, `integration`, `uat`).
-   `docker/`: Docker configurations.

---

## Core Components Implementation

### CMS Control Plane & Service Builder
The control plane handles management operations such as database connections, dynamic service creation, and schema definition. Managed by `internal/services/service_service.go` and `internal/services/dbconn_service.go`. It connects to external databases via `database_connections` and dynamically provisions tables and fields.

### Dynamic CRUD Engine & Query Engine
When a service is created, standard CRUD endpoints are dynamically served by `internal/services/dynamic_data_service.go`. It handles standard operations (`GET`, `POST`, `PUT`, `DELETE`) mapping to generic underlying databases, and applies automated filtering, dynamic joining, and pagination logic on the fly. It respects RBAC and row-level filtering.

### Schema Migration Engine
Safe schema changes are orchestrated by `internal/services/migration_service.go`. It monitors schema diffs and applies operations like `Add Column` or `Drop Column` using GORM’s Migrator. It tracks the migration status in the `migrations` table and supports rollback procedures.

### GraphQL Gateway
A generic GraphQL layer is generated on-the-fly for existing services. Implemented in `internal/graphql/gateway.go` utilizing `gqlgen`, it allows queries and mutations against dynamic data models via a centralized `POST /api/v1/graphql` endpoint.

### Backup System
The system supports snapshotting and backups for dynamic service data. Handled by `internal/services/backup_service.go`, it creates JSON or SQL-based snapshots of a specific service's data, which can later be restored, acting as an integrated disaster recovery feature.

### Observability Integration
Observability is a first-class citizen in the platform:
-   **Logging:** Structured JSON logging using `Zap` (`pkg/logger`) with context-aware trace IDs.
-   **Metrics:** Prometheus interceptors tracking request latency, rates, and errors (`internal/middleware/prometheus.go`). Exported to `/metrics`.
-   **Tracing:** OpenTelemetry combined with Jaeger (`internal/tracing/tracing.go`) to trace requests across the API, Service, and Database layers.

---

## Testing

The platform maintains comprehensive unit, integration, and UAT tests ensuring stable dynamic deployments.

### Running Tests
To run all tests with coverage reporting:
```bash
make test-coverage
```

Other available targets:
-   `make test-unit`
-   `make test-integration`
-   `make test-uat`

Coverage is configured to exceed 80% across the critical repository, service, and handler paths. In-memory SQLite is used extensively for rapid, isolated unit testing.

---

## Docker Setup

The platform is containerized using Docker and orchestrated with Docker Compose, providing a ready-to-run environment containing the API, PostgreSQL (Metadata), Redis, Jaeger, and Prometheus.

### Quick Start
```bash
make docker-up
```
This boots the full stack. The API is accessible at `http://localhost:8080`.

---

## API Documentation (Swagger)

Swagger/OpenAPI documentation is auto-generated and served directly by the platform.

### Generating Docs
```bash
make swagger
```
After starting the server, Swagger UI is available at `/swagger/index.html`.
