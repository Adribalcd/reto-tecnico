.PHONY: up down build logs test test-go test-node

# Levanta el stack completo (frontend, go-api y node-api).
up:
	docker compose up --build

down:
	docker compose down

build:
	docker compose build

logs:
	docker compose logs -f

test: test-go test-node

test-go:
	cd go-api && go test ./...

test-node:
	cd node-api && npm test
