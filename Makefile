.DEFAULT_GOAL := run
ENV ?= development

.PHONY: env-%
env-%:
	$(eval ENV := $*)

.PHONY: run
run:
	@set -a; \
	if [ -f .env ]; then . ./.env; fi; \
	if [ -f .env.$(ENV) ]; then . ./.env.$(ENV); fi; \
	set +a; \
	trap '' INT TERM; \
	go run ./cmd/api/; \
	true

.PHONY: run-worker
run-worker:
	@set -a; \
	if [ -f .env ]; then . ./.env; fi; \
	if [ -f .env.$(ENV) ]; then . ./.env.$(ENV); fi; \
	set +a; \
	trap '' INT TERM; \
	go run ./cmd/worker/; \
	true

.PHONY: build
build:
	go build -o bin/api ./cmd/api

.PHONY: test
test:
	go test ./... -v

.PHONY: swag
swag:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go
