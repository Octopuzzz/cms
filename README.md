# Dynamic CMS & API Builder

A self-hosted, Go-native Backend-as-a-Service (BaaS) platform that allows developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

This platform acts as a backend infrastructure generator, similar to Hasura or Supabase, but fully self-hosted and written in Go.

## System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

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

- **Presentation Layer**: Gin HTTP Router and Handlers. Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (Services). Controls data access, generates dynamic schemas, and performs operations requested by handlers.
- **Domain Layer**: Data models defining core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, OpenTelemetry tracing, Prometheus metrics, and Zap logging.

## Metadata Database Schema

The platform metadata is stored in a relational database (default: SQLite for local/testing, PostgreSQL recommended for production). The core tables include:

- `database_connections`: External DB configurations (id, name, type, host, port, credentials).
- `services`: User-created data models, linking to a `database_connection_id` and tracking dynamic schemas.
- `fields`: Attributes for each service (type, uniqueness, nullability, defaults).
- `service_permissions`: Connects `Role` to `Service` for RBAC.
- `migrations`: Tracks DDL executions and metadata.
- `backups`: Stores snapshot records or schema outputs.
- `users` and `roles`: Platform authentication and authorization.
- `audit_logs`: Detailed logging of structural and data modifications.

## Go Project Structure

The project follows a clean, modular structure:

```
├── cmd
│   └── server          # Application entrypoint
├── internal
│   ├── config          # Configuration loading
│   ├── database        # Database connection management
│   ├── graphql         # GraphQL gateway setup
│   ├── handlers        # HTTP/REST API handlers (Presentation Layer)
│   ├── middleware      # Gin middlewares (Auth, Rate Limiting, Observability)
│   ├── models          # Domain models
│   ├── services        # Business logic (Application Layer)
│   └── tracing         # OpenTelemetry setup
├── pkg
│   ├── logger          # Zap structured logging
│   └── response        # Standardized API responses
├── tests
│   ├── integration     # Integration tests
│   ├── uat             # User Acceptance Tests
│   └── unit            # Unit tests
├── docker              # Docker and containerization configs
├── Makefile            # Build and test commands
└── docs                # Generated Swagger documentation (ignored in git)
```

## Core Components Implementation

### CRUD Engine & Query Engine

When a service is created, the platform automatically maps endpoints (`GET /api/v1/data/{slug}`, etc.) to standard GORM operations on the fly.
The Query Engine applies automated filtering (e.g., `?email=test@example.com`), dynamic joining via foreign key mappings, and pagination logic. It natively integrates RBAC and row-level filtering before execution.

### Schema Migration Engine

The Schema Migration Engine tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM’s Migrator. It logs migration status natively into the `migrations` table and supports basic rollbacks through snapshot retention logic.

### GraphQL Gateway

The platform automatically exposes GraphQL endpoints via `POST /api/v1/graphql`. It utilizes `github.com/99designs/gqlgen` to generate generic, optional schemas based on the user-defined data models.

### Backup System

The Backup Engine generates data snapshots. It implements generic service record backup features mapping dynamic table contents to snapshots, and supports restoring rows via JSON payload decoding.

### Observability Integration

- **Logging**: Zap structured logging is injected globally, with correlation and trace IDs tied directly to context.
- **Tracing**: OpenTelemetry (Jaeger compatible) wraps SQL commands and network logic.
- **Metrics**: Prometheus metrics track latency and status codes via standard HTTP interceptors, exported at `/metrics`.

## Unit Tests

The system is tested comprehensively with a target of 80% coverage across the Repository, Service, and API handler layers.

To run the unit tests and check coverage:

```bash
make test-unit
# or for full coverage including all packages:
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Docker Setup

The platform is fully containerized. Use the provided Dockerfile and docker-compose.yml to spin up the application along with required infrastructure (e.g., databases, Redis, Jaeger, Prometheus).

```bash
docker-compose up -d
```

## Swagger Documentation

API Documentation is automatically generated using `swaggo`.

To generate the documentation:

```bash
~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
