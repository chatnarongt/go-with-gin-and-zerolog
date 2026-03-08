# AGENTS.md

This file provides guidance for AI coding agents working on this project.

## Project Overview

This is a Go REST API built with the Gin web framework, Zerolog for structured logging, and Microsoft SQL Server as the database. The API follows a modular architecture with clear separation of concerns.

## Architecture

### Module System

The application is organized into **modules** under `internal/modules/`. Each module is self-contained and exposes a `Module` struct with a `NewModule()` constructor. Modules are wired together in `cmd/api/main.go`.

The initialization order matters:

1. `config` — Loads environment variables
2. `database` — Opens the SQL Server connection pool (depends on `config`)
3. `swagger` — Configures Swagger UI (depends on `config`)
4. `health` — Provides liveness and readiness endpoints (depends on `database`)
5. `schedule` — Background jobs and cron tasks (depends on `database`)
6. `application` — Creates the Gin engine, maps routes, and starts the HTTP server (depends on `config`)

### Route Mapping

- **Root-level routes** (e.g., Swagger UI at `/swagger/*`) are registered via `app.MapRoutes(module)`
- **API routes** (under `/api/*`) are registered via `app.MapAPIRoutes(module)`
- Modules that register routes must implement a `MapRoutes(*gin.RouterGroup)` method

### Error Handling

Custom HTTP error types are defined in `internal/errs/`. These are used throughout the codebase and handled by the centralized error-handling middleware in `internal/middleware/error_handler.go`.

## Code Style & Conventions

- **Environment configuration**: All configuration values are read from environment variables via helper functions in `internal/modules/config/` (`getEnvAsString`, `getEnvAsInt`, `getEnvAsBool`). Environmental overrides are loaded from `.env.<environment>` files (e.g., `.env.development`, `.env.test`). Never hardcode configuration values.
- **Logging**: Use `github.com/rs/zerolog/log` for all logging. Use structured fields (e.g., `log.Error().Err(err).Msg("message")`), not `fmt.Printf`.
- **Graceful shutdown**: The application handles `SIGINT` and `SIGTERM` and runs cleanup functions registered via `app.OnBeforeShutdown()` (e.g., closing the database connection).
- **Swagger annotations**: API endpoints are documented using [Swaggo](https://github.com/swaggo/swag) annotations in handler functions. After modifying annotations, run `make swag` to regenerate `docs/`.
- **Guard Clauses**: Use guard clauses to handle edge cases and errors early, reducing indentation and making the main logic more readable.

## Building & Running

```bash
# Run locally (loads .env and .env.development by default)
make run

# Run with specific environment (e.g., test)
make env-test run

# Run tests
make test

# Build binary
make build

# Build Docker image
docker build -f cmd/api/Dockerfile -t go-with-gin-and-zerolog .

# Regenerate Swagger docs
make swag
```

## Key Files

| File | Purpose |
|---|---|
| `cmd/api/main.go` | Application entrypoint and module wiring |
| `cmd/worker/main.go` | Background worker entrypoint and module wiring |
| `cmd/api/Dockerfile` | Multi-stage production Docker build (scratch base) |
| `internal/modules/application/module.go` | HTTP server lifecycle and graceful shutdown |
| `internal/modules/config/module.go` | Configuration module |
| `internal/modules/database/module.go` | Database connection pool setup |
| `internal/middleware/error_handler.go` | Centralized Gin error handling middleware |
| `.env.example` | Reference for all environment variables |
| `Makefile` | Development commands |

## Testing

Run tests with:

```bash
make test
```

## Docker Notes

- The Dockerfile uses a multi-stage build: `golang:1.25-alpine` for building, `scratch` for the final image.
- The binary is statically compiled (`CGO_ENABLED=0`) and stripped (`-ldflags="-s -w"`).
- CA certificates and timezone data are copied from the builder stage into the `scratch` image.
- The container runs as a non-root user (`USER 1000`).

## Adding a New Module

1. Create a new directory under `internal/modules/<module_name>/`
2. Create a `module.go` with a `Module` struct and `NewModule()` constructor
3. If the module has HTTP endpoints, add handler files and implement `MapRoutes(router *gin.RouterGroup)`
4. Wire the module in `cmd/api/main.go`:
   - Initialize it with `NewModule()`
   - Register routes with `app.MapRoutes()` or `app.MapAPIRoutes()`
5. If the module needs cleanup on shutdown, register it with `app.OnBeforeShutdown()`
6. Update Swagger annotations and run `make swag` if new endpoints were added
