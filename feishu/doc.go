// Package feishu provides a client for sending notifications via Feishu (Lark).
//
// It supports two modes of operation:
//
//   - Webhook mode: Send text and rich-text (post) messages through a Feishu
//     bot webhook URL. This requires only a webhook URL and no app credentials.
//
//   - SDK mode: Send messages and trigger phone urgent calls through the
//     official Lark Open Platform SDK. This requires an App ID, App Secret,
//     and optionally a User Open ID for urgent calls.
//
// # Quick Start
//
// For simple webhook-only usage with the global convenience API:
//
//	feishu.Init(feishu.Config{
//	    Webhook: "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
//	})
//
//	if err := feishu.SendText("Hello from Go!"); err != nil {
//	    log.Fatal(err)
//	}
//
// For multi-instance or SDK-based usage:
//
//	client := feishu.NewClient("webhook-url", "app-id", "app-secret")
//	if err := client.SendWebhook(feishu.TextReq{...}); err != nil {
//	    log.Fatal(err)
//	}
package feishu
