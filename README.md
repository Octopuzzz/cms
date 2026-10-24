# Go Dynamic CMS + API Builder

A self-hosted, Go-native Backend-as-a-Service (BaaS) platform that dynamically generates REST and GraphQL APIs. Similar to Hasura or Supabase, it allows developers to connect databases, dynamically model data schemas, and instantly generate endpoints with robust observability, migrations, and backups.

## System Architecture

The system is built on **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

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

- **Presentation Layer**: Gin HTTP Router mapping REST routes and a GraphQL gateway via `gqlgen`.
- **Application Layer**: Business logic (Services) that orchestrates dynamic data modeling, CRUD, migrations, and schema parsing.
- **Domain Layer**: Core metadata models representing the internal state (Services, Fields, Relations, Connections).
- **Infrastructure Layer**: Cross-cutting observability (Prometheus, OpenTelemetry, Zap), database connection pools, and caching mechanisms.

## Metadata Database Schema

The core metadata drives the entire platform. Standard metadata tables (powered by PostgreSQL/SQLite) include:

- `database_connections`: Stores configurations (host, port, credentials) for connected external databases with connection pooling.
- `services`: Represents user-created data models/entities. Maps to specific tables on connected databases.
- `fields`: Defines attributes for services (e.g., string, integer, float, uuid, JSON), nullability, defaults, and uniqueness.
- `service_permissions`: Connects RBAC roles to services for fine-grained permissions (CanCreate, CanRead, row-level filters).
- `migrations`: Tracks dynamic schema modifications and allows rollbacks.
- `backups`: Records backup snapshots and SQL dumps.
- `users` / `roles`: System administration authentication and authorization.
- `audit_logs`: Detailed tracking of structural or data changes.

## Go Project Structure

A highly modular standard Go layout:

```text
.
├── cmd/
│   └── server/          # Entry point (main.go)
├── internal/
│   ├── api/             # API routes
│   ├── config/          # Configuration loader
│   ├── database/        # Connection manager and pool handler
│   ├── graphql/         # GraphQL gateway and resolvers (gqlgen)
│   ├── handlers/        # HTTP handlers (Presentation)
│   ├── middleware/      # Auth, Logging, Metrics, Tracing, Rate Limiter
│   ├── models/          # Domain layer schemas (Metadata tables)
│   ├── services/        # Application layer (CRUD, Migration, Service Builder)
│   └── tracing/         # OpenTelemetry setup
├── pkg/
│   ├── logger/          # Zap structured logging
│   └── response/        # Standardized HTTP responses
├── tests/
│   ├── integration/
│   ├── uat/
│   └── unit/
├── docker/              # Container deployment artifacts
├── Makefile             # Standard task runner
└── go.mod
```

## Core Components Implementation

### CRUD & Query Engine
When a service is created, standard REST operations are dynamically routed through the `DynamicDataService`.
- **Endpoints**: `POST /api/v1/data/{slug}`, `GET /api/v1/data/{slug}/{id}`, `PUT`, `DELETE`.
- **Query Engine**: Built-in URL filtering (`?email=test@example.com`), sorting (`?sort=created_at:desc`), pagination (`?page=1&limit=20`), and join relations (`?join=role`).
- Uses GORM to dynamically map request payloads to actual database tables in real-time, respecting row-level security and field permissions.

### Schema Migration Engine
Managed by the `MigrationService`.
- Supports safe schema mutations (Add Column, Drop Column, Rename, Change Type).
- Integrates with GORM AutoMigrate and custom Alter Table queries.
- Records migration history into the `migrations` table for safety.

### GraphQL Gateway
Automatically provisions schemas from `services` definitions.
- Generates GraphQL types and resolvers natively via `gqlgen`.
- Users can query relations and nest objects exactly like the REST Query Engine via `POST /api/v1/graphql`.

### Backup System
Handled by the `BackupService`.
- Backs up entire service datasets into JSON snapshots or native SQL dumps.
- Stores historical artifacts and allows restoration points for critical models.

### Observability Integration
Fully instrumented for modern cloud-native environments:
- **Logging**: High-performance structured logging via `Zap`, injecting request/trace context automatically.
- **Metrics**: Standard Prometheus instrumentation on all routes (`/metrics`), capturing request rates, errors, and latencies.
- **Tracing**: Distributed tracing via OpenTelemetry exported to Jaeger, mapping DB queries and network latency across boundaries.

## Unit Tests

The platform adheres to strict testing standards targeting a minimum 80% coverage on core layers (Repository, Service, API).
Tests run using isolated in-memory SQLite instances for rapid, reliable feedback.

Run tests:
```bash
make test-unit
# or coverage
make test-coverage
```

## Docker Setup

Ready for containerized deployment. Includes full compose setups mapping out backend, databases, redis, prometheus, and jaeger.

Run locally:
```bash
docker-compose up -d
```

## Swagger Documentation

API schema documentation generated natively using `swag`.

Generate docs:
```bash
make swag-init
# or
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

The Swagger UI will be served at `http://localhost:8080/swagger/index.html` (depending on configuration).
