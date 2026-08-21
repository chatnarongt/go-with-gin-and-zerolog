---
name: generate-openapi
description: "Generate and validate OpenAPI 3 contracts for applications that import the shared Swagger module."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness. Requires git and Go.
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(git:*)
---

Generate or update OpenAPI 3 contracts for this monorepo's applications.

## Invocation

- `/generate-openapi` — process every application under `cmd/` whose `main.go` imports `internal/modules/swagger`.
- `/generate-openapi <application>` — process only `cmd/<application>`; for example, `/generate-openapi api`.

Reject more than one argument. For a named application, fail if `cmd/<application>/main.go` does not exist or does not import the shared Swagger module.

## Repository contract

- Each application owns its contract at `cmd/<application>/openapi.yaml`.
- The shared implementation lives at `internal/modules/swagger`.
- The application entry point embeds its contract with `//go:embed openapi.yaml` and passes the bytes through `swagger.ModuleOptions.Spec`.
- The contract is OpenAPI `3.0.3`, source-controlled, and manually reviewed. Do not add generated files or move API-specific contracts into the shared module.
- Runtime metadata (`Title`, `Version`, `Description`) comes from the application's `swagger.ModuleOptions`; keep the YAML `info` object valid with stable placeholder/default values.

## Workflow

1. Discover targets.
   - With no argument, inspect `cmd/*/main.go` and select only applications importing `internal/modules/swagger`.
   - With an argument, select only `cmd/<argument>` after validating its Swagger import.
2. Inspect each target's composition root and all imported modules. Trace every route registration through `RegisterRoutes`, `router.GET`, `router.POST`, and equivalent Gin methods. Read request/response structs and status branches used by those handlers.
3. Locate the target contract. If the application already embeds `openapi.yaml`, update that file. If the Swagger import exists but embedding/wiring or the contract is missing, add the minimum wiring and contract required by the existing module API before documenting routes.
4. Update only the target contract's `paths`, `components`, tags, and operation metadata needed to match the current implementation. Preserve existing descriptions, examples, operation IDs, and schemas when they remain accurate. Remove paths no longer registered. Document every success and failure status actually returned. Do not invent authentication, parameters, fields, or responses.
5. Keep OpenAPI valid:
   - `openapi: 3.0.3`.
   - Required `info.title` and `info.version`.
   - JSON media types matching the Gin handlers.
   - `$ref` targets present and schemas matching JSON tags and runtime values.
   - Relative `servers` URL unless the application explicitly requires a fixed host.
6. Verify each target with `gofmt` for changed Go files, `go vet ./...`, and `go build ./cmd/<application>`. If an application can run safely, enable `APP_SWAGGER_ENABLED=true`, request its configured `APP_SWAGGER_BASE_PATH` and `/openapi.json`, and validate the returned document. Do not leave processes running.
7. Report target applications, contract files changed, routes added/removed, and verification results.

## Scope rules

- No argument means all qualifying Swagger applications, not every directory under `cmd/`. A worker or CLI without the Swagger module is out of scope.
- A named application limits edits and verification to that application's contract and necessary embedding/wiring, though repository-wide `go vet ./...` may expose unrelated existing failures.
- Do not change `internal/modules/swagger` for an application-specific route or schema.
- Do not add dependencies, authentication schemes, or code-generation tools unless explicitly requested.
- Never overwrite a contract wholesale when a focused update is sufficient.
