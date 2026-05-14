# Backend Platform (Dynamic CMS + API Builder)

A self-hosted, Go-native Backend-as-a-Service (BaaS) platform that dynamically generates REST and GraphQL APIs.

This platform acts as a backend infrastructure generator, allowing you to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor your system.

## High Level Platform Architecture

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

## System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:
- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## Project Structure

```text
.
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── config/          # Environment configuration
│   ├── database/        # Database connection management
│   ├── graphql/         # GraphQL gateway and resolvers
│   ├── handlers/        # HTTP API handlers
│   ├── middleware/      # Gin middleware (Auth, Metrics, Tracing)
│   ├── models/          # Domain models (Service, Field, User, etc.)
│   ├── services/        # Business logic and engines
│   └── tracing/         # OpenTelemetry setup
├── pkg/                 # Public packages
│   ├── logger/          # Zap structured logger
│   └── response/        # Standardized HTTP responses
├── tests/               # Testing suite
│   ├── unit/            # Unit tests
│   ├── integration/     # Integration tests
│   └── uat/             # User Acceptance Tests
├── docker/              # Docker configuration files
├── Dockerfile           # Docker image build instructions
├── docker-compose.yml   # Multi-container orchestration
├── Makefile             # Development tasks and commands
├── go.mod               # Go module dependencies
└── .env.example         # Environment variable template
```

## Core Components

### Metadata Database

The core internal configuration is stored in the **Metadata Database** (e.g., PostgreSQL). Key tables include:

- `database_connections`: External DB configurations (id, name, type, host, port, credentials).
- `services`: User-created data models referencing a DB connection.
- `fields`: Attributes for each service (type, uniqueness, nullability, defaults).
- `service_permissions`: Role-based access controls for services.
- `migrations`: DDL execution history for rollbacks.
- `backups`: Snapshot records.
- `users` and `roles`: Control plane authentication/authorization.
- `audit_logs`: Detailed logging of modifications.

### Database Connection Manager

Registers external databases (PostgreSQL, MySQL, MongoDB, SQLite) and manages connection pooling for efficient resource utilization.

### Service Builder & CMS Control Plane

Provides an API (under `/api/v1/cms/`) to define services (data models) and their fields/relations in the metadata database. Triggers schema migrations upon updates.

### CRUD & Query Engine

Maps REST endpoints (e.g., `GET /api/v1/data/{service}`) to dynamic GORM database operations. Supports advanced queries including filtering, sorting, pagination, and join queries. Enforces Role-Based Access Control and Row-Level filtering dynamically.

### Schema Migration Engine

Tracks and applies safe schema changes (Add, Drop, Rename Column) using GORM’s Migrator. Logs status and handles potential rollbacks.

### GraphQL Gateway

Automatically generates GraphQL schemas from service definitions. Exposes `POST /api/v1/graphql` for dynamic queries and mutations.

### Backup Engine

Generates data snapshots of dynamic table contents (e.g., JSON snapshots) and supports restoring rows via JSON payload decoding.

### Observability Integration

- **Logging**: Zap structured logging injected globally with correlation and trace IDs.
- **Tracing**: OpenTelemetry/Jaeger wraps SQL commands and network logic.
- **Metrics**: Prometheus tracks request latency and status codes (exposed at `/metrics`).

## Development Guide

### Docker Setup

Start the infrastructure using Docker Compose:

```bash
make docker-up
# or
docker-compose up -d
```

### Running Tests

The project includes unit, integration, and UAT tests with an 80%+ coverage target.

```bash
# Run all tests
make test

# Generate coverage report
make test-coverage
# or
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

### Swagger Documentation

Generate API documentation using Swagger/OpenAPI:

```bash
make swagger
# Start the server and visit /swagger/index.html
```
