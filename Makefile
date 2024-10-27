.PHONY: run test build migrate lint docker-up docker-down

# Go related
run:
	go run cmd/api/main.go

test:
	go test ./... -v

build:
	go build -o bin/api cmd/api/main.go

lint:
	go vet ./...
	golangci-lint run

# Docker related
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# Database related
migrate:
	mysql -h localhost -u calendar_user -pcalendar_pass calendar_db < migrations/001_create_calendar_entries.sql

# Clean and setup
clean:
	rm -rf bin/
	go clean -testcache

# Install dependencies
deps:
	go mod tidy
	go mod verify

# Development helpers
dev: docker-up migrate run