DSN ?= postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable
MIGRATIONS_DIR ?= migrations
MIGRATION ?= $(MIGRATIONS_DIR)/001_init.sql
ACCOUNT_LIKES_MIGRATION ?= $(MIGRATIONS_DIR)/002_account_post_likes.sql
AGENT_MIGRATION ?= $(MIGRATIONS_DIR)/002b_agent_queue.sql
COMMENTS_MIGRATION ?= $(MIGRATIONS_DIR)/004_post_comments.sql
AGENT_REVIEW_MIGRATION ?= $(MIGRATIONS_DIR)/005_agent_review_gate.sql
AGENT_EVENTS_MIGRATION ?= $(MIGRATIONS_DIR)/006_agent_task_events.sql
MEDIA_MIGRATION ?= $(MIGRATIONS_DIR)/007_media_assets.sql
MEDIA_FIX_MIGRATION ?= $(MIGRATIONS_DIR)/008_media_link_constraints.sql
COMMENT_MODEL_MIGRATION ?= $(MIGRATIONS_DIR)/009_comment_model.sql
USERS_AUTH_MIGRATION ?= $(MIGRATIONS_DIR)/009_users_and_auth.sql
USERS_TELEGRAM_MIGRATION ?= $(MIGRATIONS_DIR)/015_users_telegram.sql
CHAT_MIGRATION ?= $(MIGRATIONS_DIR)/016_chat.sql
CASINO_EMOJI_MIGRATION ?= $(MIGRATIONS_DIR)/017_casino_emoji_v2.sql
CASINO_MIGRATION ?= services/casino/migrations/001_init.sql
USE_DOCKER_PSQL ?= 1
GO ?= /usr/local/go/bin/go
CORE_COMPOSE_FILE ?= ops/compose/docker-compose.core.yml
CORE_COMPOSE = docker compose -f $(CORE_COMPOSE_FILE)
SERVICES_COMPOSE_FILE ?= ops/compose/docker-compose.services.yml
MICRO_COMPOSE = docker compose -f $(CORE_COMPOSE_FILE) -f $(SERVICES_COMPOSE_FILE)
ENV_COMMON ?= env/common.env
ENV_API ?= env/api.env
ENV_AGENT ?= env/agent.env
ENV_BOT ?= env/bot.env
ENV_REPORTER ?= env/reporter.env
ENV_CASINO ?= env/casino.env
MOVIE_CATALOG_MIGRATION ?= $(MIGRATIONS_DIR)/movie_catalog/0001_init.sql
RUNTIME_MIGRATIONS ?= $(MIGRATION) $(ACCOUNT_LIKES_MIGRATION) $(AGENT_MIGRATION) $(COMMENTS_MIGRATION) $(AGENT_REVIEW_MIGRATION) $(AGENT_EVENTS_MIGRATION) $(MEDIA_MIGRATION) $(MEDIA_FIX_MIGRATION) $(COMMENT_MODEL_MIGRATION) $(USERS_AUTH_MIGRATION) $(USERS_TELEGRAM_MIGRATION) $(CHAT_MIGRATION) $(CASINO_MIGRATION) $(MOVIE_CATALOG_MIGRATION) $(CASINO_EMOJI_MIGRATION)

ifeq ($(wildcard $(GO)),)
GO := go
endif

ifeq ($(OS),Windows_NT)
GUARD_MAIN_CLEAN_SCRIPT = powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/git/guard_clean_main.ps1
else
GUARD_MAIN_CLEAN_SCRIPT = bash ./scripts/git/guard_clean_main.sh
endif

ifeq ($(USE_DOCKER_PSQL),1)
PSQL_CMD = docker compose exec -T db psql -U postgres -d gallery
PSQL_CORE_CMD = $(CORE_COMPOSE) exec -T db psql -U postgres -d gallery
else
PSQL_CMD = psql "$(DSN)"
PSQL_CORE_CMD = psql "$(DSN)"
endif

up:
	docker compose up -d

down:
	docker compose down

