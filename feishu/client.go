package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// Client wraps both webhook and official Lark SDK capabilities.
type Client struct {
	webhook    string
	http       *http.Client
	larkClient *lark.Client // nil if AppID not configured
}

// NewClient creates a Client with webhook support.
// If appID and appSecret are provided, the official Lark SDK client is also initialized.
func NewClient(webhook, appID, appSecret string) *Client {
	c := &Client{
		webhook: webhook,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	if appID != "" && appSecret != "" {
		c.larkClient = lark.NewClient(appID, appSecret)
	}
	return c
}

// SendWebhook sends a message via the webhook URL (for text/rich-text bot messages).
func (c *Client) SendWebhook(req interface{}) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.doPost(data)
}

// SendMessage sends a message via the official SDK and returns the message_id.
// receiveIDType: "open_id", "user_id", "union_id", "email", "chat_id"
func (c *Client) SendMessage(ctx context.Context, receiveIDType, receiveID, msgType, content string) (string, error) {
	if c.larkClient == nil {
		return "", fmt.Errorf("feishu: lark SDK client not initialized (missing app_id/app_secret)")
	}

	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(receiveIDType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(receiveID).
			MsgType(msgType).
			Content(content).
			Build()).
		Build()

	resp, err := c.larkClient.Im.Message.Create(ctx, req)
	if err != nil {
		return "", fmt.Errorf("feishu: send message failed: %w", err)
	}
	if !resp.Success() {
		return "", fmt.Errorf("feishu: send message error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return *resp.Data.MessageId, nil
}

// UrgentPhone triggers a phone call notification for the given message.
func (c *Client) UrgentPhone(ctx context.Context, messageID string, userIDs []string) error {
	if c.larkClient == nil {
		return fmt.Errorf("feishu: lark SDK client not initialized (missing app_id/app_secret)")
	}

	req := larkim.NewUrgentPhoneMessageReqBuilder().
		MessageId(messageID).
		UserIdType("open_id").
		UrgentReceivers(larkim.NewUrgentReceiversBuilder().
			UserIdList(userIDs).
			Build()).
		Build()

	resp, err := c.larkClient.Im.Message.UrgentPhone(ctx, req)
	if err != nil {
		return fmt.Errorf("feishu: urgent phone failed: %w", err)
	}
	if !resp.Success() {
		return fmt.Errorf("feishu: urgent phone error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

func (c *Client) doPost(data []byte) error {
	resp, err := c.http.Post(c.webhook, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Drain body so the connection can be reused.
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("feishu: api status: %s", resp.Status)
	}

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("feishu: decode response failed: %w", err)
	}

	if result.Code != 0 {
		return fmt.Errorf("feishu: api error: %d - %s", result.Code, result.Msg)
	}

	return nil
}
