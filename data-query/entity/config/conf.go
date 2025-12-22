// Package conf conf
package conf

import "etrpc-go/config"

func init() {
	config.RegisterConfig("serverconf.yaml", &ServerConf, true)
}

var (
	ServerConf ServerConfStruct
)

// ServerConfStruct 云端配置
type ServerConfStruct struct {
	// 过期时间
	ExpireTimeSinceQuery int `yaml:"ExpireTimeSinceQuery"`
	// 过期时间余量
	ExpireTimeMargin int `yaml:"ExpireTimeMargin"`
	// 耗时判断阈值(常规)
	NormalCostThreshold int `yaml:"NormalCostThreshold"`
	// 耗时判断阈值（长耗时）
	ExtremelyCostThreshold int `yaml:"ExtremelyCostThreshold"`
	// 测点处理并发数
	ProcessPointConcurrencyLimit int `yaml:"ProcessPointConcurrencyLimit"`
	// 变化测点单批查询数量
	QueryChangedBatchSize int `yaml:"QueryChangedBatchSize"`
	// 变化测点并发数
	QueryChangedConcurrencyLimit int `yaml:"QueryChangedConcurrencyLimit"`
	// Interval限制
	IntervalLimit int64 `yaml:"IntervalLimit"`
	// TimeDuration限制
	TimeDurationLimit int `yaml:"TimeDurationLimit"`
}
