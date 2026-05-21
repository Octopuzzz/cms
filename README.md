# Go-native Backend Platform (Dynamic CMS + API Builder)

This is a self-hosted, production-grade Backend-as-a-Service (BaaS) platform written in Go. It works similarly to platforms like Hasura or Supabase, allowing developers to dynamically create backend services, manage schemas, automatically generate CRUD REST APIs, and optionally generate GraphQL endpoints.

The system is fully modular, scalable, cloud-ready, and adheres to Clean Architecture and Domain-Driven Design (DDD) principles.

---

## 1. System Architecture Explanation

The platform architecture connects your client applications (web, mobile, or third-party APIs) directly to database resources via an API Gateway and a dynamic set of backend engines.

### High-Level Platform Architecture

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

### Core Layers
- **Presentation Layer**: Exposes both REST endpoints (via Gin) and GraphQL endpoints (via gqlgen). Maps external requests to internal domains.
- **Application Layer**: Contains engines (CRUD, Query, Migrations, Backups) and services coordinating the platform's features.
- **Domain Layer**: Holds the core logic and models for dynamic services, fields, relations, and metadata.
- **Infrastructure Layer**: Manages the underlying database connections, observability components (Prometheus, OpenTelemetry, Zap logging), caching, and security enforcement.

---

## 2. Metadata Database Schema

The platform relies on a metadata database (typically PostgreSQL or SQLite for local development) to store configuration for dynamically generated services.

### Core Tables

- `database_connections`: Stores external database configurations, including connection types (Postgres, MySQL, Mongo), host, port, credentials, and pooling info.
- `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks the actual database table name.
- `fields`: Defines attributes for each service (name, type like string/integer/uuid, nullable, unique, default_value, indexing).
- `relations`: Defines one-to-one, one-to-many, many-to-one, and many-to-many relationships (including auto-generated join tables).
- `migrations`: Keeps track of applied schema changes and stores rollback snapshots.
- `backups`: Stores backup execution records and file references for table, schema, and snapshot backups.

---

## 3. Go Project Structure

The project employs a clean and modular structure:

```text
.
├── cmd/
│   └── server/             # Application entrypoint (main.go)
├── internal/
│   ├── config/             # Environment & app configuration
│   ├── database/           # Database connection manager & pooling
│   ├── graphql/            # GraphQL gateway and auto-generator (gqlgen)
│   ├── handlers/           # HTTP handlers & API gateway logic (REST API)
│   ├── middleware/         # Auth, rate limiting, logging, metrics middlewares
│   ├── models/             # Domain entities (Service, Field, User, Role, etc.)
│   ├── services/           # Application layer (CMS, CRUD Engine, Query Engine, Backup Engine, Schema Migrations)
│   └── tracing/            # OpenTelemetry setups
├── pkg/                    # Reusable packages
│   ├── errors/
│   ├── logger/             # Zap structured logging wrapper
│   ├── pagination/
│   └── validation/
├── tests/                  # Unit, Integration, and UAT test suites
├── docker/                 # Dockerfile and compose setups
├── go.mod                  # Go module dependencies
└── Makefile                # Task definitions (build, dev, test)
```

---

## 4. CRUD Engine Implementation

When a user creates a new service, the dynamic CRUD Engine automatically sets up endpoints utilizing the mapped database.

**Operations Handled Automatically:**
- **Create**: `POST /api/v1/data/{service}`
- **Read**: `GET /api/v1/data/{service}/{id}`
- **Update**: `PUT /api/v1/data/{service}/{id}`
- **Delete**: `DELETE /api/v1/data/{service}/{id}`
- **List**: `GET /api/v1/data/{service}`

**Implementation Details:**
- Utilizes GORM to dynamically map REST payloads into database inserts/updates using dynamic maps or `interface{}`.
- Handles validations, constraints checking, and transaction safety natively based on the service's `fields` definitions.
- Respects Role-Based Access Control (RBAC) at row and field levels.

---

## 5. Schema Migration Engine

The Schema Migration Engine handles safe, automated changes to the underlying database schemas when the user updates service fields.

**Supported Capabilities:**
- Add, Drop, Rename Column.
- Change column types.
- Auto-generate many-to-many join tables.

**Safety Features:**
- Executes an automatic backup before attempting structural DDL changes.
- Maintains a history inside the `migrations` table.
- Supports complete schema rollback capability via stored state snapshots.

---

## 6. GraphQL Gateway

For applications preferring flexible queries, the platform provides a dynamically generated GraphQL gateway using `gqlgen`.

**Capabilities:**
- **Queries**: Fetch single records or lists, complete with relation resolution (e.g., getting a `User` and their associated `Posts`).
- **Mutations**: Autogenerated create, update, and delete mutations.
- **Endpoint**: Accessible via `POST /api/v1/graphql`.

Example Query Supported:
```graphql
query {
  users {
    id
    name
    email
    roles {
      name
    }
  }
}
```

---

## 7. Backup System

A comprehensive backup mechanism ensures data durability and quick restoration.

**Supported Methods:**
- **Table & Schema Backups**: Raw SQL dump formats.
- **Service Snapshots**: JSON snapshot of dynamic table contents (useful for cross-database migrations).

**Endpoints (CMS Control Plane):**
- **Create**: `POST /cms/backup/service/{service_id}`
- **Restore**: `POST /cms/restore/service/{service_id}`

---

## 8. Observability Integration

The platform is built with full observability to easily detect bottlenecks or failures in production.

- **Logging**: Powered by `uber-go/zap` for high-performance, structured JSON logging. Request, error, and query logs include trace correlation IDs.
- **Metrics**: Integrated with **Prometheus**. Exposes standard metrics (`/metrics`) for HTTP request latency, slow queries, and error rates.
- **Tracing**: Instrumentations with **OpenTelemetry**. Traces propagate through the API gateway, caching, and down to the specific database connection layer, usually exported to Jaeger.

---

## 9. Unit Tests

The repository is rigorously tested to prevent regressions, enforcing a minimum of 80% code coverage.

**Test Focus:**
- Repositories (Database interactions)
- Services (Business logic & engine cores)
- API Handlers (Request mapping & validation)

**Execution:**
Run tests locally utilizing the in-memory SQLite setup:
```bash
make test-unit
# or
go test -v -race ./tests/...
```

---

## 10. Docker Setup

Containerization provides an easy path to deploy and scale this platform.

**Dockerfile Details:**
- Multi-stage builds are used to create a minimal scratch/alpine final image containing only the compiled binary.
- Built via standard command: `docker build -t cms-backend -f Dockerfile .`

**Docker Compose (`docker-compose.yml`):**
A comprehensive setup spinning up the Go backend, PostgreSQL (metadata DB), Redis (caching), Prometheus (metrics), and Jaeger (tracing).
Start the stack using:
```bash
docker-compose up -d
```

---

## 11. Swagger Documentation

API documentation is completely automated and generated according to the OpenAPI specification.

- Generated using `swag`.
- Command to regenerate: `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`
- Access the visual UI via `/swagger/index.html` on the running instance.
