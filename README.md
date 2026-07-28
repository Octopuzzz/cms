# CMS Backend Platform

A production-grade, self-hosted, Go-native Backend-as-a-Service (BaaS) platform. This platform allows developers to dynamically create backend services, define schemas, automatically generate CRUD REST APIs, generate optional GraphQL endpoints, and includes built-in observability, schema migrations, and backups.

## Core Features
*   **Dynamic Services**: Create data models dynamically via the CMS Control Plane.
*   **Auto-Generated APIs**: Automatically get standard REST (CRUD, list, filter, sort) endpoints for any created service.
*   **GraphQL Gateway**: Generates generic, optional GraphQL endpoints via `gqlgen` to fulfill automated GraphQL endpoints for defined data models.
*   **Multi-Database Support**: Connect and manage multiple databases (PostgreSQL, MySQL, MongoDB, SQLite).
*   **Metadata Management**: Built-in tracking for metadata, such as services, fields, relations, permissions, migrations, and backups.
*   **Schema Migration Engine**: Safely track and apply schema diffs (Add, Drop, Rename Column) using GORM's Migrator natively.
*   **Backup System**: Generate data snapshots and schema outputs, with support for restoring via JSON payloads.
*   **Observability**: Integrated Prometheus metrics, OpenTelemetry tracing (Jaeger), and Zap structured logging.
*   **Security & RBAC**: Fine-grained role-based access control, Row-Level/Field-Level permissions, input validation, rate limiting, and JWT authentication.

## System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles to ensure separation of concerns and a modular structure.

*   **Client Layer**: Consumes the exposed REST and GraphQL APIs.
*   **Presentation Layer (`internal/handlers/`)**: The Gin HTTP Router and Handlers. Acts as the API gateway mapping external requests to internal application services, including GraphQL execution via `gqlgen`.
*   **Application Layer (`internal/services/`)**: The core business logic. Defines how services interact to provide dynamic schemas, process queries, create databases, migrate schemas, and perform observability logic.
*   **Domain Layer (`internal/models/`)**: Defines the fundamental data structs and types representing core system entities such as `Service`, `Field`, `DatabaseConnection`, `User`, `Role`, etc.
*   **Infrastructure Layer**: Represents cross-cutting concerns, utilities, and integrations. Found under `internal/tracing/` (OpenTelemetry), `pkg/logger/` (Zap), `internal/database/` (Connection Manager), and Prometheus metrics.

### High-Level Architecture Diagram
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

## Metadata Database Schema

The core configuration for the platform's dynamic capabilities is persisted in the metadata database (typically PostgreSQL or SQLite).

Important entities in the schema:
*   `database_connections`: Contains ID, name, connection type, host, port, credentials, pool settings, and database name.
*   `services`: Maps to user-created data models. Tracks schema state, database ID, and table configurations.
*   `fields`: Attributes attached to services (string, integer, UUID, boolean). Tracks configurations like nullable, unique, and default values.
*   `service_permissions`: Defines role-to-service mapping for granular API authorization.
*   `migrations`: Records migration history for automated rollbacks and status tracking.
*   `backups`: Stores backup snapshots or schema states.
*   `users` / `roles`: Auth entities for users and custom role generation.
*   `audit_logs`: Activity and structural logging table.

## Go Project Structure

The structure represents standard Go layouts tailored for scalability:
```
cmd/
└── server/             # Application entrypoint (main.go)
internal/
├── config/             # Application environment/config parsing
├── database/           # Connection management, pooling, DB connectors
├── graphql/            # GraphQL dynamic gateway schema generator
├── handlers/           # HTTP controllers / Presentation Layer
├── middleware/         # Auth, tracing, rate limiting, logging
├── models/             # Domain layer models and metadata schema definitions
├── services/           # Application layer logic (CMS, Auth, Data, Migrations, etc)
└── tracing/            # OpenTelemetry setups
pkg/
├── logger/             # Zap logger implementation
└── response/           # Reusable standardized API responses
tests/
├── integration/        # External dependency integration tests
├── uat/                # End-to-end / UAT tests
└── unit/               # Service and handler level unit tests
docker/                 # Dockerfile, docker-compose, and configurations
```

## Platform Components

### CRUD Engine Implementation
The CRUD engine automatically maps dynamic REST calls to GORM operations. Utilizing standard `/api/v1/data/{slug}` routes, the dynamic data service dynamically binds input data to mapped services. The query engine handles standard parameters like `?sort`, `?limit`, filtering via specific keys, and pagination responses dynamically based on configuration.

### Schema Migration Engine
Safe schema evolution is achieved via a dedicated Migration Service. When a schema requires altering (adding/dropping/renaming columns), GORM's built-in migrator evaluates the structural diffs. Automatic migrations are logged to track state, and changes support robust rollback functionality via integrated snapshot retention logic.

### GraphQL Gateway
Built using `99designs/gqlgen`, the gateway provides a generic abstraction layer enabling client applications to hit an optional GraphQL endpoint (`POST /api/v1/graphql`). This gateway introspects the underlying generated dynamic schemas, constructing queries, mutations, and field-level relational navigation without any manual schema building.

### Backup System
A fully featured Backup Engine exports data payloads for robust system protection.
The system implements backups based on service metadata mapping dynamic table contents to robust snapshot schemas (including JSON snapshots/exports) with straightforward API hooks to initiate and restore backups directly.

### Observability Integration
Observability is seamlessly stitched across all operations:
*   **Logging**: `pkg/logger/` manages structured JSON logging utilizing `go.uber.org/zap`.
*   **Tracing**: `internal/tracing/` sets up OpenTelemetry (OTel), capturing network boundaries and SQL queries with trace and correlation IDs.
*   **Metrics**: Custom Gin middlewares expose standardized RED (Rate, Errors, Duration) Prometheus metrics under the `/metrics` endpoint.

## Usage & Commands

This project standardizes common tasks inside `Makefile`.

### Unit Tests
The project maintains a >80% testing threshold covering Repositories, Services, and Handlers.
*   `make test-unit`: Run unit tests only.
*   `make test-coverage`: Run the entire suite and generate a coverage report (`coverage.out` / `coverage.html`).
*   `go test ./...`: General command to run all test packages.

### Docker Setup
The project defines a comprehensive Docker stack combining the backend application, its metadata store, and observability stack.
*   `docker-compose up -d` or `make docker-up`: Start the entire distributed environment (Backend, DB, Redis, Prometheus, Jaeger).

### Swagger Documentation
Swagger docs (OpenAPI) are dynamically generated across the `cmd/server/main.go` annotations.
*   `make swagger`: Updates the `docs/` folder using the `swag init` command.
*   Accessible at `/swagger/index.html` on standard deployment configurations.
