.PHONY: run test build migrate lint

run:
	go run cmd/api/main.go

test:
	go test ./...

build:
	go build -o bin/api cmd/api/main.go

migrate:
	go run migrations/*.go

lint:
	go vet ./...
	golangci-lint run
