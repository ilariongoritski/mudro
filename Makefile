DSN ?= postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable
MIGRATIONS_DIR ?= migrations
MIGRATION ?= $(MIGRATIONS_DIR)/001_init.sql
AGENT_MIGRATION ?= $(MIGRATIONS_DIR)/002_agent_queue.sql
COMMENTS_MIGRATION ?= $(MIGRATIONS_DIR)/003_post_comments.sql
AGENT_REVIEW_MIGRATION ?= $(MIGRATIONS_DIR)/004_agent_review_gate.sql
AGENT_EVENTS_MIGRATION ?= $(MIGRATIONS_DIR)/005_agent_task_events.sql
MEDIA_MIGRATION ?= $(MIGRATIONS_DIR)/006_media_assets.sql
MEDIA_FIX_MIGRATION ?= $(MIGRATIONS_DIR)/007_media_link_constraints.sql
COMMENT_MODEL_MIGRATION ?= $(MIGRATIONS_DIR)/008_comment_model.sql
USERS_AUTH_MIGRATION ?= $(MIGRATIONS_DIR)/009_users_and_auth.sql
CASINO_MIGRATION ?= services/casino/migrations/001_init.sql
USE_DOCKER_PSQL ?= 1
GO ?= /usr/local/go/bin/go
ENV_COMMON ?= env/common.env
ENV_API ?= env/api.env
ENV_AGENT ?= env/agent.env
ENV_BOT ?= env/bot.env
ENV_REPORTER ?= env/reporter.env
ENV_CASINO ?= env/casino.env

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
	cat "$(USERS_AUTH_MIGRATION)" | docker compose exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1
else
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$(MIGRATION)"
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$(USERS_AUTH_MIGRATION)"
endif

