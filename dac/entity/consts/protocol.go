// Package consts 定义门禁系统的全局常量。
package consts

// 门禁协议类型常量
const (
	ProtocolHTTP     = "http"     // HTTP协议
	ProtocolXBrother = "xbrother" // XBrother协议
	ProtocolChd806d4 = "chd806d4" // CHD806D4协议
	ProtocolCACS     = "cacs"     // CACS协议

	V3ProtocolVersion      = "v3"  // V3版本
	DefaultProtocolVersion = "v2"  // 默认V2版本
	V1ProtocolVersion      = "v1"  // V1版本
	MDCProtocolVersion     = "mdc" // MDC版本
)

// 告警类型常量
const (
	AlarmTypeOpenAlarm    = 0 // 开门告警
	AlarmTypeOpenAbnormal = 1 // 开门异常
)
