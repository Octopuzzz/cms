# CMS Backend Platform Architecture

## Platform Overview
This platform acts as a self-hosted Backend-as-a-Service (BaaS) and CMS infrastructure generator, similar to Hasura or Supabase. It allows users to dynamically create database services, generate APIs, perform schema migrations, handle backups, and interact via a standard REST API or an auto-generated generic GraphQL Gateway.

## Core Technology Stack
- **Language**: Go (latest)
- **API Framework**: Gin
- **ORM**: GORM (with auto-migration features)
- **Database Engine**: Multi-db support via connection pooling (PostgreSQL, MySQL, SQLite, SQLServer)
- **GraphQL**: `gqlgen` for the optional gateway
- **Observability**: Prometheus (Metrics), OpenTelemetry/Jaeger (Tracing), Zap (Structured Logging)
- **Caching**: Redis
- **Security**: JWT tokens, Role-Based Access Control (RBAC)

## High-Level Architecture
```text
Client Applications
        │
        ├── REST API
        ├── GraphQL API
        │
        ▼
   API Gateway (Gin)
        │
        ▼
   Backend Platform Core
        │
        ├── CMS Control Plane (manage services, DBs, roles)
        ├── Service Builder (dynamic table creation via GORM DDL)
        ├── Dynamic CRUD Engine (dynamic data operations)
        ├── Schema Migration Engine (handles ADD/DROP/RENAME column operations)
        ├── Backup Engine (snapshots schema and table data)
        ├── Observability Engine (Prometheus, Jaeger, Zap)
        │
        ▼
  Database Connectors
        │
        ├── PostgreSQL (Default / Custom)
        ├── MySQL
        ├── SQLite
        └── SQL Server
```

## Metadata Database Schema
The Control Plane relies on a Metadata Database (typically PostgreSQL, defaulting to SQLite in dev) to track user configurations, services, and policies.

Key tables:
- `users`: Tracks admin and consumer users.
- `roles`, `permissions`, `user_roles`: Handles RBAC.
- `database_connections`: Manages credentials to external databases.
- `services`: Represents a dynamically created model or table. Contains `db_table_name` and points to a `database_connection_id`.
- `fields`: Belongs to a service. Defines columns (`type`, `is_nullable`, `is_unique`, `relation_config`).
- `migrations`: Audit log of DDL changes applied to a service table.
- `backups`: Stores snapshots (JSON dumps) of schema and service table data.

## GraphQL Implementation
To support the dynamic nature of the services, a generic GraphQL Gateway is generated using `gqlgen`. Instead of hardcoding all model structs, the Gateway relies on a custom `JSON` scalar and the `map[string]interface{}` bindings in Go.
Users can query data through the Gateway using dynamic `service` arguments.

Example:
```graphql
query {
  data(service: "users", id: 1)
}

query {
  list(service: "posts", page: 1, limit: 10, search: "Hello")
}
```

## Modular Project Structure
Following Clean Architecture and Domain-Driven Design (DDD):
- `/cmd/server`: Entry point for the Go application.
- `/internal/config`: Environment and startup configuration.
- `/internal/database`: Multi-DB connection manager and pooling.
- `/internal/models`: Core metadata schema structs.
- `/internal/services`: Domain logic (CRUD engine, service builder, migration, auth, backups).
- `/internal/handlers`: Gin REST controllers and payload parsing.
- `/internal/graphql`: gqlgen configuration, generic schema, and resolvers.
- `/internal/middleware`: Auth, Logging, Tracing, Metrics, Rate Limiting.
- `/pkg/logger`: Zap wrapper.
- `/tests`: Unit, Integration, and UAT tests.
- `/docker`: Docker Compose and Prometheus setup.
