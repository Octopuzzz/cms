# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. It operates similarly to platforms like Hasura or Supabase, acting as a fully self-hosted, Go-native backend infrastructure generator. It empowers developers to connect databases, dynamically create data models, automatically generate CRUD APIs and optional GraphQL endpoints, manage schema migrations, monitor logs and performance, manage backups, and scale services efficiently.

## Core Technology Stack

- **Language:** Go
- **API Layer:** REST (default), GraphQL (optional via `gqlgen`)
- **Web Framework:** Gin
- **ORM:** GORM
- **Logging:** Zap structured logging
- **Metrics:** Prometheus
- **Tracing:** OpenTelemetry
- **Database Migration:** GORM AutoMigrate based
- **Containerization:** Docker
- **API Documentation:** Swagger / OpenAPI

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
PostgreSQL    MySQL       MongoDB
```

### System Architecture Principles

The platform follows **Clean Architecture** and **Domain-Driven Design (DDD)**:
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`), acting as the API gateway. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Controls data access, dynamic schema generation, and operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## Metadata Database Schema

The CMS stores platform metadata, typically in PostgreSQL or SQLite. Core tables include:

- `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
- `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
- `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
- `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead).
- `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- `backups`: Stores snapshot records or schema outputs.
- `users` and `roles`: General authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

A clean, modular structure is used:
```
.
├── cmd
│   └── server                # Main application entry point
├── docker                    # Docker configurations
├── internal
│   ├── config                # Configuration management
│   ├── database              # Database connections and pooling
│   ├── graphql               # GraphQL gateway and resolvers
│   ├── handlers              # Presentation layer (REST handlers)
│   ├── middleware            # HTTP middlewares (Auth, metrics, tracing)
│   ├── models                # Domain models (Metadata schemas)
│   ├── services              # Business logic (Builders, Engines, Managers)
│   └── tracing               # OpenTelemetry initialization
├── pkg
│   ├── logger                # Zap structured logging wrapper
│   └── response              # Standardized API responses
└── tests                     # Unit, Integration, and UAT tests
```

## CMS Control Plane & Service Builder

Managed under `/api/v1/cms/` routes and handled by `internal/services/service_service.go` and `internal/services/dbconn_service.go`.
- Manages external database connections and connection pooling.
- Creates services dynamically with fields (string, int, float, bool, uuid, json, array, timestamp) and handles relational mapping.

## CRUD Engine & Query Engine

Managed by `internal/services/dynamic_data_service.go`.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM operations on the fly.
- Supports filtering, sorting, nested join queries, and pagination.
- Role-Based Access Control and row-level filtering applied dynamically.

## Schema Migration Engine

Managed by `internal/services/migration_service.go`.
- Automatically applies safe schema changes (Add, Drop, Rename Column, Change Type).
- Integrates automatic backups before migrations and rollback capabilities.

## GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- Provides an optional GraphQL endpoint (`/api/v1/graphql`) using `gqlgen`.
- Automatically generated from dynamic service schemas. Supports queries, mutations, and relations.

## Backup System

Managed by `internal/services/backup_service.go`.
- Takes database table snapshots and handles schema restoration. Supports JSON format.

## Observability Integration

- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.

## Unit Testing

Run unit and integration tests using standard Go tools or the Makefile.
- Command: `make test-coverage` or `go test -v -race ./tests/...`
- Test areas include Handlers, Services, and Repositories with over 80% coverage.

## Docker Setup

Containerize using the provided `Dockerfile` and `docker-compose.yml`.
- `docker-compose up -d` to start the backend alongside required infrastructure like PostgreSQL, Redis (if applicable), and observability tools.

## Swagger Documentation

Automatically generated OpenAPI specs using swag:
- Command: `make swagger` or `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`
- Documentation available in the `docs/` directory.
