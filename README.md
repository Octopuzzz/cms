# Platform BaaS - Backend-as-a-Service

Platform BaaS is a production-grade, self-hosted Backend-as-a-Service built in Go. It allows developers to connect databases, dynamically generate schemas and APIs (REST and optional GraphQL), and manage backend infrastructure autonomously, similar to Hasura or Supabase but completely Go-native.

## Architecture

The system follows Clean Architecture and Domain Driven Design, encompassing:
1.  **Presentation Layer:** REST and GraphQL Handlers, API Gateway capabilities
2.  **Application Layer:** Application Services (CMS, CRUD Engine, Query Engine)
3.  **Domain Layer:** Core Entities (Services, Fields, Relations)
4.  **Infrastructure Layer:** Database connectors (PostgreSQL, MySQL, MongoDB), Message Queues, Cache

```text
Client Applications
│
├── REST API
├── GraphQL API (optional)
│
▼
API Gateway
│
▼
Backend Platform Core
│
├── CMS Control Plane
├── Service Builder
├── CRUD Engine
├── Query Engine
├── Schema Migration Engine
├── Backup Engine
├── Observability Engine
│
▼
Database Connectors
│
├── PostgreSQL
├── MySQL
└── MongoDB
```

## Features

*   **CMS Control Plane:** Manage databases, services, relations, migrations, and backups via REST API.
*   **Database Connection Manager:** Support for connecting and pooling PostgreSQL, MySQL, and MongoDB.
*   **Service Builder & CRUD Engine:** Dynamically define data models that automatically generate CRUD APIs.
*   **Query Engine:** Built-in REST APIs support advanced filtering, sorting, pagination, and relation joins.
*   **GraphQL Gateway (Optional):** Automatically generated GraphQL schemas from defined service models using `gqlgen`.
*   **Schema Migration Engine:** Safely manage schema changes (add/drop columns) with backup and rollback capabilities.
*   **Observability:** Integrated Zap logging, Prometheus metrics, and OpenTelemetry tracing.
*   **Security:** JWT authentication, RBAC authorization, rate limiting.

## Database Schema (Metadata Database)

The metadata database stores the platform configuration:
*   `database_connections`: External database registrations.
*   `services`: Definitions of data models / schemas.
*   `fields`: Attributes belonging to services.
*   `service_permissions`: RBAC per service.
*   `migrations`: Schema migration history.
*   `backups`: Snapshot and schema backup records.

## Project Structure

```text
├── cmd
│   └── server          # Entry point for the application
├── docker              # Dockerfiles and Compose files for local development
├── internal            # Application code
│   ├── config          # Application configuration mapping
│   ├── database        # Database Connection Manager
│   ├── graphql         # Dynamic GraphQL Schema Builder
│   ├── handlers        # HTTP handlers (REST API controllers)
│   ├── middleware      # Gin middleware (Auth, Rate Limiting, Metrics)
│   ├── models          # Domain entities & GORM mappings
│   ├── services        # Business logic (Service Builder, CRUD, Migration, etc.)
│   └── tracing         # OpenTelemetry Setup
├── pkg                 # Reusable utility packages
│   ├── graphql         # GraphQL Utilities
│   ├── logger          # Zap structured logger wrapper
│   └── response        # Standardized HTTP response formatter
└── tests               # Unit, integration, and UAT tests
```

## Getting Started

### Prerequisites
*   Go 1.25+
*   Docker & Docker Compose (for local development dependencies)
*   GNU Make

### Running Locally

1.  **Set up dependencies:**
    ```bash
    docker-compose -f docker/docker-compose.yml up -d
    ```
    This will start PostgreSQL (for Metadata), Redis (Cache), and Prometheus.

2.  **Run the application:**
    ```bash
    make run
    ```
    Or manually:
    ```bash
    go run cmd/server/main.go
    ```

3.  **View Swagger Docs:**
    Navigate to `http://localhost:8080/swagger/index.html`

4.  **View GraphQL Playground:**
    Navigate to `http://localhost:8080/api/v1/graphql/playground`

## Testing

Run unit tests:
```bash
make test
```
Or manually:
```bash
go test ./... -v -cover
```
