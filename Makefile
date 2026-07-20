# ==== Config ====
APP_NAME     := golangchatapp
CMD_DIR      := ./cmd/api
BIN_DIR      := ./bin
BINARY       := $(BIN_DIR)/$(APP_NAME)

ENV_FILE     := config/dev.env
DB_PATH      := sqlite/dev/api.db
MIGRATIONS_DIR := migrations

# Load env vars from dev.env if present, otherwise fall back to DB_PATH
DB_URL       ?= sqlite3://$(DB_PATH)

GOFLAGS      :=
GO           := go

.PHONY: all build run dev clean test tidy fmt vet \
        migrate-up migrate-down migrate-force migrate-version migrate-create \
        db-reset help

all: build

## Build the API binary
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(BINARY) $(CMD_DIR)

## Run the API directly (no binary artifact)
run:
	$(GO) run $(CMD_DIR) | true

## Run with dev.env loaded into the environment
dev:
	@set -a; . ./$(ENV_FILE); set +a; $(GO) run $(CMD_DIR)

## Run tests
test:
	$(GO) test ./... -v

## Tidy go.mod/go.sum
tidy:
	$(GO) mod tidy

## Format code
fmt:
	$(GO) fmt ./...

## Vet code
vet:
	$(GO) vet ./...

## Remove build artifacts
clean:
	rm -rf $(BIN_DIR)

# ==== Migrations (golang-migrate) ====
# Requires: go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

## Apply all up migrations
migrate-up:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) up

## Roll back the last migration
migrate-down:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) down 1

## Force set a migration version (usage: make migrate-force VERSION=3)
migrate-force:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) force $(VERSION)

## Show current migration version
migrate-version:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) version

## Create a new migration pair (usage: make migrate-create NAME=add_foo_table)
migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

## Wipe and recreate the dev sqlite db, then reapply all migrations
db-reset:
	rm -f sqlite/dev/api.db sqlite/dev/api.db-shm sqlite/dev/api.db-wal
	$(MAKE) migrate-up

## Show available targets
help:
	@grep -E '^## ' -A1 $(MAKEFILE_LIST) | grep -v '^--$$' | sed -e 's/## //' | paste -d' ' - -
