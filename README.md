# Backend Platform (Dynamic CMS + API Builder)

Welcome to the self-hosted, Go-native Backend-as-a-Service (BaaS) platform. This platform acts as a backend infrastructure generator, allowing developers to dynamically create backend services, define schemas, automatically generate CRUD APIs and optional GraphQL endpoints, and manage schema migrations, all while adhering to Clean Architecture and Domain-Driven Design (DDD).

## Platform Goal

Build a Backend-as-a-Service platform that allows developers to:
- Connect external databases (PostgreSQL, MySQL, MongoDB).
- Create data models dynamically.
- Generate CRUD APIs automatically.
- Generate optional GraphQL APIs.
- Manage schema migrations.
- Monitor logs and performance.
- Manage backups.
- Scale services.

## Core Technology Stack

- **Language:** Go (1.24)
- **API Layer:** REST (default), GraphQL (optional via `gqlgen`)
- **Web Framework:** Gin
- **ORM:** GORM
- **Logging:** Zap structured logging
- **Metrics:** Prometheus
- **Tracing:** OpenTelemetry (Jaeger)
- **Cache:** Redis
- **Containerization:** Docker
- **API Documentation:** Swagger / OpenAPI

## System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles, structured into the following core layers:

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

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

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database** (typically PostgreSQL). Key tables include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials.
2. `services`: Represents user-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`). Fields: `id`, `name`, `database_id`, `created_at`, `updated_at`.
3. `fields`: Defines attributes for each service (e.g., string, integer, float, uuid, JSON). Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks.
5. `relations`: Defines one-to-one, one-to-many, many-to-one, and many-to-many relationships between services.
6. `migrations`: Keeps track of DDL executions with metadata required for rollbacks.
7. `backups`: Stores snapshot records or schema outputs.
8. `users` and `roles`: General authentication and authorization for the control plane.
9. `audit_logs`: Detailed logging of structural and data-level modifications.

## Project Structure

A clean, modular Go project structure:

```text
cmd/
└── server/             # Entry point for the application
internal/
├── config/             # Environment and global configuration
├── database/           # Database Connection Manager and pooling
├── graphql/            # GraphQL Gateway and schema generation
├── handlers/           # HTTP controllers for CMS and dynamic endpoints
├── middleware/         # Auth, Prometheus, and rate limiting middlewares
├── models/             # Domain layer models
├── services/           # Service Builder, CRUD Engine, and Core CMS Logic
└── tracing/            # OpenTelemetry configuration
pkg/
├── logger/             # Zap logger implementation
├── pagination/         # Pagination helpers
└── response/           # Standardized API responses
tests/                  # Unit, Integration, and UAT tests
docker/                 # Docker Compose and configs
```

## Core Components

### 1. CMS Control Plane & Database Connection Manager
Managed under `/api/v1/cms/` routes and handled by internal services. It allows users to register external databases and define services and schemas dynamically.
Supported databases: PostgreSQL, MySQL, MongoDB.
Endpoints:
- `POST /cms/databases`
- `GET /cms/databases`
- `POST /cms/services`

### 2. Service Builder & Relation Engine
Users dynamically create services representing data models. Supported types include string, integer, float, boolean, uuid, json, array, and timestamp. The relation engine handles One-to-One, One-to-Many, Many-to-One, and Many-to-Many relationships, automatically creating join tables where required.

### 3. Dynamic CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go`. It dynamically generates CRUD endpoints mapping to standard database operations based on defined services.
Example Endpoints:
- `POST /api/{service}`
- `GET /api/{service}/{id}`
- `GET /api/{service}?email=john@example.com&sort=created_at:desc&page=1&limit=20`

### 4. Schema Migration Engine
Managed by `internal/services/migration_service.go`. It tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM's Migrator. It inherently supports automatic backups before migrations and rollback capabilities.

### 5. GraphQL Gateway
Managed by `internal/graphql/gateway.go`. Generates a generic optional schema using `gqlgen`. It exposes a GraphQL interface for dynamic queries and mutations over the generated schemas.

### 6. Backup System
Managed by `internal/services/backup_service.go`. Supports table and schema backups, generating JSON snapshots.

### 7. Observability Integration
- **Logging:** Zap Logging is injected globally with correlation and trace IDs.
- **Metrics:** Prometheus tracks request latency and error rates via HTTP interceptors.
- **Tracing:** OpenTelemetry initialized to wrap SQL commands and network logic.

## Security & Performance
- **Security:** Input validation, SQL injection protection (handled via GORM parameterization), JWT authentication, and rate-limiting.
- **Performance:** Connection pooling, potential query caching via Redis, and pagination optimizations.

## Getting Started

### Prerequisites
- Go 1.24
- Docker and Docker Compose
- Make

### Building and Running
1. Start infrastructure dependencies (Postgres, Redis, Jaeger, Prometheus):
   ```bash
   make docker-up
   ```
2. Build and run the server:
   ```bash
   make run
   ```
   Or use `make dev` for hot reloading via `air`.

### Swagger Documentation
Generate and view Swagger API documentation:
```bash
make swagger
```
The documentation will be available at `/swagger/index.html`.

### Unit Testing
The system maintains a minimum of 80% test coverage across Repository, Service, and API handler layers.
Run tests:
```bash
make test
```
Generate coverage report:
```bash
make test-coverage
```

## Maintenance Notes
- Use `GOTOOLCHAIN=local` before `go` commands in restrictive environments to prevent unwanted toolchain downloads.
- Ensure runtime artifacts (`*.db`, `coverage.out`) and `docs/` are excluded from version control.
