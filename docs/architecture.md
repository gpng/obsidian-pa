# System Architecture

## Overview

Obsidian PA is a headless personal assistant that manages your Obsidian vault through messaging platforms (Telegram and/or Slack). It uses AI CLI tools (Claude Code or Gemini CLI) to perform intelligent operations on your markdown files.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              User Devices                                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   iPhone    │  │   Android   │  │   Laptop    │  │ Telegram / Slack    │ │
│  │  Obsidian   │  │  Obsidian   │  │  Obsidian   │  │   (Control)         │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘ │
└─────────┼────────────────┼────────────────┼────────────────────┼────────────┘
          │                │                │                    │
          └────────────────┼────────────────┘                    │
                           │                                     │
                  ┌────────▼────────┐                   ┌────────▼────────┐
                  │  Obsidian Sync  │                   │ Telegram / Slack│
                  │   (Cloud)       │                   │      API        │
                  └────────┬────────┘                   └────────┬────────┘
                           │                                     │
                           │                                     │
┌──────────────────────────┼─────────────────────────────────────┼────────────┐
│ Docker Container         │                                     │            │
│                          │                                     │            │
│  ┌───────────────────────▼─────────────────────────────────────▼─────────┐  │
│  │                        VPS / Server                                    │  │
│  │                                                                        │  │
│  │  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐ │  │
│  │  │   Obsidian App   │    │     Go Bot       │    │   Claude CLI     │ │  │
│  │  │   (KasmVNC)      │    │   (Bridge)       │    │  (Claude/Gemini) │ │  │
│  │  │                  │    │                  │    │                  │ │  │
│  │  │  - Sync Client   │    │  - Listen TG     │    │  - Read files    │ │  │
│  │  │  - Plugin Host   │    │  - Listen Slack  │    │  - Write files   │ │  │
│  │  │  - Desktop UI    │    │  - Auth user     │    │  - Run tasks     │ │  │
│  │  └────────┬─────────┘    └────────┬─────────┘    └────────┬─────────┘ │  │
│  │           │                       │                       │           │  │
│  │           └───────────────────────┼───────────────────────┘           │  │
│  │                                   │                                    │  │
│  │                          ┌────────▼────────┐                          │  │
│  │                          │  Obsidian Vault │                          │  │
│  │                          │  /config/...    │                          │  │
│  │                          │                 │                          │  │
│  │                          │  - Markdown     │                          │  │
│  │                          │  - Attachments  │                          │  │
│  │                          │  - .obsidian/   │                          │  │
│  │                          └─────────────────┘                          │  │
│  │                                                                        │  │
│  └────────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  S6 Overlay (Process Manager)                                               │
│  ├── init-services (container startup)                                      │
│  └── telegram-bot (longrun service - handles both Telegram & Slack)         │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Components

### 1. Messaging Bot (Go)

**Files:**
- `src/main.go` - Entry point, creates executor
- `src/telegram.go` - Telegram bot handler
- `src/slack.go` - Slack bot handler
- `src/executor/` - AI executor package
  - `executor.go` - Interface definition
  - `claude.go` - Claude CLI implementation
  - `gemini.go` - Gemini CLI implementation

The bridge between messaging platforms and AI CLI. Supports Telegram and Slack (Socket Mode).

**Responsibilities:**
- Connects to Telegram Bot API (long polling) and/or Slack (Socket Mode)
- Authenticates incoming messages (single user per platform)
- Forwards user messages to Claude CLI
- Returns Claude's responses to the messaging platform
- Handles errors by sending them to the chat
- Splits long messages (4096 chars for Telegram, 4000 for Slack readability)
- Maintains separate conversation sessions per platform

### 2. AI CLI (Claude or Gemini)

The AI brain. Supports two backends:

**Claude CLI** (`@anthropic-ai/claude-code`):
- Reads and writes files in the vault
- Executes with `--dangerously-skip-permissions` for autonomous operation
- Uses `AGENT.md` for agent configuration
- Has full access to the vault directory

**Gemini CLI** (`@google/gemini-cli`):
- Reads and writes files in the vault
- Executes with `--yolo` for autonomous operation
- Uses `--include-directories` for vault context
- Supports OAuth or API key authentication

### 3. Obsidian App

**Image:** `lscr.io/linuxserver/obsidian`

The official Obsidian app running in a headless Linux desktop:

