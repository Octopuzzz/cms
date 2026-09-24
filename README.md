# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade, self-hosted, Go-native Backend-as-a-Service platform. It functions similarly to platforms like Hasura or Supabase, allowing developers to dynamically create backend services, manage database connections, define schemas, and automatically generate CRUD REST endpoints and optional GraphQL APIs.

The system is built to be modular, scalable, and cloud-ready, adhering to Clean Architecture and Domain-Driven Design (DDD) principles.

---

## 1. System Architecture Explanation

The Backend Platform operates as an API Gateway and Backend Infrastructure Generator.

### High-Level Architecture

```text
       Client Applications
                │
   ┌────────────┴─────────────┐
   ▼                          ▼
REST API                 GraphQL API (optional)
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

### Core Layers (Clean Architecture)

- **Presentation Layer**: Handles incoming HTTP requests via the Gin web framework (`internal/handlers`). Maps to internal services. Includes GraphQL gateway (`internal/graphql`).
- **Application Layer**: Contains business logic (`internal/services`). Interfaces with engines for schema migrations, dynamic querying, and backups.
- **Domain Layer**: Core metadata models (`internal/models`) defining `DatabaseConnection`, `Service`, `Field`, `User`, `Role`, etc.
- **Infrastructure Layer**: Cross-cutting concerns like logging (Zap), metrics (Prometheus), tracing (OpenTelemetry), and database connection caching (`internal/database`).

---

## 2. Metadata Database Schema

The platform relies on a metadata database (recommended: PostgreSQL) to store configuration. The core tables (defined in `internal/models/models.go`) include:

- `database_connections`: Stores configurations for external databases (PostgreSQL, MySQL, MongoDB). Fields include `id`, `name`, `type`, `host`, `port`, `username`, `password`, `database`, etc.
- `services`: Represents dynamically created data models. Links to `database_connections` and tracks `db_table_name`.
- `fields`: Defines attributes for services. Fields include `name`, `type` (e.g., string, integer, json, uuid), `is_nullable`, `is_unique`, `default_value`, etc.
- `role_permissions` & `service_permissions`: Manage Role-Based Access Control (RBAC) and row/field-level permissions.
- `migrations`: Tracks schema migrations applied to dynamic services, enabling history and rollback capabilities.
- `backups`: Stores snapshot records and schema outputs.
- `users` and `roles`: Controls authentication and authorization for the CMS Control Plane.
- `audit_logs`: Detailed logging of structural and data modifications.

---

## 3. Go Project Structure

The project follows a standard Go project layout, optimized for modularity:

```
├── cmd
│   └── server          # Entry point for the application (main.go)
├── internal
│   ├── api             # API routing, handlers, and request mapping
│   ├── config          # Environment configuration loading
│   ├── contextutils    # Typed context keys
│   ├── database        # Database Connection Manager & pooling
│   ├── graphql         # GraphQL Gateway implementation (gqlgen)
│   ├── middleware      # Gin middlewares (Auth, Rate Limit, Prometheus)
│   ├── models          # Domain Layer: Metadata database schema
│   ├── services        # Application Layer: Core business logic
│   └── tracing         # OpenTelemetry tracing setup
├── pkg
│   ├── errors          # Custom error types and handling
│   ├── logger          # Structured logging setup (Zap)
│   ├── pagination      # Pagination utilities
│   └── response        # Standardized HTTP response formatting
├── tests
│   ├── integration     # Integration tests
│   ├── uat             # User Acceptance Testing
│   └── unit            # Unit tests for repositories, services, handlers
├── docker
│   ├── Dockerfile      # Containerization instructions
│   └── docker-compose.yml # Local deployment setup
├── docs                # Auto-generated Swagger documentation (gitignored)
├── Makefile            # Standardized development tasks
└── go.mod / go.sum     # Dependency management
```

---

## 4. Platform Components

### 4.1. CMS Control Plane & Service Builder
Managed via `/api/v1/cms/` endpoints.
- Allows registering external databases (`POST /cms/databases`).
- Enables dynamic creation of services/models (`POST /cms/services`) and fields.
- Implemented via `internal/services/service_service.go` and `internal/services/dbconn_service.go`.

### 4.2. Database Connection Manager
Located in `internal/database/connection_manager.go`.
- Supports PostgreSQL, MySQL, and MongoDB.
- Implements connection pooling (`max_open_conns`, `max_idle_conns`, `conn_max_lifetime`).

### 4.3. Dynamic CRUD Engine & Query Engine
Located in `internal/services/dynamic_data_service.go`.
- Automatically generates CRUD endpoints when a service is created (`GET /api/{service}`, `POST /api/{service}`, etc.).
- Supports advanced querying including filtering (`?email=...`), sorting (`?sort=...`), and pagination (`?page=1&limit=20`).
- Built-in protection against N+1 query problems via batch inserting.

### 4.4. Schema Migration Engine
Located in `internal/services/migration_service.go`.
- Tracks and applies safe schema changes (Add/Drop/Rename Column) using GORM's `Migrator()`.
- Logs status to the `migrations` table.

### 4.5. GraphQL Gateway (Optional)
Located in `internal/graphql/gateway.go`.
- Exposes `POST /api/v1/graphql`.
- Designed to dynamically generate GraphQL AST from service schemas and resolve queries.

### 4.6. Backup System
Located in `internal/services/backup_service.go`.
- Handles data snapshotting, schema backup, and generic service record backups.
- Supports restoring rows via JSON payload decoding.

### 4.7. Observability Integration
Integrated across the entire stack:
- **Logging**: Zap structured logging (`pkg/logger`) with correlation IDs.
- **Metrics**: Prometheus tracking request latency and error rates (`internal/middleware/prometheus.go`).
- **Tracing**: OpenTelemetry wrapping SQL commands and network logic (`internal/tracing/`).

### 4.8. Security & Performance
- JWT Authentication and fine-grained RBAC.
- Rate limiting to protect APIs.
- Input validation and SQL injection protection via GORM parameterization.
- Redis caching (optional) and database connection pooling.

---

## 5. Development & Testing

### Docker Setup
A complete Dockerized environment is provided in the `docker` directory, including the backend service, PostgreSQL (metadata), Redis, and Prometheus.

### Swagger Documentation
API documentation is automatically generated using Swagger.
To generate: `go install github.com/swaggo/swag/cmd/swag@latest` then `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.

### Unit Testing
The system includes comprehensive tests with an 80%+ coverage target across Repository, Service, and API handler layers.

**Run tests:**
```bash
make test-coverage
```
*Note: Unit tests use an in-memory SQLite database (`:memory:`) via `github.com/glebarez/sqlite` to mock DB connections.*
