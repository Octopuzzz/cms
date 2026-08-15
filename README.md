# Dynamic CMS & API Builder

A self-hosted, Go-native Backend-as-a-Service (BaaS) platform that allows developers to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints. This system acts as a backend infrastructure generator, similar to Hasura or Supabase but fully built in Go.

## Features
- **CMS Control Plane**: Manage database connections, create services, define schemas, and handle relations.
- **Dynamic CRUD Engine**: Automatically generates RESTful endpoints (Create, Read, Update, Delete, List) for dynamic services.
- **Query Engine**: Advanced querying capabilities including filtering, sorting, pagination, and join queries.
- **GraphQL Gateway (Optional)**: Automatically generated GraphQL endpoints for service schemas.
- **Schema Migration Engine**: Safe schema changes with automated tracking.
- **Backup Engine**: Snapshot-based service data backup and restoration.
- **Observability**: Built-in structured logging (Zap), metrics (Prometheus), and distributed tracing (OpenTelemetry).
- **Security**: JWT authentication, fine-grained Role-Based Access Control (RBAC), row-level filtering, and rate limiting.

## System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

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
- **Presentation Layer**: The Gin HTTP Router and Handlers mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (Services) controlling data access, dynamically generating schemas, and managing platform operations.
- **Domain Layer**: Core data models defining `Service`, `Field`, `DatabaseConnection`, `User`, `Role`, etc.
- **Infrastructure Layer**: Cross-cutting utilities like telemetry (`OpenTelemetry`), metrics (`Prometheus`), and structured logging (`Zap`).

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database** (e.g., PostgreSQL). Key tables include:

1. **`database_connections`**: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. **`services`**: Represents user-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. **`fields`**: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. **`relations`**: Defines relationships (One-to-One, One-to-Many, Many-to-One, Many-to-Many) between services.
5. **`service_permissions`**: Connects `Role` to `Service` providing fine-grained access checks (e.g. CanCreate, CanRead). Contains row-level and field-level capabilities.
6. **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
7. **`backups`**: Stores snapshot records or schema outputs.
8. **`users`** and **`roles`**: General authentication and authorization for the control plane.
9. **`audit_logs`**: Detailed logging of structural and data-level modifications.

## Project Structure

The project follows a clean modular structure:

```
.
├── cmd
│   └── server                # Entry point for the application
├── internal
│   ├── config                # Configuration management
│   ├── database              # Database connections and pool management
│   ├── graphql               # GraphQL gateway and schemas
│   ├── handlers              # HTTP API handlers (REST)
│   ├── middleware            # HTTP middlewares (Auth, Prometheus, Rate Limiter, etc.)
│   ├── models                # Domain models (Metadata schema)
│   ├── services              # Business logic (CMS, CRUD, Migration, Backup, etc.)
│   └── tracing               # OpenTelemetry integration
├── pkg
│   ├── logger                # Structured logging using Zap
│   └── response              # Standardized API responses
├── tests
│   ├── integration           # Integration tests
│   ├── uat                   # User Acceptance Testing
│   └── unit                  # Unit tests for repositories, services, handlers
├── docker                    # Docker and containerization setup
├── Dockerfile                # Docker build instructions
├── docker-compose.yml        # Multi-container orchestration
├── Makefile                  # Build and test automation
└── ARCHITECTURE.md           # Internal architectural overview
```

## Engines and Core Components

### CRUD Engine Implementation
When a service is created, the system automatically generates CRUD endpoints. Operations include Create, Read, Update, Delete, and List. Endpoint mappings use dynamic routes like `/api/v1/data/{slug}` mapped to standard GORM database operations on the fly.

### Schema Migration Engine
The system supports safe schema changes including adding, dropping, renaming columns, and changing column types. It uses GORM's `.Migrator()` and tracks status in the `migrations` table natively, handling history tracking and snapshot retention logic.

### GraphQL Gateway
GraphQL is automatically generated from service schemas. It supports queries, mutations, and relations dynamically. Built using `github.com/99designs/gqlgen/graphql`, the gateway exposes `POST /api/v1/graphql` for automated GraphQL endpoints.

### Backup System
The backup engine allows for snapshot-based backups and restores. It generates data snapshots for service records, mapping dynamic table contents to snapshots, and restores rows via JSON payload decoding. API examples include `POST /cms/backup/service/{service_id}`.

### Observability Integration
Observability is integrated via:
- **Zap Logging**: Request, error, and query logs with trace IDs tied to context.
- **OpenTelemetry**: Traces mapped to SQL commands and network logic.
- **Prometheus Metrics**: Request latency, slow queries, and error rates tracked via HTTP middlewares.

## Running the Platform

### Prerequisites
- Go 1.24+
- Docker and Docker Compose

### Development
Use the provided `Makefile` to manage development workflows:

- Build the backend: `make build`
- Run with hot-reload (Air): `make dev`
- Run unit tests: `make test-unit`
- Run all tests with coverage: `make test-coverage`

### Docker Deployment
```bash
docker-compose up -d
```

### Swagger Documentation
Generate API documentation using `swag`:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

## Unit Testing
The project includes a comprehensive test suite covering the Repository, Service, and API Handler layers.
Ensure an 80% test coverage minimum using:
```bash
go test ./... -coverpkg=./...
```
