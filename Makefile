DSN ?= postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable
MIGRATIONS_DIR ?= migrations

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
	@for f in $(shell ls $(MIGRATIONS_DIR)/*.sql | sort); do \
		echo "==> applying $$f"; \
		psql "$(DSN)" -X -v ON_ERROR_STOP=1 -f "$$f" || exit $$?; \
	done

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


selftest:
	go test ./cmd/vkimport ./cmd/tgimport
