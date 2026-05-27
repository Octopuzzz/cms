# Dynamic CMS & API Builder

A production-grade, self-hosted Backend-as-a-Service platform written in Go. This platform allows developers to dynamically create backend services, schemas, and APIs (both REST and GraphQL), working similarly to Hasura or Supabase but in a fully self-hosted Go-native environment.

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles, ensuring a modular, scalable, and maintainable platform.

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

### Core Layers:
- **Presentation Layer:** Gin HTTP Router and Handlers mapping requests to internal services, including a GraphQL Gateway using `gqlgen`.
- **Application Layer:** Business logic housed in services controlling data access, dynamic schemas generation, and execution of operations.
- **Domain Layer:** Core data models defining the internal representations (e.g., `Service`, `Field`, `DatabaseConnection`, `User`, `Role`).
- **Infrastructure Layer:** Cross-cutting tools such as caching, OpenTelemetry tracing, Prometheus metrics, and Zap structured logging.

## 2. Metadata Database Schema

The core configuration and platform metadata is managed within a central database. The main entities include:

- **`database_connections`**: Stores configuration for external DBs (id, name, type, host, port, credentials). Supports pooling.
- **`services`**: Defines user-created data models. Contains references to `database_connection_id` and the dynamic schema (`db_table_name`).
- **`fields`**: Defines attributes for each service (type: string, integer, float, boolean, uuid, json, array, timestamp, etc.). Configures nullability, uniqueness, and defaults.
- **`relations`**: Configures One-to-One, One-to-Many, Many-to-One, and Many-to-Many relationships.
- **`migrations`**: Keeps track of applied schema changes and DDL executions.
- **`backups`**: Stores snapshots and backup records.

## 3. Go Project Structure

The project uses a clean modular structure to separate concerns effectively:

```
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── api/             # API routing configurations
│   ├── config/          # Environment and application configurations
│   ├── database/        # Database Connection Manager & standard connections
│   ├── graphql/         # GraphQL Gateway generation and resolving
│   ├── handlers/        # HTTP controllers (REST)
│   ├── middleware/      # Application middlewares (Auth, Logging, Metrics)
│   ├── models/          # Domain layer and Database schema definitions
│   ├── services/        # Business logic (CRUD Engine, Service Builder)
│   └── tracing/         # OpenTelemetry setup
├── pkg/                 # Shared generic libraries (e.g., logger)
├── tests/               # Unit, integration, and UAT test suites
├── docker/              # Docker and container resources
├── Makefile             # Standardized task commands
└── docker-compose.yml   # Local development composition
```

## 4. CRUD Engine Implementation

When a new service is configured via the **Service Builder**, the **Dynamic CRUD Engine** immediately provides standard endpoints:

- `POST /api/v1/data/{slug}` - Create
- `GET /api/v1/data/{slug}/{id}` - Read
- `PUT /api/v1/data/{slug}/{id}` - Update
- `DELETE /api/v1/data/{slug}/{id}` - Delete
- `GET /api/v1/data/{slug}` - List

The **Query Engine** works alongside this to map GET parameters directly into GORM operations, supporting:
- Filtering (e.g., `?email=john@example.com`)
- Sorting (e.g., `?sort=created_at:desc`)
- Pagination (e.g., `?page=1&limit=20`)

## 5. Schema Migration Engine

Safe structural modifications are handled by the **Schema Migration Engine**.
- Allows modifying existing schemas (Add/Drop/Rename columns) without losing data.
- Tracks migration statuses using GORM's `Migrator` interface.
- Automates metadata recording in the `migrations` table, allowing administrators to audit or rollback recent changes dynamically.

## 6. GraphQL Gateway

An automated GraphQL endpoint is exposed at `POST /api/v1/graphql`, generated directly from the configured service definitions.
- Utilizes `github.com/99designs/gqlgen` internally.
- Features queries and mutations that reflect the active dynamic schema.
- Resolves data correctly matching the REST functionality, ensuring data parity across endpoints.

## 7. Backup System

A robust tool to snapshot and recover data:
- Exposes administrative routes such as `POST /api/v1/cms/backup/service/{service_id}` to take backups of specific services.
- Data can be snapshotted efficiently in structured formats like JSON.
- Provides rollback/restore capabilities via `POST /api/v1/cms/restore/{id}` to ensure data safety.

## 8. Observability Integration

Comprehensive tracing, logging, and metrics are built into the platform:
- **Logging**: Structured JSON logging powered by `go.uber.org/zap`. Automatically correlates HTTP requests via Request IDs.
- **Metrics**: Standardized Promethues metrics exported at `/metrics`. Tracks request latencies, error rates, and hit counts via custom middleware.
- **Tracing**: OpenTelemetry (OTel) instrumentation seamlessly integrated into HTTP middleware and database execution traces. Supports visual tracing with Jaeger.

## 9. Unit Tests

The system maintains a high standard of reliability with extensive unit testing across the Repository, Service, and API handler layers, targeting a minimum of 80% coverage.

**Running Tests:**
Use the provided Makefile to run tests:
```bash
make test-unit
make test-integration
make test-coverage
```

Or manually via Go:
```bash
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## 10. Docker Setup

The platform is container-ready out of the box.

Start the entire stack, including databases and observability services, using Docker Compose:
```bash
docker-compose up -d
```

The standard `Dockerfile` produces a lightweight, optimized Go application binary, ensuring quick startup and minimal resource usage.

## 11. Swagger Documentation

Automatic REST API documentation is provided using Swagger/OpenAPI.

To regenerate documentation:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

You can view the documentation natively by visiting:
`http://localhost:8080/swagger/index.html` (or the configured host address).
