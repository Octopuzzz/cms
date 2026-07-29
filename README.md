# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade, self-hosted Backend-as-a-Service platform written in Go. It enables dynamic service creation, automated CRUD REST and GraphQL API generation, schema migrations, and advanced observability.

## System Architecture Explanation

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer**: The Gin HTTP Router and Handlers map requests to internal services, including GraphQL endpoints.
- **Application Layer**: Business logic services control data access, generate dynamic schemas, and perform operations.
- **Domain Layer**: Core data models define entities like `Service`, `Field`, `DatabaseConnection`, `User`, and `Role`.
- **Infrastructure Layer**: Cross-cutting tools for connection caching, OpenTelemetry tracing, Prometheus metrics, and Zap structured logging.

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

## Metadata Database Schema

The core configuration is stored in a Metadata Database (PostgreSQL recommended). Key tables include:
- `database_connections`: Stores external DB configurations and connection pooling settings.
- `services`: Represents user-created data models, linking to `database_connections`.
- `fields`: Defines attributes (string, integer, float, uuid, JSON) for each service.
- `service_permissions`: Connects `Role` to `Service` for fine-grained access control.
- `migrations`: Tracks DDL executions for history and rollbacks.
- `backups`: Stores snapshot records and schema outputs.
- `users` / `roles`: Authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of structural and data modifications.

## Go Project Structure

The project uses a clean modular structure:

```text
.
├── cmd/
│   └── server/          # Main application entrypoint
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection management
│   ├── graphql/         # GraphQL gateway and schemas
│   ├── handlers/        # API route handlers (Presentation)
│   ├── middleware/      # Auth, rate limiting, metrics
│   ├── models/          # Domain data structures
│   ├── services/        # Business logic layer (Application)
│   └── tracing/         # OpenTelemetry setup
├── pkg/
│   ├── logger/          # Zap structured logging
│   └── response/        # Standardized API responses
├── tests/               # Unit, Integration, and UAT tests
└── docker/              # Docker and Docker Compose files
```

## CRUD Engine Implementation

The Dynamic CRUD Engine automatically maps REST endpoints (e.g., `GET /api/v1/data/{slug}`) to standard GORM database operations on the fly. It supports:
- Automated filtering via query parameters (e.g., `?email=test@test.com`)
- Dynamic joining via foreign key mappings
- Pagination logic
- Role-Based Access Control and Row-Level filtering dynamically applied prior to queries.

## Schema Migration Engine

The platform includes a safe schema migration engine that:
- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM's Migrator.
- Logs migration status natively into the `migrations` table.
- Handles rollbacks through snapshot retention logic.

## GraphQL Gateway

A dynamically generated GraphQL Gateway provides an optional API layer:
- Generates generic optional schemas using `gqlgen`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

## Backup System

The Backup Engine provides data snapshot capabilities:
- Generates generic service record backups mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.
- Exposes endpoints like `POST /api/v1/cms/backup/service/{service_id}`.

## Observability Integration

State-of-the-art observability is built into the platform:
- **Logging**: Zap structured logging injected globally with context-aware correlation IDs.
- **Metrics**: Prometheus tracks latency and status codes via HTTP interceptors (`/metrics`).
- **Tracing**: OpenTelemetry wraps SQL commands and network logic.

## Unit Tests

Unit tests are standard in the project, achieving over 80% coverage across Repository, Service, and API handler layers.
- Command: `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...` (Ensure to use the memory SQLite driver for DB mocking).

## Docker Setup

The platform is fully containerized and cloud-ready.
- Use `docker-compose up -d` to launch the API alongside its dependency stack.
- Configured using standard `Dockerfile` and `docker-compose.yml` for reproducible builds.

## Swagger Documentation

API Documentation is automatically generated.
- Generate docs using: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`
- Access via `GET /swagger/index.html` after the server starts.
