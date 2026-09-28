# Dynamic CMS & API Builder (Backend Platform)

This is a production-grade, self-hosted Backend-as-a-Service platform (Dynamic CMS + API Builder) built natively in Go. The platform enables developers to dynamically create backend services, schemas, and APIs (REST and optional GraphQL) on the fly. It is fully modular, scalable, and cloud-ready, serving as a comprehensive backend infrastructure generator.

## Core Features
* Connect external databases (PostgreSQL, MySQL, MongoDB, SQLite).
* Create data models dynamically via the Service Builder.
* Automatically generate CRUD REST endpoints and optional GraphQL APIs.
* Advanced Query Engine supporting filtering, sorting, pagination, and relational joins.
* Schema Migration Engine for safe, automated database migrations.
* Robust Backup Engine for table, schema, and snapshot backups.
* Built-in observability with Zap structured logging, Prometheus metrics, and OpenTelemetry tracing.

---

## System Architecture Explanation

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles to separate concerns and ensure maintainability.

* **Presentation Layer:** Managed by Gin HTTP handlers (`internal/handlers/`). Acts as the API Gateway routing REST and GraphQL requests.
* **Application Layer:** Contains business logic within `internal/services/`. It handles schema generation, validation, migration execution, and dynamic CRUD operations.
* **Domain Layer:** Defines core internal platform data models like `Service`, `Field`, `DatabaseConnection`, `User`, and `Role` (`internal/models/`).
* **Infrastructure Layer:** Implements data access, external database connection pooling, tracing (`internal/tracing/`), telemetry, logging (`pkg/logger/`), and caching (Redis).

### High-Level Components
1. **CMS Control Plane:** Management interface to connect databases, define models, and configure the platform.
2. **Service Builder:** Dynamically creates services (data models) and sets field/relation constraints.
3. **CRUD Engine:** Exposes automatically generated REST endpoints (`/api/{service}`) for operations.
4. **Query Engine:** Parses and optimizes queries supporting filtering, sorting, and joins.
5. **Schema Migration Engine:** Safely handles model changes using GORM.
6. **Observability Engine:** Integrates global logging, metrics, and distributed tracing.

---

## Metadata Database Schema

The CMS stores all platform metadata in a primary relational database (default PostgreSQL or SQLite). Below are the core tables driving the platform:

* **`databases` (`database_connections`)**: Stores configured external databases (id, name, type, host, port, credentials).
* **`services`**: Represents the user-created data models (services), containing references to `database_connection_id`.
* **`fields`**: Defines attributes for each service (type, constraints, defaults, relations).
* **`service_permissions`**: Stores Role-Based Access Control configurations linking roles to services.
* **`migrations`**: Logs schema migration history and execution metadata.
* **`backups`**: Stores backup records and JSON snapshot metadata.
* **`users` & `roles`**: Core entities for control plane access control and RBAC management.
* **`audit_logs`**: Detailed logs of schema and data-level modifications.

---

## Go Project Structure

The project adopts a modular Go layout optimized for microservice and platform development:

```text
.
├── cmd
│   └── server          # Entry point for the application
├── docker              # Docker and Docker Compose configurations
├── docs                # Auto-generated Swagger/OpenAPI documentation
├── internal
│   ├── config          # Environment configuration loader
│   ├── database        # Connection manager and pooling
│   ├── graphql         # GraphQL gateway and resolvers (gqlgen)
│   ├── handlers        # Gin HTTP handlers (Presentation Layer)
│   ├── middleware      # Auth, Rate Limiter, Prometheus, Tracing
│   ├── models          # Core domain models (Domain Layer)
│   ├── services        # Business logic, engines (Application Layer)
│   └── tracing         # OpenTelemetry setup
├── pkg
│   ├── logger          # Zap structured logging wrapper
│   └── response        # Standardized HTTP response formatting
├── tests
│   ├── integration     # Integration tests
│   ├── uat             # User Acceptance Testing
│   └── unit            # Unit tests for core logic
├── Makefile            # Build, test, and dev automation commands
└── ARCHITECTURE.md     # Primary architectural reference
```

---

## CRUD Engine Implementation

Managed dynamically via `internal/services/dynamic_data_service.go`, the CRUD Engine automatically maps incoming REST requests (e.g., `GET /api/v1/data/{slug}`) to GORM database operations on the targeted dynamic table.

It implements:
* Dynamic query building supporting basic operators (e.g., `?email=test@test.com`).
* Automated sorting and pagination extraction.
* Row-level filtering based on RBAC checks prior to query execution.
* Seamless relational joins configured via the Service Builder.

---

## Schema Migration Engine

Implemented in `internal/services/migration_service.go`. The engine translates service configurations into DDL execution dynamically:
* Tracks schema drift and applies additive changes (e.g., add/drop/rename columns).
* Leverages GORM's `.Migrator()` interface for executing changes across underlying databases.
* Keeps an audit trail natively inside the `migrations` metadata table for snapshot management and rollbacks.

---

## GraphQL Gateway

The platform optionally exposes dynamically built GraphQL queries.
* Handled by `internal/graphql/gateway.go` utilizing `github.com/99designs/gqlgen`.
* Generates generic schemas that expose data models based on the service configurations.
* Accessible via `POST /api/v1/graphql` and supports querying relationships natively.

---

## Backup System

Managed by `internal/services/backup_service.go`, the engine enables:
* Generic service-level table backups mapped to JSON payload snapshots.
* Storage of backup metadata in the `backups` table for auditing and retrieval.
* Decoding mechanisms to restore individual rows from generated snapshots on demand.

---

## Observability Integration

The platform is built with deep observability out-of-the-box:
* **Logging**: Structured, high-performance logging via `uber-go/zap` located in `pkg/logger/`. Trace IDs are injected directly into context.
* **Metrics**: Standard HTTP metrics are collected using custom middleware (`internal/middleware/prometheus.go`) and exported via the Prometheus `/metrics` endpoint.
* **Tracing**: OpenTelemetry (integrated with Jaeger) is configured within `internal/tracing/` to wrap HTTP handlers, service logic, and database spans.

---

## Unit Tests

The system maintains high quality via comprehensive unit tests targeting Repository, Service, and API handler layers.
* Uses standard Go `testing` combined with in-memory SQLite (`github.com/glebarez/sqlite`) databases to verify component logic.
* Achieves an enforced minimum coverage of **80%**.

To run the unit test suite and check coverage:
```bash
make test-coverage
# Or via native Go
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

---

## Docker Setup

The platform is designed to be cloud-ready and containerized.
* Use `docker-compose.yml` to spin up the entire stack including external dependencies (e.g., PostgreSQL, Redis).
* Standard `Dockerfile` builds a lightweight, optimized Go binary executable.

To build and run:
```bash
docker-compose up -d --build
```

---

## Swagger Documentation

The project implements automated API documentation via Swaggo (`swaggo/swag`).
* Declarative comments over API handlers generate an interactive OpenAPI 2.0 interface.
* Generated output is located in `docs/` (ignored from source control to avoid bloat, generated at build time).

To generate the documentation:
```bash
~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```