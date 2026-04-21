// Package consts 定义门禁系统的全局常量。
package consts

// 测点数据的JSON键名
const (
	KeyValue     = "pv"  // 测点值
	KeyQua       = "qua" // 测点质量
	KeyTimestamp = "tms" // 时间戳
)

// 控制器配置相关的键名和常量
const (
	KeyProtocolHTTPKey        = "http_key"                  // HTTP协议密钥
	KeyURLMode                = "url_mode"                  // URL模式: "default" 或 "specific"
	KeyDoorNum                = "door_num"                  // 门数量
	KeyFetchEventInterval     = "fetch_event_interval"      // 事件拉取间隔
	KeyFetchLoopEventInterval = "fetch_loop_event_interval" // 事件循环拉取间隔
	KeyFetchAlarmInterval     = "fetch_alarm_interval"      // 告警拉取间隔
	KeyFetchLoopAlarmInterval = "fetch_loop_alarm_interval" // 告警循环拉取间隔
	KeySyncedByTimestamp      = "synced_by_timestamp"       // 是否按时间戳同步
	OneDoorPerController      = 1                           // 单门控制器
	TwoDoorPerController      = 2                           // 双门控制器
	FourDoorPerController     = 4                           // 四门控制器
)

// EnableCompatible 启用兼容模式标识
const (
	EnableCompatible = "compatible"
)

// PointMessageMozuKey 测点消息中模组ID的键名
const (
	PointMessageMozuKey = "mozu"
)

// 控制器属性键名
const (
	AttrChannel  = "channel"  // 通道属性
	AttrProtocol = "protocol" // 协议属性
)
