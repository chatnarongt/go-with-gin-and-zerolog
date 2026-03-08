.DEFAULT_GOAL := run

.PHONY: run
run:
	@set -a; \
	if [ -f .env ]; then . ./.env; fi; \
	if [ -f .env.local ]; then . ./.env.local; fi; \
	set +a; \
	trap '' INT TERM; \
	go run ./cmd/api/; \
	true

.PHONY: run-test
run-test:
	@set -a; \
	if [ -f .env ]; then . ./.env; fi; \
	if [ -f .env.test ]; then . ./.env.test; fi; \
	set +a; \
	trap '' INT TERM; \
	go run ./cmd/api/; \
	true

.PHONY: run-staging
run-staging:
	@set -a; \
	if [ -f .env ]; then . ./.env; fi; \
	if [ -f .env.staging ]; then . ./.env.staging; fi; \
	set +a; \
	trap '' INT TERM; \
	go run ./cmd/api/; \
	true

.PHONY: run-worker
run-worker:
	@set -a; \
	if [ -f .env ]; then . ./.env; fi; \
	if [ -f .env.local ]; then . ./.env.local; fi; \
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
