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
