# CMS Backend Platform

This is a production-grade, self-hosted, Go-native Backend-as-a-Service (BaaS) platform (Dynamic CMS + API Builder). It works similarly to platforms like Hasura or Supabase. The platform allows developers to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints. It acts as a modular, scalable, and cloud-ready backend infrastructure generator.

## Core Features
* Connect external databases (PostgreSQL, MySQL, MongoDB, SQLite).
* Create data models (services) dynamically.
* Generate CRUD APIs (REST) automatically.
* Generate optional GraphQL APIs (via `gqlgen`).
* Manage schema migrations.
* Manage relation mappings and constraints.
* Monitor logs and performance (Prometheus, OpenTelemetry).
* Backup and restore services.

## Core Technology Stack
* **Language:** Go 1.24+
* **API Layer:** REST (Gin), GraphQL (`gqlgen`)
* **ORM:** GORM
* **Logging:** Zap
* **Metrics:** Prometheus
* **Tracing:** OpenTelemetry / Jaeger
* **Cache:** Redis (supported concept)
* **API Docs:** Swagger / OpenAPI
* **Containerization:** Docker

## System Architecture

The project adheres to **Clean Architecture** and **Domain-Driven Design (DDD)**.

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

- **Presentation Layer (`internal/handlers`)**: Maps HTTP requests (via Gin) to internal services, handles validation and request formatting. Exposes the optional GraphQL endpoints.
- **Application Layer (`internal/services`)**: Business logic. Generates dynamic schemas, controls data access, and executes tasks required by handlers.
- **Domain Layer (`internal/models`)**: Data entities, like `Service`, `Field`, `DatabaseConnection`, `Role`, `User`.
- **Infrastructure Layer**: Cross-cutting concerns including tracing (`internal/tracing`), metrics (Prometheus), logging (`pkg/logger`).

## Project Structure
```text
.
├── cmd
│   └── server/          # Main application entry point
├── docker/              # Docker configuration files
├── internal/
│   ├── config/          # Application configuration mapping
│   ├── database/        # Connection caching and multi-database management
│   ├── graphql/         # gqlgen gateway implementation
│   ├── handlers/        # Gin HTTP route handlers
│   ├── middleware/      # Auth, Promtheus, Security Headers, Rate Limiting
│   ├── models/          # GORM Metadata Models
│   ├── services/        # Core business logic and dynamic engines
│   └── tracing/         # OpenTelemetry / Jaeger initialization
├── pkg/
│   ├── logger/          # Zap structured logger wrapper
│   └── response/        # Standard JSON HTTP response helpers
└── tests/               # Unit, Integration, and UAT test suites
```

## Metadata Database Schema

Platform metadata is stored securely (defaults to PostgreSQL or SQLite). Essential tables (`internal/models/models.go`) include:

1. `database_connections`: Stores DB configs (type, host, credentials, pooling).
2. `services`: The user-created data models referencing `database_connections`.
3. `fields`: The attributes/columns for a `service` (string, int, uuid, etc) along with their validations and constraints.
4. `service_permissions`: Role-based fine-grained access logic.
5. `migrations`: Tracks dynamic DDL executions.
6. `backups`: Keeps backups of table rows or snapshots.
7. `users` and `roles`: Control plane authentication.
8. `audit_logs`: Detailed metadata change logging.

## Core Engines Implementation

### CRUD & Query Engine
Managed by `internal/services/dynamic_data_service.go`. Maps endpoints like `GET /api/v1/data/{slug}` to GORM commands. It supports automatic filtering (e.g. `?email=test`), sorting, and nested joins based on predefined metadata relations.

### Schema Migration Engine
Managed by `internal/services/migration_service.go`. Tracks schema diffs (add/drop/rename columns) automatically based on changes made via the `Service Builder`. Includes safety rollbacks via snapshot retention.

### GraphQL Gateway
Managed by `internal/graphql/gateway.go`. Automatically generates generic types and resolvers exposing models defined in the platform. Mounts on `POST /api/v1/graphql` and provides a playground at `/api/v1/graphql/playground`.

### Backup System
Managed by `internal/services/backup_service.go`. Creates data snapshots based on the dynamic tables. Can restore records via JSON payloads.

### Observability Integration
* **Logging**: `pkg/logger/` provides Zap structured logs globally.
* **Tracing**: `internal/tracing/` wraps SQL statements and network requests using Jaeger and OpenTelemetry.
* **Metrics**: `internal/middleware/prometheus.go` records latencies and error rates, exposed at `/metrics`.

## Unit Testing
Unit tests are extensive across handlers, services, and repositories, utilizing an in-memory SQLite database setup. Target coverage is >80%.
```bash
make test-coverage
# Or to run standard unit tests:
make test-unit
```

## Docker Setup
The project contains comprehensive Docker configuration files inside the `docker/` folder and `docker-compose.yml` to effortlessly spin up the application along with Redis, Postgres, and Jaeger.
```bash
make docker-up
```

## Swagger Documentation
API specs are auto-generated via Swaggo.
```bash
make swagger
```
The Swagger UI is accessible via `GET /api/v1/swagger/index.html`.
