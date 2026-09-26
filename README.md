# CMS Backend Platform

A production-grade Backend Platform (Dynamic CMS + API Builder) written in Go. This system acts as a backend infrastructure generator, allowing users to dynamically create backend services, schemas, APIs, and optional GraphQL endpoints.

## 🚀 Platform Goal

This Backend-as-a-Service platform enables developers to:
- Connect external databases
- Create data models dynamically
- Generate CRUD APIs automatically
- Generate optional GraphQL APIs
- Manage schema migrations
- Monitor logs and performance
- Manage backups
- Scale services

## ⚙️ Core Technology Stack

- **Language:** Go (latest stable version)
- **API Layer:** REST (default) and GraphQL (optional via gqlgen)
- **Web Framework:** Gin
- **ORM:** GORM
- **Logging:** Zap structured logging
- **Metrics:** Prometheus
- **Tracing:** OpenTelemetry
- **Cache:** Redis
- **Message Queue:** NATS / Kafka (optional)
- **Database Migration:** GORM AutoMigrate / Atlas / Goose
- **Containerization:** Docker
- **API Documentation:** Swagger / OpenAPI

## 🏗️ High-Level System Architecture

The platform follows Clean Architecture and Domain-Driven Design (DDD).

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

## 🗄️ Metadata Database Schema

The core configuration is stored in a relational Metadata Database (e.g., PostgreSQL or SQLite for local dev).

### Core Tables
1. **`database_connections`**: Stores external DB configs (id, name, type, host, port, credentials).
2. **`services`**: Represents user-created data models, tracks schemas (`db_table_name`).
3. **`fields`**: Attributes for each service (type, unique, nullable, default_value).
4. **`relations`**: Defines relationships (One-to-One, One-to-Many, Many-to-One, Many-to-Many).
5. **`service_permissions`**: Fine-grained access control.
6. **`migrations`**: Tracks DDL executions.
7. **`backups`**: Stores snapshot records.
8. **`users` / `roles`**: Auth & authorization for the control plane.
9. **`audit_logs`**: Logs structural and data modifications.

## 📁 Go Project Structure

A clean, modular Go structure:
```text
.
├── cmd/
│   └── server/          # Entry point for the application
├── internal/
│   ├── config/          # Environment configuration
│   ├── database/        # Database connection managers
│   ├── graphql/         # GraphQL Gateway
│   ├── handlers/        # Gin API controllers (Presentation Layer)
│   ├── middleware/      # Auth, Prometheus, Rate Limiter
│   ├── models/          # Domain Entities
│   ├── services/        # Business Logic (Application Layer)
│   └── tracing/         # OpenTelemetry configuration
├── pkg/
│   ├── logger/          # Zap logger wrapper
│   └── response/        # Standardized API responses
├── tests/               # Unit, Integration, UAT tests
├── docker/              # Docker configuration files
├── Dockerfile           # App container build
├── docker-compose.yml   # Services orchestration
└── Makefile             # Task automation
```

## 🛠️ Core Components

### 1. CMS Control Plane
Management interface to control the system. Provides endpoints to manage DB connections, services, schemas, and migrations. Example: `POST /api/v1/cms/services`.

### 2. Service Builder
Dynamically creates services (data models) defining database connections, fields (string, int, uuid, JSON), relations, and validation rules.

### 3. Dynamic CRUD Engine
Automatically generates endpoints when a service is created:
- `POST /api/v1/data/{service}`
- `GET /api/v1/data/{service}/{id}`
- `PUT /api/v1/data/{service}/{id}`
- `DELETE /api/v1/data/{service}/{id}`
- `GET /api/v1/data/{service}`

### 4. Query Engine
Supports advanced API queries on the generated endpoints:
- Filtering: `?email=john@example.com`
- Sorting: `?sort=created_at:desc`
- Pagination: `?page=1&limit=20`

### 5. Schema Migration Engine
Tracks and applies safe schema diffs (Add, Drop, Rename Column). Includes logging of migration statuses and handles rollbacks through snapshot retention logic.

### 6. GraphQL Gateway
Optionally generates standard GraphQL APIs from schemas (powered by `gqlgen`). Supports dynamic queries and mutations matching the underlying data models.

### 7. Backup System
Supports generating data snapshots (JSON snapshot / SQL dump). Managed via internal backup services.

### 8. Observability
State-of-the-art observability out of the box:
- **Logging:** Zap structured logs injected into context.
- **Tracing:** OpenTelemetry/Jaeger wraps DB and network calls.
- **Metrics:** Prometheus endpoint `/metrics` for latency, errors, and rates.

## 🔒 Security
- JWT-based Authentication
- Role-Based Access Control (RBAC)
- Rate Limiting Middleware
- Structured Logging / Audit Logging

## 🐳 Docker Setup

Run the entire stack locally using Docker Compose:
```bash
make docker-up
# Or
docker-compose up -d
```

## 📚 API Documentation (Swagger)

Generate and access Swagger API documentation:
```bash
make swagger
```
(Docs are output to `docs/` folder, ignored by git).

## 🧪 Unit Testing

Run the full test suite with coverage:
```bash
make test-coverage
```
Ensure >80% coverage across Repository, Service, and API handlers. Use standard Go testing tools.
