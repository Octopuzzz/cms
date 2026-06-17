# Backend Platform (Dynamic CMS + API Builder)

A production-grade, self-hosted, and Go-native Backend-as-a-Service (BaaS) platform. This platform allows developers to connect databases, dynamically create data models, automatically generate CRUD REST and optional GraphQL APIs, manage schema migrations, and perform backups—all while maintaining robust observability.

## 🏗️ System Architecture

The platform follows **Clean Architecture** and **Domain-Driven Design (DDD)**.

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

- **Presentation Layer (`internal/handlers/`, `internal/graphql/`)**: Exposes REST endpoints via Gin and GraphQL endpoints via `gqlgen`.
- **Application Layer (`internal/services/`)**: Contains the core business logic, including the CMS Control Plane, dynamic data fetching, schema migrations, and backup coordination.
- **Domain Layer (`internal/models/`)**: Defines the internal metadata schema (e.g., Services, Fields, Database Connections).
- **Infrastructure Layer (`internal/database/`, `internal/tracing/`, `pkg/logger/`)**: Manages external integrations, telemetry, and connection caching.

## 🗄️ Metadata Database Schema

The platform relies on an internal database (PostgreSQL by default) to manage state. Key tables include:

- `database_connections`: Stores configurations for user databases (PostgreSQL, MySQL, MongoDB, SQLite). Fields: `id`, `name`, `type`, `host`, `port`, `username`, `password`, `database_name`.
- `services`: Represents user-defined data models. Fields: `id`, `name`, `database_connection_id`, `db_table_name`, `created_at`, `updated_at`.
- `fields`: Schema definition for services. Fields: `id`, `service_id`, `name`, `type` (string, int, float, json, etc.), `nullable`, `unique`, `default_value`.
- `service_permissions`: Role-based access control rules mapping `roles` to `services`.
- `migrations`: Audit log for DDL operations, used for rollbacks.
- `backups`: Records of snapshots and schema dumps.
- `users` / `roles`: System administration and access control.

## 📂 Go Project Structure

The codebase is organized modularly:

```text
.
├── cmd/
│   └── server/              # Application entrypoint
├── internal/
│   ├── config/              # Configuration management
│   ├── database/            # Connection pooling & DB connectors
│   ├── graphql/             # GraphQL gateway (gqlgen)
│   ├── handlers/            # REST API controllers
│   ├── middleware/          # Auth, CORS, rate limiting, metrics
│   ├── models/              # Domain entities
│   ├── services/            # Core business logic (Service Builder, Migrations, etc.)
│   └── tracing/             # OpenTelemetry setup
├── pkg/
│   ├── logger/              # Zap structured logging
│   └── response/            # Standardized API responses
├── tests/
│   ├── integration/
│   ├── uat/
│   └── unit/
├── docker/                  # Dockerfiles and docker-compose configurations
└── docs/                    # Auto-generated Swagger documentation
```

## ⚙️ CRUD Engine & Query Engine Implementation

The **Dynamic CRUD Engine** auto-generates RESTful endpoints based on defined services:
- `POST /api/v1/data/{service}`
- `GET /api/v1/data/{service}`
- `GET /api/v1/data/{service}/{id}`
- `PUT /api/v1/data/{service}/{id}`
- `DELETE /api/v1/data/{service}/{id}`

The **Query Engine** maps URL parameters to advanced GORM queries, supporting:
- Filtering (`?email=test@example.com`)
- Sorting (`?sort=created_at:desc`)
- Pagination (`?page=1&limit=20`)
- Dynamic Joins based on configured relations (One-to-One, One-to-Many, Many-to-Many).

## 🔄 Schema Migration Engine

Handled by `internal/services/migration_service.go`, the migration engine supports safe schema mutations:
- Add, drop, rename, and change column types using GORM's `Migrator()`.
- Automatically logs migrations to the `migrations` metadata table.
- Supports rollback capabilities mapped to automated snapshots.

## 🌐 GraphQL Gateway (Optional)

Using `gqlgen`, the platform exposes a generic endpoint at `POST /api/v1/graphql`.
It dynamically parses metadata into GraphQL types, allowing complex nested queries and mutations over the generated data models.

## 💾 Backup System

Implemented in `internal/services/backup_service.go`, users can generate JSON snapshots of their service data:
- Backup API: `POST /api/v1/cms/backups/service/{service_id}`
- Restore API: `POST /api/v1/cms/restore/service/{service_id}`

## 👁️ Observability Integration

Production-ready telemetry is built-in:
- **Logging**: Structured, context-aware JSON logging via `go.uber.org/zap`.
- **Metrics**: HTTP latency, error rates, and resource usage exported via Prometheus (`/metrics` endpoint).
- **Tracing**: OpenTelemetry (OTel) traces exported to Jaeger. Spans cover database queries, external API calls, and HTTP requests.

## 🧪 Unit Tests

The system maintains an 80%+ coverage target across the Repository, Service, and Handler layers.

Run tests using the Makefile:
```bash
make test-unit
make test-integration
make test-uat
make test-coverage  # Full coverage with isolated database mocking
```

## 🐳 Docker Setup

A complete `docker-compose` environment is provided for easy local deployment, including the Go application, Postgres metadata DB, Redis cache, Prometheus, and Jaeger.

```bash
docker-compose -f docker/docker-compose.yml up -d
```

## 📖 Swagger Documentation

Interactive API documentation is automatically generated using `swaggo/swag`.
To re-generate docs:
```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
