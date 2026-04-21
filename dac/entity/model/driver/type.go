// Package driver 提供门禁驱动层的通用数据模型和工具函数。
package driver

// MethodGetDoorPositionState 获取门位置状态方法名
// MethodGetDoorState 获取门状态方法名
// MethodSetDoorState 设置门状态方法名
// MethodGetDoors 获取门列表方法名
// MethodSetDoorParameter 设置门参数方法名
// MethodGetDoorParameter 获取门参数方法名
// MethodGetEvent 获取事件方法名
// MethodGetEventByTime 按时间获取事件方法名
// MethodGetAlarm 获取告警方法名
// MethodGetAlarmByTime 按时间获取告警方法名
// MethodGetTimeGroup 获取时间组方法名
// MethodSetTimeGroup 设置时间组方法名
// MethodClearTimeGroup 清除时间组方法名
// MethodSetTime 设置时间方法名
// MethodGetTime 获取时间方法名
// MethodGetAllCards 获取所有卡方法名
// MethodGetCards 获取卡列表方法名
// MethodAddUser 添加用户方法名
// MethodAddCard 添加卡方法名
// MethodUpdateCardFlag 更新卡标志方法名
// MethodUpdateCardStaff 更新卡人员方法名
// MethodUpdateCard 更新卡方法名
// MethodDeleteCard 删除卡方法名
// MethodDeleteUser 删除用户方法名
// MethodGetCard 获取单张卡方法名
// MethodClean 清除数据方法名
// MethodReset 重置方法名
// MethodGetCurrentAlarm 获取当前告警方法名
const (
	MethodGetDoorPositionState = "get_door_position_state"

	MethodGetDoorState = "get_door_state"
	MethodSetDoorState = "set_door_state"

	MethodGetDoors         = "get_door"
	MethodSetDoorParameter = "set_door_parameter"
	MethodGetDoorParameter = "get_door_parameter"

	MethodGetEvent       = "get_event"
	MethodGetEventByTime = "get_event_by_time"
	MethodGetAlarm       = "get_alarm"
	MethodGetAlarmByTime = "get_alarm_by_time"

	MethodGetTimeGroup   = "get_time_group"
	MethodSetTimeGroup   = "set_time_group"
	MethodClearTimeGroup = "clear_time_group"

	MethodSetTime = "set_time"
	MethodGetTime = "get_time"

	MethodGetAllCards     = "get_all_cards"
	MethodGetCards        = "get_cards"
	MethodAddUser         = "add_user"
	MethodAddCard         = "add_card"
	MethodUpdateCardFlag  = "update_card_flag"
	MethodUpdateCardStaff = "update_card_staff"
	MethodUpdateCard      = "update_card"
	MethodDeleteCard      = "delete_card"
	MethodDeleteUser      = "delete_user"
	MethodGetCard         = "get_card"

	MethodClean = "clean"
	MethodReset = "reset"

	MethodGetCurrentAlarm = "get_current_alarm"
)

// DoorStateType 门状态类型
type DoorStateType int

// StateClose 关门状态
// StateOpen 开门状态
// StateNormallyOpen 常开状态
// StateNormallyClose 常闭状态
const (
	StateClose         = DoorStateType(0)
	StateOpen          = DoorStateType(1)
	StateNormallyOpen  = DoorStateType(2) // 常开
	StateNormallyClose = DoorStateType(3) // 常闭
)

// DoorStatusType 门运行状态类型
type DoorStatusType int

// StatusNormal 正常状态
// StatusKeepOpen 保持开门状态
// StatusKeepClose 保持关门状态
const (
	StatusNormal    = DoorStatusType(0)
	StatusKeepOpen  = DoorStatusType(1)
	StatusKeepClose = DoorStatusType(2)
)

// OpenModeType 开门模式类型
type OpenModeType int

// OpenModeCard 刷卡开门
// OpenModePassword 密码开门
// OpenModeCardAndCardPassword 卡+密码开门
// OpenModeCardOrDoorPassword 卡或密码开门
const (
	OpenModeCard                = OpenModeType(0)
	OpenModePassword            = OpenModeType(1)
	OpenModeCardAndCardPassword = OpenModeType(2)
	OpenModeCardOrDoorPassword  = OpenModeType(2)
)

// DoorNumberType 门编号类型
type DoorNumberType int

