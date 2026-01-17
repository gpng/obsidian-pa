// Package main provides Slack Socket Mode bot functionality.
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gpng/obsidian-pa/src/executor"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// SlackConfig holds Slack-specific configuration
type SlackConfig struct {
	AppToken      string
	BotToken      string
	AllowedUserID string
}

// runSlackBot starts the Slack bot using Socket Mode and listens for DM messages
func runSlackBot(slackConfig *SlackConfig, exec executor.Executor) {
	// Initialize Slack API client
	api := slack.New(
		slackConfig.BotToken,
		slack.OptionAppLevelToken(slackConfig.AppToken),
	)

	// Create Socket Mode client
	client := socketmode.New(api)

	// Create Socket Mode handler
	handler := socketmode.NewSocketmodeHandler(client)

	// Session ID for conversation continuity (Slack-specific)
	var sessionID string

	// Handle connection events
	handler.Handle(socketmode.EventTypeConnecting, func(evt *socketmode.Event, c *socketmode.Client) {
		log.Println("[Slack] Connecting to Slack...")
	})

	handler.Handle(socketmode.EventTypeConnected, func(evt *socketmode.Event, c *socketmode.Client) {
		log.Printf("[Slack] Connected to Slack Socket Mode (using %s)", exec.Name())
	})

	handler.Handle(socketmode.EventTypeConnectionError, func(evt *socketmode.Event, c *socketmode.Client) {
		log.Println("[Slack] Connection error, will retry...")
	})

	handler.Handle(socketmode.EventTypeHello, func(evt *socketmode.Event, c *socketmode.Client) {
		log.Println("[Slack] Received hello from Slack")
	})

	// Handle Events API events (includes messages)
	handler.Handle(socketmode.EventTypeEventsAPI, func(evt *socketmode.Event, c *socketmode.Client) {
		log.Printf("[Slack] Received Events API event: %+v", evt.Type)
		eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
		if !ok {
			return
		}

		// Always acknowledge the event
		c.Ack(*evt.Request)

		innerEvent := eventsAPIEvent.InnerEvent
		msgEvent, ok := innerEvent.Data.(*slackevents.MessageEvent)
		if !ok {
			return
		}

		// Only handle DMs (im = instant message / direct message)
		if msgEvent.ChannelType != "im" {
			return
		}

		// Ignore bot's own messages and other bots
		if msgEvent.BotID != "" {
			return
		}

		// Ignore message subtypes (edits, deletions, etc.) - only handle regular messages
		if msgEvent.SubType != "" {
			return
		}

		// Authenticate user
		if msgEvent.User != slackConfig.AllowedUserID {
			log.Printf("[Slack] Unauthorized access attempt from user ID: %s", msgEvent.User)
			return
		}

		userMsg := msgEvent.Text
		channelID := msgEvent.Channel

		if userMsg == "" {
			return
		}

		log.Printf("[Slack] Received message from authorized user: %s", userMsg)

		// Handle /reset command (also support "reset" without slash for Slack)
		if userMsg == "/reset" || userMsg == "reset" {
			sessionID = ""
			log.Println("[Slack] Session reset")
			sendSlackMessage(api, channelID, "🔄 Session reset. Starting fresh conversation.")
			return
		}

		// Handle /status command
		if userMsg == "/status" || userMsg == "status" {
			var statusMsg string
			if sessionID != "" {
				statusMsg = fmt.Sprintf("✅ Active session: %s", sessionID)
			} else {
				statusMsg = "ℹ️ No active session. Next message will start a new one."
			}
			sendSlackMessage(api, channelID, statusMsg)
			return
		}

		// Handle /start command - Read context and start daily review
		if userMsg == "/start" || userMsg == "start" {
			// Reset session for a fresh start
			sessionID = ""
			log.Printf("[Slack] Starting new session with %s context", exec.Name())

			// Send processing indicator
			processingTs := sendSlackMessage(api, channelID, "🌅 Starting your day... Reading context and reviewing tasks...")

			// Execute AI CLI with the start prompt
			response, newSessionID := exec.Execute(exec.GetStartPrompt(), sessionID)

			// Update session ID
			if newSessionID != "" {
				sessionID = newSessionID
				log.Printf("[Slack] New session started: %s", sessionID)
			}

			// Delete processing message
			if processingTs != "" {
				deleteSlackMessage(api, channelID, processingTs)
			}

			// Send response
			sendSlackResponse(api, channelID, response)
			return
		}

		// Send processing indicator
		processingTs := sendSlackMessage(api, channelID, "🧠 Processing...")

		// Execute AI CLI
		response, newSessionID := exec.Execute(userMsg, sessionID)

		// Update session ID if we got a new one
		if newSessionID != "" {
			sessionID = newSessionID
			log.Printf("[Slack] Session ID: %s", sessionID)
		}

		// Delete processing message
		if processingTs != "" {
			deleteSlackMessage(api, channelID, processingTs)
		}

		// Send response
		sendSlackResponse(api, channelID, response)
	})

	// Default handler to catch any unhandled events (for debugging)
	handler.HandleDefault(func(evt *socketmode.Event, c *socketmode.Client) {
		log.Printf("[Slack] Unhandled event type: %s", evt.Type)
	})

	log.Println("[Slack] Bot is running and listening for messages...")

	// Start the event loop (blocking)
	if err := handler.RunEventLoop(); err != nil {
		log.Fatalf("[Slack] Failed to run event loop: %v", err)
	}
}

// sendSlackMessage sends a single message to Slack and returns the timestamp (for deletion)
func sendSlackMessage(api *slack.Client, channelID, text string) string {
	_, ts, err := api.PostMessage(
		channelID,
		slack.MsgOptionText(text, false),
	)
	if err != nil {
		log.Printf("[Slack] Failed to send message: %v", err)
		return ""
	}
	return ts
}

// deleteSlackMessage deletes a message by its timestamp
func deleteSlackMessage(api *slack.Client, channelID, timestamp string) {
	_, _, err := api.DeleteMessage(channelID, timestamp)
	if err != nil {
		log.Printf("[Slack] Failed to delete message: %v", err)
	}
}

// sendSlackResponse sends a response to Slack, splitting it if necessary for readability
func sendSlackResponse(api *slack.Client, channelID, response string) {
	// Slack section blocks have a 3000 char limit for text
	const maxLength = 3000

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

		// Use blocks with mrkdwn for proper formatting
		textBlock := slack.NewTextBlockObject("mrkdwn", chunk, false, false)
		section := slack.NewSectionBlock(textBlock, nil, nil)

		_, _, err := api.PostMessage(
			channelID,
			slack.MsgOptionBlocks(section),
			slack.MsgOptionText(chunk, false), // Fallback for notifications
		)
		if err != nil {
			log.Printf("[Slack] Failed to send response: %v", err)
			// Send error message
			api.PostMessage(channelID, slack.MsgOptionText(fmt.Sprintf("❌ Failed to send response: %v", err), false))
			return
		}
	}
}
