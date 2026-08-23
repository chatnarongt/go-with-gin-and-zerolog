# Repository Guide

## Commands

Run from the repository root. Go 1.26 is declared in `go.mod`.

```bash
gofmt -w $(find cmd internal -name '*.go')
go test ./...
go vet ./...
go build ./cmd/api ./cmd/worker
go run ./cmd/api
go run ./cmd/worker --job greeting:greeting
```

Run one package or test:

```bash
go test ./internal/modules/application
go test ./internal/modules/application -run TestName -v
```

Unit tests follow a consolidated convention: all tests for a package live in a single `unit_test.go` file within that package (e.g. `internal/middleware/unit_test.go`).

The API runs on `APP_PORT` (default `8080`). Configuration loads the env files passed by the command, then process environment variables override file values. The API command currently passes `.env` and `cmd/api/.env`; both paths must be readable because the loader returns an error for a missing file. `APP_LOG_LEVEL` defaults to zerolog `info` and accepts values `-1` through `5`. Other supported settings include `APP_ENV` (`development`, `staging`, `production`), `APP_DEBUG_MODE`, `APP_CORS_ENABLED` (default `true`), `APP_CORS_ALLOWED_ORIGINS` (`*`), `APP_CORS_ALLOWED_METHODS`, `APP_CORS_ALLOWED_HEADERS`, `APP_COMPRESSION_ENABLED` (default `false`), `APP_COMPRESSION_ENCODINGS` (`zstd,br,gzip`), `APP_COMPRESSION_MIN_BYTES` (default `1024`), `APP_SWAGGER_ENABLED` (default `false`; enables `/openapi.json` and Swagger UI), `APP_SWAGGER_BASE_PATH` (default `/swagger`), `APP_SWAGGER_SERVER_URL`, as well as per-database settings for `Main` (`DB_MAIN_*`), `Analytics` (`DB_ANALYTICS_*`), and `Logging` (`DB_LOGGING_*`): `*_DRIVER` (`sqlite`, `postgres`, `pgx`, `sqlserver`, `mssql`, `mongodb`, `mongo`), `*_DSN`, `*_REQUIRED` (`true`, `false`), `*_READ_ONLY` (`true`, `false`), `*_MAX_OPEN_CONNS`, `*_MAX_IDLE_CONNS`, `*_MAX_LIFETIME`, and `*_MAX_IDLE_TIME`. The last two use Go duration strings; `0` means unlimited. MongoDB ignores `*_MAX_LIFETIME`.

Useful local checks after starting the API:

```bash
curl http://localhost:8080/probe/liveness
curl http://localhost:8080/probe/readiness
```

## Architecture

This repository is a Go foundation for HTTP services. Gin handles HTTP routing, `samber/do/v2` owns runtime dependency injection, zerolog owns application logging, and database connections are opened through `database/sql` using drivers for SQLite (`modernc.org/sqlite`), PostgreSQL (`github.com/jackc/pgx/v5/stdlib`), and MSSQL (`github.com/microsoft/go-mssqldb`), plus MongoDB via `go.mongodb.org/mongo-driver/v2`.

`cmd/api/main.go` and `cmd/worker/main.go` are composition roots. `cmd/api` runs the HTTP server with `application.Module`, and `cmd/worker` executes CLI jobs with `worker.Module` via `--job <name>`.

Shared modules under `internal/modules` implement `internal.Module`:

- `application`: coordinates HTTP server lifecycle. Registers core modules, registers command-imported feature modules, initializes lifecycle hooks, starts `http.Server`, and performs graceful shutdown.
- `worker`: coordinates CLI job execution. Registers core modules, discovers jobs via `internal.JobRegistrar` and `internal.JobRegistry`, runs the target job specified by `--job`, and handles graceful shutdown.
- `logger`: provides `*zerolog.Logger`, applying configured level, timestamps, console/json output, and hooks.
- `config`: loads env files and process environment, parses typed application/database groups, validates them, and provides `*config.Config`.
- `database`: opens configured database connections (`Main`, `Analytics`, `Logging`) supporting SQLite, PostgreSQL, MSSQL, and MongoDB, provides `*database.Databases`, named `*sql.DB`, `map[string]*sql.DB`, and `*mongo.Client` / `*mongo.Database` (named `"logging"` when configured), pings required databases during initialization, and closes/disconnects them during destruction.
- `probe`: provides liveness/readiness service and controller, then maps `/probe/liveness` and `/probe/readiness`.
- `swagger`: serves OpenAPI spec and Swagger UI when enabled.
- `greeting`: sample feature module implementing HTTP endpoints and `internal.JobRegistrar` (`greeting:greeting`).

Registration order is significant. `application.Module.coreModules()` registers logger before config because the config module logs from `OnModuleInit()` and resolves the logger after providing config. Feature imports register afterward. Providers may resolve dependencies with `do.MustInvoke` during registration/provider execution; missing services fail fast. Do not silently deduplicate module imports or routes.

Lifecycle flow:

1. `StartContext` creates a fresh DI injector.
2. Core modules register in logger-then-config order.
3. Imported feature modules register providers and routes.
4. `OnModuleInit()` hooks run in registration order.
5. Gin runs behind `http.Server` until context cancellation or server failure.
6. Shutdown drains HTTP requests, shuts down the injector, then runs imported module destruction hooks in reverse import order with a 10-second deadline.

Lifecycle interfaces (`internal.Module`, `internal.OnModuleInit`, `internal.OnModuleDestroy`, and `internal.Controller`) live in `internal/type_module.go`. Initialization and destruction errors stop or aggregate according to the application lifecycle code.

Configuration groups belong in separate `config_*.go` files within `internal/modules/config` and are embedded in `config.Config`. Parse external strings into typed values, then validate the resulting group. When adding a feature module, define its module, register its providers and routes in `Register`, implement lifecycle hooks only when needed, and add the module to the command's `Imports` rather than hardcoding it into shared module internals.
