// Package virtualpoints 管理门禁控制器的虚拟测点，
// 用于在采集失败时生成离线状态等虚拟数据。
package virtualpoints

import "dac/entity/config"

// defaultMaxAllowedFailedRequestCount 默认最大允许连续失败请求次数
const (
	defaultMaxAllowedFailedRequestCount = 10
)

// maxAllowedFailedRequestCount 当前生效的最大允许连续失败请求次数
var (
	maxAllowedFailedRequestCount = defaultMaxAllowedFailedRequestCount
)

// Init 初始化虚拟测点配置，从配置文件读取最大失败请求次数
func Init() {
	maxAllowedFailedRequestCount = config.C.GetNumber("request_failed_count", defaultMaxAllowedFailedRequestCount)
	if maxAllowedFailedRequestCount <= 0 || maxAllowedFailedRequestCount > defaultMaxAllowedFailedRequestCount {
		maxAllowedFailedRequestCount = defaultMaxAllowedFailedRequestCount
	}
}
