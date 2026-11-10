# Dynamic CMS + API Builder Backend Platform

A production-grade, self-hosted, and Go-native Backend-as-a-Service (BaaS) platform. This platform acts as a backend infrastructure generator allowing developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

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

## System Architecture Principles

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## Core Technologies

- **Language**: Go
- **API Layer**: REST (default) & GraphQL (optional, using `gqlgen`)
- **Web Framework**: Gin
- **ORM**: GORM
- **Logging**: Zap structured logging
- **Metrics**: Prometheus
- **Tracing**: OpenTelemetry
- **API Documentation**: Swagger / OpenAPI
- **Containerization**: Docker

## Project Structure

```
cmd/
└── server/
    └── main.go                  # Application entry point

internal/
├── api/                         # API route definitions
├── config/                      # Configuration management
├── database/                    # Connection manager and DB logic
├── graphql/                     # GraphQL gateway implementation
├── handlers/                    # HTTP handlers
├── middleware/                  # Auth, rate limiting, Prometheus metrics
├── models/                      # Domain models (metadata schema)
├── services/                    # Business logic (Service Builder, CRUD, Migration, etc.)
└── tracing/                     # OpenTelemetry setup

pkg/
├── logger/                      # Zap logging configuration
├── pagination/                  # Pagination helpers
└── response/                    # Standardized Gin response helpers

tests/
├── integration/                 # Integration tests
├── uat/                         # User acceptance tests
└── unit/                        # Unit tests

docker/                          # Docker configuration
```

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**.

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Component Implementation

### 1. CMS Control Plane & Service Builder
Managed under `/api/v1/cms/` routes and handled by `internal/services/service_service.go` and `internal/services/dbconn_service.go`.
- Provides an API interface to store service and field definitions.
- Connects to external databases (PostgreSQL, MySQL, MongoDB).
- Triggers migrations on service model updates.
- Supports dynamically creating fields and relationships (One-to-One, One-to-Many, Many-to-One, Many-to-Many).

### 2. Database Connection Manager
Managed by `internal/database/connection_manager.go`. Users can register external databases (PostgreSQL, MySQL, MongoDB). Connection pooling is implemented.

### 3. Dynamic CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go`.
- Auto-generates CRUD endpoints dynamically (e.g., `POST /api/v1/data/{slug}`).
- Applies automated filtering (e.g., `?email=test@example.com`), sorting (`?sort=created_at:desc`), and pagination (`?page=1&limit=20`).
- Verifies Role-Based Access Control and Row-Level filtering dynamically.

### 4. Schema Migration Engine
Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s Migrator.
- Logs migration status natively and handles rollbacks through snapshot retention.

### 5. GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
- Generates generic optional schemas dynamically using `gqlgen`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

### 6. Backup Engine
Managed by `internal/services/backup_service.go`.
- Generates data snapshots (table backup, schema backup, service snapshot).
- Supports restoring rows via JSON payload decoding.

### 7. Observability
- **Logging**: Request logs, error logs, query logs using Zap (`pkg/logger/`).
- **Metrics**: Request latency, error rate using Prometheus (`internal/middleware/prometheus.go`). Exported at `/metrics`.
- **Tracing**: OpenTelemetry/Jaeger initialized in `internal/tracing/` to wrap SQL commands and network logic.

### 8. Security
- Input validation
- JWT authentication (`internal/middleware/auth.go`)
- Rate limiting
- Role-based access control (RBAC)

## Running the Application

### Using Docker

```bash
docker-compose up -d
```

### Local Development

1. Setup environment variables by copying `.env.example` to `.env`.
2. Install dependencies: `go mod tidy`
3. Run using Makefile: `make dev` (requires `air`) or `go run cmd/server/main.go`

## Testing

The project requires a minimum of 80% unit test coverage.

Run unit tests:
```bash
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

Run all tests via Makefile:
```bash
make test-coverage
```

## API Documentation

Generate Swagger documentation using:
```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
