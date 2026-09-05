# Backend Platform (Dynamic CMS + API Builder)

This repository contains a **production-grade Backend Platform (Dynamic CMS + API Builder) written in Go**. This platform functions similarly to solutions like Hasura or Supabase, acting as a backend infrastructure generator that is fully self-hosted, Go-native, modular, scalable, and cloud-ready.

## High-Level Platform Architecture

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
- **Presentation Layer**: The API layer built with Gin Web Framework (REST) and `gqlgen` (GraphQL).
- **Application Layer**: Contains business logic (`internal/services`).
- **Domain Layer**: Contains models and data structures (`internal/models`).
- **Infrastructure Layer**: Incorporates Database Adapters, Caching, Migrations, etc.

---

## Metadata Database Schema

The CMS manages internal metadata. The primary database used is PostgreSQL, storing configurations for connected databases, created services, schema fields, and more.

Key metadata entities include:
1. `database_connections` - Users register external DBs.
2. `services` - Dynamically created data models (tables).
3. `fields` - Configuration of attributes within services.
4. `relations` - Linking models with support for One-to-One, One-to-Many, Many-to-One, and Many-to-Many.
5. `migrations` - Log of schema changes.
6. `backups` - Snapshots of tables, schemas, or service endpoints.
7. `audit_logs` - Detailed logging of metadata modifications.

*A detailed view of these tables is available inside `internal/models/models.go`.*

---

## Go Project Structure

The project employs a clean modular structure.

```
.
├── cmd/
│   └── server/
│       └── main.go                 # Entry point for the platform
├── docker/
│   ├── Dockerfile                  # Container instructions
│   ├── docker-compose.yml          # Local infrastructure via docker-compose
│   └── prometheus.yml              # Prometheus config for observability
├── internal/
│   ├── config/                     # Configuration and Environment variable handling
│   ├── database/                   # Connection Manager & Connection Pooling
│   ├── graphql/                    # Optional GraphQL Gateway using gqlgen
│   ├── handlers/                   # Presentation Layer HTTP/REST API endpoints
│   ├── middleware/                 # Rate Limiter, Authentication, Prometheus
│   ├── models/                     # Domain Data Models (Metadata schema definitions)
│   ├── services/                   # Application Layer Business Logic
│   │   ├── auth_service.go
│   │   ├── backup_service.go
│   │   ├── dynamic_data_service.go # Core CRUD Engine
│   │   ├── migration_service.go
│   │   ├── service_service.go      # Service Builder
│   │   └── user_service.go         # Identity & Access Management
│   └── tracing/                    # OpenTelemetry Tracing
├── pkg/
│   ├── logger/                     # Zap Structured Logging
│   └── response/                   # Standardized JSON response formatting
├── tests/
│   ├── integration/                # Database/API Integration tests
│   ├── uat/                        # User Acceptance Testing
│   └── unit/                       # Core component tests
├── Makefile                        # Command shortcuts
└── README.md                       # This file
```

---

## CRUD Engine & Query Engine

When a service is created, the system auto-generates REST endpoints using the `DynamicDataService` (`internal/services/dynamic_data_service.go`).

Supported Endpoints:
- `POST /api/v1/data/{slug}`
- `GET /api/v1/data/{slug}/{id}`
- `PUT /api/v1/data/{slug}/{id}`
- `DELETE /api/v1/data/{slug}/{id}`
- `GET /api/v1/data/{slug}` (Lists, Filters, Pagination, Joins, Sorting)

The query engine dynamically converts JSON parameters into performant SQL queries.

---

## Schema Migration Engine

Implemented in `internal/services/migration_service.go`, this engine provides safe, tracked schema changes on user-created services. It monitors field changes and alters table schemas accordingly, keeping a persistent track record in the `migrations` metadata table.

---

## GraphQL Gateway

An automated, optional GraphQL gateway runs over dynamically defined schemas utilizing the `gqlgen` library (`internal/graphql/gateway.go`). The system exposes a `/api/v1/graphql` endpoint translating GraphQL queries into backend REST or DB operations seamlessly.

---

## Backup System

The Platform handles snapshots and restores via the `internal/services/backup_service.go`. It can serialize data and structural schema into JSON or SQL dumps. A robust tracking table maintains historical snapshot records for reliable rollbacks.

---

## Observability

State-of-the-art telemetry integration comes built-in:
- **Logging**: Zap structured logging is standard across handlers and services (`pkg/logger`).
- **Metrics**: Prometheus hooks capture essential request latency, success rates, and handler metrics (`internal/middleware/prometheus.go`).
- **Tracing**: OpenTelemetry (`internal/tracing/tracing.go`) provides distributed tracing of queries and endpoints for deep performance analysis.

---

## Docker Setup

The system is easily run and scalable using Docker.
- A `Dockerfile` compiles the Go binary via a multi-stage, secure builder.
- The `docker-compose.yml` spins up an instance of the Backend Platform along with a PostgreSQL metadata store, Redis cache, and Prometheus for observing metrics out-of-the-box.

---

## Swagger Documentation

Auto-generated OpenAPI/Swagger documentation makes API consumption easy. The codebase provides annotations around handler functions compatible with the `swaggo/swag` CLI.

**Run**: `make swagger` to generate the `docs/` folder, then view locally via standard Swagger UI components.

---

## Unit Testing

Strict guidelines dictate minimum 80% coverage for the `internal/...` application and handlers. Tests are split into `unit`, `integration`, and `uat` layers, heavily relying on fast, parallelizable, `go test` capabilities paired with an in-memory SQLite metadata store (`tests/unit/services_test.go`).

**Run**: `make test-coverage` to run tests and output `coverage.html`.
