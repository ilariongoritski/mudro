DSN ?= postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable
MIGRATION ?= migrations/001_init.sql

up:
	docker compose up -d

down:
	docker compose down

ps:
	docker compose ps

logs:
	docker compose logs --no-color --tail=200

dbcheck:
	psql "$(DSN)" -X -v ON_ERROR_STOP=1 -c "select 1;"

migrate:
	psql "$(DSN)" -X -v ON_ERROR_STOP=1 -f "$(MIGRATION)"

tables:
	psql "$(DSN)" -X -c "\dt"

test:
	go test ./...

count-posts:
	psql "$(DSN)" -X -c "select count(*) from posts;"

health:
	$(MAKE) up
	$(MAKE) ps
	$(MAKE) dbcheck
	$(MAKE) migrate
	$(MAKE) tables
	$(MAKE) test
	$(MAKE) count-posts
