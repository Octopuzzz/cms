# Dynamic CMS & API Builder Platform

A production-grade Backend-as-a-Service (BaaS) platform written natively in Go. This system allows developers to connect databases, dynamically create data models, automatically generate CRUD REST endpoints and optional GraphQL APIs, and monitor systems with state-of-the-art observability. It operates similarly to Hasura or Supabase but is fully self-hosted.

## Platform Goals
* Connect external databases dynamically.
* Create data models (services) and relationships on the fly.
* Automatically generate CRUD REST APIs.
* Optionally generate GraphQL APIs.
* Manage schema migrations securely.
* Built-in observability (logging, metrics, tracing).
* Automated backups and restores.

## Core Technologies
* **Language:** Go (1.24)
* **Web Framework:** Gin
* **ORM:** GORM
* **GraphQL:** gqlgen
* **Logging:** Zap (structured logging)
* **Metrics:** Prometheus
* **Tracing:** OpenTelemetry
* **Containerization:** Docker
* **Documentation:** Swagger/OpenAPI

## System Architecture

The project follows Clean Architecture and Domain-Driven Design (DDD) principles:

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

- **Presentation Layer (`internal/handlers/`)**: Gin HTTP Router and Handlers, acting as the API gateway.
- **Application Layer (`internal/services/`)**: Core business logic controlling data access, dynamic schemas, and handlers requests.
- **Domain Layer (`internal/models/`)**: Core entities like `Service`, `Field`, `DatabaseConnection`, `User`.
- **Infrastructure Layer (`internal/tracing/`, `pkg/logger/`)**: Telemetry, metrics, and cross-cutting tools.

## Project Structure

```text
├── cmd/
│   └── server/          # Main application entrypoint
├── internal/
│   ├── api/             # API routes definitions
│   ├── config/          # Application configurations
│   ├── database/        # Database connection manager (PostgreSQL, MySQL, SQLite)
│   ├── graphql/         # GraphQL gateway (gqlgen)
│   ├── handlers/        # HTTP handlers for CMS, CRUD, Auth, etc.
│   ├── middleware/      # Auth, Rate Limiting, Prometheus
│   ├── models/          # Core domain models (GORM)
│   ├── services/        # Business logic for dynamic schemas, CRUD, etc.
│   └── tracing/         # OpenTelemetry setup
├── pkg/
│   ├── logger/          # Zap structured logger
│   └── response/        # Standardized Gin API responses
├── tests/
│   ├── unit/            # Unit tests for services, handlers
│   ├── integration/     # Integration tests
│   └── uat/             # User Acceptance Testing
├── docker/              # Dockerfiles and docker-compose
├── Makefile             # Development commands
├── .env.example         # Environment configuration
└── go.mod               # Go dependencies
```

## Core Components

### Metadata Database Schema
The control plane stores configuration in the internal Metadata Database. Key tables:
* `database_connections`: External DB configurations (host, port, credentials).
* `services`: User-created data models referencing database connections.
* `fields`: Attributes for each service (types like string, uuid, JSON, relations).
* `migrations`: Tracks DDL executions and state.
* `backups`: Stores snapshot records.
* `users` / `roles` / `service_permissions`: RBAC for control plane and dynamic APIs.

### Service Builder & Control Plane
Managed under `/api/v1/cms/`. Allows developers to create `services` (models) and `fields`.
* Supports dynamic field definitions (string, integer, float, boolean, uuid, json, timestamp).
* Supports relationships (One-to-One, One-to-Many, Many-to-One, Many-to-Many).
* Behind the scenes, triggers the Schema Migration Engine to alter physical schemas.

### CRUD Engine & Query Engine
Managed by `internal/services/dynamic_data_service.go` and handlers.
* Auto-generates `GET`, `POST`, `PUT`, `DELETE` endpoints for every created service (e.g., `/api/v1/data/{slug}`).
* Provides advanced queries natively:
  * **Filtering**: `?email=john@example.com`
  * **Sorting**: `?sort=created_at:desc`
  * **Pagination**: `?page=1&limit=20`
* Dynamically executes SQL using GORM based on the requested model definition.

### Schema Migration Engine
Safe schema evolution via `internal/services/migration_service.go`.
* Tracks and applies safe database changes (Add, Drop, Rename).
* Capable of basic rollback operations to previous stable metadata states.
* Interacts with GORM's `Migrator` for native DDL adjustments.

### GraphQL Gateway
Managed by `internal/graphql/gateway.go`.
* Automatically exposes a GraphQL endpoint `/api/v1/graphql`.
* Reflects the current data schemas and provides dynamic queries mapping to the underlying data stores.

### Backup System
Endpoints at `/api/v1/cms/backups/`.
* Managed by `internal/services/backup_service.go`.
* Supports backing up specific service tables to JSON snapshots.
* Allows restoring service states from previous snapshots.

### Observability Integration
End-to-end telemetry baked in:
* **Zap Logging**: Structured JSON logging.
* **OpenTelemetry**: Distributed tracing wrapping HTTP routes and database queries.
* **Prometheus Metrics**: Exposes latency, error rates, and request counts at `/metrics`.

### Security & Performance
* JWT-based Authentication and granular Role-Based Access Control (RBAC).
* Configurable Rate Limiting.
* SQL Injection protection via parameterized GORM queries.
* Database connection pooling natively supported in `DatabaseConnectionManager`.

## Development & Testing

### Running the Application

```bash
# Start required infrastructure (Database, Redis, Jaeger, Prometheus)
docker-compose -f docker/docker-compose.yml up -d

# Copy environment variables
cp .env.example .env

# Run the server (with air for hot-reload if installed)
make dev
# OR natively
make build
./bin/cms-backend
```

### Unit Tests
The project maintains a strict >80% coverage requirement across Services, Repositories, and Handlers. Tests use isolated in-memory SQLite instances.

```bash
# Run unit tests
make test-unit

# Run full test suite with coverage
make test-coverage
```

### Swagger Documentation
API documentation is generated using standard `swag`.

```bash
# Install swag if missing
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

The Swagger UI will be available at `/swagger/index.html` when running the application.

## Docker Setup
The platform is fully containerized. A multi-stage `Dockerfile` is provided for the core application, ensuring small production images.
`docker-compose.yml` configures all dependent services including PostgreSQL, Prometheus, and OpenTelemetry tools.
