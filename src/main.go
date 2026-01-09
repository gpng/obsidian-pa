// Package main implements a multi-platform bot that bridges user messages to Claude CLI
// for managing an Obsidian vault. Supports Telegram and Slack (Socket Mode).
package main

import (
	"log"
	"os"
	"strconv"
)

// DefaultVaultPath is the default path to the Obsidian vault in the container
const DefaultVaultPath = "/config/Obsidian Vault"

// DefaultClaudeModel is the default Claude model to use
const DefaultClaudeModel = "claude-haiku-4-5"

func main() {
	// Load shared configuration
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	vaultPath := os.Getenv("VAULT_PATH")
	if vaultPath == "" {
		vaultPath = DefaultVaultPath
	}
	claudeModel := os.Getenv("CLAUDE_MODEL")
	if claudeModel == "" {
		claudeModel = DefaultClaudeModel
	}

	if anthropicKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	config := &Config{
		AnthropicKey: anthropicKey,
		VaultPath:    vaultPath,
		ClaudeModel:  claudeModel,
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
		go runTelegramBot(telegramConfig, config)
	}

	if slackEnabled {
		log.Println("Starting Slack bot (Socket Mode)...")
		go runSlackBot(slackConfig, config)
	}

	// Block forever
	select {}
}

