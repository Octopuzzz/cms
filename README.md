# Backend Platform (Dynamic CMS + API Builder)

## Platform Goal
Build a Backend-as-a-Service platform allowing developers to:
- Connect databases
- Create data models dynamically
- Generate CRUD APIs automatically
- Generate optional GraphQL APIs
- Manage schema migrations
- Monitor logs and performance
- Manage backups
- Scale services

This platform acts as a backend infrastructure generator, similar to Hasura or Supabase but fully self-hosted and Go-native.

## Core Technology Requirements
- **Language**: Go
- **API Layer**: REST (default), GraphQL (optional using gqlgen)
- **Web Framework**: Gin
- **ORM**: GORM
- **Logging**: Zap structured logging
- **Metrics**: Prometheus
- **Tracing**: OpenTelemetry
- **Database Migration**: Natively implemented via GORM's `.Migrator()` and tracked in DB
- **Containerization**: Docker
- **API Documentation**: Swagger / OpenAPI

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

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**.

Tables defined in `internal/models/models.go` include:

1. **database_connections**: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. **services**: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. **fields**: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. **service_permissions**: Connects `Role` to `Service` providing fine-grained access checks (e.g., CanCreate, CanRead). Contains row-level and field-level capabilities.
5. **migrations**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. **backups**: Stores snapshot records or schema outputs.
7. **users** and **roles**: General authentication and authorization for the control plane.
8. **audit_logs**: Detailed logging of structural and data-level modifications.

## Component Implementation

### CMS Control Plane & Service Builder
Managed under `/api/v1/cms/` routes and handled by `internal/services/service_service.go` and `internal/services/dbconn_service.go`. These services:
- Provide an API interface to store service and field definitions in the metadata database.
- Utilize the `database_connections` engine to validate foreign database connectivity.
- Trigger migrations on service model updates to reflect schema changes.

### CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go`.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g., `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

### Schema Migration Engine
Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
- Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.

### GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

### Backup System
Managed by `internal/services/backup_service.go`.
- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.

### Observability Integration
Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.

## Project Structure

```text
.
├── ARCHITECTURE.md
├── Dockerfile
├── Makefile
├── cmd
│   └── server
│       └── main.go       # Entrypoint
├── docker
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── prometheus.yml
├── docker-compose.yml
├── go.mod
├── go.sum
├── internal
│   ├── config          # Configuration management
│   ├── database        # Database connections and pooling
│   ├── graphql         # GraphQL gateway
│   ├── handlers        # HTTP handlers / API layer
│   ├── middleware      # Gin middleware (Auth, Prometheus, Rate Limiter)
│   ├── models          # Domain models / DB schema
│   ├── services        # Core application logic
│   └── tracing         # OpenTelemetry setup
├── pkg
│   ├── logger          # Zap logger wrappers
│   └── response        # Standardized Gin responses
└── tests
    ├── integration     # Integration tests
    ├── uat             # End-to-end tests
    └── unit            # Unit tests
```
