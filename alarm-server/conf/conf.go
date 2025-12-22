// Package conf conf
package conf

import "etrpc-go/config"

// ServerConfStruct 云端配置
type ServerConfStruct struct {
	RuleValidConfig       RuleValidConfig       `yaml:"RuleValidConfig"`
	SyncCacheConfig       SyncCacheConfig       `yaml:"SyncCacheConfig"`
	ValidRedisCacheConfig ValidRedisCacheConfig `yaml:"ValidRedisCacheConfig"`
}

// RuleValidConfig 告警生效率上报配置
type RuleValidConfig struct {
	RegularStoreInterval int32 `yaml:"RegularStoreInterval"`
	BatchSize            int32 `yaml:"BatchSize"`
}

// SyncCacheConfig 同步缓存配置
type SyncCacheConfig struct {
	StrategyCacheInterval    int32 `yaml:"StrategyCacheInterval"`
	StrategyCacheBatchSize   int32 `yaml:"StrategyCacheBatchSize"`
	StrategyTotalIntervalCnt int32 `yaml:"StrategyTotalIntervalCnt"`
	DeviceCacheInterval      int32 `yaml:"DeviceCacheInterval"`
	DeviceCacheBatchSize     int32 `yaml:"DeviceCacheBatchSize"`
	DeviceTotalIntervalCnt   int32 `yaml:"DeviceTotalIntervalCnt"`
}

// ValidRedisCacheConfig 缓存配置
type ValidRedisCacheConfig struct {
	MGetBatchSize int32 `yaml:"MGetBatchSize"`
	MSetBatchSize int32 `yaml:"MSetBatchSize"`
	MDelBatchSize int32 `yaml:"MDelBatchSize"`
}

var (
	ServerConf ServerConfStruct
)

func init() {
	config.RegisterConfig("serverconf.yaml", &ServerConf, true)
}
