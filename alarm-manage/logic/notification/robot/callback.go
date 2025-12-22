// Package robot robot
package robot

import (
	"context"
	"encoding/xml"
	"sync"

	"etrpc-go/log"

	thttp "trpc.group/trpc-go/trpc-go/http"

	"alarm-manage/conf"
	"alarm-manage/entity/message"
)

var (
	botCrypt *WXBizMsgCrypt
	once     sync.Once
)

// GetWXBizMsgCrypt 获取企微消息加解密对象
func GetWXBizMsgCrypt() *WXBizMsgCrypt {
	once.Do(func() {
		token, AESKey := conf.RobotConfig.Token, conf.RobotConfig.EncodingAESKey
		botCrypt = NewWXBizMsgCrypt(token, AESKey, "", XmlType)
	})
	return botCrypt
}

// DecodeMsg 从请求中解析出消息
func DecodeMsg(ctx context.Context) (bool, *message.RspBotMsg) {
	head := thttp.Head(ctx)
	reqVariable := head.Request.URL.Query()
	msgSignature := reqVariable.Get("msg_signature")
	timestamp := reqVariable.Get("timestamp")
	nonce := reqVariable.Get("nonce")
	log.Info("机器人收到的url为", head.Request.URL.String())
	log.Info("机器人收到的body为", string(head.ReqBody))
	if reqVariable.Has("echostr") {
		// 企微发送的验证消息，需返回解码后的数据
		echoStr := reqVariable.Get("echostr")
		echoStrData, cryptErr := GetWXBizMsgCrypt().VerifyURL(msgSignature, timestamp, nonce, echoStr)
		if cryptErr != nil {
			log.WarnContext(ctx, "verifyUrl fail", cryptErr)
		}
		_, _ = head.Response.Write(echoStrData)
		return false, nil
	} else {
		msgStr, cryptError := GetWXBizMsgCrypt().DecryptMsg(msgSignature, timestamp, nonce, head.ReqBody)
		if cryptError != nil {
			log.WarnContext(ctx, "decryptMsg fail", cryptError)
		}
		var msg = &message.RspBotMsg{}
		_ = xml.Unmarshal(msgStr, &msg)
		return true, msg
	}
}
