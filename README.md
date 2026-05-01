# CMS Backend Platform (Dynamic CMS + API Builder)

A production-grade, self-hosted, and Go-native Backend-as-a-Service (BaaS) platform. This platform allows developers to connect external databases, create data models dynamically, generate CRUD and optional GraphQL APIs automatically, and monitor system performance with state-of-the-art observability.

---

## 1. System Architecture Explanation

The system follows **Clean Architecture** and **Domain Driven Design (DDD)** principles, separating concerns into presentation, application, domain, and infrastructure layers.

### High-Level Architecture
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

- **Presentation Layer**: Handlers and routers mapped using Gin (REST) and `gqlgen` (GraphQL).
- **Application Layer**: Business logic defined in core services (e.g., dynamic schema generation, metadata CRUD).
- **Domain Layer**: Core data models (Services, Fields, DB Connections).
- **Infrastructure Layer**: Cross-cutting concerns like logging, tracing, metrics, caching, and connection pooling.

---

## 2. Metadata Database Schema

The platform stores all dynamic metadata internally. The core schemas (often stored in PostgreSQL or an equivalent relational database) include:

- **`database_connections`**: Stores external database setups (id, name, type, host, port, username, password, database_name).
- **`services`**: Represents dynamic data models (id, name, db_table_name, database_connection_id, created_at, updated_at).
- **`fields`**: Attributes for each service (id, service_id, name, type, nullable, unique, default_value, index).
- **`relations` / `service_permissions`**: Maps one-to-one, one-to-many, many-to-one, and many-to-many relations dynamically, and maps user roles to fine-grained access constraints.
- **`migrations`**: Audit log of schema changes (add, drop, rename columns).
- **`backups`**: Metadata pointing to stored backups/snapshots.
- **`users` & `roles`**: RBAC system to manage access to the Control Plane.

---

## 3. Go Project Structure

A clean, modular layout standardizing Go service projects:

```
.
├── cmd
│   └── server          # Application entry point (main.go)
├── docker              # Dockerfile and compose configurations
├── internal
│   ├── api             # API routes and handlers (handlers/)
│   ├── config          # Application configuration loading
│   ├── database        # DB connection pooling and external DB management
│   ├── graphql         # Auto-generated GraphQL Gateway
│   ├── middleware      # Gin middleware (Auth, CORS, Rate Limit, Metrics)
│   ├── models          # Core Domain structures
│   ├── services        # Application business logic (Service Builder, Migrations, etc.)
│   └── tracing         # OpenTelemetry / Jaeger integration
├── pkg
│   ├── logger          # Zap structured logging wrapper
│   └── response        # Standardized HTTP response structures
├── tests
│   ├── integration     # Service integration tests
│   ├── uat             # End-to-end / User Acceptance Testing
│   └── unit            # Unit tests for domain and application layers
├── ARCHITECTURE.md     # In-depth architectural details
├── Makefile            # Automation for building, testing, linting
├── docker-compose.yml  # Local stack bringing up App + DBs + Observability
└── go.mod              # Go dependencies
```

---

## 4. CRUD Engine Implementation

When a Service is created, the system maps REST endpoints using a `DynamicDataService` that automatically binds generic CRUD operations against the registered database via GORM.

- **Create**: `POST /api/v1/data/{slug}`
- **Read**: `GET /api/v1/data/{slug}/{id}`
- **Update**: `PUT /api/v1/data/{slug}/{id}`
- **Delete**: `DELETE /api/v1/data/{slug}/{id}`
- **List / Query**: `GET /api/v1/data/{slug}`

### Query Engine Features:
- **Filtering**: `?email=john@example.com`
- **Sorting**: `?sort=created_at:desc`
- **Pagination**: `?page=1&limit=20`

---

## 5. Schema Migration Engine

Tracked via `internal/services/migration_service.go`, the system allows safe DDL changes.
- **Features**: Add, drop, or rename columns without breaking the metadata integrity.
- **Safety Features**: Operations leverage GORM's `.Migrator()`. Changes are recorded in the `migrations` table, allowing administrators to audit or trigger snapshots before complex rollbacks.

---

## 6. GraphQL Gateway

Located under `internal/graphql/`, the platform provides a dynamically generated GraphQL schema utilizing `github.com/99designs/gqlgen`.

- **Endpoint**: `POST /api/v1/graphql`
- **Capabilities**: Translates dynamic Service definitions into GraphQL types, providing automatic Query and Mutation generation out of the box.

---

## 7. Backup System

Managed by `internal/services/backup_service.go`, users can generate point-in-time snapshots of their service data.

- **Endpoints**:
  - `POST /api/v1/cms/backup/service/{service_id}` - Triggers a snapshot.
  - `POST /api/v1/cms/restore/{id}` - Restores from a saved snapshot.
- **Format**: JSON structured snapshots mapped to generic table contents.

---

## 8. Observability Integration

The platform provides complete tracking out of the box:

- **Logging**: Global structured logging using `go.uber.org/zap` (`pkg/logger/`). Request IDs and context correlate logs across layers.
- **Metrics**: Exposes Prometheus metrics at `GET /metrics` utilizing standard `client_golang/prometheus/promhttp`. It tracks latency, error rates, and request throughput.
- **Tracing**: OpenTelemetry (`internal/tracing/`) is set up to trace requests starting from the API gateway down to external database calls via Jaeger.

---

## 9. Unit Tests

Unit testing focuses on Repositories, Services, and API Handlers. Minimum required coverage across core layers is **80%**.

**Command to run tests:**
```bash
make test-unit
# or
go test -v -race ./tests/unit/...
```

**Full test suite coverage:**
```bash
make test-coverage
```

Unit testing uses in-memory SQLite (`github.com/glebarez/sqlite`) for efficient mocking.

---

## 10. Docker Setup

A complete `docker-compose.yml` provides an out-of-the-box environment. It starts up the Go binary alongside observability tools and a default metadata database.

**To run the stack:**
```bash
make docker-up
```

---

## 11. Swagger Documentation

API documentation is generated automatically using `swaggo`.

- **Command to Generate**:
  ```bash
  make swagger
  ```
- **Access UI**: Visit `GET /swagger/index.html` after the server starts. Provides fully interactive documentation defining endpoints, parameters, responses, and API Keys for standard operations.
