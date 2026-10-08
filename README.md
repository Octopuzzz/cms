# Backend Platform (Dynamic CMS + API Builder)

This is a production-grade Backend-as-a-Service platform, fully self-hosted and Go-native. It allows developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

---

## 1. System Architecture Explanation

The system follows **Clean Architecture** and **Domain-Driven Design (DDD)**:

*   **Presentation Layer**: Implemented using the Gin HTTP framework (`internal/handlers/`). It routes requests to internal services, handling API gateway responsibilities for REST endpoints. It also integrates `gqlgen` for optional GraphQL schemas (`internal/graphql/`).
*   **Application Layer**: Contains business logic (`internal/services/`). These services govern data access, auto-generate dynamic schemas, handle migrations, and control backups.
*   **Domain Layer**: Represents the core data models (`internal/models/`). Entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`, `Migration`, and `Backup` are mapped via GORM.
*   **Infrastructure Layer**: Provides cross-cutting tools such as database connection caching/pooling (`internal/database/`), structured logging via Zap (`pkg/logger/`), performance metrics via Prometheus (`internal/middleware/`), and distributed tracing using OpenTelemetry/Jaeger (`internal/tracing/`).

---

## 2. Metadata Database Schema

The internal configuration and platform metadata are persistently stored using a relational database (PostgreSQL recommended). Core tables include:

*   `database_connections`: Contains connection configurations to external databases (ID, Name, Type [PostgreSQL/MySQL/MongoDB], Host, Port, Credentials) with connection pooling capability.
*   `services`: User-created data models. Includes properties like `database_connection_id` and the dynamically generated `db_table_name`.
*   `fields`: Defines the fields for a specific service. Configuration includes type (string, integer, float, uuid, JSON), nullability, defaults, uniqueness, and indices.
*   `relations`: Represents connections between fields across services (e.g., One-to-One, One-to-Many).
*   `service_permissions`: Connects Roles to Services to facilitate Role-Based Access Control (RBAC), implementing row and field-level permissions.
*   `migrations`: Records executed DDL operations, containing metadata for status tracking and rollbacks.
*   `backups`: Tracks data snapshot and schema records for system restoration.

---

## 3. Go Project Structure

The project implements a clean modular structure to isolate responsibilities:

```text
├── cmd
│   └── server                # Main application entrypoint
├── internal
│   ├── config                # Environment configurations
│   ├── database              # Connection manager & DB pooling
│   ├── graphql               # Auto-generated GraphQL Gateway
│   ├── handlers              # Gin HTTP handlers (REST API)
│   ├── middleware            # Auth, Prometheus, Rate Limiter
│   ├── models                # Core Domain Models
│   ├── services              # Application Business Logic
│   └── tracing               # OpenTelemetry & Jaeger Integration
├── pkg
│   ├── logger                # Zap structured logging
│   └── response              # Standardized API responses
├── tests
│   ├── integration           # Integration testing suite
│   ├── uat                   # User Acceptance Testing
│   └── unit                  # Unit testing suite
├── docker                    # Container configs (Prometheus, Dockerfiles)
├── Makefile                  # Build, Run, Test, and Format scripts
└── ARCHITECTURE.md           # Extended architecture guidelines
```

---

## 4. CRUD Engine Implementation

The **Dynamic CRUD Engine** (`internal/services/dynamic_data_service.go`) abstracts endpoint manipulation on the fly.
When a data model (`Service`) is created:
*   Standard operations (Create, Read, Update, Delete, List) are mapped securely through the `GET/POST/PUT/DELETE /api/v1/data/{slug}` routes.
*   The system uses GORM dynamic queries utilizing `.Table(db_table_name)`.
*   It implements advanced Querying including:
    *   **Filtering**: `?field=value` maps directly to `WHERE` clauses.
    *   **Sorting**: `?sort_by=created_at&sort_order=desc`.
    *   **Pagination**: Configured via `page` and `limit` query parameters.
    *   **Joins**: Relational field configurations map to database level `JOIN` definitions.

---

## 5. Schema Migration Engine

The **Schema Migration Engine** (`internal/services/migration_service.go`) guarantees atomic and safe data changes.
*   Provides granular operations including: Add Column, Drop Column, Rename Column, and Change Column Type.
*   Leverages standard GORM `Migrator` utilities to apply incremental schema updates natively.
*   Schema drift is saved in the `migrations` metadata table, mapping operations to descriptions, success/fail statuses, and snapshot diffs.
*   Future integrations natively include automatic backup rollbacks before executing destructive DDL statements.

---

## 6. GraphQL Gateway

The **GraphQL Gateway** (`internal/graphql/gateway.go`) optionally translates dynamic services into strongly-typed GraphQL queries.
*   Powered by `github.com/99designs/gqlgen/graphql`.
*   A runtime schema generator processes `models.Service` and `models.Field` configurations and creates abstract syntactical trees (ASTs).
*   Mutations and Queries resolve directly against the dynamic data mapping layer using `DynamicDataService`.
*   Exposed universally via `POST /api/v1/graphql` to fulfill modern data fetching needs.

---

## 7. Backup System

The **Backup Engine** (`internal/services/backup_service.go`) isolates system snapshots preventing catastrophic data loss.
*   Creates backups combining both `Schema` (Service configurations) and raw `TableData`.
*   JSON structural representations easily parse out entire system dependencies allowing rapid data migration between environments.
*   Logged securely inside the `backups` table mapping the snapshot status, time, row counts, and structural dependencies.

---

## 8. Observability Integration

The platform provides complete cloud-ready observability integrated at all levels.
*   **Logging**: `pkg/logger/logger.go` injects Zap structured logging. Logs natively correlate errors across handlers by capturing user and request telemetry ID.
*   **Tracing**: `internal/tracing/tracing.go` configures `go.opentelemetry.io` with Jaeger exporters mapping specific trace contexts through network, service, and database execution times.
*   **Metrics**: `internal/middleware/prometheus.go` tracks real-time performance of request latency, error distribution rates, and query latency natively at the `/metrics` endpoint.

---

## 9. Unit Tests

System reliability is enforced through aggressive automated testing ensuring code stability.
*   A minimum of **80% coverage** is maintained globally.
*   Separated testing suites focus on Unit (`tests/unit/`), Integration (`tests/integration/`), and End-to-End (`tests/uat/`).
*   Mock frameworks utilizing an in-memory SQLite backend provide isolated testing without dependency overlaps.
*   Execute via Makefile: `make test-coverage` or `go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...`.

---

## 10. Docker Setup

A production-ready infrastructure utilizes container orchestration.
*   **Application Container**: Defined in `Dockerfile`, runs the Go binaries optimized for minimal file sizing.
*   **Docker Compose**: `docker-compose.yml` natively spins up the core API alongside dependent services like Prometheus and databases ensuring environments perfectly emulate production.

---

## 11. Swagger Documentation

API documentation is completely auto-generated reflecting live implementations.
*   Handled cleanly using `github.com/swaggo/swag`.
*   Detailed API inputs, JWT token authorizations, and definitions are integrated with Swagger syntax.
*   Can be built manually via Makefile command `make swagger` outputting directly to the un-tracked `/docs` directory.
