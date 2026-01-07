# Dockerfile for Obsidian PA
# Combines Obsidian desktop with Go bot and Claude CLI

FROM lscr.io/linuxserver/obsidian:latest

# Install build dependencies (Debian-based image)
RUN apt-get update && apt-get install -y --no-install-recommends \
    golang-go \
    nodejs \
    npm \
    ca-certificates \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Claude Code CLI globally
RUN npm install -g @anthropic-ai/claude-code

# Build the Go bot
WORKDIR /app
COPY go.mod go.sum ./
COPY vendor/ ./vendor/
COPY main.go ./
COPY CLAUDE.md ./

# Build with vendored dependencies
RUN go build -mod=vendor -o bot .

# Copy S6 service files for auto-start
COPY root/ /

# Reset workdir for runtime
WORKDIR /config
