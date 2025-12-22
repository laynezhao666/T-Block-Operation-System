// Package message message
package message

// MarkdownAction MarkdownAction
type MarkdownAction struct {
	Name        string `json:"name"`         // 唯一即可
	Text        string `json:"text"`         // 按钮文案
	Type        string `json:"type"`         // 只能是button
	Value       string `json:"value"`        // action的值
	ReplaceText string `json:"replace_text"` // 点击按钮后显示的值
}

// MarkdownAttachments MarkdownAttachments
type MarkdownAttachments struct {
	CallbackID string           `json:"callback_id"` // 用于回调的id
	Actions    []MarkdownAction `json:"actions"`     // 按钮列表
}

// MarkdownButtonContent MarkdownButtonContent
type MarkdownButtonContent struct {
	Content     string              `json:"content"`               // markdown文本内容
	Attachments MarkdownAttachments `json:"attachments,omitempty"` // 长度固定是1
}

// MarkdownButtonMsg MarkdownButtonMsg
type MarkdownButtonMsg struct {
	MsgType  string                `json:"msgtype"`
	Markdown MarkdownButtonContent `json:"markdown"`
}

// MarkdownTextContent MarkdownTextContent
type MarkdownTextContent struct {
	Content string `json:"content"` // markdown文本内容
}

// MarkdownTextMsg MarkdownTextMsg
type MarkdownTextMsg struct {
	MsgType  string              `json:"msgtype"`
	Markdown MarkdownTextContent `json:"markdown"`
}

// RspBotMsg RspBotMsg
type RspBotMsg struct {
	WebhookUrl string `json:"webhook_url"`
	MsgId      string `json:"msgid"`
	ChatId     string `json:"chatid"`
	PostId     string `json:"postid"`
	ChatType   string `json:"chattype"`
	From       struct {
		Userid string `json:"userid"`
		Name   string `json:"name"`
		Alias  string `json:"alias"`
	} `json:"from"`
	GetChatInfoUrl string                 `json:"get_chat_info_url"`
	MsgType        string                 `json:"msgtype"`
	Attachment     BotCallbackAttachments `json:"attachment"`
	Image          struct {
		ImageUrl string `json:"image_url"`
	} `json:"image"`
	Text struct {
		Content string `json:"content"`
	} `json:"text"`
	Event struct {
		EventType string `json:"event_type"`
	} `json:"event"`
}

// BotCallbackAttachments BotCallbackAttachments
type BotCallbackAttachments struct {
	CallbackID string           `json:"callbackid"` // 用于回调的id
	Actions    []MarkdownAction `json:"actions"`    // 按钮列表
}

// BotImage BotImage
type BotImage struct {
	Base64 string `json:"base64"` // 图片base64
	MD5    string `json:"md5"`    // 图片md5
}

// BotImageMessage BotImageMessage
type BotImageMessage struct {
	MsgType string   `json:"msgtype"`
	Image   BotImage `json:"image"`
}
