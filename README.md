# Go-Native Backend Platform (Dynamic CMS + API Builder)

A production-grade Backend-as-a-Service (BaaS) platform written in Go. This self-hosted, cloud-ready platform empowers developers to dynamically create backend services, define database connections, generate CRUD APIs automatically, generate optional GraphQL endpoints, and monitor everything with robust observability.

This system works similarly to Hasura or Supabase but is fully self-hosted and completely Go-native. It is built following Clean Architecture and Domain-Driven Design (DDD) principles.

## Core Technology Stack

- **Language:** Go (1.24)
- **API Layers:** REST (default) and GraphQL (optional via `gqlgen`)
- **Web Framework:** Gin
- **ORM:** GORM
- **Logging:** Zap structured logging
- **Metrics:** Prometheus
- **Tracing:** OpenTelemetry
- **Cache:** Redis (optional integration)
- **Database Migration:** Internal GORM migrator / engine
- **Containerization:** Docker & Docker Compose
- **API Documentation:** Swagger / OpenAPI

---

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles, structured around modular and scalable layers:

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

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

## 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. Tables defined in `internal/models/models.go` include:

- `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
- `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
- `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
- `relations`: (Optional / Internal mapping) Maps one-to-one, one-to-many, many-to-one, many-to-many relationships automatically via foreign keys and join tables.
- `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g., CanCreate, CanRead). Contains row-level and field-level capabilities.
- `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- `backups`: Stores snapshot records or schema outputs.
- `users` and `roles`: General authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of structural and data-level modifications.

---

## 3. Go Project Structure

The project uses a clean modular structure.

```text
.
├── cmd
│   └── server                # Main application entry point
├── internal
│   ├── config                # Configuration parsing (.env)
│   ├── database              # Connection managers and pooling
│   ├── graphql               # gqlgen integration and optional GraphQL API
│   ├── handlers              # Gin HTTP handlers
│   ├── middleware            # Auth, Telemetry, Rate limiting, CORS
│   ├── models                # Domain entities
│   ├── services              # Business logic (CRUD, CMS, Query, Schema, Backup)
│   └── tracing               # OpenTelemetry integration
├── pkg
│   ├── logger                # Zap structured logging wrapper
│   └── response              # API standard response utilities
├── tests                     # Unit, Integration, and UAT tests
│   ├── integration
│   ├── uat
│   └── unit
├── docker                    # Docker setups for components like Prometheus
├── Makefile                  # Build, run, test task runner
├── docker-compose.yml        # Development environment provisioning
└── Dockerfile                # Multi-stage build for the application
```

---

## 4. CRUD Engine & Query Engine Implementation

Managed by `internal/services/dynamic_data_service.go`, this component auto-generates operations for dynamic data models.

- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g., `?email=test@test.com`), dynamic joining via foreign key mappings, sorting (`?sort=created_at:desc`), and pagination (`?page=1&limit=20`).
- Supports nested joins and dynamic role-based access control checking before executing queries.
- Optimizes inserts using `CreateInBatches` to prevent N+1 and performance overhead.

---

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`.

- Tracks and applies safe schema diffs (Add, Drop, Rename Column, Change column type) using GORM’s `.Migrator()`.
- Logs migration status natively into the `migrations` table and handles rollbacks through snapshot retention logic.
- Ensures safety by performing automatic backups before executing migrations.

---

## 6. GraphQL Gateway

Managed by `internal/graphql/gateway.go`.

- Leverages `github.com/99designs/gqlgen/graphql` (v0.17.44).
- Auto-generates optional generic schemas for defined service models, enabling powerful queries and mutations natively via GraphQL.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL queries.

---

## 7. Backup System

Managed by `internal/services/backup_service.go`.

- Supports generating data snapshots: Table backups, Schema backups, and Full Service snapshots.
- Implements features for mapping dynamic table contents to generic JSON snapshots and executing SQL dumps.
- Offers snapshot restoration endpoints that read JSON payload decodings.

---

## 8. Observability Integration

Observability is a core pillar of the platform:

- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Request Contexts. Includes Request logs, Error logs, and Query logs.
- **Prometheus Metrics**: Defined via `internal/middleware/prometheus.go` mapping. Tracks request latency, error rates, and API status codes natively in HTTP handlers. Accessible at `/metrics`.
- **OpenTelemetry/Jaeger**: Wrapped inside `internal/tracing/` to intercept SQL commands and internal handler logic for full distributed tracing capabilities.

---

## 9. Unit Testing

The platform enforces minimum **80% coverage** across Repository, Service, and API handler layers.

- Extensive testing structure mapped out in the `tests/` directory (unit, integration, and uat).
- Built on top of isolated in-memory SQLite instances using `github.com/glebarez/sqlite` to prevent cross-contamination.
- Run tests via the command: `make test-coverage` or `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`.

---

## 10. Docker Setup

A `Dockerfile` and `docker-compose.yml` are provided at the repository root.

- Multi-stage builds are used to minimize image footprint.
- The `docker-compose.yml` file sets up external infrastructure integrations, seamlessly spinning up Redis, PostgreSQL databases, and Prometheus containers to give you a full environment out-of-the-box.

---

## 11. Swagger Documentation

Swagger is integrated for API specification and documentation.

- Generate documentation using: `~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
- The swagger output exposes comprehensive definitions of CMS control plane endpoints such as `POST /api/v1/cms/databases` or `GET /api/v1/cms/services`.
