# Notify

[![Go Reference](https://pkg.go.dev/badge/github.com/QuantProcessing/notify.svg)](https://pkg.go.dev/github.com/QuantProcessing/notify)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**English** | [中文](README_zh.md)

A lightweight Go library for sending notifications via **Feishu (Lark)** and **Telegram**.

## Features

**Feishu / Lark**
- Webhook messages (text, rich-text / post)
- SDK-based messaging via official Lark Open Platform
- Phone urgent call notifications

**Telegram**
- Bot with chat ID whitelist middleware
- Simple `Notify()` helper for one-off messages
- Long-polling command handler with `Start()`

## Requirements

- Go 1.24+

## Install

```bash
go get github.com/QuantProcessing/notify
```

Import only what you need:

```go
import "github.com/QuantProcessing/notify/feishu"
import "github.com/QuantProcessing/notify/telegram"
```

### AI Agent Integration

```bash
npx skills add QuantProcessing/notify
```

## Usage

### Feishu

#### Quick Start (Global API)

```go
package main

import (
    "log"
    "github.com/QuantProcessing/notify/feishu"
)

func main() {
    feishu.Init(feishu.Config{
        Webhook: "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
    })

    if err := feishu.SendText("Hello from Go!"); err != nil {
        log.Fatal(err)
    }
}
```

#### Rich Text (Post)

```go
err := feishu.SendRichText("Alert", [][]feishu.PostElem{
    {feishu.NewTextElem("Server "), feishu.NewAElem("down", "https://example.com")},
    {feishu.NewAtElem("ou_user_id")},
})
```

#### Urgent Phone Call

Requires App ID, App Secret, and User Open ID:

```go
feishu.Init(feishu.Config{
    Webhook:    "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
    AppID:      "cli_xxx",
    AppSecret:  "xxx",
    UserOpenID: "ou_xxx",
})

if err := feishu.SendUrgentText("CRITICAL: System Down!"); err != nil {
    log.Fatal(err)
}
```

#### Multi-Instance Usage

```go
bot := feishu.NewBot(feishu.Config{
    Webhook: "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
})

if err := bot.SendText("Hello!"); err != nil {
    log.Fatal(err)
}
```

---

### Telegram

#### Quick Start (Global API)

```go
package main

import (
    "log"
    "github.com/QuantProcessing/notify/telegram"
)

func main() {
    if err := telegram.Init(telegram.Config{
        BotToken: "123456:ABC-DEF...",
        ChatID:   "12345678",          // comma-separated for multiple
    }); err != nil {
        log.Fatal(err)
    }

    if err := telegram.Notify("Trade executed: BUY 0.01 BTC @ $65,000"); err != nil {
        log.Fatal(err)
    }
}
```

#### Start Bot (Long Polling)

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

telegram.Start(ctx) // blocks until ctx is cancelled
```

#### Multi-Instance Usage

```go
bot, err := telegram.NewBot(telegram.Config{
    BotToken: "123456:ABC-DEF...",
    ChatID:   "12345678",
})
if err != nil {
    log.Fatal(err)
}

if err := bot.Notify("Hello!"); err != nil {
    log.Fatal(err)
}
```

## Architecture

```
notify/
├── feishu/          # Feishu (Lark) notification package
│   ├── bot.go       # Global API + Bot struct
│   ├── client.go    # HTTP webhook + Lark SDK client
│   └── types.go     # Message type definitions
├── telegram/        # Telegram notification package
│   ├── bot.go       # Global API + Bot struct
│   └── middleware.go # Chat ID authorization middleware
├── go.mod
├── LICENSE          # MIT
└── README.md
```

## License

[MIT](LICENSE)
