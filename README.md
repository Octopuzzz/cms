# Dynamic CMS + API Builder (Go-Native)

## Overview
This platform is a **Backend-as-a-Service (BaaS)** that dynamically generates REST and GraphQL APIs, managing backend infrastructure directly from Go. It acts as an infrastructure generator allowing you to connect databases, create data schemas and automatically spin up robust CRUD APIs with integrated GraphQL options.

## System Architecture

The core of this system adheres to **Clean Architecture** and **Domain-Driven Design (DDD)** paradigms.

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

### Core Components
1. **CMS Control Plane**: Management of databases, schemas, services, back-ups, and migrations.
2. **Metadata Database**: Built-in persistence of internal configurations.
3. **Database Connection Manager**: Abstraction pool allowing mapping over PostgreSQL, MySQL, MongoDB, SQLite.
4. **Service Builder & Relational Engine**: Defining relations natively mapping One-to-One, One-to-Many, etc.
5. **Dynamic CRUD Engine**: Generates API endpoints synchronously when schemas change.
6. **Query Engine**: Pagination, sorting, and nested joins natively mapped.
7. **GraphQL Gateway**: Translates Dynamic Data Schemas into full-featured GraphQL Resolvers and Mutations.
8. **Schema Migration Engine**: Safe runtime migrations with automatic backups and snapshot states.
9. **Backup Engine**: Automated snapshot generation across different storage layers.
10. **Observability**: Prometheus metrics, Zap Logging, OpenTelemetry + Jaeger tracing.

## Metadata Database Schema
We rely heavily on structured metadata schemas to describe the state of user-generated components:
- **`databases`**: Tracks connection parameters (host, port, user, etc.)
- **`services`**: Tracks the individual service name and link to its respective Database.
- **`fields`**: Describing individual variables tied to a schema (Types include `string`, `int`, `uuid`, etc.)
- **`relations`**: Connecting multiple schemas, implicitly generating junction tables for Many-To-Many relations.
- **`migrations`**: Holds historical state mutations for rollbacks.
- **`backups`**: Records snapshot states and references local/remote storage paths.

## Go Project Structure
The repository defines strict modular isolation:
- `cmd/server/`: The entry point injecting dependencies and initializing the server.
- `internal/api/`: Handles specific presentation formatting.
- `internal/handlers/`: Gin router implementation parsing payloads to interact with services.
- `internal/services/`: Holds core logic including CRUD, Query, Migrations, etc.
- `internal/graphql/`: The gqlgen adapter bridging the HTTP router.
- `internal/database/`: The manager handling external database connections dynamically.
- `pkg/`: Shared utilities including standard response wrappers, pagination parsing, and logging.
- `docker/`: Holds docker-compose and generic Dockerfile generation logic.

## Deployment & Environments

### Docker Setup
We ship a `docker-compose.yml` natively integrating Jaeger and Prometheus alongside the primary Go runtime.

```sh
docker-compose up -d
```

### Testing
Testing requires Go 1.24. Unit and Integration tests mock an internal SQLite memory store.
```sh
make test-unit
```

### Swagger Documentation
The REST components use declarative Swagger configurations embedded in Handlers.
```sh
make swagger
```
