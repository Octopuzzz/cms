# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend Platform written in Go, functioning as a fully self-hosted and Go-native dynamic CMS and API builder. It is designed similarly to platforms like Hasura or Supabase but completely native to Go.

The platform allows developers to:
- Connect databases
- Create data models dynamically
- Generate CRUD APIs automatically
- Generate optional GraphQL APIs
- Manage schema migrations
- Monitor logs and performance
- Manage backups
- Scale services

## Core Technology Stack
- **Language**: Go (latest stable version)
- **API Layer**: REST (default), GraphQL (optional via `gqlgen`)
- **Web Framework**: Gin
- **ORM**: GORM
- **Logging**: Zap structured logging
- **Metrics**: Prometheus
- **Tracing**: OpenTelemetry
- **Containerization**: Docker
- **API Documentation**: Swagger / OpenAPI

## High Level Platform Architecture

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

## System Architecture

The project follows **Clean Architecture** and **Domain-Driven Design (DDD)**:

- **Presentation Layer**: `internal/handlers/` and `internal/graphql/` for Gin HTTP handlers and GraphQL resolvers.
- **Application Layer**: `internal/services/` where the core business logic resides (CMS, dynamic data handling, backups).
- **Domain Layer**: `internal/models/` detailing core configurations like database connections, user schemas, relations, and settings.
- **Infrastructure Layer**: Cross-cutting tools in `internal/database/`, `internal/tracing/`, `internal/middleware/`, and `pkg/` (logger, errors, pagination).

## Metadata Database Schema

The CMS manages its own platform metadata within a primary database. The core tables include:
- `database_connections`: Stores remote/external DB connections (PostgreSQL, MySQL, MongoDB, etc.).
- `services`: Represents user-created data models/services linked to a database connection.
- `fields`: Details the attributes (columns, types, constraints) of the generated services.
- `service_permissions`: Connects RBAC roles to service access.
- `migrations`: Tracks schema migration states and logic.
- `backups`: Logs system backups for models and definitions.
- `users`, `roles`, `audit_logs`: Controls management, authentication, and platform auditing.

## Project Structure

```
.
├── cmd
│   └── server          # Application entrypoint
├── docker              # Docker and Docker-compose definitions
├── docs                # Auto-generated Swagger documentation
├── internal
│   ├── config          # Configuration parsing
│   ├── database        # Database connection pool manager
│   ├── graphql         # GraphQL Gateway definition
│   ├── handlers        # Gin REST API handlers
│   ├── middleware      # Gin middlewares (Auth, Prometheus, Rate Limiting)
│   ├── models          # Domain models
│   ├── services        # Core business logic services
│   └── tracing         # OpenTelemetry setup
├── pkg
│   ├── errors          # Custom error types
│   ├── logger          # Zap Logger configuration
│   ├── pagination      # Query pagination helpers
│   └── response        # Standardized API response wrappers
├── tests               # Unit and integration tests
├── Makefile            # Standard tasks (build, dev, test)
└── go.mod
```

## Core Components

### CRUD & Query Engine
When a service is configured, the Dynamic CRUD Engine (`internal/services/dynamic_data_service.go`) instantly handles standard mapping to REST endpoints (`GET, POST, PUT, DELETE /api/v1/data/{slug}`).
The Query Engine supports nested joining, advanced pagination, and dynamic filtering using query strings (`?sort=created_at:desc`, `?email=x@y.z`).

### Schema Migration Engine
Safe migrations are managed via `internal/services/migration_service.go`. The engine translates service schema updates into GORM's `Migrator()` functions to transparently alter external schema definitions in target DBs.

### GraphQL Gateway
Integrated using `gqlgen`, the system parses current schemas to expose an optional standard GraphQL API endpoint.

### Backup System
`internal/services/backup_service.go` allows users to capture data snapshots of managed services and export/import via structured formats (like JSON and internal DB mapping dumps).

### Observability
Fully observable out of the box with open-standard traces (OpenTelemetry), metrics tracking (Prometheus exported to `/metrics`), and structured structured JSON logging (Zap).

## Docker Setup

The platform is designed to be fully containerized. A `docker-compose.yml` is provided at the root which sets up the backend API, Postgres DB for metadata, and Jaeger for OpenTelemetry tracing.

Run via:
```bash
docker-compose up -d
```

## Swagger Documentation

Swagger API documentation is automatically generated. The standard workflow to re-generate the docs after modifying handler comments is:

```bash
make swagger
# or
~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

## Tests

The project aims for comprehensive coverage (>80%) across the `internal/` architecture. Use standard Go testing or the included Makefile target.

```bash
make test-coverage
# Runs: go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```
