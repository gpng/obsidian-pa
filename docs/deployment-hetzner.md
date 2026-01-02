# Deployment Guide - Hetzner

This guide covers deploying Obsidian PA to a Hetzner Cloud Server.

## Prerequisites

- Hetzner Cloud account
- SSH key added to Hetzner
- Telegram bot token (from @BotFather)
- Anthropic API key
- Your Telegram user ID (from @userinfobot)
- Obsidian Sync subscription

## 1. Create Server

1. Go to [Hetzner Cloud Console](https://console.hetzner.cloud/)
2. Create a new project (or select existing)
3. Click **Add Server**
4. Choose:
   - **Location**: Closest to you (e.g., Nuremberg, Helsinki)
   - **Image**: Ubuntu 24.04
   - **Type**: CX22 (2 vCPU, 4GB RAM, €4.35/mo) or CAX11 (ARM, €3.79/mo)
   - **Networking**: ✅ **Enable IPv4** (required for SSH and Telegram API)
   - **SSH Keys**: Select your SSH key
5. Click **Create & Buy Now**
6. Note the IPv4 address

## 2. Initial Server Setup

SSH into your server:

```bash
ssh root@YOUR_SERVER_IP
```

### Install Docker

```bash
# Update system
apt update && apt upgrade -y

# Install dependencies
apt install -y make git

# Install Docker
curl -fsSL https://get.docker.com | sh

# Verify installation
docker --version
docker compose version
```

### Create non-root user (recommended)

```bash
# Create user
adduser obsidian
usermod -aG docker obsidian

# Switch to new user
su - obsidian
```

## 3. Deploy Application

### Clone repository

```bash
cd ~
git clone https://github.com/YOUR_USERNAME/obsidian-pa.git
cd obsidian-pa
```

### Configure environment variables

```bash
cp .env.example .env
nano .env
```

Fill in your credentials:

```bash
TELEGRAM_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz
ANTHROPIC_API_KEY=sk-ant-api03-xxxxx
ALLOWED_TELEGRAM_USER_ID=123456789
```

Save and exit (Ctrl+X, Y, Enter).

### Set file permissions

```bash
chmod 600 .env
```

### Build and start

```bash
make up
```

Or without Make:

```bash
docker compose up -d --build
```

Check status:

```bash
make logs
# or
docker compose ps
docker compose logs -f
```

> **Note**: If Obsidian hangs when opening, the container includes `shm_size: 1g` for Electron apps. If issues persist, increase memory in `docker-compose.yml`.

## 4. Initial Obsidian Sync Setup

This is a **one-time manual step** to authenticate with Obsidian Sync.

### SSH Tunnel (Recommended)

From your **local machine**:

```bash
ssh -L 3000:localhost:3000 obsidian@YOUR_SERVER_IP
```

Then open `http://localhost:3000` in your browser.

### Web Desktop Login

1. Login credentials: `abc` / `abc` (default for LinuxServer images)
2. You'll see a Linux desktop with Obsidian
3. Open Obsidian (should auto-start or find in menu)

### Connect Obsidian Sync

1. In Obsidian, click the **Settings** gear icon (bottom left)
2. Go to **Sync** in the sidebar
3. Click **Log in**
4. Enter your Obsidian account credentials
5. After login, click **Choose** to select your remote vault
6. Select the vault you want to sync
7. Click **Connect**
8. Wait for initial sync to complete (watch the sync icon)

### Verify Sync

- Check that your files appear in the vault
- Create a test note and verify it appears on your phone/laptop

### Close Web Desktop

After successful sync setup, just close the browser tab.

## 5. Firewall Configuration

Hetzner has a cloud firewall. Configure it in the console:

1. Go to **Firewalls** → **Create Firewall**
2. Add rules:
   - **SSH**: TCP port 22, source 0.0.0.0/0
3. Apply to your server

Or use UFW on the server:

```bash
ufw enable
ufw allow 22/tcp
ufw status
```

## 6. Test the Bot

1. Open Telegram
2. Find your bot (search by username)
3. Send a message:
   ```
   List all files in my vault
   ```
4. You should see:
   - "🧠 Processing..." briefly
   - Claude's response with your vault contents

## 7. Monitoring & Maintenance

### View logs

```bash
make logs
# or
docker compose logs -f
```

### Restart services

```bash
make restart
```

### Update to latest version

```bash
git pull
make up
```

### Check disk usage

```bash
df -h
du -sh ~/obsidian-pa/obsidian_data
```

## 8. Backup

The vault data is stored in `~/obsidian-pa/obsidian_data/`. Back this up periodically:

```bash
# Create backup
tar -czf obsidian-backup-$(date +%Y%m%d).tar.gz obsidian_data/

# Copy to local machine (run from local)
scp obsidian@YOUR_SERVER_IP:~/obsidian-pa/obsidian-backup-*.tar.gz ./
```

## Troubleshooting

### Bot not responding

```bash
# Check if container is running
docker compose ps

# Check logs for errors
make logs

# Verify environment variables are set
docker compose exec obsidian-brain env | grep -E "TELEGRAM|ANTHROPIC|ALLOWED"
```

### Obsidian not syncing

1. SSH tunnel to port 3000
2. Open web desktop
3. Check Obsidian Sync status in settings
4. Look for sync errors

### Out of memory

```bash
# Check memory usage
free -h
docker stats

# Increase swap if needed
fallocate -l 2G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
echo '/swapfile none swap sw 0 0' >> /etc/fstab
```

## Cost Estimate

| Resource | Monthly Cost |
|----------|--------------|
| Hetzner CX22 (4GB RAM) | €4.35 |
| IPv4 address | €0.50 |
| Obsidian Sync | $4 |
| Anthropic API | Variable |
| **Total** | ~€5 + $4 + API usage |
