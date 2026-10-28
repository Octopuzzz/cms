# Dynamic CMS & API Builder - Go Backend Platform

A production-grade Backend-as-a-Service (BaaS) platform built in Go. It enables dynamic creation of backend services, schema configuration, and automated REST (default) & GraphQL (optional) API generation. This platform serves as a backend infrastructure generator, similar to Hasura or Supabase, but is strictly self-hosted and Go-native.

## System Architecture

The platform follows **Clean Architecture** and **Domain-Driven Design (DDD)**.

### High-Level Components

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

- **Presentation Layer**: HTTP router (Gin) and API handlers acting as the gateway. Exposes both REST and GraphQL endpoints (via `gqlgen`).
- **Application Layer**: Contains core business logic and services (e.g., service definition, data manipulation).
- **Domain Layer**: Models core metadata entities (Services, Fields, Databases, Roles, etc.).
- **Infrastructure Layer**: Cross-cutting concerns such as logging, metrics, tracing, and DB connection pooling.

---

## Core Components

### 1. CMS Control Plane & Service Builder
The control plane handles external database connections, schema configurations, dynamic services, and their relational links.
- Defined in `/api/v1/cms/` routes and handled by `internal/services/service_service.go` & `internal/services/dbconn_service.go`.
- Services trigger automatic DB migrations to reflect data model state.

### 2. Metadata Database Schema
The metadata is internally maintained in the DB, mapping the configurations for user services.
**Tables include**:
- `database_connections` - Tracks DB connection strings and pooling config (e.g. `id, name, type, host, username`).
- `services` - Data model definitions mapping to specific databases. (e.g. `id, name, database_id, db_table_name, created_at, updated_at`).
- `fields` - Service attributes and types (string, integer, float, uuid, etc). Configurations include nullability and uniqueness.
- `service_permissions` - Fine-grained access control (Role-to-Service mappings).
- `migrations`, `backups`, `users`, `roles`, `audit_logs`.

### 3. Database Connection Manager
Allows runtime registration of external DBs (PostgreSQL, MySQL, MongoDB/SQLite). Configures connection pooling dynamically.

### 4. Relation Engine
Handles relations (One-to-One, One-to-Many, Many-to-One, Many-to-Many) between data models, executing JOIN rules for queries.

### 5. Dynamic CRUD Engine & Query Engine
Automatic REST endpoint generation mapping endpoints like `GET /api/v1/data/{slug}` to standard GORM operations via `DynamicDataService`.
- Supports Filtering (`?email=john@example.com`)
- Sorting (`?sort=created_at:desc`)
- Pagination (`?page=1&limit=20`)
- Dynamic joining through configured relations.

### 6. Schema Migration Engine
Tracks metadata and orchestrates physical schema migrations (Add, Drop, Rename Column) directly on target databases via `MigrationService`. Uses GORM's automigrate and raw SQL execution. Supports safety checks and rollbacks.

### 7. GraphQL Gateway (Optional)
Automates the exposure of service-defined data schemas into a GraphQL layer using `github.com/99designs/gqlgen/graphql` (accessible at `POST /api/v1/graphql`).

### 8. Backup Engine
Supports snapshotting schemas and table backups across different systems, generating structured snapshot outputs and allowing data restorations.

### 9. Observability Integration
Enterprise-grade observability metrics natively embedded:
- **Logging**: Zap structured logging with correlation/trace IDs via context (`pkg/logger/`).
- **Metrics**: Prometheus HTTP request interceptors for tracking latency, status codes, and errors (Exported at `/metrics`).
- **Tracing**: OpenTelemetry (OTel) + Jaeger implementations in `internal/tracing/` tracing the lifecycle of requests and SQL transactions.

### 10. Security Implementation
Includes rate-limiting, JWT-based Authentication, Row-Level & Field-Level Access Control verification, and automatic SQL Injection sanitization mechanisms.

---

## Go Project Structure

The project adopts a standard Go layout:
```text
.
├── cmd
│   └── server/               # Application entrypoint
├── docker/                   # Docker deployment scripts
├── internal/
│   ├── config/               # System and Environment Configurations
│   ├── database/             # Connection pooling and setup logic
│   ├── graphql/              # Dynamic GraphQL generation logic
│   ├── handlers/             # Gin HTTP REST logic
│   ├── middleware/           # Auth, Prom, Limiting, OTel middleware
│   ├── models/               # Domain Models (Services, Configs)
│   ├── services/             # Core engines: Builder, Query, Auth, CRUD, Migrations
│   └── tracing/              # OpenTelemetry logic integration
├── pkg/
│   ├── logger/               # Core logging library (Zap)
│   └── response/             # Standard API Response payload wrappers
├── tests/
│   ├── integration/          # System Integration verification
│   ├── uat/                  # User acceptance scenarios
│   └── unit/                 # Application isolated tests
├── ARCHITECTURE.md           # Internal architect source map
├── Makefile                  # Builder & Task commands
└── Dockerfile                # Production Multi-Stage Builder
```

---

## Unit Testing
We enforce an 80%+ test coverage threshold across the Repositories, Services, and Handlers. Tests utilize standard testing libraries mapped onto a high-performance in-memory SQLite DB.

**Run tests via Make:**
```bash
make test-unit
make test-integration
make test-coverage      # Calculates full project coverage
```
Or directly:
```bash
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

---

## Deployment & Setup

### Docker Setup
The project contains an optimized multi-stage Docker build process in `Dockerfile`, coupled with `docker-compose.yml` to spin up dependent resources (Database, Redis, Jaeger/Prometheus).

**Bootstrapping Environment:**
```bash
docker-compose up -d --build
```

### Swagger API Documentation
REST API endpoint specifications are generated via Swagger.

To update Swagger definitions:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

*(Note: The generated `docs/` folder should be gitignored)*
