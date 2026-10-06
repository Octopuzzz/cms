# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. It works similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native. The platform allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

## Core Technology Requirements
* **Language:** Go
* **API Layer:** REST (default), GraphQL (optional via `gqlgen`)
* **Web Framework:** Gin
* **ORM:** GORM
* **Logging:** Zap structured logging
* **Metrics:** Prometheus
* **Tracing:** OpenTelemetry
* **Cache:** Redis
* **Message Queue:** NATS/Kafka (optional)
* **Database Migration:** GORM auto-migrate / Schema Migration Engine
* **Containerization:** Docker
* **API Documentation:** Swagger / OpenAPI

## System Architecture

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)**.

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

### Core Layers
- **Presentation Layer**: Gin HTTP Router and Handlers mapping requests to internal services, including GraphQL endpoints.
- **Application Layer**: Business logic (Services) controlling data access, dynamic schemas, and operations.
- **Domain Layer**: Data models defining entities like Service, Field, DatabaseConnection, User, and Role.
- **Infrastructure Layer**: Telemetry, metrics, logging, and connection caching.

## Metadata Database Schema

The CMS stores platform metadata in the Metadata Database (PostgreSQL recommended, SQLite supported for testing).

1. `database_connections`: Stores external DB configurations (id, name, type, host, port, credentials).
2. `services`: Represents user-created data models, tracks dynamic schemas (`db_table_name`), and links to `database_connections`.
3. `fields`: Defines attributes for each service (type, uniqueness, nullability, defaults).
4. `service_permissions`: Connects `Role` to `Service` for fine-grained access checks.
5. `migrations`: Tracks DDL executions with metadata required for rollbacks.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure:
```
cmd/
└── server/             # Application entrypoint
internal/
├── api/                # API routes and handlers
├── config/             # Configuration management
├── database/           # Database connectors and connection manager
├── graphql/            # GraphQL gateway implementation
├── handlers/           # HTTP handlers for REST API
├── middleware/         # Gin middleware (Auth, Prometheus, Rate Limiter)
├── models/             # Domain models (Metadata schema)
├── services/           # Application layer (CMS, CRUD, Backup, Migration)
└── tracing/            # OpenTelemetry integration
pkg/
├── logger/             # Zap logger wrapper
└── response/           # Standardized API responses
tests/
├── integration/        # Integration tests
├── uat/                # User Acceptance Tests (E2E)
└── unit/               # Unit tests
docker/                 # Dockerfile and Docker Compose configurations
```

## Component Implementations

### CMS Control Plane & Service Builder
Managed via `/api/v1/cms/` endpoints. Provides an interface to store service/field definitions, validating foreign DB connectivity, and triggering migrations on service model updates to reflect schema changes.

### CRUD Engine & Query Engine
Dynamically generates REST endpoints for created services (e.g., `/api/v1/data/{slug}`).
Maps operations to standard GORM queries dynamically, applying automated filtering via query parameters, dynamic joining, and pagination.

### Schema Migration Engine
Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s migrator. Logs migration status natively into the `migrations` table and supports rollback capability.

### GraphQL Gateway
Generates generic optional schemas using `gqlgen`. Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

### Backup System
Generates data snapshots of dynamic table contents. Provides capabilities to restore records via JSON payload decoding.

### Observability Integration
* **Logging**: Zap logging injected globally with correlation and trace IDs.
* **Metrics**: Prometheus metrics export latency and status codes at `/metrics`.
* **Tracing**: OpenTelemetry (Jaeger) wraps SQL commands and network logic.

## Unit Testing
Unit tests, integration tests, and UAT are provided. Minimum target coverage is 80% across Repository, Service, and API handler layers.
Command to run full tests:
```bash
make test
make test-coverage
```

## Docker Setup
The platform is containerized using Docker. A `docker-compose.yml` is provided for running the platform along with its dependencies (Redis, Postgres, Jaeger, Prometheus).
```bash
make docker-up
```

## Swagger Documentation
API documentation is generated using Swagger.
```bash
make swagger
```
The documentation is exposed via Swagger UI.
