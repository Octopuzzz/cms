# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. It works similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native. The platform allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

The system is modular, scalable, and cloud-ready, following Clean Architecture and Domain-Driven Design (DDD).

---

## 1. System Architecture Explanation

The Backend Platform acts as an infrastructure generator and is comprised of several core layers based on Clean Architecture:

*   **Presentation Layer**: Gin HTTP Router and Handlers (`internal/handlers/`), acting as the API Gateway. Maps requests to internal services and includes the GraphQL endpoint using `gqlgen`.
*   **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, perform requested operations, and manage core mechanisms like migrations, backups, and dynamic data CRUD.
*   **Domain Layer**: Data models (`internal/models/`). Defines the system's core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`, `Migration`, and `Backup`.
*   **Infrastructure Layer**: Cross-cutting concerns including database connectivity caching, observability with OpenTelemetry tracing (`internal/tracing/`), Prometheus metrics (`internal/middleware/`), and Zap structured logging (`pkg/logger/`).

**High-Level Flow:**
Clients (REST or GraphQL API) -> API Gateway -> Backend Platform Core (CMS Control Plane, Service Builder, CRUD Engine, Query Engine, Schema Migration Engine, Backup Engine, Observability Engine) -> Database Connectors (PostgreSQL, MySQL, MongoDB, SQLite).

---

## 2. Metadata Database Schema

The core internal configuration is stored in a Metadata Database. The system supports defining dynamic services and relations. The models include:

*   `database_connections`: Stores external DB credentials and pooling settings (`id`, `name`, `type`, `host`, `port`, etc.).
*   `services`: User-created data models referencing `database_connection_id` and tracking `db_table_name`.
*   `fields`: Defines attributes for each service (type, required, uniqueness, nullability, defaults).
*   `service_permissions`: Connects `Role` to `Service` for granular RBAC (CanCreate, CanRead, etc.) and Row/Field-level security.
*   `migrations`: Records schema changes (DDL executions) and their statuses, storing schema snapshots for possible rollbacks.
*   `backups`: Snapshot records containing schema and data.
*   `users`, `roles`, `api_tokens`: Security models managing authentication and authorization.
*   `audit_logs`: Detailed tracking of structural and data changes.

---

## 3. Go Project Structure

The project uses a clean modular structure, typical for standard Go projects:

```
.
├── cmd/
│   └── server/          # Main application entry point (main.go)
├── docker/              # Docker and Docker Compose configurations, Prometheus configs
├── internal/
│   ├── config/          # Application configuration (env vars parsing)
│   ├── database/        # Database Connection Manager (GORM setup, connection pooling)
│   ├── graphql/         # GraphQL Gateway implementation (gqlgen logic)
│   ├── handlers/        # API route handlers (REST endpoints for CMS and dynamic data)
│   ├── middleware/      # Auth, Rate Limiting, Prometheus interceptors
│   ├── models/          # Core Domain models (Metadata schema)
│   ├── services/        # Application Business Logic (CMS, CRUD, Migration, Backup, User management)
│   └── tracing/         # OpenTelemetry configuration
├── pkg/
│   ├── logger/          # Zap Logger global instance
│   └── response/        # Standardized API response helpers
└── tests/
    ├── integration/     # Integration tests
    ├── uat/             # User Acceptance Tests (UAT)
    └── unit/            # Unit tests
```

---

## 4. CRUD Engine Implementation

The **Dynamic CRUD Engine** (`internal/services/dynamic_data_service.go`) automatically generates and handles REST API requests based on service definitions.
When a user defines a new `Service`, the engine routes requests to the dynamic endpoints:
*   `POST /api/v1/data/{slug}` (Create)
*   `GET /api/v1/data/{slug}/{id}` (Read)
*   `PUT /api/v1/data/{slug}/{id}` (Update)
*   `DELETE /api/v1/data/{slug}/{id}` (Delete)
*   `GET /api/v1/data/{slug}` (List with advanced filtering, sorting, pagination, and dynamic joins based on defined relations).

This engine constructs raw GORM queries safely on the fly, interpreting query params, and enforcing RBAC at the row-level and field-level dynamically.

---

## 5. Schema Migration Engine

The **Schema Migration Engine** (`internal/services/migration_service.go`) handles safe schema changes (Add Column, Drop Column, Rename, Change Type).
It tracks migration records directly into the `migrations` table in the metadata database, stores snapshots, handles rollbacks via snapshot retention, and utilizes GORM's `Migrator()` interface for executing cross-database DDL statements reliably.

---

## 6. GraphQL Gateway (Optional)

The **GraphQL Gateway** (`internal/graphql/gateway.go`) optionally translates dynamic service definitions into a GraphQL schema. It automatically structures the necessary query and mutation resolvers based on user-defined models and relations via `github.com/99designs/gqlgen`.
It exposes `POST /api/v1/graphql` to serve all GraphQL operations in a unified endpoint.

---

## 7. Backup System

The **Backup Engine** (`internal/services/backup_service.go`) manages data preservation strategies.
It supports generating structural (schema) and table data backups for specified dynamic services.
The system serializes dynamic table contents to generic JSON snapshots to support agnostic restore capabilities via `POST /cms/backup/service/{service_id}` and `POST /cms/restore/service/{service_id}`.

---

## 8. Observability Integration

The platform includes comprehensive state-of-the-art observability:
*   **Logging**: Handled globally by Zap (`pkg/logger/`), capturing request logs, errors, and traces, attaching Correlation IDs via context.
*   **Metrics**: Tracked by Prometheus through `internal/middleware/prometheus.go`, exposing latency, error rates, and request counts at the `/metrics` endpoint.
*   **Tracing**: Supported by OpenTelemetry (`internal/tracing/`), wrapping network calls and database queries to map end-to-end execution flow.

---

## 9. Unit Tests

The system maintains high code quality using a comprehensive testing suite spread across Unit, Integration, and UAT levels covering the repository, services, and API handlers. Tests utilize in-memory SQLite instances to achieve isolated environments quickly.
You can execute the entire suite and track the 80%+ coverage using:
```bash
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

---

## 10. Docker Setup

The repository is fully Dockerized for simple deployments:
*   A multi-stage `Dockerfile` handles building the Go binary optimally.
*   A `docker-compose.yml` configures the backend platform along with necessary infrastructure like PostgreSQL (if used for metadata), Redis, and observability platforms (Prometheus).

---

## 11. Swagger Documentation

API documentation is generated dynamically via Swagger (OpenAPI 3.0).
The `swag` cli reads declarative comments on the Gin handlers and models to generate `/swagger/*any` endpoint, which renders a Swagger UI containing definitions for CMS control plane endpoints and example dynamic data endpoints.

To regenerate Swagger documentation:
```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```