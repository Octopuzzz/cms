# Backend Platform (Dynamic CMS + API Builder)

Welcome to the **Self-Hosted, Go-Native Backend-as-a-Service Platform**.

This platform acts as a complete backend infrastructure generator, similar to Hasura or Supabase. It allows developers to dynamically create backend services, manage database connections, define schemas and relations, automatically generate CRUD APIs (REST and optionally GraphQL), handle database migrations, perform backups, and monitor logs, performance, and usage with integrated observability tools.

The platform is designed to be **modular, scalable, and cloud-ready**, leveraging a modern Go tech stack.

---

## 🏗️ System Architecture Explanation

The platform adheres to **Clean Architecture** and **Domain-Driven Design (DDD)** principles, structured into the following core layers:
- **Presentation Layer**: Handles REST APIs (via `Gin`) and GraphQL (via `gqlgen`).
- **Application Layer**: Contains business logic (`internal/services/`) for generating schemas, building APIs, handling queries, etc.
- **Domain Layer**: Houses the core models (`internal/models/`) corresponding to metadata configurations (Databases, Services, Fields, etc.).
- **Infrastructure Layer**: Controls database connectors, caches (Redis), logging (Zap), metrics (Prometheus), and tracing (OpenTelemetry).

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

## 🗄️ Metadata Database Schema

The CMS stores the platform metadata (which defines dynamic services, database connections, schemas, etc.) in a robust metadata database, natively structured in PostgreSQL (or SQLite for dev/testing).

Core metadata tables:
- **`database_connections`**: Stores external database configurations (id, name, type, host, port, credentials, connection pool settings).
- **`services`**: Represents dynamic data models. Includes properties like `id`, `name`, `database_connection_id`, `db_table_name`, `created_at`, `updated_at`.
- **`fields`**: Defines attributes for each service. Tracks configuration like string, integer, float, uuid, JSON types, as well as nullability, default values, and indexes.
- **`service_permissions`**: Maps RBAC (Roles to Services) indicating fine-grained permissions.
- **`migrations`**: Logs DDL execution history for schemas, aiding in history tracking and safe rollbacks.
- **`backups`**: Records generated snapshots and backups.
- **`users` and `roles`**: Manages auth and authorization for the control plane.
- **`audit_logs`**: Detailed logging of modifications to configurations and data.

---

## 📂 Go Project Structure

The project is structured modularly for maintainability and scalability:

```text
├── cmd
│   └── server                  # Entry point of the application
├── internal
│   ├── config                  # Environment variables & configurations
│   ├── database                # Connection managers, Database routing (PostgreSQL, MySQL, MongoDB)
│   ├── graphql                 # Gateway for optional GraphQL schema generation via gqlgen
│   ├── handlers                # Presentation Layer (Gin HTTP Routers & Endpoints)
│   ├── middleware              # Auth, Telemetry, Prometheus, Rate Limiting
│   ├── models                  # Domain Layer (Core Entities & Schema Definitions)
│   ├── services                # Application Layer (CMS, CRUD Engine, Schema, Backups)
│   └── tracing                 # OpenTelemetry logic & Jaeger integration
├── pkg
│   ├── logger                  # Global Zap structured logging
│   └── response                # Standardized HTTP API responses
├── tests
│   ├── unit                    # Unit tests isolated by layer
│   ├── integration             # Cross-layer integration tests
│   └── uat                     # User Acceptance & E2E flows
├── docs                        # Swagger Documentation (Generated)
├── docker                      # Docker build contexts & configs
├── docker-compose.yml          # Container orchestration (Redis, Postgres, Prometheus)
└── Makefile                    # Standardized tooling & tasks
```

---

## 🚀 CRUD Engine Implementation

The **Dynamic CRUD Engine** acts as the interface for operating against dynamically built services. Upon defining a `Service` with `Fields`, the engine exposes standardized REST interfaces.

**Supported Operations:**
- `POST /api/v1/data/{slug}` (Create)
- `GET /api/v1/data/{slug}/{id}` (Read)
- `PUT /api/v1/data/{slug}/{id}` (Update)
- `DELETE /api/v1/data/{slug}/{id}` (Delete)
- `GET /api/v1/data/{slug}` (List & Query)

**Query Engine Capabilities:**
- **Filtering:** `GET /api/v1/data/users?email=john@example.com`
- **Sorting:** `GET /api/v1/data/users?sort=created_at:desc`
- **Pagination:** `GET /api/v1/data/users?page=1&limit=20`
- **Join Queries:** `GET /api/v1/data/orders?join=user,products`

The engine validates the incoming payload against the stored schema and dynamically constructs GORM SQL operations while ensuring Role-Based Access Control and constraints.

---

## 🔄 Schema Migration Engine

The platform provides a **Schema Migration Engine** designed to handle safe diff operations (Add column, Drop column, Rename column, Change type).
- Modifying a `Service` schema automatically invokes GORM's `Migrator` against the target database connection.
- Before executing the migration, an automated backup/snapshot is initiated.
- Detailed migration history is recorded in the `migrations` metadata table natively, supporting potential rollbacks.

---

## 🌐 GraphQL Gateway (Optional)

The GraphQL gateway natively reads defined dynamic `Service` schemas and wraps them into a fully queryable GraphQL endpoint:
- Automatically handles queries and relationships via `gqlgen`.
- Users can query multiple dynamic entities simultaneously.
- Endpoints are exposed at `POST /api/v1/graphql` and respect identical authentication constraints as standard REST endpoints.

---

## 💾 Backup System

The **Backup Engine** guarantees data integrity through robust snapshots:
- Can execute JSON snapshot backups and SQL dumps for services.
- Accessible via endpoints such as `POST /api/v1/cms/backup/service/{service_id}`.
- Tracks generated backups via the metadata `backups` table, supporting point-in-time restores.

---

## 👁️ Observability Integration

Comprehensive production observability tools are heavily integrated:
- **Logging**: Highly structured and performant JSON logging via `Zap` (in `pkg/logger/`).
- **Metrics**: Standardized Prometheus metrics (`internal/middleware/prometheus.go`), tracking requests, response latency, and system status at `/metrics`.
- **Tracing**: Fully integrated Distributed Tracing utilizing `OpenTelemetry` (`internal/tracing/`), tracing SQL performance and deep request flows.

---

## 🧪 Unit Tests

The platform mandates stringent testing configurations:
- Tests encompass the Repository, Service, and API Handler layers.
- Validates the environment logic natively employing in-memory SQLite instances (`:memory:` via `github.com/glebarez/sqlite`) to decouple dependencies.
- Expected to surpass 80% minimum unit test coverage.
- To execute the comprehensive test suite and evaluate coverage, run:
  ```bash
  make test-coverage
  ```

---

## 🐳 Docker Setup

The system ensures seamless deployment natively utilizing Docker and `docker-compose`.
- **Core Platform (`Dockerfile`)**: Containerized standalone Go application configured for production (reduced image sizes).
- **Service Mesh (`docker-compose.yml`)**: Spawns necessary associated dependencies automatically: Redis (caching), PostgreSQL (metadata database), Jaeger (telemetry), and Prometheus (metrics).

To start locally:
```bash
make docker-up
```

---

## 📚 Swagger Documentation

The platform auto-generates robust API documentation covering CMS control capabilities and generic CRUD routes leveraging `Swaggo`.

To rebuild the OpenAPI specifications natively:
```bash
make swagger
```
The documentation outputs securely to the `docs/` folder (omitted from source control).
