package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Common sentinel errors.
var (
	ErrNotInitialized = errors.New("telegram: bot not initialized")
	ErrNoChatID       = errors.New("telegram: no default notify chat ID configured")
)

// Config holds Telegram bot settings.
type Config struct {
	BotToken string // Telegram bot token
	ChatID   string // Comma-separated chat IDs (first is default for Notify)
}

var (
	globalBot *Bot
	mu        sync.Mutex
)

// Bot wraps the Telegram bot client with auth and notification support.
type Bot struct {
	client         *bot.Bot
	allowedChatIDs []int64
	notifyChatID   int64 // default chat ID for Notify()
}

// NewBot creates a new Bot from the given config.
// Returns the Bot and any error encountered during creation.
func NewBot(cfg Config) (*Bot, error) {
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("telegram: bot token is empty")
	}

	b := &Bot{}
	parseChatIDs(b, cfg.ChatID)

	opts := []bot.Option{
		bot.WithMiddlewares(AuthMiddleware(b.allowedChatIDs)),
		bot.WithDefaultHandler(defaultHandler),
	}

	var err error
	b.client, err = bot.New(cfg.BotToken, opts...)
	if err != nil {
		return nil, fmt.Errorf("telegram: create bot: %w", err)
	}

	return b, nil
}

// parseChatIDs parses a comma-separated list of chat IDs and populates the Bot.
// The first valid ID becomes the default notification target.
func parseChatIDs(b *Bot, chatIDStr string) {
	ids := strings.Split(chatIDStr, ",")
	for _, idStr := range ids {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		id, parseErr := strconv.ParseInt(idStr, 10, 64)
		if parseErr != nil {
			log.Printf("[telegram] invalid chat ID: %s, error: %v", idStr, parseErr)
			continue
		}
		b.allowedChatIDs = append(b.allowedChatIDs, id)
	}

	if len(b.allowedChatIDs) > 0 {
		b.notifyChatID = b.allowedChatIDs[0]
	}
}

// Init initializes the global bot instance.
// It is safe to call multiple times; subsequent calls replace the global instance.
func Init(cfg Config) error {
	mu.Lock()
	defer mu.Unlock()

	b, err := NewBot(cfg)
	if err != nil {
		return err
	}
	globalBot = b
	return nil
}

// GetBot returns the global bot instance.
func GetBot() *Bot {
	mu.Lock()
	defer mu.Unlock()
	return globalBot
}

// Start starts the bot long-polling. This call blocks until ctx is cancelled.
func Start(ctx context.Context) {
	mu.Lock()
	b := globalBot
	mu.Unlock()

	if b == nil || b.client == nil {
		log.Println("[telegram] bot not initialized, skipping start")
		return
	}
	log.Println("[telegram] starting bot...")
	b.client.Start(ctx)
}

// Notify sends a text message to the default chat ID (first in the list)
// using the global bot.
func Notify(msg string) error {
	mu.Lock()
	b := globalBot
	mu.Unlock()

	if b == nil {
		return ErrNotInitialized
	}
	return b.Notify(msg)
}

// Notify sends a text message to the default chat ID (first in the list).
func (b *Bot) Notify(msg string) error {
	if b.notifyChatID == 0 {
		return ErrNoChatID
	}

	_, err := b.client.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: b.notifyChatID,
		Text:   msg,
	})
	return err
}

// RegisterHandler exposes the underlying client's RegisterHandler method.
func (b *Bot) RegisterHandler(handlerType bot.HandlerType, pattern string, matchType bot.MatchType, handler bot.HandlerFunc) string {
	if b.client == nil {
		return ""
	}
	return b.client.RegisterHandler(handlerType, pattern, matchType, handler)
}

// defaultHandler is the default handler for unhandled updates.
func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// do nothing
}
