# Dynamic CMS & API Builder

A production-grade, self-hosted Backend-as-a-Service (BaaS) platform written natively in Go. This platform dynamically generates REST and optional GraphQL APIs, manages database connections, defines data models, handles schema migrations, and includes built-in observability, security, and backup systems. It acts similarly to platforms like Hasura or Supabase but is entirely Go-native.

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)**.

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

### Core Layers:
- **Presentation Layer** (`internal/handlers/`): Gin HTTP Router and Handlers mapping requests to internal services, plus GraphQL endpoints using `gqlgen`.
- **Application Layer** (`internal/services/`): Business logic controlling data access, dynamic schemas, and operations.
- **Domain Layer** (`internal/models/`): Core entities defining `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## 2. Metadata Database Schema

The core internal configuration is stored in the metadata database (PostgreSQL recommended, defaults to SQLite for testing/local).

- `database_connections`: Stores external DB configs (id, name, type, host, port, credentials). Implements connection pooling settings.
- `services`: Represents user-created data models. References `database_connection_id` and tracks dynamic schemas (`db_table_name`).
- `fields`: Defines attributes for each service (string, integer, float, uuid, json, array, timestamp). Configures uniqueness, nullability, defaults.
- `relations`: Maps relations (One-to-One, One-to-Many, Many-to-One, Many-to-Many). Join tables are created automatically for Many-to-Many.
- `service_permissions`: Connects `Role` to `Service` for fine-grained access checks (e.g. CanCreate, CanRead).
- `migrations`: Tracks DDL executions and metadata for rollbacks.
- `backups`: Stores snapshot records or schema outputs.
- `users` / `roles`: Authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of structural and data-level modifications.

## 3. Go Project Structure

The project uses a clean, modular structure.

```text
.
├── cmd/
│   └── server/
│       └── main.go              # Application entrypoint
├── internal/
│   ├── config/                  # Configuration management
│   ├── database/                # Connection manager
│   ├── graphql/                 # GraphQL gateway
│   ├── handlers/                # HTTP/REST handlers
│   ├── middleware/              # Auth, Prometheus, Rate Limiter
│   ├── models/                  # Domain models (metadata schema)
│   ├── services/                # Application logic (Control Plane)
│   └── tracing/                 # OpenTelemetry tracing
├── pkg/
│   ├── logger/                  # Zap logger implementation
│   └── response/                # Standardized JSON responses
├── tests/
│   ├── integration/
│   ├── uat/
│   └── unit/
├── docker/                      # Dockerfile and Docker Compose configurations
├── Makefile                     # Standardized tasks
└── ARCHITECTURE.md              # High-level architecture documentation
```

## 4. CRUD & Query Engine Implementation

Managed by `internal/services/dynamic_data_service.go`, this engine dynamically handles:
- **Operations:** Automatically generates Endpoints for Create, Read, Update, Delete, and List (`/api/v1/data/{slug}`).
- **Advanced Queries:**
  - Filtering: `?email=john@example.com`
  - Sorting: `?sort=created_at:desc`
  - Pagination: `?page=1&limit=20`
  - Joins: Nested join queries supported natively.
- **Security:** Evaluates Role-Based Access Control and Row-Level filtering before query execution.

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema changes using GORM's `.Migrator()`.
- Operations include: Add column, Drop column, Rename column, Change column type.
- Native safety features: Automatic backups before migrations, Rollback capability, logged natively to the `migrations` table.

## 6. GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- Utilizes `github.com/99designs/gqlgen`.
- Automatically generates generic GraphQL schemas directly from defined service schemas.
- Supports Queries, Mutations, and Relations out of the box via `POST /api/v1/graphql`.

## 7. Backup System

Managed by `internal/services/backup_service.go`.
- Generates data snapshots.
- Features: Table backups, Schema backups, Service snapshots.
- Formats: SQL dump, JSON snapshot.
- Endpoints: `POST /cms/backup/service/{service_id}` and `POST /cms/restore/service/{service_id}`.

## 8. Observability Integration

Integrated deeply across the platform to ensure observability and ease of debugging.
- **Logging**: Zap structured logging (`pkg/logger/`) configured globally with request, error, and query logs. Trace and correlation IDs linked in Context.
- **Metrics**: Prometheus tracked via `internal/middleware/prometheus.go` measuring request latency, error rate, and slow queries. Exported at `/metrics`.
- **Tracing**: OpenTelemetry (integrated with Jaeger) tracking network requests, database logic, and internal handler executions (`internal/tracing/`).

## 9. Unit Tests

Robust testing practices ensuring system reliability.
- **Coverage Requirement**: The project mandates a minimum of 80% test coverage across Repository, Service, and API handlers.
- Uses `github.com/glebarez/sqlite` with `:memory:` connection for mocking test databases.
- Run tests via `make test-coverage` or `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`. Note: If local Go is <1.24, prefix with `GOTOOLCHAIN=local` to prevent download timeouts.

## 10. Docker Setup

Full containerization for simple deployments and scaling.
- `docker-compose.yml` configures the backend, a PostgreSQL database, Redis (for caching/rate limiting), Prometheus, Jaeger, and NATS/Kafka out of the box.
- Standardized builds via `Dockerfile` separating build artifacts from runtime images.
- Start infrastructure with `docker-compose up -d`.

## 11. Swagger Documentation

API documentation is generated dynamically via Swagger/OpenAPI.
- Auto-documents REST API specifications for all CMS endpoints.
- Generate via Swag CLI: `~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
- NOTE: The `docs/` directory is intentionally ignored in version control to ensure up-to-date generation during CI/CD pipelines.

---
*Built focusing on maximum modularity, scale, and performance.*
