up:
	docker compose up -d

down:
	docker compose down

test:
	go test ./...

logs:
	docker compose logs --no-color --tail=200
