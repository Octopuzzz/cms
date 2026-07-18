# CMS Backend Platform

This repository contains a production-grade **Backend-as-a-Service (BaaS) Platform (Dynamic CMS + API Builder)** written in Go. It operates similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native.

The platform allows developers to dynamically create backend services, schema models, automatically generated CRUD APIs, optional GraphQL endpoints, and includes robust management for databases, schemas, backups, and state-of-the-art observability.

## System Architecture Explanation

The CMS Backend strictly follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles to ensure the codebase is modular, scalable, and maintainable.

### Core Layers
- **Presentation Layer (`internal/handlers/`, `internal/graphql/`)**: Provides the API Gateway using the Gin HTTP framework (REST) and `gqlgen` (GraphQL). It maps incoming client requests to application services.
- **Application Layer (`internal/services/`)**: Contains the core business logic, including the Service Builder, CRUD Engine, Query Engine, Schema Migration, and Backup engines. It implements access controls and service orchestrations.
- **Domain Layer (`internal/models/`)**: Defines the internal platform metadata data models such as `Service`, `Field`, `DatabaseConnection`, `User`, and `Role`.
- **Infrastructure Layer (`internal/database/`, `internal/tracing/`, `pkg/logger/`)**: Manages external boundaries like database connection pooling, OpenTelemetry tracing, Prometheus metrics, and Zap structured logging.

### High-Level Architecture
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

The core configuration for the platform is stored in the **Metadata Database**. It tracks all the dynamic models, fields, user details, and system configurations.

Key Tables (mapped in `internal/models/models.go`):
1. **`database_connections`**: Stores external database references (type, host, port, credentials).
2. **`services`**: Represents dynamic data models created by users (linked to `database_connection_id` and the dynamic `db_table_name`).
3. **`fields`**: Attributes associated with a `service` (string, int, JSON, UUID, etc.). Tracks nullability, defaults, unique indices.
4. **`relations`**: Defines relationships (One-to-One, One-to-Many, Many-to-Many) between services.
5. **`migrations`**: Audit log of schema changes (DDL) and snapshot references for safe rollbacks.
6. **`backups`**: Records of database or service snapshots for restoration.
7. **`users` & `roles`**: RBAC configurations for the control plane.
8. **`audit_logs`**: Detailed logs of structural and data changes.

## Go Project Structure

The project maintains a standard Go layout:

```text
.
├── cmd/
│   └── server/
│       └── main.go                 # Main application entry point
├── internal/
│   ├── config/                     # Application configuration (env variables)
│   ├── database/                   # Database Connection Manager (PostgreSQL, MySQL, SQLite, MongoDB)
│   ├── graphql/                    # GraphQL Gateway using gqlgen
│   ├── handlers/                   # Gin HTTP Handlers (Presentation Layer)
│   ├── middleware/                 # Gin Middlewares (Auth, Rate Limiter, Prometheus, Security)
│   ├── models/                     # GORM Models for metadata (Domain Layer)
│   ├── services/                   # Business Logic & Engines (Application Layer)
│   └── tracing/                    # OpenTelemetry implementation
├── pkg/
│   ├── logger/                     # Zap structured logging
│   └── response/                   # Standardized JSON response utilities
├── tests/
│   ├── integration/                # Integration tests for core logic
│   ├── uat/                        # User Acceptance Tests (API endpoints)
│   └── unit/                       # Unit tests for services and models
├── docker/                         # Dockerfiles and compose configs
├── docs/                           # Auto-generated Swagger documentation
├── Makefile                        # Build, run, and test targets
└── go.mod                          # Go modules
```

## Implemented Engines

### 1. CMS Control Plane & Service Builder (`internal/services/service_service.go`, `dbconn_service.go`)
- **Control Plane**: Provides API endpoints (`/api/v1/cms/...`) to connect external databases and manage metadata.
- **Service Builder**: Dynamically registers data models (Services), configures fields, validation rules, and defines relations, pushing changes to the internal metadata DB.

### 2. Dynamic CRUD & Query Engine (`internal/services/dynamic_data_service.go`)
- Auto-generates standard CRUD endpoints for created services (e.g., `GET /api/v1/data/{service_slug}`).
- Includes an advanced Query Engine supporting:
  - **Filtering**: `?email=john@example.com`
  - **Sorting**: `?sort=created_at:desc`
  - **Pagination**: `?page=1&limit=20`
  - **Dynamic Joins**: Supported via relation mapping.

### 3. Schema Migration Engine (`internal/services/migration_service.go`)
- Handles DDL generation and execution (Add/Drop/Rename Column) using GORM's Migrator natively.
- Ensures safe migrations with snapshot capabilities to allow rollback operations.

### 4. GraphQL Gateway (`internal/graphql/gateway.go`)
- Optionally exposed at `/api/v1/graphql`.
- Generates GraphQL schemas automatically matching the defined REST services, allowing clients to query and mutate dynamic data via an integrated Playground (`/api/v1/graphql/playground`).

### 5. Backup System (`internal/services/backup_service.go`)
- Can generate JSON-based table and data snapshots dynamically.
- Implements endpoints like `POST /api/v1/cms/backup/service/{id}` and `POST /api/v1/cms/restore/{id}` to facilitate simple data safety controls.

### 6. Observability Integration (`pkg/logger`, `internal/tracing`, `internal/middleware`)
- **Logging**: High-performance JSON structured logging using **Uber Zap** (`pkg/logger`).
- **Metrics**: **Prometheus** tracks HTTP request latency, error rates, and volume via Gin middleware (`internal/middleware/prometheus.go`), exposed at `/metrics`.
- **Tracing**: Distributed tracing integrated with **OpenTelemetry** and exported to Jaeger (`internal/tracing/tracing.go`), giving end-to-end visibility.

---

## Unit Testing

The repository relies on a robust automated test suite ensuring over 80% coverage across core repository, service, and handler layers.

**To run the test suite (in memory SQLite is used for DB mocks):**
```bash
make test-unit
# or standard go testing
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Docker Setup

The platform is container-ready. Use the provided Docker setup to spin up the API and dependencies (like PostgreSQL and Redis/Jaeger if configured).

**To run via Docker Compose:**
```bash
docker-compose up -d --build
```
This boots the API server alongside necessary required infrastructure configurations based on `docker-compose.yml`.

## Swagger Documentation

Swagger is built directly into the codebase providing complete OpenAPI specifications.
Documentation is served at: `http://localhost:8080/api/v1/swagger/index.html`

**To generate or update the Swagger documentation, ensure you have the CLI installed:**
```bash
go install github.com/swaggo/swag/cmd/swag@latest
~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
