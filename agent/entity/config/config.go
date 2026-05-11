// Package config 提供了应用程序的配置管理功能
package config

import (
	"bytes"
	"errors"
	"etrpc-go/config/loader/file"

	"fmt"
	"os"
	"regexp"
	"strings"

	"agent/entity/consts"

	"gopkg.in/yaml.v3"
)

// Config 程序相关配置
type Config struct {
	Feature      map[string]int     `mapstructure:"feature"`
	Collector    CollectorConfig    `mapstructure:"collector" yaml:"collector"`
	Distributor  DistributorConfig  `mapstructure:"distributor" yaml:"distributor"`
	Plugin       PluginConfig       `mapstructure:"plugin"`
	Project      ProjectConfig      `mapstructure:"project"`
	Task         TaskConfig         `mapstructure:"task"`
	Test         TestConfig         `mapstructure:"test"`
	AuthConf     AuthConf           `mapstructure:"auth" yaml:"auth"`
	Tbox         TboxConfig         `mapstructure:"tbox" yaml:"tbox"`
	StdConf      StdConfig          `mapstructure:"std" yaml:"std"`
	MonitorProxy MonitorProxyConfig `mapstructure:"monitor_proxy" yaml:"monitor_proxy"`
}

// TboxConfig Tbox相关配置
type TboxConfig struct {
	Ip                   string   `mapstructure:"ip" yaml:"ip"`
	Gids                 []string `mapstructure:"gids" yaml:"gids"`
	HeartbeatIntervalMs  int64    `mapstructure:"heartbeat_interval_ms" yaml:"heartbeat_interval_ms"`
	HeartbeatTbusEnabled bool     `mapstructure:"heartbeat_tbus_enabled" yaml:"heartbeat_tbus_enabled"`
	SnReportEnabled      bool     `mapstructure:"sn_report_enabled" yaml:"sn_report_enabled"`
	ElvdbTarget          string   `mapstructure:"elvdb_target" yaml:"elvdb_target"`
	ElvdbSnUrl           string   `mapstructure:"elvdb_sn_url" yaml:"elvdb_sn_url"`
}

// TaskConfig 任务相关配置
type TaskConfig struct {
	Mode     string           `mapstructure:"mode"` // 任务模式：schedule 调度服务 、local 本地
	Schedule TaskScheduleData `mapstructure:"schedule"`
	Local    TaskLocalData    `mapstructure:"local"`
}

// TaskScheduleData 任务调度相关配置
type TaskScheduleData struct {
	MaxProcessCap              int64 `mapstructure:"max_process_cap" yaml:"max_process_cap"`
	RecvPort                   int64 `mapstructure:"recv_port" yaml:"recv_port"`
	CommDisconnectDurationSecs int64 `mapstructure:"comm_disconnect_duration_secs" yaml:"comm_disconnect_duration_secs"` // 通讯中断持续时间阈值（秒），默认 30s
	RemoveTasksOnUnhealthy     *bool `mapstructure:"remove_tasks_on_unhealthy" yaml:"remove_tasks_on_unhealthy"`         // 上报 UNHEALTHY 时是否移除本地任务，默认 true
}

// TaskLocalData 本地任务相关配置
type TaskLocalData struct {
	Devs           []string                 `mapstructure:"devs" yaml:"devs"`
	HotStandby     map[string]HotStandbyDev `mapstructure:"hot_standby" yaml:"hot_standby"`
	CollectSlave   bool                     `mapstructure:"collect_slave" yaml:"collect_slave"`     // 是否采集从设备
	DetectTimeout  int                      `mapstructure:"detect_timeout" yaml:"detect_timeout"`   // 探测超时时间
	DetectFailNum  int                      `mapstructure:"detect_fail_num" yaml:"detect_fail_num"` // 探测失败次数
	DetectInterval int                      `mapstructure:"detect_interval" yaml:"detect_interval"` // 探测间隔
	DirectCalc     bool                     `mapstructure:"direct_calc" yaml:"direct_calc"`
}

