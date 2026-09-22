# CMS Backend Platform

This repository contains a production-grade, self-hosted Backend-as-a-Service (BaaS) platform (Dynamic CMS + API Builder) written entirely in Go.

It allows users to dynamically create backend services, manage schemas, generate CRUD REST APIs (and optional GraphQL APIs), handle safe database migrations, perform backups, and monitor performance with state-of-the-art observability. The architecture follows Clean Architecture and Domain-Driven Design (DDD) principles.

---

## 1. System Architecture Explanation

The platform architecture is designed to be highly modular and scalable, with clean boundaries between presentation, business logic, data models, and infrastructure.

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

- **Presentation Layer (`internal/handlers/`, `internal/graphql/`)**: The Gin HTTP Router and handlers for REST. Generates optional schemas using `github.com/99designs/gqlgen/graphql` for GraphQL API.
- **Application Layer (`internal/services/`)**: Business logic, encapsulating features like dynamic CRUD, user/role management, migrations, and database connection handling.
- **Domain Layer (`internal/models/`)**: GORM data models specifying internal relationships, including user definitions, services, fields, backup, and migration logs.
- **Infrastructure Layer (`internal/tracing/`, `internal/middleware/`, `pkg/logger/`)**: Manages telemetry, Prometheus metrics export, global Zap logging, authentication, rate-limiting, and more.

---

## 2. Metadata Database Schema

Platform metadata is stored securely within the metadata database (defaulting to SQLite or PostgreSQL for metadata persistence). Key tables include:

- `database_connections`: Manages external connections for data models (id, name, type, host, port, credentials). Support for connection pooling is fully integrated.
- `services`: Represents user-created data models referencing `database_connection_id`, enabling the mapping of dynamic schemas (`db_table_name`).
- `fields`: Defines attributes for each service (type: string, integer, float, uuid, JSON), configuration for unique, null, defaults.
- `service_permissions`: Links `Role` to `Service` to enforce detailed RBAC permissions (e.g. CanCreate, CanRead).
- `migrations`: Tracks DDL updates and migration history for safety/rollback capabilities.
- `backups`: Retains service-level snapshots of data and schemas.
- `users`, `roles`, `audit_logs`: Authorization metadata and structural auditing logs.

---

## 3. Go Project Structure

The project employs a clean modular structure to separate concerns optimally.

```text
├── cmd
│   └── server                # Application entry point
├── docker                    # Container configs (Prometheus, Dockerfile, etc.)
├── internal                  # Internal application packages
│   ├── config                # App configurations/Env var mapping
│   ├── database              # Database Connection Manager (PostgreSQL, MySQL, MongoDB, SQLite)
│   ├── graphql               # Auto-generated GraphQL Gateway logic
│   ├── handlers              # API Handlers (REST controllers)
│   ├── middleware            # Middlewares (Auth, Prometheus, Rate Limiter)
│   ├── models                # Metadata Database Schema definitions
│   ├── services              # Application logic (Auth, CRUD, Migrations, Users, etc.)
│   └── tracing               # Observability & OpenTelemetry setup
├── pkg                       # Shared libraries
│   ├── logger                # Structured Zap-based custom logging
│   └── response              # HTTP Response formats
└── tests                     # Comprehensive test suites
    ├── integration
    ├── uat
    └── unit
```

---

## 4. CRUD Engine Implementation

The system automatically generates REST APIs (`GET`, `POST`, `PUT`, `DELETE`) for dynamic services.
- Managed by `internal/services/dynamic_data_service.go` and `internal/handlers/dynamic_data_handler.go`.
- Dynamic tables support `POST /api/v1/data/{slug}` to insert data.
- Built-in validation based on defined field attributes (`internal/models.FieldValidations`).
- Role-based Access Control (RBAC) validates whether a user has permissions (e.g. `CanCreate`, `CanRead`) dynamically before serving data.
- Generates batch DB operations optimally, utilizing `db.CreateInBatches` as recommended.

## 5. Query Engine

The API endpoints support advanced querying out-of-the-box:
- **Filtering**: Query parameter-based data filtering.
- **Sorting**: Append `?sort_by=field&sort_order=desc`.
- **Pagination**: Handle records effectively with `?page=1&page_size=20`.
- **Joins**: Extract linked relational structures natively using `?joins=field_name` which the CRUD engine maps through `RelationConfig` defined on a field.

## 6. Schema Migration Engine

Implemented in `internal/services/migration_service.go`.
- Tracks changes to service definitions securely.
- Applies safe schema differences (Add column, Drop column, Rename column, Change type).
- Updates are securely captured inside the `migrations` table allowing state traceability and limited rollback capabilities.

## 7. GraphQL Gateway (Optional)

The CMS generates optional GraphQL API schema handling.
- `POST /api/v1/graphql`: Acts as the main endpoint.
- Playground accessible at `/api/v1/graphql/playground`.
- Integrates with `gqlgen` utilizing a dynamic execution schema (`internal/graphql/gateway.go`) to automatically map schemas over the existing metadata structure.

## 8. Backup System

Managed securely by `internal/services/backup_service.go`.
- Captures backups as table data JSON snapshots and schema configurations.
- `POST /api/v1/cms/backup/service/{service_id}` creates a snapshot copy.
- Data can be retrieved or reverted safely utilizing built-in JSON payload restoration logic (`POST /api/v1/cms/restore/{id}`).

## 9. Observability Integration

Complete observability configuration mapped directly to the core application logic.
- **Zap Logging**: Customized JSON/text logger (`pkg/logger/`) correlating IDs dynamically.
- **OpenTelemetry/Jaeger**: Detailed SQL query and system performance tracing handled via `internal/tracing/tracing.go`.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` integrates native latency and code status monitors exposed automatically at the `/metrics` endpoint.

## 10. Security
- Robust security configurations in place.
- **JWT Authentication** for all private endpoints.
- Role-based Authorization protecting schema generation endpoints (Admin, SuperAdmin) and row-level operations (RBAC).
- Global Rate Limiter implemented via standard HTTP middleware logic.

---

## Swagger Documentation

To generate or update Swagger documentation for API endpoints:
```bash
make swagger
# Or run swag manually
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

## Unit Testing

Run unit, integration, and UAT tests using the Makefile.

```bash
make test-unit
make test-coverage
# Generates coverage in coverage.out / coverage.html
```
Test cases cover minimum threshold standards utilizing isolated memory SQLite testing infrastructures.

## Docker Setup

Use Docker to quickly containerize and spin up the complete platform configuration.

```bash
make docker-up
# Or standard compose
docker-compose up -d
```
Runs application services alongside integrated Prometheus observability pipelines out-of-the-box.
