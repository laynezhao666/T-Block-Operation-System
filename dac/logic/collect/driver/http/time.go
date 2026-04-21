// Package http 实现HTTP协议门禁控制器的驱动层。
package http

import (
	"dac/entity/utils/dhttp"

	"dac/entity/utils/ttime"
)

// TimePayload 时间同步请求/响应载荷
type TimePayload struct {
	CurrentTime string `json:"current_time"` // 当前时间字符串
}

// GetTime 从控制器获取当前时间
func (c *Controller) GetTime() (string, error) {
	var t TimePayload
	url := c.urlProducer.GetTimeURL()
	err := dhttp.GetJSON(url, c.timeout, &t)
	return t.CurrentTime, err
}

// SetTime 同步本地时间到控制器
func (c *Controller) SetTime() error {
	if c.isVersionMDC {
		// MDC 版本协议无该接口
		return nil
	}

	var req TimePayload
	// 使用本地时间格式化为控制器要求的格式
	req.CurrentTime = ttime.GetNowLocal().Format("20060102T150405+8")
	return c.postJSON(c.urlProducer.SetTimeURL(), req, nil)
}
