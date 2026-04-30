# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade, self-hosted Backend-as-a-Service platform written in Go. It enables developers to dynamically create backend services, schemas, REST APIs, and optional GraphQL endpoints on the fly. Designed with cloud-readiness in mind, it is fully containerized and includes features like dynamic schema migration, connection management, role-based access control, and state-of-the-art observability.

## 1. System Architecture Explanation

The platform is designed following **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

*   **Presentation Layer**: Composed of Gin HTTP handlers serving REST APIs and a GraphQL gateway powered by `gqlgen`. This layer handles incoming requests, standardizes responses, and routes calls to appropriate services.
*   **Application Layer**: Contains business logic encapsulated within services (`UserService`, `DynamicDataService`, `MigrationService`, etc.). This layer coordinates actions across databases, performs business validations, and dictates the core functionality of the CMS.
*   **Domain Layer**: Defines the core entities (Models) like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`, and `Menu`.
*   **Infrastructure Layer**: Handles external concerns such as managing generic database connectors (`ConnectionManager`), logging (Zap), metrics (Prometheus), and distributed tracing (OpenTelemetry/Jaeger).

### High-Level Components
*   **CMS Control Plane**: Exposes APIs under `/api/v1/cms/` to manage system metadata, connections, and service definitions.
*   **Service Builder**: Enables dynamic generation of data models and fields.
*   **CRUD Engine**: Intercepts requests for dynamic endpoints, validating RBAC and converting REST operations into raw SQL actions.
*   **Query Engine**: Parses URL query parameters for advanced filtering, sorting, pagination, and joining records.
*   **Database Connectors**: Connects to the primary metadata database as well as dynamically configured target databases (PostgreSQL, MySQL, MongoDB, SQLite).

## 2. Metadata Database Schema

The core configuration and platform states are stored in a Metadata Database. The schema includes the following key entities:

*   **`database_connections`**: Stores external database configuration (id, name, type, host, port, username, password, ssl_mode) along with connection pooling metrics.
*   **`services`**: Defines user-created data models (id, name, database_connection_id, db_table_name, created_at, updated_at).
*   **`fields`**: Represents columns/attributes within a service (id, service_id, name, type, is_nullable, is_unique, default_value).
*   **`service_permissions`**: Maps fine-grained access (CanCreate, CanRead, CanUpdate, CanDelete) to roles for a specific service.
*   **`users` & `roles`**: Base tables for authentication and authorization.
*   **`migrations`**: Audit log of all DDL schema changes executed through the platform.
*   **`backups`**: Records snapshot states for rollback purposes.
*   **`audit_logs`**: Tracks major system-level events and data manipulation.

## 3. Go Project Structure

The project adopts a clean and modular structure:

```text
├── cmd/
│   └── server/             # Application entry point (main.go)
├── internal/
│   ├── config/             # Environment & configuration loading
│   ├── database/           # Database Connection Manager (PostgreSQL, MySQL, SQLite, MongoDB)
│   ├── graphql/            # Optional GraphQL Gateway using gqlgen
│   ├── handlers/           # Gin REST API Controllers
│   ├── middleware/         # Security, Auth, Rate Limiting, Metrics
│   ├── models/             # Domain entities (Service, Field, User, Role, etc.)
│   ├── services/           # Core business logic (CRUD Engine, Service Builder, Migrations)
│   └── tracing/            # OpenTelemetry & Jaeger tracing setup
├── pkg/
│   ├── logger/             # Zap structured logger configuration
│   └── response/           # Standardized Gin HTTP response helpers
├── tests/
│   ├── integration/        # Inter-component testing
│   ├── uat/                # End-to-end user acceptance tests
│   └── unit/               # Unit testing of isolated logic
├── docker/                 # Prometheus configurations & related infrastructure files
├── Dockerfile              # Multi-stage Docker build for the application
├── docker-compose.yml      # Orchestrates Postgres, Redis, Jaeger, Prometheus, and the Backend
├── Makefile                # Standardized commands for build, run, test, and docs
└── ARCHITECTURE.md         # Detailed architectural design document
```

## 4. CRUD Engine Implementation

The Dynamic CRUD Engine is governed by the `DynamicDataService`. Once a `Service` (data model) is created in the metadata database, the platform dynamically provisions CRUD endpoints via standard routing logic (e.g., `/api/v1/data/{slug}`).

*   **Operations**: Translates REST requests into native database commands:
    *   `POST` -> Insert
    *   `GET` -> Select (Read by ID or List with Pagination)
    *   `PUT` -> Update
    *   `DELETE` -> Delete
*   **Query Capabilities**: The engine inherently supports advanced query extraction. `GET /api/v1/data/{slug}?email=test@test.com&sort=created_at:desc&page=1&limit=20` translates cleanly into native `.Where()`, `.Order()`, and `.Limit()` clauses, checking an allowed whitelist of fields.
*   **Safety**: Multi-record operations process via `db.CreateInBatches` for speed and memory efficiency, while error checks are enforced rigorously.

## 5. Schema Migration Engine

Handled by the `MigrationService`, the platform can translate abstract schema configurations from the CMS into actual DDL operations on targeted databases.

*   **Capabilities**: Allows the dynamic addition, modification, or dropping of columns on connected databases.
*   **Tracking**: Utilizing GORM's `Migrator()`, it applies structural shifts safely. All executed changes are natively logged within the `migrations` table for a complete audit trail.
*   **Safety Features**: Enables automatic database snapshots and structured rollbacks if modifications fail or require reversion.

## 6. GraphQL Gateway

The platform provides a generic GraphQL API alongside standard REST endpoints, allowing clients flexibility in querying dynamic datasets.

*   **Gateway Location**: Served under `/api/v1/graphql` and managed by `internal/graphql/gateway.go`.
*   **Implementation**: Powered by `gqlgen`, the GraphQL Executable Schema automatically interprets models generated by the Service Builder.
*   **Playground**: Available under `/api/v1/graphql/playground` for interactive querying and schema exploration.

## 7. Backup System

The `BackupService` provides system operators with the ability to take immediate snapshots of service states.

*   **Mechanism**: Accessible via `POST /api/v1/cms/backup/service/{id}`, it exports the state of a dynamically generated table.
*   **Formats**: The output can be stored as serialized JSON.
*   **Restoration**: A corresponding `/api/v1/cms/restore/{id}` endpoint decodes snapshot records back into their active database tables, assisting in disaster recovery and rollback scenarios.

## 8. Observability Integration

Comprehensive tracking is natively built into the platform:

*   **Logging**: Built using **Zap**, logs are structured (JSON) and include trace information, allowing detailed analysis of Request/Response cycles and errors.
*   **Metrics**: Integrated natively with **Prometheus**. A global middleware tracks request counts, status codes, and latency distributions, exposing them over a `/metrics` route.
*   **Tracing**: Instrumented heavily with **OpenTelemetry**. Traces cover HTTP requests and cross-boundary SQL actions. Captured telemetry is sent to an embedded **Jaeger** instance for visualization of internal system latencies.

## 9. Unit Tests

Robust testing practices ensure system stability. The project requires a minimum of 80% coverage across Repository, Service, and Handler layers.

*   **Testing Strategy**: Uses an in-memory SQLite database setup (DSN: `:memory:`) initialized manually via `database.TestConnection` to avoid side effects.
*   **Tools**: Native Go `testing` library assertions are prioritized over external dependencies to minimize test execution complexity and download time constraints in CI/CD.
*   **Execution**:
    *   Tests: `go test -v -race ./tests/...`
    *   Coverage analysis: `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`

## 10. Docker Setup

The platform is designed to be fully containerized.

*   **Dockerfile**: A multi-stage build starting with `golang:alpine` builds the Go binary (stripping debugging info via `ldflags`) and transfers it to an isolated, secure final alpine image for distribution.
*   **Docker Compose**: The `docker-compose.yml` provisions the entire ecosystem seamlessly:
    *   The platform backend (`cms-backend`)
    *   Metadata Database (`postgres`)
    *   Caching Store (`redis`)
    *   Observability Stack (`jaeger`, `prometheus`)

## 11. Swagger Documentation

Extensive OpenAPI schemas are automatically derived via code annotations.

*   **Engine**: We use `swaggo/swag` and `gin-swagger`.
*   **Generation**: The standard documentation is generated using `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
*   **Exposure**: Endpoints are natively documented, allowing consumers to interactively explore available REST API routes via `/swagger/index.html`.
