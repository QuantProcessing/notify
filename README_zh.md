# Notify

[![Go Reference](https://pkg.go.dev/badge/github.com/QuantProcessing/notify.svg)](https://pkg.go.dev/github.com/QuantProcessing/notify)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[English](README.md) | **中文**

轻量级 Go 通知库，支持 **飞书 (Lark)** 和 **Telegram**。

## 功能

**飞书 / Lark**
- Webhook 消息（文本、富文本 / post）
- 通过官方 Lark 开放平台 SDK 发送消息
- 电话加急通知

**Telegram**
- 带有 Chat ID 白名单中间件的 Bot
- 简单的 `Notify()` 消息发送
- 支持 `Start()` 长轮询接收命令

## 环境要求

- Go 1.24+

## 安装

```bash
go get github.com/QuantProcessing/notify
```

按需导入：

```go
import "github.com/QuantProcessing/notify/feishu"
import "github.com/QuantProcessing/notify/telegram"
```

## 配置

参考 [.env.example](.env.example) 了解所需的环境变量。

## 使用

### 飞书

#### 快速开始（全局 API）

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

#### 富文本消息

```go
err := feishu.SendRichText("告警", [][]feishu.PostElem{
    {feishu.NewTextElem("服务器 "), feishu.NewAElem("宕机", "https://example.com")},
    {feishu.NewAtElem("ou_user_id")},
})
```

#### 电话加急

需要 App ID、App Secret 和 User Open ID：

```go
feishu.Init(feishu.Config{
    Webhook:    "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
    AppID:      "cli_xxx",
    AppSecret:  "xxx",
    UserOpenID: "ou_xxx",
})

if err := feishu.SendUrgentText("严重: 系统宕机!"); err != nil {
    log.Fatal(err)
}
```

#### 多实例用法

```go
bot := feishu.NewBot(feishu.Config{
    Webhook: "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
})

if err := bot.SendText("你好!"); err != nil {
    log.Fatal(err)
}
```

---

### Telegram

#### 快速开始（全局 API）

```go
package main

import (
    "log"
    "github.com/QuantProcessing/notify/telegram"
)

func main() {
    if err := telegram.Init(telegram.Config{
        BotToken: "123456:ABC-DEF...",
        ChatID:   "12345678",          // 多个用逗号分隔
    }); err != nil {
        log.Fatal(err)
    }

    if err := telegram.Notify("交易已执行: 买入 0.01 BTC @ $65,000"); err != nil {
        log.Fatal(err)
    }
}
```

#### 启动 Bot（长轮询）

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

telegram.Start(ctx) // 阻塞直到 ctx 被取消
```

#### 多实例用法

```go
bot, err := telegram.NewBot(telegram.Config{
    BotToken: "123456:ABC-DEF...",
    ChatID:   "12345678",
})
if err != nil {
    log.Fatal(err)
}

if err := bot.Notify("你好!"); err != nil {
    log.Fatal(err)
}
```

## 项目结构

```
notify/
├── feishu/          # 飞书 (Lark) 通知包
│   ├── bot.go       # 全局 API + Bot 结构体
│   ├── client.go    # HTTP webhook + Lark SDK 客户端
│   └── types.go     # 消息类型定义
├── telegram/        # Telegram 通知包
│   ├── bot.go       # 全局 API + Bot 结构体
│   └── middleware.go # Chat ID 鉴权中间件
├── go.mod
├── LICENSE          # MIT
└── README.md
```

## 许可证

[MIT](LICENSE)
