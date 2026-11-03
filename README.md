# Dynamic CMS + API Builder Backend Platform

## 1. Platform Goal
A **Backend-as-a-Service platform** designed to allow developers to:
- Connect databases
- Create data models dynamically
- Generate CRUD APIs automatically
- Generate optional GraphQL APIs
- Manage schema migrations
- Monitor logs and performance
- Manage backups
- Scale services

This platform acts as a backend infrastructure generator.

## 2. Core Technology Requirements
- **Language**: Go
- **API Layer**: REST (default)
- **Optional API Layer**: GraphQL (using `gqlgen`)
- **Web Framework**: Gin
- **ORM**: GORM
- **Logging**: Zap structured logging
- **Metrics**: Prometheus
- **Tracing**: OpenTelemetry
- **Cache**: Redis
- **Database Migration**: Automatic diff via GORM `.Migrator()`
- **Containerization**: Docker
- **API Documentation**: Swagger / OpenAPI

## 3. High Level Platform Architecture

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

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## 4. Metadata Database Schema
The CMS stores platform metadata. Recommended database is PostgreSQL. Tables defined in `internal/models/models.go` include:

1. **`database_connections`**: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. **`services`**: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`). Example fields: `id`, `name`, `database_id`, `created_at`, `updated_at`.
3. **`fields`**: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. **`service_permissions`**: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. **`backups`**: Stores snapshot records or schema outputs.
7. **`users` and `roles`**: General authentication and authorization for the control plane.
8. **`audit_logs`**: Detailed logging of structural and data-level modifications.

## 5. Go Project Structure
Clean modular structure separating logic into domain-driven modules:

```text
.
├── cmd
│   └── server                # Main entry point for the application
├── docker                    # External configuration and metrics like Prometheus
├── internal
│   ├── config                # App settings and environment variables
│   ├── database              # Database connections, DSN logic, and pooling
│   ├── graphql               # GraphQL gateway schemas and execution engines
│   ├── handlers              # API Gateway (REST routes, Controllers)
│   ├── middleware            # Auth, Prometheus, OpenTelemetry Tracing
│   ├── models                # Data models and structures (Domain Layer)
│   ├── services              # Business logic (Application Layer, Core Engines)
│   └── tracing               # OpenTelemetry/Jaeger setup
├── pkg
│   ├── logger                # Zap structured logging
│   └── response              # Standardized API response utilities
└── tests
    ├── integration
    ├── uat
    └── unit                  # Unit tests isolated testing setup
```

## 6. CRUD Engine Implementation
When a service is created, the system automatically generates CRUD endpoints using the Dynamic CRUD Engine (`internal/services/dynamic_data_service.go`).

Operations:
- `POST /api/v1/data/{slug}`
- `GET /api/v1/data/{slug}/{id}`
- `PUT /api/v1/data/{slug}/{id}`
- `DELETE /api/v1/data/{slug}/{id}`
- `GET /api/v1/data/{slug}` (Lists all with filtering/sorting)

## 7. Schema Migration Engine
Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
- Supports automated synchronization and safety features like rolling back schemas via history tracking natively logged in the `migrations` table.

## 8. GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
- Generates optional generic schemas dynamically from defined data models using `gqlgen`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL queries, mutations, and relations.

## 9. Backup System
Managed by `internal/services/backup_service.go`.
- Generates schema and record snapshots. Implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports generic restore via payload reconstruction (JSON payloads).
- Example: `POST /api/v1/cms/backup/service/{service_id}`

## 10. Observability Integration
The system integrates state-of-the-art tools for full visibility:
- **Logging**: Zap logging (`pkg/logger/`) with structured output, tracking request logs, error logs, query logs.
- **Tracing**: OpenTelemetry (`internal/tracing/`) connected to Jaeger.
- **Metrics**: Prometheus metrics exported on `/metrics` (`internal/middleware/prometheus.go`) for tracking latencies and error rates.

## 11. Unit Tests
The system includes unit tests focusing on Repositories, Services, and API Handlers.
Minimum coverage standard: **80%**.

Command to run the full suite:
```bash
make test-coverage
# or
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## 12. Docker Setup
Fully containerized for easy deployment, caching, and observability tracking.
`docker-compose.yml` spins up:
- Core Platform
- PostgreSQL (Database)
- Redis (Cache)
- Jaeger (OpenTelemetry tracking)
- Prometheus (Metrics scraping)

To run the full stack:
```bash
docker-compose up -d --build
```

## 13. Swagger Documentation
API documentation is available through Swagger / OpenAPI. To generate or update the documentation, ensure you have the `swag` CLI installed, then run:

```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

*(Note: the generated `docs/` folder should not be tracked in version control, as per `.gitignore`).*
