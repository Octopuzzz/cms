# Dynamic CMS + API Builder Backend Platform

This repository contains a **production-grade Backend Platform (Dynamic CMS + API Builder)** written in Go. The system functions similarly to platforms like Hasura or Supabase but is **fully self-hosted and Go-native**.

The platform enables developers to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints. It is modular, scalable, and cloud-ready, following **Clean Architecture** and **Domain-Driven Design (DDD)**.

---

## 1. High Level Platform Architecture

The system acts as a backend infrastructure generator, sitting behind an API Gateway and translating HTTP/GraphQL requests into dynamic database operations.

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

The CMS stores platform metadata in a primary relational database (e.g., PostgreSQL or SQLite for local development). The metadata tables manage the dynamic services and configurations.

- **`database_connections`**: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
- **`services`**: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
- **`fields`**: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
- **`service_permissions`**: Connects `Role` to `Service` providing fine-grained access checks (e.g., CanCreate, CanRead). Contains row-level and field-level capabilities.
- **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- **`backups`**: Stores snapshot records or schema outputs.
- **`users`** and **`roles`**: General authentication and authorization for the control plane.
- **`audit_logs`**: Detailed logging of structural and data-level modifications.

---

## 3. Go Project Structure

The project follows standard Go layout patterns combined with Clean Architecture principles:

```text
cmd/
└── server/             # Application entry point (main.go)
internal/
├── api/                # API definition interfaces
├── config/             # Environment and configuration loading
├── database/           # Database connection manager (PostgreSQL, MySQL, MongoDB, SQLite)
├── graphql/            # GraphQL gateway generation (gqlgen)
├── handlers/           # Gin HTTP REST Handlers
├── middleware/         # Auth, rate limiting, and observability middleware
├── models/             # Domain layer models and metadata schemas
├── services/           # Application business logic (CMS Control Plane, Query Engine)
└── tracing/            # OpenTelemetry integration
pkg/
├── logger/             # Zap structured logger wrapper
└── response/           # Standardized API response payloads
tests/
├── integration/        # Integration tests
├── uat/                # User Acceptance Testing
└── unit/               # Unit tests
docker/                 # Docker Compose and external service configs
```

---

## 4. System Components

### CMS Control Plane & Service Builder
Managed under `/api/v1/cms/` routes (`internal/services/service_service.go`, `internal/services/dbconn_service.go`).
- Exposes APIs to manage external database connections (`GET/POST /cms/databases`).
- Enables dynamic creation of data models/services (`GET/POST /cms/services`).
- Definers fields, constraints, and relationships dynamically.

### Dynamic CRUD & Query Engine
Managed by `internal/services/dynamic_data_service.go`.
- Automatically generates generic endpoints (Create, Read, Update, Delete, List).
- Handles advanced query filtering (e.g., `?email=john@example.com`), sorting (`?sort=created_at:desc`), pagination, and relations dynamically using GORM mappings.
- Validates Role-Based Access Control before fulfilling requests.

### Schema Migration Engine
Managed by `internal/services/migration_service.go`.
- Supports safe schema changes like adding, dropping, or renaming columns.
- Uses GORM’s Migrator to apply changes dynamically to connected databases.
- Includes logging of migration statuses for safety.

### GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
- Provides optional GraphQL endpoints via `gqlgen` generated at runtime or build time.
- Translates GraphQL queries and mutations into dynamic data fetch operations.

### Backup Engine
Managed by `internal/services/backup_service.go`.
- Implements backup operations for table contents and schemas.
- Takes dynamic table contents and serializes them into JSON snapshots.

### Observability Integration
The platform offers full-stack state-of-the-art observability:
- **Logging**: Zap structured request logs (`pkg/logger/`).
- **Tracing**: OpenTelemetry wrapping DB spans and HTTP lifecycle requests.
- **Metrics**: Prometheus exported `/metrics` tracking latency, error rates, and requests.

---

## 5. Deployment & Containerization (Docker)

The project includes container support using Docker and Docker Compose for the main application and infrastructure services (Redis, PostgreSQL, Prometheus).

To spin up the local development stack:

```bash
docker-compose up -d
```

Ensure a valid `.env` file (copied from `.env.example`) is present at the root before starting the containers.

---

## 6. API Documentation (Swagger/OpenAPI)

REST API endpoints are documented using Swagger.
To generate the Swagger docs natively:

```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

*(Note: The generated `docs/` folder is intentionally excluded from version control).*

---

## 7. Unit Testing

The platform targets a minimum of **80% unit test coverage** across Repository, Service, and API handler layers.

Tests are powered by standard `go test` utilities with in-memory SQLite (`:memory:`) mocking for database interactions.

**Run All Unit Tests:**
```bash
make test-unit
# or
go test ./tests/unit/...
```

**Run Full Test Suite (with Coverage):**
```bash
make test-coverage
# or manually
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

---

## 8. Development Commands

A `Makefile` is included to streamline common tasks:

- `make build` - Compiles the binary to `bin/cms-backend`.
- `make dev` - Runs the application with hot-reloading (requires `air`).
- `make test-unit` - Runs unit tests.
- `make test-integration` - Runs integration tests.
- `make test-uat` - Runs user acceptance tests.

Ensure you set `GOTOOLCHAIN=local` if running Go 1.24 locally to avoid toolchain download timeout issues during builds.
