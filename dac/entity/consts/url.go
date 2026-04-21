// Package consts 定义门禁系统的全局常量。
package consts

// GID映射服务的默认URL配置
const (
	// DefaultGIDMappingPrefix GID映射服务默认地址前缀
	DefaultGIDMappingPrefix = "http://tadaptor-gidmapping:50091"
	// DefaultGIDMappingURLLocation 获取边缘位置信息URL
	DefaultGIDMappingURLLocation = "/getEdgeLocation"
	// DefaultGIDMappingURLConvertCode GID转换URL
	DefaultGIDMappingURLConvertCode = "/cgi/go/GIDMappingService/iot/collect_gid"
	// DefaultGIDMappingURLRooms 获取房间列表URL
	DefaultGIDMappingURLRooms = "/cgi/go/GIDMappingService/iot/room/all"

	// RedisKeyGetLockChannelIP 获取锁通道IP的Redis Key
	RedisKeyGetLockChannelIP = "dac_get_lock_channel_ip"
)
