# Backend Platform (Dynamic CMS + API Builder)

This repository contains a self-hosted, Go-native Backend-as-a-Service (BaaS) platform. The system acts as a dynamic infrastructure generator, allowing users to dynamically create backend services, schemas, and APIs.

## Features

- **Connect Databases:** PostgreSQL, MySQL, MongoDB, SQLite.
- **Dynamic Data Models:** Define tables, schemas, relations, and validations on the fly.
- **Auto-Generated APIs:** Generates full REST CRUD endpoints and optional GraphQL APIs automatically.
- **Schema Migrations:** Engine to manage automated and safe database schema changes.
- **Observability:** Built-in Prometheus metrics, OpenTelemetry tracing, and Zap structured logging.
- **Backup Engine:** Automated table, schema, and snapshot backups.
- **Role-Based Access Control (RBAC):** Fine-grained row and field-level permissions.

---

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer:** Gin HTTP Router and Handlers mapping REST requests and `gqlgen` for GraphQL endpoints.
- **Application Layer:** Business logic and service implementations orchestrating operations.
- **Domain Layer:** Core system models (`Service`, `Field`, `DatabaseConnection`, `User`, `Role`).
- **Infrastructure Layer:** Database access, OpenTelemetry tracing, Prometheus metrics, Zap logging, connection caching.

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

---

## 2. Metadata Database Schema

The CMS manages its own configuration internally via the Metadata Database. Key tables include:

1. `database_connections`: External DB configurations (id, name, type, host, credentials, pooling).
2. `services`: User-created data models referencing a DB connection and defining dynamic table names.
3. `fields`: Model attributes (string, integer, float, JSON, UUID, etc.) configuring unique/null constraints.
4. `relations`: Mappings between services (One-to-One, One-to-Many, Many-to-Many join tables).
5. `service_permissions`: Connects `Role` to `Service` with fine-grained access checks (e.g. CanCreate, CanRead, FieldLevelPerms, RowFilter).
6. `migrations`: DDL execution history and rollbacks.
7. `backups`: System schema and data snapshot tracking.
8. `users` and `roles`: General platform authentication and authorization.
9. `audit_logs`: Detailed logging of schema and structural operations.

---

## 3. Go Project Structure

The project uses a clean and modular layout:

```text
├── cmd
│   └── server                # Application entrypoint
├── internal
│   ├── config                # Configuration loading
│   ├── database              # Database Connection Manager
│   ├── graphql               # GraphQL Gateway generator
│   ├── handlers              # API Gateway (REST HTTP)
│   ├── middleware            # Auth, Prometheus, Rate Limiting
│   ├── models                # Domain Metadata Schema definitions
│   ├── services              # Control Plane, CRUD Engine, Schema Migration, Backup
│   └── tracing               # OpenTelemetry integration
├── pkg
│   ├── logger                # Zap structured logging
│   └── response              # API response formatting
├── tests
│   ├── integration           # Inter-service integration tests
│   ├── uat                   # E2E / UAT tests
│   └── unit                  # Unit tests (Mocking/SQLite in-memory)
├── docker
│   ├── Dockerfile            # Production container build
│   └── docker-compose.yml    # Full-stack orchestrator
├── ARCHITECTURE.md           # Core architecture design
├── Makefile                  # Build, Test, Docker tasks
└── README.md                 # Project documentation
```

---

## 4. Dynamic CRUD Engine

Managed by `internal/services/dynamic_data_service.go`, the CRUD Engine automatically maps dynamic service endpoints to standard GORM operations.

- **Endpoints Generated:**
  - `POST /api/v1/data/{slug}`
  - `GET /api/v1/data/{slug}/{id}`
  - `PUT /api/v1/data/{slug}/{id}`
  - `DELETE /api/v1/data/{slug}/{id}`
  - `GET /api/v1/data/{slug}`

- **Features:** Automated filtering (e.g., `?email=user@domain.com`), sorting (`?sort=created_at:desc`), pagination (`?page=1&limit=20`), and dynamic joining via relational configuration. It natively verifies Role-Based Access Control and Row-Level filtering.

---

## 5. Schema Migration Engine

Managed by `internal/services/migration_service.go`.

- **Operations:** Safe schema diffs for adding, dropping, renaming columns, and changing column types using GORM’s Migrator.
- **Safety Features:** Logs migration status (`pending`, `applied`, `failed`, `rolled_back`) natively into the `migrations` metadata table. Retains a snapshot of the schema for seamless rollback capabilities.

---

## 6. GraphQL Gateway

Managed by `internal/graphql/gateway.go`.

- Uses `gqlgen` to automatically expose optional endpoints for dynamically defined models.
- Generates fully-functional queries and mutations reflecting relational setups.
- Available at: `POST /api/v1/graphql`.

---

## 7. Backup System

Managed by `internal/services/backup_service.go`.

- Generates data snapshots. Currently implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.
- Status tracked natively in the `backups` metadata table.

---

## 8. Observability Integration

Implemented systematically across the platform infrastructure:

- **Logging:** Request, error, and query logs natively processed using **Zap** (`pkg/logger/`), enriched with correlation and trace IDs.
- **Metrics:** **Prometheus** tracks request latency, slow queries, error rates, and API throughput (`internal/middleware/prometheus.go`). Exported at `/metrics`.
- **Tracing:** **OpenTelemetry/Jaeger** wraps SQL commands, networking logic, and inter-component transitions (`internal/tracing/`).

---

## 9. Unit Testing

The platform enforces high testing standards with at least 80% coverage.

- **Test Areas:** Repository layer, Service implementations, and API Handler routing.
- **Execution:** Run full tests with `make test`. Run with coverage using `make test-coverage`.
- Standard isolated in-memory testing leverages `github.com/glebarez/sqlite`.

---

## 10. Docker Setup

The system is fully containerized for deployment.

- `docker-compose up -d` spins up the Go Backend, primary database configurations, and optional Prometheus metrics instances.
- Optimized multi-stage `Dockerfile` handles environment bundling for production scalability.

---

## 11. Swagger Documentation

API documentation follows the OpenAPI specification.
Generate Swagger configuration automatically:

```bash
make swagger
# Or manually using:
# swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
