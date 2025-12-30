# Deployment Guide - DigitalOcean

This guide covers deploying Obsidian PA to a DigitalOcean Droplet.

## Prerequisites

- DigitalOcean account
- SSH key added to DigitalOcean
- Telegram bot token (from @BotFather)
- Anthropic API key
- Your Telegram user ID (from @userinfobot)
- Obsidian Sync subscription

## 1. Create Droplet

1. Go to [DigitalOcean Control Panel](https://cloud.digitalocean.com/)
2. Create → Droplets
3. Choose:
   - **Region**: Closest to you
   - **Image**: Ubuntu 24.04 LTS
   - **Size**: Basic → Regular → $12/mo (2GB RAM, 1 vCPU)
   - **Authentication**: SSH Key
4. Click "Create Droplet"
5. Note the IP address

## 2. Initial Server Setup

SSH into your droplet:

```bash
ssh root@YOUR_DROPLET_IP
```

### Install Docker

```bash
# Update system
apt update && apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh

# Install Docker Compose plugin
apt install docker-compose-plugin -y

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

### Option A: SSH Tunnel (Recommended - More Secure)

From your **local machine**:

```bash
ssh -L 3000:localhost:3000 obsidian@YOUR_DROPLET_IP
```

Then open `http://localhost:3000` in your browser.

### Option B: Temporary Firewall Opening

```bash
# On the droplet
ufw allow 3000/tcp
```

Open `http://YOUR_DROPLET_IP:3000` in your browser.

> ⚠️ Remember to close this port after setup!

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

After successful sync setup:

1. Close the browser tab
2. If you opened port 3000 in firewall, close it:
   ```bash
   ufw delete allow 3000/tcp
   ```

## 5. Firewall Configuration

```bash
# Enable UFW
ufw enable

# Allow SSH
ufw allow 22/tcp

# Allow port 3000 only if you need web access
# ufw allow 3000/tcp

# Check status
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
docker compose logs -f
```

### Restart services

```bash
docker compose restart
```

### Update to latest version

```bash
git pull
docker compose up -d --build
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
scp obsidian@YOUR_DROPLET_IP:~/obsidian-pa/obsidian-backup-*.tar.gz ./
```

## Troubleshooting

### Bot not responding

```bash
# Check if container is running
docker compose ps

# Check logs for errors
docker compose logs telegram-bot

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
| DigitalOcean Droplet (2GB) | $12 |
| Obsidian Sync | $4 |
| Anthropic API | Variable (pay per use) |
| **Total** | ~$16 + API usage |
