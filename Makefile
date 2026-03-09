ENV ?= development
APP ?=
VALID_ENVS := development staging production
VALID_APPS := api worker
FIRST_VALID_APP := $(firstword $(VALID_APPS))

# Allow positional environment goal syntax, e.g. `make run api staging`.
GOAL_ENV := $(firstword $(filter $(VALID_ENVS),$(MAKECMDGOALS)))
ifneq ($(GOAL_ENV),)
ENV := $(GOAL_ENV)
endif

# Allow positional app goal syntax, e.g. `make run api`.
GOAL_APP := $(firstword $(filter $(VALID_APPS),$(MAKECMDGOALS)))
ifneq ($(GOAL_APP),)
APP := $(GOAL_APP)
endif

ifeq ($(APP),)
APP_REQUIRED_MSG := App is required. Use `make run <app>` (e.g. `make run $(FIRST_VALID_APP)`) or APP=$(FIRST_VALID_APP)
endif

ifneq ($(APP),)
ifeq (,$(filter $(APP),$(VALID_APPS)))
$(error Invalid APP '$(APP)'. Use `make run <app>` where <app> is one of [$(VALID_APPS)] (e.g. `make run $(FIRST_VALID_APP)`) or APP=$(FIRST_VALID_APP))
endif
endif

.PHONY: $(VALID_ENVS)
$(VALID_ENVS):
	@:

.PHONY: require-app
require-app:
ifeq ($(APP),)
	$(error $(APP_REQUIRED_MSG))
endif

.PHONY: run
run: require-app
	@set -a; \
	if [ -f .env ]; then . ./.env; fi; \
	if [ -f .env.$(ENV) ]; then . ./.env.$(ENV); fi; \
	if [ -f cmd/$(APP)/.env ]; then . ./cmd/$(APP)/.env; fi; \
	if [ -f cmd/$(APP)/.env.$(ENV) ]; then . ./cmd/$(APP)/.env.$(ENV); fi; \
	set +a; \
	trap '' INT TERM; \
	go run ./cmd/$(APP)/; \
	true

.PHONY: $(VALID_APPS)
$(VALID_APPS):
	@:

.PHONY: build
build: require-app
	go build -o bin/$(APP) ./cmd/$(APP)

.PHONY: test
test:
	go test ./... -v

.PHONY: swag
swag: require-app
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/$(APP)/main.go
