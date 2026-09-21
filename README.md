# Go Backend Platform (Dynamic CMS + API Builder)

## Platform Goal
A self-hosted, Go-native Backend-as-a-Service platform that acts as a backend infrastructure generator. It empowers developers to connect databases, create data models dynamically, generate CRUD APIs automatically, expose optional GraphQL APIs, manage schema migrations, monitor systems with state-of-the-art observability, and manage backups.

## System Architecture Explanation
The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: The Gin HTTP Router and Handlers mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic. Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models defining core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry, metrics (Prometheus), and logging (Zap).

### High Level Architecture
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

## Metadata Database Schema
The core internal configuration is stored in the **Metadata Database**.

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks. Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure
The project uses a clean modular structure.

```text
.
├── ARCHITECTURE.md     # Architectural source of truth
├── Dockerfile          # Root Docker image configuration
├── Makefile            # Standarized make tasks for building and testing
├── cmd
│   └── server          # Main entrypoint of the CMS Backend
├── docker
│   ├── Dockerfile
│   ├── docker-compose.yml # Containerized services for backing tech stack
│   └── prometheus.yml  # Observability tooling
├── internal
│   ├── config          # Application configuration
│   ├── database        # Database Connection Manager
│   ├── graphql         # Auto-generated GraphQL Gateway logic
│   ├── handlers        # Gin presentation layer HTTP handlers
│   ├── middleware      # Gin middleware (Auth, Prometheus, Rate Limiter)
│   ├── models          # Domain models
│   ├── services        # Application layer business logic
│   └── tracing         # OpenTelemetry / Jaeger observability configurations
├── pkg
│   ├── logger          # Global Zap structured logging
│   └── response        # Standardized HTTP JSON responses
└── tests
    ├── integration     # Multi-component integration tests
    ├── uat             # User Acceptance Testing
    └── unit            # Component-level unit tests
```

## CRUD Engine Implementation
Handled by `DynamicDataService` (`internal/services/dynamic_data_service.go`). It maps generic HTTP operations on dynamically created `services` to their respective external or internal database engine schemas via GORM on the fly. It also applies role-based access control and row-level filtering dynamically prior to queries.

## Query Engine
Also housed inside `DynamicDataService`, the Query Engine is responsible for automated filtering, sorting, nested dynamic joining, and pagination logic applied dynamically through API parameters (e.g. `?email=test@test.com`, `?sort=created_at:desc`, `?page=1&limit=20`, `?join=roles`).

## Schema Migration Engine
Handled by `MigrationService` (`internal/services/migration_service.go`). Tracks and applies safe schema diffs (Add, Drop, Rename Column, etc) directly using GORM's `Migrator()`. Changes are systematically logged natively into the `migrations` metadata table supporting historical tracking and retention logic.

## GraphQL Gateway
Located at `internal/graphql/gateway.go`. The system leverages `github.com/99designs/gqlgen` to dynamically construct generic schemas directly mapped to the generated backend models exposing `POST /api/v1/graphql` endpoints, efficiently supporting unified queries, relations, and mutations without static code generation steps at runtime.

## Backup System
Located at `internal/services/backup_service.go`. The platform's Backup Engine handles exporting table contents or full-schema JSON payload structures mapping dynamic table records into generated snapshots. It also natively supports restoring service rows via REST HTTP payload mechanisms.

## Observability Integration
Implemented systematically across all systems:
- **Zap Logging**: `pkg/logger/` initializes global structured logging tied directly to correlation and trace IDs globally injected in request Contexts.
- **OpenTelemetry/Jaeger**: Wrapped around internal database queries and network activity initialization to allow fine-grained structural bottleneck identification.
- **Prometheus Metrics**: Defined natively inside `internal/middleware/prometheus.go` exporting critical telemetry variables such as latency tracking and HTTP status rates at the `/metrics` endpoint.

## Security and Authentication
The platform includes natively implemented Input validation, automatic SQL Injection protection, strong JWT based API Token and Refresh token based authentication logic, and natively configured Rate-Limiting mechanisms supporting enterprise readiness.

## Unit Tests
Minimum coverage requirements map out to 80% with primary focuses across internal Repositories, Application level Services, and presentation Handlers. `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...` executes these suites leveraging `github.com/glebarez/sqlite` mapped directly to `:memory:` instances for isolated and scalable mocks.

## Docker Setup
A complete Docker setup has been provided for simplified orchestration:
- A root `Dockerfile` defining self-contained artifact generation.
- `/docker/docker-compose.yml` supporting dependent infrastructural dependencies (e.g., PostgreSQL, Prometheus, Jaeger).

## Swagger Documentation
Complete OpenAPI (Swagger) annotations define dynamic and static endpoints inside `cmd/server/main.go` presentation handlers enabling seamless interactive documentation natively accessible in development environments. (Generated via `swag init`).
