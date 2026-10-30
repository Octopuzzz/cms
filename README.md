# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend-as-a-Service (BaaS) platform written in Go. Similar to Hasura or Supabase, it allows developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

The system is fully self-hosted, Go-native, modular, scalable, and cloud-ready.

## System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles.

*   **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
*   **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
*   **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
*   **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

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

The core internal configuration is stored in the **Metadata Database**. The models are defined in `internal/models/models.go` and include the following main entities:

1.  `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2.  `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3.  `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4.  `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5.  `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6.  `backups`: Stores snapshot records or schema outputs.
7.  `users` and `roles`: General authentication and authorization for the control plane.
8.  `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project follows a clean, modular structure:

```text
.
├── cmd/
│   └── server/             # Application entrypoint
├── docker/                 # Docker configuration
├── internal/
│   ├── config/             # Application configuration
│   ├── database/           # Database connection manager
│   ├── graphql/            # GraphQL gateway implementation
│   ├── handlers/           # HTTP handlers (Presentation Layer)
│   ├── middleware/         # Gin middleware (Auth, Rate Limiting, Metrics)
│   ├── models/             # Domain models
│   ├── services/           # Business logic (Application Layer)
│   └── tracing/            # OpenTelemetry integration
├── pkg/
│   ├── logger/             # Zap logger wrapper
│   └── response/           # Standardized API responses
├── tests/                  # Unit, Integration, and UAT tests
│   ├── integration/
│   ├── uat/
│   └── unit/
├── .env.example
├── ARCHITECTURE.md
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── Makefile                # Task runner
```

## Component Implementations

### CRUD Engine & Query Engine

Managed by `internal/services/dynamic_data_service.go` and `internal/handlers/dynamic_data_handler.go`.
*   Dynamically maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations based on metadata.
*   Automatically generates Create, Read, Update, Delete, and List operations.
*   Applies automated filtering via query params (e.g., `?email=test@test.com`), sorting, and pagination logic.
*   Verifies Role-Based Access Control (RBAC) and Row-Level filtering dynamically prior to queries.

### Schema Migration Engine

Managed by `internal/services/migration_service.go`.
*   Provides an interface to track and apply safe schema diffs (Add, Drop, Rename Column, Change Column Type).
*   Utilizes GORM's `.Migrator()` for underlying DDL operations.
*   Logs migration status natively into the `migrations` table.
*   Supports rolling back migrations using historical metadata.

### GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
*   Generates generic optional GraphQL schemas from defined service models using `github.com/99designs/gqlgen`.
*   Exposes a central `POST /api/v1/graphql` endpoint for querying and mutating data.
*   Includes a GraphQL Playground available at `GET /api/v1/graphql/playground`.

### Backup System

Managed by `internal/services/backup_service.go`.
*   Provides capabilities to snapshot data tables.
*   Stores generated backups locally or in a configured backup store.
*   Supports restoring database tables via snapshot payloads.

### Observability Integration

*   **Logging:** Zap structured logging is injected globally (`pkg/logger/`) with correlation and trace IDs tied directly into the Context.
*   **Metrics:** Prometheus metrics track request latency and status codes via standard HTTP middleware (`internal/middleware/prometheus.go`). Exported at the `/metrics` endpoint.
*   **Tracing:** OpenTelemetry/Jaeger is initialized in `internal/tracing/` to wrap SQL commands and network logic.

### Security

*   **Authentication:** JWT-based authentication via `internal/services/auth_service.go`.
*   **Authorization:** Role-Based Access Control (RBAC) middleware ensuring precise route-level and data-level access.
*   **Rate Limiting:** IP-based rate limiting implemented in `internal/middleware/ratelimit.go`.
*   **Security Headers:** Middleware implementation for robust security headers.

## Development

### Unit Tests
The project maintains >80% test coverage across Repository, Service, and API handlers.
Tests are executed using `go test ./...` with an in-memory SQLite database setup.

To run the full suite and view coverage:
```bash
make test-coverage
```

### Docker Setup
The platform is container-ready. Use Docker Compose to spin up the application along with dependent services (like Jaeger, Prometheus, Redis, and external databases).

```bash
make docker-up
```

### Swagger Documentation
Swagger API documentation is generated via `swaggo/swag`.
To build or update the documentation:
```bash
make swagger
```
The documentation is served at `/swagger/index.html` when the server is running.
