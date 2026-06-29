APP_NAME=zhinux-db-maintainer

DB_URL?=postgres://postgres:postgres@localhost:5432/zhinux-db-maintainer?sslmode=disable

MIGRATE_CMD=go run ./cmd/migrate
SQLC_CMD=sqlc
PG_DUMP=pg_dump


.PHONY: help
help:
	@echo "Available targets:"
	@echo " make db-up            start local postgres"
	@echo " make migrate-up       apply migrations"
	@echo " make migrate-down     rollback one migration"
	@echo " make migrate-version  show migration version"
	@echo " make schema           build schema snapshot"
	@echo " make generate         run sqlc"
	@echo " make dev-setup        full local setup"
	@echo " make ci-check         CI validation"
	@echo " make docker-dev       Run dev env via docker compose"
	@echo " make docker-ci        Run ci env via docker compose"


.PHONY: db-up
db-up:
	docker compose up -d postgres


.PHONY: docker-dev
docker-dev:
		docker compose -f docker-compose.yaml -f docker-compose.dev.yaml up -d


.PHONY: docker-ci
docker-ci:
		docker compose -f docker-compose.yaml -f docker-compose.ci.yaml up


.PHONY: migrate-up
migrate-up:
	DATABASE_URL=$(DB_URL) $(MIGRATE_CMD) up


.PHONY: migrate-down
migrate-down:
	DATABASE_URL=$(DB_URL) $(MIGRATE_CMD) down


.PHONY: migrate-version
migrate-version:
	DATABASE_URL=$(DB_URL) $(MIGRATE_CMD) version


.PHONY: schema
schema:
	@echo "Building schema snapshot..."
	$(PG_DUMP) $(DB_URL) --schema-only --no-owner --no-privileges > db/postgres/schema.sql


.PHONY: generate
generate:
	$(SQLC_CMD) generate


.PHONY: dev-setup
dev-setup: docker-dev migrate-up schema generate


.PHONY: ci-check
ci-check: migrate-up schema generate
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