ps:
	docker compose ps

logs:
	docker compose logs --no-color --tail=200

core-up:
	$(CORE_COMPOSE) up -d

core-down:
	$(CORE_COMPOSE) down

core-ps:
	$(CORE_COMPOSE) ps

core-logs:
	$(CORE_COMPOSE) logs --no-color --tail=200

micro-up:
	$(MICRO_COMPOSE) up -d

micro-down:
	$(MICRO_COMPOSE) down

micro-ps:
	$(MICRO_COMPOSE) ps

micro-logs:
	$(MICRO_COMPOSE) logs --no-color --tail=200

dbcheck:
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -e -a -c "select 1;"

dbcheck-core:
	$(PSQL_CORE_CMD) -X -v ON_ERROR_STOP=1 -e -a -c "select 1;"

migrate:
ifeq ($(USE_DOCKER_PSQL),1)
	cat "$(MIGRATION)" | docker compose exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1
	cat "$(USERS_AUTH_MIGRATION)" | docker compose exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1
	cat "$(USERS_TELEGRAM_MIGRATION)" | docker compose exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1
else
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$(MIGRATION)"
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$(USERS_AUTH_MIGRATION)"
	$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$(USERS_TELEGRAM_MIGRATION)"
endif

migrate-runtime:
	@for f in $(RUNTIME_MIGRATIONS); do \
		echo "==> applying $$f"; \
		if [ "$(USE_DOCKER_PSQL)" = "1" ]; then \
			cat "$$f" | $(CORE_COMPOSE) exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1; \
		else \
			$(PSQL_CORE_CMD) -X -v ON_ERROR_STOP=1 -f "$$f"; \
		fi || exit $$?; \
	done

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

tables-core:
	$(PSQL_CORE_CMD) -X -c "\\dt"

test:
	$(GO) test ./...

test-active:
	$(GO) test ./...

lint:
	golangci-lint run ./...
	cd frontend && npm run lint

check: lint test

validate-contracts:
	$(GO) run ./tools/validate-contracts -dir ./contracts

validate-microservices:
	$(MAKE) test-active
	$(MAKE) validate-contracts
	$(MAKE) validate-systemd-templates
	docker compose -f $(CORE_COMPOSE_FILE) -f $(SERVICES_COMPOSE_FILE) config >/dev/null
	python -m py_compile ./scripts/claude/run_role_matrix.py

validate-systemd-templates:
	bash -n ./ops/scripts/install_mudro_systemd.sh
	bash -n ./ops/scripts/install_mudro_api_systemd.sh
	bash -n ./ops/scripts/install_openclaw_systemd.sh
	bash -n ./ops/scripts/harden_vps_db_auth.sh
	bash -n ./scripts/openclaw/openclaw_gateway_systemd.sh
	bash -n ./scripts/skaro/skaro_ui_linux.sh
	test -f ./ops/systemd/mudro-api.service
	test -f ./ops/systemd/mudro-bot.service
	test -f ./ops/systemd/mudro-agent-worker.service
	test -f ./ops/systemd/mudro-agent-planner.service
	test -f ./ops/systemd/mudro-agent-planner.timer
	test -f ./ops/systemd/openclaw.service
	test -f ./ops/systemd/skaro.service
	test -f ./ops/systemd/mudro-api.env.example
	test -f ./ops/systemd/mudro-bot.env.example
	test -f ./ops/systemd/mudro-agent.env.example
	test -f ./ops/systemd/openclaw.env.example
	test -f ./ops/systemd/skaro.env.example

install-mudro-systemd:
	bash ./ops/scripts/install_mudro_systemd.sh

install-openclaw-systemd:
	bash ./ops/scripts/install_openclaw_systemd.sh

guard-main-clean:
	$(GUARD_MAIN_CLEAN_SCRIPT)

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

count-posts-core:
	$(PSQL_CORE_CMD) -X -c "select count(*) from posts;"

