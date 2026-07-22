# CMS Backend Platform (Dynamic CMS + API Builder)

A production-grade Backend-as-a-Service (BaaS) platform built natively in Go. This platform enables developers to dynamically create backend services, define schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability. It acts as a comprehensive backend infrastructure generator.

## 1. System Architecture Explanation

The system is built on **Clean Architecture** and **Domain-Driven Design (DDD)** principles, ensuring a modular, scalable, and testable codebase.

- **Presentation Layer**: Built with Gin framework (`internal/handlers/`), acting as the API gateway. This layer maps REST requests to internal services and hosts optional GraphQL endpoints using `gqlgen`.
- **Application Layer**: Contains core business logic (`internal/services/`). It coordinates between components, generates dynamic schemas, and applies business rules.
- **Domain Layer**: Defines core data structures and metadata models (`internal/models/`), such as `Service`, `Field`, and `DatabaseConnection`.
- **Infrastructure Layer**: Handles cross-cutting concerns like observability with OpenTelemetry tracing (`internal/tracing/`), Prometheus metrics (`internal/middleware/prometheus.go`), structured Zap logging (`pkg/logger/`), and database connection management caching.

## 2. Metadata Database Schema

The platform stores its configuration in a core metadata database (using GORM). Core tables include:

- `database_connections`: Stores external database configurations (id, name, type, host, port, credentials) and connection pool settings. Supports PostgreSQL, MySQL, and MongoDB.
- `services`: Represents user-created data models dynamically mapped to actual database tables (`db_table_name`), linking back to `database_connection_id`.
- `fields`: Defines attributes for each service (name, type, nullable, unique, default_value, index). Supported types include string, integer, float, boolean, uuid, json, timestamp, etc.
- `relations`: Defines one-to-one, one-to-many, many-to-one, and many-to-many relationships (creating join tables).
- `migrations`: Keeps track of DDL executions for safe schema changes.
- `backups`: Stores snapshot records and schema outputs.
- `users` / `roles` / `service_permissions`: Handles Role-Based Access Control (RBAC).
- `audit_logs`: Detailed tracking of structural and data modifications.

## 3. Go Project Structure

The project follows standard Go layout conventions:

- `cmd/server/`: The entry point for the application (`main.go`).
- `internal/`: Private application and library code.
  - `handlers/`: HTTP request handlers (REST).
  - `services/`: Business logic, core engines (CRUD, Migration, Service Builder).
  - `models/`: GORM database models (Domain Layer).
  - `graphql/`: GraphQL gateway implementation.
  - `database/`: Database connection management and pooling.
  - `middleware/`: Gin middleware (Auth, Prometheus, Rate Limiting).
  - `config/`: Application configuration.
  - `tracing/`: OpenTelemetry setup.
- `pkg/`: Public library code that can be used by external applications (e.g., `logger`, `response`).
- `tests/`: Integration, Unit, and UAT tests.
- `docker/`: Docker Compose configurations and Prometheus config.

## 4. CRUD Engine Implementation

Implemented in `internal/services/dynamic_data_service.go`, the dynamic CRUD engine automatically maps REST endpoints (e.g., `GET /api/v1/data/{slug}`) to GORM database operations on the fly.

- **Dynamic Operations**: Maps standard Create, Read, Update, Delete, and List operations for dynamic services without hardcoding handlers.
- **Query Engine**: Applies dynamic filtering (e.g., `?email=john@example.com`), sorting (e.g., `?sort=created_at:desc`), pagination (`?page=1&limit=20`), and relation joining (e.g., `?join=user`).
- **Security**: Verifies Role-Based Access Control and Row-Level security dynamically prior to executing queries.

## 5. Schema Migration Engine

Implemented in `internal/services/migration_service.go`, this engine provides safe schema modifications.

- **Capabilities**: Dynamically add, drop, or rename columns, and change column types using GORM’s `.Migrator()`.
- **Safety**: Automatically tracks migrations in the `migrations` metadata table, facilitating safe rollbacks and history tracking. Integrates with the backup system to snapshot data before critical schema changes.

## 6. GraphQL Gateway

Located at `internal/graphql/gateway.go`, the GraphQL gateway provides an optional interface.

- Automatically leverages `gqlgen` to build a dynamic GraphQL executable schema based on the active service definitions.
- Exposes generic queries, mutations, and relationship fetching through a single unified `POST /api/v1/graphql` endpoint.

## 7. Backup System

Managed by `internal/services/backup_service.go`, the backup engine safeguards metadata and dynamic service data.

- **Snapshots**: Generates generic service record backups, mapping dynamic table contents to JSON snapshots or SQL dumps.
- **Restoration**: Supports restoring rows directly from JSON payloads.
- **API Access**: Accessible via endpoints like `POST /cms/backup/service/{service_id}`.

## 8. Observability Integration

The platform provides comprehensive production-ready observability:

- **Logging**: Uses `go.uber.org/zap` (`pkg/logger/`) for fast, structured JSON logging, complete with trace and correlation IDs.
- **Metrics**: Exposes Prometheus metrics via `internal/middleware/prometheus.go` at `/metrics`, tracking request latency, slow queries, error rates, and throughput.
- **Distributed Tracing**: Uses OpenTelemetry (`internal/tracing/`) to wrap HTTP handlers, database queries, and external calls, exporting spans to Jaeger.

## 9. Docker Setup

The repository is containerized for seamless deployment.

- `Dockerfile`: Multi-stage build process optimized for production, resulting in a lightweight Alpine Linux image.
- `docker-compose.yml`: Spools up the entire stack, including the Go Backend, PostgreSQL (Metadata Database), Redis (Caching), Prometheus (Metrics scraping), and Jaeger (Trace collector) across a unified `cms-network`.

## 10. Swagger Documentation

API Documentation is built directly from source using Swaggo.

To generate the documentation, run:
```bash
~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
(Or use `make swagger` if available).
The interactive OpenAPI UI is automatically served by the application when running.

## 11. Unit Tests

The platform prioritizes high reliability, enforcing a minimum of **80%** test coverage across the repository, services, and API handlers.

Run unit tests and view coverage using the Makefile:
```bash
make test-unit
```
Or directly via the Go CLI:
```bash
go test ./... -v
```
For integration tests that require a database connection, tests are configured to automatically fall back to an in-memory SQLite instance (`:memory:`) using standard database mocks.