- Provides Obsidian Sync functionality
- Hosts plugins (if needed)
- Accessible via KasmVNC web interface for setup
- Runs in the background after initial setup

### 4. S6 Overlay

**Process Manager**

Manages service lifecycle in the container:

- `telegram-bot` service: longrun process for the Go bot
- Automatically restarts the bot if it crashes
- Waits for container initialization before starting

## Data Flow

### Sending a Command

```
1. User sends message via Telegram or Slack DM
   └─▶ Platform API (Telegram / Slack Socket Mode)
       └─▶ Go Bot receives update
           └─▶ Validates user ID (per platform)
               └─▶ Sends "Processing..." indicator
                   └─▶ Executes Claude CLI with message
                       └─▶ Claude reads/modifies vault
                           └─▶ Response sent back to user
```

### Syncing Changes

```
1. Claude modifies files in /config/Obsidian Vault
   └─▶ Obsidian app detects changes
       └─▶ Obsidian Sync uploads to cloud
           └─▶ Other devices receive sync
               └─▶ Changes appear instantly
```

## Security Model

### Authentication

- Single authorized user per platform:
  - Telegram: `ALLOWED_TELEGRAM_USER_ID` (integer)
  - Slack: `ALLOWED_SLACK_USER_ID` (string, e.g., `U0123456789`)
- Unauthorized messages are silently dropped
- User ID is verified for every message

### Container Isolation

- All operations run inside Docker container
- Volume mount isolates vault data
- Network access limited to required APIs

### Sensitive Data

- API keys passed via environment variables
- `.env` file excluded from git
- No secrets stored in code or images

## Volumes

| Path | Purpose |
|------|---------|
| `/config` | Obsidian vault and app settings (persistent) |
| `/app` | Go bot binary and AGENT.md |

## Ports

| Port | Service | Purpose |
|------|---------|---------|
| 3000 | KasmVNC | Web desktop for Obsidian setup |

> **Note:** Port 3000 should be secured or closed after initial setup.

## Environment Variables

### Shared

| Variable | Used By | Purpose |
|----------|---------|----------|
| `AI_EXECUTOR` | Go Bot | Which AI to use: `claude` or `gemini` (default: `claude`) |
| `VAULT_PATH` | Go Bot | Custom vault path (default: `/config/Obsidian Vault`) |
| `PUID`, `PGID` | LinuxServer | File permissions |
| `TZ` | Container | Timezone |

### Claude (default)

| Variable | Used By | Purpose |
|----------|---------|----------|
| `ANTHROPIC_API_KEY` | Claude CLI | Anthropic API authentication (required for Claude) |
| `CLAUDE_MODEL` | Go Bot | Claude model to use (default: `claude-haiku-4-5`) |

### Gemini

| Variable | Used By | Purpose |
|----------|---------|----------|
| `GEMINI_API_KEY` | Gemini CLI | Google API key (optional, can use OAuth) |
| `GEMINI_MODEL` | Go Bot | Gemini model to use (default: `auto` - Gemini chooses best model) |

### Telegram (optional)

| Variable | Used By | Purpose |
|----------|---------|----------|
| `TELEGRAM_TOKEN` | Go Bot | Telegram Bot API token from @BotFather |
| `ALLOWED_TELEGRAM_USER_ID` | Go Bot | Authorized Telegram user ID (integer) |

### Slack (optional)

| Variable | Used By | Purpose |
|----------|---------|----------|
| `SLACK_APP_TOKEN` | Go Bot | Slack App-Level token for Socket Mode (`xapp-...`) |
| `SLACK_BOT_TOKEN` | Go Bot | Slack Bot OAuth token (`xoxb-...`) |
| `ALLOWED_SLACK_USER_ID` | Go Bot | Authorized Slack user ID (e.g., `U0123456789`) |

> **Note:** At least one platform (Telegram or Slack) must be configured. Both can be enabled simultaneously.

## AGENT.md Configuration

Both Claude CLI and Gemini CLI can use an `AGENT.md` file in the working directory for context. Place your configuration file in one of these locations:

1. **In the vault** (syncs with Obsidian Sync): `obsidian_data/<VaultName>/AGENT.md`
2. **In config directory** (parent of vault): `obsidian_data/AGENT.md`

The `AGENT.md` file customizes the AI's behavior, including vault structure, naming conventions, and task syntax preferences.
