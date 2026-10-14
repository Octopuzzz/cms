# Dynamic CMS + API Builder (Backend-as-a-Service)

A **production-grade Backend Platform** (Backend-as-a-Service) written in **Go**. This platform acts as a self-hosted, Go-native backend infrastructure generator.

The system allows developers to:
- Connect databases
- Create data models dynamically
- Generate CRUD APIs automatically
- Generate optional GraphQL APIs
- Manage schema migrations
- Monitor logs and performance
- Manage backups
- Scale services

---

## Core Technologies

- **Language:** Go (latest stable version)
- **API Layer:** REST (default) & GraphQL (optional via gqlgen)
- **Web Framework:** Gin
- **ORM:** GORM
- **Logging:** Zap structured logging
- **Metrics:** Prometheus
- **Tracing:** OpenTelemetry
- **Cache:** Redis
- **Message Queue:** NATS or Kafka (optional)
- **Database Migration:** Atlas or Goose (and dynamic GORM migrator)
- **Containerization:** Docker
- **API Documentation:** Swagger / OpenAPI

---

## High-Level Platform Architecture

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

## System Architecture Principles

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

---

## Project Structure

The project uses a clean modular structure:

```text
.
├── cmd/
│   └── server/          # Entry point for the application
├── internal/
│   ├── handlers/        # HTTP presentation layer (API & CMS routes)
│   ├── services/        # Application business logic (CRUD Engine, CMS, Backup)
│   ├── models/          # Domain layer schemas and data representations
│   ├── database/        # Infrastructure: connection managers and pooling
│   ├── graphql/         # GraphQL gateway integration
│   ├── config/          # Environment configuration
│   ├── middleware/      # Auth, rate limiting, and metrics middleware
│   └── tracing/         # OpenTelemetry instrumentation
├── pkg/
│   ├── logger/          # Zap structured logging wrapper
│   └── response/        # Standardized API response formatters
├── tests/               # Unit, integration, and UAT test suites
└── docker/              # Containerization and infrastructure setups
```

---

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. Tables defined in `internal/models/models.go` include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`). Fields: `id`, `name`, `database_id`, `created_at`, `updated_at`.
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

---

## Core Components Implementation

### 1. CMS Control Plane & Service Builder
Managed under `/api/v1/cms/` routes and handled by `internal/services/service_service.go` and `internal/services/dbconn_service.go`.
- Provides an API interface to store service and field definitions.
- Connects to external databases (PostgreSQL, MySQL, MongoDB, SQLite).
- Automatically manages dynamic schemas upon service updates.

### 2. CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go`.
- Automatically maps endpoints like `GET /api/v1/data/{slug}` to standard GORM operations.
- **Filtering:** `GET /api/v1/data/users?email=test@test.com`
- **Sorting:** `GET /api/v1/data/users?sort=created_at:desc`
- **Pagination:** `GET /api/v1/data/users?page=1&limit=20`
- Includes dynamic joining and RBAC access validation.

### 3. Relation Engine
Supports relationships: One-to-One, One-to-Many, Many-to-One, Many-to-Many.
Automatically creates join tables for many-to-many associations (e.g., `users` ↔ `roles` creates `user_roles`).

### 4. Schema Migration Engine
Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s Migrator.
- Provides capabilities for logging migration history.

### 5. GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.
- Generates dynamic schemas and resolves generic queries and mutations based on created services.

### 6. Backup System
Managed by `internal/services/backup_service.go`.
- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring records via JSON payload decoding.

### 7. Observability Integration
Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes. Exported at `/metrics`.

### 8. Security
- Complete standard input validation and SQL injection protection through parameterized GORM queries.
- JWT Authentication and Rate Limiting implemented inside `internal/middleware/`.

---

## Testing & Docker

- The system features comprehensive test coverage (`>80%`) across Repositories, Services, and API Handlers using `go test`.
- Containerization setup is configured natively in `docker/` allowing multi-node deployments of the database, Redis cache, and application layers.
- Generate Swagger Documentation easily via `swag init -g cmd/server/main.go`.