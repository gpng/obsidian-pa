// Package executor provides Claude CLI execution logic.
package executor

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Claude implements the Executor interface for Claude CLI
type Claude struct {
	config *Config
}

// claudeResponse represents the JSON response from Claude CLI
type claudeResponse struct {
	SessionID string `json:"session_id"`
	Result    string `json:"result"`
}

// NewClaude creates a new Claude executor
func NewClaude(config *Config) *Claude {
	return &Claude{config: config}
}

// Name returns the executor name for logging
func (c *Claude) Name() string {
	return "Claude"
}

// Execute runs the Claude CLI with the given prompt and returns the output and session ID
func (c *Claude) Execute(prompt string, sessionID string) (string, string) {
	args := []string{
		"-p", prompt,
		"--dangerously-skip-permissions",
		"--add-dir", c.config.VaultPath,
		"--output-format", "json",
		"--model", c.config.Model,
	}

	// Resume session if we have one
	if sessionID != "" {
		args = append(args, "--resume", sessionID)
	}

	cmd := exec.Command("claude", args...)
	cmd.Dir = c.config.VaultPath
	cmd.Env = append(os.Environ(), "ANTHROPIC_API_KEY="+c.config.APIKey)

	output, err := cmd.CombinedOutput()
	rawOutput := strings.TrimSpace(string(output))

	if err != nil {
		if rawOutput == "" {
			rawOutput = err.Error()
		}
		return fmt.Sprintf("❌ Error:\n%s", rawOutput), ""
	}

	// Try to parse JSON response
	var resp claudeResponse
	if err := json.Unmarshal([]byte(rawOutput), &resp); err != nil {
		// If JSON parsing fails, return raw output (might be plain text on error)
		log.Printf("[Claude] Failed to parse JSON response: %v", err)
		if rawOutput == "" {
			return "✅ Done (no output)", ""
		}
		return rawOutput, ""
	}

	// Extract result and session ID
	result := resp.Result
	if result == "" {
		result = "✅ Done (no output)"
	}

	return result, resp.SessionID
}

// GetStartPrompt returns the prompt used for the /start command
func (c *Claude) GetStartPrompt() string {
	return StartPrompt
}

