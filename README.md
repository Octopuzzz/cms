# Backend Platform (Dynamic CMS + API Builder)

Welcome to the Go-native Backend-as-a-Service (BaaS) platform. This platform allows you to dynamically create backend services, define schemas, generate REST and GraphQL APIs automatically, and manage database connections, schemas, migrations, backups, and observability.

## System Architecture Explanation
The system is built on **Clean Architecture** and **Domain-Driven Design (DDD)** principles. The application is logically divided into four main layers:
- **Presentation Layer**: Handles incoming HTTP and GraphQL requests. Powered by Gin and `gqlgen`.
- **Application Layer**: Contains business logic for the control plane, dynamic CRUD operations, migrations, and service building.
- **Domain Layer**: Houses core entities like Services, Fields, Database Connections, and Users.
- **Infrastructure Layer**: Implements cross-cutting concerns like logging (Zap), metrics (Prometheus), tracing (OpenTelemetry), and external database connections (GORM).

## Metadata Database Schema
The metadata database (PostgreSQL or SQLite) manages platform configuration:
- `database_connections`: Stores details for connecting to external databases (id, name, type, host, credentials).
- `services`: Tracks dynamic data models, linking to a database connection.
- `fields`: Defines attributes for services (name, type, nullable, unique, default_value).
- `relations`: Manages relationships (One-to-One, One-to-Many, Many-to-One, Many-to-Many).
- `migrations`: Logs applied schema changes.
- `backups`: Tracks database snapshots and exports.

## Go Project Structure
```text
.
├── cmd/
│   └── server/          # Application entry point
├── docker/              # Docker and Docker Compose configuration files
├── internal/
│   ├── config/          # Environment configuration
│   ├── database/        # Connection manager and GORM configurations
│   ├── graphql/         # GraphQL gateway (gqlgen)
│   ├── handlers/        # Gin HTTP handlers
│   ├── middleware/      # Auth, rate limiting, and Prometheus
│   ├── models/          # Domain structures (GORM models)
│   ├── services/        # Business logic (CRUD, CMS, migrations, backups)
│   └── tracing/         # OpenTelemetry setup
├── pkg/
│   ├── logger/          # Zap logger
│   └── response/        # Standardized HTTP responses
└── tests/               # Unit, Integration, and UAT tests
```

## CRUD Engine Implementation
The CRUD engine automatically generates REST endpoints for any dynamically created service. It handles:
- **Create/Read/Update/Delete (CRUD)** operations routed dynamically via slug.
- **Query Engine**: Built-in support for filtering (`?email=test@example.com`), sorting (`?sort=created_at:desc`), and pagination (`?page=1&limit=20`).

## Schema Migration Engine
Safe schema migrations are applied dynamically via GORM's AutoMigrate capabilities. It allows for:
- Adding, dropping, or renaming columns dynamically when a Service schema is updated.
- Safely managing rollbacks and history through the internal `migrations` tracking table.

## GraphQL Gateway
A dynamic GraphQL API operates alongside the REST endpoints. Utilizing `gqlgen`, it maps dynamic schemas to queries and mutations, providing a flexible query language over the generated services.

## Backup System
The system includes functionality for database and service-level backups. Backups generate JSON snapshots or database-specific dumps that can be retrieved and restored on demand.

## Observability Integration
The platform implements state-of-the-art observability:
- **Logging**: High-performance structured logging using Zap.
- **Tracing**: OpenTelemetry (OTEL) integration for distributed tracing of requests and database queries.
- **Metrics**: Prometheus middleware metrics integrated into Gin to track request latencies and error rates.

## Unit Tests
The project prioritizes quality with extensive testing covering Handlers, Services, and Repositories.
Run the test suite using standard Go commands:
```bash
go test ./... -cover
```
*Note: Ensure coverage exceeds 80% as required by the CI/CD pipeline.*

## Docker Setup
The platform is fully containerized. Use the provided Docker configuration to spin up the entire stack including the Go application, databases, Prometheus, and external tools:
```bash
docker-compose up -d
```

## Swagger Documentation
Interactive API documentation is generated using `swaggo/swag`. You can browse available REST APIs by navigating to `/swagger/index.html` after generating the docs:
```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```
