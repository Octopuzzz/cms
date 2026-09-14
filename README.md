# CMS Backend Platform

This repository contains a **production-grade Backend Platform (Dynamic CMS + API Builder)** written in Go. It allows developers to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints. The system is fully self-hosted, modular, scalable, and cloud-ready.

## System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

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

The core internal configuration is stored in the **Metadata Database**.

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure.

```text
├── cmd
│   └── server          # Application entrypoint
├── internal
│   ├── api             # API routing
│   ├── config          # Configuration and environments
│   ├── database        # Database Connection Manager
│   ├── graphql         # GraphQL Gateway
│   ├── handlers        # Presentation Layer
│   ├── middleware      # Observability, Auth, Rate Limiting
│   ├── models          # Domain Layer
│   ├── services        # Application Layer (CRUD, Query, Backup, Schema, etc.)
│   └── tracing         # OpenTelemetry setup
├── pkg
│   ├── logger          # Zap Logger wrappers
│   └── response        # Standardized Gin response helpers
├── tests
│   ├── integration     # Integration tests
│   ├── uat             # End-to-end tests
│   └── unit            # Unit tests
└── docker              # Containerization setups (Dockerfile, Compose)
```

## CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g. `?email=test@example.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## Schema Migration Engine

Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
- Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.

## GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

## Backup System

Managed by `internal/services/backup_service.go`.
- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.

## Observability Integration

Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.

## Unit Tests

The system includes unit tests with a minimum coverage target of 80% across Repository, Service, and API handler layers.
- Uses standard Go `testing` library and in-memory SQLite (`:memory:`) via `github.com/glebarez/sqlite`.
- Commands available in Makefile: `make test`, `make test-unit`, `make test-coverage`.

## Docker Setup

Containerization using Docker.
- A `Dockerfile` and `docker-compose.yml` are provided in the repository.
- Use `make docker-up` or `docker-compose up -d` to launch the platform with necessary services (e.g., PostgreSQL, Redis, Prometheus).

## Swagger Documentation

API documentation follows the OpenAPI specification.
- Generatable via the Swagger CLI using `make swagger`.
- Command run: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
