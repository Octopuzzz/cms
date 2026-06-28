# Go Backend Platform (Dynamic CMS + API Builder)

This repository contains a **production-grade Backend Platform** (Dynamic CMS + API Builder) written in Go. The system operates similarly to platforms like Hasura or Supabase, acting as a fully self-hosted, Go-native backend infrastructure generator. It follows **Clean Architecture and Domain-Driven Design (DDD)**.

## High-Level System Architecture

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

The platform metadata is stored in a relational database (PostgreSQL by default). Core tables include:

- **`database_connections`**: Stores external DB configs (host, port, credentials).
- **`services`**: Represents dynamic data models created by users.
- **`fields`**: Defines attributes (type, unique, nullable, default) for each service.
- **`service_permissions`**: Role-Based Access Control (RBAC) connecting roles to services.
- **`migrations`**: Tracks safe schema changes and rollback history.
- **`backups`**: Stores snapshots and schema outputs.
- **`users` & `roles`**: System authentication and authorization.
- **`audit_logs`**: Logs all structural and data-level changes.

## Go Project Structure

The project follows a modular Go structure:
```
.
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── api/             # API Gateway logic (if any)
│   ├── config/          # Configuration management
│   ├── database/        # Database Connection Manager
│   ├── graphql/         # GraphQL Gateway
│   ├── handlers/        # HTTP presentation layer (Gin)
│   ├── middleware/      # Auth, Rate Limiter, Telemetry
│   ├── models/          # Domain layer (Data models)
│   ├── services/        # Application layer (Business logic)
│   └── tracing/         # OpenTelemetry setup
├── pkg/
│   ├── logger/          # Zap structured logging
│   └── response/        # Standardized API responses
├── tests/               # Unit, Integration, and UAT tests
└── docker/              # Containerization setups (Dockerfile, Compose)
```

## Core Components

### CRUD Engine Implementation
Handled by the **Dynamic Data Service**, this engine maps dynamic API endpoints (e.g., `GET /api/v1/data/{slug}`) to GORM operations. It supports automated filtering, dynamic JOINs via foreign key mappings, sorting, and pagination.

### Schema Migration Engine
Tracks and applies schema differences such as adding, dropping, or renaming columns dynamically using GORM's `.Migrator()`. Safety mechanisms include pre-migration backups and automatic rollbacks.

### GraphQL Gateway
Automatically generates GraphQL endpoints via `gqlgen`. It translates dynamic service schemas into queries, mutations, and relations, providing an optional schema-less API gateway accessible via `POST /api/v1/graphql`.

### Backup System
Generates table and schema snapshots. Allows users to export tables as JSON or SQL formats, and supports restoring rows dynamically.

### Observability Integration
The platform offers full observability:
- **Logging**: Structured request and error logs using Zap (`pkg/logger/`).
- **Tracing**: OpenTelemetry (integrated with Jaeger) wrapping SQL and network calls (`internal/tracing/`).
- **Metrics**: Prometheus metrics via HTTP middleware to track request latency, query speed, and errors.

## Testing & Quality

The system aims for >80% test coverage spanning Repository, Service, and API handlers. Tests are run locally or in CI via:
```bash
make test-unit
make test-integration
make test-coverage
# or manually:
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Docker Setup

The platform is fully containerized. A `docker-compose.yml` file provisions the Go application alongside essential services (Database, Redis, Prometheus).
Run the platform with:
```bash
docker-compose up -d
```

## Swagger Documentation

API documentation is generated using Swag. To generate or update the docs:
```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
