# Dynamic CMS + API Builder Platform

## System Architecture Explanation

The Backend Platform is a self-hosted, Go-native Backend-as-a-Service (BaaS) following **Clean Architecture** and **Domain-Driven Design (DDD)**.

### Core Layers:
- **Presentation Layer**: The Gin HTTP Router and Handlers mapping requests to internal services, including GraphQL endpoints via `gqlgen`.
- **Application Layer**: Business logic (Services) controlling data access, dynamically generating schemas, and managing external database connections.
- **Domain Layer**: The data models defining core entities (`Service`, `Field`, `DatabaseConnection`, `User`, `Role`, `Migration`, `Backup`).
- **Infrastructure Layer**: Cross-cutting tools including connection caching, logging (`zap`), tracing (`OpenTelemetry`), and metrics (`Prometheus`).

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

The core metadata is stored in a relational database (e.g., PostgreSQL/SQLite) containing the platform configurations. Key tables include:

- `database_connections`: External DB configurations (id, name, type, host, port, credentials).
- `services`: User-created data models referencing a `database_connection_id` and tracking table names.
- `fields`: Service attributes mapping to columns (id, service_id, name, type, nullable, unique, default_value, index).
- `service_permissions`: RBAC mapping roles to service capabilities (e.g., CanCreate, CanRead).
- `migrations`: DDL execution tracking for rollbacks.
- `backups`: Stores snapshot records.
- `users` and `roles`: Platform authentication and authorization.

## Go Project Structure

The project uses a clean modular structure:

```text
.
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── api/             # API Gateway logic
│   ├── config/          # Configuration management
│   ├── database/        # Database Connection Manager
│   ├── graphql/         # GraphQL Gateway
│   ├── handlers/        # Gin HTTP Handlers (CMS Control Plane)
│   ├── middleware/      # Interceptors (Auth, Logging, Metrics)
│   ├── models/          # Domain Models
│   ├── services/        # Service Builder, CRUD Engine, Query Engine, Migrations, Backups
│   └── tracing/         # OpenTelemetry Configuration
├── pkg/
│   ├── logger/          # Zap Logger Setup
│   ├── response/        # Standard API Response Formatter
│   └── ...
├── tests/
│   ├── unit/            # Unit tests isolated with SQLite
│   ├── integration/     # Integration tests
│   └── uat/             # User Acceptance Testing
├── docker/              # Docker configuration files
├── Dockerfile           # Multi-stage production Docker build
├── Makefile             # Development tasks
└── README.md
```

## CRUD Engine Implementation

The Dynamic CRUD Engine maps REST endpoints to standard GORM database operations dynamically.
When a user defines a service, standard endpoints become immediately functional:
- `POST /api/v1/data/{slug}` (Create)
- `GET /api/v1/data/{slug}/{id}` (Read)
- `PUT /api/v1/data/{slug}/{id}` (Update)
- `DELETE /api/v1/data/{slug}/{id}` (Delete)
- `GET /api/v1/data/{slug}` (List/Query Engine)

This engine automatically translates generic JSON inputs into database records according to the dynamically defined fields and constraints.

## Schema Migration Engine

The Schema Migration engine supports safe schema changes on the linked external databases.
It supports adding, dropping, renaming columns, and changing types using GORM’s Migrator.
It automatically triggers when a service model is updated to reflect schema changes and maintains a log in the `migrations` metadata table for safety and rollbacks.

## GraphQL Gateway

A GraphQL Gateway (`internal/graphql/`) dynamically generates an optional schema using `gqlgen`. It maps defined services to GraphQL Queries and Mutations, allowing clients to fetch nested data dynamically via a single `POST /api/v1/graphql` endpoint.

## Backup System

The Backup Engine provides snapshot creation and restoration.
- Supports generic service record backups mapping table contents to JSON snapshots.
- Tracks snapshots within the `backups` metadata table.
- Can restore data via JSON payload decoding into the dynamic target tables.

## Observability Integration

The platform provides robust observability across layers:
- **Logging**: Structured Zap logging (`pkg/logger/`) configured with correlation IDs.
- **Metrics**: Prometheus interceptors (`internal/middleware/prometheus.go`) tracking HTTP latency and status codes, exported on `/metrics`.
- **Tracing**: OpenTelemetry/Jaeger initialized (`internal/tracing/`) to wrap SQL execution spans and network transactions.

## Unit Tests

The system maintains a minimum of 80% coverage across Repository, Service, and API handlers. Tests are executed seamlessly using SQLite as an in-memory test database via `github.com/glebarez/sqlite`.
To run tests:
`make test-coverage` or `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`

## Docker Setup

The platform is fully containerized. A multi-stage `Dockerfile` compiles the Go application natively. `docker-compose.yml` provides orchestrated spin-ups mapping out dependencies like PostgreSQL (Metadata), Jaeger (Tracing), and Prometheus (Metrics) cleanly.

## Swagger Documentation

API Documentation is auto-generated utilizing `swaggo/swag`.
Generate the documentation by running:
`swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`
The Swagger UI is then mounted to expose control-plane routing logic directly to users.
