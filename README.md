# CMS Backend Platform

This repository contains a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. It acts as a backend infrastructure generator, allowing developers to connect databases, create data models dynamically, generate CRUD APIs automatically, generate optional GraphQL APIs, and manage schema migrations, backups, and observability.

The system is fully self-hosted, Go-native, modular, scalable, and cloud-ready, functioning similarly to platforms like Hasura or Supabase.

## System Architecture

The system follows **Clean Architecture and Domain-Driven Design (DDD)**.

### Core Layers:
- **Presentation Layer**: Exposes REST and GraphQL APIs using the Gin framework and `gqlgen` (`internal/handlers`, `internal/graphql`).
- **Application Layer**: Contains the core business logic (`internal/services`). Services act as the engines (CMS, CRUD, Migration, etc.).
- **Domain Layer**: Defines core data structures and models (`internal/models`).
- **Infrastructure Layer**: Manages cross-cutting concerns like databases, logging, metrics, and tracing (`internal/database`, `pkg/logger`, `internal/tracing`, `internal/middleware`).

### High-Level Components
1. **CMS Control Plane**: Management interface to handle database connections, services, schemas, relations, migrations, and backups (`internal/services/service_service.go`, `internal/services/dbconn_service.go`).
2. **Metadata Database**: Uses PostgreSQL (or SQLite/MySQL depending on configuration) to store platform metadata.
3. **Database Connection Manager**: Connects to external databases (PostgreSQL, MySQL, MongoDB) and manages connection pooling (`internal/database/connection_manager.go`).
4. **Service Builder**: Dynamically creates services (data models) and defines fields, relations, constraints, and validation rules.
5. **Relation Engine**: Supports relationships (One-to-One, One-to-Many, Many-to-One, Many-to-Many) by mapping and auto-generating queries.
6. **Dynamic CRUD Engine**: Generates standard endpoints (Create, Read, Update, Delete, List) via `internal/services/dynamic_data_service.go`.
7. **Query Engine**: Enables advanced dynamic querying including filtering, sorting, pagination, and joining across services.
8. **GraphQL Gateway**: Automatically generates GraphQL schema and provides a unified endpoint (`internal/graphql/gateway.go`).
9. **Schema Migration Engine**: Ensures safe DDL (Add, Drop, Rename Column) changes mapped securely in tracking tables (`internal/services/migration_service.go`).
10. **Backup Engine**: Orchestrates backup and restore procedures capturing service snapshots or table exports (`internal/services/backup_service.go`).
11. **Observability Engine**: Centralized logging (Zap), metrics tracking (Prometheus), and distributed tracing (OpenTelemetry).

## Metadata Database Schema

The core metadata models are defined in `internal/models/models.go` and include:

- **`database_connections`**: Stores external database configurations (`id`, `name`, `type`, `host`, `port`, `username`, `password`, `database`, `ssl_mode`, pooling configurations).
- **`services`**: Represents dynamic data models (`id`, `name`, `database_connection_id`, `db_table_name`, `created_at`, `updated_at`).
- **`fields`**: Columns or properties in a service (`id`, `name`, `service_id`, `type`, `nullable`, `unique`, `default_value`).
- **`service_permissions`**: Maps fine-grained access checks between roles and services.
- **`migrations`**: Audits and stores history of schema adjustments.
- **`backups`**: Stores record references for system data snapshots.
- **`users` & `roles`**: RBAC configurations for the control plane APIs.
- **`audit_logs`**: Tracks internal modifications securely.

## Go Project Structure

The project follows a clean Go layout:

```text
├── cmd
│   └── server                # Application entry point (main.go)
├── internal
│   ├── api                   # API layer configuration
│   ├── config                # Global application configuration
│   ├── database              # DB connection manager and GORM setup
│   ├── graphql               # gqlgen integration and optional GraphQL Gateway
│   ├── handlers              # Gin HTTP handlers
│   ├── middleware            # Auth, Telemetry, and validation interceptors
│   ├── models                # Metadata database domain models
│   ├── services              # Application logic engines (CRUD, Backup, Migration, CMS)
│   └── tracing               # OpenTelemetry implementation
├── pkg
│   └── logger                # Shared zap logger package
├── tests
│   ├── integration           # Integration tests
│   ├── uat                   # E2E / User Acceptance Tests
│   └── unit                  # Standard unit tests
├── docker                    # Dockerfiles, Compose setups, and configurations
├── docs                      # Swagger OpenAPI specifications
├── Makefile                  # Tasks for build, test, run, and dev workflows
└── README.md                 # This file
```

