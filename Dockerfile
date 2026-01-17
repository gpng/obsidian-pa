# Dockerfile for Obsidian PA
# Combines Obsidian desktop with Go bot and Claude CLI

FROM lscr.io/linuxserver/obsidian:latest

# Install build dependencies (Debian-based image)
RUN apt-get update && apt-get install -y --no-install-recommends \
    golang-go \
    nodejs \
    ca-certificates \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Claude Code CLI and Gemini CLI globally
RUN npm install -g @anthropic-ai/claude-code @google/gemini-cli

# Build the Go bot
WORKDIR /app
COPY go.mod go.sum ./
COPY vendor/ ./vendor/
COPY src/ ./src/

# Build with vendored dependencies
RUN go build -mod=vendor -o bot ./src

# Copy S6 service files for auto-start
COPY root/ /

# Reset workdir for runtime
WORKDIR /config
