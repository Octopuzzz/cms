# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend-as-a-Service (BaaS) platform written in Go, acting as a backend infrastructure generator. It functions similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native. The platform allows users to connect databases, define data models dynamically, automatically generate CRUD APIs and GraphQL endpoints, and monitor logs and performance.

## 1. System Architecture

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles.

### High-Level Platform Architecture

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

### Core Layers
- **Presentation Layer**: The Gin HTTP Router and Handlers (e.g., `/api/v1/cms/`, GraphQL over `gqlgen`).
- **Application Layer**: Business logic via modular services like `DynamicDataService`, `ServiceService`, `MigrationService`, and `BackupService`.
- **Domain Layer**: The internal models and core entity structures.
- **Infrastructure Layer**: Cross-cutting tools (Redis cache, Telemetry, Prometheus metrics, Zap logging, DB connections).

## 2. Metadata Database Schema

The CMS stores platform metadata in its own internal database (e.g., PostgreSQL or SQLite).

Key tables and schemas include:
- `database_connections`: Stores configuration for external DBs (id, name, type, host, port, username, password, etc.).
- `services`: Represents user-created data models, linking to a database connection. Fields include `id`, `name`, `database_id`, `created_at`, `updated_at`, and table mappings.
- `fields`: Defines attributes for services (type, nullable, unique, default_value, indexing).
- `relations`: Defines associations like One-to-One, One-to-Many, Many-to-One, Many-to-Many.
- `migrations`: Tracks DDL executions for safe schema changes and rollbacks.
- `backups`: Stores backup records, paths, or JSON snapshots.

## 3. Go Project Structure

The project uses a clean modular layout:

```text
├── cmd
│   └── server
│       └── main.go           # Application entrypoint
├── docker
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── prometheus.yml
├── internal
│   ├── config              # Environment and configuration variables
│   ├── database            # Database connection managers
│   ├── graphql             # GraphQL Gateway logic (gqlgen)
│   ├── handlers            # Presentation Layer (Gin HTTP handlers)
│   ├── middleware          # Auth, Rate Limiter, Prometheus
│   ├── models              # Domain Layer (Database schemas)
│   ├── services            # Application Layer (Core Engines)
│   └── tracing             # OpenTelemetry integration
├── pkg
│   ├── logger              # Zap logging
│   └── response            # Standardized API response helpers
└── tests
    ├── integration         # Integration tests
    ├── uat                 # End-to-end / UAT tests
    └── unit                # Unit tests
```

## 4. CRUD Engine Implementation

The Dynamic CRUD Engine (`internal/services/dynamic_data_service.go`) maps API endpoints dynamically to database operations:
- `POST /api/v1/data/{slug}`: Create new records.
- `GET /api/v1/data/{slug}/{id}`: Read a single record.
- `PUT /api/v1/data/{slug}/{id}`: Update an existing record.
- `DELETE /api/v1/data/{slug}/{id}`: Delete a record.
- `GET /api/v1/data/{slug}`: List records.

The engine also serves as the **Query Engine**, handling filters (`?email=user@test.com`), sorting (`?sort=created_at:desc`), pagination (`?page=1&limit=20`), and table joins based on defined relations.

## 5. Schema Migration Engine

The Schema Migration Engine (`internal/services/migration_service.go`) enables safe schema changes when users create or modify services via the Service Builder.
- Evaluates differences between the defined schema and actual database tables.
- Applies GORM `.Migrator()` operations (e.g., adding/dropping/renaming columns, changing types).
- Automatically records the migration state in the metadata database to facilitate history tracking and potential rollbacks.

## 6. GraphQL Gateway (Optional)

The GraphQL Gateway (`internal/graphql/gateway.go`) runs alongside the REST API, utilizing `gqlgen` to dynamically resolve GraphQL queries and mutations against user-defined services.
- Translates dynamic models into an executable GraphQL schema.
- Handles automated queries, mutations, and deep relational fetching.

## 7. Backup System

The Backup Engine (`internal/services/backup_service.go`) manages table/schema backups and restores.
- Backs up dynamic service tables by generating JSON snapshots or database-specific dumps.
- API endpoints map directly to actions (e.g., `POST /api/v1/cms/backup/service/{service_id}`).
- Capable of full or partial restores via payload decoding.

## 8. Observability Integration

The platform provides out-of-the-box, comprehensive observability tools (`internal/middleware/`, `internal/tracing/`):
- **Logging**: Zap structured logging ensures all requests, queries, and errors are cleanly logged, equipped with tracing correlation IDs.
- **Metrics**: Prometheus middleware records request latencies, slow queries, and HTTP error rates (`/metrics` endpoint).
- **Tracing**: OpenTelemetry (integrated with Jaeger) wraps critical HTTP handlers and database commands, enabling end-to-end visualization of execution paths.

## 9. Unit Tests

The system maintains comprehensive unit, integration, and UAT tests across the Repository, Service, and API Handler layers.
- **Run all tests**: `make test` or `go test -v -race ./tests/...`
- **Run with coverage**: `make test-coverage`
- Tests run against an in-memory SQLite database environment to ensure isolation, high execution speed, and deterministic outcomes. Minimum coverage is targeted at 80%.

## 10. Docker Setup

The platform is containerized for seamless deployment and scaling.
- Provides a multi-stage `Dockerfile` to create a lean production image.
- Uses `docker-compose.yml` to orchestrate the Go application, external PostgreSQL/MySQL/MongoDB databases, Redis (for caching), Prometheus, and Jaeger.
- **Run the stack**: `make docker-up`

## 11. Swagger Documentation

Interactive Swagger/OpenAPI documentation is fully supported and auto-generated based on the internal API handler schemas.
- Uses `swaggo/swag` mapped to the Gin router.
- **Generate Docs**: `make swagger` (requires the `swag` CLI).
- Exposes API specifications for immediate client consumption.
