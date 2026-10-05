# Dynamic CMS & API Builder (Backend Platform)

A fully self-hosted, Go-native Backend-as-a-Service (BaaS) platform that allows developers to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints. This platform works as a backend infrastructure generator (similar to Hasura or Supabase) and is built using Clean Architecture and Domain-Driven Design (DDD).

## Features
- **Dynamic Data Models**: Connect databases and create data models dynamically.
- **Auto-generated APIs**: Generate CRUD REST APIs and optional GraphQL APIs automatically.
- **Schema Migrations**: Manage schema changes safely with an integrated migration engine.
- **Observability**: Built-in structured logging (Zap), metrics (Prometheus), and tracing (OpenTelemetry).
- **Backups**: Snapshot and restore services safely.
- **Scalable and Cloud-ready**: Modular design with Docker containerization support.

---

## 1. System Architecture Explanation

The system is designed following **Clean Architecture** and **Domain-Driven Design (DDD)**. The high-level architecture handles dynamic routing, dynamic schema generation, and operations over connected databases.

### Core Layers:
- **Presentation Layer**: Gin HTTP Router and Handlers (`internal/handlers/`), and the GraphQL Gateway. Acts as the API gateway mapping external requests to internal services.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, manage dynamic schema changes, and handle the logic for dynamic CRUD and queries.
- **Domain Layer**: The data models (`internal/models/`). Defines the system's core metadata entities (Services, Fields, Relations, Database Connections, Roles, Users).
- **Infrastructure Layer**: Cross-cutting tools including tracing (OpenTelemetry), metrics (Prometheus), structured logging (Zap), and connection pooling.

### High-Level Architecture Flow:
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

The CMS stores all platform metadata required to dynamically manage your services.

**Recommended Database**: PostgreSQL

### Core Metadata Tables:
- `database_connections`: Stores external database configurations (id, name, type, host, port, username, password, etc.). Manages connection pooling.
- `services`: Represents user-created data models (id, name, database_connection_id, db_table_name, created_at, updated_at).
- `fields`: Defines attributes for each service (id, name, type, nullable, unique, default_value, index). Supports types like string, integer, float, boolean, uuid, json, array, timestamp.
- `service_permissions`: Links Roles to Services for fine-grained, row-level, and field-level access control.
- `relations`: Defines one-to-one, one-to-many, many-to-one, and many-to-many relationships (including automatic join tables).
- `migrations`: Tracks DDL execution history and rollbacks.
- `backups`: Stores backup snapshot metadata.
- `users` / `roles`: For control plane authentication.

---

## 3. Go Project Structure

The project follows a clean, modular layout.

```text
.
├── cmd/
│   └── server/          # Main application entry point
├── docker/              # Docker configuration files (Prometheus, Dockerfile, etc.)
├── internal/
│   ├── config/          # Application configuration mapping
│   ├── database/        # Database connection managers
│   ├── graphql/         # GraphQL schema generation and gateway
│   ├── handlers/        # Gin HTTP Handlers (Presentation Layer)
│   ├── middleware/      # Auth, Logging, Metrics, Tracing middlewares
│   ├── models/          # Core Domain models
│   ├── services/        # Application Business Logic (CRUD, Builder, etc.)
│   └── tracing/         # OpenTelemetry tracing setup
├── pkg/
│   ├── logger/          # Structured Zap logging wrapper
│   └── response/        # Standardized API response wrappers
└── tests/
    ├── integration/     # Integration tests
    ├── uat/             # User Acceptance tests
    └── unit/            # Unit tests
```

---

## 4. CRUD Engine Implementation

When a new service (data model) is created via the Service Builder, the platform automatically generates RESTful CRUD endpoints. The **Dynamic Data Service** (`internal/services/dynamic_data_service.go`) handles routing these requests to standard GORM operations on the fly.

**Supported Operations & Example Endpoints:**
- **Create**: `POST /api/v1/data/{service_slug}`
- **Read**: `GET /api/v1/data/{service_slug}/{id}`
- **Update**: `PUT /api/v1/data/{service_slug}/{id}`
- **Delete**: `DELETE /api/v1/data/{service_slug}/{id}`
- **List / Query Engine**: `GET /api/v1/data/{service_slug}`

The **Query Engine** extends the "List" operation, supporting:
- Filtering: `?email=john@example.com`
- Sorting: `?sort=created_at:desc`
- Pagination: `?page=1&limit=20`
- Joining: `?join=user,products` (automatically maps foreign key relationships and supports nested joins).

---

## 5. Schema Migration Engine

The platform includes a Schema Migration Engine (`internal/services/migration_service.go`) that ensures safe transitions between service schema states using GORM's auto-migrator.

**Features:**
- Add, Drop, Rename, and Change column types.
- Automatic snapshot/backup capabilities prior to executing structural schema migrations.
- Rollback support in case of failure.
- Migration execution histories are logged in the `migrations` metadata table.

---

## 6. GraphQL Gateway (Optional)

The CMS features a dynamic GraphQL API powered by `gqlgen` (`internal/graphql/gateway.go`).

- Upon service creation, generic GraphQL schemas are compiled automatically.
- Accessible via the gateway: `POST /api/v1/graphql`
- The GraphQL API supports comprehensive queries, mutations, and automatic resolution of relational data between services.

---

## 7. Backup System

A comprehensive backup mechanism (`internal/services/backup_service.go`) guarantees data safety.

**Features:**
- Create data snapshots of specific service records in JSON formats.
- Triggered dynamically via endpoints, e.g., `POST /api/v1/cms/backup/service/{service_id}`.
- Restore rows via payload decoding using endpoints, e.g., `POST /api/v1/cms/restore/service/{service_id}`.

---

## 8. Observability Integration

Production-grade observability is woven into the application layers:

- **Logging**: Zap structured logging is injected globally (`pkg/logger/`), capturing request logs, errors, and traces directly tied into the Context.
- **Metrics**: A Prometheus middleware (`internal/middleware/prometheus.go`) tracks request latency, status codes, and error rates. Exported natively at `/metrics`.
- **Tracing**: OpenTelemetry (`internal/tracing/`) initialization wraps database SQL commands and network operations to detect slow queries and pinpoint bottlenecks.

---

## 9. Unit Tests

Unit tests are heavily emphasized across the Repository, Service, and API handler layers to ensure a minimum of **80% coverage**.

**Commands to run tests:**
```bash
# Run unit tests locally
make test-unit

# Run full test suite with coverage
make test-coverage
```
Note: Ensure you include `-coverpkg=./...` when tracking coverage across isolated test folders.

---

## 10. Docker Setup

The platform is completely containerized for local development and cloud deployment.

- A multi-stage `Dockerfile` handles optimized Go binary compilation.
- `docker-compose.yml` orchestrates the CMS Core, alongside sidecar dependencies (like PostgreSQL metadata database, Redis caching, Prometheus metrics exporter).
- Build the container: `make build` or standard Docker build commands.

---

## 11. Swagger Documentation

Interactive OpenAPI / Swagger documentation is generated using `swag`.

**Generation Command:**
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

*(Note: The generated `docs/` folder is intentionally ignored by version control to avoid polluting the repo with build artifacts.)*