casino-dbcheck:
	@if [ "$(USE_DOCKER_PSQL)" = "1" ]; then \
		docker compose exec -T casino-db psql -U postgres -d mudro_casino -X -v ON_ERROR_STOP=1 -e -a -c "select 1;"; \
	else \
		psql "$${CASINO_DSN:-postgres://postgres:postgres@localhost:5434/mudro_casino?sslmode=disable}" -X -v ON_ERROR_STOP=1 -e -a -c "select 1;"; \
	fi

health-runtime:
	$(MAKE) core-up
	$(MAKE) core-ps
	$(MAKE) dbcheck-core
	$(MAKE) migrate-runtime
	$(MAKE) tables-core
	$(MAKE) test-active
	$(MAKE) count-posts-core

health: health-runtime

demo-up:
	$(MAKE) health-runtime
	$(MAKE) demo-seed

demo-seed:
	@COUNT=`$(PSQL_CORE_CMD) -X -A -t -c "select count(*) from posts;"`; \
	if [ "$$COUNT" = "0" ]; then \
		if [ ! -f "$(DEMO_FEED_ITEMS)" ]; then \
			echo "demo seed file not found: $(DEMO_FEED_ITEMS)"; \
			exit 1; \
		fi; \
		echo "seeding demo posts from $(DEMO_FEED_ITEMS)"; \
		$(GO) run ./tools/importers/tgload/cmd -in "$(DEMO_FEED_ITEMS)" -dsn "$(DSN)"; \
	else \
		echo "demo seed skipped: posts=$$COUNT"; \
	fi

demo-check:
	@curl -fsS http://127.0.0.1:8080/healthz >/dev/null && echo "api healthz: ok"
	@curl -fsS "http://127.0.0.1:8080/api/front?limit=1" | grep -q '"total_posts":[[:space:]]*[1-9]' && echo "api feed: non-empty" || (echo "api feed: empty"; exit 1)
	@if curl -fsS http://127.0.0.1:5173 >/dev/null 2>&1; then \
		echo "frontend: reachable at http://127.0.0.1:5173"; \
	else \
		echo "frontend: not reachable from current shell (start in this shell or open http://127.0.0.1:5173 if frontend runs in Windows host)"; \
	fi

movie-catalog-run:
	$(GO) run ./services/movie-catalog/cmd

movie-catalog-migrate:
	@if [ "$(USE_DOCKER_PSQL)" = "1" ]; then \
		cat "$(MOVIE_CATALOG_MIGRATION)" | $(CORE_COMPOSE) exec -T db psql -U postgres -d gallery -X -v ON_ERROR_STOP=1; \
	else \
		$(PSQL_CMD) -X -v ON_ERROR_STOP=1 -f "$(MOVIE_CATALOG_MIGRATION)"; \
	fi

bff-web-run:
	$(GO) run ./services/bff-web/cmd

worker-loop:
	./ops/scripts/worker_autonomy_loop.sh

orchestration-log-init:
	@bash ./scripts/orchestration_run_init.sh "$(RUN_ID)" "$(TASK)"

claude-role-matrix:
	python ./scripts/claude/run_role_matrix.py --task "$(TASK)" --repo-root . $(CLAUDE_ROLE_MATRIX_ARGS)

openclaw-gateway-service:
	bash ./scripts/openclaw/openclaw_gateway_user_service.sh

openclaw-post-install-checks:
	bash ./scripts/openclaw/openclaw_post_install_checks.sh

bot-run:
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_BOT)" ]; then . "$(ENV_BOT)"; fi; \
	set +a; \
	$(GO) run ./services/bot/cmd

casino-run:
	@set -a; \
	if [ -f ./.env ]; then . ./.env; fi; \
	if [ -f "$(ENV_COMMON)" ]; then . "$(ENV_COMMON)"; fi; \
	if [ -f "$(ENV_CASINO)" ]; then . "$(ENV_CASINO)"; fi; \
	set +a; \
	$(GO) run ./services/casino/cmd/casino

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
	$(GO) run ./services/agent/cmd --mode planner-once

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
