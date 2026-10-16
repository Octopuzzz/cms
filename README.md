# Backend Platform (Dynamic CMS + API Builder)

This repository contains a production-grade Backend Platform written in Go. It works similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native. The platform acts as a backend infrastructure generator, allowing users to dynamically create backend services, schemas, REST APIs, and optional GraphQL endpoints.

The system is designed to be modular, scalable, cloud-ready, following Clean Architecture and Domain Driven Design principles.

## Core Technology Stack

- **Language:** Go 1.24+
- **Web Framework:** Gin
- **API Layer:** REST (default), GraphQL (optional via `gqlgen`)
- **ORM:** GORM
- **Logging:** Zap structured logging
- **Metrics:** Prometheus
- **Tracing:** OpenTelemetry
- **Cache:** Redis
- **Containerization:** Docker
- **API Documentation:** Swagger / OpenAPI

---

## 1. System Architecture Explanation

The system implements Clean Architecture with four core layers:
1. **Presentation Layer:** The API Gateway handling REST and GraphQL requests via the Gin framework (`internal/handlers`).
2. **Application Layer:** Business logic orchestration, CMS Control Plane, Service Builder, and Query Engine (`internal/services`).
3. **Domain Layer:** Core data structures and metadata database schema (`internal/models`).
4. **Infrastructure Layer:** Database Connectors (PostgreSQL, MySQL, MongoDB, SQLite), telemetry (`internal/tracing`), logging (`pkg/logger`), and metrics (`internal/middleware`).

### High-Level Architecture Diagram

```text
       Client Applications
                │
   ┌────────────┴─────────────┐
   ▼                          ▼
REST API                 GraphQL API (Optional)
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

The CMS manages platform metadata in a primary database (recommended: PostgreSQL). The core tables include:

- **`database_connections`**: Stores configuration for external databases (`id`, `name`, `type`, `host`, `port`, credentials, pooling config).
- **`services`**: Defines user-created data models, linking to a database connection (`id`, `name`, `database_id`, `db_table_name`, `created_at`, `updated_at`).
- **`fields`**: Configures attributes for each service, including types (string, integer, float, boolean, uuid, json, timestamp), nullability, and uniqueness.
- **`relations`**: Manages relationships (One-to-One, One-to-Many, Many-to-One, Many-to-Many).
- **`migrations`**: Tracks executed schema migrations for safety and rollbacks.
- **`backups`**: Stores backup snapshots.
- **`users` / `roles`**: System authentication and authorization (RBAC).

---

## 3. Go Project Structure

The project follows standard Go layout patterns:

```
├── cmd
│   └── server          # Application entrypoint (main.go)
├── docker              # Dockerfile and docker-compose configurations
├── internal
│   ├── config          # Environment and application configuration
│   ├── database        # Database Connection Manager (PostgreSQL, MySQL, SQLite)
│   ├── graphql         # GraphQL gateway generation using gqlgen
│   ├── handlers        # Gin HTTP route handlers (Presentation Layer)
│   ├── middleware      # Auth, CORS, rate limiting, Prometheus metrics
│   ├── models          # Domain entities and metadata schemas
│   ├── services        # Core business logic (CMS, Service Builder, etc.)
│   └── tracing         # OpenTelemetry initialization
├── pkg
│   ├── logger          # Zap structured logger implementation
│   └── response        # Standardized JSON response formatters
└── tests
    ├── integration     # Integration tests
    ├── uat             # User Acceptance Tests
    └── unit            # Unit tests
```

---

## 4. CRUD Engine Implementation

The Dynamic CRUD Engine automatically exposes standard REST endpoints for any dynamically created service:

- `POST /api/v1/data/{service}` (Create)
- `GET /api/v1/data/{service}/{id}` (Read)
- `PUT /api/v1/data/{service}/{id}` (Update)
- `DELETE /api/v1/data/{service}/{id}` (Delete)
- `GET /api/v1/data/{service}` (List)

### Query Engine Features
The `GET /api/v1/data/{service}` endpoint supports advanced querying parameters out-of-the-box:
- **Filtering:** `?email=john@example.com`
- **Sorting:** `?sort=created_at&order=desc`
- **Pagination:** `?page=1&pageSize=20`
- **Join Queries:** `?join=user,products` (automatically resolves foreign keys based on defined relations)

---

## 5. Schema Migration Engine

The platform includes a safe schema migration engine leveraging GORM's auto-migrator:
- Supports adding, dropping, renaming columns, and modifying column types.
- Snapshots the database schema before applying modifications.
- Tracks execution status natively within the `migrations` metadata table.

---

## 6. GraphQL Gateway

The platform optionally generates GraphQL schemas and endpoints dynamically based on the service configurations.
- Handled via `internal/graphql/gateway.go`.
- Powered by `github.com/99designs/gqlgen`.
- Mounts at `POST /api/v1/graphql` with an interactive GraphQL Playground available for testing.

---

## 7. Backup System

The Backup Engine provides snapshotting capabilities for the dynamically generated services:
- **Service Snapshot:** `POST /api/v1/backup/create` generates a JSON backup snapshot of service metadata and records.
- **Restore:** Snapshots can be restored programmatically to revert accidental data deletions.

---

## 8. Observability Integration

The platform is built with full observability out-of-the-box:
- **Logging:** Structured logging provided by Zap (`pkg/logger`). Error, Request, and Query logs are generated globally.
- **Metrics:** Prometheus metrics are tracked via Gin middleware (`internal/middleware/prometheus.go`), exposing request latency, error rates, and total requests at `/metrics`.
- **Tracing:** OpenTelemetry (OTEL) is initialized (`internal/tracing/`) to wrap SQL queries and HTTP requests, exporting traces to Jaeger.

---

## 9. Unit Tests

Unit, Integration, and UAT tests are included to ensure stability.
- The project mandates a minimum of 80% test coverage across Repository, Service, and API Handler layers.
- Uses standard Go testing libraries and in-memory SQLite (`github.com/glebarez/sqlite`) for mocking database interactions.

**Running Tests:**
```bash
make test-coverage
# or
go test ./... -v -coverpkg=./... -coverprofile=coverage.out
```

---

## 10. Docker Setup

The platform is fully containerized. A `docker-compose.yml` is provided for local deployment, which spins up the Backend Core alongside a Prometheus instance for metrics scraping.

**To run via Docker:**
```bash
docker-compose up --build
```

---

## 11. Swagger Documentation

API Documentation is managed via `swag` annotations in the handler layer.

**Generate Swagger Docs:**
```bash
~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

The documentation is accessible via the built-in Swagger UI route when the server is running.
