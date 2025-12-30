# System Architecture

## Overview

Obsidian PA is a headless personal assistant that manages your Obsidian vault through Telegram messages. It uses Claude AI via the Claude Code CLI to perform intelligent operations on your markdown files.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              User Devices                                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   iPhone    │  │   Android   │  │   Laptop    │  │   Telegram App      │ │
│  │  Obsidian   │  │  Obsidian   │  │  Obsidian   │  │   (Control)         │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘ │
└─────────┼────────────────┼────────────────┼────────────────────┼────────────┘
          │                │                │                    │
          └────────────────┼────────────────┘                    │
                           │                                     │
                  ┌────────▼────────┐                   ┌────────▼────────┐
                  │  Obsidian Sync  │                   │   Telegram API  │
                  │   (Cloud)       │                   │   (Cloud)       │
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
│  │  │   (KasmVNC)      │    │   (Bridge)       │    │   (Brain)        │ │  │
│  │  │                  │    │                  │    │                  │ │  │
│  │  │  - Sync Client   │    │  - Listen TG     │    │  - Read files    │ │  │
│  │  │  - Plugin Host   │    │  - Auth user     │    │  - Write files   │ │  │
│  │  │  - Desktop UI    │    │  - Exec Claude   │    │  - Run tasks     │ │  │
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
│  └── telegram-bot (longrun service)                                         │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Components

### 1. Telegram Bot (Go)

**File:** `main.go`

The bridge between Telegram and Claude. Responsibilities:

- Connects to Telegram Bot API using long polling
- Authenticates incoming messages (single user only)
- Forwards user messages to Claude CLI
- Returns Claude's responses to Telegram
- Handles errors by sending them to the chat
- Splits long messages to fit Telegram's 4096 char limit

### 2. Claude CLI

**Package:** `@anthropic-ai/claude-code`

The AI brain. Capabilities:

- Reads and writes files in the vault
- Executes with `--dangerously-skip-permissions` for autonomous operation
- Uses `CLAUDE.md` for agent configuration
- Has full access to the vault directory

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
1. User sends message via Telegram
   └─▶ Telegram API
       └─▶ Go Bot receives update
           └─▶ Validates user ID
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

- Single authorized user via `ALLOWED_TELEGRAM_USER_ID`
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
| `/app` | Go bot binary and CLAUDE.md |

## Ports

| Port | Service | Purpose |
|------|---------|---------|
| 3000 | KasmVNC | Web desktop for Obsidian setup |

> **Note:** Port 3000 should be secured or closed after initial setup.

## Environment Variables

| Variable | Used By | Purpose |
|----------|---------|---------|
| `TELEGRAM_TOKEN` | Go Bot | Telegram API authentication |
| `ANTHROPIC_API_KEY` | Claude CLI | Anthropic API authentication |
| `ALLOWED_TELEGRAM_USER_ID` | Go Bot | User authorization |
| `VAULT_PATH` | Go Bot | Custom vault path (default: `/config/Obsidian Vault`) |
| `PUID`, `PGID` | LinuxServer | File permissions |
| `TZ` | Container | Timezone |

## CLAUDE.md Configuration

Claude CLI automatically looks for a `CLAUDE.md` file in the working directory and parent directories. Place your configuration file in one of these locations:

1. **In the vault** (syncs with Obsidian Sync): `obsidian_data/<VaultName>/CLAUDE.md`
2. **In config directory** (parent of vault): `obsidian_data/CLAUDE.md`

The `CLAUDE.md` file customizes Claude's behavior, including vault structure, naming conventions, and task syntax preferences.
