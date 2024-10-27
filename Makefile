.PHONY: run test build migrate docker-up docker-down

# Go related
run:
	go run cmd/api/main.go

test:
	go test ./... -v

build:
	go build -o bin/api cmd/api/main.go

# Database commands
migrate:
	go run cmd/api/main.go --migrate-only

# Docker related
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# Development
dev: docker-up run

# Install dependencies
deps:
	go get -u gorm.io/gorm
	go get -u gorm.io/driver/mysql
	go mod tidy