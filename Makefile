.PHONY: migrate-up migrate-drop install generate start test generate-api generate-repo

migrate-up:
	DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable go tool dbmate migrate up
migrate-drop:
	DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable go tool dbmate migrate drop
install:
	go mod download
generate-api:
	go tool oapi-codegen -config ./oapi-codegen-cfg.yaml ./api/api.yaml
generate-repo:
	go tool sqlc generate
generate: generate-api generate-repo
start:
	go run cmd/*
test:
	go test -v ./...
