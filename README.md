# Dynamic CMS + API Builder Backend Platform

## Overview
This repository contains a production-grade Backend-as-a-Service (BaaS) platform written in Go. The platform works similarly to Hasura or Supabase but is fully self-hosted and Go-native. It allows users to dynamically create backend services, manage database connections, define schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

The system acts as a backend infrastructure generator, adhering to Clean Architecture and Domain-Driven Design (DDD) principles.

## Core Technology Stack
- **Language**: Go
- **API Layers**: REST (default) via Gin, GraphQL (optional) via gqlgen
- **ORM**: GORM
- **Logging**: Zap structured logging
- **Metrics**: Prometheus
- **Tracing**: OpenTelemetry (Jaeger)
- **Containerization**: Docker
- **API Documentation**: Swagger / OpenAPI

## System Architecture Explanation
The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)**.

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
- **Presentation Layer**: The Gin HTTP Router and Handlers map requests to internal services, including GraphQL endpoints.
- **Application Layer**: Business logic via Services controls data access, generates dynamic schemas, and executes handler requests.
- **Domain Layer**: Data models defining core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools for connection pooling, telemetry, metrics, and logging.

## Metadata Database Schema
The metadata database (e.g., PostgreSQL or SQLite for local dev) stores platform configuration. Core tables include:

- `database_connections`: External DB configurations (id, name, type, host, port, credentials).
- `services`: User-created data models (id, name, database_connection_id, db_table_name, timestamps).
- `fields`: Attributes for each service (name, type, nullable, unique, default_value).
- `service_permissions`: Connects roles to services for fine-grained access (CanCreate, CanRead).
- `migrations`: DDL execution history for schema tracking and rollbacks.
- `backups`: Records of data snapshots or schema outputs.
- `users` / `roles`: Authentication and authorization.
- `audit_logs`: Detailed logging of structural and data-level modifications.

## Go Project Structure
The project uses a clean modular structure:

```
.
├── cmd
│   └── server/          # Main application entrypoint
├── docker/              # Docker configuration files
├── internal/
│   ├── config/          # Environment configuration
│   ├── database/        # Database Connection Manager
│   ├── graphql/         # GraphQL gateway and schema generators
│   ├── handlers/        # API route controllers (Presentation Layer)
│   ├── middleware/      # Auth, security, metrics middleware
│   ├── models/          # Domain layer data structures
│   ├── services/        # Application layer business logic (Service Builder, CRUD Engine, etc.)
│   └── tracing/         # OpenTelemetry setup
├── pkg/
│   ├── logger/          # Zap logger wrapper
│   └── response/        # Standardized HTTP response helpers
├── tests/
│   ├── integration/     # Integration tests
│   ├── uat/             # User Acceptance Testing
│   └── unit/            # Unit tests
├── docs/                # Generated Swagger documentation
├── Makefile             # Build automation
├── docker-compose.yml   # Multi-container Docker deployment
└── README.md
```

## Component Implementations

### Service Builder & CMS Control Plane
Located in `internal/services/service_service.go` and `internal/services/dbconn_service.go`, these services provide the management interface. Users can connect external databases (PostgreSQL, MySQL, MongoDB), define dynamic schemas (`services` and `fields`), and manage relationships (One-to-One, One-to-Many, etc.). The system dynamically adjusts metadata tables and connects to external data sources via the Database Connection Manager.

### CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go`. The Dynamic CRUD Engine automatically exposes endpoints (e.g., `GET /api/v1/data/{slug}`) mapped to the created services. It translates incoming HTTP requests to GORM queries. The integrated Query Engine supports advanced querying:
- Filtering (`?email=john@example.com`)
- Sorting (`?sort=created_at:desc`)
- Pagination (`?page=1&limit=20`)
- Dynamic joining via foreign key mappings.

### Schema Migration Engine
Managed by `internal/services/migration_service.go`. It tracks schema modifications (Add, Drop, Rename Column) using GORM’s Migrator. The system records execution history natively in the `migrations` metadata table, facilitating automated application of schema differences and allowing for rollback capabilities.

### GraphQL Gateway
Managed by `internal/graphql/gateway.go`. The GraphQL engine automatically generates schemas using `gqlgen` from the service definitions. It exposes `POST /api/v1/graphql` to fulfill dynamic queries and mutations seamlessly mapping to the underlying relational engines.

### Backup System
Managed by `internal/services/backup_service.go`. Generates snapshots of dynamic table contents mapping to generic service records. It supports restoring data rows through JSON payload decoding, ensuring structural resilience.

### Observability Integration
Observability is integrated comprehensively:
- **Logging**: Zap logging initialized via `pkg/logger/` ensures all logs contain request, error, and trace metadata.
- **Tracing**: OpenTelemetry (Jaeger) initialized in `internal/tracing/` tracks database and application layer performance.
- **Metrics**: Prometheus metrics via `internal/middleware/prometheus.go` track request latency and error rates, exposed at the `/metrics` endpoint.

## Unit Testing
The platform enforces a minimum of 80% unit test coverage across Repository, Service, and API handler layers.

**To run unit tests:**
```bash
make test-unit
```
*(Run `make test-coverage` for full coverage reporting including the `tests/` isolated suites).*

## Docker Setup
The platform is container-ready.

**To start the full stack (API, databases, observability tools):**
```bash
make docker-up
# or
docker-compose up -d
```

## Swagger Documentation
API documentation is auto-generated using `swag` and integrated into the server.

**To generate and view docs:**
```bash
make swagger
```
The documentation is available locally via the `/swagger/*any` endpoint when the server runs.