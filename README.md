# CMS Backend Platform

A production-grade Backend-as-a-Service platform (Dynamic CMS + API Builder) written in Go. This system is a self-hosted, Go-native backend infrastructure generator that allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

## Core Features

- **Connect Databases**: Supports PostgreSQL, MySQL, MongoDB, and SQLite.
- **Dynamic Service Builder**: Create data models (services), fields, and relationships on the fly.
- **Dynamic CRUD API Engine**: Auto-generates Create, Read, Update, Delete, and List endpoints for every service.
- **Query Engine**: Advanced querying capabilities including filtering, sorting, pagination, and joining.
- **GraphQL Gateway (Optional)**: Automatically generates generic GraphQL schemas for defined services.
- **Schema Migration Engine**: Handles safe database schema changes dynamically (Add, Drop, Rename Column) with version tracking and rollbacks.
- **Backup Engine**: Built-in support for generating and restoring database snapshots.
- **Observability Integrated**: Features structured logging (Zap), open telemetry distributed tracing (Jaeger), and system metrics (Prometheus).
- **Control Plane & Role-Based Access Control**: Fully secured with JWT authentication and fine-grained permissions.

## High-Level System Architecture

The project adheres to Clean Architecture and Domain-Driven Design (DDD):
- **Presentation Layer**: HTTP Gateway (Gin).
- **Application Layer**: Business Logic & Service Controllers.
- **Domain Layer**: Metadata & Models definition.
- **Infrastructure Layer**: Cross-cutting connections, logging, metric aggregation.

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

The core platform configuration relies on an internal Metadata Database (e.g., PostgreSQL or SQLite):

- `database_connections`: Stores external DB connection config (type, host, port, credentials).
- `services`: Represents user-created data models, maps to external databases.
- `fields`: Field attributes for services (types: string, int, JSON, UUID, etc.).
- `relations`: Configures One-to-One, One-to-Many, Many-to-One, and Many-to-Many links.
- `service_permissions`: Connects Roles to Services for fine-grained authorization.
- `migrations`: DDL schema execution state and history.
- `backups`: Database records and schema backups.
- `users` / `roles`: System authentication.
- `audit_logs`: Detailed tracking of operations.

## Project Structure

```
cmd/
└── server/             # Application entrypoint
    └── main.go
internal/
├── api/                # API handler logic
├── config/             # Platform configuration setup
├── database/           # Connection pooling & ORM integration
├── graphql/            # GraphQL generator and endpoint handler
├── handlers/           # HTTP Gin Request Controllers
├── middleware/         # Auth, Prometheus, Rate Limiter
├── models/             # Domain layer & Database Metadata tables
├── services/           # Application Logic (CRUD, Builder, Migration, etc.)
└── tracing/            # OpenTelemetry setup
pkg/
└── logger/             # Global structured Zap logger
tests/
├── integration/        # Integration tests
├── uat/                # User Acceptance Tests
└── unit/               # Unit tests
docker/                 # Dockerfile, docker-compose, and configurations
```

## Setup & Running

### Requirements
- Go 1.24+
- Docker & Docker Compose (Optional, for containers)
- Make

### Running Locally
To launch the platform locally:

```bash
# Build the binary
make build

# Start the local server
make run

# Start development with hot reloading (requires air)
make dev
```

### Docker

The system supports a full Docker Compose setup including PostgreSQL, Redis, Jaeger, and Prometheus.

```bash
# Start Docker services
make docker-up

# Stop Docker services
make docker-down
```

### Tests

The project is backed by comprehensive tests across all layers (Unit, Integration, and UAT).

```bash
# Run all tests
make test

# Generate Test Coverage report
make test-coverage
```

### Documentation

API documentation is generated dynamically. To generate Swagger Docs:

```bash
make swagger
```
Once the server is running, visit `/api/v1/swagger/index.html`.
