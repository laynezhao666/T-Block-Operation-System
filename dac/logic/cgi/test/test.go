// Package test 提供门禁控制器的连通性测试功能。
package test

import (
	"fmt"
	"strconv"
	"time"

	"dac/entity/consts"
	"dac/entity/model/driver"
	"dac/entity/model/rt"
	"dac/logic/collect/template"

	"dac/entity/utils/encoding"
)

// Ping 测试门禁控制器的网络连通性。
// 根据协议类型创建临时控制器实例，发送Ping请求验证设备是否在线。
// 支持HTTP、XBrother、CACS等多种协议，HTTP协议会自动生成认证密钥。
func Ping(arg rt.PingArgs) error {
	// 解析超时时间
	timeout, err := strconv.ParseInt(arg.Timeout, 0, 64)
	if err != nil {
		return fmt.Errorf("error timeout value: \"%v\"", arg.Timeout)
	}
	info := driver.ChannelInfo{
		ChannelID:       arg.Host,
		Protocol:        arg.ProtocolName,
		ProtocolVersion: arg.ProtocolVersion,
		TimeoutMS:       time.Duration(timeout) * time.Millisecond}
	if len(info.Protocol) == 0 {
		info.Protocol = consts.ProtocolHTTP
	}

	t := template.GetManager().GetTemplate(info.Protocol)
	if t == nil {
		return fmt.Errorf("can not get template of protocol \"%v\"", info.Protocol)
	}

	switch info.Protocol {
	case consts.ProtocolHTTP:
		account := arg.Account
		password := arg.Password
		if len(account) == 0 && len(password) == 0 {
			account = consts.DefaultControllerAccount
			password = consts.DefaultControllerPassword
		}
		info.Extend = map[string]interface{}{
			consts.KeyProtocolHTTPKey: encoding.MD5String(account + password),
		}
	}

	c := t.GetDriver().CreateController(0, "test")
	r := c.Open(info)
	if r != consts.QualityOK {
		return fmt.Errorf("open %+v error", arg)
	}
	defer func() {
		_ = c.Close()
	}()

	if err := c.Ping(); err != nil {
		return fmt.Errorf("ping %+v error: %w", arg, err)
	}
	return nil
}
