// Package executor provides a common interface for AI CLI executors.
package executor

// Executor defines the interface for AI CLI executors (Claude, Gemini, etc.)
type Executor interface {
	// Execute runs the AI CLI with the given prompt and optional session ID.
	// Returns the response text and the session ID (for session continuity).
	Execute(prompt string, sessionID string) (response string, newSessionID string)

	// GetStartPrompt returns the prompt used for the /start command.
	GetStartPrompt() string

	// Name returns the name of the executor (for logging).
	Name() string
}

// Config holds common configuration for all executors
type Config struct {
	APIKey    string // API key for the AI service
	VaultPath string // Path to the Obsidian vault
	Model     string // Model to use
}

// StartPrompt is the shared prompt used for the /start command across all executors
const StartPrompt = `Read the AGENT.md file to understand your role and the project context.

Then, help me start my day by:
1. Reviewing any tasks or to-do items in the vault
2. Checking for recent notes or updates
3. Suggesting what I should focus on today

Provide a concise daily briefing.`
