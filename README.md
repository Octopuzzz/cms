# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend-as-a-Service (BaaS) platform written in Go, acting as a dynamic CMS and API builder. It is designed similarly to platforms like Hasura or Supabase, but is fully self-hosted and Go-native. The platform enables developers to connect databases, create data models dynamically, automatically generate CRUD APIs, generate optional GraphQL endpoints, manage schema migrations, monitor observability, and handle backups.

## 1. System Architecture Explanation

The system follows **Clean Architecture** and **Domain Driven Design (DDD)** principles, structured into the following layers:
- **Presentation Layer**: Handled by the Gin HTTP Router and handlers (`internal/handlers/`). Acts as the gateway for mapping REST and GraphQL requests to internal services.
- **Application Layer**: Contains business logic (`internal/services/`). Manages dynamic service schemas, database connections, backups, and migrations.
- **Domain Layer**: Core data models (`internal/models/`) which represent entities like Services, Fields, Database Connections, Users, and Roles.
- **Infrastructure Layer**: Cross-cutting concerns such as logging (`pkg/logger/` using Zap), metrics (Prometheus), tracing (`internal/tracing/` using OpenTelemetry), and caching.

**High-Level Architecture:**
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

## 2. Metadata Database Schema

The CMS stores platform configuration and metadata in a database (such as PostgreSQL or SQLite). The schema includes the following core entities:
- `database_connections`: Stores configurations for external databases (id, name, type, host, port, username, password, database_name). Supports PostgreSQL, MySQL, MongoDB.
- `services`: Represents user-created data models, linking to a `database_connection_id`, and defining dynamic table names.
- `fields`: Defines attributes for each service, specifying names, types (string, integer, float, boolean, uuid, json, array, timestamp), nullability, defaults, and indices.
- `relations`: Defines one-to-one, one-to-many, many-to-one, and many-to-many relationships, automatically maintaining join tables.
- `migrations`: Tracks executed schema changes for rollback and auditing capabilities.
- `backups`: Stores backup snapshot metadata.

## 3. Go Project Structure

The project uses a clean modular structure:

```text
├── cmd
│   └── server                # Application entrypoint
├── internal
│   ├── config                # Configuration management
│   ├── database              # Database connection pool manager
│   ├── graphql               # GraphQL gateway and schema generation
│   ├── handlers              # API presentation layer (REST/GraphQL)
│   ├── middleware            # HTTP middlewares (Auth, Tracing, Metrics)
│   ├── models                # Domain models and metadata schema
│   ├── services              # Application logic (CRUD, Migrations, CMS, Backups)
│   └── tracing               # OpenTelemetry initialization
├── pkg
│   ├── logger                # Structured Zap logging wrapper
│   └── response              # Standardized API response utilities
├── tests
│   ├── integration           # Integration tests
│   ├── uat                   # User Acceptance Testing
│   └── unit                  # Unit tests
├── docker
│   ├── Dockerfile            # Container definition
│   ├── docker-compose.yml    # Local development cluster
│   └── prometheus.yml        # Prometheus configuration
├── .env.example
├── ARCHITECTURE.md
├── Makefile                  # Automated tasks and build commands
└── go.mod
```

## 4. CRUD Engine Implementation

The **Dynamic CRUD Engine** and **Query Engine** (`internal/services/dynamic_data_service.go`) map REST endpoints directly to GORM operations based on dynamic metadata schemas. When a service is registered, the engine automatically exposes endpoints:
- `POST /api/v1/data/{slug}` (Create)
- `GET /api/v1/data/{slug}/{id}` (Read)
- `PUT /api/v1/data/{slug}/{id}` (Update)
- `DELETE /api/v1/data/{slug}/{id}` (Delete)
- `GET /api/v1/data/{slug}` (List / Query)

**Query Capabilities:**
- **Filtering**: `?email=john@example.com`
- **Sorting**: `?sort=created_at:desc`
- **Pagination**: `?page=1&limit=20`
- **Join Queries**: `?join=user,products` with nested joins support.

## 5. Schema Migration Engine

The **Migration Engine** (`internal/services/migration_service.go`) provides safe, dynamic schema transitions for underlying databases when users update service models.
- **Operations Supported**: Add column, drop column, rename column, change column type.
- **Safety Features**: The engine interfaces with GORM's `.Migrator()`. It allows automatic tracking in the `migrations` table, enabling safe operations and rollback mechanisms.

## 6. GraphQL Gateway

The **GraphQL Gateway** (`internal/graphql/gateway.go`) optionally auto-generates GraphQL schemas for user-defined services utilizing `gqlgen`.
- Exposes `POST /api/v1/graphql` to fulfill automated GraphQL queries.
- Supports dynamically structured queries, mutations, and deeply nested relations mapped to the metadata definitions.

## 7. Backup System

The **Backup Engine** (`internal/services/backup_service.go`) handles automated and manual snapshot capabilities.
- **Operations Supported**: Table backups, schema backups, and entire service snapshots in JSON format.
- Exposes control-plane endpoints like `POST /api/v1/cms/backups` and `POST /api/v1/cms/backups/{id}/restore` to orchestrate backup dumping and restoration processes over dynamic database records.

## 8. Observability Integration

Observability is deeply embedded into the infrastructure layer for monitoring system health and diagnosing errors.
- **Logging**: Structured JSON logging powered by Zap (`pkg/logger/`), capturing request logs, error logs, and query logs with correlation IDs.
- **Metrics**: Exposes Prometheus metrics via `internal/middleware/prometheus.go` at the `/metrics` endpoint, capturing request latencies, error rates, and slow queries.
- **Tracing**: OpenTelemetry (OTel) is initialized in `internal/tracing/` to provide distributed request tracking, wrapping SQL queries and network calls, exportable to Jaeger or standard OTLP collectors.

## 9. Unit Tests

The system maintains a comprehensive test suite targeting >= 80% coverage.
- Contains Unit, Integration, and UAT test directories.
- Repositories, Services, and Handlers are heavily tested.
- Unit tests run utilizing an in-memory SQLite database (`:memory:`) to allow rapid offline execution.
- **Execution Command**: Use `make test-unit` or `go test -v -race ./tests/unit/...`

## 10. Docker Setup

The repository is completely containerized for local development and scalable deployment.
- Located within `docker/`, the setup uses Docker Compose.
- **Included Services**: API Backend, Database, Redis, Prometheus (for metrics), and potentially Jaeger.
- **Commands**: `make docker-up` starts the environment in the background. `docker-build` handles application image compilation.

## 11. Swagger Documentation

The project auto-generates interactive Swagger/OpenAPI 3.0 documentation using `swaggo/swag`.
- Captures the static control plane APIs (`/api/v1/cms/...`) and authentications logic.
- Documentation output is ignored by version control to prevent commit noise, but generated dynamically.
- **Execution Command**: Run `make swagger` or `swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs`.
