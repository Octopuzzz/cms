# CMS Backend Platform

This repository contains a self-hosted, Go-native Backend-as-a-Service (BaaS) platform (Dynamic CMS + API Builder). It dynamically generates REST and GraphQL APIs, providing capabilities to create backend services, manage database connections, define schemas, automatically handle CRUD operations, and provide robust observability and scalable cloud deployment.

## System Architecture

The system follows Clean Architecture and Domain-Driven Design (DDD) principles to ensure modularity and scalability.

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

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`), acting as the API gateway. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Manages access, dynamic schemas, and operational handlers.
- **Domain Layer**: The data models (`internal/models/`).
- **Infrastructure Layer**: Cross-cutting tools (connection caching, telemetry with OpenTelemetry, Prometheus metrics, and Zap logging).

## Metadata Database Schema

The core metadata configuration is stored primarily in PostgreSQL (or SQLite/MySQL depending on configuration) and maps platform metadata including:

- `database_connections`: External DB configurations (id, name, type, host, credentials, pooling settings).
- `services`: User-created data models (mapped to database connections, tracks schema/table name).
- `fields`: Field attributes (type, nullability, uniqueness, defaults).
- `service_permissions`: RBAC connectivity (links `Role` to `Service`).
- `migrations`: Tracked DDL operations for history and rollbacks.
- `backups`: System or snapshot records.
- `users` / `roles`: Authentication and control plane authorization.
- `audit_logs`: Detailed operational logging.

## Go Project Structure

The project implements a clean modular structure:

```
cmd/
  server/         # Main entry point and server startup
internal/
  config/         # Environment and application configuration
  database/       # Connection pooling and GORM database connections
  graphql/        # GraphQL gateway implementation
  handlers/       # Gin HTTP handlers (REST API)
  middleware/     # Authentication, CORS, Rate Limiting, Prometheus
  models/         # Core Domain Models
  services/       # Application logic (Control Plane, Builders, Engines)
  tracing/        # OpenTelemetry integration
pkg/
  logger/         # Zap structured logging
  response/       # Standardized JSON response utilities
tests/
  unit/           # Unit tests
  integration/    # Integration tests
  uat/            # End-to-end / User Acceptance Tests
docker/           # Docker-specific configuration scripts
```

## Core Engines Implementation

### CRUD & Query Engine
Managed by `internal/services/dynamic_data_service.go`. It maps REST API endpoints to dynamic GORM database queries. Supports querying with filtering (`?email=test@example.com`), sorting, pagination (`?page=1&limit=20`), and dynamically validates RBAC permissions.

### Schema Migration Engine
Managed by `internal/services/migration_service.go`. Offers safe schema alterations like Add/Drop/Rename Column via GORM's automatic `.Migrator()`. Includes history tracking and rollback functionality via snapshot retention.

### GraphQL Gateway
Automatically configured in `internal/graphql/gateway.go`. Generates standard Queries and Mutations over the dynamic schemas leveraging `github.com/99designs/gqlgen/graphql`, exposing the endpoint at `/api/v1/graphql`.

### Backup System
Handled by `internal/services/backup_service.go`. Implements snapshot-style backup functionality, allowing users to back up specific services or schemas via JSON payloads and restore row-level data as needed.

## Observability Integration

Comprehensive tracing, logging, and metrics are included:
- **Logging**: Zap structured logging is utilized, with context propagation (Trace ID and Request ID).
- **Metrics**: Prometheus middleware intercepts routes. Exported at `/metrics`.
- **Tracing**: OpenTelemetry (Jaeger) is integrated across the database and HTTP handler boundaries.

## Security Features
- JWT-based authentication.
- Robust Role-Based Access Control (RBAC) via the metadata tables.
- Standard CORS policy restrictions.
- Global rate-limiting middleware configurable via environment variables.

## Unit Tests

Unit tests, integration, and UAT tests are bundled to cover over 80% coverage.
Execute with the following command:
```bash
make test-coverage
```
Unit tests utilize in-memory SQLite instances to test the generic implementation safely.

## Docker Setup

The system provides fully packaged `Dockerfile` and `docker-compose.yml` for simplified deployment:
```bash
make docker-up
```
Containerizes the Go binary along with any specified associated databases.

## Swagger Documentation

To generate Swagger API documentation dynamically from the handler comments:
```bash
make swagger
```
This produces the `docs/` folder, and the Swagger UI is served at `/api/v1/swagger/index.html`.
