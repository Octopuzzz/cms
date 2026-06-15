# Backend Platform (Dynamic CMS + API Builder)

This project is a **production-grade Backend Platform** (Dynamic CMS + API Builder) written natively in **Go**. It operates similarly to platforms like Hasura or Supabase but is fully self-hosted.

The platform allows developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

---

## 1. System Architecture

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles, separating concerns into specific layers:

- **Presentation Layer**: Built with the Gin web framework. It acts as the API gateway mapping requests to internal services, including REST APIs and GraphQL endpoints using `gqlgen`.
- **Application Layer**: Contains business logic for services, dynamic schema generation, and operations requested by the API handlers.
- **Domain Layer**: Defines core data models like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Handles cross-cutting concerns like caching (Redis), telemetry (OpenTelemetry), metrics (Prometheus), and logging (Zap).

### High-Level Architecture Flow

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

The core internal configuration is stored in a metadata database (PostgreSQL is recommended).

Key tables include:

1. `database_connections`: Stores external DB configurations (id, name, type, host, port, username, password, database_name). Handles connection pooling.
2. `services`: Represents user-created data models, linking to a `database_connection_id` and tracking dynamic schemas.
3. `fields`: Defines attributes for each service (name, type, nullable, unique, default_value, index). Supports types: string, integer, float, boolean, uuid, json, array, timestamp.
4. `service_permissions`: Connects roles to services for fine-grained access control (RBAC).
5. `migrations`: Tracks schema migration history.
6. `backups`: Stores snapshot records.
7. `users` & `roles`: Authentication and authorization.
8. `audit_logs`: Logging of modifications.

---

## 3. Go Project Structure

The project uses a clean and modular structure:

```text
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── api/             # API routing and gateways
│   ├── config/          # Configuration management
│   ├── database/        # Connection manager for databases
│   ├── graphql/         # GraphQL gateway and resolvers
│   ├── handlers/        # Gin HTTP handlers
│   ├── middleware/      # Rate limiting, auth, prometheus
│   ├── models/          # Domain models
│   ├── services/        # Business logic (CRUD, Backup, Schema, etc.)
│   └── tracing/         # OpenTelemetry setup
├── pkg/
│   └── logger/          # Shared packages like structured logging
├── tests/               # Unit and integration tests
├── docker/              # Docker and Docker Compose files
├── .env.example
├── Makefile
├── Dockerfile
└── docker-compose.yml
```

---

## 4. Components

### Dynamic CRUD & Query Engine

Managed by `internal/services/dynamic_data_service.go`, the CRUD Engine automatically provides dynamic REST API endpoints for user-defined services:

- **Endpoints**: `POST /api/{service}`, `GET /api/{service}/{id}`, `PUT /api/{service}/{id}`, `DELETE /api/{service}/{id}`, `GET /api/{service}`.
- **Advanced Querying**: Supports filtering (`?email=john@example.com`), sorting (`?sort=created_at:desc`), pagination (`?page=1&limit=20`), and join queries.

### Schema Migration Engine

Managed by `internal/services/migration_service.go`, the Schema Migration Engine handles safe schema changes directly through GORM's `Migrator()`.

- **Supported Operations**: Add column, Drop column, Rename column, Change column type.
- **Safety**: Automatically logs migration status and enables rollback capabilities via snapshot retention.

### GraphQL Gateway

Managed by `internal/graphql/gateway.go`, the system automatically exposes a generic GraphQL endpoint `POST /api/v1/graphql` generated from your service schemas using `gqlgen`.

### Backup System

Managed by `internal/services/backup_service.go`, this engine handles taking snapshots of your data and schemas.

- **Capabilities**: Back up services to JSON snapshots and restore them.
- **API**: `POST /cms/backup/service/{service_id}` and `POST /cms/restore/service/{service_id}`.

### Observability Integration

The platform includes a robust observability suite natively integrated into the application layer:

- **Logging**: Structured JSON logging using **Zap**. Correlates traces with contexts.
- **Metrics**: HTTP metrics captured via **Prometheus** interceptors, exposed at `/metrics`.
- **Tracing**: Distributed tracing via **OpenTelemetry** and exported to **Jaeger**, wrapping database and network operations.

---

## 5. Docker Setup

A complete `docker-compose.yml` is provided for running the platform along with its infrastructure dependencies:

- **CMS Backend** (Go application)
- **PostgreSQL** (Metadata database)
- **Redis** (Caching layer)
- **Jaeger** (Tracing UI & Agent)
- **Prometheus** (Metrics aggregation)

To start the platform:
```bash
docker-compose up -d --build
```

---

## 6. Unit Testing

The project implements extensive testing across the Application and Presentation layers with a target of **80% coverage**.

Run the tests using standard Go commands or via the provided Makefile:
```bash
go test ./...
# or
make test-unit
# or to get full coverage output
make test-coverage
```

---

## 7. API Documentation (Swagger)

API endpoints are documented using Swagger/OpenAPI.

To generate or update the documentation, run:
```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

The documentation provides interactive details for the CMS Control Plane, generated CRUD endpoints, and administrative APIs.