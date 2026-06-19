.PHONY: dev stop build test migrate docker-up docker-down

dev:
	@bash scripts/run-dev.sh

stop:
	@bash scripts/stop-dev.sh

build:
	cd backend && go build -o ../bin/cnow-server ./cmd/server
	cd frontend && npm run build

test:
	cd backend && go test ./...
	cd frontend && npm test

migrate:
	cd backend && go run cmd/migrate/main.go

docker-up:
	docker compose up -d

docker-down:
	docker compose down
