# Dynamic CMS & API Builder (Backend Platform)

A production-grade Backend-as-a-Service platform written in Go. This system is a self-hosted, Go-native alternative to platforms like Hasura or Supabase. It allows users to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

## System Architecture

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)**.

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
PostgreSQL    MySQL       MongoDB
```

### Core Layers
- **Presentation Layer**: Gin HTTP Router and Handlers. Serves as the API gateway mapping requests to internal services, including GraphQL endpoints via `gqlgen`.
- **Application Layer**: Business logic. Controls data access, generates dynamic schemas, and performs requested operations.
- **Domain Layer**: Core data models defining entities like `Service`, `Field`, `DatabaseConnection`, `User`, and `Role`.
- **Infrastructure Layer**: Cross-cutting tools including connection caching, telemetry (OpenTelemetry), metrics (Prometheus), and structured logging (Zap).

## Metadata Database Schema

The CMS stores platform configuration in the **Metadata Database** (e.g., PostgreSQL). Core tables include:

- `database_connections`: External DB configurations (id, name, type, host, port, credentials).
- `services`: User-created data models referencing `database_connection_id`.
- `fields`: Attributes for each service (type, nullability, unique, default_value).
- `service_permissions`: Connects `Role` to `Service` for fine-grained access control.
- `migrations`: Tracks DDL executions and metadata.
- `backups`: Snapshot records or schema outputs.
- `users` / `roles`: Authentication and authorization.

**Example `services` table:**
- `id` (UUID)
- `name` (String)
- `database_connection_id` (UUID)
- `db_table_name` (String)
- `created_at` (Timestamp)
- `updated_at` (Timestamp)

## Go Project Structure

The project uses a clean, modular layout:

```text
├── cmd
│   └── server            # Application entrypoint
├── docker                # Docker and Docker Compose files
├── internal
│   ├── config            # Configuration management
│   ├── database          # Database connection managers and drivers
│   ├── graphql           # GraphQL gateway and schemas
│   ├── handlers          # REST API handlers and presentation layer
│   ├── middleware        # HTTP middlewares (Auth, Rate Limit, etc.)
│   ├── models            # Domain models and entities
│   ├── services          # Business logic and platform engines
│   └── tracing           # OpenTelemetry and observability tools
├── pkg
│   ├── logger            # Zap structured logging wrapper
│   └── response          # Standardized HTTP response formatter
└── tests
    ├── integration       # Integration tests
    ├── uat               # User Acceptance Tests
    └── unit              # Unit tests
```

## CRUD & Query Engine Implementation

The **Dynamic CRUD Engine** maps endpoints like `GET /api/v1/data/{slug}` to standard GORM database operations on the fly. It automatically handles:
- **Create / Read / Update / Delete / List**: Core operations dynamically constructed based on the service metadata.
- **Query Engine**:
  - **Filtering**: `GET /api/v1/data/users?email=test@test.com`
  - **Sorting**: `GET /api/v1/data/users?sort=created_at:desc`
  - **Pagination**: `GET /api/v1/data/users?page=1&limit=20`
  - **Relations/Joins**: Leverages foreign key mappings to join related data.
- **Security**: Verifies Role-Based Access Control (RBAC) and row-level filtering dynamically prior to executing queries.

## Schema Migration Engine

Tracks and applies safe schema diffs dynamically:
- Supports adding, dropping, and renaming columns, as well as changing column types using GORM's Migrator.
- Logs migration statuses in the `migrations` metadata table.
- Implements fallback/rollback mechanisms to guarantee safety across schema changes.

## GraphQL Gateway

Generates GraphQL endpoints dynamically from service schemas:
- Built with `gqlgen`.
- Exposes `POST /api/v1/graphql` for automated GraphQL operations over defined data models.
- Supports generic queries, mutations, and relations.

## Backup System

Automated and manual backup operations:
- Capable of generating data snapshots and schemas.
- Maps dynamic table contents to snapshots (e.g., JSON export).
- Supports restoring rows from backups seamlessly via internal restore methods.

## Observability Integration

The platform is designed to be highly observable:
- **Logging**: Zap structured logging with correlation and trace IDs tied to context.
- **Metrics**: Prometheus metrics tracking request latency, slow queries, and error rates (exported at `/metrics`).
- **Tracing**: OpenTelemetry (integrated with Jaeger) wrapping SQL commands and network logic to track request paths.

## Unit Testing

Tests are written using the standard Go `testing` library, covering Repository, Service, and API handler layers.
- Minimum 80% coverage required.
- Uses an in-memory SQLite database setup (`:memory:`) via the `github.com/glebarez/sqlite` driver for database mocking.
- **Run Unit Tests**: `go test ./tests/unit/...`
- **Run Coverage**: `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`

## Docker Setup

The platform is containerized for easy deployment:
- `Dockerfile`: Multi-stage build for compiling the Go binary and creating a minimal runtime image.
- `docker-compose.yml`: Easily spin up the backend platform alongside necessary infrastructure (e.g., PostgreSQL, Redis, Jaeger).

**Run with Docker Compose:**
```bash
docker-compose up -d
```

## API Documentation (Swagger)

REST API endpoints are documented using Swagger/OpenAPI.
- **Generate Docs**: `~/go/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`
- Automatically updates standard request/response payloads for platform management operations.
