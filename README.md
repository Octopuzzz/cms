# CMS Backend Platform

This repository houses a **production-grade, Go-native Backend-as-a-Service (BaaS) platform**. Functioning similarly to platforms like Hasura or Supabase, it provides a self-hosted dynamic CMS and API builder. Users can dynamically create backend services, manage data models via metadata, and automatically generate CRUD APIs (REST and GraphQL), along with full observability, migrations, and backups capabilities.

---

## 1. System Architecture Explanation

The platform is designed with **Clean Architecture** and **Domain-Driven Design (DDD)** in mind, separating concerns into clearly defined layers:

*   **Presentation Layer**: Responsible for mapping HTTP requests (via Gin framework) to underlying application services. It manages standard REST APIs and optional GraphQL endpoints via `gqlgen`.
*   **Application Layer**: Contains the core business logic. The `internal/services/` logic (e.g., Service Builder, CRUD Engine, Query Engine) dynamically queries the external connections, parses incoming schemas, and constructs dynamic responses.
*   **Domain Layer**: Manages core internal configuration structs, models, metadata structures, and relations that govern the platform state.
*   **Infrastructure Layer**: Abstracts away third-party tools or external resource calls. Includes connection management, ORM configuration, telemetry via OpenTelemetry/Jaeger, caching, and robust logging via Zap.

### Core Architecture Flow

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
PostgreSQL    MySQL       MongoDB
```

---

## 2. Metadata Database Schema

The platform requires its own internal metadata database (typically PostgreSQL or SQLite for local dev) to store the platform state.

### Key Tables
*   **`database_connections`**: Stores configuration to connect to dynamic external databases (id, name, type, host, port, credentials).
*   **`services`**: Represents the dynamically created user services. Links to a `database_connection_id` and points to the dynamically generated `db_table_name`.
*   **`fields`**: Defines attributes/schema of each service (id, service_id, name, type, nullable, unique, default_value, index). Supported types: string, integer, float, boolean, uuid, json, timestamp, array.
*   **`relations`**: Defines structural relationships between services (e.g., One-to-One, One-to-Many, Many-to-Many). Join tables are auto-managed.
*   **`migrations`**: Logs executed schema migrations to track metadata for rollbacks.
*   **`backups`**: Records snapshot or schema backups with reference to services.

---

## 3. Go Project Structure

The project has a modular, cloud-ready Go structure:

```
.
├── cmd
│   └── server
│       └── main.go                  # App entry point
├── internal
│   ├── api                        # Handlers for specific modules
│   ├── config                     # App configuration (env, defaults)
│   ├── database                   # Connection manager & DB utilities
│   ├── graphql                    # GraphQL schema & gateway
│   ├── handlers                   # REST handlers mapping to services
│   ├── middleware                 # Auth, metrics, rate limiting
│   ├── models                     # Internal domain models
│   ├── services                   # Application business logic (Builders, CRUD, Migrations)
│   └── tracing                    # OpenTelemetry configuration
├── pkg
│   ├── logger                     # Zap logger wrapper
│   └── response                   # Standardized API response format
├── tests
│   ├── unit                       # Isolated layer tests
│   ├── integration                # Inter-component tests
│   └── uat                        # End-to-end / User acceptance tests
├── docker
│   └── prometheus.yml             # Prometheus config for observability
├── Makefile                       # Development tasks (build, test, docker, swagger)
├── Dockerfile                     # Platform containerization image
├── docker-compose.yml             # Full stack orchestrator (App, DB, Redis, Jaeger, Prom)
└── ARCHITECTURE.md                # Deeper architectural notes
```

---

## 4. CRUD Engine Implementation

When a user defines a new service in the Control Plane, the **Dynamic CRUD Engine** acts as the generic interface mapped to REST.

### Endpoints (Example `users` service):
*   `POST /api/users` (Create)
*   `GET /api/users/{id}` (Read)
*   `PUT /api/users/{id}` (Update)
*   `DELETE /api/users/{id}` (Delete)
*   `GET /api/users` (List)

### Query Engine Features
The list endpoint incorporates a generic **Query Engine** for powerful filtering out of the box:
*   **Filtering**: `GET /api/users?email=john@example.com`
*   **Sorting**: `GET /api/users?sort=created_at:desc`
*   **Pagination**: `GET /api/users?page=1&limit=20`
*   **Dynamic Joins**: `GET /api/orders?join=user,products` (nested joins supported).

---

## 5. Schema Migration Engine

A robust mechanism tracks schema changes when updating a `service` schema.

*   **Capabilities**: Add column, Drop column, Rename column, Change column type.
*   **Mechanism**: Handled in `internal/services/migration_service.go`, which interfaces with GORM's `Migrator()` or external utilities. It provides rollback capabilities and prompts for backups prior to structural changes.

---

## 6. GraphQL Gateway

As an alternative to REST, the platform provides an automatically generated GraphQL API using `gqlgen`.

*   **Gateway**: Requests hit `POST /api/v1/graphql` which forwards queries and mutations to the internal models.
*   **Features**: Enables deep nesting, relation querying, and dynamic typing based on the Metadata Database.

```graphql
query {
  users {
    id
    name
    email
  }
}
```

---

## 7. Backup System

A comprehensive utility to safely extract data and schemas dynamically:
*   **Table and Schema Backups**: Configurable via `internal/services/backup_service.go`.
*   **Formats**: Support for raw SQL dumps or standardized JSON payloads depending on configuration.
*   **Interface**: Controlled through the CMS Control Plane API (e.g., `POST /cms/backup/service/{service_id}`).

---

## 8. Observability Integration

Monitoring is built directly into the middleware and network layer:
*   **Logging**: High-performance structured logging with Zap (`pkg/logger/`). All requests log context, latency, error boundaries, and SQL traces.
*   **Tracing**: OpenTelemetry tracks requests. Exported natively to Jaeger (provided in `docker-compose.yml`).
*   **Metrics**: Prometheus middleware instruments request latency, rates, and 4xx/5xx patterns, available at `/metrics`.

---

## 9. Unit Tests

Testing ensures reliability using Go's native testing mechanisms.

*   **Coverage Rule**: Enforced minimum 80% coverage across Repository, Service, and API handlers.
*   **Types**: Includes Unit, Integration, and UAT (User Acceptance Tests).

**Running tests**:
```bash
# Run all tests natively
make test

# Generate a full coverage report
make test-coverage
```

*Note: For reliable local execution under constraints, tests can run against an isolated SQLite in-memory instance (`:memory:`).*

---

## 10. Docker Setup

Containerization allows for quick deployment of the full stack platform.

The `docker-compose.yml` initializes:
1.  **cms-backend**: The Go application image built in a 2-stage Dockerfile.
2.  **postgres**: The default Metadata database.
3.  **redis**: Caching backend (for Query Caching, Connection Pooling logic).
4.  **jaeger**: Local Distributed Tracing collection.
5.  **prometheus**: Local Metrics aggregation.

**Commands**:
```bash
# Boot the entire distributed stack
make docker-up

# Tear down the stack
make docker-down
```

---

## 11. Swagger Documentation

The platform has full OpenAPI representation for the internal CMS Control Plane APIs and generic structures.

**Generating the Docs**:
1. Ensure the CLI is available: `go install github.com/swaggo/swag/cmd/swag@latest`
2. Run generator: `make swagger` (Outputs to `docs/`)

The Swagger endpoint maps the generated interfaces natively via the web framework routers.

---

## Project Status
This system is fully self-hosted, modular, and cloud-ready for developers to provision dynamic backend topologies securely and effectively.