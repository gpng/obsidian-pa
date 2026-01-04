// Package main implements a Telegram bot that bridges user messages to Claude CLI
// for managing an Obsidian vault.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// DefaultVaultPath is the default path to the Obsidian vault in the container
const DefaultVaultPath = "/config/Obsidian Vault"

// DefaultClaudeModel is the default Claude model to use
const DefaultClaudeModel = "claude-haiku-4-5"

// ClaudeResponse represents the JSON response from Claude CLI
type ClaudeResponse struct {
	SessionID string `json:"session_id"`
	Result    string `json:"result"`
}

func main() {
	// Load required environment variables
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	allowedUserIDStr := os.Getenv("ALLOWED_TELEGRAM_USER_ID")
	vaultPath := os.Getenv("VAULT_PATH")
	if vaultPath == "" {
		vaultPath = DefaultVaultPath
	}
	claudeModel := os.Getenv("CLAUDE_MODEL")
	if claudeModel == "" {
		claudeModel = DefaultClaudeModel
	}

	if telegramToken == "" {
		log.Fatal("TELEGRAM_TOKEN environment variable is required")
	}
	if anthropicKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}
	if allowedUserIDStr == "" {
		log.Fatal("ALLOWED_TELEGRAM_USER_ID environment variable is required")
	}

	allowedUserID, err := strconv.ParseInt(allowedUserIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid ALLOWED_TELEGRAM_USER_ID: %v", err)
	}

	// Initialize Telegram bot
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Fatalf("Failed to create Telegram bot: %v", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Set up updates channel
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Println("Bot is running and listening for messages...")

	// Session ID for conversation continuity
	var sessionID string

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Authenticate user
		if update.Message.From.ID != allowedUserID {
			log.Printf("Unauthorized access attempt from user ID: %d", update.Message.From.ID)
			continue
		}

		userMsg := update.Message.Text
		chatID := update.Message.Chat.ID

		if userMsg == "" {
			continue
		}

		log.Printf("Received message from authorized user: %s", userMsg)

		// Handle /reset command
		if userMsg == "/reset" {
			sessionID = ""
			log.Println("Session reset")
			msg := tgbotapi.NewMessage(chatID, "🔄 Session reset. Starting fresh conversation.")
			bot.Send(msg)
			continue
		}

		// Handle /status command
		if userMsg == "/status" {
			var statusMsg string
			if sessionID != "" {
				statusMsg = fmt.Sprintf("✅ Active session: %s", sessionID)
			} else {
				statusMsg = "ℹ️ No active session. Next message will start a new one."
			}
			msg := tgbotapi.NewMessage(chatID, statusMsg)
			bot.Send(msg)
			continue
		}

		// Handle /start command - Read CLAUDE.md and start daily review
		if userMsg == "/start" {
			// Reset session for a fresh start
			sessionID = ""
			log.Println("Starting new session with CLAUDE.md context")

			// Send processing indicator
			processingMsg := tgbotapi.NewMessage(chatID, "🌅 Starting your day... Reading context and reviewing tasks...")
			sentMsg, err := bot.Send(processingMsg)
			if err != nil {
				log.Printf("Failed to send processing message: %v", err)
			}

			// Execute Claude CLI with the start prompt
			startPrompt := `Read the CLAUDE.md file to understand your role and the project context.

Then, help me start my day by:
1. Reviewing any tasks or to-do items in the vault
2. Checking for recent notes or updates
3. Suggesting what I should focus on today

Provide a concise daily briefing.`

			response, newSessionID := executeClaude(startPrompt, anthropicKey, vaultPath, claudeModel, sessionID)

			// Update session ID
			if newSessionID != "" {
				sessionID = newSessionID
				log.Printf("New session started: %s", sessionID)
			}

			// Delete processing message
			if err == nil {
				deleteMsg := tgbotapi.NewDeleteMessage(chatID, sentMsg.MessageID)
				bot.Request(deleteMsg)
			}

			// Send response
			sendResponse(bot, chatID, response)
			continue
		}

		// Send processing indicator
		processingMsg := tgbotapi.NewMessage(chatID, "🧠 Processing...")
		sentMsg, err := bot.Send(processingMsg)
		if err != nil {
			log.Printf("Failed to send processing message: %v", err)
		}

		// Execute Claude CLI
		response, newSessionID := executeClaude(userMsg, anthropicKey, vaultPath, claudeModel, sessionID)

		// Update session ID if we got a new one
		if newSessionID != "" {
			sessionID = newSessionID
			log.Printf("Session ID: %s", sessionID)
		}

		// Delete processing message
		if err == nil {
			deleteMsg := tgbotapi.NewDeleteMessage(chatID, sentMsg.MessageID)
			bot.Request(deleteMsg)
		}

		// Send response (split if too long for Telegram's 4096 char limit)
		sendResponse(bot, chatID, response)
	}
}

// executeClaude runs the Claude CLI with the given prompt and returns the output and session ID
func executeClaude(prompt, apiKey, vaultPath, model, sessionID string) (string, string) {
	args := []string{
		"-p", prompt,
		"--dangerously-skip-permissions",
		"--add-dir", vaultPath,
		"--output-format", "json",
		"--model", model,
	}

	// Resume session if we have one
	if sessionID != "" {
		args = append(args, "--resume", sessionID)
	}

	cmd := exec.Command("claude", args...)
	cmd.Dir = vaultPath
	cmd.Env = append(os.Environ(), "ANTHROPIC_API_KEY="+apiKey)

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

// sendResponse sends a message to Telegram, splitting it if necessary
func sendResponse(bot *tgbotapi.BotAPI, chatID int64, response string) {
	const maxLength = 4096

	// Split response if too long
	for len(response) > 0 {
		chunk := response
		if len(chunk) > maxLength {
			// Try to split at a newline
			splitIdx := strings.LastIndex(response[:maxLength], "\n")
			if splitIdx == -1 || splitIdx < maxLength/2 {
				splitIdx = maxLength
			}
			chunk = response[:splitIdx]
			response = response[splitIdx:]
		} else {
			response = ""
		}

		msg := tgbotapi.NewMessage(chatID, chunk)
		msg.ParseMode = "Markdown" // Enable Markdown rendering

		_, err := bot.Send(msg)
		if err != nil {
			// If Markdown parsing fails, try again without it
			log.Printf("Failed to send with Markdown, retrying as plain text: %v", err)
			msg.ParseMode = ""
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Failed to send response: %v", err)
				errMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Failed to send response: %v", err))
				bot.Send(errMsg)
				return
			}
		}
	}
}

