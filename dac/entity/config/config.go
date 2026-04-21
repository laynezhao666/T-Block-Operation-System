// Package config 提供门禁系统的配置管理。
package config

import (
	"etrpc-go/config"
)

// C 全局配置实例
var (
	C = &Config{}
)

// init 注册配置到框架，支持热更新。
func init() {
	config.RegisterConfigWithPrefix("dac.config", "dac", C, true)
}

// InterruptionToleranceMozu 某些模组的门禁性能较差，定制指定模组阈值
type InterruptionToleranceMozu struct {
	MozuId   string `yaml:"mozu_id"`
	MaxCount int    `yaml:"max_count"`
}

// Config 门禁系统配置结构体
type Config struct {
	VerifyMode   bool `yaml:"verify_mode"`
	Debug        bool `yaml:"debug"`
	ReportToTBOS bool `yaml:"report_to_tbos"`
	NotStandard  bool `yaml:"not_standard"`

	SyncFromCMDB    bool     `yaml:"sync_from_cmdb"`
	SyncGidFromCMDB bool     `yaml:"sync_gid_from_cmdb"`
	CMDBSyncMozus   []string `yaml:"cmdb_sync_mozus"`
	SyncInterval    int      `yaml:"sync_interval"`

	Enable        map[string]int `yaml:"enable"`
	EnablePooling bool           `yaml:"enable_pooling"`

	IgnoreGIDMozus   []string `yaml:"ignore_gid_mozus"`
	IgnoreFetchMozus []string `yaml:"ignore_fetch_mozus"`

	InterruptionToleranceMozus []InterruptionToleranceMozu `yaml:"interruption_tolerance_mozus"`

	ExpirationTime int `yaml:"expiration_time"`
	DeletionTime   int `yaml:"deletion_time"`

	UseGetEventsMozus []string `yaml:"use_get_events_mozus"`

	LoggingPacket struct {
		Hosts []string `yaml:"hosts"` // 指定 host: ip:port
	} `yaml:"logging_packet"`

	// 通过时间同步历史记录
	FetchByTime struct {
		Mozus []string `yaml:"mozus"` // 指定模组
		Hosts []string `yaml:"hosts"` // 指定 host: ip:port
	} `yaml:"fetch_by_time"`

	GIDMapping struct {
		Prefix string `yaml:"prefix"`
		URL    struct {
			Location    string `yaml:"location"`
			ConvertCode string `yaml:"convert_code"`
			Rooms       string `yaml:"rooms"`
		} `yaml:"url"`
	} `yaml:"gidmapping"`

	Number map[string]int `yaml:"number"`
}

// getValue 从map中获取值，不存在则返回默认值。
func getValue(m map[string]int, k string, value int) int {
	v, ok := m[k]
	if ok {
		return v
	}
	return value
}

// GetNumber 获取指定key的数值配置，不存在则返回默认值。
func (c *Config) GetNumber(k string, defaultValue int) int {
	return getValue(c.Number, k, defaultValue)
}

// IsEnable 检查指定功能是否启用。
func (c *Config) IsEnable(k string) bool {
	v, ok := c.Enable[k]
	if ok && v > 0 {
		return true
	}

	return false
}

// IsEnableCompatible 检查兼容模式是否启用。
func (c *Config) IsEnableCompatible() bool {
	return c.IsEnable("compatible")
}

// IsEnablePooling 检查连接池模式是否启用。
func (c *Config) IsEnablePooling() bool {
	return c.EnablePooling
}

// IgnoreGID 检查指定模组是否忽略GID同步。
func (c *Config) IgnoreGID(mozu string) bool {
	for _, m := range c.IgnoreGIDMozus {
		if m == mozu {
			return true
		}
	}
	return false
}

// IsSyncFromCMDB 检查是否从CMDB同步数据。
func (c *Config) IsSyncFromCMDB() bool {
	return c.SyncFromCMDB
}

// IsSyncGidFromCMDB 检查是否从CMDB同步GID。
func (c *Config) IsSyncGidFromCMDB() bool {
	return c.SyncGidFromCMDB
}

// ToleranceMozuMaxCount 获取中断容忍模组的最大中断次数
func (c *Config) ToleranceMozuMaxCount(mozu string) int {
	for _, m := range c.InterruptionToleranceMozus {
		if m.MozuId == mozu {
			return m.MaxCount
		}
	}
	return 0
}

// IgnoreFetch 检查指定模组是否忽略数据拉取。
func (c *Config) IgnoreFetch(mozu string) bool {
	for _, m := range c.IgnoreFetchMozus {
		if m == mozu {
			return true
		}
	}
	return false
}

// UseGetEvents 检查指定模组是否使用GetEvents接口。
func (c *Config) UseGetEvents(mozu string) bool {
	for _, m := range c.UseGetEventsMozus {
		if m == mozu {
			return true
		}
	}
	return false
}

// IsFetchByTime 检查指定模组或主机是否通过时间同步历史记录。
func (c *Config) IsFetchByTime(mozuID string, host string) bool {
	for _, m := range c.FetchByTime.Mozus {
		if m == mozuID {
			return true
		}
	}
	for _, h := range c.FetchByTime.Hosts {
		if h == host {
			return true
		}
	}
	return false
}

// IsLoggingPacket 检查指定主机是否开启报文日志。
func (c *Config) IsLoggingPacket(host string) bool {
	for _, h := range c.LoggingPacket.Hosts {
		if h == host {
			return true
		}
	}
	return false
}
