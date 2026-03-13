package feishu

type MsgType string

const (
	MsgTypeText        MsgType = "text"
	MsgTypePost        MsgType = "post"
	MsgTypeInteractive MsgType = "interactive"
)

type BaseReq struct {
	Timestamp string  `json:"timestamp,omitempty"`
	Sign      string  `json:"sign,omitempty"`
	MsgType   MsgType `json:"msg_type"`
}

type TextReq struct {
	BaseReq
	Content TextContent `json:"content"`
}

type TextContent struct {
	Text string `json:"text"`
}

type PostReq struct {
	BaseReq
	Content PostContentWrapper `json:"content"`
}

type PostContentWrapper struct {
	Post PostBody `json:"post"`
}

type PostBody struct {
	ZhCN *PostContent `json:"zh_cn,omitempty"`
	EnUS *PostContent `json:"en_us,omitempty"`
}

type PostContent struct {
	Title   string       `json:"title"`
	Content [][]PostElem `json:"content"`
}

type PostElem struct {
	Tag      string `json:"tag"`
	Text     string `json:"text,omitempty"`
	Href     string `json:"href,omitempty"`
	UserId   string `json:"user_id,omitempty"`
	UserName string `json:"user_name,omitempty"`
}

func NewTextElem(text string) PostElem {
	return PostElem{
		Tag:  "text",
		Text: text,
	}
}

func NewAElem(text, href string) PostElem {
	return PostElem{
		Tag:  "a",
		Text: text,
		Href: href,
	}
}

func NewAtElem(userId string) PostElem {
	return PostElem{
		Tag:    "at",
		UserId: userId,
	}
}
