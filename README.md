# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend Platform written in Go. It works similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native.
The platform allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

## High-Level Platform Architecture

```text
       Client Applications
                │
   ┌────────────┴─────────────┐
   ▼                          ▼
REST API                 GraphQL API (optional)
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

## Metadata Database Schema

The CMS stores platform metadata. The recommended database is PostgreSQL.

The core internal configuration is stored in the Metadata Database. Key tables include:

- `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
- `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`). Fields include:
  - `id`
  - `name`
  - `database_id`
  - `created_at`
  - `updated_at`
- `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
- `relations`: Handles relationship configurations between services.
- `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- `backups`: Stores snapshot records or schema outputs.
- `users` & `roles`: Authentication and authorization tables.
- `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure.

```text
.
├── ARCHITECTURE.md
├── Dockerfile
├── Makefile
├── cmd
│   └── server
│       └── main.go
├── docker
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── prometheus.yml
├── docker-compose.yml
├── go.mod
├── go.sum
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

## CRUD Engine Implementation

When a service is created, the system automatically generates CRUD endpoints.
Managed by `internal/services/dynamic_data_service.go` and `internal/handlers/dynamic_data_handler.go`.

**Operations:**
- Create
- Read
- Update
- Delete
- List

**Example endpoints:**
- `POST /api/v1/data/{slug}`
- `GET /api/v1/data/{slug}/{id}`
- `PUT /api/v1/data/{slug}/{id}`
- `DELETE /api/v1/data/{slug}/{id}`
- `GET /api/v1/data/{slug}`

Maps endpoints to standard GORM database operations on the fly.
Applies automated filtering, dynamic joining, and pagination.
Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## Schema Migration Engine

The system supports safe schema changes. Managed by `internal/services/migration_service.go`.

**Supported operations:**
- Add column
- Drop column
- Rename column
- Change column type

Tracks and applies safe schema diffs using GORM’s `.Migrator()`.
Logs migration status natively into the `migrations` table and handles rollbacks through snapshot retention logic.

## GraphQL Gateway

GraphQL is automatically generated from service schemas. Managed by `internal/graphql/gateway.go`.
Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.

Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

## Backup System

The platform supports backing up and restoring data. Managed by `internal/services/backup_service.go`.

- Table backup
- Schema backup
- Service snapshot

**Example APIs:**
- `POST /api/v1/cms/backup/service/{service_id}`
- `POST /api/v1/cms/restore/{id}`

Generates data snapshots and supports restoring rows via JSON payload decoding.

## Observability Integration

The system supports state-of-the-art observability.

- **Logging**: Zap structured logging is injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into the context.
- **Metrics**: Prometheus metrics track latency and status codes via standard HTTP interceptors. Exported at `/metrics` via `internal/middleware/prometheus.go`.
- **Tracing**: OpenTelemetry (with Jaeger support) is initialized in `internal/tracing/` to wrap SQL commands and network logic.

## Unit Tests

The system includes comprehensive unit testing.
Minimum coverage is set to 80%.
Tests cover Repository, Service, and API handler layers.

**Command:**
```sh
go test ./...
```
Or via Makefile:
```sh
make test
```

## Docker Setup

The system provides containerization via Docker.
- `Dockerfile` at root builds the application.
- `docker-compose.yml` provides a local stack, including the app, databases, Prometheus, and Jaeger.
- Run `make docker-up` or `docker-compose up -d` to start the cluster.

## Swagger Documentation

API Documentation is powered by Swagger / OpenAPI.
Generated via the `swag` CLI.

**Command:**
```sh
make swagger
```
This generates the documentation in `docs/` and is served at `/swagger/index.html`.
