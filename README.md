# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade, self-hosted Backend-as-a-Service platform written in Go. It enables dynamic creation of data models, database connections, and automatically generates REST and optional GraphQL APIs for these models.

## System Architecture Explanation

The platform follows **Clean Architecture** and **Domain-Driven Design (DDD)**.
- **Presentation Layer**: Handlers located in `internal/handlers/` acting as the API Gateway mappings, managing REST endpoints (via Gin) and GraphQL endpoints (via `gqlgen`).
- **Application Layer**: Services in `internal/services/` that encapsulate core business logic, like dynamically building schemas (`servicebuilder`), controlling data access (`crudengine`, `queryengine`), schema migrations, and backups.
- **Domain Layer**: Models defined in `internal/models/`, defining platform metadata and configurations such as `Service`, `Field`, `DatabaseConnection`, `Role`, etc.
- **Infrastructure Layer**: Utilities across `pkg/logger/`, `internal/tracing/`, and external connection integrations to power telemetry, observability, and caching.

```text
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

## Metadata Database Schema

The CMS manages internal states through a relational database schema.
Key tables:
- **`database_connections`**: ID, Name, Type (PostgreSQL, MySQL, MongoDB), Host, Port, Username, Password.
- **`services`**: ID, Name, `database_connection_id`, `db_table_name`, Timestamps.
- **`fields`**: ID, `service_id`, Name, Type, Nullable, Unique, `default_value`.
- **`relations`**: Configurations mapped between services to facilitate joins.
- **`migrations`**: Tracking schema alterations execution logs.
- **`backups`**: Records data and schema snapshot info.

## Go Project Structure

```
cmd/
└── server/             # Application entrypoint
internal/
├── config/             # Environment & configuration loading
├── database/           # Connection pooling & logic for internal/external DBs
├── graphql/            # Gateway implementation & gqlgen logic
├── handlers/           # HTTP controllers
├── middleware/         # Auth, RBAC, telemetry middleware
├── models/             # Domain definitions and DB schema definitions
├── services/           # Service Builder, CRUD Engine, Schema Migration, Backup Engine
└── tracing/            # OpenTelemetry setup
pkg/
├── logger/             # Zap structured logging wrapper
└── response/           # Pagination, error formatting and standard responses
tests/                  # Unit, Integration, and UAT specifications
docker/                 # Build configurations
```

## CMS Control Plane & Service Builder
Users define connection parameters via `DatabaseConnection` service endpoints.
The system registers dynamically constructed `services`, automatically capturing field and relation settings (string, integer, UUID, boolean, references) through its internal Metadata API.
Endpoints (examples): `POST /api/v1/cms/databases`, `POST /api/v1/cms/services`

## CRUD & Query Engine Implementation
The dynamic data service parses HTTP requests like `GET /api/v1/data/{service}?email=test&sort=created_at:desc&page=1&limit=20` and safely translates them to SQL via GORM mapping. It incorporates nested joins automatically where many-to-many relationship tables (`user_roles`) are discovered.

## Schema Migration Engine
Safe data transformations trigger on service model modifications. Internally executing `.Migrator()` against target schemas allows safe Add, Drop, Rename operations. Changes log securely inside the `migrations` system tracker table to orchestrate eventual rollbacks.

## GraphQL Gateway
Powered by `github.com/99designs/gqlgen`. Generic resolvers bridge GraphQL requests directly to the dynamic service schemas dynamically, translating operations like `query { users { id, name } }` over to the common internal query engine backend mapping.

## Backup System
Supports generating and restoring table backups, capturing structured records for generic JSON payloads inside internal service definitions and schemas via `POST /api/v1/cms/backup/service/{service_id}`.

## Observability Integration
The core logic injects telemetry:
- **Logging**: Zap handles deeply structured context aware request/response/query logs.
- **Metrics**: Prometheus hooks natively to HTTP tracking latency, error rates, slow queries.
- **Tracing**: OpenTelemetry monitors spanning distributed logic layers and DB calls.

## Unit Tests
Tests target Handlers, Repository, and Service layers aiming for >80% code coverage.
You can run testing suites with coverage reports using the built-in tooling:
```bash
make test-coverage
# Underlying execution: go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

## Docker Setup
The environment supports complete containerization out-of-the-box leveraging `Dockerfile` for the Go Application server and `docker-compose.yml` to spool infrastructure services like Redis, Prometheus, Postgres internally. Launch via:
```bash
make docker-up
```

## Swagger Documentation
APIs conform to OpenAPI specifications. Generate docs effortlessly:
```bash
make swagger
# Underlying execution: swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

## Security
- Implement JWT token authorization flow inside middlewares.
- Protect endpoints natively against SQL injection through explicit parameterized generic query builders.
- Fine-grained RBAC validation handles schema permissions natively at the DB level.
