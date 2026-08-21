# Repository Guide

## Commands

Run from the repository root. Go 1.26 is declared in `go.mod`.

```bash
gofmt -w $(find cmd internal -name '*.go')
go test ./...
go vet ./...
go build ./cmd/api ./cmd/worker
go run ./cmd/api
```

Run one package or test:

```bash
go test ./internal/modules/application
go test ./internal/modules/application -run TestName -v
```

The API runs on `APP_PORT` (default `8080`). Configuration loads the env files passed by the command, then process environment variables override file values. The API command currently passes `.env` and `cmd/api/.env`; both paths must be readable because the loader returns an error for a missing file. `APP_LOG_LEVEL` defaults to zerolog `info` and accepts values `-1` through `5`. Other supported settings include `APP_DEBUG_MODE`, `APP_SWAGGER_ENABLED` (default `false`; enables `/openapi.json` and Swagger UI), `APP_SWAGGER_BASE_PATH` (default `/swagger`), `DATABASE_DSN`, `DATABASE_MAX_OPEN_CONNS`, `DATABASE_MAX_IDLE_CONNS`, `DATABASE_MAX_LIFETIME`, and `DATABASE_MAX_IDLE_TIME`. The last two use Go duration strings; `0` means unlimited.

Useful local checks after starting the API:

```bash
curl http://localhost:8080/probe/liveness
curl http://localhost:8080/probe/readiness
```

## Architecture

This repository is a Go foundation for HTTP services. Gin handles HTTP routing, `samber/do/v2` owns runtime dependency injection, zerolog owns application logging, and SQLite is opened through `database/sql` with the modernc SQLite driver.

`cmd/api/main.go` is the composition root. It creates `application.Module`, passes config env-file paths, and imports feature modules. `cmd/worker/main.go` is currently only a package placeholder; do not assume it starts a worker.

Shared modules under `internal/modules` implement `internal.Module`:

- `application`: coordinates the complete lifecycle. It registers core modules, registers command-imported feature modules, initializes lifecycle hooks, starts `http.Server`, and performs graceful shutdown.
- `logger`: provides `*zerolog.Logger`, applying configured level, timestamps, console/json output, and hooks.
- `config`: loads env files and process environment, parses typed application/database groups, validates them, and provides `*config.Config`.
- `database`: creates the SQLite directory/file connection from config, provides `*sql.DB`, pings it during initialization, and closes it during destruction.
- `probe`: provides liveness/readiness service and controller, then maps `/probe/liveness` and `/probe/readiness`.

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
