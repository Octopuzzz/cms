# Backend Platform (Dynamic CMS + API Builder)

A production-grade Backend Platform and Dynamic CMS written in Go. This platform acts as a backend infrastructure generator, similar to Hasura or Supabase, but is fully self-hosted and Go-native. It allows users to dynamically create backend services, schemas, and automatically generated APIs (REST and optionally GraphQL).

## System Architecture Explanation

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles, separating concerns into clearly defined layers to ensure modularity, scalability, and maintainability:

- **Presentation Layer**: The API Gateway, consisting of Gin HTTP Router and Handlers (`internal/handlers/`). This layer manages incoming REST API requests and GraphQL endpoints using `gqlgen`.
- **Application Layer**: Contains business logic (`internal/services/`). Services here control data access, generate dynamic schemas, and perform complex operations requested by the handlers.
- **Domain Layer**: The data models (`internal/models/`). It defines core entities that are central to the platform's operation such as `Service`, `Field`, `DatabaseConnection`, `User`, and `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as database connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), logging (`pkg/logger/`), and external integrations.

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

The CMS stores platform metadata in the Metadata Database. Core tables include:

- `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
- `services`: Represents user-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
- `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
- `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
- `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- `backups`: Stores snapshot records or schema outputs.
- `users` and `roles`: General authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure:

```
.
├── cmd
│   └── server                # Application entrypoint
├── internal
│   ├── config                # Configuration handling
│   ├── database              # Database connection manager & pooling
│   ├── graphql               # GraphQL gateway and schemas
│   ├── handlers              # HTTP API handlers (Presentation layer)
│   ├── middleware            # HTTP middlewares (auth, observability, etc)
│   ├── models                # Domain layer models
│   ├── services              # Application layer services
│   └── tracing               # OpenTelemetry integration
├── pkg
│   └── ...                   # Reusable library packages (e.g., logger)
├── tests                     # Unit, integration, and UAT tests
├── docker                    # Docker configurations and setup
├── .env.example              # Environment variables template
├── Dockerfile                # Platform image definition
├── Makefile                  # Build, test, and run scripts
└── go.mod                    # Go module dependencies
```

## CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`, the Dynamic CRUD Engine automatically exposes endpoints (Create, Read, Update, Delete, List) for registered services.
- It maps dynamic REST endpoints (`/api/v1/data/{slug}`) to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g., `?email=test@example.com`), dynamic joining via foreign key mappings, and pagination logic.
- Enforces Role-Based Access Control and Row-Level security policies automatically before query execution.

## Schema Migration Engine

Managed by `internal/services/migration_service.go`, the system supports safe schema changes dynamically:
- Translates service updates into database schema modifications (Add column, Drop column, Rename column).
- Utilizes GORM’s `.Migrator()` functionality to enact DDL statements against external databases.
- Logs migration status natively into the `migrations` table, providing features to review applied migrations and potential rollback targets.

## GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- The system automatically generates GraphQL schemas from registered metadata services utilizing `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` allowing clients to query dynamically generated models.
- Integrates authorization middleware and translates nested GraphQL queries into optimized database joins to circumvent N+1 problems.

## Backup System

Managed by `internal/services/backup_service.go`.
- Provides functionality to capture schema structural formats and service snapshots (data rows).
- Includes snapshot and restore functionalities. Snapshots map dynamic table contents into robust JSON outputs.
- Records of backups, alongside statuses, are preserved inside the `backups` control plane table.

## Observability Integration

The platform includes state-of-the-art observability systems:
- **Logging**: Zap structured logging is injected globally (`pkg/logger/`). Request contexts carry correlation and trace IDs down to the database layers for unified log streams.
- **Tracing**: OpenTelemetry (`internal/tracing/`) wraps database connections and HTTP interceptors, capable of exporting spans directly to tools like Jaeger.
- **Metrics**: Prometheus instrumentation is defined in `internal/middleware/prometheus.go`, collecting vital statistics on latency, memory usage, query times, and status code distributions. Exposed on `/metrics`.

## Unit Tests

The repository maintains an extensive testing suite aimed at >80% code coverage across vital systems (Repository, Service, API Handlers).
Tests utilize an isolated in-memory SQLite setup via `github.com/glebarez/sqlite` to prevent cross-contamination.

Commands available in the `Makefile`:
- Unit Tests: `make test-unit`
- Integration Tests: `make test-integration`
- All Tests (with Coverage): `make test-coverage`

## Docker Setup

The repository is cloud-ready and includes Dockerization for rapid provisioning:
- The system logic builds into a minimal alpine/scratch `Dockerfile`.
- `docker-compose.yml` configures the backend alongside requisite services (e.g., PostgreSQL for the metadata DB).
- Useful commands:
  - `make docker-up` to launch all services.
  - `make docker-down` to clean up containers.

## Swagger Documentation

The project includes API documentation automatically generated via standard Go swag tooling.
The configuration and route comments output directly to the OpenAPI/Swagger format.
You can generate or update the documentation using:

```bash
make swagger
```

This compiles comments across the handlers and models into a detailed interactive web page for developer consumption.
