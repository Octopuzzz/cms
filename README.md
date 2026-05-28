# Backend Platform (Dynamic CMS + API Builder)

A production-grade Backend Platform and CMS written in Go. This platform functions similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native. It allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

The system is modular, scalable, and cloud-ready, built using **Clean Architecture** and **Domain-Driven Design (DDD)**.

## Core Features

- Connect external databases (PostgreSQL, MySQL, MongoDB, SQLite).
- Create data models dynamically via a Metadata Database.
- Generate CRUD APIs automatically.
- Generate optional GraphQL APIs natively.
- Manage schema migrations safely.
- Monitor logs and performance (Prometheus, OpenTelemetry).
- Manage backups and restorations.
- Scale services.

## Core Technologies

- **Language:** Go 1.24+
- **API Layer:** REST (default) via **Gin**, GraphQL via **gqlgen**
- **ORM:** GORM
- **Logging:** Zap structured logging
- **Metrics:** Prometheus
- **Tracing:** OpenTelemetry / Jaeger
- **Cache:** Redis
- **Containerization:** Docker & Docker Compose
- **API Documentation:** Swagger / OpenAPI

---

## High Level Platform Architecture

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

## System Components

### 1. CMS Control Plane & Service Builder
The control plane allows developers to define services, fields, database connections, and relationships dynamically.
- `internal/services/service_service.go`
- `internal/services/dbconn_service.go`

### 2. Metadata Database Schema
The platform relies on a central metadata database to track the platform's state.

**Core Tables:**
- `database_connections`: id, name, type, host, port, username, password, database_name
- `services`: id, name, database_id, created_at, updated_at
- `fields`: id, service_id, name, type, nullable, unique, default_value
- `relations`: id, service_id, related_service_id, type
- `migrations`: tracking schema changes
- `backups`: tracking backups and snapshots
- `users` / `roles`: identity and RBAC

### 3. Dynamic CRUD Engine & Query Engine
Generates and serves standard CRUD logic and advanced queries on-the-fly when a service is created.
- File: `internal/services/dynamic_data_service.go`
- **Endpoints Provided:**
  - `POST /api/v1/data/{slug}` (Create)
  - `GET /api/v1/data/{slug}` (List with pagination, sorting, filtering)
  - `GET /api/v1/data/{slug}/{id}` (Read)
  - `PUT /api/v1/data/{slug}/{id}` (Update)
  - `DELETE /api/v1/data/{slug}/{id}` (Delete)
- **Features:** Pagination (`?page=1&limit=20`), Sorting (`?sort=created_at:desc`), and Dynamic Filtering (`?email=john@example.com`).

### 4. Schema Migration Engine
Tracks and applies schema changes using GORM AutoMigrate features.
- File: `internal/services/migration_service.go`
- Handles safely adding, dropping, and renaming columns along with rollback capabilities.

### 5. GraphQL Gateway (Optional)
Automatically builds and exposes GraphQL queries based on dynamic models.
- File: `internal/graphql/gateway.go`
- Endpoint: `POST /api/v1/graphql`

### 6. Backup Engine
Supports generating data snapshots (JSON/SQL).
- File: `internal/services/backup_service.go`

### 7. Observability Integration
The platform offers full telemetry:
- **Logging:** Built with `go.uber.org/zap` for structured, leveled JSON logging, injected via `pkg/logger`.
- **Metrics:** `Prometheus` tracks metrics natively using interceptors in `internal/middleware/prometheus.go` (e.g., latency, status codes). Exposed at `/metrics`.
- **Tracing:** `OpenTelemetry` integrated via `internal/tracing/tracing.go` pushing traces to Jaeger, effectively tracing HTTP requests and internal service calls.

---

## Go Project Structure

The project structure adheres to Clean Architecture guidelines:

```
├── cmd
│   └── server             # Main application entry point
├── docker                 # Dockerfiles and infrastructure configs
├── internal
│   ├── config             # Configuration and environment setup
│   ├── database           # Database connection manager (Pools)
│   ├── graphql            # gqlgen integration and gateway
│   ├── handlers           # HTTP layer (Gin handlers)
│   ├── middleware         # Auth, Prometheus, Rate Limiting
│   ├── models             # Domain models (Services, Fields, etc)
│   ├── services           # Application business logic
│   └── tracing            # OpenTelemetry / Jaeger logic
├── pkg
│   ├── logger             # Zap logger wrapper
│   └── response           # Standardized API response formatters
├── tests                  # Unit, Integration, UAT tests
├── Makefile               # Build and test commands
├── docker-compose.yml     # Local orchestration
└── go.mod                 # Go module definition
```

---

## Docker Setup

The project includes a robust `docker-compose.yml` to set up the backend along with necessary infrastructure (PostgreSQL, Redis, Prometheus, Jaeger).

```bash
# Start all services
docker-compose up -d

# Stop services
docker-compose down
```

---

## API Documentation (Swagger)

Swagger API documentation is generated via `swaggo/swag`.

```bash
# Generate Swagger docs
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

The documentation is available locally via Swagger UI mapped to `/swagger/index.html`.

---

## Unit Testing

Minimum coverage requirement is 80%.
Tests cover the repository, service layer, and API handlers. Use standard go test or the provided `Makefile`.

```bash
# Run tests
go test ./...

# Run tests with coverage profiling
make test-coverage
```
