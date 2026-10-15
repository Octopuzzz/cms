# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade, self-hosted, Go-native Backend-as-a-Service (BaaS) platform. It allows developers to dynamically create backend services, schemas, and APIs. The platform is designed using Clean Architecture and Domain-Driven Design (DDD) to be modular, scalable, and cloud-ready.

## System Architecture

The platform is designed around several core engines that handle specific platform responsibilities.

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

### Core Components

*   **CMS Control Plane**: The management interface used to control the system. Responsibilities include managing database connections, creating services, defining schemas, and managing relations, migrations, and backups.
*   **Service Builder**: Allows dynamic creation of services (data models) including fields, relations, constraints, and validation rules.
*   **Database Connection Manager**: Supports registering external databases (PostgreSQL, MySQL, SQLite) and manages connection pooling.

## Metadata Database Schema

The core internal configuration is stored in the Metadata Database. Tables include:

1.  `database_connections`: Stores external DB configurations (id, name, type, host, port, credentials).
2.  `services`: Represents user-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3.  `fields`: Defines attributes for each service (string, integer, float, uuid, JSON). Configures uniqueness, nullability, defaults.
4.  `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks.
5.  `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6.  `backups`: Stores snapshot records or schema outputs.
7.  `users` and `roles`: General authentication and authorization for the control plane.
8.  `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project follows a clean, modular structure:

```text
.
├── cmd/
│   └── server/          # Entry point for the application
├── docker/              # Docker and Prometheus configurations
├── internal/
│   ├── config/          # Application configuration
│   ├── database/        # Database connection and pooling
│   ├── graphql/         # GraphQL gateway implementation
│   ├── handlers/        # API route handlers (Presentation Layer)
│   ├── middleware/      # Gin middleware (Auth, Rate Limiter, etc.)
│   ├── models/          # Domain models (Domain Layer)
│   ├── services/        # Business logic and platform engines (Application Layer)
│   └── tracing/         # OpenTelemetry tracing setup
├── pkg/                 # Shared utilities
│   ├── logger/          # Structured Zap/Logrus logging
│   ├── pagination/      # Pagination helpers
│   ├── response/        # Standardized API responses
│   └── utils/           # Miscellaneous utilities
└── tests/               # Unit, Integration, and UAT tests
```

## CRUD Engine & Query Engine

When a service is created, the system automatically generates CRUD endpoints. The dynamic CRUD engine handles operations using standard HTTP methods.

Example Endpoints:
*   `POST /api/v1/data/{slug}`
*   `GET /api/v1/data/{slug}/{id}`
*   `PUT /api/v1/data/{slug}/{id}`
*   `DELETE /api/v1/data/{slug}/{id}`
*   `GET /api/v1/data/{slug}`

The Query Engine supports advanced filtering, sorting, and pagination dynamically mapping query parameters to database queries (e.g., `?email=test@example.com&sort=created_at:desc&page=1&limit=20`).

## Schema Migration Engine

The platform supports safe schema changes to dynamic services. Managed by the `MigrationService`, it tracks and applies schema diffs (adding, dropping, or renaming columns) using GORM's `Migrator` interface. The engine natively logs migration status into the `migrations` table and supports rollback capabilities.

## GraphQL Gateway (Optional)

The GraphQL gateway automatically generates generic schemas from service definitions using `github.com/99designs/gqlgen/graphql`. The endpoint is exposed at `POST /api/v1/graphql` to fulfill automated GraphQL queries for the generated data models.

## Backup System

The Backup Engine manages data snapshots and database schema backups. It implements generic service record backup features mapping dynamic table contents to snapshots and supports restoring rows via JSON payload decoding. It supports table and schema backups, available through the `POST /api/v1/cms/backups` and `POST /api/v1/cms/backups/{id}/restore` endpoints.

## Observability Integration

The platform integrates state-of-the-art observability tools:

*   **Logging**: Structured logging using Zap/Logrus, injecting correlation and trace IDs directly into context (`pkg/logger/`).
*   **Tracing**: OpenTelemetry/Jaeger integration initialized in `internal/tracing/` to wrap SQL commands and network logic.
*   **Metrics**: Prometheus metrics tracking request latency and status codes via HTTP interceptors (`internal/middleware/prometheus.go`). Exported at `/metrics`.

## Unit Tests

The system includes a comprehensive suite of Unit, Integration, and UAT tests. Tests cover Repository, Service, and API handler layers.

Run tests using the Makefile:

```bash
make test-unit
make test-coverage
```

## Docker Setup

The platform is fully containerized. To spin up the platform along with necessary dependencies:

```bash
make docker-up
# or
docker-compose up -d
```

## Swagger Documentation

API documentation is automatically generated using Swagger / OpenAPI. To generate or update the documentation:

```bash
make swagger
```

This will output the documentation to the `docs/` directory, which is ignored by version control.
