# Dynamic CMS & Backend-as-a-Service Platform

A production-grade, self-hosted Backend Platform written in Go. This system dynamically creates data models, automatic CRUD REST APIs, optional GraphQL endpoints, and offers comprehensive features comparable to Hasura or Supabase.

## Features Highlight
- **Language**: Go (Latest stable version 1.24)
- **API Layers**: REST (Default via Gin) and GraphQL (Optional via gqlgen)
- **ORM**: GORM (Supporting SQLite, PostgreSQL, MySQL, SQLServer)
- **Core Capabilities**:
  - Dynamic Database Connectors
  - CMS Control Plane & Service Builder
  - Dynamic CRUD & Query Engine
  - Schema Migration Engine
  - Backup & Snapshot Engine
  - State-of-the-art Observability (Prometheus, OpenTelemetry Jaeger, Zap logging)
- **Deployment**: Fully containerized using Docker & Docker Compose.

---

## 1. System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles to separate concerns into modular components.

```text
       Client Applications
                │
   ┌────────────┴─────────────┐
   ▼                          ▼
REST API                 GraphQL API
   │                          │
   └────────────┬─────────────┘
                ▼
           API Gateway (Gin Router)
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
       Database Connectors (Connection Pool)
   ┌────────────┼─────────────┐
   ▼            ▼             ▼
PostgreSQL    MySQL       MongoDB / SQLite
```

### Core Layers:
- **Presentation Layer**: `internal/handlers/` mapping HTTP/GraphQL requests to business logic.
- **Application Layer**: `internal/services/` encapsulating all platform operational procedures (building services, executing CRUD, migrating databases).
- **Domain Layer**: `internal/models/` for strictly defining metadata data structures.
- **Infrastructure Layer**: Cross-cutting tools spanning tracing (`internal/tracing`), database adapters (`internal/database`), and logging.

---

## 2. Metadata Database Schema

The core metadata structure tracks external connections, defined microservices, fields, and access control. It defaults to storing data in an embedded SQLite database (`cms_backend.db`), but is fully portable to PostgreSQL.

### Core Tables

#### `database_connections`
Manages external DB definitions.
- `id`, `name`, `type`, `host`, `port`, `username`, `database_name`, connection pool settings (`max_open_conns`, `max_idle_conns`).

#### `services`
Represents user-created service boundaries (similar to tables or collections).
- `id`, `name`, `slug`, `db_table_name`, `database_connection_id`, `created_at`, `updated_at`.

#### `fields`
Defines columns and data types attached to a specific `service`.
- `id`, `service_id`, `name`, `type` (e.g., string, integer, boolean, relation), `is_nullable`, `is_unique`, `default_value`, `relation_config`.

#### `relations` (Abstracted)
Managed via the `RelationConfig` JSON struct inside the `fields` table. Support for one-to-one, one-to-many, and many-to-many associations. Join tables are dynamically built via the service builder.

#### Operational Tables
- `migrations`: Tracking executed schemas (`service_id`, `status`, `schema_snapshot`).
- `backups`: Storing data backups (`service_id`, `type`, `status`).
- `users` & `roles`: RBAC authentication tables.

---

## 3. Go Project Structure

The project follows idiomatic Go layouts:

```
├── cmd/
│   └── server/
│       └── main.go              # Entry point linking all components
├── internal/
│   ├── config/                  # Environment & App Config
│   ├── database/                # Connection manager & DB pooling
│   ├── graphql/                 # GraphQL gateway generator
│   ├── handlers/                # HTTP route controllers (Gin)
│   ├── middleware/              # Authentication, Metrics, Tracing interceptors
│   ├── models/                  # GORM entities & Domain schemas
│   ├── services/                # Business logic engines
│   └── tracing/                 # OpenTelemetry initializers
├── pkg/
│   ├── logger/                  # Global structured logging wrapper (Zap)
│   └── response/                # Standardized JSON response formatters
├── tests/
│   ├── integration/             # Component testing
│   ├── uat/                     # System tests
│   └── unit/                    # Fast isolated function tests
├── docker/                      # Supplementary configurations
├── docs/                        # Swagger API Documentation outputs
├── docker-compose.yml           # Local deployment file
└── Makefile                     # Developer tooling
```

---

## 4. Platform Engine Implementations

### Service Builder & Relation Engine
Located in `internal/services/service_service.go`. Dynamically builds DDL queries tailored to the target SQL dialect when users define schemas. Generates underlying physical tables based on the `db_table_name` field. Support for relations includes generating join queries via `DynamicDataService` logic.

### Dynamic CRUD & Query Engine
Located in `internal/services/dynamic_data_service.go`. Maps generic routes (`/api/v1/data/{slug}`) to safe runtime SQL compilation. Includes:
- Filtering and Pagination logic
- Safe execution avoiding SQL Injection (Strict `ORDER BY` whitelists checking fields against the database metadata structure)
- Implicit execution of `applyJoins` to enrich JSON outputs when requested via API parameters.

### Schema Migration Engine
Located in `internal/services/migration_service.go`. Provides an interface to add, drop, rename columns. It snapshots existing schemas to the `migrations` table before altering standard GORM `.Migrator()` definitions, ensuring rollbacks are possible.

### Backup System
Located in `internal/services/backup_service.go`. Exports dynamic data into JSON schemas, allowing restoration directly into rebuilt `services` from backup logs.

### Observability
Fully integrated.
- **Prometheus**: Endpoint at `/metrics` collects standard RED (Rate, Errors, Duration) statistics.
- **Zap Logging**: Output format managed through `pkg/logger`.
- **OpenTelemetry**: Tracing configurations wrap server requests routing traces dynamically to Jaeger instances.

### GraphQL Gateway
Built entirely around `gqlgen` inside `internal/graphql/gateway.go`. Dynamically processes abstract syntax trees representing GraphQL queries into metadata-aware CRUD requests behind the scenes. Exposes `POST /graphql`.

### API Documentation (Swagger)
Generated dynamically from Go doc comments via Swaggo to the `docs/` folder. Exposes a live Swagger UI dashboard at `/api/v1/swagger/index.html`.