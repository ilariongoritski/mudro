DSN ?= postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable
MIGRATION ?= migrations/001_init.sql
AGENT_MIGRATION ?= migrations/002_agent_queue.sql
USE_DOCKER_PSQL ?= 1
GO ?= /usr/local/go/bin/go

ifeq ($(shell [ -x "$(GO)" ] && echo 1 || echo 0),0)
GO := go
endif

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

migrate-agent:
ifeq ($(USE_DOCKER_PSQL),1)
	cat "$(AGENT_MIGRATION)" | docker compose exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1
else
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$(AGENT_MIGRATION)"
endif

tables:
	$(PSQL_CMD) -X -c "\dt"

test:
	$(GO) test ./...

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
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; $(GO) run ./cmd/bot

report-run:
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; $(GO) run ./cmd/reporter

agent-plan-once:
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; $(GO) run ./cmd/agent --mode once

agent-plan:
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; $(GO) run ./cmd/agent --mode planner --interval 1m

agent-work:
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; $(GO) run ./cmd/agent --mode worker --interval 15s
