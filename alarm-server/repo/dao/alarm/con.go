package alarm

// ActiveCntFilter ...
type ActiveCntFilter struct {
	MozuId      int64
	Begin       int64
	End         int64
	Level       []string
	Status      []int64
	EventStatus []int64
}

// HistoryCntFilter ...
type HistoryCntFilter struct {
	MozuId int64
	Begin  int64
	End    int64
	Level  []string
}

// ActiveAlarmFilter ...
type ActiveAlarmFilter struct {
	MozuId        int64
	AlarmId       int64
	Rid           int64 // 策略ID
	OccurBegin    string
	OccurEnd      string
	DeviceGid     []string
	DeviceNumber  []string // 设备编码
	Level         []string // 告警等级
	AlarmName     []string // 告警名称
	Content       string   // 告警内容
	Status        []int64  // 挂起状态 0:未挂起 1:挂起
	EventStatus   []int64  // 事件状态 0:未转事件 1:已转事件
	SortType      int64    // 排序类型 1:告警发生时间 2:告警挂起时间 3:告警恢复时间
	CountByMetric bool     // 是否需要按照不同的分类统计告警数量 默认false
	Page          int64    // 页数
	Size          int64    // 页大小
}

// HistoryAlarmFilter ...
type HistoryAlarmFilter struct {
	MozuId        int64
	AlarmId       int64
	Rid           int64
	OccurBegin    string
	OccurEnd      string
	DeviceGid     []string
	DeviceNumber  []string // 设备编码
	Level         []string // 告警等级
	AlarmName     []string // 告警名称
	Content       string   // 告警内容
	RestoreBegin  string
	RestoreEnd    string
	MaxDuration   int64 // 最大持续时间
	MinDuration   int64 // 最小持续时间
	SortType      int64 // 排序类型 1:告警发生时间 2:告警挂起时间 3:告警恢复时间
	CountByMetric bool  // 是否需要按照不同的分类统计告警数量 默认false
	Page          int64 // 页数
	Size          int64 // 页大小
}

// StatisticsFilter ...
type StatisticsFilter struct {
	MozuId     int64
	IsActive   bool
	Level      []string
	Status     int64
	FilterType int64
	Topk       int64
}

// AlarmStatusCon 更新活动告警状态
type AlarmStatusCon struct {
	MozuId       int64
	AlarmIds     []int64
	UserId       uint64
	AlarmStatus  int64
	HangupReason string
	EventStatus  int64
	UpdateTime   string
}

// CloseStatusCon 关闭告警
type CloseStatusCon struct {
	MozuId      int64
	AlarmIds    []int64
	UserId      uint64
	CloseReason string
}

// DelHistoryAlarmCon 删除历史告警
type DelHistoryAlarmCon struct {
	MozuId    int64
	EndTime   string
	Rid       []int64
	DeviceGid []string
	Level     []string
}
