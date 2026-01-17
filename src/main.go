// Package main implements a multi-platform bot that bridges user messages to AI CLI
// (Claude or Gemini) for managing an Obsidian vault. Supports Telegram and Slack (Socket Mode).
package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gpng/obsidian-pa/src/executor"
)

// DefaultVaultPath is the default path to the Obsidian vault in the container
const DefaultVaultPath = "/config/Obsidian Vault"

// DefaultClaudeModel is the default Claude model to use
const DefaultClaudeModel = "claude-haiku-4-5"

// DefaultGeminiModel is the default Gemini model to use
const DefaultGeminiModel = "auto"

func main() {
	// Determine which AI executor to use (default: claude)
	executorType := strings.ToLower(os.Getenv("AI_EXECUTOR"))
	if executorType == "" {
		executorType = "claude" // Default to Claude for backward compatibility
	}

	// Load vault path (shared across executors)
	vaultPath := os.Getenv("VAULT_PATH")
	if vaultPath == "" {
		vaultPath = DefaultVaultPath
	}

	// Create the appropriate executor
	var exec executor.Executor
	switch executorType {
	case "gemini":
		apiKey := os.Getenv("GEMINI_API_KEY") // Optional - can use OAuth
		model := os.Getenv("GEMINI_MODEL")
		if model == "" {
			model = DefaultGeminiModel
		}
		exec = executor.NewGemini(&executor.Config{
			APIKey:    apiKey,
			VaultPath: vaultPath,
			Model:     model,
		})
		log.Printf("Using Gemini executor with model: %s", model)

	case "claude":
		fallthrough
	default:
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			log.Fatal("ANTHROPIC_API_KEY environment variable is required for Claude executor")
		}
		model := os.Getenv("CLAUDE_MODEL")
		if model == "" {
			model = DefaultClaudeModel
		}
		exec = executor.NewClaude(&executor.Config{
			APIKey:    apiKey,
			VaultPath: vaultPath,
			Model:     model,
		})
		log.Printf("Using Claude executor with model: %s", model)
	}

	// Load Telegram configuration (optional)
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	telegramUserIDStr := os.Getenv("ALLOWED_TELEGRAM_USER_ID")
	telegramEnabled := telegramToken != "" && telegramUserIDStr != ""

	var telegramConfig *TelegramConfig
	if telegramEnabled {
		telegramUserID, err := strconv.ParseInt(telegramUserIDStr, 10, 64)
		if err != nil {
			log.Fatalf("Invalid ALLOWED_TELEGRAM_USER_ID: %v", err)
		}
		telegramConfig = &TelegramConfig{
			Token:         telegramToken,
			AllowedUserID: telegramUserID,
		}
	}

	// Load Slack configuration (optional)
	slackAppToken := os.Getenv("SLACK_APP_TOKEN")
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	slackUserID := os.Getenv("ALLOWED_SLACK_USER_ID")
	slackEnabled := slackAppToken != "" && slackBotToken != "" && slackUserID != ""

	var slackConfig *SlackConfig
	if slackEnabled {
		slackConfig = &SlackConfig{
			AppToken:      slackAppToken,
			BotToken:      slackBotToken,
			AllowedUserID: slackUserID,
		}
	}

	// Ensure at least one platform is enabled
	if !telegramEnabled && !slackEnabled {
		log.Fatal("At least one platform must be configured. Set TELEGRAM_TOKEN + ALLOWED_TELEGRAM_USER_ID for Telegram, or SLACK_APP_TOKEN + SLACK_BOT_TOKEN + ALLOWED_SLACK_USER_ID for Slack.")
	}

	// Start enabled platforms
	if telegramEnabled {
		log.Println("Starting Telegram bot...")
		go runTelegramBot(telegramConfig, exec)
	}

	if slackEnabled {
		log.Println("Starting Slack bot (Socket Mode)...")
		go runSlackBot(slackConfig, exec)
	}

	// Block forever
	select {}
}
