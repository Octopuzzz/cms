# CMS Backend Platform

This repository contains a **production-grade Backend Platform (Dynamic CMS + API Builder) written in Go**. It is a self-hosted, Go-native platform that dynamically generates REST and GraphQL APIs, allowing users to connect databases, create data models dynamically, generate CRUD APIs, manage schema migrations, monitor systems, and scale services.

## Core Technologies
*   **Language:** Go (1.24)
*   **API Layer:** REST (Gin), GraphQL (gqlgen)
*   **ORM:** GORM
*   **Logging:** Zap structured logging
*   **Metrics:** Prometheus
*   **Tracing:** OpenTelemetry
*   **Containerization:** Docker
*   **API Documentation:** Swagger / OpenAPI

## System Architecture

The platform follows **Clean Architecture** and **Domain-Driven Design (DDD)**.

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
1.  **Presentation Layer:** Gin HTTP Router and Handlers mapping requests to internal services, including GraphQL endpoints.
2.  **Application Layer:** Business logic handling data access, dynamic schemas, and operations.
3.  **Domain Layer:** Data models for core entities like Service, Field, DatabaseConnection, User, Role.
4.  **Infrastructure Layer:** Cross-cutting concerns like caching, telemetry, metrics, and logging.

## Metadata Database Schema

The CMS stores platform metadata in a primary database (PostgreSQL/SQLite). Core tables include:

*   **`database_connections`**: Stores external database configurations (host, port, credentials).
*   **`services`**: Represents user-created data models, tracking dynamic schemas.
*   **`fields`**: Defines attributes for each service (type, unique, nullable).
*   **`service_permissions`**: Role-based access control linking roles to services.
*   **`migrations`**: Tracks execution history and schema diffs.
*   **`backups`**: Stores snapshot records.
*   **`users` & `roles`**: Auth and authorization.

## Go Project Structure

The project uses a clean modular structure:

```
cmd/
  server/           # Application entrypoint
internal/
  config/           # Environment configuration
  database/         # Connection management
  graphql/          # gqlgen Gateway implementation
  handlers/         # Gin HTTP Handlers (Presentation)
  middleware/       # Auth, Rate limiting, Tracing
  models/           # Domain Entities
  services/         # Business Logic (Application Layer)
  tracing/          # OpenTelemetry setup
pkg/
  logger/           # Zap logging setup
  response/         # Standardized API responses
tests/
  unit/             # Unit tests
  integration/      # Integration tests
  uat/              # End-to-end tests
docker/             # Docker configurations
```

## Component Implementations

### 1. CMS Control Plane & Service Builder
Managed under `/api/v1/cms/` routes, this component allows users to define schemas and store them in the metadata database. It utilizes the connection manager to validate database connectivity and triggers migrations to reflect new data models.

### 2. CRUD Engine & Query Engine
When services are defined, dynamic REST endpoints are automatically resolved via `internal/services/dynamic_data_service.go`. It dynamically maps endpoints to GORM database operations, providing filtering (e.g., `?email=test`), sorting, and pagination logic on the fly while verifying RBAC.

### 3. Schema Migration Engine
Tracks and applies schema modifications (Add, Drop, Rename Column) using GORM’s Migrator. It securely applies diffs to connected databases and records statuses in the internal `migrations` table.

### 4. GraphQL Gateway
A dynamic GraphQL server using `gqlgen`. It parses defined schemas and resolves standard CRUD operations via the same dynamic query engine that powers the REST API, available at `/api/v1/graphql`.

### 5. Backup System
Implemented in `internal/services/backup_service.go`, it supports creating JSON snapshots of dynamic data, storing backup metadata, and restoring records back into the dynamic tables.

### 6. Observability Integration
The platform offers full telemetry:
*   **Structured Logging:** Request, query, and error logs using `zap`.
*   **Metrics:** Prometheus endpoints at `/metrics` tracking latency, slow queries, and error rates via Gin middleware.
*   **Tracing:** OpenTelemetry initialized to wrap SQL operations and HTTP requests.

## Unit Tests
Run the robust test suite providing >= 80% coverage across Repository, Service, and API handler layers:
```bash
make test-coverage
```

## Docker Setup
The platform is fully containerized. A `docker-compose.yml` is provided to spin up the application along with Prometheus for metrics.
```bash
make docker-up
```

## Swagger Documentation
API specs are available using swaggo. Generate them using:
```bash
make swagger
```
