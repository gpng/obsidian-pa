// Package main provides Telegram bot functionality.
package main

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gpng/obsidian-pa/src/executor"
)

// TelegramConfig holds Telegram-specific configuration
type TelegramConfig struct {
	Token         string
	AllowedUserID int64
}

// runTelegramBot starts the Telegram bot and listens for messages
func runTelegramBot(tgConfig *TelegramConfig, exec executor.Executor) {
	// Initialize Telegram bot
	bot, err := tgbotapi.NewBotAPI(tgConfig.Token)
	if err != nil {
		log.Fatalf("Failed to create Telegram bot: %v", err)
	}

	log.Printf("[Telegram] Authorized on account %s (using %s)", bot.Self.UserName, exec.Name())

	// Set up updates channel
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Println("[Telegram] Bot is running and listening for messages...")

	// Session ID for conversation continuity (Telegram-specific)
	var sessionID string

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Authenticate user
		if update.Message.From.ID != tgConfig.AllowedUserID {
			log.Printf("[Telegram] Unauthorized access attempt from user ID: %d", update.Message.From.ID)
			continue
		}

		userMsg := update.Message.Text
		chatID := update.Message.Chat.ID

		if userMsg == "" {
			continue
		}

		log.Printf("[Telegram] Received message from authorized user: %s", userMsg)

		// Handle /reset command
		if userMsg == "/reset" {
			sessionID = ""
			log.Println("[Telegram] Session reset")
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

		// Handle /start command - Read context and start daily review
		if userMsg == "/start" {
			// Reset session for a fresh start
			sessionID = ""
			log.Printf("[Telegram] Starting new session with %s context", exec.Name())

			// Send processing indicator
			processingMsg := tgbotapi.NewMessage(chatID, "🌅 Starting your day... Reading context and reviewing tasks...")
			sentMsg, err := bot.Send(processingMsg)
			if err != nil {
				log.Printf("[Telegram] Failed to send processing message: %v", err)
			}

			// Execute AI CLI with the start prompt
			response, newSessionID := exec.Execute(exec.GetStartPrompt(), sessionID)

			// Update session ID
			if newSessionID != "" {
				sessionID = newSessionID
				log.Printf("[Telegram] New session started: %s", sessionID)
			}

			// Delete processing message
			if err == nil {
				deleteMsg := tgbotapi.NewDeleteMessage(chatID, sentMsg.MessageID)
				bot.Request(deleteMsg)
			}

			// Send response
			sendTelegramResponse(bot, chatID, response)
			continue
		}

		// Send processing indicator
		processingMsg := tgbotapi.NewMessage(chatID, "🧠 Processing...")
		sentMsg, err := bot.Send(processingMsg)
		if err != nil {
			log.Printf("[Telegram] Failed to send processing message: %v", err)
		}

		// Execute AI CLI
		response, newSessionID := exec.Execute(userMsg, sessionID)

		// Update session ID if we got a new one
		if newSessionID != "" {
			sessionID = newSessionID
			log.Printf("[Telegram] Session ID: %s", sessionID)
		}

		// Delete processing message
		if err == nil {
			deleteMsg := tgbotapi.NewDeleteMessage(chatID, sentMsg.MessageID)
			bot.Request(deleteMsg)
		}

		// Send response (split if too long for Telegram's 4096 char limit)
		sendTelegramResponse(bot, chatID, response)
	}
}

// sendTelegramResponse sends a message to Telegram, splitting it if necessary
func sendTelegramResponse(bot *tgbotapi.BotAPI, chatID int64, response string) {
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
			log.Printf("[Telegram] Failed to send with Markdown, retrying as plain text: %v", err)
			msg.ParseMode = ""
			if _, err := bot.Send(msg); err != nil {
				log.Printf("[Telegram] Failed to send response: %v", err)
				errMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Failed to send response: %v", err))
				bot.Send(errMsg)
				return
			}
		}
	}
}
