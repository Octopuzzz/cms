# Dynamic CMS + API Builder Platform

## System Architecture Explanation

The platform is a self-hosted, Go-native Backend-as-a-Service (BaaS) that dynamically generates REST and GraphQL APIs. It follows **Clean Architecture** and **Domain-Driven Design (DDD)**.

- **Presentation Layer**: Gin HTTP Router and Handlers mapping requests to internal services, including GraphQL endpoints via `gqlgen`.
- **Application Layer**: Business logic via internal services controlling data access, dynamic schema generation, and operations.
- **Domain Layer**: Core models (`Service`, `Field`, `DatabaseConnection`, `User`, `Role`).
- **Infrastructure Layer**: Telemetry (OpenTelemetry), metrics (Prometheus), logging (Zap), and connection pooling.

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

Stored centrally, tracking platform configuration. Tables include:
- `database_connections`: External DB configs (host, port, credentials).
- `services`: User-created data models referencing DB connections.
- `fields`: Attributes for services (types, uniqueness, nullability).
- `service_permissions`: RBAC roles mapping permissions to services.
- `migrations`: DDL execution history.
- `backups`: Snapshot records.
- `users` / `roles`: Control plane authentication.
- `audit_logs`: Detailed operational logging.

## Go Project Structure

Clean modular structure reflecting best practices:
- `cmd/server/`: Main application entry point.
- `internal/`:
  - `config/`: Application configuration.
  - `database/`: Connection manager for external and metadata databases.
  - `graphql/`: Generated GraphQL schema and resolvers.
  - `handlers/`: API layer.
  - `middleware/`: Auth, tracing, metrics.
  - `models/`: Domain models.
  - `services/`: Business logic implementations.
  - `tracing/`: OpenTelemetry setup.
- `pkg/`: Reusable components (logging, response formatting).
- `tests/`: Integration, unit, and UAT tests.
- `docker/`: Docker and Prometheus configuration.

## CRUD Engine Implementation

Automatically handles `Create`, `Read`, `Update`, `Delete`, and `List` operations via dynamic routing in `internal/services/dynamic_data_service.go`. Converts requests to standard GORM operations on the fly, applies query parameter filtering (e.g. `?email=test@test.com`), dynamic joining via foreign keys, and handles RBAC prior to querying.

## Schema Migration Engine

Implemented in `internal/services/migration_service.go`. Tracks and applies schema diffs safely using GORM's `Migrator`. Supports automatic backups before migrations and handles logging of status for potential rollbacks.

## GraphQL Gateway

Powered by `github.com/99designs/gqlgen/graphql`, the gateway dynamically exposes schemas mapping to services created in the platform. Provides a central `POST /api/v1/graphql` endpoint to execute queries and mutations seamlessly across generated data models.

## Backup System

Managed by `internal/services/backup_service.go`. Generates data snapshots (table/schema/service) and provides endpoints to capture (`POST /cms/backup/service/{id}`) and restore (`POST /cms/restore/service/{id}`) records seamlessly.

## Observability Integration

Comprehensive tracing and metric collection:
- **Logging**: Zap structured logging with correlation and trace IDs.
- **Tracing**: OpenTelemetry wrapping SQL queries and network logic.
- **Metrics**: Prometheus tracking request latency and status codes via `internal/middleware/prometheus.go` at `/metrics`.

## Unit Tests

Targeting 80% minimum coverage across repositories, services, and API handlers. Tests utilize in-memory SQLite (`:memory:`) setups for mocking, executed via `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`.

## Docker Setup

Full containerization for development and production deployment via Docker and `docker-compose`. Includes application instances, Redis caching, Prometheus, and external DB provisions natively located in the `docker/` folder.

## Swagger Documentation

Automatically generated OpenAPI 3 specifications covering all CMS control plane routes and dynamically generated APIs. Built via `swag init` and exposed natively in the application router.
