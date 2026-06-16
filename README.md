# Go Backend Platform (Dynamic CMS + API Builder)

A production-grade Backend-as-a-Service (BaaS) platform written in Go. This platform acts as a backend infrastructure generator, allowing developers to connect databases, dynamically create data models, automatically generate REST and GraphQL APIs, manage schema migrations, and handle backups, all while maintaining state-of-the-art observability.

It is fully self-hosted, Go-native, modular, scalable, and cloud-ready, comparable to platforms like Hasura or Supabase.

## System Architecture

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles, separating concerns into clearly defined layers:

- **Presentation Layer**: Exposes REST endpoints (using Gin) and optional GraphQL endpoints (using gqlgen). Acts as the API gateway.
- **Application Layer**: Contains core business logic. Defines services to control data access, generate schemas dynamically, and perform domain operations.
- **Domain Layer**: Defines core platform entities (`Service`, `Field`, `DatabaseConnection`, `User`, `Role`, etc.).
- **Infrastructure Layer**: Handles external connections, telemetry (OpenTelemetry), metrics (Prometheus), logging (Zap), and connection pooling.

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

## Core Components

### 1. CMS Control Plane & Service Builder
The core management interface for the system.
- Manages external database connections (PostgreSQL, MySQL, MongoDB).
- Dynamically defines data services, schemas, and fields.
- Triggers automatic migrations on model updates to reflect schema changes.

### 2. Metadata Database Schema
The platform uses a metadata database (typically PostgreSQL or SQLite) to store internal configurations.
Key tables include:
- `database_connections`: Stores external DB configs, credentials, and pooling settings.
- `services`: Represents user-created data models, linking to `database_connection_id` and holding the `db_table_name`.
- `fields`: Defines attributes for services (types, uniqueness, nullability, defaults).
- `relations`: Defines one-to-one, one-to-many, many-to-one, and many-to-many relations between services.
- `service_permissions`: Links `Role` to `Service` for fine-grained RBAC.
- `migrations`: Tracks DDL executions for rollbacks and history.
- `backups`: Stores backup snapshot records.
- `users` / `roles`: For control plane authentication.

### 3. Dynamic CRUD Engine & Query Engine
Automatically maps HTTP endpoints to database operations dynamically based on service definitions.
- Generates standard CRUD endpoints (`POST`, `GET`, `PUT`, `DELETE` to `/api/v1/data/{slug}`).
- Supports advanced queries including filtering (`?email=...`), sorting (`?sort=...`), pagination (`?page=1&limit=20`), and dynamic joins.
- Enforces Role-Based Access Control and Row-Level Security prior to executing queries.

### 4. Schema Migration Engine
Provides safe schema modifications.
- Supports adding, dropping, and renaming columns via GORM’s Migrator.
- Logs statuses natively into the `migrations` table and supports rollback operations.

### 5. GraphQL Gateway
Automatically provisions a GraphQL endpoint.
- Uses `gqlgen` to generate an optional GraphQL schema dynamically representing user-defined data models.
- Available at `/api/v1/graphql`.

### 6. Backup Engine
Supports taking snapshots of data.
- Capable of backing up generic service records and generating JSON data payloads for table state.
- Supports restoring rows via decoding backup payloads.

### 7. Observability
State-of-the-art monitoring tools integrated out-of-the-box.
- **Logging**: Structured Zap logging (`pkg/logger/`) with injected correlation IDs.
- **Tracing**: OpenTelemetry/Jaeger (`internal/tracing/`) wrapping SQL and network requests.
- **Metrics**: Prometheus metrics integrated via Gin middleware tracking HTTP request latency and codes, exposed at `/metrics`.

## Project Structure

A clean, modular layout standard for large Go applications:

```
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── api/             # API Handlers and routing
│   ├── config/          # Environment and configuration
│   ├── database/        # Database Connection Managers and pooling
│   ├── graphql/         # GraphQL Gateway
│   ├── handlers/        # HTTP controllers mapping requests to services
│   ├── middleware/      # Gin middlewares (Auth, Prometheus, etc.)
│   ├── models/          # Domain Entities (Services, Fields, DB Conns, etc.)
│   ├── services/        # Application Business Logic (CRUD, Migration, Service Builder)
│   └── tracing/         # OpenTelemetry configuration
├── pkg/
│   ├── logger/          # Structured Zap logging setup
│   ├── pagination/      # Common pagination helpers
│   ├── response/        # Standard HTTP JSON response helpers
│   └── validation/      # Input validation utilities
├── tests/
│   ├── unit/            # Isolated unit tests
│   ├── integration/     # Integration tests
│   └── uat/             # User Acceptance Testing
├── docker/              # Docker deployment configurations
├── Dockerfile           # Multi-stage container build file
├── Makefile             # Command runner
└── ARCHITECTURE.md      # Detailed system architecture specification
```

## Running the Platform

### Docker Setup

The platform is fully containerized using a multi-stage `Dockerfile` and `docker-compose`.

```bash
# Start all services (Backend, DBs, Redis, Observability)
make docker-up

# View logs
make docker-logs

# Tear down
make docker-down
```

### Local Development

Uses `air` for hot-reloading during development.

```bash
# Install dependencies
make deps

# Run with hot reload
make dev

# Or run natively
make run
```

## Testing

The platform enforces high test coverage across Repository, Service, and API handler layers. Tests are run using an in-memory SQLite database (`:memory:`) to ensure isolation and speed.

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run full tests with coverage output (Coverage target >= 80%)
make test-coverage
```

## Swagger Documentation

API documentation is generated using Swaggo. The docs map all standard endpoints for the CMS control plane.

```bash
# Generate swagger docs
make swagger
```

## Technology Stack Summary
- **Language**: Go 1.24
- **Framework**: Gin
- **ORM**: GORM
- **GraphQL**: gqlgen
- **Database**: PostgreSQL, MySQL, MongoDB, SQLite (testing)
- **Logging**: Zap
- **Metrics**: Prometheus
- **Tracing**: OpenTelemetry
- **Documentation**: Swagger/OpenAPI
- **Containerization**: Docker
