# Dynamic CMS & API Builder Backend Platform

A production-grade, self-hosted Backend-as-a-Service (BaaS) platform written natively in Go. This platform allows developers to dynamically construct data models, connect to external databases, and instantly generate robust CRUD REST APIs and optional GraphQL endpoints, closely mirroring the functionality of platforms like Hasura or Supabase.

Built entirely in Go, it features a scalable, modular architecture rooted in Clean Architecture and Domain-Driven Design (DDD).

---

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles to separate concerns, ensure testability, and provide high scalability:

*   **Presentation Layer (`internal/handlers/`, `internal/graphql/`)**: The API Gateway (Gin HTTP Router). This layer exposes the dynamically generated REST endpoints and GraphQL schemas (`gqlgen`). It manages HTTP communication, intercepts payloads, and passes data to internal services.
*   **Application Layer (`internal/services/`)**: Contains the core business logic. The `Service Builder`, `CRUD Engine`, `Schema Migration Engine`, and `Backup Engine` live here. They translate HTTP requests into internal operational commands (like querying dynamic tables or applying migrations).
*   **Domain Layer (`internal/models/`)**: Defines the fundamental entities of the platform (e.g., `Service`, `Field`, `DatabaseConnection`, `User`, `Role`). These models are strictly decoupled from business logic and database adapters.
*   **Infrastructure Layer (`internal/database/`, `internal/tracing/`, `pkg/logger/`)**: Manages external dependencies including database connectivity (connection pooling/GORM), structured logging (Zap), and telemetry (Prometheus, OpenTelemetry).

---

## 2. Metadata Database Schema

The core internal configuration is stored securely in a **Metadata Database**. Currently modeled in `internal/models/models.go`, the key tables are:

*   `database_connections`: Stores configurations for registered external DBs (id, name, type, host, port, credentials).
*   `services`: Represents dynamically created data models (tables). References `database_connection_id` and tracks the underlying `db_table_name`.
*   `fields`: Defines the attributes of each service (string, integer, uuid, JSON). It also tracks metadata like nullability, unique constraints, and default values.
*   `service_permissions`: Links `Role`s to `Service`s for RBAC, determining capabilities like `CanCreate` or `CanRead`.
*   `migrations`: Records executed DDL statements (like adding or dropping columns) with metadata required for auditability and rollbacks.
*   `backups`: Stores backup snapshot metadata and status.
*   `users` and `roles`: Manages platform authentication and authorization.
*   `audit_logs`: Maintains an immutable history of structural changes and data modifications.

---

## 3. Go Project Structure

The project adopts a clean, standard Go layout to maintain modularity:

```text
.
├── cmd/
│   └── server/             # Application entrypoint (main.go)
├── internal/
│   ├── config/             # Environment variable and settings management
│   ├── database/           # DB Connection Manager & GORM configuration
│   ├── graphql/            # GraphQL gateway (schema generation & resolvers)
│   ├── handlers/           # HTTP handlers (REST API controllers)
│   ├── middleware/         # Auth (JWT), Rate Limiting, Prometheus interceptors
│   ├── models/             # Domain definitions (Services, Fields, DB Conns)
│   ├── services/           # Application business logic (CRUD Engine, CMS Plane)
│   └── tracing/            # OpenTelemetry & Jaeger initialization
├── pkg/
│   ├── logger/             # Zap structured logger configuration
│   └── response/           # Standardized API JSON response helpers
├── tests/
│   ├── integration/        # DB and API integration tests
│   ├── uat/                # User Acceptance tests
│   └── unit/               # Service & Handler unit tests
├── docker/                 # Docker Compose and Prometheus configuration files
├── Dockerfile              # Multi-stage Docker build for the backend
├── Makefile                # Standardized build, test, and dev commands
└── ARCHITECTURE.md         # Detailed architectural documentation
```

---

## 4. CRUD Engine Implementation

The Dynamic CRUD Engine (`internal/services/dynamic_data_service.go`) intercepts requests to `GET /api/v1/data/{slug}` (and POST, PUT, DELETE) and translates them into live queries against the dynamic tables.

*   **Dynamic Operations**: Automatically supports Create, Read, Update, Delete, and List operations.
*   **Query Engine Integration**: Handles dynamic query parameters for sorting (e.g., `?sort=created_at:desc`), filtering (e.g., `?email=user@example.com`), and pagination (e.g., `?page=1&limit=20`).
*   **Access Control**: Evaluates `service_permissions` before executing queries, ensuring row and field-level security based on the authenticated user's role.

---

## 5. Schema Migration Engine

The platform handles real-time data modeling safely via the Schema Migration Engine (`internal/services/migration_service.go`).

*   **Safe Operations**: When a user modifies a service via the CMS plane, the engine dynamically executes DDL queries using GORM’s `.Migrator()`.
*   **Capabilities**: Supports Adding, Dropping, and Renaming columns dynamically on registered external databases.
*   **Auditability**: Every schema change is logged directly into the `migrations` metadata table, facilitating future rollback and auditing workflows.

---

## 6. GraphQL Gateway

A unified, automatically generated GraphQL endpoint provides developers with an alternative to REST.

*   **Implementation (`internal/graphql/gateway.go`)**: Powered by `gqlgen`, it dynamically constructs queries and mutations based on the user-defined `services` and `fields`.
*   **Endpoint**: Exposes `POST /api/v1/graphql` to fulfill queries seamlessly, leveraging the underlying dynamic CRUD engine to retrieve requested relationships and deeply nested fields.

---

## 7. Backup System

The Backup Engine (`internal/services/backup_service.go`) ensures data durability and snapshotting.

*   **Snapshots**: Provides generic snapshot features capable of mapping dynamic table data to restorable formats (e.g., JSON snapshots).
*   **Management**: Backup actions and results are recorded in the `backups` table within the metadata DB. Restoration routines process the stored snapshot payloads back into live data.

---

## 8. Observability Integration

Production-readiness is guaranteed through deep, systemic observability:

*   **Logging**: Implemented using **Uber's Zap logger** (`pkg/logger/`). Request, error, and system logs are output in structured JSON formats. Context extraction allows correlation/trace IDs to be stitched into every log line.
*   **Metrics**: Handled via **Prometheus** interceptors (`internal/middleware/prometheus.go`). Endpoints automatically track request latencies, status codes, and error rates, exposing them on the `/metrics` endpoint for scraping.
*   **Tracing**: Utilizing **OpenTelemetry** (`internal/tracing/`), deep request flows (including downstream DB transactions) are traced and exportable to tools like **Jaeger**.

---

## 9. Unit Tests

Testing is prioritized, targeting an 80%+ coverage metric across Handlers, Services, and Repositories.

*   **Execution**: Tests are run using `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`
*   **Structure**: Uses in-memory SQLite instances (`github.com/glebarez/sqlite`) to rapidly test service logic, validations, and dynamic endpoints without requiring heavy external dependencies. Includes integration and UAT tests.

---

## 10. Docker Setup

The repository is built for instant containerization:

*   **Multi-stage `Dockerfile`**: Compiles the Go binary securely in an Alpine environment, reducing the final image size and attack surface.
*   **`docker-compose.yml`**: Allows users to instantly spin up the backend platform alongside supportive infrastructure like Prometheus and Jaeger.

---

## 11. Swagger Documentation

API Documentation is auto-generated using standard Swaggo comments scattered throughout the presentation layer (`internal/handlers/`).

*   **Command**: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`
*   **Integration**: Integrated tightly into the Gin router, rendering the OpenAPI spec dynamically for interactive developer exploration.
