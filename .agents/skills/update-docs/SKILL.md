---
name: update-docs
description: Update README.md, AGENTS.md, and this SKILL.md when new modules or code changes are introduced
---

# Update Documentation Skill

## Overview
This skill provides a systematic workflow for updating `README.md`, `AGENTS.md`, and itself (`SKILL.md`) whenever new modules, configuration variables, or significant code changes are added to the `go-with-gin-and-zerolog` project.

## Workflow

### 1. Analyze Recent Changes
Before making any updates, investigate the current state of the project.
- **Modules**: Run `list_dir` on `internal/modules/` to identify any new modules.
- **Configuration**: Review `.env.example` and `internal/modules/config/` for any new environment variables or setup steps.
- **Wiring**: Check both `cmd/api/main.go` and `cmd/worker/main.go` to see how modules are injected and whether startup order has changed.
- **Dependencies**: Look at `go.mod` for any new major libraries that should be documented.

### 2. Update `README.md`
Edit `README.md` to reflect the new state of the project.
- **Tech Stack**: Append any new critical technologies or frameworks used.
- **Project Structure**: Update the folder structure diagram if new module directories or core layers were added.
- **Environment Variables**: Make sure the environment variables table perfectly matches `.env.example`.
- **Development Commands**: If new `make` commands or Docker scripts were introduced, document them in the instructions.

### 3. Update `AGENTS.md`
Edit `AGENTS.md` to ensure future AI agents have accurate context about the architecture and rules.
- **Initialization Order**: Verify and update startup order for both `cmd/api/main.go` and `cmd/worker/main.go`.
- **Code Style & Conventions**: If a new coding standard was introduced (like a specific way to handle errors safely or a new logging sub-package requirement), append it.
- **Adding a New Module**: Update the step-by-step instructions if new boilerplate is required for creating a module.
- **Key Files**: Update the table with any new fundamental components (e.g., a new core middleware file).

### 4. Update This Skill (`SKILL.md`)
Edit `.agents/skills/update-docs/SKILL.md` if the workflow needs to be adapted.
- **New Documentation Files**: If new core organizational documentation files are introduced (like a `CONTRIBUTING.md`), update this workflow so future agents know to keep them in sync.
- **Architecture Shifts**: If the project significantly changes how things are wired, update the "Analyze Recent Changes" steps with new instructions.

### 5. Review and Finalize
- Use the `replace_file_content` or `multi_replace_file_content` tools to make atomic, precise changes to the markdown files without destroying existing unrelated documentation.
- Ensure all markdown tables are properly aligned and formatted.
