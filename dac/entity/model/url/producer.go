// Package url 定义门禁控制器URL生成接口。
package url

// BaseInfo URL生成所需的基础信息
type BaseInfo struct {
	ChannelID string // 通道ID
	ApiKey    string // API密钥
}

// Producer URL生成器接口，定义各类门禁操作的URL生成方法
type Producer interface {
	// GetDoorPositionStateURL 获取门位置状态URL
	GetDoorPositionStateURL() string

	// GetDoorStateURL 获取门状态URL
	GetDoorStateURL() string
	// SetDoorStateURL 设置门状态URL
	SetDoorStateURL() string

	// GetDoorsURL 获取门列表URL
	GetDoorsURL() string

	// GetDoorParameterURL 获取门参数URL
	GetDoorParameterURL() string
	// SetDoorParameterURL 设置门参数URL
	SetDoorParameterURL() string

	// GetHistoryEventHisURL 获取历史事件URL（按索引）
	GetHistoryEventHisURL(recordIndex interface{}) string
	// GetHistoryEventByTimestampURL 获取历史事件URL（按时间戳）
	GetHistoryEventByTimestampURL(begin, end string, recordIndex interface{}) string
	// GetMDCEventURL 获取MDC事件URL
	GetMDCEventURL(recordIndex interface{}) string
	// GetHistoryAlarmURL 获取历史告警URL（按索引）
	GetHistoryAlarmURL(alarmIndex interface{}) string
	// GetHistoryAlarmByTimestampURL 获取历史告警URL（按时间戳）
	GetHistoryAlarmByTimestampURL(begin, end string, alarmIndex interface{}) string
	// GetMDCAlarmURL 获取MDC告警URL
	GetMDCAlarmURL(alarmIndex interface{}) string

	// GetTimeGroupURL 获取时间组URL
	GetTimeGroupURL(groupNo interface{}) string
	// SetTimeGroupURL 设置时间组URL
	SetTimeGroupURL() string
	// ClearTimeGroupURL 清除时间组URL
	ClearTimeGroupURL() string

	// GetTimeURL 获取时间URL
	GetTimeURL() string
	// SetTimeURL 设置时间URL
	SetTimeURL() string

	// GetCardsURL 获取卡列表URL
	GetCardsURL(cardIndex interface{}) string
	// GetAllCardsURL 获取所有卡URL
	GetAllCardsURL() string
	// AddCardURL 添加卡URL
	AddCardURL() string
	// GetCardURL 获取单张卡URL
	GetCardURL(cardNo string) string
	// UpdateCardURL 更新卡URL
	UpdateCardURL() string
	// DeleteCardURL 删除卡URL
	DeleteCardURL() string

	// AddUserURL 添加用户URL
	AddUserURL() string
	// DeleteUserURL 删除用户URL
	DeleteUserURL() string

	// CleanURL 清除数据URL
	CleanURL() string

	// ResetURL 重置URL
	ResetURL() string

	// GetCurrentAlarmURL 获取当前告警URL
	GetCurrentAlarmURL() string

	// NeedDataPrefix 返回是否需要在POST body中添加"data="前缀
	NeedDataPrefix() bool
}
