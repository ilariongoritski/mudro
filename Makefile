DSN ?= postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable
<<<<<<< ours
MIGRATION ?= migrations/001_init.sql
AGENT_MIGRATION ?= migrations/002_agent_queue.sql
COMMENTS_MIGRATION ?= migrations/003_post_comments.sql
AGENT_REVIEW_MIGRATION ?= migrations/004_agent_review_gate.sql
AGENT_EVENTS_MIGRATION ?= migrations/005_agent_task_events.sql
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
=======
MIGRATIONS_DIR ?= migrations
>>>>>>> theirs

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
<<<<<<< ours
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
=======
	@for f in $(shell ls $(MIGRATIONS_DIR)/*.sql | sort); do \
		echo "==> applying $$f"; \
		psql "$(DSN)" -X -v ON_ERROR_STOP=1 -f "$$f" || exit $$?; \
	done
>>>>>>> theirs

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

<<<<<<< ours
worker-loop:
	./scripts/worker_autonomy_loop.sh

bot-run:
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_BOT)" ]; then . "$(ENV_BOT)"; fi; \
	set +a; \
	$(GO) run ./cmd/bot

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
=======

selftest:
	go test ./cmd/vkimport ./cmd/tgimport
>>>>>>> theirs
