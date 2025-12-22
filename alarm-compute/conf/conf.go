// Package conf conf
package conf

import (
	"etrpc-go/config"
)

// ServerConfStruct 云端配置
type ServerConfStruct struct {
	//单次查询数据模块测点数量限制
	MaxPointCountForDataQuery int32                `yaml:"MaxPointCountForDataQuery"`
	RealTimeConfig            RealTimeConfig       `yaml:"RealTimeConfig"`
	DelayTimeConfig           DelayTimeConfig      `yaml:"DelayTimeConfig"`
	VirtualConfig             VirtualConfig        `yaml:"VirtualConfig"`
	ActiveAlarmCache          ActiveAlarmCache     `yaml:"ActiveAlarmCache"`
	HeartBeatConfig           HeartBeatConfig      `yaml:"HeartBeatConfig"`
	ValidateRecordConfig      ValidateRecordConfig `yaml:"ValidateRecordConfig"`
	MysqlTableName            MysqlTableName       `yaml:"MysqlTableName"`
}

// RealTimeConfig 实时策略Config
type RealTimeConfig struct {
	ParallelExecWpSize       int32 `yaml:"ParallelExecWpSize"`
	VaryPointBatchSize       int32 `yaml:"VaryPointBatchSize"`
	VaryPointPoolSize        int32 `yaml:"VaryPointPoolSize"`
	VaryPointQueryTimeSpan   int32 `yaml:"VaryPointQueryTimeSpan"`
	EmptyVaryPointCountLimit int32 `yaml:"EmptyVaryPointCountLimit"`
	RealTimeTaskInterval     int32 `yaml:"RealTimeTaskInterval"`
	RealTimeBatchSize        int32 `yaml:"RealTimeBatchSize"`
	// 每隔多少个小周期，全量分析实时策略
	TotalAnalyzeCycleCount int32 `yaml:"TotalAnalyzeCycleCount"`
	RealTimeTaskPoolSize   int32 `yaml:"RealTimeTaskPoolSize"`
}

// ActiveAlarmCache 活动告警缓存Config
type ActiveAlarmCache struct {
	CacheKeyTimeDuration     int32 `yaml:"CacheKeyTimeDuration"`
	ActiveNormalSyncInterval int32 `yaml:"ActiveNormalSyncInterval"`
	ActiveRequestBatchSize   int32 `yaml:"ActiveRequestBatchSize"`
}

// DelayTimeConfig 延迟策略Config
type DelayTimeConfig struct {
	ParallelExecWpSize       int32 `yaml:"ParallelExecWpSize"`
	VaryPointBatchSize       int32 `yaml:"VaryPointBatchSize"`
	VaryPointPoolSize        int32 `yaml:"VaryPointPoolSize"`
	VaryPointQueryTimeSpan   int32 `yaml:"VaryPointQueryTimeSpan"`
	EmptyVaryPointCountLimit int32 `yaml:"EmptyVaryPointCountLimit"`
	DelayTimeTaskInterval    int32 `yaml:"DelayTimeTaskInterval"`
	TotalAnalyzeCycleCount   int32 `yaml:"TotalAnalyzeCycleCount"`
	DelayTimeTaskPoolSize    int32 `yaml:"DelayTimeTaskPoolSize"`
	// 查询测点历史数据时，并发请求数量
	IntervalRequestPoolSize int32 `yaml:"IntervalRequestPoolSize"`
	// 批量查询历史数据时，设置的batch大小
	IntervalBatchPointSize  int32 `yaml:"IntervalBatchPointSize"`
	DurationRequestPoolSize int32 `yaml:"DurationRequestPoolSize"`
	DurationBatchPointSize  int32 `yaml:"DurationBatchPointSize"`
	JPRangeSec              int32 `yaml:"JPRangeSec"`
}

// VirtualConfig 延迟策略Config
type VirtualConfig struct {
	VaryPointBatchSize       int32 `yaml:"VaryPointBatchSize"`
	VaryPointPoolSize        int32 `yaml:"VaryPointPoolSize"`
	VaryPointQueryTimeSpan   int32 `yaml:"VaryPointQueryTimeSpan"`
	EmptyVaryPointCountLimit int32 `yaml:"EmptyVaryPointCountLimit"`
	VirtualTaskInterval      int32 `yaml:"VirtualTaskInterval"`
	TotalAnalyzeCycleCount   int32 `yaml:"TotalAnalyzeCycleCount"`
	VirtualTaskPoolSize      int32 `yaml:"VirtualTaskPoolSize"`
	IntervalRequestPoolSize  int32 `yaml:"IntervalRequestPoolSize"`
	// 批量查询历史数据时，设置的batch大小
	IntervalBatchPointSize  int32 `yaml:"IntervalBatchPointSize"`
	DurationRequestPoolSize int32 `yaml:"DurationRequestPoolSize"`
	DurationBatchPointSize  int32 `yaml:"DurationBatchPointSize"`
	JPRangeSec              int32 `yaml:"JPRangeSec"`
	RoundPrecision          int32 `yaml:"RoundPrecision"`
	PointKafkaBatchSize     int32 `yaml:"PointKafkaBatchSize"`
	FlushInterval           int32 `yaml:"FlushInterval"`
}

// HeartBeatConfig 心跳配置
type HeartBeatConfig struct {
	HeartBeatInterval int32 `yaml:"HeartBeatInterval"`
}

// ValidateRecordConfig ValidateRecordConfig
type ValidateRecordConfig struct {
	BatchSize              int32 `yaml:"BatchSize"`
	FlushInterval          int32 `yaml:"FlushInterval"`
	FailedDispatchInterval int32 `yaml:"FailedDispatchInterval"`
}

// MysqlTableName MysqlTableName
type MysqlTableName struct {
	ActiveAlarmTable  string `yaml:"ActiveAlarm"`
	HistoryAlarmTable string `yaml:"HistoryAlarm"`
}

var (
	ServerConf ServerConfStruct
)

func init() {
	config.RegisterConfig("serverconf.yaml", &ServerConf, true)
}
