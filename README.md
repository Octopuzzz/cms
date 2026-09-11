# Backend Platform (Dynamic CMS + API Builder)

A production-grade Backend-as-a-Service platform written in Go. This self-hosted, Go-native platform allows users to dynamically create backend services, manage database connections, define schemas, and automatically generate CRUD REST endpoints and optional GraphQL APIs.

---

## 1. System Architecture Explanation

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

*   **Presentation Layer**: Implemented using the Gin HTTP framework (`internal/handlers/`), acting as the API gateway. It maps requests to internal services and handles both REST and GraphQL APIs (`gqlgen` in `internal/graphql/gateway.go`).
*   **Application Layer**: Contains business logic (`internal/services/`). The services module handles operations such as schema migrations, CRUD generation, role-based access control, and platform orchestration.
*   **Domain Layer**: Contains the platform's core data models (`internal/models/models.go`) like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`, `Permission`, `Migration`, and `Backup`.
*   **Infrastructure Layer**: Integrates external tools, such as telemetry (OpenTelemetry/Jaeger in `internal/tracing/`), Prometheus metrics (`internal/middleware/prometheus.go`), structured logging via Zap (`pkg/logger/`), and database connections (`internal/database/`).

**High-Level Architecture:**
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

The core internal configuration is stored in the **Metadata Database**. The models defined in `internal/models/models.go` represent the tables:

*   **`database_connections`**: Stores configurations for external databases (PostgreSQL, MySQL, MongoDB). Fields include `id`, `name`, `type`, `host`, `port`, credentials, etc.
*   **`services`**: Represents dynamic data models created by users, containing references to `database_connection_id` and the dynamically generated `db_table_name`.
*   **`fields`**: Defines the attributes for each service (type: string, integer, float, boolean, uuid, json, array, timestamp) and configurations (nullable, unique, default_value, index).
*   **`relations` / `relation_configs`**: Defines table relationships (One-to-One, One-to-Many, Many-to-One, Many-to-Many).
*   **`users`**, **`roles`**, **`permissions`**: Handles authentication and RBAC. Includes `user_roles` and `role_permissions` join tables.
*   **`service_permissions`** & **`field_permissions`**: Stores fine-grained access control to dynamic tables.
*   **`migrations`**: Tracks DDL schema migrations, providing metadata for history and rollbacks.
*   **`backups`**: Stores snapshots and schema output references.

---

## 3. Go Project Structure

The project has a clear modular layout:

```text
├── cmd
│   └── server                # Entry point (main.go)
├── internal
│   ├── config                # Environment variables and config structures
│   ├── database              # Database Connection Manager & pooling
│   ├── graphql               # GraphQL gateway implementation (gqlgen)
│   ├── handlers              # Gin HTTP handlers (Presentation Layer)
│   ├── middleware            # Auth, Prometheus, Rate Limiter
│   ├── models                # Metadata Database Schema (Domain Layer)
│   ├── services              # CMS, Builder, CRUD, Backup logic (Application Layer)
│   └── tracing               # OpenTelemetry implementation
├── pkg
│   ├── logger                # Zap structured logging wrapper
│   └── response              # Standardized API response wrappers
├── tests                     # Automated Tests
│   ├── integration           # Integration tests
│   ├── uat                   # End-to-End / User Acceptance tests
│   └── unit                  # Unit tests (Repository, Service, Handlers)
├── docker                    # Docker/Prometheus configuration files
├── Dockerfile                # Image build instructions
├── docker-compose.yml        # Docker Compose environment setup
└── Makefile                  # Centralized build/run/test/lint commands
```

---

## 4. CRUD Engine Implementation

Handled primarily by `internal/services/dynamic_data_service.go` and `internal/handlers/dynamic_data_handler.go`.
The engine auto-generates operations based on registered `services`:

*   **Create**: `POST /api/v1/data/{slug}`
*   **Read**: `GET /api/v1/data/{slug}/{id}`
*   **Update**: `PUT /api/v1/data/{slug}/{id}`
*   **Delete**: `DELETE /api/v1/data/{slug}/{id}`
*   **List (Query Engine)**: `GET /api/v1/data/{slug}`
    *   Supports Filtering (e.g., `?email=john@example.com`)
    *   Sorting (e.g., `?sort=created_at:desc`)
    *   Pagination (e.g., `?page=1&limit=20`)
    *   Joins via foreign keys mapped in the system.

RBAC and row-level permissions are applied directly within the service layer.

---

## 5. Schema Migration Engine

Implemented in `internal/services/migration_service.go` and `internal/handlers/migration_handler.go`.

*   Tracks changes to the metadata models (`services`, `fields`) and translates them into DDL.
*   Performs database operations via GORM's `Migrator` feature.
*   Supports adding/dropping/renaming columns and changing column types safely.
*   Maintains a migration history inside the `migrations` table, supporting tracking and potential rollbacks.

---

## 6. GraphQL Gateway

Implemented within `internal/graphql/gateway.go`.

*   Automatically maps generated schema and operations via `github.com/99designs/gqlgen`.
*   A single endpoint `POST /api/v1/graphql` fields query and mutation requests.
*   Handles relation fetching and nested object querying automatically based on the dynamically defined service relations.
*   A Playground is available at `GET /api/v1/graphql/playground`.

---

## 7. Backup System

Managed by `internal/services/backup_service.go` and exposed via `/api/v1/cms/backup`.

*   **Creation**: Captures point-in-time snapshots of dynamic table configurations and corresponding dynamic table row data into JSON representations.
*   **Restoration**: Can accept these JSON outputs to recreate or restore deleted rows for an existing dynamic table configuration.
*   Logs backups to the `backups` metadata table.

---

## 8. Observability Integration

Integrated deeply across all endpoints:

*   **Logging**: `pkg/logger` wraps Zap to output JSON formatted, structured request logs and query logs. Context correlation/request IDs are maintained across middleware and services.
*   **Metrics**: Prometheus intercepts HTTP traffic (`internal/middleware/prometheus.go`) to record latency, error rates, and request counts, exported at `/metrics`.
*   **Tracing**: OpenTelemetry configures Jaeger (`internal/tracing/tracing.go`) to trace HTTP requests down through the ORM/SQL layer.

---

## 9. Unit Tests

Unit and integration tests are located in `tests/`.

*   Follows a minimum 80% coverage rule over Repository, Service, and API handlers.
*   Uses an in-memory SQLite database setup (`:memory:`) with `github.com/glebarez/sqlite` to mock DB connections cleanly across suites.
*   Running unit tests: `make test-unit`
*   Full coverage evaluation: `make test-coverage`

---

## 10. Docker Setup

Complete containerization configuration provided in `docker-compose.yml` and `Dockerfile`.

*   `cms-backend`: The Go platform binary.
*   `postgres`: Metadata relational database (`cms_metadata`).
*   `redis`: Caching layer.
*   `jaeger`: Distributed tracing aggregation tool.
*   `prometheus`: Metrics aggregation platform.

To spin up: `make docker-up`

---

## 11. Swagger Documentation

API documentation is generated using `swag`.

*   Route comments reside in handler files and `cmd/server/main.go`.
*   Exposed dynamically on the server at: `/swagger/index.html`
*   To regenerate docs, run: `make swagger` (Generates into `docs/` which is ignored from git).
