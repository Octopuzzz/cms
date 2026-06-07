# CMS Backend Platform

This repository contains a production-grade, self-hosted Backend Platform (Dynamic CMS + API Builder) written natively in Go.

The platform allows developers to dynamically create backend services, schemas, REST and GraphQL APIs, manage database connections, handle migrations, perform backups, and monitor observability.

It is designed following Clean Architecture and Domain-Driven Design (DDD) principles and acts as a robust backend infrastructure generator.

## Core Technology Stack

- **Language**: Go (v1.24)
- **API Layers**: REST (default via Gin), GraphQL (optional via gqlgen)
- **ORM**: GORM
- **Database Support**: PostgreSQL, MySQL, MongoDB, SQLite (in-memory tests)
- **Logging**: Zap structured logging
- **Metrics**: Prometheus
- **Tracing**: OpenTelemetry (Jaeger)
- **API Documentation**: Swagger / OpenAPI
- **Containerization**: Docker & Docker Compose

## System Architecture Explanation

The CMS Backend strictly adheres to **Clean Architecture** and **Domain-Driven Design (DDD)**:

- **Presentation Layer** (`internal/handlers/`, `internal/graphql/`): Maps incoming HTTP REST and GraphQL requests to internal services via Gin.
- **Application Layer** (`internal/services/`): Core business logic handling service lifecycle, dynamic CRUD operations, migrations, backups, and user authorization.
- **Domain Layer** (`internal/models/`): Defines robust core data models (Services, Fields, DB Connections, Users, Roles).
- **Infrastructure Layer** (`internal/database/`, `internal/tracing/`, `pkg/logger/`): Provides cross-cutting concerns like DB connection pooling, OpenTelemetry, Prometheus metrics, and global JSON structured logging.

### High Level Architecture Diagram

```
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
PostgreSQL    MySQL       MongoDB
```

## Metadata Database Schema

The platform stores its configuration in an internal metadata database. The key tables include:

- `database_connections`: External database configurations (PostgreSQL, MySQL, etc.) supporting connection pooling metrics.
- `services`: Dynamic models mapped to actual database tables. Links to a specific `database_connection`.
- `fields`: Definitions of the columns/attributes for a `service` (string, int, uuid, json, relations). Tracks uniqueness, nullability, and default values.
- `service_permissions`: Fine-grained Role-Based Access Control (RBAC) connecting `Role` to `Service`, supporting row-level and field-level permissions.
- `menus`: Tree structure for mapping services to UI interfaces.
- `custom_validations`: Reusable custom rules applied during data mutation.
- `migrations`: DDL execution history for safe tracking and rollback.
- `backups`: Data and schema snapshot records.
- `users`, `roles`, `permissions`, `user_roles`, `role_permissions`: Comprehensive identity and access management.
- `audit_logs`: Detailed tracking of structural and data changes across the system.

## Go Project Structure

The structure emphasizes modularity and cloud readiness:

```text
.
├── cmd
│   └── server          # Application entry point (main.go)
├── docker              # Dockerfile and Docker Compose configurations
├── docs                # Auto-generated Swagger/OpenAPI documentation
├── internal
│   ├── config          # Environment configuration loader
│   ├── database        # Database connection managers and pooling
│   ├── graphql         # GraphQL gateway and resolvers (gqlgen)
│   ├── handlers        # Gin REST API controllers
│   ├── middleware      # Auth, CORS, Rate Limiting, Metrics
│   ├── models          # Domain entity definitions
│   ├── services        # Application business logic (CRUD Engine, CMS Plane, etc.)
│   └── tracing         # OpenTelemetry initialization
├── pkg
│   └── logger          # Zap structured logging wrapper
└── tests
    ├── integration     # Service and database integration testing
    ├── uat             # End-to-end / User Acceptance Testing
    └── unit            # Isolated unit tests utilizing in-memory SQLite
```

## Core Engine Implementations

### CRUD Engine & Query Engine
Housed within `internal/services/dynamic_data_service.go`, the dynamic engine:
- Intercepts requests like `GET /api/v1/data/{slug}`.
- Maps the dynamic service schema to standard GORM database operations on the fly.
- Applies automated filtering (e.g., `?email=test@test.com`), dynamic joining for complex relations, sorting, and pagination.
- Dynamically enforces role-based, row-level, and field-level permission structures prior to execution.

### Schema Migration Engine
The `internal/services/migration_service.go` performs safe schema changes utilizing GORM's `Migrator()` interface.
- Supported operations: adding, dropping, renaming columns, and modifying types.
- Natively tracks migration status in the `migrations` table and handles rollbacks through schema snapshot retention.

### GraphQL Gateway
Managed in `internal/graphql/`, it dynamically generates optional GraphQL schemas driven by the service metadata.
- Exposes `POST /api/v1/graphql` and playground `GET /api/v1/graphql/playground`.
- Supports automated queries and mutations reflective of current dynamic models.

### Backup System
`internal/services/backup_service.go` supports extracting table schemas and data sets into generic JSON formats, preserving dynamic data structures across any configured database type.
- Allows capturing system state and restoring specific records via API (`POST /api/v1/cms/backup/service/{id}`, `POST /api/v1/cms/restore/{id}`).

### Observability Integration
The entire application is wrapped with standard observability tools:
- **Logging**: Zap handles JSON structured logging (`pkg/logger/`), inherently capturing trace and correlation IDs.
- **Tracing**: OpenTelemetry (Jaeger configuration via `internal/tracing/`) traces requests and tracks deeply nested function latency.
- **Metrics**: Prometheus middleware intercepts all traffic via `internal/middleware/prometheus.go` and exposes `/metrics` detailing API response times and throughput.

## Unit Testing

The repository relies strictly on Go's standard `testing` framework utilizing an in-memory SQLite (`:memory:`) driver for high performance and reliable unit tests.

*Coverage requirements are set at a minimum of 80% across Respository, Service, and API handler layers.*

To run tests:
```bash
make test
```
Or for isolated unit testing:
```bash
make test-unit
```

## Docker Setup

A complete `docker-compose.yml` is provided alongside a multi-stage `Dockerfile`.

To start the platform, along with Prometheus:
```bash
make docker-up
# OR
docker-compose up -d
```

## Swagger Documentation

API documentation is generated dynamically utilizing `swaggo`.
To parse and rebuild the API schemas:
```bash
make swagger
```
The output is stored in `/docs` and is served live at `http://localhost:8080/swagger/index.html`.

---
*This repository and architecture act as a high-performance, dynamic Backend-as-a-Service.*
