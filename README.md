# Dynamic CMS & API Builder (Go Backend Platform)

This is a production-grade Backend-as-a-Service (BaaS) platform, written in Go, functioning as a fully self-hosted, Go-native backend infrastructure generator. It empowers developers to connect databases, build data models, and dynamically generate both REST and GraphQL APIs.

---

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer**: The Gin HTTP Router and Handlers. Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
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
PostgreSQL    MySQL       MongoDB / SQLite
```

---

## 2. Metadata Database Schema

The platform metadata is managed internally in the CMS. Supported databases include PostgreSQL, MySQL, and SQLite.
The core metadata tables include:

1. **`database_connections`**: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. **`services`**: Represents user-created data models. Contains references to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. **`fields`**: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. **`service_permissions`**: Connects `Role` to `Service` providing fine-grained access checks. Contains row-level and field-level capabilities.
5. **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. **`backups`**: Stores snapshot records or schema outputs.
7. **`users`** & **`roles`**: General authentication and authorization for the control plane.
8. **`audit_logs`**: Detailed logging of structural and data-level modifications.

---

## 3. Go Project Structure

The project uses a clean, modular structure following standard Go project layouts:

```text
.
├── ARCHITECTURE.md      # Architecture documentation
├── cmd
│   └── server           # Main application entrypoint
├── docker               # Docker configuration files
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── prometheus.yml
├── internal
│   ├── config           # Configuration management
│   ├── database         # Database connection management
│   ├── graphql          # GraphQL gateway and resolvers
│   ├── handlers         # HTTP API handlers (Gin)
│   ├── middleware       # Auth, telemetry, Prometheus metrics
│   ├── models           # Domain models and schema structures
│   ├── services         # Application layer logic (CRUD, CMS, Auth)
│   └── tracing          # OpenTelemetry integration
├── pkg
│   ├── logger           # Zap structured logging
│   └── response         # Standardized API response helpers
└── tests                # Automated testing
    ├── integration      # Integration tests
    ├── uat              # User Acceptance Testing
    └── unit             # Unit tests
```

---

## 4. CRUD Engine Implementation

The Dynamic CRUD Engine is handled via `internal/services/dynamic_data_service.go`. When a service is defined dynamically, the platform maps generic REST endpoints to the created schemas.
- Supports automated querying, filtering (e.g., `?email=test@example.com`), sorting, and pagination logic via standard GORM database operations.
- Dynamic endpoints automatically follow the CRUD lifecycle: Create, Read, Update, Delete, List.
- Applies Role-Based Access Control and Row-Level filtering dynamically prior to executing queries.

---

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`, the Schema Migration Engine supports safe structure changes:
- Tracks and applies safe schema modifications using GORM's `.Migrator()`, allowing for adding, dropping, renaming, and changing column types.
- Supports robust metadata tracking stored directly in the `migrations` metadata table for history tracking.
- Safety features like snapshots and rollbacks are deeply integrated.

---

## 6. GraphQL Gateway

The GraphQL Gateway generates generic optional schemas seamlessly:
- Utilizing `github.com/99designs/gqlgen/graphql`, the gateway allows fetching structured relations dynamically.
- Managed by `internal/graphql/gateway.go`, users can query dynamic metadata efficiently from generated schemas.
- Exposed natively via `POST /api/v1/graphql`.

---

## 7. Backup System

The Backup Engine (`internal/services/backup_service.go`) guarantees resilience against data loss:
- Safely generates and stores data snapshots.
- Implements generic service record backup features mapping dynamic table contents to snapshots natively.
- Seamlessly supports restoring rows via JSON payload decoding.

---

## 8. Observability Integration

Comprehensive observability covers metrics, tracing, and structured logging:
- **Logging**: Zap logging (`pkg/logger/`) is globally utilized with request logging and error tracking, correlating logs with trace IDs.
- **Metrics**: Prometheus hooks natively to HTTP requests in `internal/middleware/prometheus.go` to measure request latency, error rates, and load. Metrics are exported at `/metrics`.
- **Tracing**: OpenTelemetry (integrated via `internal/tracing/`) traces complex database interactions and API calls securely.

---

## 9. Unit Tests

The backend emphasizes high reliability through automated testing (`tests/unit/`, `tests/integration/`, `tests/uat/`):
- Aiming for over 80% coverage across Repository, Service, and API handler layers.
- Uses standard Go testing tools and seamlessly validates against an in-memory SQLite database (`:memory:`) via the `github.com/glebarez/sqlite` driver for fast execution.
- Command to run all tests with coverage: `make test-coverage`.

---

## 10. Docker Setup

The system includes production-ready Docker configurations in the `docker/` directory:
- Contains an optimized multi-stage `Dockerfile`.
- Uses `docker-compose.yml` to coordinate the Go application, databases (like Postgres/MySQL/MongoDB/Redis), Prometheus, and Jaeger.
- Seamlessly start everything using: `make docker-up` or `docker-compose up -d`.

---

## 11. Swagger Documentation

API documentation is autogenerated natively using `swag`. It builds robust Swagger endpoints mapped from declarative comments on the API handlers.
- Easily available standard interface for the REST APIs.
- Generates swagger documentation executing the command: `make swagger`.
