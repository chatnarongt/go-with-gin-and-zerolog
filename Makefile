.DEFAULT_GOAL := run

-include .env .env.local
export

.PHONY: run
run:
	@trap '' INT TERM; go run ./cmd/api/; true

.PHONY: build
build:
	go build -o bin/api ./cmd/api

.PHONY: test
test:
	go test ./... -v

.PHONY: swag
swag:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go
