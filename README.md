# Go Backend Platform (Dynamic CMS + API Builder)

This project is a production-grade Backend-as-a-Service (BaaS) platform written in Go, acting as a dynamic CMS and API builder. It is designed to be fully self-hosted, modular, scalable, and cloud-ready, serving as a backend infrastructure generator.

## System Architecture Explanation

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: Handled by the Gin HTTP router (`internal/handlers/`), serving as the API Gateway for incoming REST requests, and `gqlgen` for optional GraphQL endpoints.
- **Application Layer**: Contains core business logic (`internal/services/`). The services control data access, generate dynamic schemas, and orchestrate the operations requested by handlers.
- **Domain Layer**: The core data models (`internal/models/`) define entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`, `Migration`, and `Backup`.
- **Infrastructure Layer**: Cross-cutting concerns such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

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

The CMS stores platform metadata in the metadata database (PostgreSQL/MySQL/SQLite). Key tables include:

1. `database_connections`: Stores external database configurations (id, name, type, host, port, credentials).
2. `services`: Represents user-created data models, references `database_connection_id`, tracks schemas.
3. `fields`: Defines attributes for each service (type, uniqueness, nullability, defaults).
4. `service_permissions`: Connects `Role` to `Service` for RBAC.
5. `migrations`: Tracks DDL executions and schema changes.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` & `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of modifications.

## Go Project Structure

The project follows a standard modular Go structure:

```
.
├── cmd/
│   └── server/                # Application entrypoint
├── docker/                    # Docker configs (Prometheus, etc.)
├── internal/
│   ├── config/                # Environment configuration
│   ├── database/              # DB connection manager and routing
│   ├── graphql/               # GraphQL gateway and schemas
│   ├── handlers/              # Gin HTTP handlers (Presentation)
│   ├── middleware/            # Auth, metrics, telemetry middleware
│   ├── models/                # GORM domain models
│   ├── services/              # Core business logic (Application)
│   └── tracing/               # OpenTelemetry integration
├── pkg/
│   ├── logger/                # Zap logging implementation
│   └── response/              # Standardized API responses
└── tests/
    ├── integration/           # Cross-component tests
    ├── uat/                   # User acceptance tests
    └── unit/                  # Unit tests for handlers and services
```

## Core Components Implementation

### CRUD & Query Engine
Managed by `DynamicDataService` (`internal/services/dynamic_data_service.go`), the system dynamically generates CRUD operations for user-defined services. It maps API requests (e.g., `GET /api/v1/data/{slug}`) to standard GORM operations on the fly. It supports automated filtering (`?email=test@example.com`), sorting, pagination, and dynamic joining based on foreign key mappings.

### Schema Migration Engine
Managed by `MigrationService` (`internal/services/migration_service.go`), the platform handles safe schema changes (Add/Drop/Rename Column) using GORM's AutoMigrator. Migrations are tracked in the `migrations` table, allowing for status monitoring, snapshots, and rollback functionality.

### GraphQL Gateway
The platform automatically generates optional GraphQL APIs from service schemas. Built using `gqlgen`, the gateway is configured in `internal/graphql/` and exposes `POST /api/v1/graphql` to serve dynamic queries, mutations, and relationship resolution.

### Backup System
The Backup Engine (`internal/services/backup_service.go`) supports taking data snapshots, schema backups, and specific service snapshots. It serializes dynamic table contents to JSON payloads and manages the records within the `backups` metadata table.

### Observability Integration
The platform incorporates state-of-the-art observability:
- **Logging**: Structured JSON logging using Zap (`pkg/logger/`) with correlation IDs.
- **Metrics**: Prometheus metrics exposed via `/metrics` and tracked through `internal/middleware/prometheus.go` for request latency and error rates.
- **Tracing**: OpenTelemetry (Jaeger) initialized in `internal/tracing/` wraps SQL commands and HTTP requests.

### Docker Setup
The project is fully containerized. A `Dockerfile` is provided for the main Go application, and a `docker-compose.yml` orchestrates the backend platform along with necessary infrastructure like PostgreSQL, Redis, Prometheus, and Jaeger.

### Swagger Documentation
API documentation is generated using `swag`. The definitions are parsed from handler annotations and output to the `docs/` folder (ignored by version control). It can be generated by running `swag init -g cmd/server/main.go --parseDependency --parseInternal`.

### Unit Tests
The system includes comprehensive unit testing with over 80% coverage. Tests cover the repository, service, and API handler layers. They are located in `tests/unit/`, `tests/integration/`, and `tests/uat/`, and can be executed via `make test-coverage` or `go test ./...`. Unit tests typically mock database connections using an in-memory SQLite setup.
