# Backend Platform (Dynamic CMS + API Builder)

A fully self-hosted, Go-native Backend-as-a-Service (BaaS) platform designed for dynamically generating REST and GraphQL APIs, managing backend infrastructure, and handling dynamic schema modifications. This platform operates similarly to Hasura or Supabase but is entirely Go-native.

## System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles to maintain high modularity, scalability, and cloud-readiness.

- **Presentation Layer**: The Gin HTTP Router and Handlers act as an API Gateway routing incoming client requests to business logic. It handles both traditional REST requests and GraphQL endpoint queries via `gqlgen`.
- **Application Layer**: Contains business logic (`internal/services/`) responsible for access control, dynamic schema generation, database connectivity handling, and coordinating CRUD operations requested by handlers.
- **Domain Layer**: Houses data models (`internal/models/`) which form the core entities of the platform configuration, such as `Service`, `Field`, `DatabaseConnection`, `User`, and `Role`.
- **Infrastructure Layer**: Manages external integrations, cross-cutting telemetry (OpenTelemetry), metrics (Prometheus), structured logging (Zap), and underlying persistent storage mapping.

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

The CMS Control Plane stores all configuration details about connected databases, dynamically built services, fields, and access policies in the Metadata Database (typically PostgreSQL).

Core tables include:

1. `database_connections`: Stores configurations for registered external DB connections (ID, Name, Type, Host, Port, Credentials) with connection pooling features.
2. `services`: Represents user-created data models, linking to `database_connections` and capturing dynamically created schema tables.
3. `fields`: Defines attributes belonging to a service (e.g., string, integer, float, json). Includes configurations for nullability, default values, indices, and uniqueness.
4. `service_permissions`: Maps user roles to generated services to enforce granular access restrictions (create, read, update, delete).
5. `migrations`: Tracks Schema Migration Engine changes (DDL statements) to support rollback safety.
6. `backups`: Stores metadata and data logs for platform backups and snapshots.
7. `users` and `roles`: General authentication, authorization, and role management for access to the CMS Control Plane.
8. `audit_logs`: Detailed logging of structural, configuration, and data-level modifications for security tracking.

## Go Project Structure

The project implements a clean modular structure supporting microservice-like independent layers:

```
.
├── cmd
│   └── server                  # Entry point for the backend platform
├── docker                      # Docker configurations (e.g., Prometheus)
├── internal
│   ├── config                  # Configuration loading and environment parsing
│   ├── database                # Connection manager and generic DB interfaces
│   ├── graphql                 # GraphQL Gateway and generic schema generation
│   ├── handlers                # HTTP Presentation Layer (REST and GraphQL endpoints)
│   ├── middleware              # Application middleware (Auth, Metrics, Tracing, Rate Limiter)
│   ├── models                  # Domain models defining the metadata schema structure
│   ├── services                # Core logic engines (ServiceBuilder, CRUDEngine, MigrationEngine)
│   └── tracing                 # OpenTelemetry initializers and observability layers
├── pkg
│   ├── logger                  # Standardized Zap structured logger
│   └── response                # HTTP response mapping utilities
├── tests
│   ├── integration             # Integration tests for connected components
│   ├── uat                     # User Acceptance Tests bypassing normal flow paths
│   └── unit                    # Unit tests isolated by Mock DBs (SQLite Memory)
├── .env.example
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── Makefile                    # Standardized build, test, run, and dev tasks
```

## Engine Implementations

### CRUD Engine & Query Engine
Managed via `DynamicDataService` mapping routes like `GET /api/v1/data/{slug}` to dynamic GORM database operations on connected databases. It handles API filtering (e.g., `?email=john@example.com`), sorting (`?sort=created_at:desc`), deep nested foreign key joining, and standardized pagination limits seamlessly.

### Schema Migration Engine
Managed via `MigrationService`. Uses native GORM `.Migrator()` implementations to track, append, and safely transition database structural changes based on modifications defined in the `Service Builder`. Includes safety mechanisms storing operations in the internal `migrations` metadata table.

### GraphQL Gateway
Automatically converts dynamically created REST endpoints into GraphQL structures utilizing `github.com/99designs/gqlgen/graphql`. It maps incoming `queries`, `mutations`, and deeply-nested `relations` directly to existing GORM DB models securely.

### Backup System
Implements database, schema, and structural snapshots via `BackupService`. Capable of executing system-level exports to JSON/SQL formatting for rapid deployment and restoration via `POST /api/v1/cms/backup/service/{service_id}`.

### Observability Integration
The platform incorporates deep visibility across all logical layers:
- **Logging**: Zap handles deeply correlated request/response structures.
- **Metrics**: Prometheus intercepts HTTP traffic evaluating latency, slow queries, and standard load indicators, exported natively over `/metrics`.
- **Tracing**: OpenTelemetry (integrated with Jaeger) wraps critical business logic blocks and nested SQL execution loops.

## Deployment Setup

### Unit Tests
The codebase supports >80% code coverage. Testing includes isolated database mocking via in-memory SQLite layers. Run using:
```bash
make test-unit
# or standard go tooling
go test -v -race ./tests/unit/...
```

### Docker Setup
The project contains an Alpine-based robust `Dockerfile` executing standard Go mult-stage builds reducing container footprints. Launch the full stack using `docker-compose.yml`:
```bash
make docker-up
# Equivalent: docker-compose up -d
```

### Swagger API Documentation
REST definitions are continuously synchronized. Generate Swagger Docs easily using:
```bash
make swagger
```
