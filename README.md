# Dynamic CMS + API Builder Backend Platform

## Overview

This is a **production-grade Backend Platform (Dynamic CMS + API Builder)** written in **Go**. It works similarly to platforms like Hasura or Supabase but is **fully self-hosted and Go-native**. The platform acts as a **backend infrastructure generator**, allowing developers to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

## Core Features

- Connect databases (PostgreSQL, MySQL, MongoDB, SQLite)
- Create data models dynamically via the Service Builder
- Generate CRUD APIs automatically with the Dynamic CRUD Engine
- Generate optional GraphQL APIs natively with `gqlgen`
- Manage schema migrations securely and traceably
- Monitor logs and performance using OpenTelemetry, Prometheus, and Zap
- Manage service backups and restorations
- Scale services globally with Docker

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
PostgreSQL    MySQL       MongoDB / SQLite
```

### Core Layers

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

## Metadata Database Schema

The core internal configuration is stored in the **Metadata Database**. Tables include:

1. `database_connections`: Stores external DB configurations with fields for id, name, type, host, port, credentials. Includes connection pooling settings.
2. `services`: Represents user-created data models. Contains reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. `fields`: Defines attributes for each service, such as string, integer, float, uuid, JSON. Configures uniqueness, nullability, defaults.
4. `relations`: Defines one-to-one, one-to-many, many-to-one, and many-to-many relationships (including join tables).
5. `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks.
6. `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
7. `backups`: Stores snapshot records or schema outputs.
8. `users` and `roles`: General authentication and authorization.
9. `audit_logs`: Detailed logging of structural and data-level modifications.

## Component Implementation

### CRUD Engine & Query Engine

When a service is created, the system automatically generates standard endpoints mapping to standard GORM database operations.
- `POST /api/v1/data/{slug}`
- `GET /api/v1/data/{slug}/{id}`
- `PUT /api/v1/data/{slug}/{id}`
- `DELETE /api/v1/data/{slug}/{id}`
- `GET /api/v1/data/{slug}`

It features an advanced **Query Engine** allowing for:
- **Filtering**: `?email=test@test.com`
- **Sorting**: `?sort=created_at:desc`
- **Pagination**: `?page=1&limit=20`
- **Join Queries**: Automatic data loading via mapped relations.

### Schema Migration Engine

Managed natively within the Go layer using GORM auto-migrations merged with detailed tracking, supporting:
- Add column, Drop column, Rename column, Change column type.
- Automatic rollback capabilities and logging into the `migrations` table.

### GraphQL Gateway

Managed natively by the platform using `github.com/99designs/gqlgen`.
- Dynamic schema mapping exposing optional automated capabilities to fulfill GraphQL queries and mutations.

### Backup System

- Snapshot tables and metadata dynamically.
- Support for generic JSON service backup dumping and row restoration.

### Observability Integration

Integrated deeply into the architecture:
- **Logging**: Zap structured logging injected contextually per request.
- **Metrics**: Prometheus HTTP latency, request count, and error rate tracking. Exported at `/metrics`.
- **Tracing**: OpenTelemetry wrapping SQL, DB, and network layers tracking latency cascades across the system structure.

## Project Structure

```text
.
├── cmd
│   └── server          # Entry point for the backend platform
├── docker              # Docker containerization files
├── internal
│   ├── api             # Core API logic
│   ├── config          # Application configuration loader
│   ├── database        # Database Connector / DB Manager
│   ├── graphql         # GraphQL schema generation and gateway
│   ├── handlers        # Gin REST handlers
│   ├── middleware      # Gin middlewares (Auth, metrics, tracing)
│   ├── models          # Domain core entities (services, fields)
│   ├── services        # Core business logic (ServiceBuilder, CRUD Engine)
│   └── tracing         # Observability / OpenTelemetry integration
├── pkg
│   ├── errors          # Error standardizations
│   ├── logger          # Zap Logger utility
│   ├── pagination      # Query pagination helpers
│   ├── response        # HTTP response writers
│   └── validation      # Struct and dynamic input validation
├── tests               # Unit and Integration test suite
├── .env.example
├── ARCHITECTURE.md     # In-depth architectural breakdown
├── Dockerfile
├── Makefile            # Standard tasks automation
└── docker-compose.yml
```

## Setup & Deployment

### Dependencies

- Go 1.24+
- Docker & Docker Compose
- Target Database Platform (PostgreSQL, MySQL, MongoDB)

### Run Locally

Use the `Makefile` to simplify development tasks:

```bash
# Build the application
make build

# Run the app locally with hot reloading (air)
make dev
```

### Run via Docker

```bash
docker-compose up --build -d
```

### Swagger Documentation

API documentation is autogenerated using Swaggo. To generate docs:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

### Unit Tests

We enforce 80%+ unit test coverage for Repository, Service, and Handler layers.

```bash
# Run all unit tests
make test-unit

# Run full coverage (including all packages)
make test-coverage
```
