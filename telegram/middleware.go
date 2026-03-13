package telegram

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// AuthMiddleware creates a middleware that allows only specific chat IDs.
func AuthMiddleware(allowedIds []int64) bot.Middleware {
	allowedMap := make(map[int64]bool)
	for _, id := range allowedIds {
		allowedMap[id] = true
	}

	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			var chatID int64
			var fromUser string

			// Extract Chat ID based on update type
			if update.Message != nil {
				chatID = update.Message.Chat.ID
				fromUser = update.Message.From.Username
			} else if update.CallbackQuery != nil {
				// update.CallbackQuery.Message is a value (struct), so we cannot check nil.
				// We check if the inner Message field is present.
				if update.CallbackQuery.Message.Message != nil {
					chatID = update.CallbackQuery.Message.Message.Chat.ID
				} else {
					// Fallback or ignore if InaccessibleMessage
					log.Println("[telegram] auth: inaccessible message in callback query, ignoring")
					return
				}
				fromUser = update.CallbackQuery.From.Username
			} else {
				log.Println("[telegram] auth: could not determine chat ID, ignoring update")
				return
			}

			if !allowedMap[chatID] {
				log.Printf("[telegram] unauthorized access attempt, chat_id=%d, user=%s", chatID, fromUser)
				return
			}

			next(ctx, b, update)
		}
	}
}
