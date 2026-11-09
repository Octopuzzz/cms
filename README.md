# CMS Backend Platform

## Overview
A production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. This self-hosted, Go-native platform allows developers to dynamically create backend services, schemas, REST APIs, and optional GraphQL endpoints. It is modular, scalable, and cloud-ready, adhering to Clean Architecture and Domain-Driven Design (DDD).

## System Architecture Explanation

The system follows a layered architecture to cleanly separate concerns:
- **Presentation Layer**: Handlers and API gateway processing incoming REST and GraphQL requests.
- **Application Layer**: Business logic, including CMS operations, service builders, dynamic CRUD execution, and migration services.
- **Domain Layer**: Core data models representing schemas, fields, roles, relations, and metadata.
- **Infrastructure Layer**: Cross-cutting modules like logging (Zap), metrics (Prometheus), tracing (OpenTelemetry), and external database connectivity (PostgreSQL, MySQL, MongoDB, SQLite).

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

## Metadata Database Schema

The core metadata schema configures the dynamic nature of the platform:
- `database_connections`: Stores external DB configuration (type, host, port, credentials).
- `services`: Represents dynamically created models/tables. Tied to a specific database connection.
- `fields`: Schema definitions for services (type, nullable, unique, default values).
- `relations`: Manages relationship mappings (One-to-One, One-to-Many, Many-to-Many) and join tables.
- `migrations`: Audit log of schema changes and migrations applied to services.
- `backups`: Tracks snapshots of service data and schemas.
- `users` / `roles`: For platform access control (RBAC).

## Go Project Structure

The codebase is organized cleanly:
```
cmd/
  server/         # Entry point for the CMS backend
internal/
  api/            # API definitions
  config/         # App configuration management
  database/       # Connection manager and GORM integrations
  graphql/        # GraphQL Gateway using gqlgen
  handlers/       # HTTP controllers mapping routes to services
  middleware/     # Auth, tracing, rate-limiting, metrics
  models/         # Domain structures
  services/       # Business logic (CRUD, Service Builder, Migration, etc.)
  tracing/        # OpenTelemetry integration
pkg/
  logger/         # Zap structured logging
  response/       # Standardized JSON responses
tests/            # Unit, integration, and UAT tests
docker/           # Docker setup and configurations
```

## CRUD Engine Implementation

The Dynamic CRUD Engine (implemented via `DynamicDataService` and handlers) automates data manipulation for dynamically created services:
- **Endpoints:** `POST /api/v1/data/{slug}`, `GET /api/v1/data/{slug}/{id}`, `GET /api/v1/data/{slug}`, `PUT /api/v1/data/{slug}/{id}`, `DELETE /api/v1/data/{slug}/{id}`.
- **Operations:** Generates queries mapping standard REST payloads to database operations via GORM.
- **Query Engine:** Supports filtering, pagination, and sorting (e.g., `?sort=created_at:desc`, `?page=1&limit=20`), dynamically validating `ORDER BY` fields.

## Schema Migration Engine

Implemented in `MigrationService`, ensuring safe schema evolution:
- Detects field additions, modifications, or removals on `Service` updates.
- Executes DDL commands on the target database automatically.
- Logs migrations to support history tracking and basic rollback functionality based on prior snapshots.

## GraphQL Gateway

Powered by `gqlgen`:
- Automatically builds GraphQL schema corresponding to registered `services`.
- Generates generic queries and mutations that reflect the REST endpoints.
- Hosted at `/api/v1/graphql` with an interactive playground at `/api/v1/graphql/playground`.

## Backup System

Implemented in `BackupService`:
- Provides endpoints to back up generic service data into snapshots (JSON/SQL).
- Manages snapshots for potential migration rollbacks and data recovery tasks.
- API endpoints map to `/api/v1/cms/backup/service/{id}` and `/api/v1/cms/restore/{id}`.

## Observability Integration

Comprehensive observability via industry-standard tooling:
- **Logging**: Zap handles structured JSON request and application logs.
- **Tracing**: OpenTelemetry setup integrates with Jaeger, providing distributed tracing across API calls and database queries.
- **Metrics**: Prometheus metrics intercept requests, tracking latency, response codes, and errors at `/metrics`.

## Unit Testing

Strict unit testing strategy:
- Target of 80% coverage across Repository, Service, and API handler layers.
- Uses standard Go `testing` to prevent dependency download failures.
- Leverages in-memory SQLite (with DSN `:memory:`) via `github.com/glebarez/sqlite` to verify DB-driven services.
- Tested using command: `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`

## Docker Setup

Containerized for scalability and straightforward deployments:
- `Dockerfile`: Multi-stage build producing an optimized binary.
- `docker-compose.yml`: Spawns the CMS platform, underlying databases, Redis for caching, Jaeger for tracing, and Prometheus for metrics collection.

## Swagger Documentation

Auto-generated API documentation:
- Accessible on `/swagger/*any`.
- Generated using `swag`.
- Represents the statically typed administrative endpoints and available routes.
