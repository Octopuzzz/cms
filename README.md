# Dynamic CMS & API Builder

A self-hosted, Go-native Backend-as-a-Service (BaaS) platform that dynamically generates REST and GraphQL APIs.

## System Architecture

This platform follows **Clean Architecture** and **Domain-Driven Design (DDD)**.

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
- **Presentation Layer**: Gin HTTP Router and Handlers mapping requests to internal services. GraphQL endpoints generated using `gqlgen`.
- **Application Layer**: Business logic. Services control data access, dynamic schemas, and operations.
- **Domain Layer**: Core data models defining the platform state (`Service`, `Field`, `DatabaseConnection`, `User`, `Role`).
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (OpenTelemetry), metrics (Prometheus), and logging (Zap).

## Metadata Database Schema

The CMS stores platform metadata in the Metadata Database (e.g., PostgreSQL).

Core tables include:
- `database_connections`: External DB configs (id, name, type, host, port, credentials).
- `services`: User-created data models referencing `database_connection_id`.
- `fields`: Attributes for each service (id, name, type, nullable, unique).
- `relations`: Manages relationships between services.
- `service_permissions`: Fine-grained access checks (Role to Service).
- `migrations`: Tracks DDL executions and metadata.
- `backups`: Stores snapshot records or schema outputs.
- `users` & `roles`: General authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure

The project uses a clean modular structure.

```text
.
├── cmd/
│   └── server/          # Main application entrypoint
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection management
│   ├── graphql/         # GraphQL gateway and schemas
│   ├── handlers/        # Gin REST API handlers
│   ├── middleware/      # Auth, Rate Limiter, Prometheus
│   ├── models/          # Core Domain Models
│   ├── services/        # Business logic (CRUD, Service Builder)
│   └── tracing/         # OpenTelemetry tracing setup
├── pkg/
│   ├── logger/          # Zap structured logger
│   └── response/        # Standardized API responses
├── tests/
│   ├── integration/     # Integration test suite
│   ├── uat/             # User Acceptance Testing
│   └── unit/            # Unit test suite
├── docker/              # Docker configuration files
├── Makefile             # Standardized build/test tasks
└── Dockerfile           # Application containerization
```

## Component Implementations

### CMS Control Plane & Service Builder
Users can dynamically create services (data models) and define schemas via the CMS control plane. Handled by `ServiceService` and `DBConnService`, mapping the API definitions directly into the metadata database.

### CRUD Engine & Query Engine
The Dynamic Data Service automatically maps endpoints (e.g., `GET /api/v1/data/{service}`) to GORM database operations. It supports dynamic SQL queries, advanced filtering, sorting, and pagination logic on the fly. Many-to-many relationships automatically create necessary join tables.

### Schema Migration Engine
The `MigrationService` tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM's `.Migrator()`. It supports automatic backup before migration and rollback capabilities, natively logging status into the `migrations` table.

### GraphQL Gateway
The platform automatically generates optional GraphQL endpoints from service schemas using `github.com/99designs/gqlgen/graphql`. The `POST /api/v1/graphql` route exposes the generic optional schemas generated.

### Backup System
The Backup System (`BackupService`) supports generic service record back up, creating data snapshots (SQL dump or JSON snapshot). Supports both table and schema backup with restore capabilities.

## Observability Integration

- **Logging**: Global structured logging using **Zap**, with correlation and trace IDs tied into context.
- **Metrics**: **Prometheus** tracks request latency, status codes, and error rates via standard HTTP interceptors. Available at `/metrics`.
- **Tracing**: **OpenTelemetry** with Jaeger wraps SQL commands and network logic to track distributed traces and performance bottlenecks.

## Security
- Input validation
- SQL injection protection
- JWT authentication
- Rate limiting
- Role-based Access Control (RBAC) and row-level security.

## Unit Testing

The platform includes comprehensive test suites.
- Minimum coverage requirement is 80%.
- Areas tested: Repository, Service, API Handlers.
- To execute the full test suite with coverage, run:
  ```sh
  make test-coverage
  ```
  *(Note: Internally runs `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`)*

## Docker Setup

The platform is fully containerized. A `docker-compose.yml` is provided in the `docker` directory to spin up the application along with Prometheus.

To build and run:
```sh
docker-compose -f docker/docker-compose.yml up -d
```

## Swagger Documentation

Swagger/OpenAPI documentation is automatically generated.
To generate documentation:
```sh
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
