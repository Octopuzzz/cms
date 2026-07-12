# Dynamic CMS & API Builder

A self-hosted, Go-native Backend-as-a-Service (BaaS) platform that allows developers to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

This platform acts as a backend infrastructure generator, working similarly to Hasura or Supabase.

## System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: The Gin HTTP Router and Handlers. Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic. Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models. Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry, metrics (Prometheus), and logging.

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
PostgreSQL    MySQL       MongoDB
```

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. Tables defined in the models include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean, modular structure:

```
├── cmd
│   └── server             # Main application entrypoint
├── internal
│   ├── api                # API routes and controllers
│   ├── config             # Configuration management
│   ├── database           # Database connection manager
│   ├── graphql            # GraphQL gateway implementation
│   ├── handlers           # HTTP handlers
│   ├── middleware         # Middleware (Auth, Telemetry, etc.)
│   ├── models             # Domain models (Metadata DB schema)
│   ├── services           # Application business logic
│   └── tracing            # OpenTelemetry and Tracing
├── pkg
│   ├── logger             # Zap structured logging
│   ├── pagination         # Pagination utilities
│   ├── response           # Standardized API responses
│   └── validation         # Input validation logic
├── tests                  # Unit, integration, and UAT tests
├── docker                 # Docker configurations
├── Makefile               # Standardized tasks
├── docker-compose.yml     # Local services (DB, Redis, Observability)
└── ARCHITECTURE.md        # Architectural decisions
```

## Core Components Implementation

### CRUD Engine & Query Engine

Managed by dynamic data services.
- Maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly.
- Applies automated filtering via query params (e.g. `?email=test@test.com`), dynamic joining via foreign key mappings, and pagination logic.
- Verifies Role-Based Access Control and Row-Level filtering dynamically prior to queries.

### Schema Migration Engine

- Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM's Migrator.
- Logs migration status natively into `migrations` table and handles rollbacks through snapshot retention logic.

### GraphQL Gateway

- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

### Backup System

- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.

### Observability Integration

Implemented across the stack:
- **Zap Logging**: Injected globally with correlation and trace IDs tied directly into Context.
- **OpenTelemetry/Jaeger**: Initialized to wrap SQL commands and network logic.
- **Prometheus Metrics**: Tracks latency and status codes via standard HTTP interceptors. Exported at `/metrics`.

## Unit Tests

The system includes comprehensive tests with a minimum coverage requirement of 80%.
Tests focus on Repository, Service, and API handlers.
Run unit tests with:
```bash
make test-unit
# or
go test -v -race ./tests/unit/...
```

To run all tests and generate a coverage report:
```bash
make test-coverage
```

## Docker Setup

The platform is containerized and includes a `docker-compose.yml` for local development.

```bash
# Start Docker services
make docker-up

# Stop Docker services
make docker-down

# View logs
make docker-logs
```

## Swagger Documentation

Swagger documentation is automatically generated from source code annotations.
To generate the latest Swagger docs:

```bash
make swagger
```

This will run `swag init` and output to the `docs/` folder, which can be viewed via the Swagger UI endpoint when running the server.
