# Dynamic CMS + API Builder Platform

## Overview
This repository contains a production-grade Backend-as-a-Service (BaaS) platform written natively in Go. It acts as a dynamic backend infrastructure generator, allowing developers to connect databases, build data models, and automatically provision REST & GraphQL APIs.

---

## 1. System Architecture Explanation
The system follows **Clean Architecture** and **Domain-Driven Design (DDD)**.
- **Presentation Layer:** The API Gateway utilizing Gin for REST and gqlgen for GraphQL mapping.
- **Application Layer:** Business logic engines handling dynamic service creation, CRUD operations, relationships, and caching.
- **Domain Layer:** Core internal entities governing the metadata state (Services, Fields, Relationships, RBAC Roles).
- **Infrastructure Layer:** Database connectors (PostgreSQL, MySQL, MongoDB, SQLite), Redis caching, OpenTelemetry tracing, and Zap structured logging.

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

---

## 2. Metadata Database Schema
The control plane relies on a metadata database (PostgreSQL recommended, SQLite supported for dev) to manage platform configurations.

### Core Tables
1. **`database_connections`**: Stores external target database credentials (`id`, `name`, `type`, `host`, `port`, `username`, `password`, `database_name`). Connection pooling configs are also defined here.
2. **`services`**: Tracks dynamically generated backend models (`id`, `name`, `database_connection_id`, `db_table_name`, `created_at`, `updated_at`).
3. **`fields`**: Defines model attributes (e.g., `id`, `name`, `type`, `nullable`, `unique`, `default_value`).
4. **`relations`**: Defines One-to-One, One-to-Many, Many-to-One, and Many-to-Many associations.
5. **`migrations`**: Logs schema migration history and rollback capabilities.
6. **`backups`**: Records snapshot states.
7. **`users` & `roles`**: RBAC permissions for the CMS Control Plane.

---

## 3. Go Project Structure
The structure strictly adheres to Go modular standards:

```text
├── cmd
│   └── server          # Main entrypoint (main.go)
├── internal
│   ├── api             # HTTP handlers and routers
│   ├── cms             # Control plane logic
│   ├── crudengine      # Dynamic data operations
│   ├── queryengine     # Parsing, filtering, sorting, pagination
│   ├── database        # Connection managers and poolers
│   ├── servicebuilder  # DDL translation and schema definitions
│   ├── schema          # Models & domain logic
│   ├── backup          # Backup and restoration services
│   ├── graphql         # gqlgen schema definitions and resolvers
│   └── logger          # Zap structured logging wrapper
├── pkg
│   ├── pagination      # Pagination utilities
│   ├── validation      # Input validation logic
│   └── errors          # Standardized error formatting
├── tests               # Unit and Integration tests
├── docker              # Deployment configuration
├── docker-compose.yml
├── Makefile
└── README.md
```

---

## 4. CRUD Engine Implementation
The dynamic CRUD engine (`internal/crudengine`) automatically intercepts requests for registered services.
- **Create:** `POST /api/v1/{service}`
- **Read:** `GET /api/v1/{service}/{id}`
- **Update:** `PUT /api/v1/{service}/{id}`
- **Delete:** `DELETE /api/v1/{service}/{id}`
- **List (Query Engine):** `GET /api/v1/{service}` supports advanced operations:
  - Filtering: `?email=john@example.com`
  - Sorting: `?sort=created_at:desc`
  - Pagination: `?page=1&limit=20`
  - Joins: `?join=user,products`

---

## 5. Schema Migration Engine
The schema migration engine allows safe schema mutations via the Control Plane API:
- **Supported DDL:** Add column, drop column, rename column, modify types.
- **Safety Features:** Integration with the `backup` engine creates an automated schema/table snapshot before executing structural changes.
- **Execution:** Triggered implicitly by the Service Builder or explicitly via `POST /cms/migrations`.

---

## 6. GraphQL Gateway
The platform supports automated, on-the-fly GraphQL endpoints via `gqlgen`.
- **Location:** Managed under `internal/graphql/`.
- **Functionality:** Dynamically generates queries, mutations, and deep relational traversals mapped directly from the `services` metadata.
- **Endpoint:** `POST /api/v1/graphql`

---

## 7. Backup System
Implemented in `internal/backup/`, the backup engine supports:
- **Table Backup:** Extracting raw rows into JSON payloads.
- **Service Snapshot:** Capturing schema states + data logic concurrently.
- **Endpoints:**
  - `POST /cms/backup/service/{service_id}`
  - `POST /cms/restore/service/{service_id}`

---

## 8. Observability Integration
The platform offers out-of-the-box, full-stack observability.
- **Logging:** Structured JSON request, error, and query logging via `go.uber.org/zap`.
- **Tracing:** Distributed tracing mapped via OpenTelemetry (`go.opentelemetry.io`). Trace IDs inject seamlessly into database layers.
- **Metrics:** Prometheus endpoint exposed at `/metrics` tracking request latency, slow queries, and active HTTP connections.

---

## 9. Unit Tests
The project maintains a **>80% test coverage** requirement.
- Coverage includes `Repository`, `Service`, and `API handlers`.
- Use the provided Makefile to run tests:
  ```bash
  make test-unit
  make test-coverage
  ```
- Uses in-memory SQLite (`:memory:`) for high-speed, isolated unit and integration testing.

---

## 10. Docker Setup
Fully containerized for rapid, platform-agnostic deployments.
- **Dockerfile:** Multi-stage build targeting scratch/alpine containers for minimal footprint.
- **Docker Compose:** `docker-compose.yml` orchestrates the API container, Redis caching layer, and a target PostgreSQL metadata database.
  ```bash
  docker-compose up -d --build
  ```

---

## 11. Swagger Documentation
Standard OpenAPI specifications are auto-generated.
- **Implementation:** Integrated via `github.com/swaggo/swag/cmd/swag`.
- **Generation Command:**
  ```bash
  swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
  ```
- **Live Interface:** Available under `GET /swagger/index.html`.

---
