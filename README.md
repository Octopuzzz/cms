# CMS Backend Platform

This repository contains a **production-grade Backend Platform (Dynamic CMS + API Builder)** written in Go. It operates similarly to Hasura or Supabase, but is fully self-hosted and Go-native. The platform enables users to dynamically create backend services, schemas, APIs (REST and optional GraphQL), manage schema migrations, monitor logs, and manage backups. It adheres strictly to Clean Architecture and Domain-Driven Design (DDD).

## System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer**: Gin HTTP Router and Handlers map requests to internal services, alongside optional GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic resides in services. These control data access, dynamic schemas, migrations, backups, and user logic.
- **Domain Layer**: The core entity models (`Service`, `Field`, `DatabaseConnection`, `User`, etc.).
- **Infrastructure Layer**: Cross-cutting tooling such as database connectors, observability metrics, tracing, and logging.

### High-Level Architecture Diagram
```
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

Platform metadata is stored in a relational database (defaulting to PostgreSQL, SQLite for local/testing). Key tables include:

- `database_connections`: Stores external database configurations (id, name, type, host, port, username, password, etc.).
- `services`: Represents user-created data models referencing a `database_connection_id` and the dynamic schema (`db_table_name`).
- `fields`: Defines attributes for each service (type: string, integer, float, uuid, JSON; nullable, unique, default_value, index).
- `relations`: (Logical model relationships) One-to-One, One-to-Many, Many-to-One, Many-to-Many.
- `service_permissions`: Links Roles to Services (e.g. CanCreate, CanRead, CanUpdate, CanDelete).
- `migrations`: Keeps track of DDL executions with metadata for history and rollbacks.
- `backups`: Stores backup snapshot metadata.
- `users` / `roles`: System authentication and RBAC for platform administrators.

## Go Project Structure

The project adopts a clean, modular structure:
```
cmd/
  └── server/             # Entrypoint for the CMS backend
internal/
  ├── config/             # Environment configuration
  ├── database/           # Database connections and pooling
  ├── graphql/            # GraphQL gateway generation
  ├── handlers/           # Gin HTTP REST handlers (Presentation)
  ├── middleware/         # Auth, Prometheus, Rate Limiter
  ├── models/             # Domain layer models
  ├── services/           # Application business logic (CRUD, Migrations, etc.)
  └── tracing/            # OpenTelemetry integration
pkg/
  ├── logger/             # Zap structured logging
  └── response/           # Standardized API responses
tests/
  ├── integration/        # Integration tests
  ├── uat/                # End-to-end / UAT tests
  └── unit/               # Isolated unit tests
docker/                   # Container configurations (Prometheus, etc.)
```

## CRUD Engine & Query Engine

The Dynamic CRUD Engine (in `internal/services/dynamic_data_service.go`) intercepts dynamic requests and translates them into live GORM database operations on the connected service schemas.

Endpoints generated:
- `POST /api/v1/data/{slug}` (Create)
- `GET /api/v1/data/{slug}/{id}` (Read)
- `PUT /api/v1/data/{slug}/{id}` (Update)
- `DELETE /api/v1/data/{slug}/{id}` (Delete)
- `GET /api/v1/data/{slug}` (List/Query)

Advanced Query Support includes filtering (`?email=john@example.com`), sorting (`?sort=created_at:desc`), pagination (`?page=1&limit=20`), and relation joins.

## Schema Migration Engine

The Schema Migration Engine handles safe schema changes via GORM’s Migrator. It tracks applied schema changes natively into the `migrations` metadata table. Operations supported include adding, dropping, renaming columns, and changing column types safely.

## GraphQL Gateway

A GraphQL API gateway (`internal/graphql/gateway.go`) is automatically supported alongside REST. Implemented via `gqlgen`, it allows queries and mutations conforming to the dynamically built service schemas.

## Backup System

The Backup Engine (`internal/services/backup_service.go`) implements mechanisms to perform table data backups and schema snapshots. It outputs JSON and structural snapshots mapping back to the external services for secure restoration and backup continuity.

## Observability Integration

Observability is a core pillar:
- **Logging**: Zap structured logging is used globally (`pkg/logger/`), supporting request and error logs.
- **Metrics**: Prometheus middleware (`internal/middleware/prometheus.go`) tracks latency, error rates, and status codes.
- **Tracing**: OpenTelemetry/Jaeger wraps queries and network requests in `internal/tracing/`.

## Unit Tests

Unit testing mandates a minimum 80% coverage focusing on Repositories, Services, and Handlers. In-memory SQLite is leveraged via `:memory:` for high-speed, isolated execution.
Run tests via:
```bash
make test-unit
make test-coverage
```

## Docker Setup

The platform is fully containerized. A robust `docker-compose.yml` spins up:
- The CMS Backend platform
- PostgreSQL metadata database
- Redis for query caching and rate limiting
- Jaeger for OpenTelemetry distributed tracing
- Prometheus for application metrics

Run the full stack:
```bash
docker-compose up -d
```

## API Documentation (Swagger)

Swagger (OpenAPI) documentation is auto-generated and served directly from the application. To generate the docs manually from the codebase:
```bash
make swagger
# or
~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