migrate-all:
	@for f in $(shell ls $(MIGRATIONS_DIR)/*.sql | sort); do \
		echo "==> applying $$f"; \
		if [ "$(USE_DOCKER_PSQL)" = "1" ]; then cat "$$f" | docker compose exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1; else $(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$$f"; fi || exit $$?; \
	done

migrate-agent:
ifeq ($(USE_DOCKER_PSQL),1)
	cat "$(AGENT_MIGRATION)" | docker compose exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1
else
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$(AGENT_MIGRATION)"
endif

migrate-comments:
ifeq ($(USE_DOCKER_PSQL),1)
	cat "$(COMMENTS_MIGRATION)" | docker compose exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1
else
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$(COMMENTS_MIGRATION)"
endif

migrate-agent-review:
ifeq ($(USE_DOCKER_PSQL),1)
	cat "$(AGENT_REVIEW_MIGRATION)" | docker compose exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1
else
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$(AGENT_REVIEW_MIGRATION)"
endif

migrate-agent-events:
ifeq ($(USE_DOCKER_PSQL),1)
	cat "$(AGENT_EVENTS_MIGRATION)" | docker compose exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1
else
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$(AGENT_EVENTS_MIGRATION)"
endif

migrate-media:
ifeq ($(USE_DOCKER_PSQL),1)
	cat "$(MEDIA_MIGRATION)" | docker compose exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1
	cat "$(MEDIA_FIX_MIGRATION)" | docker compose exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1
else
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$(MEDIA_MIGRATION)"
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$(MEDIA_FIX_MIGRATION)"
endif

migrate-comment-model:
ifeq ($(USE_DOCKER_PSQL),1)
	cat "$(COMMENT_MODEL_MIGRATION)" | docker compose exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1
else
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$(COMMENT_MODEL_MIGRATION)"
endif

migrate-users-auth:
ifeq ($(USE_DOCKER_PSQL),1)
	cat "$(USERS_AUTH_MIGRATION)" | docker compose exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1
else
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$(USERS_AUTH_MIGRATION)"
endif

migrate-casino:
	bash ./scripts/migrate-casino.sh

tables:
	$(PSQL_CMD) -X -c "\\dt"

test:
	$(GO) test ./...

lint:
	golangci-lint run ./...
	cd frontend && npm run lint

check: lint test

selftest:
	$(GO) test ./cmd/vkimport ./cmd/tgimport

media-backfill:
	$(GO) run ./cmd/mediabackfill

comment-backfill:
	$(GO) run ./cmd/commentbackfill

tg-csv-import:
	@if [ -z "$(CSV)" ]; then echo "CSV is required"; exit 1; fi
	$(GO) run ./cmd/tgcsvimport -in "$(CSV)" -dsn "$(DSN)"

tg-comments-csv-import:
	@if [ -z "$(CSV)" ]; then echo "CSV is required"; exit 1; fi
	$(GO) run ./cmd/tgcommentscsvimport -in "$(CSV)" -dsn "$(DSN)"

tg-comment-media-import:
	@if [ -z "$(DIR)" ]; then echo "DIR is required"; exit 1; fi
	$(GO) run ./cmd/tgcommentmediaimport -dir "$(DIR)" -dsn "$(DSN)"

count-posts:
	$(PSQL_CMD) -X -c "select count(*) from posts;"

casino-dbcheck:
	@if [ "$(USE_DOCKER_PSQL)" = "1" ]; then \
		docker compose exec -T casino-db psql -U postgres -d mudro_casino -X -v ON_ERROR_STOP=1 -e -a -c "select 1;"; \
	else \
		psql "$${CASINO_DSN:-postgres://postgres:postgres@localhost:5434/mudro_casino?sslmode=disable}" -X -v ON_ERROR_STOP=1 -e -a -c "select 1;"; \
	fi

health:
	$(MAKE) up
	$(MAKE) ps
	$(MAKE) dbcheck
	$(MAKE) casino-dbcheck
	$(MAKE) migrate
	$(MAKE) migrate-casino
	$(MAKE) tables
	$(MAKE) test
	$(MAKE) count-posts

worker-loop:
	./scripts/worker_autonomy_loop.sh

orchestration-log-init:
	@bash ./scripts/orchestration_run_init.sh "$(RUN_ID)" "$(TASK)"

bot-run:
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_BOT)" ]; then . "$(ENV_BOT)"; fi; \
	set +a; \
	$(GO) run ./cmd/bot

casino-run:
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_CASINO)" ]; then . "$(ENV_CASINO)"; fi; \
	set +a; \
	$(GO) run ./services/casino/cmd/casino

report-run:
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_REPORTER)" ]; then . "$(ENV_REPORTER)"; fi; \
	set +a; \
	$(GO) run ./cmd/reporter

memento:
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	set +a; \
	GOCACHE=/tmp/go-build-cache $(GO) run ./cmd/memento

agent-plan-once:
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_AGENT)" ]; then . "$(ENV_AGENT)"; fi; \
	set +a; \
	$(GO) run ./cmd/agent --mode once

agent-plan:
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_AGENT)" ]; then . "$(ENV_AGENT)"; fi; \
	set +a; \
	$(GO) run ./cmd/agent --mode planner --interval 1m

agent-work:
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_AGENT)" ]; then . "$(ENV_AGENT)"; fi; \
	set +a; \
	$(GO) run ./cmd/agent --mode worker --interval 15s

agent-approve:
	@if [ -z "$(TASK_ID)" ]; then echo "TASK_ID is required"; exit 1; fi
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_AGENT)" ]; then . "$(ENV_AGENT)"; fi; \
	set +a; \
	$(GO) run ./cmd/agent --mode approve --task-id "$(TASK_ID)"

agent-reject:
	@if [ -z "$(TASK_ID)" ]; then echo "TASK_ID is required"; exit 1; fi
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_AGENT)" ]; then . "$(ENV_AGENT)"; fi; \
	set +a; \
	$(GO) run ./cmd/agent --mode reject --task-id "$(TASK_ID)" --reason "$(REASON)"
