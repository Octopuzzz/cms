# Dynamic CMS & API Builder

A production-grade, self-hosted Backend-as-a-Service (BaaS) platform written in Go. This platform allows developers to dynamically create backend services, schemas, and APIs, generating both REST and GraphQL endpoints dynamically.

It is modular, scalable, and cloud-ready, designed to replace generic backend development overhead with instantaneous API generation and robust control-plane management.

---

## System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles to maintain high cohesion and low coupling across the system components:

- **Presentation Layer**: Handled via the Gin HTTP Router (`internal/handlers/`), this layer acts as an API gateway, mapping incoming REST and GraphQL requests to the appropriate internal services.
- **Application Layer**: Contains business logic (`internal/services/`). The services control data access, compute dynamic schemas, and orchestrate the underlying engines (CRUD, Migration, Query).
- **Domain Layer**: The foundational data models (`internal/models/`). It defines the core metadata entities: `Service`, `Field`, `DatabaseConnection`, `User`, and `Role`.
- **Infrastructure Layer**: Cross-cutting concerns such as logging (`pkg/logger/`), observability & tracing (`internal/tracing/`), metrics (Prometheus), and connection caching.

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
PostgreSQL    MySQL       MongoDB / SQLite
```

---

## Metadata Database Schema

The CMS manages platform metadata securely in a centralized PostgreSQL (or SQLite/MySQL) instance. The models are defined in `internal/models/models.go` and include:

- **database_connections**: Stores configurations for external user databases. Tracks connection details (host, port, credentials, pool settings) to seamlessly map user tables.
- **services**: The abstraction for dynamically generated data models. Each service references a `database_connection_id` and tracks a specific `db_table_name`.
- **fields**: Defines individual attributes for each service. Maps basic types (string, integer, UUID, JSON, etc.) alongside configurable constraints like nullability, default values, and uniqueness.
- **service_permissions**: Manages fine-grained Role-Based Access Control (RBAC). Enables permission toggles (CanCreate, CanRead, CanUpdate, CanDelete) and advanced Row-Level and Field-Level access constraints.
- **migrations**: Audits executed Data Definition Language (DDL) schemas to provide rollback capabilities and a verifiable history of data-model changes.
- **backups**: Manages and records service snapshots, holding both structural schemas and dataset snapshots.
- **users** & **roles**: Provides standard control-plane authentication and role mapping.
- **audit_logs**: High-resolution logging of structural actions and data modifications across the platform.

---

## Go Project Structure

The project strictly follows a standard modular Go structure for enterprise projects:

```
.
├── cmd/
│   └── server/          # Entry point for the CMS application
├── docker/              # Deployment manifests and prometheus configurations
├── internal/            # Core private platform logic
│   ├── config/          # Environment configuration loading
│   ├── database/        # Database connection pool manager and adapters
│   ├── graphql/         # GraphQL gateway (gqlgen integration)
│   ├── handlers/        # Gin HTTP request handlers
│   ├── middleware/      # Rate-limiting, authentication, logging middlewares
│   ├── models/          # Core Domain models (Metadata Schema)
│   ├── services/        # Business Logic (Auth, CRUD, Service Building)
│   └── tracing/         # OpenTelemetry configuration
├── pkg/
│   ├── logger/          # Structured Zap logger wrapper
│   └── response/        # Gin API response standardization utilities
└── tests/
    ├── integration/     # Inter-service communication testing
    ├── uat/             # End-to-end user acceptance tests
    └── unit/            # Isolated business logic testing
```

---

## Core Engines & Implementations

### CRUD Engine Implementation
Located within `internal/services/dynamic_data_service.go`, the CRUD Engine dynamically translates API endpoints (`GET /api/v1/data/{slug}`, `POST /api/v1/data/{slug}`) to raw SQL and GORM database operations. It:
- Automates list, retrieval, creation, modification, and deletion.
- Employs advanced **Query Engine** capabilities like dynamic sorting (`?sort=created_at:desc`), filtering (`?email=test@example.com`), and limit/offset pagination.
- Resolves foreign-key relationships dynamically for complex nested joins across services.

### Schema Migration Engine
Located within `internal/services/migration_service.go`, the Schema Migration Engine ensures safe structural modifications:
- Applies automatic tracking of schema diffs (Adding, Dropping, or Renaming columns) relying on GORM's `.Migrator()`.
- Captures status directly within the `migrations` table.
- Orchestrates snapshot mechanisms prior to DDL modifications to ensure rollback integrity.

### GraphQL Gateway
Configured in `internal/graphql/gateway.go`, the platform allows optionally skipping REST.
- Integrated with `github.com/99designs/gqlgen`.
- Auto-generates resolvers based on the user-defined `services` configuration.
- Exposes queries and mutations at `POST /api/v1/graphql` to satisfy advanced relational querying dynamically.

### Backup System
Handled by `internal/services/backup_service.go`, ensuring complete data durability.
- Enables snapshotting of specific services.
- Backs up schema structures and full underlying row-level table data as JSON snapshots.
- Readily supports data restoration endpoints.

---

## Observability Integration

Observability is a core pillar of the platform, enabling robust monitoring of APIs and background services:
- **Logging**: Implemented with **Zap** (`pkg/logger/`), providing ultra-fast structured JSON logging, mapping Trace IDs across requests.
- **Metrics**: Endpoints are automatically monitored using **Prometheus** (`internal/middleware/prometheus.go`). Exported at `/metrics`, tracking total requests, error rates, and request latencies.
- **Tracing**: Deep internal tracing is achieved using **OpenTelemetry** (`internal/tracing/`), wrapping both network requests and SQL database transactions for holistic bottleneck discovery.

---

## Testing & Quality Assurance

A strong emphasis on testing is required for all components. The testing suite focuses on maintaining at least 80% coverage across business critical units (Repositories, Services, and API Handlers).

- **Unit Tests**: Ensure atomic logic is correct (`tests/unit/`). Uses in-memory SQLite (`:memory:`) via `github.com/glebarez/sqlite` to validate dynamic operations locally without external DB constraints.
- **Integration Tests**: Validates interactions across modules.
- **UAT**: Emulates actual super-admin actions mapping APIs directly.

**Executing Tests:**
```bash
make test-coverage
```
*Note: Due to Go 1.24 local toolchain specifics, dependency resolution may take time without a locally cached environment.*

---

## Docker Setup

The system provides out-of-the-box infrastructure deployment using Docker:
- **Dockerfile**: Compiles the Go application minimizing the attack surface (using distroless-like multi-stage builds).
- **docker-compose.yml**: Orchestrates the local platform. It stands up:
  - `cms-backend`: The compiled Go binary platform.
  - `postgres`: The primary Metadata Database.
  - `redis`: High-speed caching server for API requests and session storage.
  - `prometheus`: Gathers application metrics natively.

To bring up the stack:
```bash
make docker-up
```

---

## Swagger API Documentation

The platform fully documents the internal control plane endpoints via **OpenAPI / Swagger**.

Documentation generation relies on standard `swag` annotations over the Gin handlers. To regenerate documentation after modifications:
```bash
make swagger
```
The output is written directly to the ignored `/docs` folder and is served automatically by the API gateway on application launch for interactive developer exploration.
