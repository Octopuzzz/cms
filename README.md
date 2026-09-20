# CMS Backend Platform

A production-grade, self-hosted Backend-as-a-Service platform (Dynamic CMS + API Builder) written in Go. This system is designed to automatically generate backend services, manage schemas and migrations, and serve dynamic REST and optional GraphQL APIs from external databases, similar to platforms like Hasura or Supabase.

## High-Level System Architecture

The system is built on **Clean Architecture and Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: Exposes both REST API (via Gin) and GraphQL (via gqlgen). Acts as the API gateway mapping external requests to internal systems.
- **Application Layer**: Contains business logic encapsulated within specialized services (e.g., Service Builder, CRUD Engine, Query Engine, Backup Engine, Migration Engine).
- **Domain Layer**: The core metadata data models representing Services, Fields, Relationships, etc.
- **Infrastructure Layer**: Cross-cutting tools including connection caching, telemetry (OpenTelemetry), metrics (Prometheus), logging (Zap), and standard GORM configurations.

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

The core internal configuration is stored in the **Metadata Database**. By default, it manages the configuration for external datasets and APIs. Tables include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project follows standard Go conventions for clean, modular organization:

```text
├── cmd
│   └── server                # Main application entrypoint
├── internal
│   ├── api                   # Extracted route logic
│   ├── config                # Platform configuration via ENV
│   ├── database              # Database Connection Manager & GORM wrappers
│   ├── graphql               # gqlgen code and generated GraphQL Gateway
│   ├── handlers              # Gin HTTP handlers
│   ├── middleware            # Auth, Rate Limiter, Telemetry, etc.
│   ├── models                # Metadata schema domain definitions
│   ├── services              # Application logic (CRUD Engine, Service Builder, Migrations, etc.)
│   └── tracing               # OpenTelemetry initializations
├── pkg
│   ├── errors                # Centralized error handling
│   ├── logger                # Zap structured logger configuration
│   ├── pagination            # Standardized offset/limit pagination
│   ├── response              # Consistent JSON HTTP response writer
│   └── validation            # Shared data validation logic
├── tests                     # Automated tests
│   ├── integration
│   ├── uat
│   └── unit
├── docker                    # Docker ecosystem setup
├── Makefile                  # Build orchestration
└── docs                      # Swagger / OpenAPI outputs
```

## Component Implementations

### CRUD & Query Engine
Managed by `internal/services/dynamic_data_service.go`. The system automatically generates CRUD endpoints.
- Maps endpoints (e.g., `POST /api/v1/data/{slug}`, `GET /api/v1/data/{slug}/{id}`) to standard GORM operations natively.
- Supports advanced dynamic querying via query parameters (e.g. filtering `?email=john@example.com`, sorting `?sort=created_at:desc`, and offset/limit pagination).
- Processes RBAC and applies row-level policies directly on-the-fly.
- Facilitates relation inclusion mappings (One-to-One, One-to-Many, Many-to-Many via auto-join tables).

### Schema Migration Engine
Managed by `internal/services/migration_service.go`.
- Tracks and applies safe schema diffs incrementally (Add, Drop, Rename Column, Modify Type) via GORM’s Migrator interface natively.
- Automatically triggers database structural changes when a Service model is updated via the CMS Control Plane.
- Stores historical tracking for potential rollback auditing.

### GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
- Auto-generates optional queries and mutations matching dynamic service definitions via `github.com/99designs/gqlgen`.
- Allows unified nested relation fetching on top of the CRUD Engine underlying connections.

### Backup System
Managed by `internal/services/backup_service.go`.
- Enables manual and automated snapshot captures.
- Capable of backing up data structurally (e.g. JSON snapshots) for individual services.
- Facilitates point-in-time state recovery or automated backup routines prior to executing schema migrations.

### Observability Integration
End-to-End observability is enforced:
- **Logging:** Structured JSON logs produced by `pkg/logger/` utilizing Uber’s Zap framework. Emits context-aware trace IDs.
- **Metrics:** `internal/middleware/prometheus.go` tracks API request latencies, status codes, and error rates natively.
- **Tracing:** `internal/tracing/` sets up OpenTelemetry (OTel), capturing network spans, and query execution delays across the API Gateway.

## Unit Testing

Strict minimum coverage of 80% is targeted across Repository, Service, and API handlers. Tests utilize in-memory SQLite instances for high execution speeds.

```bash
# Run full unit tests
make test-unit

# Run full coverage analysis across all packages
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Docker Setup

Containerization supports scaling and simplified deployments:
- Contains an optimized multi-stage `Dockerfile`.
- Uses `docker-compose.yml` to orchestrate dependent infrastructure like PostgreSQL (metadata database), Redis, and Observability tools alongside the Go application platform.

```bash
# Bring up local platform environment
make docker-up
```

## Swagger Documentation

Auto-generated OpenAPI (Swagger) interface provides documentation for dynamic models and standard CMS endpoints:
- Use `swag` annotations on handlers natively.
- Run `make swagger` to generate definitions matching live logic.
