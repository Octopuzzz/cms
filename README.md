# Backend Platform (Dynamic CMS + API Builder)

This platform is a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go, functioning similarly to platforms like Hasura or Supabase but fully self-hosted and Go-native. The platform acts as a backend infrastructure generator allowing users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

The system is modular, scalable, and cloud-ready. It follows Clean Architecture and Domain Driven Design.

## System Architecture

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

The core internal configuration is stored in the **Metadata Database**. The metadata tables are:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure.

```text
cmd/
  server/         # Main application entry point
internal/
  api/            # API handlers and routes
  cms/            # CMS logic
  crudengine/     # Dynamic CRUD handlers
  queryengine/    # Dynamic querying and filtering
  database/       # Database connection manager
  servicebuilder/ # Schema building logic
  schema/         # Data structures and migrations
  backup/         # Backup tools
  graphql/        # GraphQL optional gateway
  logger/         # Logging wrapper
  models/         # Domain models
  handlers/       # Web layer routers and presentation
  services/       # Application logic
  tracing/        # OpenTelemetry integration
  middleware/     # Auth, Telemetry, and metrics middleware
  config/         # Configuration variables mapping
pkg/
  pagination/     # Pagination helpers
  validation/     # Input validation wrappers
  errors/         # Standardized error formats
  response/       # Gin response helpers
tests/            # Unit, integration and UAT tests
  unit/
  integration/
  uat/
docker/           # Dockerfiles and compose setups
```

## Component Explanations

### CMS Control Plane & Service Builder
Managed under `/api/v1/cms/` routes and handled by the Service Builder (`internal/services/service_service.go` and `internal/services/dbconn_service.go`). It provides an API interface to store service and field definitions in the metadata database, utilizes the database connection engine to validate external db connectivity, and triggers migrations on service model updates to reflect schema changes.

### Dynamic CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go`. Maps generic dynamic endpoints to underlying GORM database operations on the fly. It applies automated filtering via query params, dynamic joining via foreign key mappings, pagination, and verifies Role-Based Access Control and Row-Level filtering.

### Schema Migration Engine
Managed by `internal/services/migration_service.go`. It tracks and applies safe schema changes using GORM's `.Migrator()`. It logs migration statuses and handles changes to external databases automatically when models are updated.

### GraphQL Gateway
Managed by `internal/graphql/gateway.go`. The GraphQL gateway generates generic optional schemas using `github.com/99designs/gqlgen/graphql`, allowing automatically generated and interactive access to the data models via GraphQL queries and mutations.

### Backup System
Managed by `internal/services/backup_service.go`. It generates data snapshots, mapping dynamic table contents into backup files. Supports restoring rows via JSON payload decoding.

### Observability Integration
The application uses Zap logging injected globally. OpenTelemetry with Jaeger is configured (`internal/tracing/`) for tracking operations. Prometheus metrics are collected via `internal/middleware/prometheus.go` and exported at `/metrics`.

### Unit Tests
The project features a suite of unit, integration, and UAT tests aiming for 80% coverage. Tests use standard Go testing framework and can be run with `make test-coverage` or `go test ./...`.

### Docker Setup
The platform is containerized utilizing the files in the `docker/` directory, and `Dockerfile`. It can easily be launched using `docker-compose`.

### Swagger Documentation
The project includes fully generated OpenAPI / Swagger documentation utilizing `swaggo`. The documentation can be generated through the CLI utilizing `swag init` targeting `cmd/server/main.go`.