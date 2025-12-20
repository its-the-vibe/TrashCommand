# GitHub Copilot Instructions for TrashCommand

## Project Overview
TrashCommand is a Go service that listens to Slack reaction events via Redis and automatically deletes messages when they receive a wastebasket (🗑️) emoji reaction from a non-bot user.

## Tech Stack
- **Language**: Go 1.23+
- **Dependencies**: Redis client (go-redis/v9), Slack client (slack-go/slack)
- **Deployment**: Docker with multi-stage builds (scratch-based for minimal image size)
- **Configuration**: Environment variables

## Development Guidelines

### Code Style
- Follow standard Go conventions and idiomatic Go patterns
- Use `gofmt` for code formatting
- Keep functions small and focused on a single responsibility
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Prefer composition over inheritance

### Building and Testing
- Build: `go build -o trashcommand .`
- Build with vendor: `go build -mod=vendor -o trashcommand .`
- Test: `go test ./...` (if tests exist)
- Docker build: `docker build -t trashcommand:latest .`
- The project uses vendored dependencies for reproducible builds

### Project Structure
- `main.go`: Single-file application containing all core logic
- `vendor/`: Vendored dependencies (managed by `go mod vendor`)
- `Dockerfile`: Multi-stage build using scratch base image
- `docker-compose.yml`: Local development with Redis
- `.env.example`: Template for environment configuration

### Key Components
1. **Event Handling**: Listens to Redis pub/sub channel for Slack reaction events
2. **Message Deletion**: Uses Slack API to delete messages when wastebasket reaction is detected
3. **Bot Filtering**: Ensures only non-bot users can trigger message deletion
4. **Configuration**: All settings via environment variables (SLACK_BOT_TOKEN, REDIS_ADDR, etc.)

### Security Best Practices
- Never log or expose secrets (SLACK_BOT_TOKEN)
- Always validate event payloads before processing
- Keep dependencies up-to-date and free from vulnerabilities
- Use minimal Docker base images (currently using scratch)
- Validate that users are not bots before performing actions

### Error Handling
- Log errors with context using `log.Printf`
- Continue processing on non-fatal errors (e.g., malformed payloads)
- Exit on fatal errors (e.g., missing SLACK_BOT_TOKEN, Redis connection failure)
- Use graceful shutdown on interrupt signals

### Dependencies
- Use vendored dependencies (`-mod=vendor` flag)
- Run `go mod tidy` and `go mod vendor` when adding new dependencies
- Keep go.mod minimal with only necessary direct dependencies

### Documentation
- Keep README.md up-to-date with configuration changes
- Document environment variables in both README.md and code comments
- Include example payloads and event structures in documentation
- Maintain clear deployment instructions for both local and Docker environments

### Docker Best Practices
- Use multi-stage builds to minimize final image size
- Build static binaries with CGO_ENABLED=0
- Use scratch base image for security and size benefits
- Include CA certificates for HTTPS API calls
- Keep image size minimal (current target: ~7MB)

### When Making Changes
- Ensure backward compatibility with existing Redis event payloads
- Update README.md if adding new environment variables or features
- Test locally with docker-compose before containerized deployment
- Verify that graceful shutdown still works properly
- Consider edge cases in event handling (malformed JSON, missing fields, etc.)
