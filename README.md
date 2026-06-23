# Dynamic CMS & API Builder

A self-hosted, Go-native Backend-as-a-Service (BaaS) platform that dynamically generates REST and GraphQL APIs.

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

## System Architecture Explanation

The system follows **Clean Architecture and Domain Driven Design**.
- **Presentation Layer**: Handlers and Routers (`internal/handlers`, `internal/graphql`) using Gin and gqlgen.
- **Application Layer**: Business logic (`internal/services`).
- **Domain Layer**: Models (`internal/models`).
- **Infrastructure Layer**: Database connectors (`internal/database`), Configuration, Tracing, and Logging.

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. Tables defined include:

- `database_connections`: id, name, type, host, port, username, password, database_name.
- `services`: id, name, database_connection_id, db_table_name, created_at, updated_at.
- `fields`: id, service_id, name, type, nullable, unique, default_value, index.
- `service_permissions`: Connects `Role` to `Service` providing access control.
- `migrations`: Keeps track of DDL executions.
- `backups`: Stores backup snapshots.
- `users` and `roles`: General authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of modifications.

## Go Project Structure

The project uses a clean modular structure:

```
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── api/             # API configuration
│   ├── config/          # Application configuration
│   ├── database/        # Database Connection Manager
│   ├── graphql/         # GraphQL Gateway
│   ├── handlers/        # Presentation Layer (HTTP Handlers)
│   ├── middleware/      # Middleware (Auth, Rate Limiting, Prom)
│   ├── models/          # Domain Layer
│   ├── services/        # Application Layer (CMS Control Plane, Service Builder, CRUD Engine, Query Engine)
│   └── tracing/         # OpenTelemetry Setup
├── pkg/
│   └── logger/          # Zap structured logging
├── tests/               # Unit, integration, and UAT tests
├── docker/              # Docker setup
├── docker-compose.yml   # Multi-container orchestrator
└── Makefile             # Task automation
```

## CRUD Engine Implementation

Managed by `DynamicDataService` (in `internal/services/dynamic_data_service.go`).
It automatically generates endpoints for dynamically created services:
- `POST /api/v1/data/{slug}` - Create
- `GET /api/v1/data/{slug}/{id}` - Read
- `PUT /api/v1/data/{slug}/{id}` - Update
- `DELETE /api/v1/data/{slug}/{id}` - Delete
- `GET /api/v1/data/{slug}` - List

## Query Engine

The list endpoint `GET /api/v1/data/{slug}` supports advanced queries:
- **Filtering**: e.g. `?email=john@example.com`
- **Sorting**: e.g. `?sort=created_at:desc`
- **Pagination**: e.g. `?page=1&limit=20`
- **Joins**: e.g. `?join=user,products`

## Schema Migration Engine

Managed by `MigrationService` (in `internal/services/migration_service.go`).
- Supports safe schema changes: Add column, Drop column, Rename column, Change column type.
- Keeps track of migration status natively into the `migrations` table and handles rollbacks.

## GraphQL Gateway

Managed by `internal/graphql/gateway.go` utilizing `gqlgen`.
- Automatically generates optional schema mutations and queries for defined data models.
- Exposed at `POST /api/v1/graphql`.
- Includes a Playground at `GET /api/v1/graphql/playground`.

## Backup System

Managed by `BackupService` (in `internal/services/backup_service.go`).
- Supports table backups, schema backups, and service snapshots.
- Endpoints:
  - `POST /api/v1/cms/backup/service/{service_id}`
  - `POST /api/v1/cms/restore/{id}`

## Observability Integration

- **Logging**: Zap structured logging with correlation/request IDs.
- **Metrics**: Prometheus metrics exported at `/metrics`.
- **Tracing**: OpenTelemetry/Jaeger initialized in `internal/tracing/`.

## Security

Implements security best practices:
- Input validation
- SQL injection protection (using GORM parameterized queries)
- JWT authentication
- Rate limiting

## Unit Tests

Run tests using the Makefile to achieve over 80% coverage:

```bash
make test-unit
# To run full test suite with coverage
make test-coverage
```

## Docker Setup

Run the entire stack with Docker Compose:

```bash
docker-compose up -d
```

## Swagger Documentation

Generate and access OpenAPI/Swagger documentation. To generate:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

Swagger UI is accessible at `GET /api/v1/swagger/index.html`.
