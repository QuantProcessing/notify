---
name: notify
description: Use when sending notifications via Feishu (Lark) or Telegram in Go projects, integrating webhook messages, rich-text posts, urgent phone calls, or Telegram bot commands
---

# Notify

## Overview

Lightweight Go library for sending notifications via **Feishu (Lark)** and **Telegram**. Both packages use a **Dual-API pattern** — global convenience functions for simple use, plus `NewBot()` constructors for multi-instance scenarios.

Module: `github.com/QuantProcessing/notify`

## Quick Reference

### feishu package

| Function | Description |
|----------|-------------|
| `Init(cfg Config)` | Initialize global Feishu bot |
| `SendText(text)` | Send plain text via webhook (global) |
| `SendRichText(title, content)` | Send rich-text post via webhook (global) |
| `SendUrgentText(text)` | Send text + trigger phone call (global, needs AppID/AppSecret/UserOpenID) |
| `NewBot(cfg Config) *Bot` | Create independent Bot instance |

**Config:** `Webhook` (required), `AppID`, `AppSecret`, `UserOpenID` (optional, for SDK/urgent calls)

### telegram package

| Function | Description |
|----------|-------------|
| `Init(cfg Config) error` | Initialize global Telegram bot |
| `Notify(msg) error` | Send message to default chat (global) |
| `Start(ctx)` | Start long-polling, blocks until ctx cancelled |
| `NewBot(cfg Config) (*Bot, error)` | Create independent Bot instance |

**Config:** `BotToken` (required), `ChatID` (comma-separated, first = default for Notify)

**Bot methods:** `Notify(msg)`, `RegisterHandler(handlerType, pattern, matchType, handler)`

## Usage Patterns

### Global API

```go
import "github.com/QuantProcessing/notify/feishu"

feishu.Init(feishu.Config{Webhook: "https://open.feishu.cn/open-apis/bot/v2/hook/xxx"})
feishu.SendText("Hello!")
```

```go
import "github.com/QuantProcessing/notify/telegram"

telegram.Init(telegram.Config{BotToken: "123456:ABC-DEF...", ChatID: "12345678"})
telegram.Notify("Trade executed!")
```

### Multi-Instance

```go
bot := feishu.NewBot(feishu.Config{Webhook: "https://..."})
bot.SendText("Hello from instance!")
```

### Rich Text (Feishu)

```go
feishu.SendRichText("Alert", [][]feishu.PostElem{
    {feishu.NewTextElem("Server "), feishu.NewAElem("down", "https://example.com")},
    {feishu.NewAtElem("ou_user_id")},
})
```

### Telegram Long Polling

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
telegram.Start(ctx) // blocks until ctx is cancelled
```

## Common Mistakes

- **Forgetting `Init()` before global functions** → returns `ErrNotInitialized`
- **Missing `AppID`/`AppSecret` for urgent calls** → silently falls back to webhook `SendText`
- **Invalid Telegram chat ID format** → silently skipped, check logs
- **Calling `Start()` without `Init()`** → silently returns, check logs