// SetDoorStateRequest 设置门状态请求，key为门编号，value为目标状态
type SetDoorStateRequest map[DoorNumberType]DoorStateType

// DoorParameter 门参数配置
type DoorParameter struct {
	Number         DoorNumberType `json:"number" gorm:"column:number"`
	Name           string         `json:"name" gorm:"column:name"`
	Password       string         `json:"password" gorm:"column:password"`
	KeepOpenTime   int            `json:"keep_open_time" gorm:"column:keep_open_time"`   // 单位：秒
	OpenTimeout    int            `json:"open_timeout" gorm:"column:open_timeout"`       // 单位：秒
	LockCount      int            `json:"lock_count" gorm:"column:lock_count"`           // 非法卡允许的最长失败次数
	LockTime       int            `json:"lock_time" gorm:"column:lock_time"`             // 非法卡的封锁时间，单位：秒
	VerifyInterval int            `json:"verify_interval" gorm:"column:verify_interval"` // 非法卡刷卡间隔，单位：秒
	OpenMode       OpenModeType   `json:"open_mode" gorm:"column:open_mode"`
	FireSignalMode int            `json:"fire_signal_mode" gorm:"column:fire_signal_mode"`
}

// DirectionType 进出方向类型
type DirectionType int

// DirectionEnter 进入方向
// DirectionLeave 离开方向
const (
	DirectionEnter = DirectionType(0)
	DirectionLeave = DirectionType(1)
)

// EventType 事件类型
type EventType int

// EventTypeRemoteOpen 远程开门事件
const (
	EventTypeRemoteOpen = EventType(3)
)

// Event 门禁事件记录
type Event struct {
	Index       int            `json:"index" column:"column:index"`
	Timestamp   int64          `json:"timestamp" column:"column:timestamp"` // 秒级时间戳
	CardNumber  string         `json:"card_number" column:"column:card_number"`
	Username    string         `json:"username" column:"column:username"`
	DoorNumber  DoorNumberType `json:"door_number" column:"column:door_number"`
	Direction   DirectionType  `json:"direction" column:"column:direction"`
	Type        EventType      `json:"type" column:"column:type"`
	Description string         `json:"description" column:"column:description"`
}

// EventData 事件数据集合，包含分页信息
type EventData struct {
	Offset int // 当前待获取的第一条记录的索引
	Last   int // 最后一条记录的索引
	// 当前获取到的记录数据所在区间的结束时间戳
	// 仅通过时间戳获取记录时使用该字段
	EndTimestamp int64
	Events       []Event
}

// AlarmType 告警类型
type AlarmType int

// AlarmStateType 告警状态类型
type AlarmStateType int

// AlarmStateRecovery 告警恢复状态
// AlarmStateAlarming 告警中状态
const (
	AlarmStateRecovery = AlarmStateType(0)
	AlarmStateAlarming = AlarmStateType(1)
)

// Alarm 门禁告警记录
type Alarm struct {
	Index       int            `json:"index" column:"column:index"`
	Timestamp   int64          `json:"timestamp" column:"column:timestamp"` // 秒级时间戳
	DoorNumber  DoorNumberType `json:"door_number" column:"column:door_number"`
	Type        AlarmType      `json:"type" column:"column:type"`
	State       AlarmStateType `json:"state" column:"column:state"`
	Description string         `json:"description" column:"column:channel_id"`
}

// AlarmData 告警数据集合，包含分页信息
type AlarmData struct {
	Offset int // 当前待获取的第一条记录的索引
	Last   int // 最后一条记录的索引
	// 当前获取到的记录数据所在区间的结束时间戳
	// 仅通过时间戳获取记录时使用该字段
	EndTimestamp int64
	Alarms       []Alarm
}

// CardData 卡数据集合，包含分页信息
type CardData struct {
	Offset int
	Total  int
	Cards  []Card
}

// CurrentAlarmEvent 当前告警事件
type CurrentAlarmEvent struct {
	Type int
	Desc string
}

// CurrentAlarmData 当前告警数据，按门分组
type CurrentAlarmData struct {
	Door   int
	Alarms []CurrentAlarmEvent
}

// TimeInterval 时间区间，[BeginTimestamp, EndTimestamp)
type TimeInterval struct {
	BeginTimestamp int64
	EndTimestamp   int64
}

// IsValid 判断时间区间是否有效
func (t *TimeInterval) IsValid() bool {
	return t != nil && t.BeginTimestamp < t.EndTimestamp
}
