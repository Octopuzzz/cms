# Go Backend Platform (Dynamic CMS + API Builder)

This is a **Backend-as-a-Service platform** built entirely in Go. It allows developers to dynamically create backend services, manage database connections, define data schemas, auto-generate CRUD REST endpoints and optional GraphQL endpoints, and monitor systems with state-of-the-art observability.

It acts similarly to platforms like Hasura or Supabase but is **fully self-hosted and Go-native**.

---

## 🚀 Platform Goal

The platform acts as a **backend infrastructure generator**, allowing developers to:
- Connect databases
- Create data models dynamically
- Generate CRUD APIs automatically
- Generate optional GraphQL APIs
- Manage schema migrations
- Monitor logs and performance
- Manage backups
- Scale services

---

## 🛠️ Core Technology Stack

- **Language:** Go 1.24+
- **API Layer:** REST (default)
- **Optional API Layer:** GraphQL (via `gqlgen`)
- **Web Framework:** Gin
- **ORM:** GORM
- **Logging:** Zap structured logging
- **Metrics:** Prometheus
- **Tracing:** OpenTelemetry (Jaeger)
- **Cache:** Redis
- **Containerization:** Docker
- **API Documentation:** Swagger / OpenAPI

---

## 🏗️ System Architecture

The CMS Backend follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

- **Presentation Layer**: The Gin HTTP Router and Handlers (`internal/handlers/`). Acts as the API gateway mapping requests to internal services. Includes GraphQL endpoints using `gqlgen`.
- **Application Layer**: Business logic (`internal/services/`). Services control data access, generate dynamic schemas, and perform operations requested by handlers.
- **Domain Layer**: The data models (`internal/models/`). Defines core entities like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure Layer**: Cross-cutting tools such as connection caching, telemetry (`internal/tracing/`), metrics (Prometheus), and logging (`pkg/logger/`).

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

---

## 📂 Project Structure

```text
.
├── cmd
│   └── server          # Application entrypoint
├── docker              # Docker-related files
├── internal
│   ├── api             # API routes
│   ├── config          # Configuration and environment setup
│   ├── database        # Database Connection Manager
│   ├── graphql         # GraphQL Gateway
│   ├── handlers        # Presentation Layer (Gin HTTP Handlers)
│   ├── middleware      # Gin middlewares (Auth, Tracing, Metrics)
│   ├── models          # Domain Layer (Database schemas)
│   ├── services        # Application Layer (Business logic)
│   └── tracing         # OpenTelemetry configuration
├── pkg
│   ├── errors          # Custom error definitions
│   ├── logger          # Zap structured logging
│   ├── pagination      # Pagination utilities
│   ├── response        # Standardized API responses
│   └── validation      # Input validation logic
├── tests               # Unit and Integration tests
├── docs                # Swagger API documentation
├── Makefile            # Standardized tasks
└── README.md           # This file
```

---

## 🗄️ Metadata Database Schema

The core internal configuration is stored in the **Metadata Database** (e.g., PostgreSQL or SQLite for local dev). Key tables include:

1. **`database_connections`**: Stores external DB configurations with fields for id, name, type, host, port, credentials.
2. **`services`**: Represents user-created data models. Contains a reference to `database_connection_id` and tracks dynamic schemas (`db_table_name`).
3. **`fields`**: Defines attributes for each service (string, integer, float, uuid, JSON). Configures uniqueness, nullability, defaults.
4. **`service_permissions`**: Connects `Role` to `Service` providing fine-grained access checks.
5. **`migrations`**: Keeps track of DDL executions with metadata required for rollbacks or history tracking.
6. **`backups`**: Stores snapshot records or schema outputs.
7. **`users` and `roles`**: General authentication and authorization.
8. **`audit_logs`**: Detailed logging of modifications.

---

## 🧩 Core Components

### 1. CMS Control Plane & Service Builder
Provides an interface to manage database connections, create services (data models), and define fields dynamically. Managed under `/api/v1/cms/` routes. Validates foreign database connectivity and triggers migrations on service model updates.

### 2. Database Connection Manager
Users can register external databases (PostgreSQL, MySQL, MongoDB/SQLite). The system pools connections to these external sources to ensure performance and reliability.

### 3. Relation Engine
Supports relational data modeling including One-to-One, One-to-Many, Many-to-One, and Many-to-Many relationships.

### 4. Dynamic CRUD & Query Engine
Automatically maps HTTP endpoints to dynamic queries.
- Example endpoints: `POST /api/v1/data/{slug}`, `GET /api/v1/data/{slug}/{id}`
- Supports dynamic filtering (`?email=john@example.com`), sorting (`?sort=created_at:desc`), pagination (`?page=1&limit=20`), and relation joins.

### 5. GraphQL Gateway (Optional)
Automatically generates optional GraphQL schemas for user-defined services via `gqlgen`. Supports dynamic queries, mutations, and relations. Exposed at `POST /api/v1/graphql`.

### 6. Schema Migration Engine
Tracks and applies safe schema diffs (Add, Drop, Rename Column) using GORM's `Migrator`. Supports automatic backups before migrations and rollback capabilities.

### 7. Backup Engine
Generates and restores data snapshots. Supports generic service record backup mapping dynamic tables to JSON snapshots.

### 8. Observability
- **Logging:** Request, error, and query logs via Zap.
- **Metrics:** Request latency, slow queries, and error rates exported to Prometheus via `/metrics`.
- **Tracing:** OpenTelemetry/Jaeger wraps SQL commands and network logic.

### 9. Security & Performance
- Implements JWT authentication, RBAC, input validation, and SQL injection protection.
- Optimized with database connection pooling and robust pagination handling.

---

## 🧪 Testing

The system includes comprehensive tests ensuring at least 80% coverage across Repository, Service, and API handler layers.

**Run all tests:**
```bash
make test-coverage
# or
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
```

---

## 🐳 Docker Setup

The platform is fully containerized. A `docker-compose.yml` file is provided to spin up the application along with Redis, PostgreSQL (for metadata), Prometheus, and Jaeger.

**Start the platform:**
```bash
docker-compose up -d
```

---

## 📚 API Documentation

The platform uses Swagger for API documentation.

**Generate Swagger docs:**
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go --parseDependency --parseInternal --output docs
```

The documentation will be available at `http://localhost:8080/swagger/index.html`.