// HotStandbyDev 热备相关配置
type HotStandbyDev struct {
	IsMaster bool   `mapstructure:"is_master" yaml:"is_master"` // 是否为主设备
	Ip       string `mapstructure:"ip" yaml:"ip"`
	Port     int64  `mapstructure:"port" yaml:"port"`
}

// PluginConfig 插件相关配置
type PluginConfig struct {
	InterruptionJudgeThreshold int `mapstructure:"interruption_judge_threshold" yaml:"interruption_judge_threshold"`
	PluginCallInterval         int `mapstructure:"plugin_call_interval" yaml:"plugin_call_interval"`
}

// CollectorConfig 采集相关配置
type CollectorConfig struct {
	Common       DriverCommonConfig `mapstructure:"common"`
	Snmp         SnmpConfig         `mapstructure:"snmp" yaml:"snmp"`
	Modbus       ModbusConfig       `mapstructure:"modbus"`
	StartupProbe StartupProbeConfig `mapstructure:"startup_probe" yaml:"startup_probe"`
}

// StartupProbeConfig k8s 启动探针相关配置
type StartupProbeConfig struct {
	// DisconnectThreshold 通讯断开设备占比阈值（整数，0~100），超过该阈值时探针返回服务不可用
	DisconnectThreshold int `mapstructure:"disconnect_threshold" yaml:"disconnect_threshold"`
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
	StdWorkerCount     int `mapstructure:"std_worker_count" yaml:"std_worker_count"`
	StdLoopTime        int `mapstructure:"std_loop_time" yaml:"std_loop_time"`
}

// SnmpConfig snmp相关配置
type SnmpConfig struct {
	Timeout             int `mapstructure:"timeout" yaml:"timeout"`
	Retry               int `mapstructure:"retry" yaml:"retry"`
	LogInterval         int `mapstructure:"log_interval" yaml:"log_interval"`
	MaxCoroutine        int `mapstructure:"max_coroutine" yaml:"max_coroutine"`
	PerCoroutinePoints  int `mapstructure:"per_coroutine_points" yaml:"per_coroutine_points"`
	RequestTotalTimeout int `mapstructure:"request_total_timeout" yaml:"request_total_timeout"`
	MaxOidCount         int `mapstructure:"max_oid_count" yaml:"max_oid_count"`
}

// DistributorConfig 分发相关配置
type DistributorConfig struct {
	Common     DistributorCommonConfig `mapstructure:"common"`
	Kafka      KafkaInstanceInfo       `mapstructure:"kafka"`
	MsgSender  MsgSenderInfo           `mapstructure:"msg_sender" yaml:"msg_sender"`
	TBus       TBusInfo                `mapstructure:"tbus" yaml:"tbus"`
	Http       HttpInfo                `mapstructure:"http"`
	Bypass     BypassInfo              `mapstructure:"bypass" yaml:"bypass"`
	Tlink      TlinkInfo               `mapstructure:"tlink"`
	Test       TestInfo                `mapstructure:"test"`
	Forwards   []KafkaInstanceInfo     `mapstructure:"forwards"`
	MqttConfig MqttConfig              `mapstructure:"deviceModel" yaml:"deviceModel"`
}

