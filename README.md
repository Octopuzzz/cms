# Dynamic CMS + API Builder Backend Platform

A production-grade, self-hosted Backend Platform and API Builder written natively in Go. Designed similarly to platforms like Hasura or Supabase, it allows developers to dynamically create backend services, database schemas, generate CRUD APIs, and expose optional GraphQL endpoints.

The platform acts as a backend infrastructure generator, emphasizing clean architecture, modularity, scalability, and observability.

## Table of Contents
- [Platform Features](#platform-features)
- [System Architecture Explanation](#system-architecture-explanation)
- [Metadata Database Schema](#metadata-database-schema)
- [Go Project Structure](#go-project-structure)
- [Core Components](#core-components)
  - [CRUD Engine Implementation](#crud-engine-implementation)
  - [Schema Migration Engine](#schema-migration-engine)
  - [GraphQL Gateway](#graphql-gateway)
  - [Backup System](#backup-system)
- [Observability Integration](#observability-integration)
- [Docker Setup](#docker-setup)
- [Swagger Documentation](#swagger-documentation)
- [Unit Tests & Local Development](#unit-tests--local-development)

## Platform Features

* **Database Agnostic:** Connect PostgreSQL, MySQL, MongoDB, or use internal SQLite.
* **Service Builder:** Dynamically generate data models, defining fields, relationships, constraints, and validation rules.
* **Dynamic CRUD APIs:** Auto-generated Create, Read, Update, Delete, and List REST endpoints for every service.
* **Query Engine:** Powerful advanced queries (Filtering, Sorting, Pagination, Nested Joins).
* **GraphQL Gateway:** Auto-generated optional GraphQL endpoints matching dynamic service schemas.
* **Schema Migrations:** Safe schema modifications with native rollbacks and backups.
* **Observability:** Built-in Prometheus metrics, OpenTelemetry traces, and structured Zap logging.
* **Performance:** Implements database connection pooling, query optimization, and rate-limiting.
* **Security:** JWT authentication, RBAC (Role-Based Access Control), input validation, and SQL injection protection.

## System Architecture Explanation

The system is built upon **Clean Architecture** and **Domain-Driven Design (DDD)** principles to separate business rules from implementation details:

1. **Presentation Layer (`internal/handlers`, `internal/graphql`):**
   Handles incoming REST and GraphQL requests using the Gin web framework. Includes JWT verification, Rate Limiting, and CORS middleware.
2. **Application Layer (`internal/services`):**
   Core business logic orchestrating dynamic schema building, executing queries via the CRUD engine, and handling migrations/backups.
3. **Domain Layer (`internal/models`):**
   The internal entities required to map out configurations, definitions, fields, relations, and internal users/roles.
4. **Infrastructure Layer (`internal/database`, `internal/tracing`, `pkg/logger`):**
   Responsible for generic database operations (GORM), observability tools, caching (Redis), and connection management.

### Data Flow
```
Client (REST/GraphQL)
   -> API Gateway (Gin Middleware: Auth, Telemetry, Rate-Limiting)
      -> Handlers (Map routes/mutations to services)
         -> Services (Apply RBAC, prepare schemas, generate queries)
            -> Connection Manager (Route to appropriate tenant/metadata DB)
               -> Database (PostgreSQL/MySQL/MongoDB)
```

## Metadata Database Schema

The CMS manages platform state using an internal Metadata Database. Key tables include:

*   **`users` / `roles`**: Core authentication and RBAC platform control.
*   **`database_connections`**:
    *   `id` (UUID), `name`, `type`, `host`, `port`, `username`, `password`, `database_name`
    *   Manages connection string generation and pooling configs.
*   **`services`**:
    *   `id` (UUID), `name`, `database_connection_id`, `db_table_name`, `created_at`, `updated_at`
    *   Represents a dynamically created data model.
*   **`fields`**:
    *   `id` (UUID), `service_id`, `name`, `type`, `nullable`, `unique`, `default_value`, `index`
    *   Attributes of a service (e.g., string, integer, float, uuid, json, array, timestamp).
*   **`service_permissions`**:
    *   Fine-grained row/column level access bound to a `role_id` and `service_id`.
*   **`relations`**:
    *   Defines One-to-One, One-to-Many, Many-to-One, and Many-to-Many logic.
*   **`migrations` & `backups`**:
    *   State tracking for DDL executions and snapshots.

## Go Project Structure

The platform uses a standard, modular Go layout:

```text
├── cmd
│   └── server/          # Main application entrypoint
├── internal
│   ├── config/          # Environment configuration & validation
│   ├── database/        # Connection Manager & multi-DB GORM logic
│   ├── graphql/         # gqlgen gateway, query/mutation resolvers
│   ├── handlers/        # Gin HTTP controllers (REST API)
│   ├── middleware/      # Auth, Logging, Metrics, Rate Limit
│   ├── models/          # Domain entities (Metadata Database Schema)
│   ├── services/        # Business logic (CRUD, Builder, Migration)
│   └── tracing/         # OpenTelemetry setup
├── pkg
│   ├── logger/          # Structured Zap logger wrapper
│   └── response/        # Standardized HTTP API responses
├── tests/               # Unit, Integration, and UAT (E2E) tests
├── docker/              # Additional Docker services (Prometheus)
├── Dockerfile           # Multi-stage production container image
├── docker-compose.yml   # Local deployment cluster
├── Makefile             # Development automation tasks
└── ARCHITECTURE.md      # Detailed system architecture notes
```

## Core Components

### CRUD Engine Implementation

Implemented in `internal/services/dynamic_data_service.go`, the engine handles standard API generation dynamically:

*   **Endpoints Supported:** `GET /api/v1/data/{slug}`, `POST /api/v1/data/{slug}`, `PUT`, `DELETE`.
*   **Query Processing:** Evaluates request context to dynamically apply GORM `.Where()`, `.Order()`, and `.Limit()` clauses based on API queries (e.g., `?email=john@example.com&sort=created_at:desc&page=1&limit=20`).
*   **Relations:** Supports recursive joins via `?join=user,products`. Many-to-many connections automatically generate and traverse join tables.

### Schema Migration Engine

Implemented in `internal/services/migration_service.go`.
*   Tracks field definitions and triggers GORM's `Migrator()` for structural table generation.
*   Supports Safe Schema Changes: Add column, Drop Column, Rename Column, Change Type.
*   Ensures safety by integrating directly with the Backup Engine before any structural adjustments, permitting rollback functionality.

### GraphQL Gateway

Implemented in `internal/graphql/gateway.go` utilizing `gqlgen`.
*   Provides a central `/api/v1/graphql` gateway.
*   Translates dynamic Service queries and schemas into standard GraphQL query schemas on the fly.
*   Allows client-side optimization to request deeply nested relation trees via standard GQL queries.

### Backup System

Implemented in `internal/services/backup_service.go`.
*   Supports service snapshot generation via SQL Dumps and JSON Snapshots.
*   Accessible through the CMS control plane endpoints (`POST /api/v1/cms/backup/service/{id}`).
*   Fully integrates with the `Restore` function to revert tables from captured snapshots during critical migration failures.

## Observability Integration

Comprehensive observability ensures maximum runtime visibility.
*   **Logging:** Zap structured JSON logger used globally (`pkg/logger/`), injecting correlation IDs to trace individual web requests.
*   **Metrics:** Prometheus hooks injected via Gin middleware (`internal/middleware/prometheus.go`). Monitors request latency, error rates, and slow queries. Exposed at `/metrics`.
*   **Tracing:** OpenTelemetry initialized (`internal/tracing/`). Exported to Jaeger to trace service layer spans, database calls, and latency waterfalls.

## Docker Setup

The platform is fully containerized and orchestration-ready.
*   **`Dockerfile`**: A multi-stage build that compiles a minimal static Go binary to reduce footprint and improve security.
*   **`docker-compose.yml`**: Provisions the complete stack locally including the CMS Backend, a persistent Postgres instance, Redis for caching, Jaeger for trace visualization, and Prometheus for metrics gathering.
    *   Command: `make docker-up`

## Swagger Documentation

API Documentation is auto-generated using standard Swaggo comments.
*   **Access:** Live API playground accessible at `/api/v1/swagger/index.html`.
*   **Generation:** Managed via the Makefile (`make swagger`), which parses comments inside `cmd/server/main.go` and `internal/handlers` to output JSON definitions into the `docs/` directory. (Note: `docs/` is ignored by Git to prevent merge conflicts).

## Unit Tests & Local Development

The platform is rigorously tested focusing on Repositories, Services, and Handlers. The target is 80%+ coverage.

*   **Test Suite:** Unit Tests use in-memory SQLite (via `github.com/glebarez/sqlite`) mapping `database.TestConnection` to ensure no state conflicts.
*   **Run All Tests:** `make test`
*   **Run Coverage:** `make test-coverage`
*   **Run Development Server:** `make dev` (requires `air` for hot reloading).

> Note: To build or test locally, ensure `GOTOOLCHAIN=local` is set or Go version `1.24` is utilized as per `go.mod`.
