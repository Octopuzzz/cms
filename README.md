# Dynamic CMS & API Builder Backend

A production-grade Backend-as-a-Service (BaaS) platform written in Go. This system acts as a backend infrastructure generator allowing developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems.

It works similarly to platforms like Hasura or Supabase but is **fully self-hosted and Go-native**.

## Features

- **Dynamic Data Models**: Create services and schemas dynamically.
- **Auto-generated APIs**: Instantly get REST CRUD endpoints for your services.
- **GraphQL Gateway (Optional)**: Automatically generated GraphQL APIs from your schemas.
- **Database Connection Manager**: Connect to external databases (PostgreSQL, MySQL, MongoDB).
- **Schema Migration Engine**: Safe schema changes (Add, Drop, Rename columns) with rollback capability.
- **Advanced Querying**: Built-in support for filtering, sorting, pagination, and joining.
- **Observability**: Built-in integration with Zap (structured logging), Prometheus (metrics), and OpenTelemetry (tracing).
- **Backup Engine**: Table/schema backups and service snapshots.
- **Security**: JWT authentication, rate limiting, and Role-Based Access Control (RBAC).

## Architecture Overview

The system follows **Clean Architecture** and **Domain Driven Design (DDD)** principles.

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

### Core Layers

- **Presentation Layer**: Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen` (`internal/graphql/gateway.go`).
- **Application Layer**: Business logic (`internal/services/`). Controls data access, generates dynamic schemas, and performs operations requested by handlers.
- **Domain Layer**: Data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting concerns such as connection caching, telemetry (`internal/tracing/`), metrics (`internal/middleware/prometheus.go`), and logging (`pkg/logger/`).

## Metadata Database Schema

The CMS stores its platform metadata (configuration) in a relational database. Core tables include:

- `database_connections`: External DB configurations (id, name, type, host, port, credentials).
- `services`: User-created data models (services). Contains relations to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
- `fields`: Attributes for each service (type: string, integer, etc., nullable, unique, default_value).
- `service_permissions`: Connects `Role` to `Service` for RBAC.
- `migrations`: Tracks DDL executions and handles rollbacks.
- `backups`: Stores snapshot records or schema outputs.
- `users` / `roles`: Authentication and authorization tables for the control plane.
- `audit_logs`: Detailed logging of structural and data-level modifications.

## Component Implementation

### 1. CMS Control Plane & Service Builder
Managed via `/api/v1/cms/` endpoints by `internal/services/service_service.go` and `internal/services/dbconn_service.go`. These allow you to store service definitions, validate external DB connections, and trigger schema migrations.

### 2. Dynamic CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go`. When a service is defined, standard GORM operations are dynamically mapped to endpoints like `GET /api/v1/data/{slug}`. The query engine supports advanced features like filtering (`?email=john@example.com`), sorting, pagination, and dynamic foreign key joining.

### 3. Schema Migration Engine
Managed by `internal/services/migration_service.go`. Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s Migrator.

### 4. GraphQL Gateway (Optional)
Handled in `internal/graphql/gateway.go` utilizing `gqlgen`. It exposes a `POST /api/v1/graphql` endpoint to fulfill queries, mutations, and relations for dynamically defined data models.

### 5. Backup System
Managed by `internal/services/backup_service.go`. Allows for data snapshots (JSON) and restoring rows.

### 6. Observability Integration
Implemented across the stack:
- **Logging**: Zap structured logging with correlation/trace IDs.
- **Metrics**: Prometheus metrics exposed via `/metrics`.
- **Tracing**: OpenTelemetry (Jaeger) wrapping SQL commands and network logic.

## Project Structure

```
.
├── cmd
│   └── server                # Application entrypoint
├── docker                    # Docker configurations
├── internal
│   ├── config                # Environment and app configuration
│   ├── database              # Database connection management
│   ├── graphql               # GraphQL gateway and resolvers
│   ├── handlers              # API Route handlers (REST & Webhooks)
│   ├── middleware            # HTTP Middlewares (Auth, Rate Limiter, Prometheus)
│   ├── models                # Metadata Database Schema Models
│   ├── services              # Business logic (CRUD, Migrations, etc.)
│   └── tracing               # OpenTelemetry integration
├── pkg
│   ├── logger                # Zap structured logger wrapper
│   └── response              # API response standardizers
└── tests
    ├── integration           # Integration tests
    ├── uat                   # User Acceptance Testing
    └── unit                  # Unit tests (Mocked dependencies)
```

## Getting Started

### Prerequisites
- Go 1.24
- Docker & Docker Compose (for infrastructure)
- Make

### Running Locally

1. **Clone the repository:**
   ```bash
   git clone <repo-url>
   cd <repo-directory>
   ```

2. **Copy environment variables:**
   ```bash
   cp .env.example .env
   ```

3. **Start infrastructure (Redis, Database, Jaeger, Prometheus):**
   ```bash
   docker-compose up -d
   ```

4. **Run the backend server:**
   ```bash
   make dev
   ```
   Or using standard go run:
   ```bash
   GOTOOLCHAIN=local go run cmd/server/main.go
   ```

### Documentation & Swagger
Swagger documentation is automatically generated.
1. Install swag: `go install github.com/swaggo/swag/cmd/swag@latest`
2. Generate docs: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`
3. The API documentation will be available at `/swagger/index.html`.

### Testing

The project has robust unit testing targeting 80% coverage.

```bash
# Run unit tests
make test-unit

# Run all tests with coverage
make test-coverage
# Or natively:
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## License
MIT
