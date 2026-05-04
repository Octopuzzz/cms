# Dynamic CMS + API Builder Platform

## 1. System Architecture Explanation

The Backend Platform acts as a backend infrastructure generator, similar to Hasura or Supabase. It allows developers to dynamically create backend services, APIs, schemas, and more. It is fully self-hosted and Go-native.

The architecture follows Clean Architecture and Domain-Driven Design (DDD):
- **Presentation Layer**: Handles incoming HTTP and GraphQL requests. Utilizes Gin for REST and gqlgen for GraphQL. Acts as an API Gateway.
- **Application Layer**: Contains business logic to orchestrate models, repositories, and various engines (Service Builder, CRUD Engine, Migration Engine).
- **Domain Layer**: Houses data models representing platform concepts like Services, Fields, Database Connections, and Users.
- **Infrastructure Layer**: Connects to the external world, such as Database Connectors (PostgreSQL, MySQL, MongoDB), Observability platforms (Prometheus, OpenTelemetry, Zap Logging), and Caching engines.

### High Level Architecture Flow

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

The internal CMS metadata is stored in a relational database (defaulting to SQLite or PostgreSQL) holding configuration for the dynamic systems:

- **`database_connections`**: Stores credentials and connection configuration for external user databases.
- **`services`**: Represents dynamic data models created via the Service Builder.
  - Fields include: `id`, `name`, `database_connection_id`, `db_table_name`, `created_at`, `updated_at`.
- **`fields`**: Represents attributes mapped to columns within a `service`.
  - Fields include: `id`, `service_id`, `name`, `type`, `nullable`, `unique`, `default_value`, `is_indexed`.
- **`service_permissions`**: Maps roles to services to handle RBAC (Role-Based Access Control).
- **`migrations`**: Logs migration states and schema diffs applied to dynamic services.
- **`backups`**: Stores references to JSON or SQL dumps for service snapshots.
- **`users`** & **`roles`**: Tracks platform administrators and user roles.

---

## 3. Go Project Structure

The project uses a standard, modular Go layout:

```text
.
├── cmd/
│   └── server/
│       └── main.go               # Application entrypoint
├── internal/
│   ├── api/                      # REST Handlers/Controllers
│   ├── config/                   # Configuration parsing
│   ├── database/                 # Connection manager and GORM instance logic
│   ├── graphql/                  # gqlgen implementation
│   ├── handlers/                 # HTTP routers
│   ├── middleware/               # Auth, Telemetry, and validation middleware
│   ├── models/                   # GORM entities
│   ├── services/                 # Core logic, CRUD Engine, Migration Engine, Service Builder
│   └── tracing/                  # OpenTelemetry configuration
├── pkg/
│   ├── errors/                   # Custom error wrappers
│   ├── logger/                   # Zap logger initialization
│   ├── pagination/               # Query pagination helpers
│   ├── response/                 # Standardized JSON response formatting
│   └── validation/               # Request validation helpers
├── tests/                        # Unit, Integration, and UAT test suites
├── docker/                       # Dockerfile, Compose configs
├── docs/                         # Generated Swagger definitions
├── Makefile                      # Make commands
├── .env.example                  # Example Environment Config
└── go.mod / go.sum
```

---

## 4. CRUD Engine Implementation

The **Dynamic CRUD Engine** acts as the core interpreter bridging HTTP paths to dynamic database queries.
When a user defines a Service, the platform automatically exposes standard endpoints:
- `POST /api/v1/data/{service_name}`
- `GET /api/v1/data/{service_name}`
- `GET /api/v1/data/{service_name}/{id}`
- `PUT /api/v1/data/{service_name}/{id}`
- `DELETE /api/v1/data/{service_name}/{id}`

The CRUD Engine leverages `database_connections` to route the query to the correct external database. It dynamically formats GORM statements based on incoming parameters, ensuring correct typing and security validation.

### Query Engine Capabilities
- **Filtering**: `GET /api/v1/data/users?email=test@example.com`
- **Sorting**: `GET /api/v1/data/users?sort=created_at:desc`
- **Pagination**: `GET /api/v1/data/users?page=1&limit=20`

---

## 5. Schema Migration Engine

The **Schema Migration Engine** guarantees database consistency when developers modify their dynamic services.

- **Automated DDL**: When adding or removing a field via the CMS Control Plane, the Migration Engine runs safe database alterations (e.g., `ADD COLUMN`, `DROP COLUMN`) utilizing GORM's `.Migrator()`.
- **Safety Checks**: Actions are recorded natively into the `migrations` metadata table, recording timestamp, service ID, and the exact operation (schema diff).

---

## 6. GraphQL Gateway

The platform auto-generates a unified GraphQL Gateway (`/api/v1/graphql`) using the `gqlgen` library.
It supports optional fallback query patterns without developers needing to manually define `.graphqls` files. The gateway resolves queries by introspecting the existing dynamic `services` and their corresponding `fields`, mapping GraphQL types directly to the CRUD engine's querying mechanism.

---

## 7. Backup System

The Backup system protects the developer's dynamic data directly through platform APIs:
- Provides mechanisms to snapshot an entire service's tables.
- Supports taking JSON backups.
- Keeps track of snapshots via the `backups` table.
- API endpoints allow triggering (`POST /api/v1/cms/backup/service/{id}`) and restoring data states.

---

## 8. Observability Integration

The platform integrates deep observability ensuring production readiness:
- **Logging**: Uses `go.uber.org/zap` for structured JSON logging. Request IDs and correlation IDs are injected into contexts.
- **Metrics**: Exposes a `/metrics` Prometheus endpoint mapping HTTP request latencies, status codes, and database interaction durations.
- **Tracing**: Fully integrated with OpenTelemetry. Telemetry contexts are passed through the HTTP middleware down to GORM database execution logs.

---

## 9. Unit Tests

Testing is enforced strictly, covering the Application (Services), Infrastructure (Database), and Presentation (Handlers) layers.
- Uses `testing` standard library combined with in-memory SQLite instances (`:memory:`) using the `github.com/glebarez/sqlite` driver for fast execution.
- Includes integration tests and user acceptance tests (UAT).
- Aimed at >= 80% coverage.

**Running tests:**
```bash
# Run unit tests
make test-unit

# Run full test suite with coverage
make test-coverage
```
*Note: Depending on environment restrictions, backgrounding tests or compiling with `GOTOOLCHAIN=local` might be required.*

---

## 10. Docker Setup

The system provides fully containerized environments using Docker and Docker Compose.
- **`Dockerfile`**: A multi-stage build that compiles the Go binary and packages a minimal alpine or distroless image.
- **`docker-compose.yml`**: Spins up the entire platform including the primary application, default PostgreSQL database, and necessary monitoring dependencies like Prometheus/Redis (if enabled).

**Running via Docker:**
```bash
docker-compose up --build -d
```

---

## 11. Swagger Documentation

API Documentation is auto-generated based on declarative annotations inside handler functions.
- Generates OpenAPI 3 specifications into the `docs/` folder.
- Available visually via Swagger UI paths if configured.

**Re-generating Swagger Docs:**
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
