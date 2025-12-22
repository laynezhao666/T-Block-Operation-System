// Package config 提供了应用程序的配置管理功能
package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"agent/entity/consts"
)

// Config 程序相关配置
type Config struct {
	Feature     map[string]int    `mapstructure:"feature"`
	Collector   CollectorConfig   `mapstructure:"collector" yaml:"collector"`
	Distributor DistributorConfig `mapstructure:"distributor" yaml:"distributor"`
	Plugin      PluginConfig      `mapstructure:"plugin"`
	Project     ProjectConfig     `mapstructure:"project"`
	Task        TaskConfig        `mapstructure:"task"`
	Test        TestConfig        `mapstructure:"test"`
	AuthConf    AuthConf          `mapstructure:"auth" yaml:"auth"`
	Tbox        TboxConfig        `mapstructure:"tbox" yaml:"tbox"`
}

// TboxConfig Tbox相关配置
type TboxConfig struct {
	Ip                   string   `mapstructure:"ip" yaml:"ip"`
	Gids                 []string `mapstructure:"gids" yaml:"gids"`
	HeartbeatIntervalMs  int64    `mapstructure:"heartbeat_interval_ms" yaml:"heartbeat_interval_ms"`
	HeartbeatTbusEnabled bool     `mapstructure:"heartbeat_tbus_enabled" yaml:"heartbeat_tbus_enabled"`
}

// TaskConfig 任务相关配置
type TaskConfig struct {
	Mode     string           `mapstructure:"mode"` // 任务模式：schedule 调度服务 、local 本地
	Schedule TaskScheduleData `mapstructure:"schedule"`
	Local    TaskLocalData    `mapstructure:"local"`
}

// TaskScheduleData 任务调度相关配置
type TaskScheduleData struct {
	MaxProcessCap int64 `mapstructure:"max_process_cap" yaml:"max_process_cap"`
}

// TaskLocalData 本地任务相关配置
type TaskLocalData struct {
	Devs []string `mapstructure:"devs"`
}

// PluginConfig 插件相关配置
type PluginConfig struct {
	InterruptionJudgeThreshold int `mapstructure:"interruption_judge_threshold" yaml:"interruption_judge_threshold"`
	PluginCallInterval         int `mapstructure:"plugin_call_interval" yaml:"plugin_call_interval"`
}

// CollectorConfig 采集相关配置
type CollectorConfig struct {
	Common DriverCommonConfig `mapstructure:"common"`
	Snmp   SnmpConfig         `mapstructure:"snmp" yaml:"snmp"`
	Modbus ModbusConfig       `mapstructure:"modbus"`
}

// ModbusConfig Modbus相关配置
type ModbusConfig struct {
	SerialsMap SerialsMapData `yaml:"serials_map"`
}

// SerialsMapData 串口相关配置
type SerialsMapData struct {
	COMs map[string]COMConfig `yaml:",inline"`
}

// COMConfig 串口相关配置
type COMConfig struct {
	Baud    string `yaml:"baud"`
	Databit string `yaml:"databit"`
	Dev     string `yaml:"dev"`
	ID      string `yaml:"id"`
	Mode    string `yaml:"mode,omitempty"` // omitempty 表示如果没有值则不输出该字段
	Parity  string `yaml:"parity"`
	Stopbit string `yaml:"stopbit"`
}

// DriverCommonConfig 驱动通用配置
type DriverCommonConfig struct {
	PacketFailedCount  int `mapstructure:"packet_failed_count" yaml:"packet_failed_count"`
	RequestFailedCount int `mapstructure:"request_failed_count" yaml:"request_failed_count"`
	RequestFailedTime  int `mapstructure:"request_failed_time" yaml:"request_failed_time"`
	CollectionInterval int `mapstructure:"collection_interval" yaml:"collection_interval"`
}

// SnmpConfig snmp相关配置
type SnmpConfig struct {
	Timeout             int `mapstructure:"timeout" yaml:"timeout"`
	Retry               int `mapstructure:"retry" yaml:"retry"`
	LogInterval         int `mapstructure:"log_interval" yaml:"log_interval"`
	MaxCoroutine        int `mapstructure:"max_coroutine" yaml:"max_coroutine"`
	PerCoroutinePoints  int `mapstructure:"per_coroutine_points" yaml:"per_coroutine_points"`
	RequestTotalTimeout int `mapstructure:"request_total_timeout" yaml:"request_total_timeout"`
}

