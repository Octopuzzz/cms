# CMS Backend Platform

This repository implements a modular, self-hosted Backend Platform and API Builder written in Go, acting as a dynamic CMS and infrastructure generator akin to Hasura or Supabase.

The platform provides functionality for:
- Database connection management
- Dynamic CRUD engine with auto-generated REST APIs
- Extensible API via a generated GraphQL Gateway
- Database metadata and schema migrations tracking
- Automatic Service Backups
- Role-Based Access Control

## Requirements
- Go 1.25+
- Docker and Docker Compose

## Quick Start

1.  **Run with Docker Compose**
    ```bash
    make docker-up
    ```
    This starts the CMS platform, a PostgreSQL database, Redis cache, and Prometheus.

2.  **API Endpoints**
    - REST Base Path: `http://localhost:8080/api/v1`
    - GraphQL Endpoint: `http://localhost:8080/api/v1/graphql`
    - Swagger UI: `http://localhost:8080/swagger/index.html`

## Generating the Static Typed GraphQL Schema (Optional feature illustration)
The provided GraphQL gateway supports querying any generic dynamic service using untyped JSON data. To generate statically typed Go handlers based on a live dynamic database (to support schemas like `query { users { id name email } }`), you would generally generate `.graphqls` based on the database at runtime and then run the included `gqlgen` generator programmatically.

Due to Go's compiled nature, a fully static GraphQL endpoint with strongly typed structs must be compiled, which is typical for `gqlgen`. Thus, this platform implements a dynamic gateway layer exposing data generically.

## Commands
Check the `Makefile` for available commands.
