# Dynamic CMS Backend & API Builder

A self-hosted, Go-native Backend-as-a-Service (BaaS) platform that dynamically generates REST and GraphQL APIs.

## System Architecture

The platform acts as a backend infrastructure generator, similar to Hasura or Supabase but completely native to Go. It follows Clean Architecture and Domain-Driven Design (DDD).

### Core Components
- **CMS Control Plane**: A management API to create services (data models), manage relations, and manage database connections.
- **Service Builder**: Dynamically registers metadata representing schemas and constraints.
- **Schema Migration Engine**: Generates and applies DDL operations to alter the physical database schema dynamically.
- **Dynamic CRUD Engine**: Automatically provides generic CRUD REST endpoints for dynamically generated services.
- **GraphQL Gateway**: Dynamically builds GraphQL schemas at runtime from service metadata to allow complex querying.
- **Backup Engine**: Facilitates schema and data backups as JSON or SQL.
- **Observability**: Integrates OpenTelemetry tracing, Prometheus metrics, and structured Zap logging.

### Tech Stack
- **Language**: Go
- **Web Framework**: Gin
- **ORM**: GORM
- **GraphQL**: `graphql-go/graphql`
- **Database Support**: PostgreSQL, MySQL, SQL Server, SQLite
- **Logging**: Zap / Logrus
- **Metrics/Tracing**: Prometheus & OpenTelemetry

## Project Structure

```text
├── cmd/
│   └── server/       # Main entry point and Gin server setup
├── internal/
│   ├── config/       # Environment config
│   ├── database/     # Multi-DB dynamic connection pool manager
│   ├── graphql/      # Dynamic GraphQL schema generation
│   ├── handlers/     # REST HTTP Controllers
│   ├── logger/       # Zap logging wrapper
│   ├── metrics/      # Prometheus metrics definitions
│   ├── middleware/   # Gin middlewares (Auth, Tracing, Metrics)
│   ├── models/       # Metadata Domain Models
│   ├── services/     # Core Business Logic (Engines)
│   └── tracing/      # OpenTelemetry provider setup
├── pkg/              # Utility packages
├── docker/           # Docker resources
└── tests/            # Unit & Integration tests
```

## Metadata Database Schema

The CMS stores its structural information in tables. Core tables:
- `database_connections`: External database configs.
- `services`: Dynamic models created by users.
- `fields`: Columns in the dynamic models.
- `migrations`: Tracked schema changes.
- `backups`: Stored backup data snapshots.
- `users`, `roles`, `permissions`: Access control.

*(See `internal/models/models.go` for the detailed GORM definitions)*

## Deployment & Usage

### Running Locally (Docker)
The easiest way to run the CMS platform is via Docker.

```sh
docker-compose up -d --build
```

This will start:
- CMS Go Backend (Port: 8080)
- PostgreSQL Metadata DB
- Redis (for caching)

### Running Locally (Native)

1. **Install Dependencies**
   ```sh
   go mod download
   ```

2. **Configure Environment**
   ```sh
   cp .env.example .env
   # Update config as necessary
   ```

3. **Run Server**
   ```sh
   make run
   # Or using air for hot-reload
   make dev
   ```

### Running Tests

```sh
make test
make test-unit
```

## API Documentation

The platform includes auto-generated Swagger documentation.

When running the platform, visit:
`http://localhost:8080/swagger/index.html`

## GraphQL Playground

Access the interactive GraphQL interface:
`http://localhost:8080/api/v1/graphql/playground`
