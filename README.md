# Backend Platform (Dynamic CMS + API Builder)

A production-grade Backend Platform and Dynamic CMS written in Go. This system operates similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native. It empowers developers to connect databases, create data models dynamically, and automatically generate REST and GraphQL APIs.

---

## 1. System Architecture Explanation

The system is built on **Clean Architecture** and **Domain-Driven Design (DDD)** principles, ensuring modularity, scalability, and maintainability.

### High-Level Components:
- **Presentation Layer**: Handles incoming HTTP/REST requests (via Gin) and GraphQL queries (via gqlgen). Acts as the API gateway.
- **Application Layer**: Contains business logic (services). Coordinates interactions between handlers and domain models.
- **Domain Layer**: Defines core data structures and interfaces (`internal/models`), establishing the business logic entities (e.g., Services, Fields, DatabaseConnections).
- **Infrastructure Layer**: Manages cross-cutting concerns like logging (Zap), metrics (Prometheus), tracing (OpenTelemetry/Jaeger), caching (Redis), and external database connections (GORM).

### High-Level Platform Architecture

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

---

## 2. Metadata Database Schema

The platform requires a metadata database (typically PostgreSQL or SQLite) to manage its internal state. The primary entities include:

- **databases** (`database_connections`): Stores configurations for external data sources (id, name, type, host, port, credentials).
- **services** (`services`): Represents dynamically created data models (tables) configured by users.
- **fields** (`fields`): Defines attributes for each service, configuring types, nullability, uniqueness, and defaults.
- **relations** (via `relation_config` in fields): Defines one-to-one, one-to-many, many-to-one, and many-to-many relationships.
- **permissions** (`service_permissions`): Defines RBAC rules and row-level security per service and role.
- **migrations** (`migrations`): Tracks schema changes applied to the databases to support safe rollbacks.
- **backups** (`backups`): Stores metadata regarding snapshots and schema backups.
- **users & roles**: System users and Role-Based Access Control settings.
- **audit_logs**: Keeps an immutable trail of actions performed within the CMS control plane.

---

## 3. Go Project Structure

The project is structured modularly following standard Go community conventions:

```text
.
├── cmd/
│   └── server/          # Main application entrypoint
├── docker/              # Dockerfile and docker-compose configurations
├── internal/
│   ├── api/             # API routing and handlers logic
│   ├── config/          # Application configuration (env, YAML)
│   ├── database/        # Database connection managers (PostgreSQL, MySQL, MongoDB, SQLite)
│   ├── graphql/         # GraphQL gateway (gqlgen)
│   ├── handlers/        # Gin HTTP handlers
│   ├── middleware/      # HTTP middleware (Auth, CORS, Prometheus, Security)
│   ├── models/          # Core Domain Models and database schemas
│   ├── services/        # Application services (CRUD Engine, CMS, Migrations, etc.)
│   └── tracing/         # OpenTelemetry & Jaeger integration
├── pkg/                 # Shared utilities
│   └── logger/          # Structured Zap logger wrapper
├── tests/               # Testing suite
│   ├── integration/     # Integration tests
│   ├── uat/             # User Acceptance Testing
│   └── unit/            # Unit tests for services and repositories
├── .env.example         # Example environment variables
├── ARCHITECTURE.md      # Detailed architecture documentation
├── Dockerfile           # Standard Docker build configuration
├── docker-compose.yml   # Multi-container orchestration setup
├── go.mod               # Go module dependencies
└── Makefile             # Make targets for building, testing, running
```

---

## 4. CRUD Engine Implementation

The Dynamic CRUD Engine (`internal/services/dynamic_data_service.go`) maps standard HTTP endpoints to dynamically configured database tables.

When a user defines a service in the CMS, the platform automatically exposes REST endpoints:

- `POST /api/v1/data/{service_slug}` - Create new records.
- `GET /api/v1/data/{service_slug}` - List records with pagination and filtering.
- `GET /api/v1/data/{service_slug}/{id}` - Retrieve a specific record by ID.
- `PUT /api/v1/data/{service_slug}/{id}` - Update an existing record.
- `DELETE /api/v1/data/{service_slug}/{id}` - Delete a record.

