// Package telegram provides a Telegram bot client for sending notifications
// and handling commands with built-in chat ID authorization.
//
// It wraps the [github.com/go-telegram/bot] library and adds:
//
//   - Chat ID whitelist middleware for access control
//   - A simple Notify function for sending messages to a default chat
//   - Global convenience API via Init/Start/Notify
//   - Multi-instance support via the Bot struct
//
// # Quick Start
//
// Send a one-off notification:
//
//	if err := telegram.Init(telegram.Config{
//	    BotToken: "123456:ABC-DEF...",
//	    ChatID:   "12345678",
//	}); err != nil {
//	    log.Fatal(err)
//	}
//
//	if err := telegram.Notify("Hello from Go!"); err != nil {
//	    log.Fatal(err)
//	}
//
// Start the bot for receiving commands (blocking):
//
//	telegram.Start(ctx)
package telegram
