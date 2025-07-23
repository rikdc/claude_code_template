# Prompt Manager

A Go web service for managing and analyzing Claude Code conversation prompts and responses.

## Development

```bash
# Run the server
go run cmd/main.go

# Build the application
go build -o bin/prompt-manager cmd/main.go

# Run tests
go test ./...
```

## API Endpoints

- `GET /health` - Health check
- `GET /api/v1/conversations` - List conversations (TODO)
- `POST /api/v1/conversations/{id}/rating` - Rate conversation (TODO)

## Database

Uses SQLite for development, with migration support for production PostgreSQL.