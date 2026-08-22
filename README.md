# Backend Platform (Dynamic CMS + API Builder)

Welcome to the self-hosted, Go-native Backend-as-a-Service (BaaS) platform. This platform functions similarly to Hasura or Supabase, allowing users to dynamically create backend services, schemas, REST APIs, optional GraphQL endpoints, and more.

## Overview

The platform acts as a backend infrastructure generator, enabling you to:
- Connect databases
- Create data models dynamically
- Generate CRUD APIs automatically
- Generate optional GraphQL APIs
- Manage schema migrations
- Monitor logs and performance
- Manage backups
- Scale services

---

## 1. System Architecture Explanation

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)**.

### Core Layers
- **Presentation Layer**: Gin HTTP Router and Handlers (`internal/handlers/`), including GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Controls data access, dynamic schemas, and operations.
- **Domain Layer**: Data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, etc.
- **Infrastructure Layer**: Cross-cutting tools (connection caching, telemetry with OpenTelemetry, Prometheus metrics, Zap logging).

### High Level Platform Architecture
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

The CMS stores platform metadata in the core database (e.g., PostgreSQL).

### Core Tables
- `database_connections`: External DB configurations (id, name, type, host, port, credentials).
- `services`: User-created data models (id, name, database_connection_id, db_table_name, created_at, updated_at).
- `fields`: Attributes for each service (name, type, nullable, unique, default_value, index).
- `service_permissions`: Connects `Role` to `Service` for RBAC.
- `relations`: Defines one-to-one, one-to-many, many-to-one, and many-to-many relations.
- `migrations`: DDL execution history and metadata.
- `backups`: Backup records.
- `users` / `roles`: Auth for control plane.

---

## 3. Go Project Structure

A clean, modular structure is used across the project:

```
├── cmd/
│   └── server/            # Main application entrypoint
├── internal/
│   ├── api/               # API Gateway / Routing
│   ├── config/            # Configuration management
│   ├── database/          # Database Connection Manager
│   ├── graphql/           # GraphQL Gateway
│   ├── handlers/          # HTTP Handlers (Presentation)
│   ├── middleware/        # Auth, Rate Limiter, Prometheus
│   ├── models/            # Core Domain Models
│   ├── services/          # Application Logic (CRUD, Migration, etc.)
│   └── tracing/           # OpenTelemetry integration
├── pkg/
│   └── logger/            # Zap structured logger wrapper
├── tests/                 # Unit, Integration, and UAT tests
├── docker/                # Docker configuration files
├── Dockerfile             # Container definition
├── docker-compose.yml     # Local development environment setup
└── Makefile               # Task runner for build/test
```

---

## 4. CRUD Engine Implementation

The Dynamic CRUD Engine (`internal/services/dynamic_data_service.go`) automatically maps operations to GORM based on dynamically created services.

- **Endpoints**:
  - `POST /api/v1/data/{slug}` (Create)
  - `GET /api/v1/data/{slug}/{id}` (Read)
  - `PUT /api/v1/data/{slug}/{id}` (Update)
  - `DELETE /api/v1/data/{slug}/{id}` (Delete)
  - `GET /api/v1/data/{slug}` (List with pagination, sorting, filtering)

### Query Engine
Supports advanced queries:
- **Filtering**: `?email=test@example.com`
- **Sorting**: `?sort=created_at:desc`
- **Pagination**: `?page=1&limit=20`

---

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`, the engine tracks and applies schema changes dynamically to registered databases.

- **Capabilities**: Add, Drop, Rename columns, and Change column types via GORM's Migrator.
- **Safety**: Records migrations in the `migrations` table to provide history and support rollbacks.

---

## 6. GraphQL Gateway

The system implements a generic GraphQL gateway (`internal/graphql/gateway.go`) using `gqlgen` to fulfill automated GraphQL endpoints.

- Automatically maps dynamic service schemas to GraphQL queries and mutations.
- Accessible via `POST /api/v1/graphql`.
- Provides an alternative to the REST API for fetching deeply nested relations or specific fields.

---

## 7. Backup System

The Backup Engine (`internal/services/backup_service.go`) allows users to safeguard their data models.

- **Features**: Generates data snapshots (JSON/SQL formats).
- **Endpoints**:
  - `POST /api/v1/cms/backup/service/{service_id}`
  - `POST /api/v1/cms/restore/service/{service_id}`

---

## 8. Observability Integration

Comprehensive observability is integrated out-of-the-box.

- **Logging**: Zap structured logging (`pkg/logger/`) with correlation/trace IDs.
- **Metrics**: Prometheus metrics exposed via `/metrics` intercepting HTTP status, latencies, and request rates.
- **Tracing**: OpenTelemetry (integrated via `internal/tracing/`) traces requests from the HTTP layer down to SQL execution.

---

## 9. Unit Tests

The system maintains high unit test coverage (target >80%) across Repositories, Services, and API Handlers.
Tests heavily utilize in-memory SQLite (via `github.com/glebarez/sqlite`) for mocking database interactions.

Run tests using:
```bash
make test-coverage
# or manually:
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

---

## 10. Docker Setup

The platform is fully containerized. A `Dockerfile` is provided for the core application, and `docker-compose.yml` spins up required infrastructure (like the external databases, Redis, etc., depending on configuration).

To run locally with Docker:
```bash
docker-compose up --build
```

---

## 11. Swagger Documentation

API Documentation is auto-generated using standard OpenAPI / Swagger tools.

Generate docs with:
```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
(Note: The generated `docs/` folder is git-ignored to prevent version control pollution).
