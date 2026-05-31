# Backend Platform (Dynamic CMS + API Builder)

This repository contains a production-grade Backend-as-a-Service platform written in Go. The platform allows developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability. It follows Clean Architecture and Domain-Driven Design (DDD) principles.

## Platform Features

*   **CMS Control Plane:** Manage databases, services, schemas, relations, migrations, and backups via a RESTful API.
*   **Database Connection Manager:** Connect to PostgreSQL, MySQL, and MongoDB external databases with connection pooling.
*   **Service Builder & Relation Engine:** Dynamically create data models (services), configure fields, and establish One-to-One, One-to-Many, Many-to-One, and Many-to-Many relationships.
*   **Dynamic CRUD & Query Engine:** Automatically generated REST endpoints for defined services with advanced query support (filtering, sorting, pagination, and joins).
*   **GraphQL Gateway (Optional):** Automatically generated GraphQL schemas for dynamic querying.
*   **Schema Migration Engine:** Safe schema updates with automatic tracking and rollback capabilities via GORM's Migrator.
*   **Backup Engine:** Snapshot and restore capabilities for service data and schemas.
*   **Observability:** Integrated Zap logging, Prometheus metrics, and OpenTelemetry/Jaeger tracing.

---

## 1. System Architecture Explanation

The system is built as a highly modular, cloud-ready application:

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

The system employs **Clean Architecture**:
*   **Presentation Layer:** Gin HTTP Handlers (`internal/handlers/`) mapping requests to internal services, and `gqlgen` for the GraphQL gateway.
*   **Application Layer:** Business logic encapsulated in services (`internal/services/`), including the Dynamic Data Service and Service Builder.
*   **Domain Layer:** Core data models (`internal/models/`) like `Service`, `Field`, `DatabaseConnection`, etc.
*   **Infrastructure Layer:** Cross-cutting concerns such as logging (`pkg/logger/`), metrics (`internal/middleware/prometheus.go`), telemetry (`internal/tracing/`), and connection management (`internal/database/`).

---

## 2. Metadata Database Schema

The CMS stores its internal configuration in a metadata database (PostgreSQL recommended, SQLite supported for dev/test). The core tables defined in `internal/models/models.go` include:

*   **`database_connections`:** Tracks external DB credentials, connection pools, and types (id, name, type, host, port, username, password, database, max_open_conns).
*   **`services`:** Represents user-defined data models. Links to a `database_connection_id` and maps to a dynamic table (`db_table_name`).
*   **`fields`:** Defines attributes (string, integer, float, uuid, json, etc.), constraints (nullable, unique), and relations for services.
*   **`service_permissions` & `roles` / `users`:** RBAC configuration mapping roles to service-level capabilities (CanCreate, CanRead, etc.) and field-level/row-level permissions.
*   **`migrations`:** Tracks schema migration history (status, schema snapshot, applied_at).
*   **`backups`:** Records snapshot and schema backup metadata.
*   **`audit_logs`:** Tracks structural and data-level modifications across the system.

---

## 3. Go Project Structure

The repository uses a standard, clean Go modular structure:

```
├── cmd/
│   └── server/          # Main application entrypoint
├── docker/              # Docker configurations (Prometheus, Dockerfile, etc.)
├── internal/
│   ├── api/             # API routes
│   ├── config/          # Environment configuration
│   ├── database/        # Connection manager and DB logic
│   ├── graphql/         # gqlgen integration and dynamic schema execution
│   ├── handlers/        # Gin HTTP controllers
│   ├── middleware/      # Auth, rate limiting, and Prometheus middlewares
│   ├── models/          # GORM domain models
│   ├── services/        # Application logic (CRUD, Service, Auth, Migration, Backup)
│   └── tracing/         # OpenTelemetry setup
├── pkg/
│   ├── logger/          # Zap structured logging wrapper
│   └── response/        # Standardized API responses
├── tests/               # Unit, integration, and UAT tests
├── Makefile             # Standardized tasks
├── docker-compose.yml
├── go.mod
└── README.md
```

---

