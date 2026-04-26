# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade, self-hosted, Go-native Backend-as-a-Service platform. It allows users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

## System Architecture Explanation

The CMS Backend follows Clean Architecture and Domain-Driven Design (DDD) principles:

- **Presentation Layer**: Gin HTTP Router and Handlers. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Controls data access, generates dynamic schemas, and performs operations requested by handlers.
- **Domain Layer**: Data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (OpenTelemetry), metrics (Prometheus), and logging (Zap).

## Metadata Database Schema

The core metadata database stores platform configuration and dynamically created schemas. Tables include:

- `database_connections`: External DB configurations (PostgreSQL, MySQL, MongoDB, SQLite).
- `services`: User-created data models (services).
- `fields`: Attributes for each service (string, integer, UUID, JSON, relations).
- `users`, `roles`, `role_permissions`, `user_roles`: RBAC authentication.
- `migrations`: Tracks DDL executions and schema history.
- `backups`: Stores backup snapshots.
- `custom_validations`: Field-level validation rules.

## Go Project Structure

```
├── cmd/
│   └── server/          # Entry point (main.go)
├── docker/              # Docker Compose and Prometheus configs
├── internal/
│   ├── config/          # Environment configuration
│   ├── database/        # Connection pooling and management
│   ├── graphql/         # GraphQL Gateway
│   ├── handlers/        # Gin REST API Handlers
│   ├── middleware/      # Auth, Rate Limiter, Prometheus
│   ├── models/          # Core Domain Models
│   ├── services/        # Business Logic & Core Engines
│   └── tracing/         # OpenTelemetry Setup
├── pkg/
│   ├── logger/          # Structured Zap Logging
│   └── response/        # Standardized API Responses
├── tests/               # Unit, Integration, and UAT tests
├── Dockerfile           # Docker container configuration
└── Makefile             # Make targets (build, test, run)
```

## Core Engines

- **CRUD Engine**: Automatically generates generic `Create`, `Read`, `Update`, `Delete`, and `List` endpoints (`internal/services/dynamic_data_service.go`).
- **Schema Migration Engine**: Supports safe schema changes (Add/Drop/Rename Column) with `gorm.Migrator` and rollback capability (`internal/services/migration_service.go`).
- **GraphQL Gateway**: Dynamically generates GraphQL AST using `gqlgen` from service schemas (`internal/graphql/gateway.go`).
- **Backup System**: Supports snapshot backups and restoration of dynamic tables (`internal/services/backup_service.go`).
- **Observability**: Prometheus metrics, Jaeger distributed tracing, and structured Zap logging are fully integrated.

## Getting Started

### Docker Setup

To run the full stack locally:

```bash
docker-compose up -d
```

### Swagger Documentation

The project includes automatically generated OpenAPI/Swagger documentation. To view it, start the server and navigate to:

```
http://localhost:8080/swagger/index.html
```

To regenerate Swagger docs after making changes, run:

```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

### Running Tests

The platform includes Unit, Integration, and User Acceptance Testing (UAT). Minimum code coverage is 80%.

```bash
make test-coverage
```

Or directly using Go:

```bash
go test ./...
```
