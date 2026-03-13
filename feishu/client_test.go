package feishu

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Run("webhook only", func(t *testing.T) {
		c := NewClient("https://example.com/hook", "", "")
		if c.webhook != "https://example.com/hook" {
			t.Errorf("expected webhook to be set, got %q", c.webhook)
		}
		if c.larkClient != nil {
			t.Error("expected larkClient to be nil without app credentials")
		}
	})

	t.Run("with app credentials", func(t *testing.T) {
		c := NewClient("https://example.com/hook", "app-id", "app-secret")
		if c.larkClient == nil {
			t.Error("expected larkClient to be initialized with app credentials")
		}
	})
}

func TestNewBot(t *testing.T) {
	t.Run("nil when unconfigured", func(t *testing.T) {
		b := NewBot(Config{})
		if b != nil {
			t.Error("expected nil Bot when neither webhook nor app_id is set")
		}
	})

	t.Run("created with webhook", func(t *testing.T) {
		b := NewBot(Config{Webhook: "https://example.com/hook"})
		if b == nil {
			t.Fatal("expected non-nil Bot")
		}
		if b.client == nil {
			t.Error("expected client to be initialized")
		}
	})

	t.Run("stores user open id", func(t *testing.T) {
		b := NewBot(Config{
			Webhook:    "https://example.com/hook",
			UserOpenID: "ou_test123",
		})
		if b == nil {
			t.Fatal("expected non-nil Bot")
		}
		if b.userOpenID != "ou_test123" {
			t.Errorf("expected userOpenID=%q, got %q", "ou_test123", b.userOpenID)
		}
	})
}

func TestInit(t *testing.T) {
	// Init should allow multiple calls (unlike sync.Once)
	Init(Config{Webhook: "https://example.com/hook1"})
	Init(Config{Webhook: "https://example.com/hook2"})

	mu.Lock()
	b := globalBot
	mu.Unlock()

	if b == nil {
		t.Fatal("expected global bot to be initialized")
	}
	// Second Init should have replaced the first
	if b.client.webhook != "https://example.com/hook2" {
		t.Errorf("expected webhook from second Init, got %q", b.client.webhook)
	}
}

func TestTextSerialization(t *testing.T) {
	req := TextReq{
		BaseReq: BaseReq{MsgType: MsgTypeText},
		Content: TextContent{Text: "hello"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if m["msg_type"] != "text" {
		t.Errorf("expected msg_type=text, got %v", m["msg_type"])
	}
	content := m["content"].(map[string]interface{})
	if content["text"] != "hello" {
		t.Errorf("expected text=hello, got %v", content["text"])
	}
}

func TestPostSerialization(t *testing.T) {
	req := PostReq{
		BaseReq: BaseReq{MsgType: MsgTypePost},
		Content: PostContentWrapper{
			Post: PostBody{
				ZhCN: &PostContent{
					Title: "Test",
					Content: [][]PostElem{
						{NewTextElem("Hello "), NewAElem("World", "https://example.com")},
						{NewAtElem("ou_user1")},
					},
				},
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if m["msg_type"] != "post" {
		t.Errorf("expected msg_type=post, got %v", m["msg_type"])
	}
}

func TestPostElemConstructors(t *testing.T) {
	text := NewTextElem("hello")
	if text.Tag != "text" || text.Text != "hello" {
		t.Errorf("NewTextElem: got tag=%q text=%q", text.Tag, text.Text)
	}

	a := NewAElem("click", "https://example.com")
	if a.Tag != "a" || a.Text != "click" || a.Href != "https://example.com" {
		t.Errorf("NewAElem: got tag=%q text=%q href=%q", a.Tag, a.Text, a.Href)
	}

	at := NewAtElem("ou_user1")
	if at.Tag != "at" || at.UserId != "ou_user1" {
		t.Errorf("NewAtElem: got tag=%q userId=%q", at.Tag, at.UserId)
	}
}

func TestDoPost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"msg":  "ok",
		})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "", "")
	if err := c.SendWebhook(map[string]string{"msg_type": "text"}); err != nil {
		t.Errorf("expected success, got %v", err)
	}
}

func TestDoPostError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "", "")
	err := c.SendWebhook(map[string]string{"msg_type": "text"})
	if err == nil {
		t.Error("expected error for 500 status")
	}
}

func TestSendTextNotInitialized(t *testing.T) {
	// Reset global state
	mu.Lock()
	old := globalBot
	globalBot = nil
	mu.Unlock()

	defer func() {
		mu.Lock()
		globalBot = old
		mu.Unlock()
	}()

	err := SendText("test")
	if err != ErrNotInitialized {
		t.Errorf("expected ErrNotInitialized, got %v", err)
	}
}

// --- Integration tests below (require environment variables) ---

func TestIntegrationSendRichText(t *testing.T) {
	webhook := os.Getenv("FEISHU_WEBHOOK")
	if webhook == "" {
		t.Skip("FEISHU_WEBHOOK not set, skipping integration test")
	}

	Init(Config{Webhook: webhook})

	if err := SendRichText("Test Title", [][]PostElem{
		{NewTextElem("Hello")},
		{NewAElem("Link", "http://example.com")},
	}); err != nil {
		t.Fatalf("SendRichText failed: %v", err)
	}
}

func TestIntegrationSendUrgentText(t *testing.T) {
	webhook := os.Getenv("FEISHU_WEBHOOK")
	appID := os.Getenv("FEISHU_APP_ID")
	appSecret := os.Getenv("FEISHU_APP_SECRET")
	userOpenID := os.Getenv("FEISHU_USER_OPEN_ID")

	if webhook == "" || appID == "" {
		t.Skip("FEISHU_WEBHOOK/FEISHU_APP_ID not set, skipping integration test")
	}

	Init(Config{
		Webhook:    webhook,
		AppID:      appID,
		AppSecret:  appSecret,
		UserOpenID: userOpenID,
	})

	if err := SendUrgentText("🚨 Test urgent message"); err != nil {
		t.Fatalf("SendUrgentText failed: %v", err)
	}
}
