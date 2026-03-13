package telegram

import (
	"testing"
)

func TestNewBot(t *testing.T) {
	t.Run("empty token returns error", func(t *testing.T) {
		_, err := NewBot(Config{})
		if err == nil {
			t.Error("expected error for empty bot token")
		}
	})

	t.Run("invalid token returns error", func(t *testing.T) {
		// go-telegram/bot validates token format
		_, err := NewBot(Config{BotToken: "invalid"})
		if err == nil {
			t.Error("expected error for invalid token format")
		}
	})
}

func TestParseChatIDs(t *testing.T) {
	tests := []struct {
		name     string
		chatID   string
		wantLen  int
		wantFirst int64
	}{
		{"single", "12345", 1, 12345},
		{"multiple", "111,222,333", 3, 111},
		{"with spaces", " 111 , 222 ", 2, 111},
		{"empty", "", 0, 0},
		{"mixed valid invalid", "111,abc,222", 2, 111},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't fully create a Bot without a valid token,
			// so we test the parsing logic via Init behavior.
			// Instead, reuse the same logic inline to verify.
			b := &Bot{}
			parseChatIDs(b, tt.chatID)

			if len(b.allowedChatIDs) != tt.wantLen {
				t.Errorf("expected %d chat IDs, got %d", tt.wantLen, len(b.allowedChatIDs))
			}
			if tt.wantLen > 0 && b.notifyChatID != tt.wantFirst {
				t.Errorf("expected first chat ID=%d, got %d", tt.wantFirst, b.notifyChatID)
			}
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	// Test that AuthMiddleware can be constructed without panicking
	mw := AuthMiddleware([]int64{123, 456})
	if mw == nil {
		t.Error("expected non-nil middleware")
	}

	mw = AuthMiddleware(nil)
	if mw == nil {
		t.Error("expected non-nil middleware even with nil allowed IDs")
	}
}

func TestNotifyNotInitialized(t *testing.T) {
	mu.Lock()
	old := globalBot
	globalBot = nil
	mu.Unlock()

	defer func() {
		mu.Lock()
		globalBot = old
		mu.Unlock()
	}()

	err := Notify("test")
	if err != ErrNotInitialized {
		t.Errorf("expected ErrNotInitialized, got %v", err)
	}
}