// DistributorCommonConfig 分发通用配置
type DistributorCommonConfig struct {
	IntervalReportSecond int `mapstructure:"interval_report_second" yaml:"interval_report_second"`
}
type MqttConfig struct {
	Enable   []string `yaml:"enable"`
	Broker   string   `yaml:"broker"`
	ClientID string   `yaml:"client_id"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	Qos      int      `yaml:"qos"`
	Retain   bool     `yaml:"retain"`
	TimeoutC int      `yaml:"timeout_connect"`
	TimeoutR int      `yaml:"timeout_rw"`
	Debug    bool     `yaml:"debug"`
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

// BypassInfo 旁路分发配置
type BypassInfo struct {
	Enable   []string `yaml:"enable"` // collect_change, collect_interval, std_change, std_interval
	Target   string   `yaml:"target"`
	ClientId string   `yaml:"client_id"`
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

type TestInfo struct {
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

// StdConfig 标准层相关配置
type StdConfig struct {
	HideCodeList []string `mapstructure:"hide_code" yaml:"hide_code"`
}

// MonitorProxyConfig 监控代理配置
type MonitorProxyConfig struct {
	Enabled             bool   `mapstructure:"enabled" yaml:"enabled"`                               // 是否启用代理上报
	ProxyTarget         string `mapstructure:"proxy_target" yaml:"proxy_target"`                     // 代理服务 target（trpc client 目标）
	AppMark             string `mapstructure:"app_mark" yaml:"app_mark"`                             // 当前服务的 appMark
	MetricGroup         string `mapstructure:"metric_group" yaml:"metric_group"`                     // 指标组名称
	Timeout             int    `mapstructure:"timeout" yaml:"timeout"`                               // 请求超时时间(ms)，默认 3000
	QueueSize           int    `mapstructure:"queue_size" yaml:"queue_size"`                         // 队列大小，默认 10000
	BatchSize           int    `mapstructure:"batch_size" yaml:"batch_size"`                         // 批量上报大小，默认 100
	FlushInterval       int    `mapstructure:"flush_interval" yaml:"flush_interval"`                 // 定时刷新间隔(ms)，默认 1000
	TimeSkewThresholdMs int    `mapstructure:"time_skew_threshold_ms" yaml:"time_skew_threshold_ms"` // 时间偏差阈值(ms)，默认 5000
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

// IsStdCalOnlyChangedPoints 标准计算是否只计算变化点
func (c *Config) IsStdCalOnlyChangedPoints() bool {
	return c.IsFeatureEnable("std_cal_only_changed")
}

// IsStdValueRangeCheckEnable 是否启用标准点值有效范围检测
func (c *Config) IsStdValueRangeCheckEnable() bool {
	return c.IsFeatureEnable("std_range_check")
}

// IsCollectReportEnable 是否启用采集上报
func (c *Config) IsCollectReportEnable() bool {
	return c.IsFeatureEnable("collect_report")
}

// IsSimulationEnable 是否启用模拟
func (c *Config) IsSimulationEnable() bool {
	return c.IsFeatureEnable("simulation")
}

// IsSimpleBridgeLansEnable 是否启用简版网桥网口配置
func (c *Config) IsSimpleBridgeLansEnable() bool {
	return c.IsFeatureEnable("simple_bridge_lans")
}

// IsAlarmTestEnable 是否启用告警测试
func (c *Config) IsAlarmTestEnable() bool {
	return c.IsFeatureEnable("alarm_test")
}

// IsLocalDebugEnable 是否启用本地调试
func (c *Config) IsLocalDebugEnable() bool {
	return c.IsFeatureEnable("local_debug")
}

// IsDevTaskLocalEnable 1 - 采集任务启用本地配置模式； 0 - 任务由调度中心下发
func (c *Config) IsDevTaskLocalEnable() bool {
	return c.IsFeatureEnable("devs_local")
}

// IsBackupPushEnable 是否启用备份推送
func (c *Config) IsBackupPushEnable() bool {
	return c.IsFeatureEnable("backup_push")
}

// IsDisableSetTime 是否禁用时间设置接口
func (c *Config) IsDisableSetTime() bool {
	return c.IsFeatureEnable("disable_set_time")
}

// IsOpcSpecialId 是否特殊opc配置
func (c *Config) IsOpcSpecialId() bool {
	return c.IsFeatureEnable("opc_special_id")
}

// IsDriverCommOptimizeEnable 是否启用采集驱动通讯状态优化
// 开启后：1. 驱动设备为空时更新通讯状态 2. 驱动Open成功后才清空统计数据
func (c *Config) IsDriverCommOptimizeEnable() bool {
	return c.IsFeatureEnable("driver_comm_optimize")
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

// GetProjectPath 获取项目根目录路径
func (c *Config) GetProjectPath() string {
	return consts.ProjectPath
}

// GetProjectLocalPath 获取项目本地路径
func (c *Config) GetProjectLocalPath() string {
	return c.GetProjectPath() + "/" + c.Project.ModuleGroup
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

// HotStandbyEnable 判断是否开启热备
func (c *Config) HotStandbyEnable() bool {
	return c.IsDevTaskLocalEnable() && len(c.Task.Local.HotStandby) > 0
}

// GetDefaultIP4 获取默认ip
func GetDefaultIP4() string {
	// TODO 在tbox模式下正确获取需要的ip（配置的ip）
	return "127.0.0.1"
}

// Update 更新trpc_go.yaml里的配置
func (c *Config) Update(key []string, value any) error {
	fileName := file.GetLocalConfigPath()
	// 读取trpc_go.yaml文件
	content, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read trpc_go.yaml: %v", err)
	}
	newContent, err := updateYaml(key, value, content)
	if err != nil {
		return err
	}
	if err = os.WriteFile(fileName, newContent, 0644); err != nil {
		return fmt.Errorf("failed to write trpc_go.yaml: %v", err)
	}
	return nil
}

func updateYaml(key []string, value any, content []byte) ([]byte, error) {
	// 使用yaml.v3的Node结构来解析yaml，保留格式和注释
	var node yaml.Node
	if err := yaml.Unmarshal(content, &node); err != nil {
		return content, fmt.Errorf("failed to unmarshal yaml: %v", err)
	}

	// 递归更新指定路径的字段
	if err := updateNestedFieldNode(&node, key, value); err != nil {
		return content, fmt.Errorf("failed to update field: %v", err)
	}

	// 将修改后的内容写回文件，保留格式，使用2个空格的缩进
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(&node); err != nil {
		return content, fmt.Errorf("failed to marshal yaml: %v", err)
	}

	return buf.Bytes(), nil
}

// updateNestedFieldNode 递归更新嵌套字段，使用yaml.Node结构
func updateNestedFieldNode(node *yaml.Node, keys []string, value any) error {
	if len(keys) == 0 {
		return errors.New("empty key path")
	}

	// 如果node是文档节点，获取其内容
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		return updateNestedFieldNode(node.Content[0], keys, value)
	}

	// 如果node是映射节点，查找对应的key
	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			if keyNode.Value == keys[0] {
				// 如果是最后一个key，直接设置值
				if len(keys) == 1 {
					// 创建新的值节点
					newValueNode := &yaml.Node{}
					if err := newValueNode.Encode(value); err != nil {
						return fmt.Errorf("failed to encode value: %v", err)
					}
					// 替换原有的值节点
					node.Content[i+1] = newValueNode
					return nil
				}

				// 递归处理嵌套字段
				return updateNestedFieldNode(valueNode, keys[1:], value)
			}
		}

		// 如果key不存在，需要创建新的字段
		if len(keys) == 1 {
			// 创建新的key-value对
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: keys[0],
			}
			valueNode := &yaml.Node{}
			if err := valueNode.Encode(value); err != nil {
				return fmt.Errorf("failed to encode value: %v", err)
			}
			node.Content = append(node.Content, keyNode, valueNode)
			return nil
		} else {
			// 创建嵌套的映射节点
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: keys[0],
			}
			nestedNode := &yaml.Node{
				Kind: yaml.MappingNode,
			}
			node.Content = append(node.Content, keyNode, nestedNode)
			return updateNestedFieldNode(nestedNode, keys[1:], value)
		}
	}

	return fmt.Errorf("unexpected node kind: %v at key '%s'", node.Kind, keys[0])
}

// updateNestedField 递归更新嵌套字段
func updateNestedField(data map[string]any, keys []string, value any) error {
	if len(keys) == 0 {
		return errors.New("empty key path")
	}

	currentKey := keys[0]

	// 如果是最后一个key，直接设置值
	if len(keys) == 1 {
		data[currentKey] = value
		return nil
	}

	// 处理嵌套字段
	if nestedData, exists := data[currentKey]; exists {
		if nestedMap, ok := nestedData.(map[string]any); ok {
			return updateNestedField(nestedMap, keys[1:], value)
		}
		return fmt.Errorf("field '%s' is not a map", currentKey)
	}

	// 如果字段不存在，创建新的嵌套map
	newMap := make(map[string]any)
	data[currentKey] = newMap
	return updateNestedField(newMap, keys[1:], value)
}
