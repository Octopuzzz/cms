# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade, self-hosted Backend-as-a-Service (BaaS) platform written in Go. It enables developers to dynamically create backend services, schemas, and REST/GraphQL APIs. It acts similarly to platforms like Hasura or Supabase but is fully native to Go and highly modular, scalable, and cloud-ready.

## System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles. The codebase is organized logically to separate concerns and ensure maintainability.

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
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database** (typically PostgreSQL). The schema definitions can be found in `internal/models/models.go` and include the following key tables:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project employs a clean modular structure to keep logic decoupled.

```
├── cmd
│   └── server
│       └── main.go                 # Application entrypoint
├── internal
│   ├── config
│   │   └── config.go               # Configuration management
│   ├── database
│   │   └── connection_manager.go   # Database connection pool manager
│   ├── graphql
│   │   └── gateway.go              # GraphQL schema auto-generation
│   ├── handlers
│   │   └── ...                     # HTTP request handlers (Gin)
│   ├── middleware
│   │   ├── auth.go
│   │   ├── prometheus.go           # Metrics middleware
│   │   └── rate_limiter.go         # Rate limiting logic
│   ├── models
│   │   └── models.go               # GORM models (Domain entities)
│   ├── services
│   │   ├── dynamic_data_service.go # CRUD and Query engines
│   │   ├── service_service.go      # Service Builder
│   │   ├── migration_service.go    # Schema Migration Engine
│   │   ├── backup_service.go       # Backup Engine
│   │   └── ...                     # Other core services
│   └── tracing
│       └── tracing.go              # OpenTelemetry integration
├── pkg
│   ├── logger
│   │   └── logger.go               # Zap logging wrapper
│   └── response
│       └── response.go             # Standardized API response wrappers
├── tests
│   ├── integration                 # Integration tests
│   ├── uat                         # User Acceptance / E2E tests
│   └── unit                        # Unit tests
└── docker
    ├── Dockerfile                  # Application Dockerfile
    ├── docker-compose.yml          # Local development stack
    └── prometheus.yml              # Prometheus configuration
```

## Core Components Implementation

### 1. CMS Control Plane & Service Builder
Managed under `/api/v1/cms/` routes and handled by `internal/services/service_service.go` and `internal/services/dbconn_service.go`. These services:
- Provide an API interface to store service and field definitions in the metadata database.
- Utilize the `database_connections` engine to validate foreign database connectivity.
- Trigger migrations on service model updates to reflect schema changes.

### 2. CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go`.
- Maps standard CRUD operations (`GET`, `POST`, `PUT`, `DELETE` over `/api/v1/data/{slug}`) to dynamic GORM database interactions.
- The Query Engine applies automated filtering via query params (e.g., `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

### 3. Relation Engine
- Supports dynamic relations out of the box in `internal/services/dynamic_data_service.go` using schema fields logic.

### 4. Schema Migration Engine
Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
- Logs migration status natively into the `migrations` metadata table and handles rollbacks through snapshot retention logic.

### 5. GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
- Dynamically generates generic, optional GraphQL schemas utilizing `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL queries and mutations for defined data models seamlessly alongside the REST API.

### 6. Backup System
Managed by `internal/services/backup_service.go`.
- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.

### 7. Observability Integration
Observability is integrated natively across the stack to ensure the platform is production-ready.
- **Zap Logging**: Injected globally via `pkg/logger/` with correlation and trace IDs tied directly into the Go Context.
- **OpenTelemetry/Jaeger**: Initialized in `internal/tracing/` to wrap SQL commands, queries, and network logic for detailed distributed tracing.
- **Prometheus Metrics**: `internal/middleware/prometheus.go` tracks request latency and status codes via standard HTTP interceptors. Exported at the `/metrics` endpoint.

## Getting Started

### Prerequisites
- Go 1.24+
- Docker & Docker Compose (for local services)
- Make (for running build and test commands)

### Docker Setup
To bring up the entire platform, including any necessary database containers (like PostgreSQL) and observability tools (Prometheus, Jaeger):

```bash
make docker-up
```

To take the platform down:

```bash
make docker-down
```

### Building the Platform
To build the Go binary locally:

```bash
make build
```

To run the platform:

```bash
make run
```

### Unit Tests
The system includes comprehensive test coverage. Unit testing uses an in-memory SQLite setup via `github.com/glebarez/sqlite`.

To run all unit tests:
```bash
make test-unit
```

To run all tests and view a coverage report:
```bash
make test-coverage
```

### Swagger Documentation
The platform automatically generates Swagger / OpenAPI documentation using `swaggo`.

To generate the documentation:
```bash
make swagger
```

*Note: Ensure `swag` is installed (`go install github.com/swaggo/swag/cmd/swag@latest`).*
