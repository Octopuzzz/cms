# Dynamic CMS & API Builder Platform

## Overview
A production-grade Backend-as-a-Service platform (Dynamic CMS + API Builder) written in Go. This system acts as a backend infrastructure generator, allowing users to dynamically create backend services, schemas, REST/GraphQL APIs, and manage schemas securely. It's built for scale and extensibility natively in Go, incorporating industry best practices in observability, architecture, and deployment.

## System Architecture

The platform is designed around **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: The API layer is powered by Gin (REST) and `gqlgen` (GraphQL).
- **Application Layer**: Business logic handling orchestration (Services, Generators).
- **Domain Layer**: Core platform entities (Services, Connections, Fields, Roles).
- **Infrastructure Layer**: Cross-cutting capabilities (Tracing, Metrics, Logging, Database Connectors).

### High-Level Architecture Diagram
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

## Metadata Database Schema
The platform leverages a central metadata database to persist configuration state. Key entities include:

- `database_connections`: Configuration for external data stores (id, type, host, credentials, pooling logic).
- `services`: Dynamic service definitions linked to a connection (id, name, table mapping).
- `fields`: Schema definition details (type, nullable, unique, default values) bound to services.
- `service_permissions`: RBAC implementation mapping roles to CRUD operations on specific services.
- `migrations`: DDL change tracking and history.
- `backups`: Records of generated database or schema backups.
- `users` & `roles`: Core platform authentication.
- `audit_logs`: Tracking modifications for compliance.

## Project Structure
```
cmd/
  server/         # Main application entry point
internal/
  config/         # Configuration and env parsing
  database/       # DB Connection manager and adapters
  graphql/        # GraphQL gateway and auto-resolvers
  handlers/       # HTTP/REST presentation layer
  middleware/     # Auth, Observability, Rate Limiting
  models/         # Domain entity definitions
  services/       # Application logic (Control Plane, Builders, Engines)
  tracing/        # OpenTelemetry initialization
pkg/
  logger/         # Zap structured logging wrapper
  response/       # Standardized HTTP JSON responses
tests/
  unit/           # Isolated component tests
  integration/    # End-to-end component interaction
  uat/            # User Acceptance Testing
docker/           # Infrastructure scaffolding (Docker, Prometheus)
docs/             # Auto-generated Swagger documentation
```

## Component Overview

### CMS Control Plane
Exposes `/api/v1/cms/` routes to manage database connections, schema updates, access control, and other administrative workflows.

### Service Builder & CRUD Engine
Users define new data services and their fields. The system maps these into physical database tables via the underlying GORM models. Generates automatic endpoints under `/api/v1/data/{slug}`.

### Query & Relation Engine
- Dynamically parses query strings (e.g., `?email=test@example.com&sort=created_at:desc&page=1&limit=20`).
- Resolves foreign keys to automatically build JOIN queries and populate relationships dynamically based on the underlying schema.

### Schema Migration Engine
Safe data-definition transitions handled programmatically via GORM’s Migrator. Automatically records states into the metadata DB to maintain a rollback history if required.

### GraphQL Gateway
Exposes a single `/api/v1/graphql` endpoint. Dynamically generates types and resolvers in memory utilizing `gqlgen` to mirror the defined service schema, enabling complex object graphing without writing boilerplate.

### Backup System
Supports generating JSON/SQL representations of dynamically generated service tables for safety and point-in-time recovery via administrative endpoints.

### Observability Engine
The platform has deep native integration with:
- **Prometheus**: Real-time operational metrics for application throughput/latency.
- **OpenTelemetry**: Trace propagation across internal system boundaries and outbound DB queries.
- **Zap Logging**: High-performance structured JSON logging.

### Docker & Deployment
Fully containerized setup utilizing `docker-compose`. Brings up the API Server, Prometheus, and any necessary dependent services effortlessly. Includes built-in multi-stage Dockerfiles.

### Unit Testing & CI
Testing strategy built into the core design with coverage across Handlers, Services, and DB implementations utilizing standard Go tooling and in-memory SQLite instances. Run tests locally utilizing `make test-coverage`.
