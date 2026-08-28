# Dynamic CMS + API Builder Platform

Welcome to the **Dynamic CMS + API Builder Platform**, a production-grade Backend-as-a-Service (BaaS) system designed to act as your ultimate backend infrastructure generator. Built with Go, it dynamically creates backend services, defines data schemas, auto-generates REST CRUD endpoints and optional GraphQL APIs, and includes built-in observability, schema migrations, and backup capabilities.

---

## 1. System Architecture Explanation

The platform is designed following **Clean Architecture** and **Domain-Driven Design (DDD)** principles, ensuring a decoupled, scalable, and highly modular foundation.

### Core Layers:
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

### High-Level Platform Architecture:

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

The core internal configuration is stored in the **Metadata Database**, which can be backed by PostgreSQL. The tables store all information needed for the dynamic data models:

1. **database_connections**: Stores external DB configurations (`id`, `name`, `type`, `host`, `port`, `credentials`). Implements connection pooling parameters.
2. **services**: Represents user-created data models (e.g., `id`, `name`, `database_connection_id`, `db_table_name`, `created_at`, `updated_at`).
3. **fields**: Defines attributes for each service, such as type (string, integer, float, uuid, JSON). Configures uniqueness, nullability, defaults.
4. **service_permissions**: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. **migrations**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. **backups**: Stores snapshot records or schema outputs.
7. **users** and **roles**: General authentication and authorization for the control plane.
8. **audit_logs**: Detailed logging of structural and data-level modifications.

---

## 3. Go Project Structure

The platform implements a clean, modular structure aligned with standard Go layout practices:

```text
.
├── cmd
│   └── server                # Application entrypoint (main.go)
├── internal
│   ├── handlers              # API Handlers (REST & GraphQL)
│   ├── services              # Business logic (CMS, CRUD, Migration, etc.)
│   ├── database              # Database Connection Manager
│   ├── models                # Domain layer schemas and entities
│   ├── graphql               # GraphQL Gateway logic
│   ├── tracing               # OpenTelemetry setup
│   ├── middleware            # Auth, Rate Limiter, Prometheus
│   └── config                # Environment and configuration
├── pkg
│   ├── logger                # Zap structured logging
│   ├── response              # HTTP Response helpers
│   └── errors                # Custom error types
├── tests
│   ├── unit                  # Unit tests (Services, Repo)
│   ├── integration           # Integration tests
│   └── uat                   # End-to-end / User Acceptance Testing
├── docker                    # Docker setups and scripts
├── Dockerfile                # Image build configuration
├── docker-compose.yml        # Local multi-container setup
└── Makefile                  # Task runner definition
```

---

## 4. Dynamic CRUD & Query Engine Implementation

### Dynamic CRUD Engine
When a service is created, the system auto-generates CRUD endpoints via the `DynamicDataService`. Operations are dynamically mapped to database instructions through GORM.

- **Create**: `POST /api/v1/data/{slug}`
- **Read**: `GET /api/v1/data/{slug}/{id}`
- **Update**: `PUT /api/v1/data/{slug}/{id}`
- **Delete**: `DELETE /api/v1/data/{slug}/{id}`
- **List**: `GET /api/v1/data/{slug}`

### Query Engine
Advanced querying is natively supported, parsing URL parameters into SQL execution blocks.
- **Filtering**: `GET /api/v1/data/users?email=john@example.com`
- **Sorting**: `GET /api/v1/data/users?sort=created_at:desc`
- **Pagination**: `GET /api/v1/data/users?page=1&limit=20`
- **Join Queries**: `GET /api/v1/data/orders?join=user,products`
The Engine seamlessly verifies Role-Based Access Control and Row-Level filtering dynamically prior to executing these queries.

---

## 5. Schema Migration Engine

Managed by `MigrationService`, the system safely applies schema changes to external databases using GORM's Migrator.
- **Supported Operations**: Add column, Drop column, Rename column, Change column type.
- **Safety Features**: Tracks migration status natively into the `migrations` table, triggers snapshots, and supports rollbacks to maintain schema integrity.

---

## 6. GraphQL Gateway

The platform automatically builds an optional GraphQL layer based on defined service schemas using the `gqlgen` library.
- Managed by `internal/graphql/gateway.go`.
- Exposes queries, mutations, and dynamic relations over `POST /api/v1/graphql`.

Example Query:
```graphql
query {
  users {
    id
    name
    email
  }
}
```

---

## 7. Backup System

Managed by `BackupService`, this engine guarantees data preservation during migrations and on-demand calls.
- **Operations Supported**: Snapshot data from a specific service.
- **Mechanics**: Implements generic service record backup features mapping dynamic table contents to snapshots.
- **Restoration**: Supports restoring rows via JSON payload decoding.

API Examples:
- Backup: `POST /api/v1/cms/backup/service/{service_id}`
- Restore: `POST /api/v1/cms/restore/service/{service_id}`

---

## 8. Observability Integration

Modern production platforms demand extreme visibility. This is embedded across the stack:
- **Logging**: Zap structured logging injected globally with correlation/trace IDs via `pkg/logger`.
- **Tracing**: OpenTelemetry (Jaeger compatible) wrappers around SQL commands and HTTP endpoints located in `internal/tracing/`.
- **Metrics**: Prometheus instrumentation (`internal/middleware/prometheus.go`) tracks latency, request rates, and status codes. Accessible at `/metrics`.

---

## 9. Unit Tests

Quality is enforced via stringent test requirements.
- **Minimum Coverage**: 80% coverage across Repository, Service, and API handler layers.
- Unit tests use in-memory SQLite mapping to prevent dependencies on persistent state.
- **Command**: Run the full suite using `make test-coverage` or `go test ./...`.

---

## 10. Docker Setup

A complete `docker-compose.yml` and `Dockerfile` are provided to spin up the entire cluster natively, integrating:
- The Go Backend (using distroless / alpine bases for size).
- A Postgres / MySQL Database.
- Redis (For Query Caching).
- Prometheus & Grafana (For Monitoring).
- Jaeger (For Distributed Tracing).

**Commands**:
- `docker-compose up -d` (To start the stack).
- `docker-compose down` (To tear down the stack).

---

## 11. Swagger Documentation

API Documentation is auto-generated using standard OpenAPI / Swagger annotations.
- Using standard tools: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`
- Exposed natively during runtime, enabling easy developer onboarding.
