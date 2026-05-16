# CMS Backend Platform

This repository contains the implementation of a self-hosted, production-grade **Backend Platform (Dynamic CMS + API Builder)** written in **Go**. The platform dynamically generates REST and optional GraphQL APIs based on user-defined data models.

## Platform Goal

Act as a scalable, cloud-ready backend infrastructure generator enabling developers to:
- Connect to various external databases (PostgreSQL, MySQL, MongoDB, SQLite).
- Create dynamic schema definitions and relational metadata.
- Automatically expose secure CRUD APIs via REST and GraphQL.
- Ensure safe schema migrations.
- Manage database and service backups.
- Track performance and health via built-in observability engines.

## Core Technology Stack

- **Language:** Go 1.24+
- **API Layer:** REST (Gin Framework) and GraphQL (gqlgen)
- **ORM:** GORM (Dynamic & Static mapping)
- **Logging:** Uber Zap (Structured Logging)
- **Metrics:** Prometheus
- **Tracing:** OpenTelemetry / Jaeger
- **Caching:** Redis (Connection Manager and Query Layer support)
- **Documentation:** Swagger / OpenAPI
- **Containerization:** Docker & Docker Compose

## System Architecture

The project enforces **Clean Architecture** and **Domain-Driven Design (DDD)** concepts mapping clearly to:
1. **Presentation Layer:** `cmd/server/`, `internal/handlers/`, and `internal/graphql/`.
2. **Application Layer:** `internal/services/`.
3. **Domain Layer:** `internal/models/`.
4. **Infrastructure Layer:** `internal/database/`, `internal/tracing/`, `internal/middleware/`.

### High-Level Components

1. **CMS Control Plane:** Manages platform configuration stored in the Metadata DB (`/api/v1/cms/`).
2. **Service Builder:** Dynamically registers fields, types, constraints, and relationships.
3. **Database Connection Manager:** Abstracts and pools multi-tenant connections (`internal/database/connection_manager.go`).
4. **Relation Engine:** Builds one-to-one, one-to-many, and many-to-many SQL table links automatically.
5. **Dynamic CRUD Engine:** Instantiates handlers resolving data at `/api/v1/data/{service_name}` (`internal/services/dynamic_data_service.go`).
6. **Query Engine:** Parses and injects filters, limits, and joins securely into GORM abstractions.
7. **GraphQL Gateway:** Bridges static schemas into dynamically available queries.
8. **Schema Migration Engine:** Uses `.Migrator()` logic to track and apply delta changes securely.
9. **Backup Engine:** Facilitates cross-service snapshot creation and payload restoration.
10. **Observability Engine:** Pre-configured telemetry exporters outputting context-tied traces and standard Prometheus HTTP histograms.

## Metadata Database Schema

The core metadata requires an initial SQL-compliant database configured natively (via `internal/models/models.go`):

- **databases:** Stores `database_connections` including target type, host, credentials, and connection limits.
- **services:** Stores definitions binding dynamic names to a specific `database_connection_id`.
- **fields:** Associates schema parameters (name, type, limits) back to `services`.
- **service_permissions:** Ties RBAC (roles like admin/viewer) directly to data constraints.
- **users** & **roles**: Defines internal admin accounts and fine-grained roles.
- **migrations**: A logging table maintaining schema changes for reliable application.
- **backups**: Saves snapshot metadata mapping local storage paths to user services.

## Go Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go                 # Entrypoint
├── internal/
│   ├── config/                     # Environment configuration mapping
│   ├── database/                   # Pool and connection registry engine
│   ├── graphql/                    # gqlgen definitions and gateway
│   ├── handlers/                   # Presentation REST layer
│   ├── middleware/                 # Rate limits, Auth, Prometheus bindings
│   ├── models/                     # Static core schema structs
│   ├── services/                   # Application layer (CRUD, CMS, Builder, Migration)
│   └── tracing/                    # OpenTelemetry integrations
├── pkg/
│   ├── logger/                     # Zap structured wrapper
│   └── response/                   # Standardized HTTP JSON output structures
├── tests/
│   ├── unit/                       # Isolated domain & service tests
│   ├── integration/                # Database-bound logic tests
│   └── uat/                        # High-level full API tests
├── docker/                         # Supplemental deployment artifacts
├── Dockerfile                      # Application build definition
├── docker-compose.yml              # Sandbox initialization
└── Makefile                        # Standard build directives
```

## Running the Platform

Ensure Go, Docker, and Make are available locally.

### Startup
1. Launch dependencies: `docker-compose up -d postgres redis jaeger prometheus`
2. Run backend: `make dev` (requires Air) or `go run ./cmd/server/main.go`

### Observability Output
- **API Swagger:** `http://localhost:8080/api/v1/swagger/index.html`
- **Prometheus Metrics:** `http://localhost:8080/metrics`
- **GraphQL Playground:** `http://localhost:8080/api/v1/graphql/playground`
- **Jaeger Dashboard:** `http://localhost:16686`

### Testing
Execute unit, integration, and coverage tests via standard tooling:
```bash
make test-coverage
```
