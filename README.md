# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go, functioning as a fully self-hosted, Go-native alternative to platforms like Hasura or Supabase.

The platform enables developers to:
- Connect multiple external databases (PostgreSQL, MySQL, MongoDB).
- Dynamically create data models and schemas via a CMS Control Plane.
- Auto-generate complete CRUD REST APIs.
- Auto-generate optional GraphQL APIs.
- Execute safe schema migrations with rollback capabilities.
- Generate system and database backups.
- Utilize deep observability (Prometheus, OpenTelemetry, Zap).

## System Architecture Explanation

The system follows **Clean Architecture and Domain-Driven Design (DDD)**.

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

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`), routing REST requests to services. Includes GraphQL endpoints via `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Manages data access, schema generation, and service operations.
- **Domain Layer**: Core data models (`internal/models/`) defining entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting concerns such as database connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## Metadata Database Schema

Platform metadata is stored in a relational database (SQLite/PostgreSQL by default). The key tables are:

1. **`database_connections`**: Stores external DB configurations (type, host, port, credentials).
2. **`services`**: User-created data models. Links to a `database_connection_id` and maps to a dynamic `db_table_name`.
3. **`fields`**: Defines attributes for each service (type, unique, nullable, default).
4. **`service_permissions`**: Maps fine-grained Role-Based Access Control (RBAC) to Services.
5. **`relations`**: Defined inside `fields` via `RelationConfig` for 1:1, 1:N, N:1, N:M mappings.
6. **`migrations`**: History of schema changes (pending, applied, failed).
7. **`backups`**: Records of snapshots and schema dumps.
8. **`users` & `roles`**: Auth data for the platform.

## Go Project Structure

```text
.
├── cmd/
│   └── server/
│       └── main.go                 # Application entrypoint
├── internal/
│   ├── config/                     # Configuration and environment setup
│   ├── database/                   # Database connection manager & pooling
│   ├── graphql/                    # GraphQL Gateway and schema generator
│   ├── handlers/                   # HTTP / REST Controllers
│   ├── middleware/                 # Auth, Rate Limiter, Observability middleware
│   ├── models/                     # GORM domain entities & Metadata Schema
│   ├── services/                   # Core engines (CRUD, CMS, Migrations, etc.)
│   └── tracing/                    # OpenTelemetry configuration
├── pkg/
│   ├── logger/                     # Zap structured logger
│   └── response/                   # Standardized HTTP response helpers
├── tests/
│   ├── unit/                       # Unit tests
│   ├── integration/                # Integration tests
│   └── uat/                        # End-to-End User Acceptance Tests
├── docker/                         # Dockerfiles and Prometheus configs
├── docker-compose.yml              # Multi-container local environment
├── Makefile                        # Standard build & test tasks
└── go.mod                          # Dependencies (Go 1.24)
```

## CRUD Engine Implementation

The **Dynamic CRUD Engine** (`internal/services/dynamic_data_service.go`) intercepts generic REST calls (e.g., `/api/v1/data/{slug}`) and translates them into targeted GORM queries.

- Automatically maps the `{slug}` to its corresponding `Service` and metadata database connection.
- Auto-generates standard Create, Read, Update, Delete, and List endpoints.
- **Query Engine Features**:
  - Filtering: `?email=john@example.com`
  - Sorting: `?sort=created_at:desc`
  - Pagination: `?page=1&limit=20`
  - Relational Joins: Dynamically resolves Foreign Keys via `RelationConfig`.

## Schema Migration Engine

The **Schema Migration Engine** (`internal/services/migration_service.go`) ensures safe structural database updates:

- **Operations Supported**: Add column, Drop column, Rename column, Change type.
- **Safety**: Automatically takes snapshots of the model before running DDL queries, allowing you to rollback or audit changes.
- **Execution**: Changes execute directly against the target database via SQL `ALTER TABLE` commands.

## GraphQL Gateway

The **GraphQL Gateway** (`internal/graphql/gateway.go`) sits atop the REST-centric core to provide flexible query capabilities:

- Built using `github.com/99designs/gqlgen`.
- Can be dynamically mapped to existing services to auto-generate queries, mutations, and deep relational traversals based on the defined schemas.
- Exposed under `/api/v1/graphql`.

## Backup System

The **Backup Engine** (`internal/services/backup_service.go`) handles snapshotting system data:

- Provides snapshot-based data backups for dynamic service records.
- Serializes rows into JSON format for secure storage and restoration.
- Integrates with the platform's RBAC so only authorized users can trigger dumps or restorations.

## Observability Integration

High-performance observability is built into the infrastructure:

- **Logging**: Zap structured logging with Context correlation/trace IDs (`pkg/logger/`).
- **Metrics**: Prometheus middleware tracking latency, request rates, and response status codes, exposed on `/metrics` (`internal/middleware/prometheus.go`).
- **Tracing**: OpenTelemetry (OTEL) distributed tracing via Jaeger (`internal/tracing/tracing.go`).

## Unit Tests

The system maintains high test coverage. Tests are separated into Unit, Integration, and UAT.

- **Command**: `make test` or `make test-unit`
- **Coverage**: To ensure >= 80% coverage across all packages, use `make test-coverage`.
- Tests run against an in-memory SQLite database (`:memory:`) to guarantee speed and isolation.

## Docker Setup

The platform is fully containerized and orchestration is handled via `docker-compose.yml`.

- **Services**: The Go Backend, an external PostgreSQL database for metadata, and Prometheus for metrics.
- **Commands**:
  - `make docker-up` / `docker-compose up -d`
  - `make docker-build`

## Swagger Documentation

API documentation is handled via standard OpenAPI / Swagger annotations on handlers.

- **Command**: `make swagger`
- Generates `docs/` which is served at `/swagger/index.html`.
- Documents all core endpoints: CMS metadata control, authentication, backups, and migrations.
