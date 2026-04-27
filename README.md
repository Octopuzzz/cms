# Dynamic CMS + API Builder (Go-native BaaS)

This is a production-grade Backend-as-a-Service (BaaS) platform written in Go. It enables developers to dynamically connect to databases, build data models/schemas, automatically generate CRUD REST/GraphQL APIs, run schema migrations, and collect comprehensive observability metrics.

## System Architecture

The platform follows Domain-Driven Design and Clean Architecture with clearly defined layers:
- **Presentation**: `internal/handlers` maps REST routes via Gin framework. Includes a GraphQL Gateway using `gqlgen`.
- **Application**: `internal/services` implements core functionality (Service Builder, Dynamic CRUD Engine, Database Connection Manager, Backup Engine, Schema Migrations).
- **Domain**: `internal/models` represents core platform abstractions like `Service`, `Field`, `DatabaseConnection`, `User`, `Role`.
- **Infrastructure**: Distributed system tooling includes Redis connection caching, Jaeger OpenTelemetry tracing (`internal/tracing`), structured Zap logging, and Prometheus endpoints.

It's a Go-based self-hosted backend infrastructure generator similar to Supabase or Hasura.

## Metadata Database Schema

The control plane stores configuration metadata within its internal metadata database. The schema includes:
1. `users` and `roles` - For authentication and role-based access control mapping.
2. `database_connections` - Credentials, host settings, connection pooling parameters for external databases.
3. `services` - Metadata defining models generated for end users dynamically.
4. `fields` - Configuration variables attached to models (string, int, JSON) and relational configurations.
5. `service_permissions` - Fine-grained capability declarations connecting models and roles.
6. `migrations` and `backups` - Tracking of DDL status, and data snapshot metadata.
7. `audit_logs` - Structured history of metadata changes and platform events.

## Features

- **CMS Control Plane:** Interface to create models and handle configuration.
- **Service Builder:** Define new services, fields, unique constraints, and schema features programmatically.
- **CRUD Engine:** Exposes dynamic routes (`GET/POST /api/v1/data/{model}`) applying zero-touch database transactions. Supports sorting, pagination, and relation joins.
- **GraphQL Gateway:** Exposes an optional dynamic GraphQL API wrapping auto-generated models.
- **Database Connection Manager:** Manage and connect to PostgreSQL, MySQL, SQLServer, MongoDB, and SQLite databases from the dashboard.
- **Schema Migration Engine:** Apply controlled operations (add, drop, rename columns) securely.
- **Observability Integration:** Structured logging, Prometheus metrics, and OpenTelemetry Jaeger traces out of the box.
- **Docker Setup:** Provided `docker-compose.yml` launches API, DBs, Redis, Prometheus, and Jaeger seamlessly.
- **Swagger Documentation:** Auto-generated API documentation.

## Running the Platform

Ensure Docker is installed, then spin up the environment:
```bash
docker-compose up -d
```

## Unit Testing
We enforce a strict >80% test coverage threshold covering Repository, Service, and API Layers. Run tests with:
```bash
go test -v -race -coverprofile=coverage.out -coverpkg=./... ./...
```
