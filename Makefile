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
USE_DOCKER_PSQL ?= 1
GO ?= /usr/local/go/bin/go
ENV_COMMON ?= env/common.env
ENV_API ?= env/api.env
ENV_AGENT ?= env/agent.env
ENV_BOT ?= env/bot.env
ENV_REPORTER ?= env/reporter.env

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

tables:
	$(PSQL_CMD) -X -c "\\dt"

test:
	$(GO) test ./...

test-no-tmp:
	@PKGS=`$(GO) list ./... | grep -v '/tmp$$'`; \
	$(GO) test $$PKGS

selftest:
	$(GO) test ./tools/importers/vkimport/app ./tools/importers/tgimport/app

media-backfill:
	$(GO) run ./tools/backfill/mediabackfill/cmd

comment-backfill:
	$(GO) run ./tools/backfill/commentbackfill/cmd

tg-csv-import:
	@if [ -z "$(CSV)" ]; then echo "CSV is required"; exit 1; fi
	$(GO) run ./tools/importers/tgcsvimport/cmd -in "$(CSV)" -dsn "$(DSN)"

tg-comments-csv-import:
	@if [ -z "$(CSV)" ]; then echo "CSV is required"; exit 1; fi
	$(GO) run ./tools/importers/tgcommentscsvimport/cmd -in "$(CSV)" -dsn "$(DSN)"

tg-comment-media-import:
	@if [ -z "$(DIR)" ]; then echo "DIR is required"; exit 1; fi
	$(GO) run ./tools/importers/tgcommentmediaimport/cmd -dir "$(DIR)" -dsn "$(DSN)"

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

worker-loop:
	./ops/scripts/worker_autonomy_loop.sh

bot-run:
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_BOT)" ]; then . "$(ENV_BOT)"; fi; \
	set +a; \
	$(GO) run ./services/bot/cmd

report-run:
	@echo "DEPRECATED: report-run is Old. Use legacy-report-run if you really need reporter."
	$(MAKE) legacy-report-run

legacy-report-run:
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_REPORTER)" ]; then . "$(ENV_REPORTER)"; fi; \
	set +a; \
	$(GO) run ./legacy/old/services/reporter-old/cmd

memento:
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	set +a; \
	GOCACHE=/tmp/go-build-cache $(GO) run ./tools/maintenance/memento/cmd

agent-plan-once:
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_AGENT)" ]; then . "$(ENV_AGENT)"; fi; \
	set +a; \
	$(GO) run ./services/agent/cmd --mode once

agent-plan:
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_AGENT)" ]; then . "$(ENV_AGENT)"; fi; \
	set +a; \
	$(GO) run ./services/agent/cmd --mode planner --interval 1m

agent-work:
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_AGENT)" ]; then . "$(ENV_AGENT)"; fi; \
	set +a; \
	$(GO) run ./services/agent/cmd --mode worker --interval 15s

agent-approve:
	@if [ -z "$(TASK_ID)" ]; then echo "TASK_ID is required"; exit 1; fi
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_AGENT)" ]; then . "$(ENV_AGENT)"; fi; \
	set +a; \
	$(GO) run ./services/agent/cmd --mode approve --task-id "$(TASK_ID)"

agent-reject:
	@if [ -z "$(TASK_ID)" ]; then echo "TASK_ID is required"; exit 1; fi
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_AGENT)" ]; then . "$(ENV_AGENT)"; fi; \
	set +a; \
	$(GO) run ./services/agent/cmd --mode reject --task-id "$(TASK_ID)" --reason "$(REASON)"
