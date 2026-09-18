# Dynamic CMS + API Builder (Backend Platform)

This project is a **production-grade Backend Platform (Dynamic CMS + API Builder) written in Go**. It acts as a self-hosted, Go-native Backend-as-a-Service (BaaS) platform, similar to Hasura or Supabase. It allows users to dynamically create backend services, database schemas, and automatically generates REST and GraphQL APIs.

The system is designed to be **modular, scalable, and cloud-ready**, following **Clean Architecture and Domain-Driven Design (DDD)**.

## Core Features
*   **CMS Control Plane**: Manage database connections, create services, define schemas, and handle relations.
*   **Database Connection Manager**: Connect to external PostgreSQL, MySQL, and MongoDB databases.
*   **Service Builder**: Dynamically create data models with various field types and constraints.
*   **Relation Engine**: Support for One-to-One, One-to-Many, Many-to-One, and Many-to-Many relationships.
*   **Dynamic CRUD Engine**: Automatically generate RESTful API endpoints for created services.
*   **Query Engine**: Advanced querying capabilities including filtering, sorting, pagination, and joins.
*   **GraphQL Gateway (Optional)**: Automatically generated GraphQL queries and mutations.
*   **Schema Migration Engine**: Safe schema changes with rollback capabilities.
*   **Backup Engine**: Support for database, schema, and service backups and restoration.
*   **Observability**: Integrated structured logging (Zap), metrics (Prometheus), and tracing (OpenTelemetry/Jaeger).
*   **Performance & Security**: Connection pooling, query caching (Redis), rate limiting, JWT authentication, and SQL injection protection.

## System Architecture

The platform architecture follows Clean Architecture with distinct layers:

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

*   **Presentation Layer**: REST handlers (Gin) and GraphQL endpoints (gqlgen).
*   **Application Layer**: Core business logic and service orchestration.
*   **Domain Layer**: System entities, metadata schemas, and configuration.
*   **Infrastructure Layer**: Database connectors (GORM), caching (Redis), messaging, metrics, tracing, and logging.

## Metadata Database Schema

The platform metadata is stored in a relational database (default: PostgreSQL). Core tables include:

*   **`users` / `roles`**: System authentication and RBAC.
*   **`database_connections`**: Stores configurations for external data sources.
*   **`services`**: Represents dynamic data models created by users.
*   **`fields`**: Defines schema columns for services (type, constraints).
*   **`service_permissions`**: Maps roles to specific service operations.
*   **`migrations`**: Audit log of applied schema changes.
*   **`backups`**: Records of database or schema snapshots.

## Project Structure

```text
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── config/          # Environment configuration
│   ├── database/        # Database Connection Manager and initialization
│   ├── graphql/         # GraphQL Gateway and schema generation
│   ├── handlers/        # Gin HTTP handlers (REST API)
│   ├── middleware/      # Auth, CORS, rate limiting, metrics
│   ├── models/          # Domain layer schemas
│   ├── services/        # Application logic (CRUD Engine, Service Builder, Migrations)
│   └── tracing/         # OpenTelemetry setup
├── pkg/
│   ├── logger/          # Zap structured logging wrapper
│   └── response/        # Standardized HTTP response formatter
├── tests/               # Unit, Integration, and UAT test suites
├── docker/              # Deployment manifests
├── Makefile             # Build and operational tasks
└── README.md            # You are here
```

## Quick Start

### Prerequisites
*   Go 1.24+
*   Docker & Docker Compose (for supporting infrastructure)

### Run with Docker

1.  Start the infrastructure (PostgreSQL, Redis, Jaeger, Prometheus):
    ```bash
    make docker-up
    ```
2.  Start the application:
    ```bash
    make run
    ```

### Development

*   **Build**: `make build`
*   **Dev Mode (Hot Reload)**: `make dev`
*   **Run Tests**: `make test-coverage`
*   **Generate Swagger Docs**: `make swagger`

### Observability

*   **Prometheus Metrics**: `http://localhost:8080/metrics`
*   **Swagger Documentation**: `http://localhost:8080/swagger/index.html`
*   **GraphQL Playground**: `http://localhost:8080/api/v1/graphql`
