# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade, self-hosted Backend-as-a-Service (BaaS) platform written natively in Go. Acting as a backend infrastructure generator, it dynamically builds REST and optional GraphQL APIs on top of connected databases while ensuring observability, scalability, and robust security through Role-Based Access Control (RBAC).

## 1. System Architecture Explanation

The platform adheres to **Clean Architecture** and **Domain-Driven Design (DDD)** principles to separate concerns, improve testability, and simplify scaling:

- **Presentation Layer**: Implemented using the Gin web framework (`internal/handlers/`). Acts as the API gateway mapping requests to underlying internal services. Includes GraphQL endpoints powered by `gqlgen`.
- **Application Layer**: Contains business logic (`internal/services/`). Services orchestrate schema creation, data manipulation, validation, and permissions management.
- **Domain Layer**: Defines core platform entities (`internal/models/`) like `Service`, `Field`, `DatabaseConnection`, `User`, and `Role`.
- **Infrastructure Layer**: Cross-cutting capabilities, including connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and structured logging (`pkg/logger/`).

### High Level Diagram

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

## 2. Metadata Database Schema

The CMS manages platform definitions via a Metadata Database (typically PostgreSQL, though testing uses SQLite). Core tables include:

- **`database_connections`**: Stores external database settings (id, name, type, host, port, credentials).
- **`services`**: Represents dynamic data models created by users (id, name, database_connection_id, db_table_name, created_at, updated_at).
- **`fields`**: Defines attributes for each service (name, type, nullable, unique, default_value, index). Supports types: string, integer, float, boolean, uuid, json, array, timestamp.
- **`service_permissions`**: Ties `Role` to `Service` for granular API access (CanCreate, CanRead, etc.).
- **`migrations`**: Audits applied schema changes (add, drop, rename) to allow safe schema evolution and rollbacks.
- **`backups`**: Records snapshot information and schema outputs.
- **`users` / `roles`**: Identity and RBAC data for the control plane.
- **`audit_logs`**: Logs configuration mutations and sensitive data accesses.

## 3. Go Project Structure

The platform uses a standard modular Go layout:

```text
.
├── cmd/
│   └── server/          # Application entry point (main.go)
├── docker/              # Dockerfile and compose configurations
├── internal/
│   ├── config/          # Environment/configuration management
│   ├── database/        # Database connection pool manager
│   ├── graphql/         # GraphQL gateway (gqlgen)
│   ├── handlers/        # Gin HTTP route handlers
│   ├── middleware/      # Auth, CORS, rate-limiting, metrics
│   ├── models/          # Core domain models (GORM definitions)
│   ├── services/        # Application business logic (ServiceBuilder, CRUDEngine, etc)
│   └── tracing/         # OpenTelemetry Jaeger configuration
├── pkg/
│   ├── logger/          # Structured Zap logging wrapper
│   └── response/        # Standardized Gin API response formatters
├── tests/               # Comprehensive test suites
│   ├── unit/            # Isolated logic tests
│   ├── integration/     # Database-reliant tests
│   └── uat/             # User Acceptance / E2E tests
├── Makefile             # Development automation tasks
└── go.mod               # Go dependencies
```

## 4. CRUD Engine Implementation

The Dynamic CRUD Engine (`internal/services/dynamic_data_service.go`) intercepts generic routes (`/api/v1/data/{slug}`) and translates them into parameterized GORM database queries.

- Operations generated automatically include: Create, Read, Update, Delete, and List.
- The **Query Engine** interprets URL query parameters:
  - **Filtering**: `?email=john@example.com`
  - **Sorting**: `?sort=created_at:desc`
  - **Pagination**: `?page=1&limit=20` (Handled centrally with secure row-level constraints).
- N+1 query optimization is enforced by using bulk insert routines (`db.CreateInBatches`) where applicable.

## 5. Schema Migration Engine

Managed inside `internal/services/migration_service.go`, the Schema Migration Engine safely evolves the target database schema based on Service Field configurations.

- **Capabilities**: Dynamically add, rename, drop, or alter column definitions using GORM’s `Migrator()`.
- **Safety**: Changes are natively audited within the `migrations` table, providing metadata tracking necessary for rolling back unsupported schema discrepancies.
- Migrations trigger on-the-fly during service definitions or modifications.

## 6. GraphQL Gateway

The GraphQL API is automatically provisioned via `github.com/99designs/gqlgen`. It mounts a unified GraphQL handler (`internal/graphql/gateway.go`) to `/api/v1/graphql` and a playground to `/api/v1/graphql/playground`.

- The platform translates defined standard Services into a single executable GraphQL Schema.
- Automatically handles nested associations for queries and mutations.

## 7. Backup System

The Backup Engine (`internal/services/backup_service.go`) guarantees platform resilience:

- **Endpoints**: `POST /api/v1/cms/backup/service/{service_id}` generates a comprehensive JSON data snapshot based on a dynamic table's contents.
- **Restoration**: `POST /api/v1/cms/restore/{id}` decodes and reinstates the backup payload.
- Ensures all generated schema artifacts maintain historical parity.

## 8. Observability Integration

Comprehensive observability prevents blind spots in production:

- **Logging**: `pkg/logger/` utilizes Zap for structured JSON logging. Request IDs and correlation IDs are attached contextually.
- **Metrics**: `internal/middleware/prometheus.go` records request latency, slow queries, and HTTP error rates using Prometheus (`GET /metrics`).
- **Tracing**: OpenTelemetry configures Jaeger tracing (`internal/tracing/`) to capture deep network and SQL call stacks.

## 9. Unit Tests

Testing requires a minimum 80% coverage across the Repository, Service, and API Handler layers. Tests are run isolated in an in-memory SQLite database (`:memory:`).

**Run standard tests:**
```bash
make test
```

**Run unit tests with full codebase coverage:**
```bash
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## 10. Docker Setup

A robust, production-ready environment setup.

**Start the full stack (App + Databases + Metrics/Tracing):**
```bash
make docker-up
# Equivalent to: docker-compose up -d
```

**Containers configured in `docker-compose.yml` include:**
- API Platform Container (`Dockerfile`)
- PostgreSQL Database
- Redis (Caching)
- Prometheus & Jaeger (Metrics & Tracing)

## 11. Swagger Documentation

The platform auto-generates REST API OpenAPI specification endpoints.

**Generate documentation:**
```bash
make swagger
# Or manually: swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

**View documentation:**
After starting the server, access the UI at `http://localhost:8080/swagger/index.html`.
