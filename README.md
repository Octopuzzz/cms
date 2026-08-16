# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend-as-a-Service platform (Dynamic CMS + API Builder) written in Go. It operates similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native. The platform allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

The system is highly modular, scalable, and cloud-ready, adhering to Clean Architecture and Domain-Driven Design (DDD) principles.

---

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: The Gin HTTP Router and Handlers. Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (Services). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models. Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry, metrics (Prometheus), and logging.

### High-Level Architecture Diagram

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

The core internal configuration is stored in the **Metadata Database**. By default, it runs on SQLite for development and PostgreSQL for production.

Core tables:
1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks (e.g., CanCreate, CanRead). Contains row-level and field-level capabilities.
5. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. `backups`: Stores snapshot records or schema outputs.
7. `users` and `roles`: General authentication and authorization for the control plane.
8. `audit_logs`: Detailed logging of structural and data-level modifications.

---

## 3. Go Project Structure

The project uses a clean modular structure:

```text
.
├── cmd
│   └── server                # Application entrypoint
├── internal
│   ├── config                # Environment and configuration
│   ├── database              # Database connections and pool management
│   ├── graphql               # GraphQL schemas and resolvers
│   ├── handlers              # Gin HTTP handlers (REST API)
│   ├── middleware            # Authentication, rate limiting, telemetry
│   ├── models                # GORM models (domain layer)
│   ├── services              # Business logic (Service Builder, Migration, etc.)
│   └── tracing               # OpenTelemetry initialization
├── pkg
│   ├── logger                # Zap structured logging wrapper
│   └── response              # Standardized API response formatters
├── tests
│   ├── integration           # Integration tests
│   ├── uat                   # User Acceptance Testing
│   └── unit                  # Unit tests
├── docker                    # Dockerfiles and docker-compose configs
└── docs                      # Auto-generated Swagger documentation
```

---

## 4. CRUD Engine Implementation

When a service is created dynamically, the system automatically exposes RESTful endpoints for CRUD operations.
Managed by `DynamicDataService` mapping routes like `GET /api/v1/data/{slug}` to standard GORM operations on the fly.

Features:
- Automated filtering via query params (e.g., `?email=john@example.com`).
- Dynamic joining via foreign key mappings.
- Built-in pagination and sorting (e.g., `?sort=created_at:desc&page=1&limit=20`).
- Validates Role-Based Access Control and Row-Level filtering dynamically prior to queries.

---

## 5. Schema Migration Engine

Managed by `MigrationService`.
The system supports safe schema changes for dynamic services:
- Supports Add column, Drop column, Rename column, Change column type.
- Utilizes GORM's `.Migrator()`.
- Logs migration status natively into the `migrations` table and handles rollbacks through snapshot retention logic.
- Automatic backup execution before migration execution is supported.

---

## 6. GraphQL Gateway (Optional)

Managed by `Gateway` in the `graphql` package.
- Generates generic optional schemas using `github.com/99designs/gqlgen/graphql`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.
- Supports dynamically resolving relations.

---

## 7. Backup System

Managed by `BackupService`.
The platform supports backup and recovery functionalities:
- Generates data snapshots (table backup, schema backup, service snapshot).
- Output format: JSON snapshots (and SQL dumps).
- Implements generic service record backup features mapping dynamic table contents to snapshots.
- Supports restoring rows via JSON payload decoding.

---

## 8. Observability Integration

The system natively includes top-tier observability tools:
- **Logging**: Zap structured logging injected globally with correlation and trace IDs tied directly into Context.
- **Tracing**: OpenTelemetry/Jaeger initialized to wrap SQL commands and network logic.
- **Metrics**: Prometheus metrics track request latency, status codes, and error rates via standard HTTP middleware. Exposed at `/metrics`.

---

## 9. Unit Tests

Unit tests are written with `testing` (in-memory SQLite) ensuring components are decoupled from the infrastructure layer.

Test areas covered:
- Repository (Models)
- Service (Business logic, Dynamic schema creation, Backup, Migration)
- API handlers (Routing, middleware)

Coverage requirement: **>80%**

To run unit tests and ensure coverage across the whole project:
```bash
make test-unit
# or
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

---

## 10. Docker Setup

The system provides fully containerized environments using Docker and Docker Compose.
The provided `docker-compose.yml` configures:
- The Go Backend Server (API Builder / CMS)
- PostgreSQL (Metadata Database)
- Redis (Caching and rate-limiting)
- Prometheus (Metrics scraping)
- Jaeger / OpenTelemetry Collector (Tracing)

To start the platform:
```bash
docker-compose up -d
```

---

## 11. Swagger Documentation

API Documentation is auto-generated using standard OpenAPI 3.0 / Swagger mappings.
Annotations are added to all Gin handlers, which are then compiled by the `swag` CLI.

To generate or update Swagger documentation:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

The Swagger UI is exposed at `GET /swagger/index.html` when the server is running.
