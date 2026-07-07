# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. It operates similarly to platforms like Hasura or Supabase, providing a fully self-hosted and Go-native solution. It allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

The system is highly modular, scalable, and cloud-ready, serving as a comprehensive backend infrastructure generator.

## Core Technology Stack

- **Language:** Go 1.24+
- **API Layers:** REST (default) and GraphQL (optional via `gqlgen`)
- **Web Framework:** Gin
- **ORM:** GORM
- **Logging:** Zap structured logging
- **Metrics:** Prometheus
- **Tracing:** OpenTelemetry
- **Cache / Query Optimization:** Supported (Redis ready)
- **Containerization:** Docker & Docker Compose
- **API Documentation:** Swagger / OpenAPI

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
PostgreSQL    MySQL       MongoDB / SQLite
```

The platform adheres to **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer:** Handled by Gin HTTP Router and Handlers (`internal/handlers/`), including the GraphQL endpoints.
- **Application Layer:** Contains business logic within `internal/services/`.
- **Domain Layer:** Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, and `Role` within `internal/models/`.
- **Infrastructure Layer:** Implements telemetry, metrics, logging, and connection caching (`internal/tracing/`, `pkg/logger/`, etc.).

## Metadata Database Schema

The core configuration is stored in the **Metadata Database** (managed via GORM in `internal/models/models.go`):

- **database_connections:** Stores external DB configs (id, name, type, host, port, credentials).
- **services:** Represents user-created data models. Links to a DB connection and tracks the dynamic table name.
- **fields:** Defines attributes for services (e.g., string, integer, uuid, JSON). Handles constraints like uniqueness and nullability.
- **service_permissions:** Connects Roles to Services for RBAC (e.g., CanCreate, CanRead).
- **migrations:** Logs schema DDL executions for tracking and rollback.
- **backups:** Stores schema and data snapshots.
- **users** & **roles:** System authentication and authorization.
- **audit_logs:** Tracks modifications across the platform.

## Key Components

### 1. CMS Control Plane & Service Builder
Provides management interfaces (`/api/v1/cms/` and `/api/v1/services/`) to define databases, services, and schemas dynamically. Backed by `ServiceService` and `DBConnService`.

### 2. Database Connection Manager
Allows connection to external databases (PostgreSQL, MySQL, MongoDB, SQLite). Implements robust connection pooling.

### 3. Dynamic CRUD Engine & Query Engine
Automatically generates REST endpoints for models via `DynamicDataService` (e.g., `GET /api/v1/data/{slug}`). Supports advanced features like pagination, filtering (e.g., `?email=test@example.com`), and dynamic joins.

### 4. Schema Migration Engine
Tracks and applies safe schema alterations (Add/Drop/Rename columns) via GORM’s `.Migrator()`, with built-in rollback logic managed by `MigrationService`.

### 5. GraphQL Gateway
Dynamically serves a generic GraphQL schema from `github.com/99designs/gqlgen/graphql` at `POST /api/v1/graphql`, offering flexible querying.

### 6. Backup Engine
Supports service snapshot and table backups via JSON serialization within the `BackupService`. Includes restore functionalities.

### 7. Observability
Fully instrumented using Zap (structured logging with correlation IDs), OpenTelemetry/Jaeger (SQL and HTTP tracing), and Prometheus (metrics exported at `/metrics`).

## Project Structure

```text
├── cmd
│   └── server/main.go       # Application entrypoint
├── internal
│   ├── config/              # Configuration logic
│   ├── database/            # Database connection manager
│   ├── graphql/             # GraphQL gateway implementation
│   ├── handlers/            # HTTP and API handlers (Presentation)
│   ├── middleware/          # Auth, CORS, rate-limiting, metrics
│   ├── models/              # Domain models and DB schema definitions
│   ├── services/            # Business logic and engines (Application)
│   └── tracing/             # OpenTelemetry integration
├── pkg
│   ├── logger/              # Zap logger wrapper
│   └── response/            # Standardized API responses
├── tests
│   ├── integration/         # Integration tests
│   ├── uat/                 # User Acceptance / E2E tests
│   └── unit/                # Unit tests
├── docker
│   └── ...                  # Dockerfiles and Prometheus configs
├── docker-compose.yml       # Complete platform orchestration
└── Makefile                 # Build, test, and utility tasks
```

## Running the Platform

### Docker Compose
Start the complete stack (app, db, redis, jaeger, prometheus):
```bash
make docker-up
```

### Local Development
```bash
# Copy example environment variables
cp .env.example .env

# Run the backend
make run

# Alternatively, run with hot reload (requires air)
make dev
```

## Testing
The project enforces high unit test coverage across Repository, Service, and API handlers.

```bash
# Run all tests
make test

# Generate coverage report
make test-coverage
```

## API Documentation
Swagger/OpenAPI documentation is automatically generated.

```bash
# Generate docs (requires swag)
make swagger
```
The documentation will be available locally via `/swagger/index.html`.
