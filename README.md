# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade, self-hosted Backend-as-a-Service (BaaS) platform written in Go. It works similarly to platforms like Hasura or Supabase, allowing developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

The system is fully self-hosted, modular, scalable, and cloud-ready. It follows Clean Architecture and Domain-Driven Design (DDD) principles.

## 1. System Architecture Explanation

The platform connects to various databases and generates an API Gateway that routes requests through dynamically constructed schemas. It uses the Gin web framework for the Presentation layer and GORM for database interactions.

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

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## 2. Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. Tables defined in the Domain Layer include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## 3. Go Project Structure

The project uses a clean modular structure.

```text
├── cmd
│   └── server             # Main application entrypoint
├── internal
│   ├── config             # Configuration and environment variable management
│   ├── database           # Database connection manager and pooling
│   ├── graphql            # Auto-generated GraphQL Gateway logic
│   ├── handlers           # Gin REST API endpoints mapping
│   ├── middleware         # Authentication, rate-limiting, and observability middlewares
│   ├── models             # Domain definitions and Metadata Schema definitions
│   ├── services           # Business logic (CRUD, Service Builder, Migration, Backups)
│   └── tracing            # OpenTelemetry setups
├── pkg
│   ├── errors             # Standardized error formats
│   ├── logger             # Zap structured logger implementation
│   ├── pagination         # Dynamic pagination logic
│   ├── response           # Standard HTTP response helpers
│   └── validation         # Input validation
├── tests
│   ├── unit               # Unit testing layer (Repositories, Services, Handlers)
│   ├── integration        # Integration tests
│   └── uat                # End-to-end / User Acceptance Tests
├── docker                 # Docker configuration files
├── Dockerfile             # Core application container
├── docker-compose.yml     # Stack orchestrator mapping Postgre/Redis/Jaeger/Prometheus
└── Makefile               # CLI helper for testing, building, and running
```

## 4. CRUD Engine Implementation

Managed by `internal/services/dynamic_data_service.go` and associated handlers.

When a service is created, the system automatically exposes REST CRUD endpoints under dynamic routes like `GET /api/v1/data/{slug}`.
- It maps these requests to standard GORM database operations on the fly using reflection or dynamic struct generation.
- The Engine automatically parses filtering via query params (e.g. `?email=test@example.com`), sorting (`?sort=created_at:desc`), and advanced query constructs.
- It verifies Role-Based Access Control and Row-Level security dynamically prior to executing queries.

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`.

The system tracks and applies safe schema diffs (Add Column, Drop Column, Rename Column, Type modifications) dynamically relying on GORM's `.Migrator()`.
It logs migration statuses natively into the `migrations` table and handles fail-safes such as rollback capability via snapshot retentions.

## 6. GraphQL Gateway

Managed by `internal/graphql/gateway.go`.

An optional GraphQL layer is automatically generated from the dynamic service schemas. Using `github.com/99designs/gqlgen`, the gateway fulfills nested queries and mutations, providing an alternative to the REST interface. It listens on `POST /api/v1/graphql`.

## 7. Backup System

Managed by `internal/services/backup_service.go`.

The platform features built-in table and service snapshot capabilities. It exports dynamic table states into JSON payloads and persists them into the `backups` table or as files. The restore process ingests these JSON payloads to rebuild row states for specific services.

## 8. Observability Integration

The system implements state-of-the-art telemetry and observability:
- **Logging**: Zap structured logging is injected globally with correlation and trace IDs tied directly into Context.
- **Metrics**: Exposes Prometheus metrics via `internal/middleware/prometheus.go` (e.g., request latency, error rate, status codes) exported at the `/metrics` endpoint.
- **Tracing**: OpenTelemetry and Jaeger integrate closely with database handlers and network layers, wrapping SQL commands for distributed tracing visibility.

## 9. Unit Tests

The system maintains comprehensive unit testing coverage across Repository, Service, and API handler layers.

- Unit tests use isolated in-memory SQLite database setups via `github.com/glebarez/sqlite` to emulate PostgreSQL/MySQL behaviors quickly.
- Standard Golang tools (`go test`) are mapped through the `Makefile` providing distinct profiles for unit, integration, and user-acceptance testing.
- Target coverage is >80%.
- To run the full test suite and calculate coverage:
  ```bash
  go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
  ```

## 10. Docker Setup

The system provides fully containerized environments using `Dockerfile` and `docker-compose.yml`.

The compose stack spins up:
- The Go Backend Container (`cms-backend`)
- A PostgreSQL Database (`postgres:15-alpine`) for the Metadata repository
- A Redis instance (`redis:7-alpine`) for caching layers
- A Jaeger Container (`jaegertracing/all-in-one`) for OpenTelemetry consumption
- A Prometheus instance for metrics scraping

Start the environment with:
```bash
docker-compose up -d
```

## 11. Swagger Documentation

API Documentation is auto-generated using standard Swaggo annotations placed on handler routines.
The Swagger documentation covers the static CMS Control plane routes (e.g., `/cms/databases`, `/cms/services`) and structures the dynamic expectations.

To generate or update the documentation:
```bash
make swagger
```
