// Package main provides shared Claude CLI execution logic.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// ClaudeResponse represents the JSON response from Claude CLI
type ClaudeResponse struct {
	SessionID string `json:"session_id"`
	Result    string `json:"result"`
}

// Config holds shared configuration for the application
type Config struct {
	AnthropicKey string
	VaultPath    string
	ClaudeModel  string
}

// executeClaude runs the Claude CLI with the given prompt and returns the output and session ID
func executeClaude(prompt string, config *Config, sessionID string) (string, string) {
	args := []string{
		"-p", prompt,
		"--dangerously-skip-permissions",
		"--add-dir", config.VaultPath,
		"--output-format", "json",
		"--model", config.ClaudeModel,
	}

	// Resume session if we have one
	if sessionID != "" {
		args = append(args, "--resume", sessionID)
	}

	cmd := exec.Command("claude", args...)
	cmd.Dir = config.VaultPath
	cmd.Env = append(os.Environ(), "ANTHROPIC_API_KEY="+config.AnthropicKey)

	output, err := cmd.CombinedOutput()
	rawOutput := strings.TrimSpace(string(output))

	if err != nil {
		if rawOutput == "" {
			rawOutput = err.Error()
		}
		return fmt.Sprintf("❌ Error:\n%s", rawOutput), ""
	}

	// Try to parse JSON response
	var claudeResp ClaudeResponse
	if err := json.Unmarshal([]byte(rawOutput), &claudeResp); err != nil {
		// If JSON parsing fails, return raw output (might be plain text on error)
		log.Printf("Failed to parse Claude JSON response: %v", err)
		if rawOutput == "" {
			return "✅ Done (no output)", ""
		}
		return rawOutput, ""
	}

	// Extract result and session ID
	result := claudeResp.Result
	if result == "" {
		result = "✅ Done (no output)"
	}

	return result, claudeResp.SessionID
}

// getStartPrompt returns the prompt used for the /start command
func getStartPrompt() string {
	return `Read the CLAUDE.md file to understand your role and the project context.

Then, help me start my day by:
1. Reviewing any tasks or to-do items in the vault
2. Checking for recent notes or updates
3. Suggesting what I should focus on today

Provide a concise daily briefing.`
}
