# Backend Platform (Dynamic CMS + API Builder)

This repository contains a **production-grade Backend Platform** (Dynamic CMS + API Builder) written in Go. It operates similarly to platforms like Hasura or Supabase but is fully **self-hosted and Go-native**. The system allows developers to dynamically create backend services, schemas, and APIs, generating both REST and optional GraphQL endpoints automatically.

## System Architecture Explanation

The system is built upon **Clean Architecture** and **Domain-Driven Design (DDD)**.

### High-Level Architecture

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

- **Presentation Layer**: Handled by the Gin HTTP framework (in `internal/handlers/`), routing API requests and exposing GraphQL via `gqlgen` (`internal/graphql/`).
- **Application Layer**: Contains business logic (`internal/services/`). The services control data access, generate schemas, and implement dynamic engine logic.
- **Domain Layer**: Houses core platform entities (`internal/models/`) like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`, etc.
- **Infrastructure Layer**: Incorporates database integration (`internal/database/`), observability tracing (`internal/tracing/`), middleware (`internal/middleware/`), and logging (`pkg/logger/`).

## Metadata Database Schema

The core internal metadata configuration is stored in the CMS's primary database (typically PostgreSQL or SQLite). Tables include:

1. **`database_connections`**: Stores user-configured external databases. Fields: id, name, type (PostgreSQL, MySQL, MongoDB), host, port, credentials.
2. **`services`**: Represents dynamic data models (services). Stores configurations and links to the `database_connections`.
3. **`fields`**: Configures attributes for each service (type: string, integer, boolean, uuid, json, timestamp, etc.), along with rules for uniqueness, default values, and nullability.
4. **`relations`**: (and join tables) Configures One-to-One, One-to-Many, Many-to-One, and Many-to-Many entity relationships.
5. **`service_permissions`**: Maps Role capabilities to Services.
6. **`migrations`**: Keeps track of applied DDL schema changes and rollbacks.
7. **`backups`**: Records created data snapshots and schemas.
8. **`users`** & **`roles`**: RBAC system for the control plane.
9. **`audit_logs`**: Logs of structural modifications.

## Go Project Structure

The project uses a clean modular structure:

```
.
├── cmd
│   └── server                # Application entrypoint
├── docker                    # Docker and Prometheus configurations
├── internal
│   ├── config                # Environment configurations
│   ├── database              # Database Connection Manager
│   ├── graphql               # Auto-generated GraphQL Gateway via gqlgen
│   ├── handlers              # Gin REST Handlers
│   ├── middleware            # Auth, Prometheus, Rate Limiting
│   ├── models                # Core Domain Models
│   ├── services              # Application Business Logic
│   └── tracing               # OpenTelemetry Configuration
├── pkg
│   ├── logger                # Structured Zap Logger
│   └── response              # API Standard Responses
└── tests
    ├── integration           # Integration tests
    ├── uat                   # E2E / UAT tests
    └── unit                  # Unit tests (Mock DBs, Coverage >80%)
```

## CRUD Engine Implementation

The Dynamic CRUD Engine is primarily managed by `internal/services/dynamic_data_service.go` and its associated handlers.
When a service is created, it automatically maps generic REST endpoints to dynamic DB queries:

- **Create**: `POST /api/v1/data/{slug}`
- **Read**: `GET /api/v1/data/{slug}/{id}`
- **Update**: `PUT /api/v1/data/{slug}/{id}`
- **Delete**: `DELETE /api/v1/data/{slug}/{id}`
- **List / Query Engine**: `GET /api/v1/data/{slug}`

**Query Engine Capabilities**:
- **Filtering**: `?email=john@example.com`
- **Sorting**: `?sort=created_at:desc`
- **Pagination**: `?page=1&limit=20`

All operations enforce RBAC and execute secure, parametrized queries through GORM.

## Schema Migration Engine

Managed by `internal/services/migration_service.go`.
The engine tracks schema changes made via the Service Builder (Add, Drop, Rename, Change Type). It uses GORM's `Migrator()` interface to apply real-time DDL modifications to target databases safely. It logs the applied migrations natively and can trigger backups to ensure rollback capability prior to altering the schema.

## GraphQL Gateway (Optional)

The GraphQL gateway is implemented using `github.com/99designs/gqlgen`. The entry point resides in `internal/graphql/gateway.go`.
When services and fields are created, the platform maps their underlying schemas into GraphQL generic types, exposing them automatically at:
- `POST /api/v1/graphql`

This supports queries, mutations, and deeply nested relational joins out of the box.

## Backup System

The Backup Engine (`internal/services/backup_service.go`) implements operations to preserve system states:
- **Snapshots**: Generates serialized JSON and data state snapshots of specific services.
- Records of backups are stored in the metadata DB and can be queried or restored via backup handlers (e.g., `POST /api/v1/cms/backup/service/{service_id}`).

## Observability Integration

Modern, state-of-the-art observability is implemented across the stack:
- **Logging**: Uses Uber's `Zap` logger (`pkg/logger/`) for high-performance structured JSON logging. It injects context and request IDs.
- **Metrics**: Standard HTTP metrics (request latency, status codes, rates) are captured using the `Prometheus` middleware (`internal/middleware/prometheus.go`), exposed at `/metrics`.
- **Tracing**: Fully integrated with `OpenTelemetry` (`internal/tracing/tracing.go`). The application emits spans tracing SQL queries, HTTP interactions, and internal processes directly to Jaeger/OTel Collectors.

## Unit Tests

The system requires strict test coverage. Tests reside in the `tests/` directory:
- Contains `unit`, `integration`, and `uat` (E2E) suites.
- Minimum expected unit test coverage is **80%** across Repositories, Services, and Handlers.
- Run tests using the included Makefile:
  - `make test-unit`
  - `make test-coverage`
- Unit tests run against an in-memory SQLite database utilizing `github.com/glebarez/sqlite` to mock connections efficiently without requiring an external service.

## Docker Setup

The platform is completely container-ready.
- A standard **`Dockerfile`** at the root compiles the Go binary into an optimized Linux image.
- **`docker-compose.yml`** orchestrates the platform alongside essential backing services:
  - The API Service (CMS Backend)
  - PostgreSQL (Metadata DB)
  - Redis (Caching / Pooling / Rate Limiter)
  - Prometheus (Metrics scraping)
  - Jaeger (Distributed tracing collector)

Run using: `docker-compose up -d` or `make docker-up`.

## Swagger Documentation

Auto-generated Swagger (OpenAPI) documentation is fully supported.
- Gin-Swagger serves the UI exposing CMS Control Plane API definitions.
- Generate docs using: `make swagger` (which runs `swag init`).
- The output resides in the `docs/` folder (ignored from version control to prevent artifact clutter).
