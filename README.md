# Go-Native Backend Platform (Dynamic CMS + API Builder)

This project is a production-grade Backend Platform written in Go. It acts as a Backend-as-a-Service (BaaS) and backend infrastructure generator, similar to platforms like Hasura or Supabase, but fully self-hosted and Go-native.

## Overview

The platform allows developers to:
- Connect databases (PostgreSQL, MySQL, MongoDB, SQLite)
- Create data models dynamically via a Service Builder
- Generate CRUD REST APIs automatically
- Generate optional GraphQL APIs
- Manage schema migrations safely
- Monitor logs and performance metrics
- Manage backups
- Scale services efficiently

## System Architecture Explanation

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
PostgreSQL    MySQL       MongoDB / SQLite
```

- **Presentation Layer**: Built with Gin framework for REST and `gqlgen` for GraphQL. Handles routing, authentication, and request parsing.
- **Application Layer**: Contains business logic (`internal/services/`). Manages dynamic schemas, data access, and orchestration.
- **Domain Layer**: Defines core entities and interfaces (`internal/models/`).
- **Infrastructure Layer**: Cross-cutting tools including connection caching, OpenTelemetry tracing, Prometheus metrics, and Zap structured logging.

## Metadata Database Schema

The CMS stores platform metadata, primarily in PostgreSQL (or SQLite for development/testing).

Core tables:
- **`database_connections`**: Stores external DB configurations (`id`, `name`, `type`, `host`, `port`, `username`, `password`, `database_name`).
- **`services`**: Represents dynamic data models (`id`, `name`, `database_connection_id`, `db_table_name`, `created_at`, `updated_at`).
- **`fields`**: Defines attributes for each service, configuring `type`, `nullable`, `unique`, `default_value`, etc.
- **`relations`**: Defines relationships (One-to-One, One-to-Many, Many-to-One, Many-to-Many).
- **`migrations`**: Tracks DDL executions and rollback metadata.
- **`backups`**: Stores snapshot records or schema outputs.
- **`users` & `roles`**: Manages RBAC and platform security.

## Go Project Structure

The project uses a clean modular structure standard for Go applications:

```
.
├── cmd/
│   └── server/          # Main application entrypoint
├── internal/
│   ├── config/          # Configuration and environment setup
│   ├── database/        # Database Connection Manager (PostgreSQL, MySQL, MongoDB, SQLite)
│   ├── graphql/         # GraphQL Gateway implementation using gqlgen
│   ├── handlers/        # Gin HTTP Handlers (Presentation Layer)
│   ├── middleware/      # Auth, Rate Limiter, Prometheus metrics
│   ├── models/          # Core Domain Models
│   ├── services/        # Application Layer (Service Builder, CMS, CRUD, Migration, etc.)
│   └── tracing/         # OpenTelemetry tracing setup
├── pkg/
│   ├── logger/          # Zap structured logging
│   └── response/        # Standardized API responses
├── tests/
│   ├── integration/     # Integration test suite
│   ├── uat/             # User Acceptance Testing
│   └── unit/            # Unit tests covering Repositories, Services, and Handlers
├── docker/              # Dockerfiles and docker-compose configurations
└── docs/                # Generated Swagger/OpenAPI documentation
```

## CMS Control Plane & Service Builder

The **CMS Control Plane** (`internal/services/service_service.go`, `internal/services/dbconn_service.go`) allows users to register databases, dynamically create services (data models), and manage metadata.

Users can define schema properties such as fields (string, integer, float, boolean, uuid, json, timestamp) and relations natively through REST endpoints.

## CRUD Engine & Query Engine Implementation

The **Dynamic CRUD Engine** (`internal/services/dynamic_data_service.go`) automatically maps standard operations for any dynamically generated service:

- `POST /api/v1/data/{slug}`
- `GET /api/v1/data/{slug}/{id}`
- `PUT /api/v1/data/{slug}/{id}`
- `DELETE /api/v1/data/{slug}/{id}`
- `GET /api/v1/data/{slug}`

The **Query Engine** builds upon GORM to support:
- Filtering (e.g., `?email=test@test.com`)
- Sorting (e.g., `?sort=created_at:desc`)
- Pagination (e.g., `?page=1&limit=20`)
- Dynamic Joining for relations.

## Schema Migration Engine

The **Schema Migration Engine** (`internal/services/migration_service.go`) allows safe modifications to generated schemas.
It supports adding, dropping, renaming columns, and changing types using GORM’s Migrator. It inherently records migration status into the `migrations` metadata table, offering a foundation for robust version tracking and safe DDL deployments.

## GraphQL Gateway

An **Optional GraphQL Gateway** (`internal/graphql/gateway.go`) runs alongside the REST API. Built with `gqlgen`, it maps dynamic schemas to GraphQL definitions, allowing for powerful nested queries and aggregations from a single entry point at `POST /api/v1/graphql`.

## Backup System

The **Backup Engine** (`internal/services/backup_service.go`) implements data safety protocols:
- Table backups and JSON snapshots
- Allows dumping service contents into persistent records
- Facilitates restoring exact row states from JSON payload decodings.

## Observability Integration

The platform integrates deep observability:
- **Logging**: Zap structured logging (`pkg/logger/`) with correlation IDs to track requests contextually.
- **Metrics**: Prometheus middleware (`internal/middleware/prometheus.go`) tracks request latency, errors, and throughput, exposed via `/metrics`.
- **Tracing**: OpenTelemetry (OTel) + Jaeger integration (`internal/tracing/`) for detailed span traces covering network and SQL layer events.

## Security & Performance

- **Security**: Implements JWT Authentication, fine-grained Role-Based Access Control (RBAC), SQL injection protection (via GORM parameterized queries), and a Rate Limiter middleware.
- **Performance**: Incorporates DB connection pooling and is built natively to integrate with optional Redis caching. Batch insertions (`db.CreateInBatches`) mitigate N+1 query patterns.

## Unit Tests

The system maintains high coverage with comprehensive unit testing using the standard `testing` package and in-memory SQLite (`:memory:`) databases to mock complex behaviors.

To run the full suite (Unit, Integration, UAT) and generate a coverage profile:
```bash
make test-coverage
# or
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Docker Setup

The platform is containerized for seamless deployments.
- Refer to `docker-compose.yml` to spin up the backend along with necessary infrastructure.
- Uses multi-stage builds (`Dockerfile`) to produce minimal deployment artifacts.

To run:
```bash
docker-compose up -d
```

## Swagger Documentation

The project includes standard Swagger / OpenAPI documentation detailing all CMS management and data APIs.
To generate or update Swagger docs:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
(Note: `docs/` is explicitly ignored by version control but should be built in CI/CD).
