# Dynamic CMS & API Builder - Backend Platform

This platform is a **fully self-hosted and Go-native Backend-as-a-Service**. It acts as a backend infrastructure generator, allowing users to dynamically create backend services, schemas, REST APIs, and optional GraphQL endpoints.

The system is modular, scalable, and cloud-ready, comparable to platforms like Hasura or Supabase.

---

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles, separating concerns across distinct layers to ensure maintainability and testability.

### Core Layers
*   **Presentation Layer (`internal/handlers/`, `internal/graphql/`)**: The Gin HTTP Router and Handlers act as the API gateway mapping external requests to internal services. Includes optional GraphQL endpoints using `gqlgen`.
*   **Application Layer (`internal/services/`)**: Business logic. Services control data access, generate dynamic schemas, build dynamic queries, and perform operations requested by handlers.
*   **Domain Layer (`internal/models/`)**: The data models. Defines core entities for the platform metadata like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
*   **Infrastructure Layer (`internal/tracing/`, `pkg/logger/`, `internal/middleware/`)**: Cross-cutting tools such as connection caching, telemetry (OpenTelemetry), metrics (Prometheus), and structured logging (Zap).

### High-Level Architecture Diagram

```text
       Client Applications
                │
   ┌────────────┴─────────────┐
   ▼                          ▼
REST API                 GraphQL API (optional)
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

---

## 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database** (recommended: PostgreSQL). The platform uses this schema to manage dynamic resources.

### Core Tables
1.  **`database_connections`**: Stores external DB configurations with fields for `id`, `name`, `type`, `host`, `port`, `username`, `password`, and `database_name`. Manages connection pooling settings.
2.  **`services`**: Represents user-created data models. Includes fields like `id`, `name`, `database_id`, `created_at`, `updated_at`, tracking dynamic schemas.
3.  **`fields`**: Defines attributes for each service (e.g., string, integer, float, boolean, uuid, json). Configures `nullable`, `unique`, `default_value`, and `index`.
4.  **`relations`**: Defines relationships between services (One-to-One, One-to-Many, Many-to-One, Many-to-Many).
5.  **`migrations`**: Keeps track of DDL executions for safe schema changes, rollback capability, and history tracking.
6.  **`backups`**: Stores records of generated backups and snapshot outputs.

---

## 3. Go Project Structure

The project employs a clean modular structure, organizing the codebase logically.

```text
.
├── cmd
│   └── server                # Application entrypoint
├── internal
│   ├── config                # Configuration loader
│   ├── database              # Database connection management
│   ├── graphql               # GraphQL gateway implementation
│   ├── handlers              # REST API controllers
│   ├── middleware            # HTTP middlewares (Auth, Prometheus, Rate limiting)
│   ├── models                # Domain models
│   ├── services              # Business logic (CMS, CRUD, Migrations, Backups)
│   └── tracing               # OpenTelemetry integration
├── pkg
│   ├── logger                # Zap structured logger
│   └── response              # Standardized API responses
├── tests
│   ├── integration           # Integration tests
│   ├── uat                   # End-to-end / User Acceptance Tests
│   └── unit                  # Unit tests
├── docker
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── prometheus.yml
├── Dockerfile
├── Makefile                  # Build and utility commands
└── ARCHITECTURE.md           # Architecture overview
```

---

## 4. CMS Control Plane & Service Builder

The management interface controls the system, handled via `/api/v1/cms/` routes (`internal/services/service_service.go`, `internal/services/dbconn_service.go`).

*   **Database Connectors**: Users register PostgreSQL, MySQL, or MongoDB databases.
*   **Service Builder**: Dynamically creates services containing connection mapping, fields, constraints, and validation rules. Supported field types include string, integer, float, boolean, uuid, json, array, timestamp.
*   **Relation Engine**: Supports relationships including automatically creating join tables for Many-to-Many relationships.

---

## 5. CRUD Engine & Query Engine Implementation

Managed by `internal/services/dynamic_data_service.go`.

*   **Dynamic CRUD Engine**: Automatically exposes RESTful CRUD endpoints for generated services (`POST /api/{service}`, `GET /api/{service}/{id}`, etc.). It maps requests to on-the-fly GORM database operations.
*   **Query Engine**: Applies advanced database queries dynamically. Supports:
    *   **Filtering**: `GET /api/users?email=john@example.com`
    *   **Sorting**: `GET /api/users?sort=created_at:desc`
    *   **Pagination**: `GET /api/users?page=1&limit=20`
    *   **Joins**: Connects records dynamically using defined mappings `GET /api/orders?join=user,products`

---

## 6. Schema Migration Engine

Implemented in `internal/services/migration_service.go`.

*   Provides safe schema changes by tracking and executing schema diffs using GORM’s `.Migrator()`.
*   Supported Operations: Add column, Drop column, Rename column, Change column type.
*   Safety Features: Logs migration status in the `migrations` table natively, and handles rollbacks through pre-migration automatic backup capability.

---

## 7. GraphQL Gateway (Optional)

Managed by `internal/graphql/gateway.go`.

*   A generic, automatically generated GraphQL endpoint (`POST /api/v1/graphql`) using `github.com/99designs/gqlgen/graphql`.
*   Fulfills automated GraphQL APIs for all dynamically defined data models.
*   Supports full Queries, Mutations, and deep Relations.

---

## 8. Backup System

Managed by `internal/services/backup_service.go`.

*   **Backup Engine**: Handles table backups, schema backups, and service snapshots.
*   Snapshots map dynamic table contents into exportable structures (SQL dump / JSON snapshot).
*   Enables endpoint-based backup creation and schema restoration processes.

---

## 9. Observability Integration

Observability is built directly into the infrastructure layer.

*   **Logging**: Uses `go.uber.org/zap` for structured JSON logging. Request logs, error logs, and query logs are injected with correlation and trace IDs tied directly into the Context.
*   **Metrics**: Prometheus (`github.com/prometheus/client_golang`) is used via `internal/middleware/prometheus.go` to expose metrics like request latency, slow queries, and error rates at the `/metrics` endpoint.
*   **Tracing**: OpenTelemetry/Jaeger (`internal/tracing/`) traces and wraps SQL commands, service logic, and network requests.

---

## 10. Performance Optimization & Security

*   **Optimization**: Database connection pooling (in `internal/database/`), Query caching mechanisms using Redis, automated index implementations, and optimized database pagination.
*   **Security**: Implementations include input validation mappings, SQL injection protection through ORM parameterization, JWT authentication, rate limiting, and dynamic Role-Based Access Control (RBAC).

---

## 11. Unit Testing

The system includes a robust test suite covering Repository, Service, and API handlers. The minimum coverage goal is 80%.

To run all tests:
```bash
make test
```
To run unit tests specifically:
```bash
make test-unit
```
To view coverage reports:
```bash
make test-coverage
```

*(Note: Unit tests use an in-memory SQLite database setup `DSN: :memory:` for complete isolation.)*

---

## 12. Docker Setup

The platform is containerized for deployment.

*   `Dockerfile`: Provides the steps to compile the application and build a minimal, production-ready Go binary container.
*   `docker-compose.yml`: (in `docker/` and root directory) Sets up the platform ecosystem, combining the Backend Platform, Postgres database, Redis cache, Jaeger (OpenTelemetry tracing), and Prometheus (Metrics aggregation).

To launch the entire stack:
```bash
make docker-up
# or
docker-compose up -d
```

---

## 13. Swagger API Documentation

API documentation is generated using Swaggo.

To generate the documentation into the `docs/` folder:
```bash
make swagger
# or
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
*(The `docs/` folder is explicitly ignored by `.gitignore`.)*
