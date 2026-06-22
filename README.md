# Backend Platform (Dynamic CMS + API Builder)

A production-grade Backend-as-a-Service (BaaS) platform written in Go. This system operates similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native. It empowers developers to dynamically create backend services, schema definitions, APIs (REST & GraphQL), and manage databases, migrations, and backups.

## System Architecture

The platform follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles, ensuring a modular, scalable, and cloud-ready system.

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

## Metadata Database Schema

The CMS stores platform metadata in a primary database (PostgreSQL recommended). Key tables include:

1. `database_connections`: External DB configurations (id, name, type, host, port, credentials).
2. `services`: User-created data models referencing database connections.
3. `fields`: Attributes for each service (name, type, nullable, unique, default_value).
4. `service_permissions`: Connects roles to services for fine-grained access control.
5. `migrations`: Tracks DDL executions and metadata for rollbacks.
6. `backups`: Stores snapshot records and schema outputs.
7. `users` & `roles`: General authentication/authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure:

```text
.
├── cmd/
│   └── server/          # Application entry point
├── docker/              # Docker and containerization files
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection manager
│   ├── graphql/         # GraphQL gateway and schemas
│   ├── handlers/        # HTTP handlers (Presentation Layer)
│   ├── middleware/      # Auth, Rate Limiter, Prometheus
│   ├── models/          # Data models (Domain Layer)
│   ├── services/        # Business logic (Application Layer)
│   └── tracing/         # OpenTelemetry integration
├── pkg/
│   ├── logger/          # Zap structured logging
│   └── response/        # Standardized API responses
└── tests/               # Unit, integration, and UAT tests
```

## CMS Control Plane & Service Builder

Managed under `/api/v1/cms/` routes, providing an interface to register external databases and dynamically create services (data models) including fields, relations, constraints, and validation rules. Supported databases include PostgreSQL, MySQL, and MongoDB.

## Dynamic CRUD Engine & Query Engine

Automatically generates CRUD operations for dynamically created services.
- **CRUD Operations**: Handled via endpoints like `GET /api/v1/data/{slug}`.
- **Querying**: Supports advanced querying including filtering (`?email=test@test.com`), sorting, pagination, and dynamic joining.

## Schema Migration Engine

Safely applies schema changes (add, drop, rename columns) utilizing GORM's `Migrator`. It natively logs migration status into the `migrations` table and supports automatic backups and rollbacks.

## GraphQL Gateway

An optional API layer exposing auto-generated GraphQL queries and mutations from service schemas. Built using `gqlgen` and accessible at `POST /api/v1/graphql`.

## Backup Engine

Generates data snapshots and supports restoring rows via JSON payloads. Future expansions include full table and schema backups as well as SQL dumps.

## Observability Integration

- **Logging**: Structured request/error/query logging via Zap (`pkg/logger/`).
- **Metrics**: Request latency, error rates, and slow queries exported via Prometheus at `/metrics`.
- **Tracing**: OpenTelemetry/Jaeger wrapping SQL commands and network logic in `internal/tracing/`.

## Unit Testing

Run unit, integration, and UAT tests:

```bash
make test-unit
make test-integration
make test-uat
```

Or run all tests with coverage (minimum 80% expected):

```bash
make test-coverage
```

## Docker Setup

The platform is containerized and cloud-ready.

```bash
make docker-up
```

Start the services detached using `docker-compose`.

## Swagger Documentation

Generate OpenAPI/Swagger documentation natively:

```bash
make swagger
```

Outputs documentation to `docs/` providing standard interactive API specs.