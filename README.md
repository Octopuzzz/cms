# Go Dynamic CMS & API Builder

A self-hosted, Go-native Backend-as-a-Service (BaaS) platform that dynamically generates REST and GraphQL APIs. It allows developers to dynamically create backend services, schemas, and APIs.

## System Architecture Explanation

The platform acts as a backend infrastructure generator, following Clean Architecture and Domain-Driven Design (DDD).

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

### Core Layers
- **Presentation Layer**: The Gin HTTP Router and Handlers mapping requests to internal services, and generating GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic controlling data access, dynamic schemas generation, and operations requested by handlers.
- **Domain Layer**: The core entity models (`Service`, `Field`, `DatabaseConnection`, `User`, `Role`).
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry, metrics (Prometheus), and logging (Zap).

## Metadata Database Schema

The CMS stores platform metadata in a dedicated database (PostgreSQL recommended).

Tables include:
- `database_connections`: Stores external DB configurations (id, name, type, host, port, credentials).
- `services`: Represents user-created data models (id, name, database_id, created_at, updated_at).
- `fields`: Defines attributes for each service (name, type, nullable, unique, default_value, index).
- `relations`: Configures One-to-One, One-to-Many, Many-to-One, and Many-to-Many relations.
- `migrations`: Keeps track of DDL executions with metadata required for rollbacks.
- `backups`: Stores snapshot records or schema outputs.
- `users` / `roles`: General authentication and authorization (RBAC) for the control plane.
- `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure.

```text
.
├── cmd
│   └── server
│       └── main.go
├── internal
│   ├── config
│   ├── database
│   ├── graphql
│   ├── handlers
│   ├── middleware
│   ├── models
│   ├── services
│   └── tracing
├── pkg
│   ├── logger
│   └── response
├── tests
├── docker
└── Makefile
```

## CRUD Engine Implementation

When a service is created via the Service Builder, the Dynamic CRUD Engine automatically generates endpoints for standard operations on the fly mapped to standard GORM database operations.

Example endpoints:
- `POST /api/{service}`
- `GET /api/{service}/{id}`
- `PUT /api/{service}/{id}`
- `DELETE /api/{service}/{id}`
- `GET /api/{service}`

The system supports advanced queries via the **Query Engine**:
- **Filtering**: `GET /api/users?email=john@example.com`
- **Sorting**: `GET /api/users?sort=created_at:desc`
- **Pagination**: `GET /api/users?page=1&limit=20`
- **Join Queries**: `GET /api/orders?join=user,products`

## Schema Migration Engine

The platform handles safe schema changes via the Schema Migration Engine.

- Tracks and applies safe schema diffs (Add column, Drop column, Rename column, Change column type) using GORM's automatic migrations.
- Logs migration status natively into the `migrations` table.
- Safety features include automatic backups before migrations and rollback capabilities.

## GraphQL Gateway

A generic optional GraphQL schema is generated dynamically using `gqlgen`.
The system exposes a `POST /api/v1/graphql` endpoint fulfilling automated GraphQL endpoints for defined data models, supporting queries, mutations, and relations.

Example query:
```graphql
query {
  users {
    id
    name
    email
  }
}
```

## Backup System

The Backup Engine supports multiple backup formats (SQL dump, JSON snapshot) for data preservation.
- Generates data snapshots mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.
- Endpoints: `POST /cms/backup/service/{service_id}` and `POST /cms/restore/service/{service_id}`

## Observability Integration

The system natively supports comprehensive observability:
- **Logging**: Zap structured logging (request logs, error logs, query logs) with correlation and trace IDs.
- **Metrics**: Prometheus tracking request latency, slow queries, and error rates via standard HTTP interceptors.
- **Tracing**: OpenTelemetry (Jaeger) initialized in the infrastructure layer wrapping SQL commands and network logic.

## Unit Tests

The system includes comprehensive unit testing focused on Repository, Service, and API handlers.

- Target Minimum Coverage: 80%
- Framework: standard `testing` library with in-memory SQLite database setup (`:memory:`) via `github.com/glebarez/sqlite` driver for database mocking.
- Run tests: `make test-unit`, or for full coverage `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`.

## Docker Setup

The platform is fully containerized and cloud-ready.

- Uses standard `Dockerfile` for the Go application.
- Uses `docker-compose.yml` to orchestrate the backend, external databases (PostgreSQL, MySQL, MongoDB), Redis (for query caching), and observability tools (Prometheus, Jaeger).

## Swagger Documentation

API Documentation is generated using Swagger/OpenAPI.
- Generated via `swag` CLI.
- Command: `~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`
- Exposed natively within the platform to interact with the CMS Control Plane and dynamically generated REST endpoints.

---

**Built with Go, Gin, GORM, gqlgen, Zap, Prometheus, and OpenTelemetry.**
