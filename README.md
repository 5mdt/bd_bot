# Birthday Bot

A simple HTMX-powered web application for managing birthday records with Telegram bot notifications.

## ⚠️ Security Warning

**This application has no authentication or authorization mechanisms.**

- **Anyone with access to the web interface can view, edit, or delete all birthday data**
- **Do not expose this service to the public internet without proper security measures**
- **Recommended for local networks, private deployments, or behind authentication proxies only**

### Secure Deployment Options

1. **Local Network Only**: Deploy on localhost or private network
2. **VPN Access**: Place behind a VPN for remote access
3. **Reverse Proxy Auth**: Use nginx, Traefik, or similar with authentication
4. **Firewall Rules**: Restrict access to specific IP addresses
5. **Container Networks**: Use Docker networks to limit exposure

## Features

- 🎂 Web interface for managing birthday records
- 🤖 Telegram bot notifications for birthday reminders
- 📱 Mobile-responsive design
- 🔄 Real-time updates with HTMX
- 🐳 Multi-architecture Docker support (amd64, arm64, i386)

## Quick Start

### Using Docker

```bash
# Pull and run the latest image
docker run -d \
  -p 8080:8080 \
  -v ./data:/data \
  -e TELEGRAM_BOT_TOKEN=your_bot_token \
  ghcr.io/5mdt/bd_bot:latest
```

### Using Docker Compose

```bash
# Clone the repository
git clone https://github.com/5mdt/bd_bot
cd bd_bot

# Set environment variables
echo "TELEGRAM_BOT_TOKEN=your_bot_token" > .env

# Start the application
docker-compose up -d
```

### From Source

```bash
# Install Go 1.21+
# Clone the repository
git clone https://github.com/5mdt/bd_bot
cd bd_bot

# Build and run
go build -o bd_bot ./cmd/app
./bd_bot
```

## Configuration

### Environment Variables

- `PORT`: Server port (default: 8080)
- `YAML_PATH`: Path to birthday data file (default: `/data/birthdays.yaml`)
- `TELEGRAM_BOT_TOKEN`: Telegram bot token
- `NOTIFICATION_START_HOUR`: Start hour for notifications in UTC (default: 8)
- `NOTIFICATION_END_HOUR`: End hour for notifications in UTC (default: 20)

### Logging

- `DEBUG`: Set to `true` for verbose logging
- `LOG_LEVEL`: Set to `DEBUG`, `INFO`, `WARN`, or `ERROR`

## Telegram Bot Setup

1. Create a bot with [@BotFather](https://t.me/botfather)
2. Get the bot token
3. Set the `TELEGRAM_BOT_TOKEN` environment variable
4. Add the bot to your Telegram chats
5. Use `/update_birth_date YYYY-MM-DD` to set birthdays

## License

This project is open source. See the [LICENSE](./LICENSE) for details.

## Security Notice

**This software is provided "as is" without any warranty. Users are responsible for securing their deployment according to their security requirements.**
