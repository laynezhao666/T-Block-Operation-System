package rpc

import (
	"context"

	"etrpc-go/log"
	"etrpc-go/util/httputil"

	"alarm-manage/entity/message"
)

// SendMarkdownButton 发送Markdown按钮消息
//
//	@param ctx
//	@param url
//	@param message 文本消息
//	@param stra_key 用于机器人回调
func SendMarkdownButton(ctx context.Context, url string, msg string, attachment message.MarkdownAttachments) {
	log.Infof("[SendMarkdown] message: %s, attachment: %+v", msg, attachment)
	markdownMsg := message.MarkdownButtonMsg{
		MsgType: "markdown",
		Markdown: message.MarkdownButtonContent{
			Content:     msg,
			Attachments: attachment,
		},
	}
	res := make(map[string]any)
	err := httputil.PostJson(ctx, url, nil, markdownMsg, &res)
	if err != nil {
		log.Errorf("[SendMarkdown] post hook fail %s", err.Error())
		return
	}
}

// SendMarkdown 发送Markdown文本消息
//
//	@param ctx
//	@param url
//	@param message 文本消息
//	@param stra_key 用于机器人回调
func SendMarkdown(ctx context.Context, url string, msg string) {
	log.Infof("[SendMarkdown] message: %s", msg)
	markdownMsg := message.MarkdownTextMsg{
		MsgType: "markdown",
		Markdown: message.MarkdownTextContent{
			Content: msg,
		},
	}
	res := make(map[string]any)
	err := httputil.PostJson(ctx, url, nil, markdownMsg, &res)
	if err != nil {
		log.Errorf("[SendMarkdown] post hook fail %s", err.Error())
		return
	}
}
