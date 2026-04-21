// Package http 实现HTTP门禁控制器协议的驱动层。
package http

import (
	"dac/entity/config"
	"dac/entity/utils/dhttp"
)

// Reset 重置控制器（MDC版本不支持此操作）
func (c *Controller) Reset() error {
	if c.isVersionMDC {
		// MDC 版本协议无该接口
		return nil
	}

	url := c.urlProducer.ResetURL()
	if config.C.IsEnableCompatible() {
		// 兼容垃圾厂家不规范的 http 实现
		// ContentLength 为 0 的请求可能会被某些垃圾厂商的门禁丢弃
		return dhttp.PostJSONWithoutContentLength(url, c.timeout)
	}
	// POST 方法，以文档中的示例为准
	return dhttp.PostJSONWithoutData(url, c.timeout)
}
