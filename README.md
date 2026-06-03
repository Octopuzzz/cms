# CMS Backend Platform

A production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. This self-hosted, Go-native Backend-as-a-Service allows developers to connect databases, dynamically create data models, automatically generate CRUD and GraphQL APIs, manage schema migrations, monitor observability, manage backups, and easily scale.

---

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

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

## 2. Metadata Database Schema

The CMS stores platform metadata in the Metadata Database (default: PostgreSQL, MySQL, SQLite supported).
Tables include:

- **databases**: Stores external database configurations (id, name, type, host, port, username, password, database_name).
- **services**: Represents user-created data models, references a database connection, tracks tables. (id, name, database_connection_id, created_at, updated_at)
- **fields**: Defines attributes for each service (name, type, nullable, unique, default_value, index).
- **service_permissions**: Connects `Role` to `Service` for RBAC.
- **migrations**: Tracks DDL executions for rollback/history.
- **backups**: Stores snapshot records.
- **users / roles**: Authentication and authorization.

## 3. Go Project Structure

The project uses a clean modular structure:
```text
cmd/
 └── server/        # Main application entry point
internal/
 ├── api/           # Handlers and routes
 ├── config/        # Environment configurations
 ├── database/      # Database connections and pool manager
 ├── graphql/       # GraphQL Gateway (gqlgen)
 ├── handlers/      # Gin HTTP controllers
 ├── middleware/    # Auth, logging, metrics interceptors
 ├── models/        # Domain entities
 ├── services/      # Business logic (Service Builder, Migration, etc.)
 └── tracing/       # OpenTelemetry setup
pkg/
 ├── logger/        # Zap structured logging
 ├── pagination/    # Pagination logic
 └── response/      # Standardized HTTP responses
tests/
 ├── unit/          # Unit tests
 ├── integration/   # Integration tests
 └── uat/           # User Acceptance Testing
docker/             # Docker configuration
```

## 4. CRUD Engine Implementation

The **Dynamic CRUD Engine** (managed by `internal/services/dynamic_data_service.go`) maps standard HTTP methods to auto-generated REST endpoints:
- **POST** `/api/v1/data/{service}` - Create
- **GET** `/api/v1/data/{service}/{id}` - Read
- **PUT** `/api/v1/data/{service}/{id}` - Update
- **DELETE** `/api/v1/data/{service}/{id}` - Delete
- **GET** `/api/v1/data/{service}` - List (supports pagination `?page=1&limit=20`, sorting `?sort=created_at:desc`, filtering `?email=john@example.com`, and joining).

## 5. Schema Migration Engine

The **Schema Migration Engine** (`internal/services/migration_service.go`) manages safe schema changes (Add/Drop/Rename Column, Change type) via GORM's `.Migrator()`.
It supports safe tracking, automatic backup prior to migrating, and rollback capabilities natively logged into the `migrations` metadata table.

## 6. GraphQL Gateway

The **GraphQL Gateway** (`internal/graphql/gateway.go`) automatically generates GraphQL schemas from your service definitions using `gqlgen`.
It exposes `POST /api/v1/graphql` for automated query and mutation fulfillment over defined data models, and provides a playground at `GET /api/v1/graphql/playground`.

## 7. Backup System

The **Backup Engine** (`internal/services/backup_service.go`) manages data snapshots.
It generates service record backups (JSON snapshots) and supports mapping dynamic table contents. Backups and restoration processes can be handled dynamically through the platform.

## 8. Observability Integration

The platform provides comprehensive observability out of the box:
- **Logging**: Zap structured logging with correlation and trace IDs. Request logs, error logs, query logs.
- **Metrics**: Prometheus metrics exported via `/metrics` capturing request latency, error rates, etc.
- **Tracing**: OpenTelemetry (integrated with Jaeger) wraps SQL commands and network requests for deep tracing.

## 9. Unit Tests

The system maintains high code quality with comprehensive testing spanning Repository, Service, and API handlers.
Minimum coverage is 80%. Tests run against an in-memory SQLite database.
Run all tests using:
```sh
go test ./...
```
Or via make:
```sh
make test-unit
make test-integration
make test-uat
make test-coverage
```

## 10. Docker Setup

A complete `Dockerfile` and `docker-compose.yml` are provided for rapid deployment.
To spin up the platform (along with optionally required backing services like Redis or DBs):
```sh
docker-compose up -d
```
The Docker setup ensures a containerized, cloud-ready execution environment.

## 11. Swagger Documentation

The platform auto-generates API documentation using Swagger / OpenAPI.
To generate/update the documentation:
```sh
make swagger
```
The generated documentation is hosted on standard Swagger UI endpoints within the Gin server, detailing payload formats, authentication (JWT), and parameters.
