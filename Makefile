ifneq (,$(wildcard ./.env))
    include .env
    export
endif

ifneq (,$(wildcard ./.env.local))
    include .env.local
    export
endif

.PHONY: run migrate migrate-down migrate-status generate test

run:
	go run ./cmd/trip-service

migrate:
	go tool goose -dir ./migrations postgres $(DATABASE_URL) up

migrate-down:
	go tool goose -dir ./migrations postgres $(DATABASE_URL) down

migrate-status:
	go tool goose -dir ./migrations postgres $(DATABASE_URL) status

generate:
	go tool oapi-codegen \
		-generate types,chi-server \
		-package api \
		-o internal/generated/api.gen.go \
		contracts/openapi/trip-service.openapi.yaml

test:
	go test -race ./...