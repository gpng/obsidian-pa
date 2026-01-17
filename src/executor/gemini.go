// Package executor provides Gemini CLI execution logic.
package executor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Gemini implements the Executor interface for Gemini CLI
type Gemini struct {
	config *Config
}

// geminiStreamEvent represents an event from Gemini CLI stream-json output
type geminiStreamEvent struct {
	Type      string `json:"type"`
	SessionID string `json:"session_id,omitempty"`
	Model     string `json:"model,omitempty"`
	// For message events
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
	// For result events
	Response string `json:"response,omitempty"`
	// For error events
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// NewGemini creates a new Gemini executor
func NewGemini(config *Config) *Gemini {
	return &Gemini{config: config}
}

// Name returns the executor name for logging
func (g *Gemini) Name() string {
	return "Gemini"
}

// Execute runs the Gemini CLI with the given prompt and returns the output and session ID
func (g *Gemini) Execute(prompt string, sessionID string) (string, string) {
	args := []string{
		"-p", prompt,
		"--yolo",                                    // Auto-accept all permissions
		"--include-directories", g.config.VaultPath, // Add vault as context
		"--output-format", "stream-json",            // Streaming JSON to capture session_id
	}

	// Only pass model flag if not "auto" or empty (let Gemini CLI use its default)
	if g.config.Model != "" && g.config.Model != "auto" {
		args = append(args, "-m", g.config.Model)
	}

	// Resume session if we have one
	if sessionID != "" {
		args = append(args, "--resume", sessionID)
	}

	cmd := exec.Command("gemini", args...)
	cmd.Dir = g.config.VaultPath

	// Set API key environment variable if provided
	if g.config.APIKey != "" {
		cmd.Env = append(os.Environ(), "GEMINI_API_KEY="+g.config.APIKey)
	} else {
		cmd.Env = os.Environ() // Use existing environment (OAuth-based auth)
	}

	output, err := cmd.CombinedOutput()
	rawOutput := strings.TrimSpace(string(output))

	if err != nil {
		if rawOutput == "" {
			rawOutput = err.Error()
		}
		return fmt.Sprintf("❌ Error:\n%s", rawOutput), ""
	}

	// Parse streaming JSON (newline-delimited JSON)
	return parseStreamOutput(rawOutput)
}

// parseStreamOutput parses newline-delimited JSON events from Gemini CLI
// Returns thinking steps and final response with clear visual markers
func parseStreamOutput(output string) (string, string) {
	var sessionID string
	var thinkingSteps []string
	var finalResult string

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Skip non-JSON lines (e.g., "YOLO mode is enabled...", "Loaded cached credentials.")
		if !strings.HasPrefix(line, "{") {
			continue
		}

		var event geminiStreamEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			log.Printf("[Gemini] Failed to parse JSON event: %v", err)
			continue
		}

		switch event.Type {
		case "init":
			// Capture session ID from init event
			sessionID = event.SessionID
		case "message":
			// Collect assistant messages as thinking steps
			if event.Role == "assistant" && event.Content != "" {
				thinkingSteps = append(thinkingSteps, event.Content)
			}
		case "result":
			// Final result - this is the main response
			if event.Response != "" {
				finalResult = event.Response
			}
		case "error":
			// Handle errors
			errMsg := event.Error
			if errMsg == "" {
				errMsg = event.Message
			}
			if errMsg != "" {
				return fmt.Sprintf("❌ Error: %s", errMsg), sessionID
			}
		}
	}

	// Build the response with clear sections
	var result strings.Builder

	// Add thinking section if there are thinking steps
	if len(thinkingSteps) > 0 {
		result.WriteString("🧠 *Thinking:*\n")
		for _, step := range thinkingSteps {
			result.WriteString("• ")
			result.WriteString(step)
			result.WriteString("\n")
		}
		result.WriteString("\n---\n\n")
	}

	// Add final response
	if finalResult != "" {
		result.WriteString("📋 *Response:*\n")
		result.WriteString(finalResult)
	} else if len(thinkingSteps) > 0 {
		// If no final result but we have thinking, use last thinking step as response
		result.WriteString("📋 *Response:*\n")
		result.WriteString(thinkingSteps[len(thinkingSteps)-1])
	} else {
		result.WriteString("✅ Done (no output)")
	}

	return result.String(), sessionID
}

// GetStartPrompt returns the prompt used for the /start command
func (g *Gemini) GetStartPrompt() string {
	return StartPrompt
}

