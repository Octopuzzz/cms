# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend Platform (Dynamic CMS + API Builder) written natively in Go. It operates similarly to platforms like Hasura or Supabase but is fully self-hosted. The platform allows users to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

The system is designed to be **clean, modular, scalable, and cloud-ready**, following **Clean Architecture and Domain-Driven Design (DDD)**.

## System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

### High-Level Architecture

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
PostgreSQL    MySQL       MongoDB
```

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**, which can be hosted on PostgreSQL.
Key tables and components:

1. **`database_connections`**: Stores external DB configurations with fields for `id`, `name`, `type`, `host`, `port`, and credentials. Includes connection pooling settings.
2. **`services`**: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`). Fields include `id`, `name`, `database_id`, `created_at`, `updated_at`.
3. **`fields`**: Defines attributes for each service, such as string, integer, float, boolean, uuid, json, array, timestamp. Configures uniqueness, nullability, defaults, and indexing.
4. **`relations`**: Defines relationships between tables (One-to-One, One-to-Many, Many-to-One, Many-to-Many with automatic join tables).
5. **`service_permissions`**: Connects `Role` to `Service` providing fine-grained access checks.
6. **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
7. **`backups`**: Stores snapshot records or schema outputs.
8. **`users` and `roles`**: Authentication and authorization tables.
9. **`audit_logs`**: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure.

```text
├── cmd
│   └── server          # Application entry point
├── internal            # Core business logic and internal packages
│   ├── api             # API routing setup
│   ├── config          # Application configuration
│   ├── database        # Connection manager and caching
│   ├── graphql         # GraphQL gateway and schemas
│   ├── handlers        # HTTP controllers/handlers
│   ├── middleware      # Gin middlewares (Auth, Rate Limiting, Prometheus)
│   ├── models          # Domain models
│   ├── services        # Application services (CRUD Engine, CMS, Migrations, Backups)
│   └── tracing         # OpenTelemetry tracing setup
├── pkg                 # Reusable public packages
│   ├── errors          # Error handling
│   ├── logger          # Zap structured logger
│   ├── pagination      # Pagination logic
│   └── validation      # Validation logic
├── tests               # Test suites
│   ├── integration     # Integration tests
│   ├── uat             # User Acceptance Testing
│   └── unit            # Unit tests
└── docker              # Docker and deployment configurations
```

## CRUD Engine Implementation

When a service is created, the system automatically generates CRUD endpoints mapped to standard GORM database operations on the fly.
Managed by `internal/services/dynamic_data_service.go` and `internal/handlers/dynamic_data_handler.go`.

**Generated Endpoints:**
- **Create**: `POST /api/v1/data/{slug}`
- **Read**: `GET /api/v1/data/{slug}/{id}`
- **Update**: `PUT /api/v1/data/{slug}/{id}`
- **Delete**: `DELETE /api/v1/data/{slug}/{id}`
- **List**: `GET /api/v1/data/{slug}`

**Query Engine Capabilities:**
- **Filtering**: `GET /api/v1/data/users?email=john@example.com`
- **Sorting**: `GET /api/v1/data/users?sort=created_at:desc`
- **Pagination**: `GET /api/v1/data/users?page=1&limit=20`
- **Join Queries**: `GET /api/v1/data/orders?join=user,products` (with nested joins support)

## Schema Migration Engine

Tracks and applies safe schema changes using GORM's `.Migrator()`.
Managed by `internal/services/migration_service.go`.

**Supported Operations:**
- Add column
- Drop column
- Rename column
- Change column type

**Safety Features:**
- Automatic backup before migration
- Native status logging in the `migrations` table
- Rollback capabilities through snapshot retention

## GraphQL Gateway

Generates generic optional schemas dynamically using `github.com/99designs/gqlgen/graphql`.
Managed by `internal/graphql/gateway.go`.

Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

**Features Supported:**
- Queries
- Mutations
- Relations resolving

Example Query:
```graphql
query {
  users {
    id
    name
    email
  }
}
```

## Backup System

Generates data snapshots and manages table backups.
Managed by `internal/services/backup_service.go`.

**Supported Features:**
- Table backup
- Schema backup
- Service snapshot (JSON format)

**API Endpoints:**
- **Backup**: `POST /api/v1/cms/backup/service/{service_id}`
- **Restore**: `POST /api/v1/cms/restore/service/{service_id}` (restores rows via JSON payload decoding)

## Observability Integration

Integrated deeply to ensure system health and monitoring.

- **Logging**: Zap structured logging (`pkg/logger/`) with correlation and trace IDs tied directly into the Context.
  - Includes Request logs, Error logs, and Query logs.
- **Tracing**: OpenTelemetry (`internal/tracing/`) integrated with Jaeger. Wraps SQL commands and network logic.
- **Metrics**: Prometheus tracking via `internal/middleware/prometheus.go`. Exports Request latency, Slow queries, and Error rate at `/metrics`.

## Unit Tests

The system maintains high quality via comprehensive testing (Unit, Integration, and UAT).
- Minimum Coverage: 80%
- Covered Layers: Repository, Service, API handlers.

**Command to run tests:**
```bash
go test ./...
```
*(For complete coverage reporting including isolated packages, run: `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`)*

## Docker Setup

The platform is containerized using Docker and Docker Compose.
Configuration is in `docker-compose.yml` and `Dockerfile`.

Services included in the stack:
- **cms-backend**: The Go application
- **postgres**: The metadata database
- **redis**: Caching for performance
- **jaeger**: Distributed tracing backend
- **prometheus**: Metrics scraping

Run the stack:
```bash
docker-compose up -d
```

## Swagger Documentation

API documentation is generated using Swaggo.

To generate/update documentation:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
The documentation is served locally and provides interactive API exploring. (Note: `docs/` is ignored from version control).
