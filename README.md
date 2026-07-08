# Dynamic CMS + API Builder (Go-Native BaaS Platform)

Welcome to the self-hosted **Backend-as-a-Service (BaaS) Platform**, a robust, production-grade Go-native alternative to platforms like Hasura and Supabase. This platform dynamically generates backend services, REST and optional GraphQL APIs, manages schema migrations, and integrates comprehensive observability, backups, and security.

This project follows **Clean Architecture** and **Domain-Driven Design (DDD)**.

## 1. System Architecture Explanation

The platform architecture is designed to decouple presentation, business logic, domain models, and external infrastructure.

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

- **Presentation Layer**: Handles incoming HTTP and GraphQL requests (Gin router, `gqlgen`).
- **Application Layer**: Contains business logic, orchestrates workflows, manages the dynamic CRUD engine and migrations.
- **Domain Layer**: Contains core models and entities.
- **Infrastructure Layer**: Connects to the database, caching (Redis), logging (Zap), and telemetry (OpenTelemetry, Prometheus).

## 2. Metadata Database Schema

The CMS manages internal states and metadata representing the configured services using a relational model. Below is the metadata database schema structure typically managed inside PostgreSQL or SQLite:

* **`databases`** (Database Connections)
  * `id`, `name`, `type`, `host`, `port`, `username`, `password`, `database_name`
* **`services`** (Data Models)
  * `id`, `name`, `database_id`, `db_table_name`, `created_at`, `updated_at`
* **`fields`** (Service Attributes)
  * `id`, `service_id`, `name`, `type`, `nullable`, `unique`, `default_value`, `index`
* **`relations`**
  * Tracks relationships (One-to-One, One-to-Many, Many-to-Many) between services.
* **`service_permissions`**
  * Connects `Role` to `Service` for RBAC.
* **`migrations`**
  * Tracks schema DDL changes for rollback and versioning.
* **`backups`**
  * Records snapshot statuses.
* **`users`** & **`roles`**
  * System authentication and authorization.

## 3. Go Project Structure

The project implements a clean modular structure to ensure scalability:

```text
.
├── cmd/
│   └── server/             # Main application entrypoint
├── docker/                 # Container and Compose configurations
├── internal/
│   ├── api/                # Interfaces & API controllers (if abstracted)
│   ├── config/             # Configuration management (.env)
│   ├── database/           # DB Connection management, caching
│   ├── graphql/            # GraphQL Gateway (gqlgen)
│   ├── handlers/           # HTTP handlers / REST presentation
│   ├── middleware/         # Auth, Prometheus metrics, Tracing, Rate-limits
│   ├── models/             # Domain entities (Service, Field, User, etc.)
│   ├── services/           # Application business logic (CRUD Engine, CMS, Backups)
│   └── tracing/            # OpenTelemetry integration
├── pkg/
│   ├── logger/             # Zap logger initialization
│   ├── response/           # Standardized API responses
│   └── errors/             # Common error types
├── tests/
│   ├── unit/               # Isolated logic tests
│   ├── integration/        # Database-dependent tests
│   └── uat/                # End-to-End user acceptance tests
├── ARCHITECTURE.md         # Detailed architectural specifications
├── Makefile                # Build, Dev, and Test automation commands
└── Dockerfile              # Docker build instructions
```

## 4. Components

### CRUD Engine Implementation
Dynamically handles RESTful operations.
- `POST /api/{service}`
- `GET /api/{service}/{id}`
- `PUT /api/{service}/{id}`
- `DELETE /api/{service}/{id}`
- `GET /api/{service}`

### Query Engine
Supports filtering (`?email=...`), sorting (`?sort=created_at:desc`), pagination (`?page=1&limit=20`), and joins.

### Schema Migration Engine
Tracks metadata and securely applies DDL changes (Add, Drop, Rename Column) using GORM’s Migrator. Supports snapshot-based rollbacks.

### GraphQL Gateway
Built using `gqlgen`, generating dynamic queries and mutations mapped back to defined services and relations.

### Backup System
Supports snapshot creation and restoration. Generates structured JSON representations of dynamic table contents for simple storage.

### Observability Integration
* **Logging**: Structured Zap logs.
* **Metrics**: Prometheus middleware tracking latency and request rates.
* **Tracing**: OpenTelemetry wrapping SQL execution and request lifecycles.

## 5. Security & Performance

* **Security**: Enforces strict input validation, protects against SQL injections (using parameterized queries in GORM), utilizes JWT authentication, and implements rate limiting.
* **Performance**: Connection pooling natively managed via GORM, optional query caching, and pagination optimization.

## 6. Testing

The platform enforces high testing standards with a minimum of 80% coverage across Repository, Service, and API handlers.
Run unit tests via:

```bash
make test-unit
# or
go test ./tests/unit/...
```

Run complete test suite with coverage:
```bash
make test-coverage
```

## 7. Setup & Docker

To run the platform via Docker:
```bash
docker-compose up -d
```
The application will expose REST on port `:8080` (or specified in `.env`) and metrics on `:9090`.

## 8. API Documentation

Swagger / OpenAPI documentation is supported. Generate docs using:
```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