### Advanced Query Engine Features
The engine intercepts URL queries to construct sophisticated GORM actions:
- **Filtering**: `?email=john@example.com` constructs standard WHERE clauses.
- **Sorting**: `?sort=created_at:desc` handles ORDER BY modifiers.
- **Pagination**: Uses standard `page` and `limit` query parameters.
- **Joins**: Uses configured relational metadata (`RelationConfig`) to automatically eager-load associations based on user input.

---

## 5. Schema Migration Engine

The Schema Migration Engine (`internal/services/migration_service.go`) ensures safe structural modifications to connected databases.

- **Capabilities**: Dynamically applies changes such as adding, renaming, dropping columns, and modifying column data types using GORM's `Migrator()`.
- **Safety**: Automatically records the state before executing DDLS, logs success/failure in the `migrations` table, and supports rolling back schemas if necessary.
- **Management API**: Accessible via `POST /api/v1/cms/migrations` and `POST /api/v1/cms/migrations/{id}/rollback`.

---

## 6. GraphQL Gateway

The system integrates a fully functional, optional **GraphQL Gateway** mapped at `/api/v1/graphql`.

- **Implementation**: Utilizes `github.com/99designs/gqlgen`.
- **Dynamic Generation**: Reads the internal metadata schema (Services, Fields, Relations) to construct generic GraphQL schemas allowing dynamic queries and mutations.
- **Playground**: A GraphQL playground is exposed at `/api/v1/graphql/playground` to provide a seamless development and testing environment.

---

## 7. Backup System

A comprehensive backup mechanism (`internal/services/backup_service.go`) guarantees data safety:

- **Features**: Supports structural schema backups, full data snapshots (exported as JSON), and record-level row counts.
- **State Management**: Asynchronous backup jobs are tracked in the `backups` table with standard states (`pending`, `completed`, `failed`).
- **Restoration**: Users can hit `POST /api/v1/cms/restore/{id}` to automatically deserialize and insert JSON snapshot data back into active service tables.

---

## 8. Observability Integration

Production readiness requires robust observability:

- **Logging**: Uses `go.uber.org/zap` for high-performance structured JSON logging. All HTTP requests are tagged with a unique `X-Correlation-ID`.
- **Metrics**: Exposes standard performance metrics on `/metrics` via Prometheus (`github.com/prometheus/client_golang`). Tracks latency, error rates, and request counts.
- **Tracing**: Integrates **OpenTelemetry** connected to **Jaeger**. Traces map the entire request lifecycle across HTTP handlers and internal database operations, enabling precise bottleneck identification.

---

## 9. Unit Tests

The platform ensures a robust foundation through comprehensive testing.

- **Scope**: Minimum 80% test coverage target. Covers internal services, data access patterns, and dynamic API handlers.
- **Implementation**: Utilizes Go's built-in `testing` library. Unit tests leverage an in-memory SQLite database setup (`DSN: :memory:`) to execute tests independently of an external database server.
- **Execution**: Run the full unit testing suite via:
  ```bash
  make test-unit
  # or standard Go
  go test -v -race ./tests/unit/...
  ```

---

## 10. Docker Setup

The platform is designed to be fully containerized. A multi-container Docker Compose setup handles everything needed for a local production-like environment:

- **Images**: The main API server builds into a minimal alpine image using the multi-stage `Dockerfile`.
- **Docker Compose**: The `docker-compose.yml` orchestrates the system, standing up the Go API server, a Redis instance for caching, a Postgres database for testing, Jaeger for tracing, and Prometheus for metrics.
- **Execution**:
  ```bash
  make docker-up
  # or
  docker-compose up -d
  ```

---

## 11. Swagger Documentation

All control plane and REST endpoints are heavily documented using **Swagger/OpenAPI 2.0**.

- **Implementation**: API configurations, payload models, and response codes are annotated in the Go source using `swaggo`.
- **Viewing Docs**: Accessible directly from the running server at `/api/v1/swagger/index.html`.
- **Generation**: To regenerate the documentation after code modifications, run:
  ```bash
  make swagger
  ```

---

This system is engineered for extensibility, robustness, and cloud-native deployments, delivering an out-of-the-box infrastructure backbone for modern digital experiences.