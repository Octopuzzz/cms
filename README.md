# CMS Backend Platform

A production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. This platform functions as a self-hosted, Go-native Backend-as-a-Service (BaaS), allowing users to dynamically create backend services, schemas, REST APIs, and optional GraphQL endpoints.

The system is modular, scalable, and cloud-ready, adhering to Clean Architecture and Domain-Driven Design (DDD).

## Platform Features
* **Database Connection Manager:** Connect and manage external databases (PostgreSQL, MySQL, MongoDB).
* **Service Builder:** Dynamically define data models (services), fields, and relationships.
* **Dynamic CRUD Engine:** Auto-generate Create, Read, Update, Delete, and List endpoints for defined services.
* **Query Engine:** Support advanced queries including filtering, sorting, pagination, and join queries.
* **Schema Migration Engine:** Apply and track schema changes safely with rollback capabilities.
* **GraphQL Gateway:** Auto-generate optional GraphQL APIs based on defined schemas.
* **Backup Engine:** Perform table and schema backups with JSON and SQL snapshotting capabilities.
* **Observability:** Comprehensive logging (Zap), metrics (Prometheus), and distributed tracing (OpenTelemetry).
* **API Documentation:** Auto-generate Swagger/OpenAPI documentation.

## Core Technologies
* **Language:** Go (1.24)
* **Web Framework:** Gin
* **ORM:** GORM
* **GraphQL Library:** gqlgen
* **Database:** SQLite (Default for Metadata), PostgreSQL, MySQL, MongoDB
* **Observability:** Zap, Prometheus, OpenTelemetry
* **Containerization:** Docker

---

## 1. System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

* **Presentation Layer:** The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
* **Application Layer:** Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
* **Domain Layer:** Data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
* **Infrastructure Layer:** Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

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

## 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. These tables manage the state and configuration of the dynamically generated platform.

* **`database_connections`**: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
* **`services`**: Represents user-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
* **`fields`**: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, and defaults.
* **`service_permissions`**: Connects `Role` to `Service` providing fine-grained access checks. Contains row-level and field-level capabilities.
* **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
* **`backups`**: Stores snapshot records or schema outputs.
* **`users`** and **`roles`**: General authentication and authorization for the control plane.
* **`audit_logs`**: Detailed logging of structural and data-level modifications.

## 3. Go Project Structure

The project follows a clean, modular structure:
```text
.
├── cmd/
│   └── server/             # Application entrypoint (main.go)
├── internal/
│   ├── config/             # Environment & configuration management
│   ├── database/           # Database connection manager & pooling
│   ├── graphql/            # GraphQL schema generation and gateway
│   ├── handlers/           # HTTP handlers (REST API controllers)
│   ├── middleware/         # Auth, Prometheus, Tracing, Rate Limiter
│   ├── models/             # Domain entities and metadata schema
│   ├── services/           # Application business logic (CRUD, Builder)
│   └── tracing/            # OpenTelemetry initialization
├── pkg/
│   ├── logger/             # Structured Zap logging wrapper
│   └── response/           # Standardized API response helpers
├── tests/
│   ├── integration/        # Integration tests
│   ├── uat/                # User Acceptance Tests (E2E)
│   └── unit/               # Unit tests
├── docker/                 # Docker Compose and configs (Prometheus, etc.)
├── Makefile                # Standardized task runner
└── go.mod                  # Go module definition (Go 1.24)
```

## 4. CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go`, the dynamic CRUD engine handles operations without predefined static models.
* Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
* Intercepts and parses dynamic JSON bodies to insert or update data dynamically based on the metadata defined in the `fields` table.
* Handles auto-generated endpoints such as:
  * `POST /api/v1/data/{slug}`
  * `GET /api/v1/data/{slug}/{id}`
  * `PUT /api/v1/data/{slug}/{id}`
  * `DELETE /api/v1/data/{slug}/{id}`
  * `GET /api/v1/data/{slug}`

## 5. Query Engine

The dynamic Query engine is also embedded within `internal/services/dynamic_data_service.go` and handles reading operations.
* **Filtering:** Automatically parses URL parameters (e.g., `?email=test@test.com`) into SQL WHERE clauses.
* **Pagination:** Supports `?page=1&limit=20` using custom pagination logic to retrieve chunked datasets.
* **Sorting:** Support `?sort=created_at:desc`. Validates against allowed columns and dynamically applies ORDER BY clauses.

## 6. Schema Migration Engine

Managed by `internal/services/migration_service.go`.
* Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s `.Migrator()`.
* Creates underlying database tables when a new service is built.
* Logs migration status natively into the `migrations` table and handles rollbacks through snapshot retention logic to ensure database consistency.

## 7. GraphQL Gateway

Managed by `internal/graphql/gateway.go`.
* Dynamically parses metadata schemas (Services, Fields, Relations) to construct a GraphQL schema in-memory.
* Utilizes `github.com/99designs/gqlgen/graphql` (v0.17.44) for query and mutation execution.
* Exposes `POST /api/v1/graphql` to fulfill automated GraphQL queries.
* Supports standard querying of dynamic resources (e.g., `query { users { id, name, email } }`).

## 8. Backup System

Managed by `internal/services/backup_service.go`.
* Allows for complete snapshots of a given service's data.
* Currently implements generic service record backup features mapping dynamic table contents to JSON snapshots.
* Enables restore functionalities by reading snapshot objects, mapping payload data back to SQL queries, and inserting missing rows to reconstruct data state.

## 9. Observability Integration

Modern cloud-native observability is integrated into every layer of the platform:
* **Logging:** Utilizes Zap structured logging (`pkg/logger/`). Request IDs, correlation IDs, and context details are natively chained to ensure clear error tracking.
* **Metrics:** Embedded Prometheus metrics (`internal/middleware/prometheus.go`). Automatically tracks incoming HTTP request latency, status codes, and throughput. Exported via `/metrics`.
* **Tracing:** OpenTelemetry (OTEL) integration (`internal/tracing/`). Spans are automatically initialized upon HTTP requests, wrapping complex logic sequences, and database transactions to isolate performance bottlenecks visually via Jaeger or similar OTEL collectors.

## 10. Unit Tests

The system maintains high code quality standards. Minimum test coverage for core components (Repositories, Services, API Handlers) is > 80%.
* Uses in-memory SQLite (`:memory:`) setups for mocking databases in tests to avoid connection constraints.
* Test commands:
  * Unit Tests: `make test-unit`
  * Full Suite: `make test`
  * Coverage: `make test-coverage`

## 11. Docker Setup

A Docker-ready infrastructure is available.
* A `Dockerfile` provides multi-stage builds optimized for size and security.
* `docker-compose.yml` configures the backend, alongside observability stacks (like Prometheus).
* Start the stack using: `make docker-up` or `docker-compose up -d`.

## 12. Swagger Documentation

API Documentation is auto-generated using standard Swaggo annotations over Gin handlers.
* Generating docs: `make swagger` (requires `swag` CLI installed globally).
* Exposes standard OpenAPI definitions to test and consume standard Platform APIs, Auth endpoints, and Metadata configuration.
