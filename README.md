# Dynamic CMS + API Builder Platform

A production-grade Backend-as-a-Service (BaaS) platform written in Go, acting as a dynamic backend infrastructure generator. This system allows users to dynamically create backend services, schemas, REST/GraphQL APIs, manage schema migrations, monitor logs/performance, and manage backups. It is fully self-hosted, modular, scalable, and cloud-ready.

## System Architecture Explanation

The platform follows **Clean Architecture and Domain-Driven Design (DDD)**.
Core layers include:
- **Presentation Layer:** REST (Gin) and optional GraphQL (gqlgen) endpoints exposed via API Gateway.
- **Application Layer:** Business logic handling data access, dynamic schemas, and handler requests.
- **Domain Layer:** Core entities (Service, Field, DatabaseConnection, User, Role).
- **Infrastructure Layer:** Database connectors, connection caching, telemetry (OpenTelemetry), metrics (Prometheus), and logging (Zap).

### High-Level Platform Architecture

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

Platform metadata is stored securely in a relational database (PostgreSQL recommended). Core tables:

- **databases:** Stores external DB configs (id, name, type, host, port, credentials).
- **services:** Represents user-created data models, linking to `database_connection_id` and tracking table names.
- **fields:** Attributes for each service (id, name, type, nullable, unique, default_value).
- **relations / service_permissions:** Connects Roles to Services for fine-grained access (RBAC).
- **migrations:** History of executed DDL operations for schema safety.
- **backups:** Snapshot records and schema outputs.
- **users / roles:** Authentication and authorization for CMS Control Plane.
- **audit_logs:** Detailed logging of modifications.

## Go Project Structure

Clean modular structure:
```
cmd/
└── server/             # Entry point
internal/
├── config/             # Configuration management
├── database/           # Connection manager and DB pooling
├── graphql/            # GraphQL gateway implementation
├── handlers/           # REST API handlers
├── middleware/         # Auth, Prometheus, Rate Limiter
├── models/             # Domain layer entities
├── services/           # Application layer logic
└── tracing/            # OpenTelemetry setup
pkg/
├── logger/             # Zap structured logging
└── response/           # Standardized API responses
tests/                  # Unit, Integration, and UAT tests
docker/                 # Docker setup
```

## CRUD Engine Implementation

Automatically generates RESTful endpoints upon service creation using `DynamicDataService`. Endpoints support:
- `POST /api/v1/data/{slug}`
- `GET /api/v1/data/{slug}/{id}`
- `PUT /api/v1/data/{slug}/{id}`
- `DELETE /api/v1/data/{slug}/{id}`
- `GET /api/v1/data/{slug}`

It uses dynamic GORM logic to validate schema metadata, apply automated filtering via query params, perform dynamic join mapping for relational queries, and handle pagination.

## Schema Migration Engine

Powered by `migration_service`, tracks and safely applies structural differences (add/drop/rename columns, alter types) via GORM’s Migrator. Features include tracking state in `migrations` table and snapshot-based rollbacks to maintain schema safety.

## GraphQL Gateway (Optional)

An automatically generated GraphQL API powered by `gqlgen`. Exposes `POST /api/v1/graphql`, routing generic optional queries and mutations matching user-defined service data models directly into backend logic.

## Backup System

Handles automated and manual backups via `backup_service`. Generates data snapshots matching dynamic table structures to robust JSON outputs. Supports precise table restoration and entire structural rollbacks.

## Observability Integration

Comprehensive observability for distributed monitoring:
- **Logging:** Zap-based request, error, and query logs with correlation IDs.
- **Metrics:** Prometheus endpoint (`/metrics`) capturing request latency, slow queries, and error rates.
- **Tracing:** OpenTelemetry exported to Jaeger, tracking SQL calls and HTTP lifecycles in real-time.

## Unit Tests & Testing Strategy

Covering Repository, Service, and API handler layers, the testing strategy guarantees robust core functionalities. Built on an in-memory SQLite setup via `github.com/glebarez/sqlite`, allowing quick tests.

Execute suite:
```bash
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Docker Setup & Containerization

Docker and `docker-compose.yml` natively support containerized deployment, wrapping the Golang binary along with foundational infrastructure dependencies like Prometheus and PostgreSQL.

## Swagger Documentation

Swagger/OpenAPI documentation is auto-generated using `swag`. It exposes route structures securely while keeping API footprints dynamically up to date.
