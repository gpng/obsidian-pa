# Telegram Bot Setup Guide

This guide walks you through creating a Telegram bot for Obsidian PA.

> **Note:** Telegram is optional. You can also use [Slack](slack-setup.md) instead, or enable both platforms simultaneously.

## Prerequisites

- A Telegram account
- The Telegram app (mobile or desktop)

## Step 1: Create a Bot with BotFather

1. Open Telegram and search for [@BotFather](https://t.me/BotFather)
2. Start a chat with BotFather
3. Send the command:
   ```
   /newbot
   ```
4. BotFather will ask for a name. Enter something like:
   ```
   Obsidian PA
   ```
5. BotFather will ask for a username. This must be unique and end with `bot`:
   ```
   my_obsidian_pa_bot
   ```
6. BotFather will respond with your **bot token**. It looks like:
   ```
   123456789:ABCdefGHIjklMNOpqrsTUVwxyz
   ```

> ⚠️ **Keep this token secret!** Anyone with this token can control your bot.

## Step 2: Get Your Telegram User ID

Your bot needs your user ID to know who is authorized to use it.

1. Search for [@userinfobot](https://t.me/userinfobot) in Telegram
2. Start a chat with it
3. It will immediately reply with your user information:
   ```
   Id: 123456789
   First: Your Name
   Lang: en
   ```
4. Copy your **Id** (the number, e.g., `123456789`)

## Step 3: Configure Environment Variables

Create your `.env` file:

```bash
cp .env.example .env
```

Edit `.env` and fill in your values:

```bash
# The token from BotFather (Step 1)
TELEGRAM_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz

# Your Anthropic API key
ANTHROPIC_API_KEY=sk-ant-...

# Your Telegram user ID from userinfobot (Step 2)
ALLOWED_TELEGRAM_USER_ID=123456789
```

## Step 4: Configure Bot Settings (Optional)

You can customize your bot's profile in BotFather:

### Set Bot Description
```
/setdescription
```
Then select your bot and enter:
```
Personal assistant for managing your Obsidian vault. Send me tasks, notes, or questions about your vault!
```

### Set Bot About Text
```
/setabouttext
```
Then select your bot and enter:
```
Manages Obsidian vault using Claude AI
```

### Set Bot Profile Picture
```
/setuserpic
```
Then select your bot and upload an image.

### Set Bot Commands
```
/setcommands
```
Then select your bot and enter:
```
help - Show available commands
status - Check bot status
```

> Note: The bot doesn't actually have commands—it processes natural language. But setting these helps users understand what the bot does.

## Step 5: Verify Setup

1. Start your container:
   ```bash
   docker compose up -d --build
   ```

2. Open your bot in Telegram (search for the username you created)

3. Send a test message:
   ```
   Hello, are you working?
   ```

4. You should see:
   - A "🧠 Processing..." message appear briefly
   - A response from Claude

## Troubleshooting

### "Unauthorized access attempt" in logs

Your Telegram user ID doesn't match `ALLOWED_TELEGRAM_USER_ID`. Double-check:
1. The user ID from @userinfobot
2. The value in your `.env` file

### Bot not responding at all

1. Check container logs:
   ```bash
   docker compose logs -f
   ```
2. Verify `TELEGRAM_TOKEN` is correct
3. Ensure the container is running:
   ```bash
   docker compose ps
   ```

### "Missing TELEGRAM_TOKEN" error

The environment variable isn't being passed to the container. Check:
1. `.env` file exists and has the token
2. `docker-compose.yml` references `${TELEGRAM_TOKEN}`

## Security Best Practices

1. **Never commit `.env`** - It's in `.gitignore` for a reason
2. **Rotate token if exposed** - Use `/revoke` with BotFather to generate a new token
3. **Use a dedicated bot** - Don't share this bot token with other applications
4. **Keep user ID private** - Your user ID can be used to target you for spam

## BotFather Command Reference

| Command | Description |
|---------|-------------|
| `/newbot` | Create a new bot |
| `/mybots` | List your bots |
| `/setname` | Change bot's display name |
| `/setdescription` | Set the bot's description |
| `/setabouttext` | Set the "About" text |
| `/setuserpic` | Set profile picture |
| `/setcommands` | Define bot commands |
| `/deletebot` | Delete a bot |
| `/token` | Generate new token |
| `/revoke` | Revoke current token |
