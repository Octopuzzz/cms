# CMS Backend Platform

A production-grade Backend-as-a-Service platform written in Go. This platform acts as a backend infrastructure generator, allowing developers to dynamically create backend services, manage database connections, define schemas, and auto-generate REST/GraphQL APIs.

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

### High-Level Components
- **CMS Control Plane**: Management interface to control the system (DB connections, services, schemas).
- **Service Builder**: Engine to dynamically create services (data models) and their fields/relations.
- **CRUD Engine**: Automatically generates and handles CRUD operations for defined services.
- **Query Engine**: Supports advanced queries (filtering, sorting, pagination, joins).
- **Schema Migration Engine**: Safely manages database schema changes and rollbacks.
- **Backup Engine**: Handles table/schema backups and snapshots.
- **Observability Engine**: Integrates Zap logging, Prometheus metrics, and OpenTelemetry tracing.

## 2. Metadata Database Schema

The platform metadata is stored in the primary database (PostgreSQL by default).

**Core Tables:**
- `database_connections`: Stores external DB configs (id, name, type, host, port, credentials).
- `services`: Represents user-created data models. Tracks dynamic schemas (`db_table_name`).
- `fields`: Defines attributes for services (type, nullable, unique, default, index).
- `service_permissions`: Connects `Role` to `Service` for fine-grained RBAC.
- `migrations`: Tracks applied DDL executions and rollback metadata.
- `backups`: Stores snapshot records or schema outputs.
- `users` / `roles`: Authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of modifications.

## 3. Go Project Structure

```
.
├── ARCHITECTURE.md
├── Dockerfile
├── Makefile
├── cmd/
│   └── server/
│       └── main.go           # Application entrypoint
├── docker/
│   └── (Database init scripts, etc)
├── docker-compose.yml
├── go.mod
├── go.sum
├── internal/
│   ├── config/               # Configuration loading
│   ├── database/             # Connection manager
│   ├── graphql/              # GraphQL gateway (gqlgen)
│   ├── handlers/             # HTTP Handlers (Gin)
│   ├── middleware/           # Auth, metrics, tracing
│   ├── models/               # Domain models
│   ├── services/             # Core business logic
│   └── tracing/              # OpenTelemetry setup
├── pkg/
│   └── (Shared utilities)
└── tests/
    ├── integration/
    ├── uat/
    └── unit/
```

## 4. CRUD Engine Implementation

The Dynamic CRUD Engine (`internal/services/dynamic_data_service.go`) maps generic endpoints to GORM database operations on the fly.

**Endpoints:**
- `POST /api/v1/data/{slug}`: Create record
- `GET /api/v1/data/{slug}/{id}`: Read record
- `PUT /api/v1/data/{slug}/{id}`: Update record
- `DELETE /api/v1/data/{slug}/{id}`: Delete record
- `GET /api/v1/data/{slug}`: List records

It automatically applies filtering via query parameters, dynamic joining via foreign key mappings, pagination, and verifies Role-Based Access Control dynamically before executing queries.

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`, the migration engine allows safe schema changes to dynamically generated services.

- **Operations Supported**: Add column, Drop column, Rename column.
- **Safety**: Utilizes GORM's `Migrator()`. Tracks migration status in the `migrations` table and handles rollbacks through snapshot retention logic.

## 6. GraphQL Gateway

The system includes an optional GraphQL gateway managed by `internal/graphql/gateway.go`.

- **Implementation**: Uses `github.com/99designs/gqlgen/graphql`.
- **Endpoint**: Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.

## 7. Backup System

Managed by `internal/services/backup_service.go`.

- **Features**: Generates data snapshots of dynamic table contents. Supports restoring rows via JSON payload decoding. Tracks backups in the metadata database.

## 8. Observability Integration

- **Logging**: Structured logging using Uber's `Zap`. Logs contain correlation and trace IDs tied directly to the Context.
- **Metrics**: Prometheus metrics exported at `/metrics`. Tracks request latency, status codes, and HTTP interceptors.
- **Tracing**: OpenTelemetry (Jaeger) initialized in `internal/tracing/` to trace SQL commands and network logic.

## 9. Unit Tests

The system includes comprehensive unit tests with a target minimum coverage of 80%.

- **Coverage**: Repository, Service, and API handlers.
- **Command**: Run the full test suite using `make test-coverage`.

## 10. Docker Setup

The platform is containerized and ready for cloud deployment.

- `Dockerfile`: Contains the multi-stage build instructions for the Go application.
- `docker-compose.yml`: Spins up the CMS Backend along with necessary infrastructure (e.g., PostgreSQL, Redis, Jaeger/Prometheus).
- **Commands**:
  - `make docker-up`: Start Docker services
  - `make docker-down`: Stop Docker services

## 11. Swagger Documentation

API documentation is generated using `swag`.

- **Command**: `make swagger` (executes `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`).
- **Output**: Generates OpenAPI/Swagger documentation in the `docs/` directory (ignored by git).
