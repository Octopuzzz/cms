# Go Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend-as-a-Service platform (Dynamic CMS + API Builder) written in Go. The platform works similarly to Hasura or Supabase but is fully self-hosted and Go-native. It allows developers to connect databases, create data models dynamically, generate CRUD APIs automatically, generate optional GraphQL APIs, manage schema migrations, monitor logs, perform backups, and scale services.

## System Architecture

The system follows **Clean Architecture and Domain Driven Design (DDD)**.

### Layers:
- **Presentation Layer (`internal/handlers`)**: Handles incoming HTTP requests using Gin. Exposes standard REST endpoints and an optional GraphQL gateway.
- **Application Layer (`internal/services`)**: Business logic layer orchestrating database connections, service building, CRUD operations, query engine execution, backups, and schema migrations.
- **Domain Layer (`internal/models`)**: Defines core models representing metadata (Services, Fields, Relations, Database Connections, Users, Permissions, etc.).
- **Infrastructure Layer**: Connectors for various databases (PostgreSQL, MySQL, MongoDB, SQLite), tracing (OpenTelemetry), metrics (Prometheus), logging (Zap), and standard caching mechanisms.

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

## Metadata Database Schema

The core system stores its metadata in the internal database. This tracks everything needed to dynamically route data, enforce RBAC schemas, and manage platform stability.

- **`database_connections`**: Stores configuration parameters (host, port, credentials) to connect to internal or external databases (Postgres, MySQL, MongoDB). Handles connection pooling.
- **`services`**: Defines user-created data models / entities. Tracks the target `database_connection_id` and the generated `db_table_name`.
- **`fields`**: Represents the schema columns of a service. Contains field type (string, integer, float, json, uuid, etc.), nullability, default values, validations, and relationship mappings.
- **`service_permissions`**: Controls RBAC logic at the entity level as well as providing granular field-level permissions and Row-Level security via raw filters.
- **`migrations`**: Audits and stores DDL migrations associated with creating or altering tables across services.
- **`backups`**: Stores payload snapshots and schema dumps for services.
- **`users`, `roles`, `permissions`, `audit_logs`**: Core platform security tables controlling CMS access.

## Go Project Structure

The project implements a highly modular structure suitable for microservice expansion or monolithic deployment:

```text
├── cmd
│   └── server          # Application entrypoint (main.go)
├── docker              # Docker containerization files
├── internal
│   ├── config          # Environment configuration loading
│   ├── database        # Database Connection Manager (pools, external DB wrappers)
│   ├── graphql         # gqlgen-powered Optional GraphQL Gateway
│   ├── handlers        # Gin REST API endpoints
│   ├── middleware      # Auth, RateLimiting, Prometheus Interceptors
│   ├── models          # Domain core & metadata schemas
│   ├── services        # Core engines: CMS, Builders, Migrations, Backups
│   └── tracing         # OpenTelemetry & Request Correlation
├── pkg
│   ├── logger          # Zap Structured Logger
│   └── response        # Standardized JSON response formatting
├── tests
│   ├── integration     # Service-level integration tests
│   ├── uat             # User Acceptance Testing
│   └── unit            # Core unit tests
├── .env.example
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## Platform Components Implementation

### 1. CMS Control Plane & Service Builder
Controlled via `ServiceService` and `DBConnService`, exposed through `/api/v1/cms/` endpoints.
- Allows registering multiple database sources.
- Enables dynamic creation of Data Services and schema configurations. Users configure columns, relationships (one-to-one, one-to-many, many-to-many), validations, and constraints via API requests.
- Validates changes and triggers the `MigrationService`.

### 2. CRUD Engine & Query Engine
Managed by `DynamicDataService` and exposed under `GET/POST/PUT/DELETE /api/v1/data/{slug}`.
- Maps standard HTTP requests seamlessly to the underlying mapped `database_connection`.
- Features an advanced Query Engine executing:
  - **Filtering**: Query parameter driven (e.g. `?email=user@example.com`).
  - **Sorting**: Order clauses (e.g. `?sort=created_at:desc`).
  - **Pagination**: Efficient offsets based on `?page=1&limit=20`.
  - **Relational Joins**: Capable of nesting related dynamic models.
- Applies Role-Based Access Control and Row-Level security filters per query execution natively.

### 3. Schema Migration Engine
Managed by `MigrationService`.
- Automatically executes safe schema updates directly to the connected databases.
- Applies standard operations such as creating tables, adding columns, dropping columns, or renaming structure using GORM `.Migrator()` implementations mapping to raw SQL DDL depending on the target database driver.
- Logs migrations into the metadata table with snapshots for safe history tracking and possible rollbacks.

### 4. Backup Engine
Managed by `BackupService`.
- Exposes endpoints to trigger asynchronous backups of any user-created Service.
- Captures schema representations alongside full record payload snapshots encoding massive volumes of data via JSON streams.
- Facilitates simple restoration protocols.

### 5. GraphQL Gateway
A standard integration utilizing `github.com/99designs/gqlgen/graphql` inside `internal/graphql/gateway.go`.
- Maps generic GraphQL queries dynamically against active Service structures.
- Allows clients to bypass REST and run nested relation mappings via a centralized API Gateway endpoint: `POST /api/v1/graphql`.

### 6. Observability Integration
Enterprise-ready operational readiness via:
- **Zap Logging**: Structured JSON logging. Traces and correlates request IDs across boundaries.
- **Prometheus Metrics**: Injected standard Gin Middlewares exporting route timing, request latencies, and total request capacities exposed at the `/metrics` endpoint.
- **OpenTelemetry**: Integrated throughout the infrastructure ensuring distributed trace capabilities for advanced performance bottleneck detection.

### 7. Security Best Practices
- Implements comprehensive JWT authentication mapping to custom Role & Permissions sets.
- Employs strict request rate limiters via `internal/middleware/rate_limiter.go`.
- Safeguards raw query inputs securely preventing structural SQL injections when querying dynamic engines.

### 8. Unit Testing and Verification
Unit Tests are tightly integrated targeting the Repository, Service, and API Handler layers maintaining over 80% baseline coverage. Tests are orchestrated locally mapping against transient in-memory SQLite (`:memory:`) instances to guarantee test isolation. Run unit tests via standard commands:
`make test-unit` or `go test -v -race -coverpkg=./... ./tests/...`.

### 9. Docker Setup
Fully cloud-ready via robust Dockerization.
- **Dockerfile**: Minimal multi-stage builder creating statically linked binaries to optimize image size and startup latency.
- **docker-compose.yml**: Connects the central platform to default surrounding infrastructure dependencies like standard PostgreSQL or Redis services if applicable locally.

### 10. API Documentation (Swagger)
Complete API schemas and REST pathways are continuously documented leveraging standard Swagger / OpenAPI configurations on Gin. Standard UI routes expose the API interface intuitively for operators. Generated via `swag init`.
