APP_NAME=zhinux-db-maintainer

# Try to get from environment, fallback to defaults
POSTGRES_USER?=postgres
POSTGRES_PASSWORD?=postgres
POSTGRES_DB_NAME?=zhinux-db-maintainer
POSTGRES_PORT?=5432
POSTGRES_HOST?=localhost
POSTGRES_DB_URL?=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB_NAME)?sslmode=disable

# Use Docker if local pg_dump not found
PG_DUMP_CMD=$(shell command -v pg_dump 2>/dev/null)

MIGRATE_CMD=go run ./cmd/migrate
SQLC_CMD=sqlc
# PG_DUMP=pg_dump


.PHONY: help
help:
	@echo "Available targets:"
	@echo " make db-up            start local postgres(with env support)"
	@echo " make migrate-up       apply migrations(with Docker fallback)"
	@echo " make migrate-down     rollback one migration"
	@echo " make migrate-version  show migration version"
	@echo " make migrate-reset    full reset db migrations and apply them from scratch" # TODO: Implement
	@echo " make schema           build schema snapshot(with Docker fallback)"
	@echo " make generate         run sqlc"
	@echo " make dev-setup        full local setup"
	@echo " make ci-check         CI validation with Docker"
	@echo " make docker-dev       Run dev env via docker compose"
	@echo " make docker-ci        Run ci env via docker compose"


.PHONY: db-up
db-up:
	docker compose -f docker-compose.dev.yaml up -d postgres
	# docker compose up -d postgres


.PHONY: db-up-ci
db-up-ci:
	docker compose -f docker-compose.ci.yaml up -d postgres


.PHONY: docker-dev
docker-dev:
		docker compose -f docker-compose.yaml -f docker-compose.dev.yaml up -d


.PHONY: docker-ci
docker-ci:
		docker compose -f docker-compose.yaml -f docker-compose.ci.yaml up


.PHONY: migrate-up
migrate-up:
	POSTGRES_DB_URL=$(POSTGRES_DB_URL) $(MIGRATE_CMD) up


.PHONY: migrate-down
migrate-down:
	POSTGRES_DB_URL=$(POSTGRES_DB_URL) $(MIGRATE_CMD) down


.PHONY: migrate-version
migrate-version:
	POSTGRES_DB_URL=$(POSTGRES_DB_URL) $(MIGRATE_CMD) version


.PHONY: schema
schema:
	@echo "Building schema snapshot..."
	# $(PG_DUMP) $(POSTGRES_DB_URL) --schema-only --no-owner --no-privileges > db/postgres/schema.sql
	@if [ -x "$(PG_DUMP_CMD)" ]; then \
		pg_dump $(POSTGRES_DB_URL) --schema-only --no-owner --no-privileges > db/postgres/schema.sql.tmp; \
	else \
		docker run --rm -e PGPASSWORD=$(POSTGRES_PASSWORD) \
			--network host postgres:15 \
			pg_dump -U $(POSTGRES_USER) -h $(POSTGRES_HOST) -p $(POSTGRES_PORT) $(POSTGRES_DB_NAME) \
			--schema-only --no-owner --no-privileges > db/postgres/schema.sql.tmp; \
	fi
	# Filter out lines that start with a backslash
	grep -v '^\\' db/postgres/schema.sql.tmp > db/postgres/schema.sql
	rm db/postgres/schema.sql.tmp


.PHONY: generate
generate:
	$(SQLC_CMD) generate


.PHONY: dev-setup
dev-setup: db-up migrate-up schema generate
# dev-setup: docker-dev migrate-up schema generate


.PHONY: ci-check
ci-check: db-up-ci migrate-up schema generate
	git diff --exit-code

# .PHONY: db-up db-down migrate-up migrate-down migrate-reset generate test ci

# db-up:
# 	docker compose up -d postgres

# db-down:
# 	docker compose down

# migrate-up:
# 	go run ./cmd/migrate up

# migrate-down:
# 	go run ./cmd/migrate down 1

# migrate-reset:
# 	go run ./cmd/migrate reset

# generate:
# 	sqlc generate

# test:
# 	go test ./...

# ci: generate test
# 	git diff --exit-code
