# Dynamic CMS & API Builder Backend Platform

This repository contains a **production-grade Backend Platform (Dynamic CMS + API Builder) written in Go**. It is a fully self-hosted, Go-native Backend-as-a-Service (BaaS) platform that allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints, similar to Hasura or Supabase.

## System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer (`internal/handlers/`)**: The Gin HTTP Router and Handlers map incoming REST and GraphQL requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer (`internal/services/`)**: Business logic. Services control data access, generate dynamic schemas, validate permissions, and perform operations requested by handlers.
- **Domain Layer (`internal/models/`)**: The data models, defining core entities like `Service`, `Field`, `DatabaseConnection`, `User`, and `Role`.
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

The platform metadata is managed and stored in the internal database. The tables include:

1. **`database_connections`**: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. **`services`**: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. **`fields`**: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. **`service_permissions`**: Connects `Role` to `Service` providing fine-grained access checks (e.g., CanCreate, CanRead). Contains row-level and field-level capabilities.
5. **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. **`backups`**: Stores snapshot records or schema outputs.
7. **`users`** and **`roles`**: General authentication and authorization for the control plane.
8. **`audit_logs`**: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure to ensure maintainability and testability:

```text
.
├── cmd
│   └── server
│       └── main.go               # Application entry point
├── docker
│   ├── Dockerfile                # Docker build definitions
│   ├── docker-compose.yml        # Docker compose configuration
│   └── prometheus.yml            # Prometheus configuration
├── internal
│   ├── config                    # Configuration loading and setup
│   ├── database                  # Connection manager and DB logic
│   ├── graphql                   # GraphQL gateway and resolvers
│   ├── handlers                  # HTTP controllers (Presentation Layer)
│   ├── middleware                # Auth, Rate Limiter, Prometheus
│   ├── models                    # Domain models and entities
│   ├── services                  # Business logic (Application Layer)
│   └── tracing                   # OpenTelemetry / Jaeger setup
├── pkg
│   ├── logger                    # Structured Zap logger wrapper
│   └── response                  # Standardized Gin response helpers
└── tests
    ├── integration               # Integration tests
    ├── uat                       # User Acceptance Tests / End-to-End
    └── unit                      # Unit tests
```

## CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`, the dynamic CRUD engine handles operations on created services.
- It maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- It dynamically processes operations including **Create, Read, Update, Delete**, and **List**.
- Applies automated filtering via query params (e.g., `?email=test@test.com`), sorting (`?sort=created_at:desc`), pagination (`?page=1&limit=20`), and dynamic joining via foreign key mappings (`?join=...`).
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## Schema Migration Engine

Managed by `internal/services/migration_service.go`. The platform tracks and applies safe schema diffs using GORM’s auto-migration capabilities.
- Supports operations such as adding, dropping, and renaming columns, and changing column types.
- Provides rollback capabilities by logging migration status directly into the `migrations` table and handling snapshot retentions.
- Automatically handles backing up table structures before significant modifications.

## GraphQL Gateway (Optional)

Managed by `internal/graphql/gateway.go`.
- Generated from service schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models, providing support for queries, mutations, and relations.

## Backup System

Managed by `internal/services/backup_service.go`.
- The system supports creating backups for specific services using `POST /api/v1/cms/backup/service/{service_id}`.
- Generates JSON snapshot backups mapping dynamic table contents.
- Supports restoring records using `POST /api/v1/cms/restore/{id}` via JSON payload decoding.

## Observability Integration

The platform provides deep visibility into its operations natively across the stack:
- **Logging**: Structured, high-performance request logs, error logs, and query logs using **Zap Logging** (`pkg/logger/`). Injected globally with correlation and trace IDs tied directly into Context.
- **Tracing**: Request tracing wrapped around SQL commands and network logic initialized with **OpenTelemetry** and exported to **Jaeger** (`internal/tracing/`).
- **Metrics**: **Prometheus** metrics are tracked using `internal/middleware/prometheus.go` to measure request latency, error rates, and throughput. Exported at `/metrics`.

## Unit Tests

The system maintains comprehensive unit testing coverage:
- Includes tests for the `Repository`, `Service`, and `API handlers` layers.
- Achieves a minimum code coverage of 80 percent.
- Uses `github.com/glebarez/sqlite` driver for fast, in-memory SQLite database (`:memory:`) mocking.
- Run tests via standard Go tooling: `go test ./...` or via `make test`.
- Coverage can be calculated using `make test-coverage`.

## Docker Setup

The repository is containerized and ready for scalable deployments using Docker.
- A multi-stage `Dockerfile` ensures optimal image sizing.
- A `docker-compose.yml` is provided out of the box to orchestrate the API backend alongside required services: PostgreSQL (Metadata DB), Redis (Caching), Jaeger (Tracing), and Prometheus (Metrics).
- Run `make docker-up` or `docker-compose up -d` to spin up the local development/production environment.

## Swagger Documentation

API Documentation is auto-generated using standard OpenAPI / Swagger specifications via the `swaggo/swag` library.
- Documentation annotations are embedded directly within controller methods.
- Accessed via `GET /api/v1/swagger/*any`.
- Regenerate the docs by running `make swagger`.
