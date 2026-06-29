# Dynamic CMS + API Builder Platform

## Overview
This repository contains a **production-grade Backend Platform (Dynamic CMS + API Builder) written in Go**. It acts as a backend infrastructure generator allowing developers to dynamically create backend services, connect external databases, manage schemas, automatically generate CRUD APIs (REST & optional GraphQL), and ensure systems are observable, robust, and scalable.

It is designed following **Clean Architecture and Domain-Driven Design (DDD)** principles and provides a Go-native, fully self-hosted alternative to platforms like Hasura or Supabase.

---

## 1. System Architecture Explanation

The system is built as a monolithic service composed of several key components interacting to offer Backend-as-a-Service capabilities:

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

### Core Layers:
- **Presentation Layer (`internal/handlers/`, `internal/graphql/`)**: Handles incoming HTTP requests via the Gin web framework and GraphQL queries via gqlgen. It provides the gateway to internal engines.
- **Application Layer (`internal/services/`)**: Contains the business logic orchestrating the Dynamic Data Service, Schema Migration Service, Database Connection Service, and Backup Service.
- **Domain Layer (`internal/models/`)**: Defines the internal metadata definitions that manage the state of dynamically created services.
- **Infrastructure Layer (`internal/database/`, `internal/tracing/`, `pkg/logger/`)**: Manages external database connections, observability pipelines (OpenTelemetry, Prometheus), caching, and logging (Zap).

---

## 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. Tables defined in `internal/models/models.go` include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, username, password, database_name. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, boolean, timestamp, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

---

## 3. Go Project Structure

The project uses a clean modular structure, separating application logic from framework details:

```
.
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── config/          # Configuration loading
│   ├── database/        # Database Connection Manager
│   ├── graphql/         # GraphQL schema & resolvers
│   ├── handlers/        # Gin API controllers (REST & CMS Control Plane)
│   ├── middleware/      # Rate limiting, auth, metrics middlewares
│   ├── models/          # Domain layer schemas (GORM models)
│   ├── services/        # Business logic (CRUD Engine, Service Builder, Migrations, etc.)
│   └── tracing/         # OpenTelemetry setup
├── pkg/
│   ├── logger/          # Zap structured logging
│   └── response/        # Standardized API responses
├── tests/
│   ├── unit/            # Unit testing
│   ├── integration/     # Integration testing
│   └── uat/             # User Acceptance Testing
└── docker/              # Dockerfile and compose configurations
```

---

## 4. CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`, the **Dynamic CRUD Engine** auto-generates REST endpoints for any built service.

**Operations Map**:
- `POST /api/v1/data/{service}` -> Create record
- `GET /api/v1/data/{service}/{id}` -> Read record
- `PUT /api/v1/data/{service}/{id}` -> Update record
- `DELETE /api/v1/data/{service}/{id}` -> Delete record
- `GET /api/v1/data/{service}` -> List records (via Query Engine)

The **Query Engine** works alongside the CRUD engine, mapping standard URL parameters to GORM queries:
- **Filtering**: `?email=john@example.com`
- **Sorting**: `?sort=created_at:desc`
- **Pagination**: `?page=1&limit=20`

---

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`, the system automatically maintains the state of dynamic schemas in the target databases using `GORM`'s migrator.

- Tracks schema versions natively inside the `migrations` metadata table.
- Implements safety by allowing schema snapshots and error tracking.
- Supported operations include adding columns, altering types, and resolving relations for Join tables automatically.

---

## 6. GraphQL Gateway Implementation Details

Handled under `internal/graphql/gateway.go`, the system exposes a single GraphQL endpoint that dynamically resolves queries against user-created models.

- **Library**: `github.com/99designs/gqlgen`
- **Endpoint**: `POST /api/v1/graphql`
- Provides capabilities to query data, mutate records, and query related data from joined tables, adapting the types automatically based on metadata definitions.

---

## 7. Backup System Overview

Managed by `internal/services/backup_service.go`, the backup engine provides resilience to the data models.

- Capable of creating Table backups, Schema backups, and Service Snapshots in JSON format.
- Logs snapshots to the `backups` table natively in the metadata database for simple restore capabilities.

---

## 8. Observability Integration

The platform provides an integrated observability pipeline built in:

- **Logging**: Zap structured logging is utilized (`pkg/logger/`), tracking request logs, error logs, and queries.
- **Metrics**: Prometheus middleware (`internal/middleware/prometheus.go`) collects request latency, error rates, and system performance, exposed via `/metrics`.
- **Tracing**: OpenTelemetry (`internal/tracing/`) ensures network requests and deep database logic are tracked with proper Trace Context injection for easy debugging.

---

## 9. Unit Testing Guidelines

The platform maintains strong robustness with extensive unit, integration, and UAT tests.

- **Minimum Coverage Requirements**: 80%
- Test areas include Repositories, Services, and API Handlers.
- **Command to Run Tests**:
  ```bash
  go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
  ```
- **Note**: The codebase leverages SQLite in-memory databases (`:memory:`) to accurately and rapidly perform isolated tests.

---

## 10. Docker Setup

A multi-container Docker structure guarantees production-readiness. Find configurations inside `/docker` and the root directory.

- `Dockerfile`: Multi-stage build process generating optimized Go binaries.
- `docker-compose.yml`: Easily launches the platform with its associated databases (Postgres, etc.), Prometheus, and caching instances.

---

## 11. Swagger Documentation

API Documentation is powered by Swagger / OpenAPI. To regenerate documentation, you can run:

```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

The UI typically exposes endpoints defined in the `internal/handlers` to allow an interactive view of all CMS APIs and metadata endpoints.
