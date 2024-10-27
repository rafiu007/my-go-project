#!/bin/bash
# setup.sh

# Create main project directory structure
mkdir -p \
    cmd/api \
    internal/domain/{entity,repository} \
    internal/infrastructure/{db,queue} \
    internal/interfaces/http/{handlers,middleware,router} \
    internal/interfaces/queue \
    internal/service \
    internal/scheduler \
    pkg/{logger,validator} \
    config \
    migrations

# Create placeholder files to ensure git tracks empty directories
touch \
    cmd/api/main.go \
    internal/domain/entity/calendar_entry.go \
    internal/domain/repository/calendar_repository.go \
    internal/infrastructure/db/mysql.go \
    internal/infrastructure/queue/sqs.go \
    internal/interfaces/http/handlers/calendar_handler.go \
    internal/interfaces/http/middleware/logging.go \
    internal/interfaces/http/router/router.go \
    internal/interfaces/queue/consumer.go \
    internal/service/calendar_service.go \
    internal/scheduler/queue_scheduler.go \
    pkg/logger/logger.go \
    pkg/validator/validator.go \
    config/config.go \
    migrations/001_create_calendar_entries.sql \
    .gitignore \
    README.md \
    Makefile \
    docker-compose.yml

# Initialize go module (replace 'your-module-name' with your actual module name)
go mod init my-go-project

# Add basic .gitignore
cat > .gitignore << 'EOF'
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with 'go test -c'
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (remove the comment below to include it)
vendor/

# Environment files
.env

# IDE specific files
.idea/
.vscode/
*.swp
*.swo

# OS specific files
.DS_Store
EOF

# Add basic README
cat > README.md << 'EOF'
# Calendar Entry Service

A service for managing calendar entries with MySQL storage and SQS integration.

## Project Structure

```
.
├── cmd/
│   └── api/              # Application entrypoints
├── internal/             # Private application and library code
│   ├── domain/          # Enterprise business rules
│   ├── infrastructure/  # Database and external services
│   ├── interfaces/      # Interface adapters
│   ├── service/         # Application business rules
│   └── scheduler/       # Scheduling services
├── pkg/                 # Public library code
├── config/             # Configuration
└── migrations/         # Database migrations
```

## Getting Started

1. Copy `.env.example` to `.env` and fill in the values
2. Run `docker-compose up -d` to start required services
3. Run `make migrate` to run database migrations
4. Run `make run` to start the service

## Development

- `make test` - Run tests
- `make lint` - Run linters
- `make build` - Build the application
EOF

# Add basic Makefile
cat > Makefile << 'EOF'
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
EOF

# Make the script executable
chmod +x setup.sh

echo "Project structure created successfully!"