## 4. CRUD Engine Implementation

Implemented primarily in `internal/services/dynamic_data_service.go`, the CRUD Engine handles on-the-fly SQL generation and execution based on user-defined schemas.

*   **Create/Update/Delete:** Translates generic JSON payloads into direct SQL execution (e.g., `INSERT INTO <dynamic_table>...`) mapped to the target database connection.
*   **Read & List:** Generates `SELECT` queries utilizing standard GORM operations or raw queries.
*   **Query Engine:** The `ListDataRequest` struct parses URL query parameters (`?email=john@example.com&sort=created_at:desc&page=1&limit=20`) to dynamically apply `WHERE` clauses, `ORDER BY`, offsets, and limits.
*   **Relation Engine:** Handles `JOIN` clauses automatically when `joins` query parameters are provided, navigating the mappings defined in the `fields` metadata table.

---

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`.

*   **Safety & Tracking:** Records all migration actions in the `migrations` metadata table with statuses (`pending`, `applied`, `failed`).
*   **Operations:** Leverages GORM's `.Migrator()` interface for safe execution of schema diffs (Adding, Dropping, or Renaming columns/tables) when a `Service` definition is updated via the CMS Control Plane.
*   **Snapshots:** Stores schema snapshots before modifications to enable potential rollback capabilities.

---

## 6. GraphQL Gateway

Implemented in `internal/graphql/gateway.go`.

*   **Library:** Utilizes `github.com/99designs/gqlgen`.
*   **Dynamic Schema:** Implements the `graphql.ExecutableSchema` interface dynamically. Currently generates an AST at runtime based on defined services to allow querying of arbitrary, user-defined data structures without recompilation.
*   **Endpoint:** Exposed at `/api/v1/graphql` alongside an integrated GraphQL Playground endpoint for developer testing.

---

## 7. Backup System

Managed by `internal/services/backup_service.go` and `internal/handlers/backup_handler.go`.

*   **Snapshots:** Generates point-in-time JSON snapshots or schema backups of user-defined services.
*   **Storage:** Backup records are tracked in the metadata `backups` table, detailing row counts and tracking status.
*   **Restore:** Supports parsing JSON snapshot payloads and restoring records directly to the dynamic service tables.

---

## 8. Observability Integration

*   **Logging:** `pkg/logger/` wraps `go.uber.org/zap` for highly performant structured JSON logging. Logs include contextual information like request IDs and correlation IDs.
*   **Metrics:** `internal/middleware/prometheus.go` implements an interceptor to track HTTP request latencies, counts, and status codes. Exposed for scraping at `/metrics`.
*   **Tracing:** `internal/tracing/` configures OpenTelemetry and Jaeger to provide distributed tracing across network calls and database queries.

---

## 9. Security & Performance

*   **Security:** JWT-based authentication (`internal/middleware/auth.go`), rate limiting middlewares, SQL injection protection through parameterized queries (`DynamicDataService`), and robust field/row-level permissions.
*   **Performance:** Implements connection pooling configurable per-database via the `DatabaseConnectionManager`, enabling optimal resource utilization when querying multiple distinct database backends.

---

## 10. Unit Tests & Verification

The project includes a comprehensive test suite covering Repositories, Services, and API Handlers.

*   **Coverage:** Minimum 80% coverage enforced.
*   **Execution:** Run using `make test-coverage` or `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`.
*   **Setup:** Tests utilize an isolated, in-memory SQLite database (`:memory:`) to guarantee speed and stability.

---

## 11. Docker Setup & Deployment

The platform is containerized for easy deployment.

*   `Dockerfile`: Multi-stage build process generating a lightweight Alpine-based Go binary.
*   `docker-compose.yml`: Spins up the backend service alongside required infrastructure components (e.g., PostgreSQL for metadata, Redis, Prometheus, Jaeger).

---

## 12. API Documentation (Swagger)

API documentation is generated using `swaggo/swag`.

*   **Generation:** Run `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
*   **Access:** Exposed natively within the Go binary, typically accessible at `/swagger/index.html` (depending on router configuration).