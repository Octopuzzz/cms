# CMS Backend Platform

A production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. This platform acts as a backend infrastructure generator, allowing users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints. It is fully self-hosted, modular, scalable, and cloud-ready.

---

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

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

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

---

## 2. Metadata Database Schema

The CMS stores platform metadata in a core database (PostgreSQL by default). Key tables include:

- **`database_connections`**: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
- **`services`**: Represents user-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
- **`fields`**: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults, relations, and validations.
- **`service_permissions`**: Connects `Role` to `Service` providing fine-grained access checks (e.g., CanCreate, CanRead). Contains row-level and field-level capabilities.
- **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- **`backups`**: Stores snapshot records or schema outputs.
- **`users`** and **`roles`**: General authentication and authorization for the control plane.
- **`audit_logs`**: Detailed logging of structural and data-level modifications.

---

## 3. Go Project Structure

The project uses a clean modular structure:

```text
.
├── cmd
│   └── server          # Entry point for the application (main.go)
├── docker              # Docker and containerization configuration (Dockerfile, docker-compose.yml)
├── internal
│   ├── config          # Application configuration mapping
│   ├── database        # Database connection manager (PostgreSQL, MySQL, MongoDB, SQLite)
│   ├── graphql         # GraphQL gateway generation using gqlgen
│   ├── handlers        # HTTP presentation layer (Gin handlers)
│   ├── middleware      # Auth, rate limiting, metrics, CORS, tracing
│   ├── models          # Domain models (Metadata database schema)
│   ├── services        # Application business logic (Service Builder, CRUD Engine, etc.)
│   └── tracing         # OpenTelemetry setup
├── pkg
│   ├── logger          # Zap structured logging
│   └── response        # Standardized API response formatters
└── tests
    ├── integration     # Integration tests
    ├── uat             # User Acceptance Testing
    └── unit            # Unit tests with high coverage
```

---

## 4. Components Documentation

### CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go` and exposed via `/api/v1/data/{slug}`.
- Maps endpoints to standard GORM database operations dynamically.
- Automatically generates CRUD operations (Create, Read, Update, Delete, List) for registered services.
- **Query Engine Features**: Supports advanced filtering (`?email=test@example.com`), sorting, and pagination mapping. It handles RBAC validation before executing dynamic SQL queries.

### Schema Migration Engine
Managed by `internal/services/migration_service.go` and exposed via `/api/v1/cms/migrations`.
- Safely applies schema modifications (Add, Drop, Rename Column, Modify Type) via GORM’s Migrator.
- Logs execution history natively into the `migrations` table.
- Supports Rollback capabilities through snapshot retention.

### GraphQL Gateway
Managed by `internal/graphql/gateway.go` and exposed via `/api/v1/graphql`.
- Automatically generates generic GraphQL schemas and endpoints using `github.com/99designs/gqlgen/graphql`.
- Fulfills GraphQL queries and mutations mapping to the dynamically created data models.
- Includes a Playground endpoint for query testing.

### Backup System
Managed by `internal/services/backup_service.go` and exposed via `/api/v1/cms/backup/service/{service_id}`.
- Generates data snapshots (table backups, schema backups, service snapshots).
- Uses a generic JSON payload approach to store large dataset backups into the metadata database.
- Implements `RestoreBackup` functionality to repopulate data dynamically from a saved JSON snapshot.

### Observability Integration
Observability is integrated comprehensively across the system:
- **Logging (Zap)**: Structured JSON logging configured in `pkg/logger/`. Captures request, error, and query logs.
- **Metrics (Prometheus)**: `internal/middleware/prometheus.go` tracks request latency, errors, and rates via standard HTTP interceptors. Exported at `/metrics`.
- **Tracing (OpenTelemetry)**: Managed in `internal/tracing/` to track and propagate traces globally via Jaeger.

### Security & Performance
- **Security**: JWT Authentication (`internal/middleware/auth.go`), Role-Based Access Control (RBAC), API Rate Limiting, and CORS management.
- **Performance**: Robust Database Connection Pooling (`internal/database/connection_manager.go`) to optimize multiple simultaneous database accesses.

---

## 5. Unit Testing
The project guarantees high test coverage across Repository, Service, and API handler layers.
- Uses standard Go testing tools and in-memory SQLite (`:memory:`) to emulate database operations efficiently.
- Run tests via the Makefile:
  ```bash
  make test-coverage
  # or manually:
  go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
  ```

---

## 6. Docker Setup
A complete Docker setup is provided for production-ready deployment:
- **`Dockerfile`**: Compiles the Go binary and builds a lean Alpine Linux container.
- **`docker-compose.yml`**: Spins up the entire stack, including the CMS Backend, PostgreSQL metadata database, Redis (optional), Prometheus, Jaeger, and other microservices.
  ```bash
  docker-compose up --build
  ```

---

## 7. Swagger Documentation
API Documentation is auto-generated using standard OpenAPI definitions.
- Viewable at `/api/v1/swagger/index.html` on the running application.
- To regenerate docs when modifying API structures:
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
  ```
