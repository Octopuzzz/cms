# Dynamic CMS & API Builder

A self-hosted, production-grade Backend-as-a-Service platform written natively in Go. This platform dynamically creates backend services, defines data models, maps databases, and auto-generates REST (CRUD) and GraphQL APIs.

---

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles to remain modular and highly scalable.

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

- **Presentation Layer**: Exposes routes via Gin Framework (`internal/handlers/`). Handles REST traffic as well as GraphQL endpoint connections.
- **Application Layer**: Contains business rules (`internal/services/`), managing CMS logic, database validations, and dynamic queries.
- **Domain Layer**: Core data models representing metadata definitions (`internal/models/`).
- **Infrastructure Layer**: Metrics via Prometheus, logging via Zap, tracing via OpenTelemetry, and GORM database adapters.

---

## 2. Metadata Database Schema

The core state of the platform relies on a heavily relational metadata schema.

The tables include:

- `database_connections` - Stores configuration and credentials for external user databases.
- `services` - Represents a dynamically created backend schema or model definition.
- `fields` - Columns associated with a `service` (string, int, JSON, UUID, etc.)
- `relations` - Keeps track of connections between services (One-to-One, One-to-Many, etc.).
- `migrations` - Audit log and structural history for rolling forward and back database schemas.
- `backups` - Points to database snapshot records.
- `users` / `roles` - Internal Auth and RBAC definitions for Control Plane operations.

---

## 3. Go Project Structure

The project utilizes a standard Go modular layout:

```text
.
├── cmd/
│   └── server/          # Main entry points
├── internal/
│   ├── handlers/        # API route handlers (REST & Control Plane)
│   ├── services/        # Application and domain logic (CRUD engine, Schema builders)
│   ├── models/          # Core Metadata DB structures
│   ├── database/        # External connection pooling & initialization
│   ├── graphql/         # gqlgen dynamic resolving gateway
│   ├── tracing/         # OpenTelemetry configuration
│   ├── middleware/      # Auth, observability and prometheus interceptors
│   ├── config/          # Environment configuration
│   └── validation/      # Input and structure verifiers
├── pkg/
│   ├── logger/          # Structured Zap logging wrapper
│   ├── errors/          # Common application errors
│   └── response/        # Gin API standardized JSON response helpers
├── tests/               # Unit, Integration, and UAT specifications
├── docker/              # Deployment configurations and environments
├── docs/                # Generated Swagger Definitions
└── Makefile             # Standardization commands (build, test, dev)
```

---

## 4. CRUD Engine Implementation

The Dynamic CRUD Engine mapping REST commands to backend external or internal queries is managed in `internal/services/dynamic_data_service.go` and its respective handler.

- **Automated Mapping**: Takes URL patterns `GET /api/v1/data/{service_slug}` and applies them to specific mapped relational queries in the target metadata DB.
- **Query Engine**: Converts query parameters natively into GORM where clauses (`?email=user@domain.com`), handles Sorting (`?sort=created_at:desc`), and deals with structured Pagination constraints (`?page=1&limit=20`).

---

## 5. Schema Migration Engine

Located at `internal/services/migration_service.go`.

- **Safe Alterations**: Utilizes GORM’s Migrator to apply changes (Add Column, Drop Column, Change Types) automatically to external target tables when a user edits their `service` definition via the Control Plane.
- **Rollbacks**: Schema definitions natively hook into snapshots enabling database restoration.

---

## 6. GraphQL Gateway

Located at `internal/graphql/gateway.go`.

- By integrating `github.com/99designs/gqlgen/graphql`, the platform automatically compiles GraphQL APIs.
- Queries and mutations map back onto the standard Dynamic Service implementation, bridging GraphQL inputs into raw mapped SQL operations.

---

## 7. Backup System

Managed by `internal/services/backup_service.go`.

- **Platform Snapshots**: Performs entire schema extraction directly via JSON snapshot formatting.
- Facilitates backup and restore commands, enabling safe data migration across environments or rollback scenarios if users disrupt their dynamic DB layouts.

---

## 8. Observability Integration

Modern production readiness relies on rich observability:
- **Logging**: Configured `pkg/logger` implements Uber's **Zap** structured logger natively parsing traces.
- **Metrics**: Endpoints include Prometheus wrappers calculating HTTP durations and application layer timings (`internal/middleware/prometheus.go`).
- **Tracing**: Native **OpenTelemetry** hooks map directly into Context spans, tracking query depths.

---

## 9. Unit Testing

The system enforces rigorous automated checks with an 80%+ coverage mandate.

- Tests utilize an in-memory SQLite backend (DSN: `:memory:`) for fast iterative suite running.
- Includes Controller, Service, and Repository layer breakdowns mapped within `tests/unit/`.
- Run standard coverage using: `make test-coverage`.

---

## 10. Docker Setup

Fully containerized and Cloud Ready.

- `Dockerfile` utilizes multi-stage distroless compilation for minimal secure footprint execution.
- `docker-compose.yml` configures Postgres metadata bases alongside Redis caches and Jaeger monitoring nodes out of the box.

---

## 11. Swagger Documentation

API standardizations are generated automatically utilizing the `swaggo/swag` CLI.

- Run: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`
- Produces automated OpenAPI configurations exposed to clients wanting to interface directly with the Control Plane.
