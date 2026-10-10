# Dynamic CMS & API Builder

A production-grade Backend Platform written in Go. This system is a self-hosted, Go-native Backend-as-a-Service (BaaS) platform that allows users to dynamically create backend services, manage database schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

It acts as a backend infrastructure generator, similar to platforms like Hasura or Supabase.

## Table of Contents
- [Platform Goal](#platform-goal)
- [Core Technology Requirements](#core-technology-requirements)
- [System Architecture Explanation](#system-architecture-explanation)
- [Metadata Database Schema](#metadata-database-schema)
- [Go Project Structure](#go-project-structure)
- [CMS Control Plane & Service Builder](#cms-control-plane--service-builder)
- [Database Connection Manager](#database-connection-manager)
- [CRUD Engine Implementation & Query Engine](#crud-engine-implementation--query-engine)
- [Relation Engine](#relation-engine)
- [Schema Migration Engine](#schema-migration-engine)
- [GraphQL Gateway](#graphql-gateway)
- [Backup System](#backup-system)
- [Observability Integration](#observability-integration)
- [Performance Optimization & Security](#performance-optimization--security)
- [Unit Tests](#unit-tests)
- [Docker Setup](#docker-setup)
- [Swagger Documentation](#swagger-documentation)

## Platform Goal
Build a Backend-as-a-Service platform that allows developers to:
- Connect databases
- Create data models dynamically
- Generate CRUD APIs automatically
- Generate optional GraphQL APIs
- Manage schema migrations
- Monitor logs and performance
- Manage backups
- Scale services

## Core Technology Requirements
- **Language**: Go
- **API Layer**: REST (default), GraphQL (optional)
- **GraphQL Library**: `gqlgen`
- **Web Framework**: Gin
- **ORM**: GORM
- **Logging**: Zap structured logging
- **Metrics**: Prometheus
- **Tracing**: OpenTelemetry
- **Cache**: Redis
- **Database Migration**: GORM AutoMigrate / Custom Engine
- **Containerization**: Docker
- **API Documentation**: Swagger / OpenAPI

## System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles.

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
PostgreSQL    MySQL       MongoDB
```

### Core Layers
- **Presentation Layer**: The Gin HTTP Router and Handlers. Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (Services). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models. Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry, metrics, and logging.

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database** (e.g., PostgreSQL).

Tables include:
- `databases` / `database_connections`: Stores external DB configurations (id, name, type, host, port, credentials).
- `services`: Represents user-created data models (id, name, database_id, db_table_name, created_at, updated_at).
- `fields`: Defines attributes for each service (name, type, nullable, unique, default_value, index).
- `relations`: Configures relationships between services.
- `migrations`: Keeps track of DDL executions.
- `backups`: Stores snapshot records or schema outputs.
- `users` and `roles`: General authentication and authorization.

## Go Project Structure

The project uses a clean modular structure:
```text
cmd/
└── server/          # Application entrypoint
internal/
├── api/             # API routes
├── config/          # Configuration loading
├── database/        # Database connection manager
├── graphql/         # GraphQL gateway
├── handlers/        # Gin HTTP handlers
├── middleware/      # Auth, rate limiting, telemetry
├── models/          # Domain models
├── services/        # Business logic (CRUD, Migrations, etc.)
└── tracing/         # OpenTelemetry setup
pkg/
├── errors/          # Custom error definitions
├── logger/          # Zap logger wrapper
├── pagination/      # Pagination helpers
└── response/        # API response formatters
tests/               # Unit, integration, and UAT tests
docker/              # Docker setup
```

## CMS Control Plane & Service Builder

The management interface controls the system.
- Manage database connections
- Create services and define schemas
- Manage relations, migrations, and backups

**API Examples:**
- `POST /api/v1/cms/databases`
- `GET /api/v1/cms/databases`
- `POST /api/v1/cms/services`
- `GET /api/v1/cms/services`

Users can dynamically create services (data models) containing fields, relations, and validation rules. Supported field types include string, integer, float, boolean, uuid, json, array, timestamp.

## Database Connection Manager

Users can register external databases: PostgreSQL, MySQL, and MongoDB. Connection pooling is implemented for performance optimization.

## CRUD Engine Implementation & Query Engine

When a service is created, the system automatically generates CRUD endpoints.
- `POST /api/v1/data/{service}`
- `GET /api/v1/data/{service}/{id}`
- `PUT /api/v1/data/{service}/{id}`
- `DELETE /api/v1/data/{service}/{id}`
- `GET /api/v1/data/{service}`

**Query Engine features:**
- Filtering: `GET /api/v1/data/users?email=john@example.com`
- Sorting: `GET /api/v1/data/users?sort=created_at:desc`
- Pagination: `GET /api/v1/data/users?page=1&limit=20`
- Join queries mapping to foreign keys.

## Relation Engine

The system supports relationships: One-to-One, One-to-Many, Many-to-One, Many-to-Many (automatically creates join tables).

## Schema Migration Engine

Supports safe schema changes:
- Add column, Drop column, Rename column, Change column type
- Automatic tracking in `migrations` table
- Rollback capability

## GraphQL Gateway

GraphQL is automatically generated from service schemas using `gqlgen`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL endpoints for defined data models.
- Supports queries, mutations, and relations.

## Backup System

The platform supports:
- Table backup / Service snapshot
- Restore capabilities

**Example:**
- `POST /api/v1/cms/backup/service/{service_id}`
- `POST /api/v1/cms/restore/{id}`

## Observability Integration

Implemented across the stack:
- **Logging**: Request, Error, and Query logs using Zap (correlated with Trace IDs).
- **Metrics**: Request latency, slow queries, error rates tracked via Prometheus (`/metrics`).
- **Tracing**: OpenTelemetry (Jaeger) wraps SQL commands and network logic.

## Performance Optimization & Security

- **Performance**: Database connection pooling, query caching (Redis), pagination optimizations.
- **Security**: JWT authentication, Role-Based Access Control (RBAC), input validation, SQL injection protection via GORM, Rate limiting.

## Unit Tests

The system includes comprehensive unit tests using the standard Go `testing` library and in-memory SQLite (`:memory:`).
Minimum coverage expected is 80% across Repository, Service, and API handler layers.

**Command:**
```bash
make test-coverage
# or
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Docker Setup

The system is fully containerized using Docker and `docker-compose`. This spins up the backend platform alongside necessary infrastructure (Redis, PostgreSQL metadata DB, Jaeger, Prometheus).

## Swagger Documentation

API documentation is available via Swagger/OpenAPI.
Generated using `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
Accessible at `/swagger/index.html`.
