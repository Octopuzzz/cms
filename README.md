# Dynamic CMS & API Builder Platform

## Overview
This platform is a production-grade Backend-as-a-Service (BaaS) built dynamically in Go. It operates similarly to platforms like Hasura or Supabase but is fully self-hosted and natively implements dynamic schema creation and API endpoints in Go.

## System Architecture

The project adheres to Clean Architecture and Domain-Driven Design (DDD):
- **Presentation Layer**: Handled by the Gin HTTP Router, acting as the API gateway. This layer maps REST and GraphQL requests to internal services.
- **Application Layer**: Contains business logic (`internal/services/`). These services govern data access, auto-generate CRUD REST endpoints (and optional GraphQL endpoints), schema migrations, and more.
- **Domain Layer**: Houses data models (`internal/models/`), such as `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Incorporates essential tools like database connection caching, OpenTelemetry tracing (`internal/tracing/`), Prometheus metrics (`internal/middleware/`), and Zap structured logging (`pkg/logger/`).

## Metadata Database Schema

The core platform configuration relies on a Metadata Database (typically PostgreSQL, but SQLite is supported for testing). Important tables include:
- `database_connections`: External DB configurations.
- `services`: Dynamic user-created data models.
- `fields`: Attributes for each service.
- `service_permissions`: Role-Based Access Control logic linking `Role` to `Service`.
- `migrations`: Tracking of schema diffs (Add, Drop, Rename Column) for applying and rolling back DDL executions.
- `backups`: Records of data snapshots.
- `users` / `roles`: Handles core platform authentication/authorization.

## Go Project Structure

The project has a clear and modular layout:
```
.
├── cmd
│   └── server                # Entrypoint for the application
├── docker                    # Contains Docker and monitoring configurations
├── internal                  # Core platform implementation
│   ├── config                # Environment variables and configuration logic
│   ├── database              # Connection manager for self and external databases
│   ├── graphql               # Auto-generated GraphQL Gateway logic (`gqlgen`)
│   ├── handlers              # Gin HTTP handlers for APIs (CRUD, Auth, CMS)
│   ├── middleware            # Request interceptors (Auth, Prometheus, Rate Limiter)
│   ├── models                # GORM models (Base models and metadata schemas)
│   ├── services              # Business logic (CRUD Engine, Migrations, CMS Builder)
│   └── tracing               # OpenTelemetry setup
├── pkg
│   ├── logger                # Zap structured logging
│   └── response              # Standardized API responses
├── tests                     # Automated tests (Unit, Integration, UAT)
└── Makefile                  # Development aliases and build commands
```

## Component Implementations

### CRUD Engine & Query Engine
Located primarily in `internal/services/dynamic_data_service.go`, this component dynamically parses HTTP endpoints and translates them into GORM database operations on the fly. It natively supports advanced querying like filtering (`?email=test`), sorting (`?sort=created_at:desc`), pagination (`?page=1&limit=20`), and dynamically joined relations based on mapping metadata.

### Schema Migration Engine
The `internal/services/migration_service.go` performs safe schema changes (Add, Drop, Rename Column, Change type). It natively implements GORM's `Migrator()` to directly apply data definition language commands securely on the connected database. It also integrates auto-backups before migrations and handles structured rollback logging.

### GraphQL Gateway
Built using `99designs/gqlgen`, the gateway located at `internal/graphql/` dynamically generates GraphQL schemas based on created services. The gateway leverages `POST /api/v1/graphql` to satisfy queries and mutations for optional interaction with services.

### Backup Engine
The `internal/services/backup_service.go` allows generating records of data snapshots, supporting structured rollback capabilities or exports mapping dynamic table contents to snapshots natively.

### Observability Integration
Extensive telemetry is woven natively into the application:
- **Logging**: Zap structured logging is injected globally (`pkg/logger/`), incorporating trace and request IDs directly within Context.
- **Tracing**: OpenTelemetry (integrated via Jaeger) handles deep tracing of HTTP requests and SQL commands (`internal/tracing/`).
- **Metrics**: Prometheus middleware tracks metrics like request latency, error rate, and query execution mapping standard HTTP codes automatically to `/metrics`.

### Unit Tests
The project prioritizes stability with automated tests targeting > 80% coverage on the Repository, Service, and API Handler layers.
Run tests using:
```bash
make test-coverage
# or manually
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

### Docker Setup
The platform is fully containerized. A `docker-compose.yml` file is provided, spinning up:
- The backend application (`Dockerfile`)
- Core databases (PostgreSQL/MySQL/Redis)
- Observability dashboards (Prometheus, Jaeger)

### Swagger Documentation
Comprehensive API documentation (Swagger / OpenAPI) is generated automatically. To update or create the swagger docs:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```