// DistributorConfig 分发相关配置
type DistributorConfig struct {
	Common    DistributorCommonConfig `mapstructure:"common"`
	Kafka     KafkaInstanceInfo       `mapstructure:"kafka"`
	MsgSender MsgSenderInfo           `mapstructure:"msg_sender" yaml:"msg_sender"`
	TBus      TBusInfo                `mapstructure:"tbus" yaml:"tbus"`
	Http      HttpInfo                `mapstructure:"http"`
	Tlink     TlinkInfo               `mapstructure:"tlink"`
	Forwards  []KafkaInstanceInfo     `mapstructure:"forwards"`
}

// DistributorCommonConfig 分发通用配置
type DistributorCommonConfig struct {
	IntervalReportSecond int `mapstructure:"interval_report_second" yaml:"interval_report_second"`
}

// MsgSenderInfo 公网转发通道配置
type MsgSenderInfo struct {
	Enable []string `yaml:"enable"`
}

// TBusInfo 旧版边端tbus转发通道配置
type TBusInfo struct {
	Enable []string `yaml:"enable"`
}

// WhiteRange 白名单范围
type WhiteRange struct {
	LowerLimit int64 `yaml:"lower_limit"`
	UpperLimit int64 `yaml:"upper_limit"`
}

// HttpInfo http 北向上报配置
type HttpInfo struct {
	Enable         []string `yaml:"enable"`
	NorthWhitelist []string `yaml:"north_whitelist"`
}

// TlinkInfo tlink 上报配置
type TlinkInfo struct {
	Enable []string `yaml:"enable"`
}

// KafkaInstanceInfo kafka配置
type KafkaInstanceInfo struct {
	Brokers []string `mapstructure:"brokers" yaml:"brokers"`
	Log     string   `mapstructure:"log" yaml:"log"`
	Topic   struct {
		Points string `mapstructure:"points"`
	} `mapstructure:"topic"`
	MaxAttempt     int    `mapstructure:"max_attempt" yaml:"max_attempt"`
	WriteTimeoutMs int    `mapstructure:"write_timeout" yaml:"write_timeout"`
	Compression    string `mapstructure:"compression"`
	SASL           struct {
		Mechanism string `mapstructure:"mechanism"`
		Username  string `mapstructure:"username"`
		Password  string `mapstructure:"password"`
	} `mapstructure:"sasl"`
}

// ProjectConfig 工程配置
type ProjectConfig struct {
	Source      string `mapstructure:"source"`
	ModuleGroup string `mapstructure:"module_group" yaml:"module_group"`
	Mode        string `mapstructure:"mode"` // 运行模式：agent、agent-gw
}

// TestConfig 测试相关模拟数据
type TestConfig struct {
	ModbusIP string `mapstructure:"modbus_ip" yaml:"modbus_ip"`
	SnmpAddr string `mapstructure:"snmp_addr" yaml:"snmp_addr"`
}

// AuthConf 鉴权配置
type AuthConf struct {
	SysIdKey       map[string]string `mapstructure:"sys_id_key" yaml:"sys_id_key"`
	CloseAuth      bool
	Debug          bool
	Refer2Bus      map[string]string
	BanRefer       bool
	SecurityConf   SecurityConf `mapstructure:"security" yaml:"security"`
	RequiredRoutes []string     `mapstructure:"required_routes" yaml:"required_routes"`
}

// SecurityConf 安全配置
type SecurityConf struct {
	Enabled     bool
	Debug       bool
	BanRefer    bool
	MaxBodySize int
}

// GetSysKeyById 获取系统ID
func (c *Config) GetSysKeyById(id string) string {
	if c == nil {
		return ""
	}

	key, ok := c.AuthConf.SysIdKey[id]
	if !ok {
		return ""
	}
	return key
}

// LoadIntOrDefault 加载int值
func LoadIntOrDefault(v int, d int) int {
	if v == 0 {
		return d
	}
	return v
}

// IsFeatureEnable 是否启用特性
func (c *Config) IsFeatureEnable(k string) bool {
	if c == nil {
		return false
	}

	enable, ok := c.Feature[k]
	if !ok {
		return false
	}
	return enable == 1
}

// IsKafkaLogEnable 是否启用kafka日志
func (c *Config) IsKafkaLogEnable() bool {
	return c.Distributor.Kafka.Log == "1"
}

// IsStdCalEnable 是否启用标准计算
func (c *Config) IsStdCalEnable() bool {
	return c.IsFeatureEnable("standard_calculation")
}

