# Obsidian PA

Personal Assistant for Obsidian via Telegram, powered by Claude AI.

Obsidian PA enables you to manage your Obsidian vault through Telegram messages. Send a message to your bot, and Claude AI will read, write, and organize your notes—synced instantly to all your devices via Obsidian Sync.

## Features

- 🤖 **Natural Language Interface** - Ask Claude to create notes, manage tasks, or query your vault
- 🔄 **Real-time Sync** - Changes sync immediately via Obsidian Sync to all your devices
- 🐳 **Dockerized** - Runs the official Obsidian app in a container with full plugin support
- 🔐 **User Authentication** - Only responds to your authorized Telegram account

## Prerequisites

- Docker and Docker Compose
- Obsidian Sync subscription (for cross-device sync)
- Telegram account
- Anthropic API key

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/obsidian-pa.git
cd obsidian-pa
```

### 2. Set Up Telegram Bot

See [docs/telegram-bot-setup.md](docs/telegram-bot-setup.md) for detailed instructions.

Quick summary:
1. Message [@BotFather](https://t.me/BotFather) on Telegram
2. Send `/newbot` and follow the prompts
3. Save the bot token
4. Message [@userinfobot](https://t.me/userinfobot) to get your user ID

### 3. Configure Environment Variables

```bash
cp .env.example .env
```

Edit `.env` with your credentials:

```bash
TELEGRAM_TOKEN=your_bot_token_from_botfather
ANTHROPIC_API_KEY=your_anthropic_api_key
ALLOWED_TELEGRAM_USER_ID=your_telegram_user_id
```

### 4. Build and Start

```bash
make up
```

Or without Make:

```bash
docker compose up -d --build
```

### 5. Initial Obsidian Setup (One-Time)

1. Open `http://localhost:3000` in your browser
2. Default login: `abc` / `abc`
3. You'll see a Linux desktop with Obsidian
4. Open Obsidian and log in to Obsidian Sync
5. Select your vault to sync
6. Wait for files to sync, then close the browser

### 6. Start Using

Send a message to your Telegram bot:

> "Create a note called 'Project Ideas' with a list of things I want to build"

The bot will respond with the result, and the note will appear in your Obsidian apps!

## Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `TELEGRAM_TOKEN` | Yes | Bot token from @BotFather |
| `ANTHROPIC_API_KEY` | Yes | Anthropic API key from console.anthropic.com |
| `ALLOWED_TELEGRAM_USER_ID` | Yes | Your Telegram user ID (numbers only) |
| `VAULT_PATH` | No | Custom vault path (default: `/config/Obsidian Vault`) |
| `CLAUDE_MODEL` | No | Claude model to use (default: `claude-haiku-4-5`) |

### Customizing Claude's Behavior

Create a `CLAUDE.md` file to customize how Claude interacts with your vault. Place it in one of these locations:

1. **In your vault** (syncs via Obsidian Sync): `obsidian_data/<VaultName>/CLAUDE.md`
2. **In config directory**: `obsidian_data/CLAUDE.md`

Use it to:

- Define your vault structure
- Set naming conventions
- Configure task syntax preferences
- Add domain-specific instructions

### Makefile Commands

| Command | Description |
|---------|-------------|
| `make up` | Build and start container |
| `make down` | Stop container |
| `make restart` | Restart container |
| `make logs` | Follow all logs |
| `make logs-bot` | Follow bot logs only |
| `make shell` | Open bash in container |
| `make clean` | Remove container and image |
| `make clean-all` | Remove everything including data |

## Architecture

See [docs/architecture.md](docs/architecture.md) for detailed system architecture.

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  Telegram   │────▶│   Go Bot     │────▶│ Claude CLI  │
│    App      │◀────│   (Bridge)   │◀────│   (Brain)   │
└─────────────┘     └──────────────┘     └─────────────┘
                            │                    │
                            │              ┌─────▼─────┐
                            │              │  Obsidian │
                            │              │   Vault   │
                            │              └─────┬─────┘
                            │                    │
                    ┌───────▼────────────────────▼───────┐
                    │           Docker Container          │
                    │  (Obsidian + KasmVNC + Go Bot)     │
                    └───────────────────┬────────────────┘
                                        │
                              ┌─────────▼─────────┐
                              │   Obsidian Sync   │
                              └─────────┬─────────┘
                                        │
                              ┌─────────▼─────────┐
                              │   Your Devices    │
                              │ (Phone, Laptop)   │
                              └───────────────────┘
```

## Development

### Local Development

```bash
# Install Go (use mise for version management)
mise install

# Run tests
go test ./...

# Build locally
go build -o bot .
```

### Project Structure

```
obsidian-pa/
├── main.go              # Telegram bot implementation
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── vendor/              # Vendored dependencies
├── Dockerfile           # Container build instructions
├── docker-compose.yml   # Container orchestration
├── CLAUDE.md            # Claude agent configuration
├── .env.example         # Environment variable template
├── .mise.toml           # Go version management
├── docs/                # Documentation
│   ├── architecture.md
│   ├── design-decisions.md
│   └── telegram-bot-setup.md
└── root/                # S6 overlay service files
    └── etc/s6-overlay/...
```

## Deployment

### VPS Deployment (Hetzner)

See [docs/deployment-hetzner.md](docs/deployment-hetzner.md) for a complete step-by-step guide covering:

- Droplet creation and Docker setup
- Environment variable configuration
- SSH tunnel for secure Obsidian Sync setup
- Firewall configuration
- Monitoring and maintenance

**Quick overview:**

1. SSH into your VPS
2. Install Docker and Docker Compose
3. Clone the repository
4. Configure `.env`
5. Run `docker compose up -d --build`
6. Complete the one-time Obsidian setup via SSH tunnel

> ⚠️ **Security Note**: After initial setup, keep port 3000 closed. Use SSH tunneling for web desktop access.

## Troubleshooting

### Bot not responding

1. Check container logs: `docker compose logs -f`
2. Verify environment variables are set correctly
3. Ensure your Telegram user ID matches `ALLOWED_TELEGRAM_USER_ID`

### Sync not working

1. Access the web desktop at `http://localhost:3000`
2. Open Obsidian and verify Sync is connected
3. Check for sync errors in Obsidian's settings

## License

MIT
