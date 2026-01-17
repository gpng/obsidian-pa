# Design Decisions

This document explains the key design decisions made for Obsidian PA.

## Why Docker + LinuxServer Obsidian Image?

### Decision
Use the official LinuxServer.io Obsidian Docker image which runs the full Obsidian desktop application.

### Rationale

**Considered Alternatives:**
1. Direct file manipulation without Obsidian
2. Obsidian REST API (via plugin)
3. Custom sync implementation

**Why LinuxServer Image:**
- **Official Obsidian Sync**: The most reliable sync method. Community plugins can't fully replicate this.
- **Full Plugin Support**: Any Obsidian plugin can run (Dataview, Tasks, etc.)
- **Future Compatibility**: Updates to Obsidian work automatically via image updates
- **Visual Debugging**: KasmVNC allows you to see exactly what Obsidian sees

**Trade-offs:**
- Higher resource usage (~2GB RAM)
- Requires one-time GUI setup for Sync login
- Heavier Docker image

---

## Why Go for the Messaging Bot?

### Decision
Implement the messaging-to-Claude bridge in Go.

### Rationale

**Considered Alternatives:**
1. Python (with python-telegram-bot)
2. Node.js
3. Shell script

**Why Go:**
- **Single Binary**: Easy to deploy, no runtime dependencies
- **Low Resource Usage**: Minimal memory footprint
- **Excellent Libraries**: `telegram-bot-api` and `slack-go/slack` are mature
- **Concurrency**: Easy to run multiple platform listeners in goroutines
- **Strong Typing**: Catches errors at compile time
- **Fast Startup**: Bot is ready in milliseconds

**Trade-offs:**
- More verbose than Python
- Requires compilation step

---

## Why Claude CLI vs Anthropic API?

### Decision
Use the Claude Code CLI (`@anthropic-ai/claude-code`) instead of direct API calls.

### Rationale

**Considered Alternatives:**
1. Direct Anthropic API with tool use
2. Claude Code as library
3. LangChain or similar framework

**Why Claude CLI:**
- **Agent Capabilities**: Built-in file operations, search, and editing
- **Context Awareness**: Automatically includes relevant files
- **No Tool Implementation**: File read/write/search handled by CLI
- **AGENT.md Support**: Easy customization via markdown file
- **Battle-tested**: Same tool developers use

**Trade-offs:**
- External process execution overhead
- Less control over individual API calls
- Requires Node.js in container

---

## Why Support Multiple AI Backends?

### Decision
Support both Claude CLI and Gemini CLI through a common `Executor` interface.

### Rationale

**Why Multiple Backends:**
- **Flexibility**: Users can choose based on pricing, performance, or preference
- **Redundancy**: If one service has issues, switch to the other
- **Future-proofing**: Easy to add more backends (OpenAI, local models, etc.)
- **Cost optimization**: Gemini offers a generous free tier

**Why an Interface:**
- **Clean separation**: Bot logic doesn't care which AI it's using
- **Testability**: Easy to mock for testing
- **Single code path**: No conditional logic in bot handlers

**Implementation:**
```go
type Executor interface {
    Execute(prompt, sessionID string) (response, newSessionID string)
    GetStartPrompt() string
    Name() string
}
```

**Trade-offs:**
- Slightly more complex architecture
- Each CLI has different capabilities/quirks
- Must maintain feature parity across backends

---

## Why Single User Authentication?

### Decision
Support only one authorized Telegram user via environment variable.

### Rationale

**Considered Alternatives:**
1. Allow multiple users
2. Password-based authentication
3. No authentication (public bot)

**Why Single User:**
- **Simplicity**: No user management needed
- **Security**: Personal vault = personal access
- **Privacy**: No risk of data leakage between users
- **Performance**: No database or session management

**Trade-offs:**
- Cannot share bot with others
- Must redeploy to change user

---

## Why `--dangerously-skip-permissions`?

### Decision
Run Claude CLI with `--dangerously-skip-permissions` flag.

### Rationale

**Why Dangerous Mode:**
- **Autonomous Operation**: No human confirmation prompts
- **Telegram Context**: Can't interact with CLI prompts via chat
- **Personal Use**: You trust yourself with your own vault

**Mitigations:**
- Single authorized user only
- Operations confined to vault directory
- Obsidian Sync provides version history/restore
- Regular backups recommended

**Trade-offs:**
- Claude can delete/modify any file in vault
- No safety prompts for destructive operations

---

## Why Run Bot as Non-Root User?

### Decision
Run the Telegram bot as the `abc` user instead of root using `s6-setuidgid`.

### Rationale

**Why Required:**
- Claude CLI refuses `--dangerously-skip-permissions` when running as root
- This is a security feature of Claude CLI to prevent accidental damage

**Implementation:**
```bash
# In root/etc/s6-overlay/s6-rc.d/telegram-bot/run
#!/command/execlineb -P
with-contenv
s6-setuidgid abc
/app/bot
```

**Benefits:**
- Bot runs with minimal privileges
- Matches LinuxServer container conventions
- Vault files owned by `abc` user (PUID/PGID)

---

## Why S6 Overlay for Process Management?

### Decision
Use S6 Overlay to manage the Telegram bot as a `longrun` service.

### Rationale

**Considered Alternatives:**
1. Run bot as container entrypoint
2. supervisord
3. systemd (not available in Alpine)

**Why S6 Overlay:**
- **Built-in**: LinuxServer images already use S6
- **Auto-restart**: Bot restarts if it crashes
- **Dependencies**: Can wait for other services to start
- **Logging**: Integrated with container logging
- **Single Process**: Go bot handles both Telegram and Slack internally

**Trade-offs:**
- Learning curve for S6 configuration
- More files to maintain

---

## Why Vendored Dependencies?

### Decision
Vendor Go dependencies using `go mod vendor`.

### Rationale

**Why Vendor:**
- **Reproducible Builds**: Same deps on every build
- **No Network Required**: Build works offline
- **Audit Trail**: Dependencies are in version control
- **Build Speed**: No download step in Docker

**Trade-offs:**
- Larger repository size
- Must update vendor when deps change

---

## Why Environment Variables for Config?

### Decision
Use environment variables for all runtime configuration.

### Rationale

**Considered Alternatives:**
1. Config file (YAML/JSON)
2. Command-line flags
3. Hardcoded values

**Why Environment Variables:**
- **12-Factor App**: Standard for containerized apps
- **Docker Integration**: Easy to set in docker-compose
- **Secret Management**: Works with Docker secrets, Kubernetes secrets
- **No Config Parsing**: Built-in `os.Getenv()`

**Trade-offs:**
- Less structured than config files
- Limited to string values
