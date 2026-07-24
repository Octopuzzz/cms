# Dynamic CMS + API Builder Backend Platform

## Overview
This is a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. It works similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native. The platform allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

The system is modular, scalable, and cloud-ready, acting as a backend infrastructure generator.

## Core Technology Requirements
- **Language**: Go
- **API Layer**: REST (default), GraphQL (optional via `gqlgen`)
- **Web Framework**: Gin
- **ORM**: GORM
- **Logging**: Zap structured logging
- **Metrics**: Prometheus
- **Tracing**: OpenTelemetry
- **Cache**: Redis
- **Database Migration**: GORM AutoMigrate / Native
- **Containerization**: Docker
- **API Documentation**: Swagger / OpenAPI

## System Architecture

The system follows **Clean Architecture and Domain-Driven Design (DDD)**.

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

### Core Layers
1. **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
2. **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
3. **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
4. **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## Metadata Database Schema

The CMS stores platform metadata in a primary database (e.g., PostgreSQL). Tables include:

- `database_connections`: Stores external DB configurations (id, name, type, host, port, credentials). Includes connection pooling settings.
- `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas.
- `fields`: Defines attributes for each service (string, integer, float, uuid, JSON). Configures uniqueness, nullability, defaults.
- `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks.
- `relations`: Defines relations between services (One-to-One, One-to-Many, Many-to-One, Many-to-Many).
- `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- `backups`: Stores snapshot records or schema outputs.
- `users` & `roles`: General authentication and authorization.
- `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure:

```text
.
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── config/          # Environment and configuration
│   ├── database/        # Database connection manager
│   ├── graphql/         # GraphQL gateway and resolvers
│   ├── handlers/        # HTTP presentation layer (API)
│   ├── middleware/      # Auth, Rate Limiter, Prometheus
│   ├── models/          # Domain models (Metadata DB schema)
│   ├── services/        # Application business logic (CMS, CRUD Engine, Backup, Migration)
│   └── tracing/         # OpenTelemetry configuration
├── pkg/
│   ├── logger/          # Zap structured logging
│   └── response/        # Standardized API responses
├── tests/
│   ├── integration/     # Integration tests
│   ├── uat/             # User Acceptance Testing
│   └── unit/            # Unit tests
├── docker/              # Docker configurations
├── docs/                # Swagger generated documentation
└── Makefile             # Make targets for build, test, docker
```

## CRUD Engine & Query Engine Implementation

Managed by `internal/services/dynamic_data_service.go`, the system automatically generates CRUD endpoints when a service is created:
- `POST /api/v1/data/{slug}` (Create)
- `GET /api/v1/data/{slug}/{id}` (Read)
- `PUT /api/v1/data/{slug}/{id}` (Update)
- `DELETE /api/v1/data/{slug}/{id}` (Delete)
- `GET /api/v1/data/{slug}` (List / Query)

### Query Capabilities
- **Filtering**: `?email=john@example.com`
- **Sorting**: `?sort=created_at:desc`
- **Pagination**: `?page=1&limit=20`
- **Join Queries**: Dynamic joining via foreign key mappings (e.g., `?join=user,products`)

## Schema Migration Engine

Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column, Change column type) using GORM’s `.Migrator()`.
- Provides automatic backup before migration and rollback capability.
- Logs migration status natively into the `migrations` table.

## GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- Automatically generates generic schemas from service definitions using `gqlgen`.
- Exposes `POST /api/v1/graphql`.
- Supports queries, mutations, and relations natively based on the defined data models.

## Backup System

Managed by `internal/services/backup_service.go`.
- Generates data snapshots (table backup, schema backup, service snapshot).
- Supports formats like SQL dump and JSON snapshot.
- Endpoints allow triggering backups (`POST /cms/backup/service/{service_id}`) and restoring data (`POST /cms/restore/service/{service_id}`).

## Observability Integration

Implemented across the stack:
- **Logging**: Request, error, and query logs via Zap (`pkg/logger/`), injected with correlation and trace IDs.
- **Tracing**: OpenTelemetry/Jaeger initialized in `internal/tracing/` wraps SQL commands and network logic.
- **Metrics**: Prometheus tracks request latency, error rates, and slow queries via standard HTTP interceptors (`internal/middleware/prometheus.go`), exposed at `/metrics`.

## Unit Tests

The system maintains a minimum of 80% coverage across Repository, Service, and API handlers.
- Run tests: `go test -v -race ./tests/...`
- Or using Makefile: `make test-unit`, `make test-coverage`.

## Docker Setup

- Start services (DBs, Redis, App) via Docker Compose: `docker-compose up -d`
- Stop services: `docker-compose down`
- Dockerfile provided for the main Go backend service for modular deployment.

## Swagger Documentation

Generate Swagger API documentation using the provided Makefile command:
```bash
make swagger
```
Or manually:
```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
