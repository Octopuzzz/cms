# Dynamic CMS + API Builder Backend Platform

This repository contains a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. It acts as a backend infrastructure generator, allowing developers to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

The system is modular, scalable, cloud-ready, and follows Clean Architecture and Domain-Driven Design (DDD) principles.

## System Architecture Explanation

The platform architecture comprises several core components handling dynamic API and data layer generation.

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

- **Presentation Layer**: Handled by the Gin HTTP framework (in `internal/handlers`), generating endpoints dynamically and routing them. GraphQL is managed by `gqlgen` (in `internal/graphql`).
- **Application Layer**: Contains business logic (`internal/services`). Includes engines for dynamic data reading/writing, schema building, and migrations.
- **Domain Layer**: Models (`internal/models`) outlining entities.
- **Infrastructure Layer**: Incorporates open telemetry, logging, metrics, connection caching, and authentication middleware.

## Metadata Database Schema

The CMS must store platform metadata. The core configuration is maintained via a Postgres/SQLite database.

Core tables/models (defined in `internal/models/models.go`):

1. `database_connections`
   - `id`, `name`, `type` (postgres, mysql, mongodb), `host`, `port`, `username`, `password`, `database_name`.
2. `services`
   - `id`, `name`, `database_id`, `db_table_name`, `created_at`, `updated_at`.
3. `fields`
   - `id`, `service_id`, `name`, `type` (string, integer, float, uuid, etc.), `nullable`, `unique`, `default_value`, `index`.
4. `service_permissions`
   - Ties services to roles with granular control.
5. `migrations`
   - Logs DDL execution for safe schema tracking.
6. `backups`
   - Data and schema snapshots.
7. `users` & `roles`
   - For Control Plane RBAC.

## Go Project Structure

The project structure enforces clean boundaries and modularity:

- `cmd/server/`: The application entry point (`main.go`), setting up configs, DB connections, HTTP routers, and starting the web server.
- `internal/`: Application-specific logic.
  - `config/`: Configuration variables.
  - `database/`: Database connection manager and pooling logic.
  - `graphql/`: GraphQL gateway using `gqlgen`.
  - `handlers/`: Gin REST HTTP handlers mapping routes to application services.
  - `middleware/`: HTTP middlewares (e.g., prometheus, authentication, rate limiting).
  - `models/`: Domain entities and metadata schemas.
  - `services/`: Business operations (CRUD, Backup, Schema, Auth, Migrations).
  - `tracing/`: OpenTelemetry setup.
- `pkg/`: Reusable, cross-project utility code.
  - `logger/`: Zap structured logging.
  - `response/`: Standardized JSON API responses.
- `docker/`: Dockerfiles and compose setups.
- `tests/`: Extensive testing suites for isolated and integrated verification (`unit/`, `integration/`, `uat/`).

## CRUD Engine Implementation

The Dynamic CRUD Engine (`internal/services/dynamic_data_service.go`) implements generic operations on user-generated services.
When a service is created, it dynamically exposes CRUD endpoints at `/api/v1/data/{slug}`.

- **Create, Read, Update, Delete**: Uses `gorm`'s map-backed functions (e.g. `Table(tableName).Create(&data)`) dynamically.
- **Query Engine**: Features filtering, sorting, pagination, and join support via robust URL query parameter parsing and `gorm` clause generation.

## Schema Migration Engine

Implemented in `internal/services/migration_service.go`.
Provides safe schema modifications via the GORM migrator (`db.Migrator()`).

- **Supported Operations**: Add column, Drop column, Rename column.
- **Safety**: Integrates with the `migrations` table and handles execution logs. Recommends manual rollback strategies based on generated logs.

## GraphQL Gateway

Implemented in `internal/graphql/gateway.go`.
Automatically constructs GraphQL schemas from the registered dynamic services and generates resolver bindings dynamically for queries, mutations, and object relations via `github.com/99designs/gqlgen/graphql`.

## Backup System

Implemented in `internal/services/backup_service.go`.
Supports creating snapshots of service data and schemas to JSON or SQL dumps, allowing system-level extraction and restoration of generated content.

## Observability Integration

Ensures the platform is continuously monitored.

- **Logging**: Zap structured logging in `pkg/logger/logger.go`. Logs are attached to the request context.
- **Tracing**: OpenTelemetry config in `internal/tracing/tracing.go`, tracking latency of network and database requests via Jaeger.
- **Metrics**: Prometheus middleware (`internal/middleware/prometheus.go`) tracks latency, slow queries, and HTTP status codes, exposed on `/metrics`.

## Unit Tests

The system targets a minimum coverage of 80% with extensive suites spanning Repositories, Services, and Handlers.

- **Unit Tests**: Found in `tests/unit/services_test.go`. Uses SQLite in-memory mocking (`:memory:`).
- **Integration Tests**: In `tests/integration/service_integration_test.go`.
- **Command**: Run with `make test-coverage` or `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`.

## Docker Setup

The repository contains containerization infrastructure:

- `Dockerfile`: Or `docker/Dockerfile` provides the build step for compiling the Go executable into a lightweight production container.
- `docker-compose.yml`: Local orchestrator binding the Application, PostgreSQL/MySQL targets, Jaeger for OpenTelemetry, and Prometheus for metrics.

## Swagger Documentation

API documentation is generated dynamically via `swaggo/swag`.

- **Generation**: To generate, install `swag` and run `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
- **Access**: The UI is typically hosted at `/swagger/index.html`.
