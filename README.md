# Dynamic CMS + API Builder Platform

## Overview

A production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. This system acts as a fully self-hosted, Go-native Backend-as-a-Service (BaaS) similar to platforms like Hasura or Supabase.

It allows developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

The platform is modular, scalable, and cloud-ready, adhering to **Clean Architecture** and **Domain-Driven Design (DDD)** principles.

## System Architecture Explanation

The system is structured into four main layers based on Clean Architecture:

1. **Presentation Layer (`internal/handlers/`, `internal/graphql/`)**: The Gin HTTP Router and Handlers. It acts as the API gateway mapping external REST and GraphQL requests to internal services.
2. **Application Layer (`internal/services/`)**: Business logic layer. Services control data access, generate dynamic schemas, validate configurations, and perform all operations requested by handlers.
3. **Domain Layer (`internal/models/`)**: The data models layer defining core platform metadata entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`, and `Permission`.
4. **Infrastructure Layer (`internal/database/`, `pkg/logger/`, `internal/tracing/`)**: Cross-cutting infrastructure tools such as connection pooling, OpenTelemetry tracing, Prometheus metrics, Zap logging, and caching.

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

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**, which can be backed by PostgreSQL or SQLite. Key models include:

- **`database_connections`**: Stores external database configurations (id, name, type, host, port, credentials) and connection pooling settings.
- **`services`**: Represents dynamically created user data models. Tracks dynamic schemas (`db_table_name`) and foreign keys to `database_connections`.
- **`fields`**: Defines attributes for each service (type: string, integer, uuid, JSON, boolean). Configures constraints like uniqueness, nullability, and default values.
- **`roles` & `users`**: Manages RBAC authentication and authorization.
- **`service_permissions`**: Connects roles to services, providing fine-grained access checks (e.g., CanCreate, CanRead). Contains row-level and field-level capabilities.
- **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- **`backups`**: Stores snapshot records or schema outputs.

## Go Project Structure

The project uses a clean modular structure to ensure maintainability:

```text
.
├── cmd/
│   └── server/             # Application entrypoint
├── docker/                 # Dockerfile, compose, and Prometheus configurations
├── internal/
│   ├── config/             # Configuration loading (viper, .env)
│   ├── database/           # DB Connection Manager, pooling logic
│   ├── graphql/            # GraphQL Gateway and schema generation
│   ├── handlers/           # REST API presentation layer
│   ├── middleware/         # Auth, Prometheus, Tracing, Rate Limiter
│   ├── models/             # Domain layer models (metadata DB)
│   ├── services/           # Application layer logic
│   └── tracing/            # OpenTelemetry setup
├── pkg/
│   ├── logger/             # Zap structured logging
│   └── response/           # Gin JSON response helpers
├── tests/                  # Unit, Integration, and UAT tests
│   ├── unit/
│   ├── integration/
│   └── uat/
├── Makefile                # Standardized build, test, run tasks
└── ARCHITECTURE.md         # Architecture documentation
```

## Core Components Implementation

### CMS Control Plane & Service Builder
Provides management interfaces (`/api/v1/cms/`) to control the system. Responsibilities include defining database connections, configuring services, managing schema relations, and access control. It validates foreign database connectivity and triggers automatic migrations when service definitions are updated.

### CRUD Engine & Query Engine
When a service is created, the system auto-generates RESTful CRUD endpoints. Mapped dynamically by `DynamicDataService`, queries support advanced operations:
- **Filtering**: `GET /api/v1/data/{slug}?email=john@example.com`
- **Sorting**: `GET /api/v1/data/{slug}?sort=created_at:desc`
- **Pagination**: `GET /api/v1/data/{slug}?page=1&limit=20`
- **Joins**: Automated foreign key mappings and related field projections based on service relations.

### Schema Migration Engine
Tracks and safely applies schema changes to remote databases. It supports adding, dropping, and renaming columns safely by managing operations through GORM's `Migrator()`. Migration status is natively logged and tracks execution times and success flags.

### GraphQL Gateway
Automatically maps defined services into GraphQL schemas. Using `gqlgen`, it generates generic types on the fly to support dynamic Queries and Mutations. It is exposed at `/api/v1/graphql` and resolves fields directly using the Query Engine.

### Backup System
Implements a snapshot engine for backups. It can generate data snapshots of specific service tables, storing row counts and encoded table data. Restore mechanisms map encoded JSON payloads back to relational databases via the dynamic engine.

### Observability Integration
Extensive observability is built into the system core:
- **Logging**: Configured via Zap (`pkg/logger/`). Request logs, error logs, and SQL traces contain injected correlation IDs.
- **Metrics**: Standard HTTP metrics (latency, status codes, rates) are captured using Prometheus via `internal/middleware/prometheus.go` and exported at `/metrics`.
- **Tracing**: Fully instrumented using OpenTelemetry (`internal/tracing/`), wrapping both network boundaries and database drivers.

## Security & Performance
- **Security**: JWT-based authentication for user management, role-based access control (RBAC), automatic row-level filters, SQL injection protection through parameterized execution, and rate limiting middleware.
- **Performance**: Connection pooling managed per-database in `internal/database/connection_manager.go`. Query processing maps generic types minimizing serialization overhead.

## Unit Tests
Tests are grouped into unit, integration, and uat directories under `/tests`.
They utilize in-memory SQLite (`:memory:`) to decouple dependencies.
Minimum target coverage is 80%.

Commands:
```bash
# Run all tests
make test

# Generate coverage
make test-coverage
```

## Docker Setup
The platform is container-ready. Use `docker-compose` to spin up the application along with support systems (Prometheus, Postgres, etc).

Commands:
```bash
make docker-up
make docker-down
```

## Swagger Documentation
API definitions are standardized using Swagger/OpenAPI. Docs are accessible when running the API.

To re-generate them:
```bash
make swagger
```
