# Slack Bot Setup Guide

This guide walks you through creating a Slack bot for Obsidian PA using Socket Mode.

> **Note:** Slack is optional. You can also use [Telegram](telegram-bot-setup.md) instead, or enable both platforms simultaneously.

## Prerequisites

- A Slack workspace where you have permission to install apps
- Admin access or ability to request app installation

## Why Socket Mode?

Socket Mode uses WebSockets instead of HTTP webhooks, which means:
- **No public URL required** - Works behind firewalls
- **Easier setup** - No need to configure SSL or expose ports
- **Lower latency** - Persistent connection

## Step 1: Create a Slack App

1. Go to [api.slack.com/apps](https://api.slack.com/apps)
2. Click **Create New App**
3. Choose **From scratch**
4. Enter:
   - **App Name**: `Obsidian PA` (or your preferred name)
   - **Workspace**: Select your workspace
5. Click **Create App**

## Step 2: Enable Socket Mode

1. In the left sidebar, click **Socket Mode**
2. Toggle **Enable Socket Mode** to ON
3. You'll be prompted to create an App-Level Token:
   - **Token Name**: `socket-token` (or any name)
   - **Scopes**: Add `connections:write`
4. Click **Generate**
5. Copy the token (starts with `xapp-`) → This is your `SLACK_APP_TOKEN`

> ⚠️ **Keep this token secret!** Anyone with this token can connect to your app.

## Step 3: Add Bot Token Scopes

1. In the left sidebar, click **OAuth & Permissions**
2. Scroll to **Scopes** → **Bot Token Scopes**
3. Click **Add an OAuth Scope** and add:

| Scope | Purpose |
|-------|---------|
| `chat:write` | Send messages to users |
| `im:history` | Receive DM messages |

That's all you need - just 2 scopes!

## Step 4: Install App to Workspace

1. Scroll up to **OAuth Tokens for Your Workspace**
2. Click **Install to Workspace**
3. Review permissions and click **Allow**
4. Copy the **Bot User OAuth Token** (starts with `xoxb-`) → This is your `SLACK_BOT_TOKEN`

## Step 5: Subscribe to Events

1. In the left sidebar, click **Event Subscriptions**
2. Toggle **Enable Events** to ON
3. Expand **Subscribe to bot events**
4. Click **Add Bot User Event**
5. Select `message.im` (messages in DMs with the bot)
6. Click **Save Changes**

## Step 6: Get Your Slack User ID

The bot needs your user ID to authenticate you.

1. Open Slack
2. Click your profile picture (bottom-left corner)
3. Click **Profile**
4. Click the **⋮** (more) menu
5. Select **Copy member ID**
6. This is your `ALLOWED_SLACK_USER_ID` (e.g., `U0ABC123DEF`)

## Step 7: Configure Environment Variables

Add to your `.env` file:

```bash
# Slack Configuration
SLACK_APP_TOKEN=xapp-1-A0123456789-1234567890123-abcdef...
SLACK_BOT_TOKEN=xoxb-1234567890-1234567890123-abcDEF...
ALLOWED_SLACK_USER_ID=U0ABC123DEF
```

## Step 8: Verify Setup

1. Start (or restart) your container:
   ```bash
   make restart
   # or
   docker compose up -d --build
   ```

2. Check logs to confirm Slack connected:
   ```bash
   make logs
   ```
   You should see:
   ```
   [Slack] Connecting to Slack...
   [Slack] Connected to Slack Socket Mode
   [Slack] Bot is running and listening for messages...
   ```

3. Open Slack and send a DM to your bot:
   ```
   Hello, are you working?
   ```

4. You should see:
   - A "🧠 Processing..." message appear briefly
   - A response from Claude

## Available Commands

Send these as DMs to the bot:

| Command | Description |
|---------|-------------|
| `start` | Read AGENT.md and start daily review |
| `status` | Check if there's an active session |
| `reset` | Clear session and start fresh |

> **Note:** Unlike Telegram, Slack commands work with or without the `/` prefix.

## Troubleshooting

### "Unauthorized access attempt" in logs

Your Slack user ID doesn't match `ALLOWED_SLACK_USER_ID`. Double-check:
1. The member ID from your Slack profile
2. The value in your `.env` file

### Bot not responding

1. Check container logs:
   ```bash
   docker compose logs -f | grep Slack
   ```
2. Verify all three Slack env vars are set correctly
3. Ensure Socket Mode is enabled in the Slack app settings

### "Connection error, will retry..." in logs

This is normal during startup or if there's a network blip. Socket Mode will automatically reconnect.

### Bot responds to others

Check that you set `ALLOWED_SLACK_USER_ID` correctly. Only DMs from this user ID will be processed.

## Security Best Practices

1. **Never commit `.env`** - It's in `.gitignore` for a reason
2. **Regenerate tokens if exposed** - Delete and recreate in Slack app settings
3. **Use a dedicated workspace** - Or create the app in a workspace you control
4. **Keep user ID private** - It can be used to identify you

## Summary of Required Tokens

| Token | Prefix | Where to Get |
|-------|--------|--------------|
| App-Level Token | `xapp-` | Socket Mode → App-Level Tokens |
| Bot OAuth Token | `xoxb-` | OAuth & Permissions → Bot User OAuth Token |
| User ID | `U...` | Your Slack profile → Copy member ID |
