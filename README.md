# Go Backend Platform (Dynamic CMS + API Builder)

A production-grade Backend-as-a-Service (BaaS) platform built in Go. It allows developers to dynamically create backend services, schemas, and APIs (REST and optional GraphQL) while offering schema migration, backups, monitoring, and database management—all in a fully self-hosted, modular, and cloud-ready environment.

---

## 1. System Architecture Explanation

The platform is designed using **Clean Architecture** and **Domain-Driven Design (DDD)** principles to separate concerns, improve maintainability, and ensure testability.

### High-Level Architecture

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
1. **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
2. **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
3. **Domain Layer**: Core data models (`internal/models/`) defining entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
4. **Infrastructure Layer**: Cross-cutting tools including connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

---

## 2. Metadata Database Schema

The core metadata configuration is stored in a relational database (default PostgreSQL). The key tables manage the configuration of the dynamic platform:

- **`database_connections`**: Stores external DB configurations (`id`, `name`, `type`, `host`, `port`, `username`, `password`, `database_name`). Handles connection pooling settings.
- **`services`**: Represents user-created data models, linking to a `database_connection_id` and tracking dynamic schemas (`db_table_name`). Fields include `id`, `name`, `database_id`, `created_at`, `updated_at`.
- **`fields`**: Defines attributes for each service (types: string, integer, float, boolean, uuid, json, array, timestamp). Configures `name`, `type`, `nullable`, `unique`, `default_value`, `index`.
- **`service_permissions` / `relations`**: Connects roles to services for access checks and manages dynamic table relationships (One-to-One, One-to-Many, Many-to-One, Many-to-Many).
- **`migrations`**: Tracks DDL executions with metadata required for rollbacks or history tracking.
- **`backups`**: Stores snapshot records and schema outputs.
- **`users` & `roles`**: General authentication and authorization for the CMS control plane.

---

## 3. Go Project Structure

The codebase is organized in a highly modular way to support scalability:

```text
.
├── cmd
│   └── server          # Application entry point (main.go)
├── internal
│   ├── handlers        # API Gateway (REST & GraphQL mapping)
│   ├── services        # Core engines (CMS, CRUD, Migration, Backup)
│   ├── models          # Domain layer and Database schema models
│   ├── database        # Database Connection Manager
│   ├── graphql         # Auto-generated GraphQL Gateway
│   ├── middleware      # Auth, Rate limiting, OpenTelemetry, Prometheus
│   ├── config          # Application configuration mapping
│   └── tracing         # OpenTelemetry & Jaeger instrumentation
├── pkg
│   ├── logger          # Zap structured logging wrapper
│   └── response        # Standardized API response handlers
├── tests
│   ├── unit            # Unit tests for services and models
│   └── integration     # API and integration testing
├── docker              # Docker container configuration
├── Makefile            # Automation for builds, tests, running
├── Dockerfile          # Docker build instructions
└── README.md           # Project documentation
```

---

## 4. CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`, the **Dynamic CRUD & Query Engine** dynamically maps API endpoints to underlying database queries.

- **Automated Generation**: When a service is defined via the CMS control plane, the system dynamically exposes CRUD operations (Create, Read, Update, Delete, List).
  - `POST /api/v1/data/{slug}`
  - `GET /api/v1/data/{slug}/{id}`
  - `PUT /api/v1/data/{slug}/{id}`
  - `DELETE /api/v1/data/{slug}/{id}`
  - `GET /api/v1/data/{slug}`
- **Query Engine Features**: Supports filtering (`?email=john@example.com`), sorting (`?sort=created_at:desc`), pagination (`?page=1&limit=20`), and foreign-key join mapping (`?join=user,products`).
- **Data Validation & Security**: Applies Role-Based Access Control and Row-Level filtering dynamically prior to executing queries. Uses parameterized queries to prevent SQL injection.

---

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`, ensuring safe application of schema changes to external databases.

- **Supported Operations**: Add column, drop column, rename column, change column type.
- **Safety Mechanisms**: Leverages GORM's `.Migrator()`. Tracks executed migrations in the metadata database for status and rollback capabilities. Enables seamless schema updates as users modify service fields in the CMS.

---

## 6. GraphQL Gateway (Optional)

Managed by `internal/graphql/gateway.go`, acting as an alternative presentation layer.

- **Dynamic Schema Generation**: Utilizes `gqlgen` (`github.com/99designs/gqlgen/graphql`) to generate GraphQL schemas directly from defined services.
- **Endpoint**: Exposes `POST /api/v1/graphql`.
- **Capabilities**: Supports complex queries, mutations, and deep relational fetching using dynamically compiled resolvers.

---

## 7. Backup System

Managed by `internal/services/backup_service.go`, securing both data and schema.

- **Snapshots**: Generates snapshot records mapping dynamic table contents to JSON payloads.
- **Restoration**: Decodes JSON payloads to restore service rows.
- **Endpoints**:
  - Backup: `POST /api/v1/cms/backup/service/{service_id}`
  - Restore: `POST /api/v1/cms/restore/service/{service_id}`

---

## 8. Observability Integration

Comprehensive observability embedded throughout the application for robust monitoring:

- **Logging**: Zap structured logging integrated via `pkg/logger/`. Request logs, error logs, and query logs are enriched with correlation/trace IDs from the Context.
- **Tracing**: OpenTelemetry (Jaeger) is initialized in `internal/tracing/` to wrap SQL commands and track network/request lifecycles for latency tracking.
- **Metrics**: Prometheus instrumentation is configured in `internal/middleware/`. It exports metrics (`/metrics`) including request latency, status codes, and error rates via standard HTTP interceptors.

---

## 9. Unit Tests

Testing ensures reliability and regressions prevention with a target of **80% minimum coverage**.

- **Scope**: Repository functions, Services logic, and API Handlers.
- **Database Testing**: Uses an in-memory SQLite database setup (`github.com/glebarez/sqlite` via `DSN: :memory:`) to mock database interactions reliably.
- **Running Tests**:
  - Standard unit tests: `make test-unit`
  - Full coverage check: `make test-coverage` (executes `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`)

---

## 10. Docker Setup

The platform includes full containerization for consistent local development and cloud deployment.

- **Dockerfile**: Optimized multi-stage build ensuring a lean, secure, and production-ready Go binary (`cms-backend`).
- **Docker Compose**: Provided in `docker-compose.yml` to spin up the entire stack seamlessly, including the core API, metadata database (PostgreSQL/MySQL), Redis for caching, and observability tools (Jaeger, Prometheus).
- **Execution**: Easily start the full platform using `docker-compose up --build`.

---

## 11. Swagger Documentation

API documentation is auto-generated using Swagger/OpenAPI.

- **Generation**: Run `~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs` to compile comments into standard OpenAPI definitions.
- **Access**: The interactive Swagger UI is exposed at `/swagger/index.html` allowing developers to visually explore and test both the CMS Control Plane APIs and dynamically generated CRUD APIs.
