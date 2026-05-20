# CMS Backend Platform

## Overview
This is a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go, acting as a self-hosted backend infrastructure generator.

## Architecture
The platform follows Clean Architecture and Domain Driven Design (DDD).

### High Level Architecture
```
       Client Applications
                │
   ┌────────────┴─────────────┐
   ▼                          ▼
REST API                 GraphQL API (optional)
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

### Components
1. **CMS Control Plane**: Management interface to control the system.
2. **Metadata Database**: Stores platform metadata (`database_connections`, `services`, `fields`, `service_permissions`, `migrations`, `backups`, `users`, `roles`, `audit_logs`).
3. **Database Connection Manager**: Manages connections to external user databases (PostgreSQL, MySQL, MongoDB).
4. **Service Builder**: Creates dynamic services/models.
5. **Relation Engine**: Supports One-to-One, One-to-Many, Many-to-One, and Many-to-Many relationships.
6. **Dynamic CRUD Engine**: Generates RESTful CRUD endpoints automatically based on service definitions.
7. **Query Engine**: Supports filtering, sorting, pagination, and join queries.
8. **GraphQL Gateway**: Auto-generates optional GraphQL endpoints.
9. **Schema Migration Engine**: Safely performs database schema modifications (add/drop/rename columns).
10. **Backup Engine**: Handles table backups and snapshots.
11. **Observability**: Zap for logging, Prometheus for metrics, OpenTelemetry for tracing.
12. **Performance Optimization**: Connection pooling, query caching, and more.
13. **Security**: Role-Based Access Control, JWT Auth, Input validation.

## Project Structure
```
.
├── cmd/
│   └── server/          # Main application entry point
├── docker/              # Docker and Prometheus configurations
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # DB connection manager and dialects
│   ├── graphql/         # GraphQL gateway and schemas
│   ├── handlers/        # HTTP controllers/handlers
│   ├── middleware/      # Gin middlewares (Auth, Metrics, etc.)
│   ├── models/          # Core domain models
│   ├── services/        # Business logic and engines (CRUD, Service Builder)
│   └── tracing/         # OpenTelemetry setup
├── pkg/
│   ├── logger/          # Zap structured logger
│   └── response/        # Standardized API response format
├── tests/               # Unit, Integration, and UAT tests
└── docs/                # Auto-generated Swagger documentation
```

## Running the Application

### Using Docker (Recommended)
The platform is containerized and can be started via Docker Compose:
```bash
make docker-up
```
This starts the Go backend, PostgreSQL (metadata), Redis (cache), Jaeger (tracing), and Prometheus (metrics).

### Local Development
To run locally with hot-reload (requires `air`):
```bash
make dev
```
Or build and run standard:
```bash
make build
make run
```

### Testing
Run all unit, integration, and UAT tests:
```bash
make test
```
Or individually:
```bash
make test-unit
make test-integration
make test-uat
```
To check coverage:
```bash
make test-coverage
```

### API Documentation
Swagger is used for REST documentation. Run:
```bash
make swagger
```
The documentation will be available at `http://localhost:8080/swagger/index.html`.
