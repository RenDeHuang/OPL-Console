.PHONY: test build migrate-up dev-api dev-web

test:
	go test ./...
	npm --prefix apps/web run typecheck
	npm --prefix apps/web run build

build:
	go build ./cmd/console-api
	npm --prefix apps/web run build

migrate-up:
	go run ./cmd/migrate up

dev-api:
	go run ./cmd/console-api

dev-web:
	npm --prefix apps/web run dev -- --host 127.0.0.1
