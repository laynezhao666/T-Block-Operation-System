// Package http 实现HTTP门禁控制器协议的驱动层。
package http

import (
	"dac/entity/utils/dhttp"
	"fmt"
	"strings"
	"time"

	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/model/driver"
	"dac/entity/model/url"
	"dac/entity/utils"

	"dac/entity/utils/flog"
	"dac/entity/utils/tlog"
)

// Controller HTTP协议门禁控制器，通过HTTP接口与门禁设备通信
type Controller struct {
	timeout  time.Duration              // 请求超时时间
	baseInfo driver.ControllerBasicInfo // 控制器基本信息
	chanInfo driver.ChannelInfo         // 通道信息

	urlProducer url.Producer // URL生成器

	isVersion1     bool         // 是否V1版本协议
	isVersionMDC   bool         // 是否MDC版本协议
	needDataPrefix bool         // 是否需要data=前缀
	logger         tlog.Logger  // 日志记录器
	filterLogger   *flog.Filter // 过滤日志记录器

	eventFetchWaitTime time.Duration // 事件采集等待时间
	alarmFetchWaitTime time.Duration // 告警采集等待时间
}

// Open 打开门控器连接，初始化协议版本、URL生成器等配置
func (c *Controller) Open(chanInfo driver.ChannelInfo) consts.Quality {
	c.logger = tlog.NewPrefixLogger(fmt.Sprintf("[%v@%v]", chanInfo.ChannelID, c.baseInfo.ID), config.Log)
	c.filterLogger = flog.NewFilterLogger(time.Minute*10, c.logger)

	c.eventFetchWaitTime = utils.GetEventFetchWaitTime(chanInfo.Extend)
	c.alarmFetchWaitTime = utils.GetAlarmFetchWaitTime(chanInfo.Extend)

	c.chanInfo = chanInfo

	version := strings.ToLower(chanInfo.ProtocolVersion)
	if strings.Index(version, consts.V1ProtocolVersion) >= 0 {
		c.isVersion1 = true
	} else if strings.Index(version, consts.MDCProtocolVersion) >= 0 {
		c.isVersionMDC = true
	}

	c.timeout = chanInfo.TimeoutMS

	key := ""
	if chanInfo.Extend != nil {
		key, _ = chanInfo.Extend[consts.KeyProtocolHTTPKey].(string)
	}

	// 优先从门控器配置读取 URL 模式，如果未配置则使用全局配置
	useSpecificProducer := config.C.NotStandard
	if chanInfo.Extend != nil {
		if urlMode, ok := chanInfo.Extend[consts.KeyURLMode].(string); ok {
			switch urlMode {
			case "1": // 非标准北向http协议，url顺序强校验，apikey放最后且-d部分没有“data=”
				useSpecificProducer = true
			case "0": // 标准北向http协议
				useSpecificProducer = false
			}
		}
	}

	c.urlProducer = url.NewDefaultProducer(chanInfo.ChannelID, key)
	if useSpecificProducer {
		c.urlProducer = url.NewSpecificProducer(chanInfo.ChannelID, key)
	}
	c.needDataPrefix = c.urlProducer.NeedDataPrefix()

	return consts.QualityOK
}

// Close 关闭门控器连接，停止过滤日志
func (c *Controller) Close() consts.Quality {
	if c.filterLogger != nil {
		go c.filterLogger.Stop()
	}
	return consts.QualityOK
}

// Ping 检测门控器连接是否正常（通过获取门参数验证）
func (c *Controller) Ping() error {
	_, err := c.GetDoorParameter()
	return err
}

// IsReady 检查门控器是否就绪（HTTP协议始终就绪）
func (c *Controller) IsReady() bool {
	return true
}

// postJSON 根据 urlProducer 的类型选择是否在 POST body 中添加 "data=" 前缀
func (c *Controller) postJSON(url string, reqBody interface{}, dataPointer interface{}) error {
	if c.needDataPrefix {
		return dhttp.PostJSON(url, c.timeout, reqBody, dataPointer)
	}
	return dhttp.PostJSONWithoutDataPrefix(url, c.timeout, reqBody, dataPointer)
}

// getBody 根据 urlProducer 的类型选择是否在请求体中添加 "data=" 前缀
func (c *Controller) getBody(req interface{}) (string, error) {
	if c.needDataPrefix {
		return dhttp.GetBody(req)
	}
	return dhttp.GetBodyWithoutDataPrefix(req)
}
