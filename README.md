# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade, self-hosted Backend Platform and API Builder written in Go. It allows developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

It functions similarly to platforms like Hasura or Supabase but is fully Go-native, modular, scalable, and cloud-ready.

## System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles.

### Core Layers:
1. **Presentation Layer**: Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
2. **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
3. **Domain Layer**: Data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
4. **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

### High Level Architecture
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

The core internal configuration is stored in the **Metadata Database**. Tables include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead).
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure.

```
cmd/
  server/
    main.go           # Application entry point
internal/
  config/             # Configuration management
  database/           # Connection manager & DB utilities
  graphql/            # GraphQL Gateway and schema generation
  handlers/           # REST API endpoints (Presentation Layer)
  middleware/         # Auth, Rate Limiting, Metrics
  models/             # Domain Models (Metadata DB schemas)
  services/           # Business logic (Application Layer)
  tracing/            # OpenTelemetry integration
pkg/                  # Reusable components (e.g., logger, errors)
tests/                # Unit, integration, and UAT tests
docker/               # Docker configurations
```

## Core Components Implementation

### CMS Control Plane & Service Builder
Provides an API interface to store service and field definitions in the metadata database, validates foreign database connectivity, and triggers migrations on service model updates to reflect schema changes. Handles databases, services, and relations.

### Dynamic CRUD Engine & Query Engine
Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly. Applies automated filtering via query params (e.g., `?email=test@example.com`), dynamic joining via foreign key mappings, pagination logic, and sorting. Verifies RBAC dynamically prior to queries.

### Schema Migration Engine
Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`. Logs migration status natively into the `migrations` table and handles rollbacks through snapshot retention logic.

### GraphQL Gateway
Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`. Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

### Backup System
Generates data snapshots. Supports table backup, schema backup, and service snapshots. Implements generic service record backup features mapping dynamic table contents to snapshots, and supports restoring rows via JSON payload decoding.

### Observability Integration
Implemented across the stack:
- **Logging**: Zap structured logging injected globally with correlation and trace IDs tied directly into Context.
- **Tracing**: OpenTelemetry/Jaeger initialized to wrap SQL commands and network logic.
- **Metrics**: Prometheus metrics via standard HTTP interceptors tracking request latency, slow queries, and error rates. Exported at `/metrics`.

### Security
Implements input validation, SQL injection protection, JWT authentication, rate limiting, and Role-Based Access Control (RBAC).

## Unit Testing

The system includes a comprehensive suite of unit tests for Repositories, Services, and API Handlers. Target minimum coverage is 80%.
Tests use an in-memory SQLite database setup.

Run the test suite:
```bash
go test ./...
```
Generate coverage report:
```bash
make test-coverage
# or
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Docker Setup

The platform is containerized. Use Docker Compose to spin up the entire stack, including PostgreSQL, Redis, Jaeger, and Prometheus.
```bash
docker-compose up -d
```
See `docker-compose.yml` and `Dockerfile` for configurations.

## API Documentation

API documentation is generated using Swagger / OpenAPI.

Generate the Swagger documentation:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
