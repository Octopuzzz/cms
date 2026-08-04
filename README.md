# Backend Platform (Dynamic CMS + API Builder)

Welcome to the Backend-as-a-Service (BaaS) platform! This system is a fully self-hosted, Go-native platform that dynamically generates REST and GraphQL APIs, designed as a production-grade infrastructure generator. It works similarly to platforms like Hasura or Supabase.

The platform allows developers to:
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

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles to ensure it is modular, scalable, and cloud-ready.

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
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

---

## 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**, which can be backed by PostgreSQL. The tables store platform metadata including databases, services, fields, relations, migrations, and backups.

Key tables defined in `internal/models/models.go`:
- `database_connections`: Stores external DB configurations (id, name, type, host, port, credentials). Includes connection pooling settings.
- `services`: Represents user-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
- `fields`: Defines attributes for each service (id, name, type, nullable, unique, default_value, index). Configures uniqueness, nullability, defaults.
- `service_permissions`: Connects `Role` to `Service`, providing fine-grained access checks (e.g. CanCreate, CanRead).
- `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- `backups`: Stores snapshot records or schema outputs.
- `users` and `roles`: General authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of structural and data-level modifications.

---

## 3. Go Project Structure

The project uses a clean modular structure, separating the entry point, internal business logic, shared packages, testing, and deployment configurations.

```
.
├── cmd/
│   └── server/               # Application entry point
├── internal/
│   ├── api/                  # API routes and gateway
│   ├── config/               # Configuration management
│   ├── database/             # Database connection managers
│   ├── graphql/              # GraphQL gateway
│   ├── handlers/             # HTTP Handlers / Controllers
│   ├── middleware/           # Auth, Prometheus, Rate Limiter, OpenTelemetry
│   ├── models/               # Domain models (Metadata schemas)
│   ├── services/             # Application business logic (CMS, CRUD, Migration, Backup)
│   └── tracing/              # OpenTelemetry and Tracing implementation
├── pkg/
│   ├── logger/               # Zap structured logging
│   └── response/             # Standardized HTTP responses
├── tests/
│   ├── unit/                 # Unit tests
│   ├── integration/          # Integration tests
│   └── uat/                  # User Acceptance Tests
├── docker/
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── prometheus.yml
├── ARCHITECTURE.md           # Internal architectural documentation
├── Makefile                  # Build, run, and test targets
└── README.md                 # Project documentation
```

---

## 4. CRUD Engine Implementation

The Dynamic CRUD Engine is managed by `internal/services/dynamic_data_service.go`. When a service (data model) is created, the system automatically generates CRUD endpoints.

- **Automated Operations**: Exposes Create, Read, Update, Delete, and List endpoints dynamically (e.g., `POST /api/v1/data/{slug}`, `GET /api/v1/data/{slug}`).
- **Query Engine**: Maps endpoints to standard GORM database operations on the fly.
  - **Filtering**: `GET /api/v1/data/{slug}?email=john@example.com`
  - **Sorting**: `GET /api/v1/data/{slug}?sort=created_at:desc`
  - **Pagination**: `GET /api/v1/data/{slug}?page=1&limit=20`
  - **Join Queries**: Dynamic joining via foreign key mappings.
- **Security**: Verifies Role-Based Access Control and Row-Level filtering dynamically prior to executing queries.

---

## 5. Schema Migration Engine

The Schema Migration Engine is managed by `internal/services/migration_service.go` and is responsible for safe schema changes.

- **Capabilities**: Tracks and applies safe schema diffs (Add column, Drop column, Rename column, Change column type) using GORM’s `.Migrator()`.
- **Safety Features**:
  - Logs migration status natively into the `migrations` table.
  - Supports automatic backups before migrations.
  - Handles rollbacks through snapshot retention logic.

---

## 6. GraphQL Gateway

The system provides an optional GraphQL gateway managed by `internal/graphql/gateway.go`.

- **Auto-Generation**: Uses `github.com/99designs/gqlgen/graphql` to automatically generate optional schemas based on the dynamic service metadata.
- **Features**: Exposes `POST /api/v1/graphql` to fulfill automated GraphQL queries and mutations for defined data models.
- **Relationships**: Automatically resolves relationships between different dynamic services.

---

## 7. Backup System

The Backup Engine is managed by `internal/services/backup_service.go` and ensures data safety.

- **Supported Operations**: Table backup, schema backup, and service snapshots.
- **Format**: Generates generic service record backup features mapping dynamic table contents to snapshots (JSON payloads).
- **Restore**: Supports restoring rows via JSON payload decoding.

---

## 8. Observability Integration

The platform includes state-of-the-art observability and monitoring:

- **Logging**: Zap structured logging is injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context. Request, error, and query logs are collected.
- **Tracing**: OpenTelemetry (`internal/tracing/`) is initialized to wrap SQL commands and network logic.
- **Metrics**: Prometheus (`internal/middleware/prometheus.go`) tracks request latency, slow queries, error rates, and HTTP status codes via standard HTTP interceptors. Exported at `/metrics`.

---

## 9. Unit Tests

The system enforces rigorous testing standards using the standard Go `testing` library.

- **Test Areas**: Coverage focuses on the Repository, Service, and API handlers.
- **Database Testing**: Unit tests use an in-memory SQLite database setup (`:memory:`) via the `github.com/glebarez/sqlite` driver and standard GORM configuration for database mocking.
- **Commands**:
  - `make test-unit` for isolated tests.
  - `make test-coverage` or `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...` for full test execution across all packages.
- **Requirement**: Minimum of 80% coverage is strictly required.

---

## 10. Docker Setup

The platform is fully containerized and ready for deployment using Docker.

- **Dockerfile**: Optimized multi-stage builds.
- **Docker Compose**: The `docker/docker-compose.yml` file sets up the complete environment including the Go application, databases, Redis (optional for caching), and observability tools (Prometheus).
- **Execution**: Run `docker-compose -f docker/docker-compose.yml up` to launch the entire stack.

---

## 11. Swagger Documentation

API Documentation is auto-generated using Swagger/OpenAPI.

- **Generation**: Run `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs` to generate the latest docs. (Requires `swag` CLI: `go install github.com/swaggo/swag/cmd/swag@latest`).
- **Structure**: The generated `docs/` directory is explicitly ignored by version control to avoid bloat.
- **Viewing**: The documentation is exposed via a Swagger UI endpoint within the application.
