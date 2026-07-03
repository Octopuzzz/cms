# CMS Backend Platform

This repository contains the source code for a production-grade Backend-as-a-Service (BaaS) platform. It provides a dynamic CMS and API Builder written in Go, acting as a fully self-hosted, Go-native alternative to platforms like Hasura or Supabase.

The platform allows developers to dynamically connect databases, define data models, and automatically generate CRUD REST endpoints and optional GraphQL APIs, complete with observability, migrations, backups, and more.

## Expected Output Implementations

The following requirements have been implemented and are documented below:

### System Architecture Explanation

The CMS Backend strictly adheres to **Clean Architecture** and **Domain-Driven Design (DDD)** principles to ensure it is modular, scalable, and testable:

- **Presentation Layer**: Handled by the Gin HTTP framework (located in `internal/handlers/`), this acts as the API gateway for mapping REST requests to internal services, and includes an optional GraphQL gateway via `gqlgen`.
- **Application Layer**: Contains business logic (`internal/services/`). Services orchestrate data flow, generate dynamic queries, control database connections, and handle schema management.
- **Domain Layer**: Defines core platform entities (`internal/models/`) such as `Service`, `Field`, `DatabaseConnection`, `User`, and `Role`.
- **Infrastructure Layer**: Cross-cutting concerns are handled here, including robust database connector logic, OpenTelemetry tracing (`internal/tracing/`), Prometheus metrics (`internal/middleware/prometheus.go`), and structured Zap logging (`pkg/logger/`).

### Metadata Database Schema

The core internal configuration is stored in a structured **Metadata Database**, defaulting to SQLite/PostgreSQL. The key metadata tables defined in `internal/models/` include:

- `database_connections`: Stores external DB configurations with id, name, type (PostgreSQL, MySQL, MongoDB), host, port, credentials, and connection pooling settings.
- `services`: Represents user-created data models. Includes fields for id, name, reference to a `database_connection_id`, and tracking metadata.
- `fields`: Defines column-level attributes for each service, such as type (string, integer, uuid, etc.), uniqueness, nullability, and default values.
- `relations`: Manages relationship configurations (One-to-One, One-to-Many, Many-to-Many) between dynamically generated schemas.
- `migrations`: Keeps track of DDL executions with metadata required for tracking schema history.
- `backups`: Stores backup snapshot metadata.
- `users` and `roles`: Provides authentication and authorization (RBAC) to secure the control plane.

### Go Project Structure

The project utilizes a clean and modular directory structure common in Go applications:

- `cmd/server/`: Contains the main application entry point.
- `internal/`: Houses the private application code.
  - `config/`: Application configuration loading.
  - `database/`: Database connection managers and pooling logic.
  - `graphql/`: GraphQL schema generation and resolvers.
  - `handlers/`: HTTP request handlers (REST).
  - `middleware/`: HTTP middlewares (Auth, Rate Limiting, Metrics).
  - `models/`: Domain models and metadata schema definitions.
  - `services/`: Core business logic (CRUD Engine, Service Builder, Migration Engine, Backup Engine).
  - `tracing/`: OpenTelemetry setup.
- `pkg/`: Publicly importable utility packages.
  - `logger/`: Zap logging configuration.
  - `response/`: Standardized API responses.
- `tests/`: Organized test suites (`unit/`, `integration/`, `uat/`).
- `docker/`: Dockerfiles and Prometheus configs.
- `Makefile`: Automates common build, run, test, and swagger generation tasks.

### CRUD Engine Implementation

The system automatically generates REST API operations when a service is created, handled natively by the `DynamicDataService` located in `internal/services/dynamic_data_service.go`.

- **Operations**: Provides dynamic mapping for Create, Read, Update, Delete, and List endpoints (`/api/v1/data/{slug}`).
- **Dynamic Queries**: The `DynamicDataService` parses HTTP query parameters to construct complex GORM database queries on the fly.
  - **Filtering**: e.g., `?email=john@example.com`
  - **Sorting**: e.g., `?sort=created_at:desc`
  - **Pagination**: Supports standard offset/limit pagination parameters.
- **Relational Joins**: Supports automated query joins based on the defined relation metadata schemas.

### Schema Migration Engine

Safe dynamic schema modifications are handled by the `MigrationService` in `internal/services/migration_service.go`.

- **Supported Operations**: The engine dynamically invokes standard GORM auto-migration capabilities allowing for adding columns, dropping columns, renaming, and changing data types.
- **Tracking**: Migration statuses and history are natively logged into the `migrations` metadata table.
- **Safety**: Supports rollback capabilities and hooks into the backup engine before executing structural changes.

### GraphQL Gateway

The platform implements an automatically generated optional GraphQL endpoint via `internal/graphql/gateway.go`.

- Uses `github.com/99designs/gqlgen` to map dynamic service models into generic schema configurations.
- Resolves GraphQL Queries and Mutations to provide flexible data retrieval matching the REST CRUD Engine logic.
- Exposed natively on `/api/v1/graphql`.

### Backup System

Data resilience is managed by the `BackupService` in `internal/services/backup_service.go`.

- **Table / Schema Backup**: Generates data snapshots mapping dynamic table contents to structured outputs (JSON).
- **Service Snapshot**: Capable of saving entire service configurations and related relational schemas.
- **Restore**: Provides capabilities to safely restore metadata configuration and row data.

### Observability Integration

Enterprise-grade observability is fully integrated across the stack:

- **Logging**: Configured via `pkg/logger/` using Uber's `zap` structured logger to capture request logs, error logs, and SQL query executions with associated trace IDs.
- **Metrics**: Standard `Prometheus` HTTP interceptors wrap endpoints inside `internal/middleware/prometheus.go` to measure request latency, request count, and error rates. Exposed natively on `/metrics`.
- **Tracing**: End-to-end distributed tracing is managed by `OpenTelemetry` inside `internal/tracing/`, configured to output to standard OTLP collectors (like Jaeger). Trace IDs are propagated correctly through standard Go Contexts.

### Unit Tests

Robust testing practices ensure code reliability, aiming for a minimum 80% coverage:

- **Structure**: Tests are split into `tests/unit`, `tests/integration`, and `tests/uat`.
- **Scope**: Covers standard Repositories, Service Layers, and API Handlers using an in-memory SQLite database (`:memory:`) configured through `database.ConnectionManager`.
- **Execution**: Can be executed via the Make target `make test`. Full coverage generation including `internal/` packages is achieved via `make test-coverage`.

### Docker Setup

The application is containerized and cloud-ready.

- `Dockerfile`: A multi-stage build definition found in the root directory ensuring a minimal, secure production image compiling the static Go binary.
- `docker-compose.yml`: Local orchestrator which spins up the `cms-backend` along with required infrastructure pieces.

### Swagger Documentation

Standard OpenAPI / Swagger 2.0 documentation is natively supported via `swaggo`.

- Endpoints are annotated inside `internal/handlers/`.
- Docs are generated using `make swagger`, which invokes `swag init` building out JSON/YAML definitions in the `docs/` directory.

---
*This README serves as documentation for the implemented Backend-as-a-Service features fulfilling the primary architectural design requirements.*
