---
name: review-code
description: Perform structured code reviews focused on bugs, regressions, and missing tests
---

# Review Code Skill

## Overview
This skill defines a consistent workflow for reviewing code changes in `go-with-gin-and-zerolog`.

Use it when the goal is to assess correctness and risk, not to redesign architecture.
Prioritize actionable findings over summaries.

## Review Principles
- Focus on behavior, safety, and regressions first.
- Prefer concrete, reproducible findings over style preferences.
- Report findings with file and line references.
- Keep severity explicit: `high`, `medium`, `low`.
- If no issues are found, state that clearly and call out residual risks or testing gaps.

## Workflow

### 1. Gather Change Context
- Inspect changed files with `get_changed_files`.
- Read relevant files and nearby code paths, not only changed lines.
- If behavior changed at runtime boundaries, review handlers, middleware, module wiring, and config loading.

### 2. Validate Risk Areas
For each changed area, check:
- **Runtime behavior**: panic paths, nil handling, error propagation, status codes.
- **Module wiring**: initialization order and dependencies in `cmd/api/main.go` and `cmd/worker/main.go`.
- **Configuration safety**: defaults and validation consistency with `.env.example` and `internal/modules/config/`.
- **Database interactions**: connection lifecycle, cleanup hooks, timeout/pool settings, query safety.
- **Logging and observability**: structured logs with actionable context, no sensitive data leakage.
- **API contract drift**: request/response changes, Swagger annotation updates, backward compatibility.

### 3. Check Verification Coverage
- Run or inspect tests when feasible (`make test`).
- Identify missing tests for newly introduced branches, failure paths, and edge cases.
- If tests are not run, explicitly state that limitation.

### 4. Produce Findings-First Output
Report findings ordered by severity and include:
- Severity (`high`, `medium`, `low`)
- Impact (what can break and when)
- Evidence (`path:line`)
- Suggested fix (short and specific)

Use this output shape:

1. `[high] <title>`
   - Impact: <impact>
   - Evidence: `<path:line>`
   - Fix: <recommended change>

After findings, add:
- **Open Questions / Assumptions**
- **Change Summary** (brief)
- **Residual Risks / Testing Gaps**

## Repository-Specific Notes
- Prefer guard clauses for clearer control flow in handlers/services.
- Use `github.com/rs/zerolog/log` for logging, not `fmt.Printf`.
- Keep route registration split correctly:
  - Root routes via `app.MapRoutes(...)`
  - API routes via `app.MapAPIRoutes(...)`
- Keep startup-order documentation accurate for both entrypoints:
  - API: `config -> database -> swagger -> health -> application`
  - Worker: `config -> database -> schedule -> application`
