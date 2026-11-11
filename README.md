# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend-as-a-Service (BaaS) platform written in Go. It operates similarly to platforms like Hasura or Supabase, acting as a fully self-hosted, Go-native backend infrastructure generator.

## Core Capabilities
- Connect to external databases (PostgreSQL, MySQL, MongoDB, SQLite).
- Create data models dynamically via the CMS Control Plane.
- Generate CRUD APIs automatically.
- Generate optional GraphQL APIs.
- Manage schema migrations.
- Monitor logs and performance via observability tools.
- Manage backups and restore snapshots.
- Scale services in a cloud-ready, Dockerized environment.

## Technology Stack
- **Language**: Go (latest stable)
- **API Layer**: REST (Gin Framework) and GraphQL (gqlgen)
- **ORM**: GORM
- **Logging**: Zap structured logging
- **Metrics**: Prometheus
- **Tracing**: OpenTelemetry (Jaeger)
- **Cache**: Redis
- **Containerization**: Docker
- **API Documentation**: Swagger / OpenAPI

---

## 1. High Level Platform Architecture

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

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)**:
- **Presentation Layer**: Gin HTTP Router/Handlers and GraphQL endpoints.
- **Application Layer**: Business logic via Services controlling data access, schema generation, etc.
- **Domain Layer**: Core data models (Service, Field, DB Connection, User).
- **Infrastructure Layer**: Cross-cutting concerns (connection pooling, telemetry, metrics, logging).

---

## 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. The models include:

1. **`database_connections`**: Stores external DB configurations with fields for id, name, type, host, port, credentials.
2. **`services`**: Represents user-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. **`fields`**: Defines attributes for each service (e.g., string, integer, float, uuid, JSON). Configures uniqueness, nullability, defaults.
4. **`service_permissions`**: Connects `Role` to `Service` providing fine-grained access checks (e.g., CanCreate, CanRead).
5. **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. **`backups`**: Stores snapshot records or schema outputs.
7. **`users`** and **`roles`**: General authentication and authorization for the control plane.
8. **`audit_logs`**: Detailed logging of structural and data-level modifications.

---

## 3. Project Structure

```text
.
├── cmd
│   └── server                # Main application entrypoint
├── docker                    # Docker configurations and setup files
├── internal                  # Private application and library code
│   ├── config                # Configuration loading
│   ├── database              # Database Connection Manager (PostgreSQL, MySQL, MongoDB, SQLite)
│   ├── graphql               # GraphQL Gateway (gqlgen)
│   ├── handlers              # Presentation Layer (Gin Handlers)
│   ├── middleware            # Auth, Prometheus, OpenTelemetry Tracing
│   ├── models                # Domain Layer models
│   ├── services              # Application Layer (CMS Control Plane, Service Builder, Backup)
│   └── tracing               # OpenTelemetry initialization
├── pkg                       # Public library code
│   ├── logger                # Zap structured logger integration
│   └── response              # Standardized API response helpers
└── tests                     # Test suite
    ├── integration
    ├── uat
    └── unit
```

---

## 4. CRUD Engine & Query Engine

Managed by `internal/services/dynamic_data_service.go`, this engine:
- Maps endpoints (e.g., `GET /api/v1/data/{slug}`) to standard GORM database operations on the fly.
- Handles automated filtering via query params (e.g., `?email=test@test.com`), sorting, dynamic joining via foreign key mappings, and pagination.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

---

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`, this component:
- Tracks and applies safe schema diffs (Add, Drop, Rename Column, Change column type) using GORM’s `.Migrator()`.
- Logs migration status natively into the `migrations` table and handles rollbacks through snapshot retention logic.

---

## 6. GraphQL Gateway

Managed by `internal/graphql/gateway.go`, this optional gateway:
- Generates generic schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes a `POST /api/v1/graphql` endpoint to fulfill automated GraphQL queries, mutations, and relations for defined data models.

---

## 7. Backup System

Managed by `internal/services/backup_service.go`, the platform provides:
- Table backup, schema backup, and service snapshots.
- Supports generic service record backup features mapping dynamic table contents to snapshots (e.g., JSON snapshots).
- Restoring functionality via JSON payload decoding.

---

## 8. Observability Integration

Implemented across the stack:
- **Logging**: Zap logging (`pkg/logger/`) with correlation and trace IDs tied directly into the Context.
- **Metrics**: Prometheus (`internal/middleware/prometheus.go`) tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.
- **Tracing**: OpenTelemetry/Jaeger (`internal/tracing/`) initialized to wrap SQL commands and network logic.

---

## 9. Unit Testing

The platform maintains high test coverage across Repository, Service, and API handler layers.

**Commands**:
- Run all tests: `make test`
- Run unit tests: `make test-unit`
- Run integration tests: `make test-integration`
- Generate test coverage: `make test-coverage`

To accurately capture coverage across all packages when running tests, the `-coverpkg=./...` flag is utilized.
Unit tests rely on an in-memory SQLite database (`:memory:`) to ensure isolation.

---

## 10. Docker Setup

A complete `docker-compose.yml` is provided for immediate deployment of the platform and its dependencies (PostgreSQL, Redis, Jaeger, Prometheus).

**Commands**:
- Start all services: `make docker-up` or `docker-compose up -d`
- Stop services: `make docker-down` or `docker-compose down`

---

## 11. Swagger / OpenAPI Documentation

API documentation is generated using `swag`.

**Generation Command**:
```bash
make swagger
```
Alternatively:
```bash
~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

The resulting OpenAPI specification is placed in the `docs/` directory, exposing endpoints for the CMS Control Plane, Services, and Dynamic CRUD engines.