## CRUD Engine Implementation

Handled by `internal/services/dynamic_data_service.go` and `internal/handlers/dynamic_data_handler.go`.
- Automatically mounts endpoints like `/api/v1/data/{slug}`.
- Generates GORM SQL queries dynamically by converting REST requests and query parameters (e.g., `?limit=10&page=2&sort=created_at:desc&email=user@example.com`) into underlying database operations.
- Automatically handles join statements for related tables mapped in the `Relation Engine`.
- Performs dynamic row-level RBAC checks based on user roles and `service_permissions`.

## Schema Migration Engine

Implemented in `internal/services/migration_service.go`.
- Translates `Service` and `Field` updates (received by the CMS) into underlying GORM `.Migrator()` commands.
- Supports `Add Column`, `Drop Column`, `Change Type`, and table generation on the target dynamic database.
- Tracks migration statuses and handles rollbacks to preserve database integrity during failures.

## GraphQL Gateway

Implemented in `internal/graphql/gateway.go` utilizing `gqlgen`.
- Mounts at `POST /api/v1/graphql`.
- Generates GraphQL representations on the fly by resolving mapped tables.
- Supports recursive relation queries allowing deep nesting of output JSON for linked services (e.g. users and roles).

## Backup System

Implemented in `internal/services/backup_service.go` and `internal/handlers/backup_handler.go`.
- Facilitates backup and restore processes via `/api/v1/cms/backup` routes.
- Serializes rows securely using generic struct implementations (`[]map[string]any`).
- Implements Snapshot creation for services ensuring data continuity without strict SQL DDL exports.

## Observability Integration

1. **Logging**: Implemented using `go.uber.org/zap` in `pkg/logger/logger.go`. Provides structured JSON logs and correlates them with Request/Trace IDs.
2. **Metrics**: Implemented natively in Go using `github.com/prometheus/client_golang` (`internal/middleware/prometheus.go`). Generates metrics like `http_request_duration_seconds` and `http_requests_total`. Exposes them at `/metrics`.
3. **Tracing**: Implemented using OpenTelemetry (`internal/tracing/tracing.go`). Exports spans to standard OTLP endpoints (e.g. Jaeger or Zipkin). Auto-instruments HTTP handlers via OpenTelemetry Gin middleware.

## Unit Tests

The system maintains 80%+ test coverage. Tests are separated logically:
- Unit (`tests/unit/`): Validates algorithms, mapping, and isolated database logic. Uses SQLite in-memory databases (`:memory:`).
- Integration (`tests/integration/`): Tests full internal components without the network layer.
- UAT (`tests/uat/`): E2E API tests hitting endpoints and verifying full system interaction.

Run tests using the included Makefile:
```bash
make test          # Runs all tests
make test-unit     # Runs unit tests specifically
make test-coverage # Generates coverage output
```

## Docker Setup

The repository is containerization ready.
- **Dockerfile**: Located in `docker/Dockerfile`, implements a multi-stage optimized Go build (`alpine` based image) for production deployment.
- **docker-compose.yml**: Provided in `docker/docker-compose.yml`, initializes the CMS Application, a PostgreSQL backend for Metadata, and an accompanying Prometheus server for monitoring setup.

Run the environment using:
```bash
make docker-up
```

## Swagger Documentation

API Documentation is auto-generated using standard Swaggo comments.
- Source annotations exist directly on HTTP Handlers (e.g. `@Summary`, `@Description`, `@Tags`).
- Re-generate documentation running: `make swagger`
- The `docs/` folder contains generated openapi/swagger JSON and YAML setups mapped via `/swagger/*any`.
