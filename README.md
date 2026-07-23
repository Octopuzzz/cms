# Dynamic CMS + API Builder Backend Platform

A production-grade, self-hosted Backend Platform written in Go. It allows developers to dynamically create backend services, schemas, REST APIs, and optional GraphQL endpoints, similar to platforms like Hasura or Supabase. The system is modular, scalable, cloud-ready, and strictly follows Clean Architecture and Domain-Driven Design (DDD).

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

*   **Presentation Layer**: The Gin HTTP Router and Handlers. Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
*   **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
*   **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
*   **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

### High-Level Platform Architecture

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

## 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. The tables include:

1.  `database_connections`: Stores external DB configurations (id, name, type, host, port, username, password, database_name). Includes connection pooling settings.
2.  `services`: Represents user-created data models (id, name, database_id, created_at, updated_at). Tracks dynamic schemas (`db_table_name`).
3.  `fields`: Defines attributes for each service (name, type, nullable, unique, default_value, index). Configures uniqueness, nullability, defaults. Supported types: string, integer, float, boolean, uuid, json, array, timestamp.
4.  `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks.
5.  `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6.  `backups`: Stores snapshot records or schema outputs.
7.  `users` and `roles`: General authentication and authorization for the control plane. Join table `user_roles` supports many-to-many.
8.  `audit_logs`: Detailed logging of structural and data-level modifications.

## 3. Go Project Structure

The project uses a clean, modular Go structure:

```
├── cmd
│   └── server
│       └── main.go       # Entry point
├── internal
│   ├── config          # Application configuration
│   ├── database        # Database Connection Manager
│   ├── graphql         # GraphQL gateway (gqlgen)
│   ├── handlers        # Presentation Layer (API Gateway / REST)
│   ├── middleware      # Gin middlewares (Auth, Rate Limiter, Prometheus)
│   ├── models          # Domain Layer (Metadata schemas)
│   ├── services        # Application Layer (CMS Control Plane, Service Builder, CRUD Engine, Query Engine, Schema Migration, Backup)
│   └── tracing         # OpenTelemetry tracing
├── pkg
│   ├── logger          # Zap structured logging
│   └── response        # API Response helpers
├── tests
│   ├── integration     # Integration tests
│   ├── uat             # UAT / E2E tests
│   └── unit            # Unit tests
├── docker
│   ├── docker-compose.yml
│   └── Dockerfile
├── .env.example
├── ARCHITECTURE.md
├── go.mod
├── go.sum
└── Makefile
```

## 4. CRUD Engine Implementation

The **Dynamic CRUD Engine** and **Query Engine** map REST endpoints to standard GORM database operations on the fly.

*   Operations: Create, Read, Update, Delete, List.
*   Endpoints are automatically generated when a service is created:
    *   `POST /api/v1/data/{slug}`
    *   `GET /api/v1/data/{slug}/{id}`
    *   `PUT /api/v1/data/{slug}/{id}`
    *   `DELETE /api/v1/data/{slug}/{id}`
    *   `GET /api/v1/data/{slug}`
*   Advanced Queries:
    *   Filtering: `?email=john@example.com`
    *   Sorting: `?sort=created_at:desc`
    *   Pagination: `?page=1&limit=20`
    *   Join queries for relations.
*   Security: Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

## 5. Schema Migration Engine

The platform includes a safe schema migration engine that tracks and applies DDL operations.

*   **Supported Operations:** Add column, Drop column, Rename column, Change column type.
*   **Safety features:**
    *   Uses GORM's `.Migrator()`.
    *   Automatic backup before migration.
    *   Logs migration status natively into the `migrations` metadata table.
    *   Rollback capability.

## 6. GraphQL Gateway

The **GraphQL Gateway** uses `github.com/99designs/gqlgen` to automatically generate GraphQL schemas from service definitions.

*   Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.
*   Supports queries, mutations, and relations.

## 7. Backup System

The backup system manages backups and snapshots of user data schemas and services.

*   Supports: Table backup, Schema backup, Service snapshot (generic service record backup mapping dynamic table contents to snapshots).
*   Backup Formats: SQL dump, JSON snapshot.
*   API endpoints:
    *   Backup: `POST /api/v1/cms/backup/service/{service_id}`
    *   List Backups: `GET /api/v1/cms/backup/service/{service_id}`
    *   Restore: `POST /api/v1/cms/restore/{id}` (restores rows via JSON payload decoding)

## 8. Observability Integration

The platform provides state-of-the-art observability:

*   **Logging**: `pkg/logger/` provides Zap structured logging, injecting request, error, and query logs globally with correlation and trace IDs tied directly to Context.
*   **Metrics**: `internal/middleware/prometheus.go` tracks request latency, slow queries, and error rates via standard HTTP interceptors. Exported at `/metrics` using Prometheus.
*   **Tracing**: OpenTelemetry/Jaeger is initialized in `internal/tracing/` to wrap SQL commands and network logic.

## 9. Unit Tests Structure and Execution

The system includes robust unit, integration, and UAT tests ensuring quality and correctness.

*   **Test Areas**: Repository, Service, API handlers.
*   **Minimum coverage requirement**: > 80%.
*   **Test execution**:
    *   Unit tests: `make test-unit`
    *   Integration tests: `make test-integration`
    *   UAT tests: `make test-uat`
    *   All tests: `make test` or `go test ./...`
    *   Coverage: `make test-coverage` (Use `-coverpkg=./...` for full coverage reporting)

## 10. Docker Setup

The platform is fully containerized using Docker and Docker Compose.

*   `docker/Dockerfile`: Builds the Go server binary for deployment.
*   `docker/docker-compose.yml`: Provisions the entire platform stack locally, including the backend service, databases, and potentially observability tools (Jaeger, Prometheus).
*   Run the platform using: `make docker-up`

## 11. Swagger Documentation

API documentation is generated using Swagger (OpenAPI).

*   Generate Swagger docs using the command: `make swagger` (or `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`).
*   The API endpoints and models are automatically documented based on Go annotations in the code.
