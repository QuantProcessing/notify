package feishu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// Common sentinel errors.
var (
	ErrNotInitialized = errors.New("feishu: client not initialized")
	ErrNoUserOpenID   = errors.New("feishu: user_open_id not configured, cannot send urgent message")
)

// Config holds Feishu notification settings.
type Config struct {
	Webhook    string // Webhook URL for bot messages
	AppID      string // Lark App ID (optional, for SDK features)
	AppSecret  string // Lark App Secret (optional, for SDK features)
	UserOpenID string // User Open ID (optional, for urgent calls)
}

// Bot holds the Feishu client and user configuration.
// Use NewBot for multi-instance usage or Init for the global convenience API.
type Bot struct {
	client     *Client
	userOpenID string
}

var (
	globalBot *Bot
	mu        sync.Mutex
)

// NewBot creates a new Bot instance from the given Config.
// Returns nil if neither Webhook nor AppID is configured.
func NewBot(cfg Config) *Bot {
	if cfg.Webhook == "" && cfg.AppID == "" {
		return nil
	}
	return &Bot{
		client:     NewClient(cfg.Webhook, cfg.AppID, cfg.AppSecret),
		userOpenID: cfg.UserOpenID,
	}
}

// Init initializes the global Feishu bot. It is safe to call multiple times;
// subsequent calls replace the global instance.
func Init(cfg Config) {
	mu.Lock()
	defer mu.Unlock()

	b := NewBot(cfg)
	if b == nil {
		log.Println("[feishu] not configured (no webhook or app_id), skipping init")
		return
	}
	globalBot = b
	log.Println("[feishu] bot initialized")
}

// SendText sends a simple text message via webhook using the global bot.
func SendText(text string) error {
	mu.Lock()
	b := globalBot
	mu.Unlock()

	if b == nil {
		return ErrNotInitialized
	}
	return b.SendText(text)
}

// SendRichText sends a post message with title and content lines via webhook
// using the global bot.
func SendRichText(title string, content [][]PostElem) error {
	mu.Lock()
	b := globalBot
	mu.Unlock()

	if b == nil {
		return ErrNotInitialized
	}
	return b.SendRichText(title, content)
}

// SendUrgentText sends a text message and triggers a phone call notification
// using the global bot. Falls back to SendText if SDK is not configured.
func SendUrgentText(text string) error {
	mu.Lock()
	b := globalBot
	mu.Unlock()

	if b == nil {
		return ErrNotInitialized
	}
	return b.SendUrgentText(text)
}

// SendText sends a simple text message via webhook.
func (b *Bot) SendText(text string) error {
	req := TextReq{
		BaseReq: BaseReq{MsgType: MsgTypeText},
		Content: TextContent{Text: text},
	}
	return b.client.SendWebhook(req)
}

// SendRichText sends a post message with title and content lines via webhook.
func (b *Bot) SendRichText(title string, content [][]PostElem) error {
	req := PostReq{
		BaseReq: BaseReq{MsgType: MsgTypePost},
		Content: PostContentWrapper{
			Post: PostBody{
				ZhCN: &PostContent{
					Title:   title,
					Content: content,
				},
			},
		},
	}
	return b.client.SendWebhook(req)
}

// SendUrgentText sends a text message to the configured user via SDK,
// then triggers a phone call notification on that message.
// Falls back to SendText if SDK is not configured.
func (b *Bot) SendUrgentText(text string) error {
	// Fallback: if SDK not available, just send via webhook
	if b.client.larkClient == nil {
		log.Println("[feishu] lark SDK not configured, falling back to webhook SendText for urgent message")
		return b.SendText(text)
	}

	if b.userOpenID == "" {
		return ErrNoUserOpenID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Build message content JSON
	contentJSON, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return fmt.Errorf("feishu: marshal message content: %w", err)
	}

	// Step 1: Send message via SDK to get message_id
	messageID, err := b.client.SendMessage(ctx, "open_id", b.userOpenID, "text", string(contentJSON))
	if err != nil {
		return fmt.Errorf("feishu: send message for urgent: %w", err)
	}

	log.Printf("[feishu] message sent, triggering phone urgent, message_id: %s", messageID)

	// Step 2: Trigger phone urgent on the message
	if err := b.client.UrgentPhone(ctx, messageID, []string{b.userOpenID}); err != nil {
		return fmt.Errorf("feishu: urgent phone call failed (message was sent, id=%s): %w", messageID, err)
	}

	log.Println("[feishu] phone urgent triggered successfully")
	return nil
}
