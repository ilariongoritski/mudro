DSN ?= postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable
MIGRATION ?= migrations/001_init.sql
USE_DOCKER_PSQL ?= 1

ifeq ($(USE_DOCKER_PSQL),1)
PSQL_CMD = docker compose exec -T db psql -U postgres -d gallery
else
PSQL_CMD = psql "$(DSN)"
endif

up:
	docker compose up -d

down:
	docker compose down

ps:
	docker compose ps

logs:
	docker compose logs --no-color --tail=200

dbcheck:
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -e -a -c "select 1;"

migrate:
ifeq ($(USE_DOCKER_PSQL),1)
	cat "$(MIGRATION)" | docker compose exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1
else
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$(MIGRATION)"
endif

tables:
	$(PSQL_CMD) -X -c "\dt"

test:
	go test ./...

count-posts:
	$(PSQL_CMD) -X -c "select count(*) from posts;"

health:
	$(MAKE) up
	$(MAKE) ps
	$(MAKE) dbcheck
	$(MAKE) migrate
	$(MAKE) tables
	$(MAKE) test
	$(MAKE) count-posts

bot-run:
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; go run ./cmd/bot
