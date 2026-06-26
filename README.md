# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. It works similarly to platforms like Hasura or Supabase but is fully self-hosted and Go-native. The platform allows developers to dynamically create backend services, manage database connections, define schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

The system is designed to be modular, scalable, and cloud-ready, following Clean Architecture and Domain-Driven Design (DDD).

---

## 1. System Architecture Explanation

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: The Gin HTTP Router and Handlers. Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (Services). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models. Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry, metrics (Prometheus), and logging.

### High Level Architecture Diagram
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

---

## 2. Metadata Database Schema

The CMS stores platform metadata in a dedicated PostgreSQL database. The core internal configuration tables include:

- `database_connections`: Stores external DB configurations (id, name, type, host, port, username, password, database_name).
- `services`: Represents user-created data models. Contains a reference to `database_connection_id`, `name`, and tracks dynamic schemas (`db_table_name`).
- `fields`: Defines attributes for each service (id, name, type, nullable, unique, default_value, index). Supports types like string, integer, float, boolean, uuid, json, array, timestamp.
- `service_permissions`: Connects `Role` to `Service` providing fine-grained access checks.
- `migrations`: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
- `backups`: Stores snapshot records or schema outputs.
- `users` and `roles`: General authentication and authorization for the control plane.
- `audit_logs`: Detailed logging of structural and data-level modifications.

---

## 3. Go Project Structure

The project uses a clean modular structure:

```
.
├── cmd
│   └── server
│       └── main.go               # Entry point of the application
├── docker
│   ├── Dockerfile                # Multi-stage Docker build
│   ├── docker-compose.yml        # Docker Compose for complete platform deployment
│   └── prometheus.yml            # Prometheus configuration
├── internal
│   ├── config
│   │   └── config.go             # Environment and configuration loading
│   ├── database
│   │   └── connection_manager.go # External DB connection pooling
│   ├── graphql
│   │   └── gateway.go            # Dynamically generated GraphQL API
│   ├── handlers
│   │   └── ...                   # Gin HTTP handlers for APIs
│   ├── middleware
│   │   ├── auth.go               # JWT Authentication
│   │   ├── prometheus.go         # Metrics tracking
│   │   └── rate_limiter.go       # API rate limiting
│   ├── models
│   │   └── models.go             # Metadata models
│   ├── services
│   │   ├── auth_service.go       # Auth and JWT logic
│   │   ├── backup_service.go     # Backup Engine implementation
│   │   ├── dynamic_data_service.go # CRUD Engine and Query Engine
│   │   ├── migration_service.go  # Schema Migration Engine
│   │   ├── service_service.go    # Service Builder
│   │   └── user_service.go       # Users and Roles
│   └── tracing
│       └── tracing.go            # OpenTelemetry integration
├── pkg
│   ├── logger
│   │   └── logger.go             # Zap structured logging
│   └── response
│       └── response.go           # Standardized API responses
└── tests
    ├── integration
    ├── uat
    └── unit
```

---

## 4. CRUD Engine Implementation

The Dynamic CRUD Engine automatically generates operations (Create, Read, Update, Delete, List) for dynamically created services.

When a service is created, the system maps REST endpoints to GORM database operations on the fly:
- `POST /api/v1/data/{slug}`
- `GET /api/v1/data/{slug}/{id}`
- `PUT /api/v1/data/{slug}/{id}`
- `DELETE /api/v1/data/{slug}/{id}`
- `GET /api/v1/data/{slug}`

It uses `internal/services/dynamic_data_service.go` to parse the dynamic metadata and map it into SQL queries on the corresponding database connection.

---

## 5. Query Engine

The Query Engine supports advanced queries directly on dynamically created services:
- **Filtering**: `GET /api/v1/data/users?email=john@example.com`
- **Sorting**: `GET /api/v1/data/users?sort=created_at:desc`
- **Pagination**: `GET /api/v1/data/users?page=1&limit=20`

These features are automatically applied within the dynamic data handler before querying the database, leveraging optimized connection pooling.

---

## 6. Schema Migration Engine

The Schema Migration Engine (`internal/services/migration_service.go`) tracks and applies safe schema diffs using GORM’s `.Migrator()`.
Supported operations:
- Add column
- Drop column
- Rename column
- Change column type

It logs migration status natively into the `migrations` table and supports rollback capability through historical states and snapshot retention logic.

---

## 7. GraphQL Gateway (Optional)

A dynamically generated GraphQL API is available via `internal/graphql/gateway.go`.
It utilizes `github.com/99designs/gqlgen` to automatically expose queries, mutations, and relations for the dynamically created services.

Example query:
```graphql
query {
  users {
    id
    name
    email
  }
}
```

---

## 8. Backup System

The Backup Engine (`internal/services/backup_service.go`) supports taking table backups, schema backups, and service snapshots.
The backups can be retrieved in standard formats like JSON.

Example Endpoints:
- Backup: `POST /api/v1/cms/backup/service/{service_id}`
- Restore: `POST /api/v1/cms/restore/service/{service_id}`

---

## 9. Observability Integration

The platform includes a robust observability stack out-of-the-box:
- **Logging**: Zap structured logging is injected globally (`pkg/logger/`) for request logs, error logs, and query logs with contextual trace IDs.
- **Metrics**: Prometheus metrics are exported at `/metrics` tracking request latency, status codes, and error rates via standard HTTP interceptors (`internal/middleware/prometheus.go`).
- **Tracing**: OpenTelemetry (Jaeger) tracing is initialized in `internal/tracing/` to wrap SQL commands and network logic.

---

## 10. Security & Performance Optimization

**Security:**
- Input validation on all dynamic payloads.
- SQL injection protection via GORM parameters.
- JWT authentication for protected API routes.
- Rate limiting to protect API endpoints.

**Performance Optimization:**
- Connection pooling for user database connectors.
- Pagination optimizations for large datasets.
- N+1 query optimization using bulk batched processing in GORM.

---

## 11. Docker Setup

The application is fully containerized with an included `docker-compose.yml`.

Services included:
- `cms-backend`: The Go platform core
- `postgres`: Metadata database
- `redis`: Query caching and message brokering
- `jaeger`: OpenTelemetry tracing backend
- `prometheus`: Metrics collection

Run the stack using:
```bash
docker-compose up -d
```

---

## 12. Unit Testing

The system includes comprehensive unit testing with a minimum 80% coverage across Repository, Service, and API handler layers.

Tests are located in `tests/`.
Command to run tests:
```bash
make test-coverage
# or manually
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

---

## 13. API Documentation (Swagger)

Swagger (OpenAPI) documentation is auto-generated for the core platform APIs.

To generate the docs:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

The docs can be found mapped to a handler on the web server or as static JSON output in the `docs/` folder.
