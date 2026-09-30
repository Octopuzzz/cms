# Dynamic CMS & API Builder (Backend Platform)

This repository contains a **production-grade Backend Platform (Dynamic CMS + API Builder)** written in Go. The platform works similarly to Hasura or Supabase, providing a self-hosted, Go-native Backend-as-a-Service (BaaS) solution.

It allows developers to dynamically create backend services, schemas, and APIs, generating both REST and optional GraphQL endpoints on the fly. The system is designed to be modular, scalable, and cloud-ready, adhering to Clean Architecture and Domain-Driven Design (DDD) principles.

---

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles, structured into distinct layers:

- **Presentation Layer**: The API Gateway mapping HTTP and GraphQL requests to internal services using Gin (`internal/handlers/`) and `gqlgen` for GraphQL endpoints.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, build queries, handle migrations, manage backups, and fulfill operations requested by handlers.
- **Domain Layer**: The core data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, and `Role`.
- **Infrastructure Layer**: Cross-cutting utilities such as database connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and structured logging (`pkg/logger/`).

### High-Level Architecture Diagram
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

The core internal configuration is stored in the **Metadata Database**. Tables include:

- `database_connections`: Stores external DB configurations with connection pooling settings (id, name, type, host, port, credentials).
- `services`: Represents user-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
- `fields`: Defines attributes for each service (string, integer, float, uuid, JSON) and configures uniqueness, nullability, and defaults.
- `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
- `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- `backups`: Stores snapshot records or schema outputs.
- `users` and `roles`: General authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of structural and data-level modifications.

---

## 3. Go Project Structure

The project uses a clean modular structure standard for Go applications:

```
.
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── api/             # API routing
│   ├── config/          # Application configuration
│   ├── database/        # Database connection managers
│   ├── graphql/         # GraphQL gateway and schemas
│   ├── handlers/        # Gin HTTP handlers
│   ├── middleware/      # Auth, rate limiting, Prometheus metrics
│   ├── models/          # Domain models
│   ├── services/        # Application business logic (CRUD, backup, migrations)
│   └── tracing/         # OpenTelemetry tracing setup
├── pkg/
│   ├── logger/          # Zap structured logging
│   └── response/        # Standardized API responses
├── tests/
│   ├── integration/     # Integration tests
│   ├── uat/             # User Acceptance Testing
│   └── unit/            # Unit tests
├── docker/              # Docker configuration files
├── Dockerfile           # App Dockerfile
├── docker-compose.yml   # Multi-container orchestration
└── Makefile             # Task automation
```

---

## 4. CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`, the **Dynamic CRUD Engine** automatically maps REST endpoints to standard GORM database operations on the fly:

- **Create**: Inserts dynamic payloads into database tables mapped by services.
- **Read**: Retrieves specific records by ID.
- **Update**: Performs dynamic updates on existing records.
- **Delete**: Removes records.
- **List (Query Engine)**: Applies automated filtering via query params (e.g., `?email=test@test.com`), dynamic joining via foreign key mappings, sorting (`?sort=created_at:desc`), and pagination (`?page=1&limit=20`).

The CRUD engine natively verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

---

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`, the **Schema Migration Engine** supports safe schema changes across external connected databases.

- Tracks and applies safe schema diffs (Add Column, Drop Column, Rename Column, Change Column Type) using GORM’s `.Migrator()`.
- Logs migration status natively into the `migrations` metadata table.
- Supports rollback capabilities through snapshot retention logic.
- Creates automatic backups before migration runs for data safety.

---

## 6. GraphQL Gateway

Managed by `internal/graphql/gateway.go`, the system features an optional, fully automated **GraphQL Gateway**:

- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL queries and mutations for defined data models.
- Introspects the generated schemas from the metadata database.
- Fully handles complex nested queries and relations through resolvers.

---

## 7. Backup System

Managed by `internal/services/backup_service.go`, the platform offers native data snapshotting capabilities:

- Generates data snapshots mapping dynamic table contents to snapshots.
- Supports storing generic service record backup features.
- Enables restoring rows via JSON payload decoding.
- APIs for backup (`POST /cms/backup/service/{service_id}`) and restore (`POST /cms/restore/service/{service_id}`).

---

## 8. Observability Integration

The platform provides complete visibility into its operations, implemented across the stack:

- **Logging**: Zap structured logging injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into the Go Context for request/error tracking.
- **Tracing**: OpenTelemetry (`internal/tracing/`) wrapping SQL commands and network logic. Exportable to tools like Jaeger.
- **Metrics**: Prometheus metrics integrated via Gin middleware (`internal/middleware/prometheus.go`). Tracks request latency, error rates, and status codes. Exported at `/metrics`.

---

## 9. Unit Tests

The system includes a comprehensive suite of unit tests for Repositories, Services, and API handlers. The project requires a minimum of **80% code coverage**.

To run all unit tests:
```bash
go test ./...
```

For coverage reporting across the entire project (including `internal/` packages):
```bash
make test-coverage
# or
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

---

## 10. Docker Setup

Containerization is fully supported. The repository includes a `Dockerfile` and `docker-compose.yml` that orchestrates the backend platform alongside dependencies like Prometheus and PostgreSQL.

To build and start the entire stack:
```bash
docker-compose up -d
```
To view logs:
```bash
docker-compose logs -f
```
To stop the services:
```bash
docker-compose down
```

---

## 11. Swagger Documentation

API Documentation is generated using standard OpenAPI specifications via Swagger (`swag`).

To generate/update the Swagger documentation:
```bash
make swagger
# or natively:
~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

The output will be generated in the `docs/` directory, serving as the source of truth for the platform's REST endpoints.
