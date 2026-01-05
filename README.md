# TrashCommand
Listen for the trashcan emoji reaction and delete the associated slack message, or listen for the bomb emoji and schedule deletion via TimeBomb

## Overview
TrashCommand is a Go service that listens to Slack reaction events via Redis and:
- Automatically deletes messages when they receive a wastebasket (🗑️) emoji reaction from a non-bot user
- Publishes messages to TimeBomb for scheduled deletion when they receive a bomb (💣) emoji reaction from a non-bot user

## Features
- Subscribes to Redis channel for Slack reaction events
- Filters for wastebasket and bomb emoji reactions from non-bot users
- Automatically deletes messages with wastebasket reactions
- Publishes bomb-reacted messages to TimeBomb for scheduled deletion
- Configurable via environment variables
- Containerized deployment with Docker

## Requirements
- Go 1.23+
- Redis server (hosted externally)
- Slack Bot Token with `chat:write` permission

## Configuration
The service is configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SLACK_BOT_TOKEN` | Slack Bot OAuth Token (required) | - |
| `REDIS_ADDR` | Redis server address | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password (if required) | - |
| `REDIS_CHANNEL` | Redis channel to subscribe to | `slack-relay-reaction-added` |
| `TIMEBOMB_REDIS_CHANNEL` | Redis channel to publish bomb reactions to | `timebomb-messages` |
| `TIMEBOMB_TTL_SECONDS` | TTL in seconds for TimeBomb messages | `5` |
| `LOG_LEVEL` | Logging level: DEBUG, INFO, or ERROR | `INFO` |

### Log Levels
- **DEBUG**: Logs all messages including skipped events (e.g., non-reaction events, bot reactions, unsupported reactions)
- **INFO**: Logs important operational messages (startup, connections, message deletions, TimeBomb publishes)
- **ERROR**: Logs only errors and fatal messages

## Running Locally

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Edit `.env` and set your configuration values

3. Build and run:
```bash
go build -o trashcommand .
./trashcommand
```

## Running with Docker

1. Build the Docker image:
```bash
docker build -t trashcommand:latest .
```

2. Run the container:
```bash
docker run -d \
  -e SLACK_BOT_TOKEN=xoxb-your-token \
  -e REDIS_ADDR=redis.example.com:6379 \
  --name trashcommand \
  trashcommand:latest
```

## Running with Docker Compose

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Edit `.env` and set your configuration values

3. Start the service:
```bash
docker-compose up -d
```

4. View logs:
```bash
docker-compose logs -f
```

## Event Payload Format
The service expects messages on the Redis channel to be JSON payloads with the following structure:
```json
{
  "event": {
    "type": "reaction_added",
    "user": "U123456789",
    "reaction": "wastebasket",
    "item": {
      "type": "message",
      "channel": "C123456789",
      "ts": "1766236581.981479"
    }
  },
  "authorizations": [
    {
      "user_id": "U123456789",
      "is_bot": false
    }
  ]
}
```

## How It Works
1. The service connects to Redis and subscribes to the configured channel
2. When a message is received, it parses the JSON payload
3. It checks if:
   - The event type is `reaction_added`
   - The reaction is `wastebasket` or `bomb`
   - The item is a `message`
   - The user is not a bot
4. For wastebasket reactions:
   - Calls the Slack API to delete the message immediately
5. For bomb reactions:
   - Publishes the message details to the TimeBomb Redis channel with the configured TTL
   - TimeBomb will handle the scheduled deletion

## Security
- All dependencies are free from known vulnerabilities
- Uses minimal scratch-based Docker image (7.07MB)
- Requires authentication via SLACK_BOT_TOKEN
- No secrets are logged or exposed

## License
MIT
