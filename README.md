# Backend Platform (Dynamic CMS + API Builder)

This repository contains a production-grade Backend Platform written in Go. It is a self-hosted, Go-native Backend-as-a-Service (BaaS) platform that dynamically generates REST and GraphQL APIs, following Clean Architecture and Domain-Driven Design (DDD).

The platform allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints. It is modular, scalable, and cloud-ready, functioning similarly to platforms like Hasura or Supabase but completely self-hosted.

## Core Technology Requirements

- **Language:** Go 1.24
- **API Layer:** REST (default)
- **Optional API Layer:** GraphQL (using `gqlgen`)
- **Web Framework:** Gin
- **ORM:** GORM
- **Logging:** Zap structured logging
- **Metrics:** Prometheus
- **Tracing:** OpenTelemetry
- **Database Migration:** Internal Migrator engine based on GORM (`internal/services/migration_service.go`)
- **Containerization:** Docker
- **API Documentation:** Swagger / OpenAPI

## System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

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

## Core Components Implementation

### 1. CMS Control Plane & Service Builder

Managed under `/api/v1/cms/` routes and handled by `internal/services/service_service.go` and `internal/services/dbconn_service.go`. These services:
- Provide an API interface to store service and field definitions in the metadata database.
- Utilize the `database_connections` engine to validate foreign database connectivity.
- Trigger migrations on service model updates to reflect schema changes.

Users can register external databases (PostgreSQL, MySQL, MongoDB) and build dynamic schemas with field configurations like type, nullability, uniqueness, etc.

### 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. Tables defined in `internal/models/models.go` include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

### 3. Database Connection Manager

Users must be able to register external databases. The connection manager supports PostgreSQL, MySQL, and MongoDB. Connection pooling is implemented out of the box using GORM connection configurations.

### 4. Dynamic CRUD Engine & Query Engine

Managed by `internal/services/dynamic_data_service.go`.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- When a service is created, the system automatically generates REST operations (Create, Read, Update, Delete, List).
- Applies automated filtering via query params (e.g. `?email=test@example.com`), dynamic joining via foreign key mappings, and pagination logic.
- Nested joins and relation definitions (One-to-One, One-to-Many, Many-to-One, Many-to-Many) are supported.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

### 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column, Change column type) using GORM’s `.Migrator()`.
- Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.

### 6. GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models. Supports queries, mutations, and relations.

### 7. Backup System

Managed by `internal/services/backup_service.go`.
- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots in JSON snapshot or SQL dump formats.
- Supports restoring rows via JSON payload decoding.

### 8. Observability Integration

Implemented across the stack:
- **Zap Logging**: Injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into Context.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands and network logic.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.

### 9. Project Structure

The project uses a clean modular structure:

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

### 10. Security

The platform implements security best practices including:
- Input validation
- SQL injection protection
- JWT authentication
- Rate limiting

## Local Development & Docker Setup

Docker is configured for containerization. The main configuration can be found in `docker-compose.yml`.

### Docker Setup

```bash
docker-compose up -d
```
This sets up the entire platform including the backend, databases, Redis, Prometheus, and any observability tooling.

### Local Run

The project uses a Makefile for standardized tasks.

To run the server in development mode (with hot-reloading using `air`):
```bash
make dev
```

To build the executable:
```bash
make build
```

## Unit Testing

The system includes unit, integration, and user acceptance tests (UAT), ensuring standard functionality and avoiding regressions.

Minimum required test coverage across Repository, Service, and API handler layers is **80%**.

To run all tests:
```bash
go test ./...
```
Or use the provided Makefile:
```bash
make test-unit
make test-integration
make test-uat
make test-coverage
```

## Swagger Documentation

API Documentation is available via Swagger/OpenAPI.

To generate or update Swagger documentation locally:
```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

Ensure the `swag` CLI is installed:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```
