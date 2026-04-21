package http

import (
	"dac/entity/config"
	"dac/entity/utils/dhttp"
)

func (c *Controller) Clean() error {
	if c.isVersionMDC {
		// MDC 版本协议无该接口
		return nil
	}

	url := c.urlProducer.CleanURL()
	if config.C.IsEnableCompatible() {
		// 兼容厂家不规范的 http 实现
		// ContentLength 为 0 的请求可能会被某些垃圾厂商的门禁丢弃
		return dhttp.PostJSONWithoutContentLength(url, c.timeout)
	}
	// POST 方法，以文档中的示例为准
	return dhttp.PostJSONWithoutData(url, c.timeout)
}
