# Go Backend Platform (Dynamic CMS + API Builder)

A production-grade Backend Platform (Backend-as-a-Service) written in Go, acting as a backend infrastructure generator. Similar to Supabase or Hasura, this system allows users to dynamically create backend services, schemas, and auto-generated APIs (REST and optional GraphQL) completely from scratch, while managing database connections, migrations, observability, and backups.

## High-Level Platform Architecture

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

## System Architecture Explanation

This platform is structured using Clean Architecture and Domain-Driven Design (DDD):
- **Presentation Layer (`internal/handlers/`, `internal/graphql/`)**: Handles REST requests using Gin and GraphQL requests via `gqlgen`. Act as the API Gateway.
- **Application Layer (`internal/services/`)**: Orchestrates business logic across multiple microservice components including Service Builder, Dynamic CRUD Engine, Backup, and Migration engines.
- **Domain Layer (`internal/models/`)**: Defines core entity schema logic mapped to database records (`services`, `fields`, `database_connections`, `migrations`, `backups`, etc.).
- **Infrastructure Layer (`internal/database/`, `pkg/logger/`, `internal/tracing/`)**: Cross-cutting tools ensuring observability and stable telemetry (Zap logging, OpenTelemetry, Prometheus metrics), plus caching (Redis) and message queue components.

## Metadata Database Schema

The CMS manages platform configuration in a PostgreSQL metadata database mapping the following entities:

1. `database_connections`: External database registrations (PostgreSQL, MySQL, MongoDB). Fields: `id`, `name`, `type`, `host`, `port`, `username`, `password`, `database`, etc.
2. `services`: Data models linked to an external database. Fields: `id`, `name`, `database_connection_id`, `created_at`, `updated_at`.
3. `fields`: Dynamic columns belonging to a service. Support types: `string`, `integer`, `float`, `boolean`, `uuid`, `json`, `array`, `timestamp`. Defines `nullable`, `unique`, `default_value`, and `index`.
4. `service_permissions`: Connects `Role` to `Service` for Role-Based Access Control (RBAC).
5. `migrations`: Tracks schema migration history.
6. `backups`: Logs system snapshots.
7. `users` & `roles`: System authentication.

## Core Components Implementation

### CMS Control Plane & Service Builder
Users define configurations via `/api/v1/cms/`. Handled by the **Service Builder** (`internal/services/service_service.go`), allowing the registration of dynamic databases, services, schemas, fields, and constraints. When a service model is created, relationships (One-to-One, One-to-Many, Many-to-One, Many-to-Many) are generated within the **Relation Engine**.

### Dynamic CRUD Engine & Query Engine
Defined in `internal/services/dynamic_data_service.go`. Auto-generates standard CRUD operations and maps them to endpoints:
- `POST /api/v1/data/{slug}`
- `GET /api/v1/data/{slug}/{id}`
- `PUT /api/v1/data/{slug}/{id}`
- `DELETE /api/v1/data/{slug}/{id}`
- `GET /api/v1/data/{slug}`

Advanced query support using the **Query Engine**:
- Filters: `?email=john@example.com`
- Sorting: `?sort=created_at:desc`
- Pagination: `?page=1&limit=20`
- Relational Joins: `?joins=user,products`

### Schema Migration Engine
Safe structural changes to the dynamically created schemas. Managed natively by GORM tracking system (`internal/services/migration_service.go`), capturing Add/Drop/Rename Column logic with automated snapshots generated via the Backup Engine prior to DDL execution.

### GraphQL Gateway
Handled via `internal/graphql/gateway.go`. Dynamically builds an Executable Schema utilizing the `github.com/99designs/gqlgen/graphql` library mapped dynamically to service schemas.

### Backup System
Handled via `internal/services/backup_service.go`. Creates data snapshots mapping relational database information or dynamically generated schemas.
- Backup: `POST /api/v1/cms/backup/service/{service_id}`
- Restore: `POST /api/v1/cms/restore/service/{service_id}`

### Observability Integration
End-to-End telemetry integration out of the box.
- Request/Error Logs via **Zap Structured Logging** (`pkg/logger/`).
- Network Distributed Tracing and Query latency telemetry mapped using **OpenTelemetry** (`internal/tracing/`).
- Standard Metrics via **Prometheus** intercepting the Gin HTTP framework (`internal/middleware/prometheus.go`).

## Security & Optimization
- **Security**: Robust Input Validation, JSON Web Token (JWT) Authentication, Gin-based Rate Limiter mapping to the CMS control plane, and automatic SQL Injection protection provided by parameterized SQL queries within the Application Layer (GORM).
- **Optimization**: Connection caching via Database Connection pooling (`internal/database/connection_manager.go`), scalable generic query pagination structures mapping directly to Redis or database caching nodes.

## Project Structure
```
├── cmd
│   └── server
├── docker
├── internal
│   ├── config
│   ├── database
│   ├── graphql
│   ├── handlers
│   ├── middleware
│   ├── models
│   ├── services
│   └── tracing
├── pkg
│   ├── logger
│   └── response
└── tests
    ├── integration
    ├── uat
    └── unit
```

## Unit Testing
Minimum 80% coverage testing Repository, Service, and API handlers. Use `-coverpkg=./...` locally or execute via standard Makefile.
```bash
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Docker Setup
Run the distributed architecture via Compose.
```bash
docker-compose up -d
```

## Swagger Documentation
Generates standardized OpenAPI document specs matching REST handlers.
```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