// IsCollectReportEnable 是否启用采集上报
func (c *Config) IsCollectReportEnable() bool {
	return c.IsFeatureEnable("collect_report")
}

// IsSimulationEnable 是否启用模拟
func (c *Config) IsSimulationEnable() bool {
	return c.IsFeatureEnable("simulation")
}

// IsAlarmTestEnable 是否启用告警测试
func (c *Config) IsAlarmTestEnable() bool {
	return c.IsFeatureEnable("alarm_test")
}

// IsDevTaskLocalEnable 1 - 采集任务启用本地配置模式； 0 - 任务由调度中心下发
func (c *Config) IsDevTaskLocalEnable() bool {
	return c.IsFeatureEnable("devs_local")
}

// IsBackupPushEnable 是否启用备份推送
func (c *Config) IsBackupPushEnable() bool {
	return c.IsFeatureEnable("backup_push")
}

func (c *Config) errorFiled(field string) error {
	return fmt.Errorf("%v of %+v is empty", field, *c)
}

func (c *Config) validateKafka() error {
	if len(c.Distributor.Kafka.Brokers) == 0 {
		return c.errorFiled("Distributor.Kafka.Brokers")
	}

	if len(c.Distributor.Kafka.Topic.Points) == 0 {
		return c.errorFiled("Distributor.Kafka.Topic.Points")
	}

	if c.Distributor.Kafka.MaxAttempt <= 0 {
		c.Distributor.Kafka.MaxAttempt = consts.DefaultKafkaMaxAttempt
	}
	if c.Distributor.Kafka.WriteTimeoutMs <= 0 {
		c.Distributor.Kafka.WriteTimeoutMs = consts.DefaultKafkaWriteTimeoutMs
	}

	// todo forward功能
	for i := range c.Distributor.Forwards {
		if len(c.Distributor.Forwards[i].Topic.Points) == 0 {
			c.Distributor.Forwards[i].Topic.Points = c.Distributor.Kafka.Topic.Points
		}
	}

	return nil
}

// IsGatewayMode 是否网关模式
func (c *Config) IsGatewayMode() bool {
	if c.Project.Mode == "agent-gw" {
		return true
	}
	return false
}

func (c *Config) validate() error {
	if c == nil {
		return errors.New("receiver is nil")
	}

	var err error
	if err = c.loadProjectModuleGroup(); err != nil {
		return err
	}
	if err = c.validateKafka(); err != nil {
		return err
	}

	return nil
}

// LoadProjectModuleGroup 获取模组ID，如果是正则表达式，从环境变量HOSTNAME中获取模组ID，否则直接作为模组ID
func (c *Config) loadProjectModuleGroup() error {
	moduleGroupID := GetStringValue(c.Project.ModuleGroup, consts.ModuleGroupRegex)
	if strings.HasPrefix(moduleGroupID, consts.ModuleGroupRegexPrefix) {
		re := regexp.MustCompile(strings.TrimPrefix(moduleGroupID, consts.ModuleGroupRegexPrefix))
		hostname := os.Getenv("HOSTNAME")
		match := re.FindStringSubmatch(hostname)
		if len(match) < 2 {
			return fmt.Errorf("获取环境变量中的模组信息失败, HOSTNAME:%v", hostname)
		}
		moduleGroupID = match[1]
	}
	c.Project.ModuleGroup = moduleGroupID
	return nil
}

// GetProjectModuleGroup 获取模组ID
func (c *Config) GetProjectModuleGroup() string {
	return c.Project.ModuleGroup
}

// GetProjectLocalPath 获取项目本地路径
func (c *Config) GetProjectLocalPath() string {
	return consts.ProjectPath + "/" + c.Project.ModuleGroup
}

// GetRB 获取全局配置
func GetRB() *Config {
	return &Conf
}

// GetStringValue 获取字符串值
func GetStringValue(value string, defaultValue string) string {
	if len(value) > 0 {
		return value
	}
	return defaultValue
}

// GetSerialConfig 获取串口配置
func (c *Config) GetSerialConfig() map[string]COMConfig {
	return c.Collector.Modbus.SerialsMap.COMs
}

// GetCheckTime 获取检测时间间隔
func (c *Config) GetCheckTime() string {
	s, ok := c.Feature["check_time"]
	if !ok {
		return "30"
	}
	return fmt.Sprint(s)
}

// GetDefaultIP4 获取默认ip
func GetDefaultIP4() string {
	// TODO 在tbox模式下正确获取需要的ip（配置的ip）
	return "127.0.0.1"
}
