-include .env
export $(shell sed 's/=.*//' .env)

APP_NAME := estatehub_api
CMD_PATH := cmd/api/main.go
BIN_PATH := bin/api
MIGRATIONS_DIR := $(MIGRATION_DIR)
MIGRATE_BIN := migrate
PG_URL := "postgres://$(PG_USER):$(PG_PASS)@${PG_HOST}:$(PG_PORT)/$(PG_NAME)?sslmode=disable"
GREEN := "\033[1;32m"
RESET := "\033[0m"

help: ## Mostra esta ajuda
	@printf "\n\033[1;34mComandos disponíveis:\033[0m\n\n"
	@grep -E '^[a-zA-Z0-9_-]+:.*## ' Makefile \
	| sed -E 's/(.*):.*## (.*)/  \1@\2/' \
	| column -t -s '@'
	@printf "\n"

migrate-create: ## Cria nova migration: make migrate-create name=create_users_table
	@echo $(GREEN)" Criando nova migration: $(name)"$(RESET)
	$(MIGRATE_BIN) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

migrate-up: ## Executa migrations UP
	docker compose run --rm migrate \
		-path=/migrations \
		-database "$(PG_URL)" \
		up

migrate-down: ## Executa migrations DOWN
	docker compose run --rm migrate \
		-path=/migrations \
		-database "$(PG_URL)" \
		down

migrate-redo: ## Reverte e reaplica a última migration
	docker compose run --rm migrate \
		-path=/migrations \
		-database "$(PG_URL)" \
		down 1

	docker compose run --rm migrate \
		-path=/migrations \
		-database "$(PG_URL)" \
		up 1

migrate-version: ## Mostra versão do schema
	docker compose run --rm migrate \
		-path=/migrations \
		-database "$(PG_URL)" \
		version

migrate-rollback: ## Reverte apenas a última migration
	docker compose run --rm migrate \
		-path=/migrations \
		-database "$(DOCKER_PG_URL)" \
		down 1