// Package db 定义门禁系统的数据库模型和表结构。
package db

import (
	"dac/entity/consts"

	"dac/entity/utils/tgorm"
)

// DriverCard 驱动层卡信息数据库模型
type DriverCard struct {
	ID           IDType `json:"id" gorm:"primaryKey;autoIncrement"`
	ControllerID IDType `json:"controller_id" gorm:"column:controller_id;uniqueIndex:driver_idx_channel_id_card_no;uniqueIndex:driver_idx_channel_id_card_index"`
	ChannelID    string `json:"channel_id" gorm:"column:chid;size:128;uniqueIndex:driver_idx_channel_id_card_no;uniqueIndex:driver_idx_channel_id_card_index"`
	CardNo       string `json:"card_no" gorm:"column:card_no;uniqueIndex:driver_idx_channel_id_card_no;size:30"`
	CardFlag     int    `json:"card_flag" gorm:"column:card_flag"`
	DoorNos      []int  `json:"door,omitempty" gorm:"column:door;serializer:json"`
	TimeGroupNo  int    `json:"group_no" gorm:"column:group_no"`
	UserName     string `json:"user_name" gorm:"column:user_name"`
	Password     string `json:"password" gorm:"column:password"`
	CardIndex    int    `json:"card_index" gorm:"column:card_index;uniqueIndex:driver_idx_channel_id_card_index"`
	Status       int    `json:"status" gorm:"column:status"` // 0表示禁用 1表示启用 2表示删除
	tgorm.Model
}

// TableName 返回驱动层卡信息表名
func (*DriverCard) TableName() string {
	return "t_dac_driver_card"
}

// WeekDay 星期列表类型
type WeekDay = []int

// TimeZone 时间段（起止时间）
type TimeZone struct {
	Begin string `json:"begin"` // 开始时间
	End   string `json:"end"`   // 结束时间
}

// DriverTimeGroup 驱动层时间组数据库模型
type DriverTimeGroup struct {
	ID           IDType     `json:"id" gorm:"primaryKey;autoIncrement"`
	ControllerID IDType     `json:"controller_id" gorm:"column:controller_id;uniqueIndex:driver_idx_channel_id_group_no"`
	ChannelID    string     `json:"channel_id" gorm:"column:chid;uniqueIndex:driver_idx_channel_id_group_no;size:128"`
	GroupNo      int        `json:"group_no" gorm:"column:group_no;uniqueIndex:driver_idx_channel_id_group_no"`
	Week         WeekDay    `json:"week,omitempty" gorm:"column:week;serializer:json"`
	TimeZone     []TimeZone `json:"timezone,omitempty" gorm:"column:timeZone;serializer:json"`
	tgorm.Model
}

// TableName 返回驱动层时间组表名
func (*DriverTimeGroup) TableName() string {
	return "t_dac_driver_time_group"
}

// DoorNumberType 门编号类型
type DoorNumberType int

// OpenModeType 开门模式类型
type OpenModeType int

// DriverDoorParameter 驱动层门参数数据库模型
type DriverDoorParameter struct {
	ID             IDType         `json:"id" gorm:"primaryKey;autoIncrement"`
	ControllerID   IDType         `json:"controller_id" gorm:"column:controller_id;uniqueIndex:driver_idx_channel_id_number"`
	ChannelID      string         `json:"channel_id" gorm:"column:chid;size:128;uniqueIndex:driver_idx_channel_id_number"`
	Number         DoorNumberType `json:"number" gorm:"column:number;uniqueIndex:driver_idx_channel_id_number"`
	Name           string         `json:"name" gorm:"column:name"`
	Password       string         `json:"password" gorm:"column:password"`
	KeepOpenTime   int            `json:"keep_open_time" gorm:"column:keep_open_time"`   // 单位：秒
	OpenTimeout    int            `json:"open_timeout" gorm:"column:open_timeout"`       // 单位：秒
	LockCount      int            `json:"lock_count" gorm:"column:lock_count"`           // 非法卡允许的最长失败次数
	LockTime       int            `json:"lock_time" gorm:"column:lock_time"`             // 非法卡的封锁时间，单位：秒
	VerifyInterval int            `json:"verify_interval" gorm:"column:verify_interval"` // 非法卡刷卡间隔，单位：秒
	OpenMode       OpenModeType   `json:"open_mode" gorm:"column:open_mode"`
	FireSignalMode int            `json:"fire_signal_mode" gorm:"column:fire_signal_mode"`
	tgorm.Model
}

// NewDriverDoorParameter 创建驱动层门参数实例（使用默认值）
func NewDriverDoorParameter(controllerID IDType, channelID string, number int, name string) DriverDoorParameter {
	return DriverDoorParameter{
		ControllerID:   controllerID,
		ChannelID:      channelID,
		Number:         DoorNumberType(number),
		Name:           name,
		KeepOpenTime:   consts.DefaultDoorKeepOpenTime,
		OpenTimeout:    consts.DefaultDoorOpenTimeout,
		LockCount:      consts.DefaultDoorLockCount,
		LockTime:       consts.DefaultDoorLockTime,
		VerifyInterval: consts.DefaultDoorVerifyInterval,
		OpenMode:       consts.DefaultDoorOpenMode,
		FireSignalMode: consts.DefaultDoorFireSignalMode,
	}
}

// TableName 返回驱动层门参数表名
func (DriverDoorParameter) TableName() string {
	return "t_dac_driver_door_parameter"
}

// DirectionType 进出方向类型
type DirectionType int

// EventType 事件类型
type EventType int

// DriverEvent 驱动层事件数据库模型
type DriverEvent struct {
	ID           IDType         `json:"id" gorm:"primaryKey;autoIncrement"`
	ControllerID IDType         `json:"controller_id" gorm:"column:controller_id;uniqueIndex:driver_idx_channel_id_index_timestamp_door_number_type"`
	ChannelID    string         `json:"channel_id" gorm:"column:chid;size:128;uniqueIndex:driver_idx_channel_id_index_timestamp_door_number_type"`
	Index        int            `json:"index" gorm:"column:index;uniqueIndex:driver_idx_channel_id_index_timestamp_door_number_type"`
	Timestamp    int64          `json:"timestamp" gorm:"column:timestamp;uniqueIndex:driver_idx_channel_id_index_timestamp_door_number_type"` // 秒级时间戳
	CardNumber   string         `json:"card_number" gorm:"column:card_number"`
	Username     string         `json:"username" gorm:"column:username"`
	DoorNumber   DoorNumberType `json:"door_number" gorm:"column:door_number;uniqueIndex:driver_idx_channel_id_index_timestamp_door_number_type"`
	Direction    DirectionType  `json:"direction" gorm:"column:direction"`
	Type         EventType      `json:"type" gorm:"column:type;uniqueIndex:driver_idx_channel_id_index_timestamp_door_number_type"`
	Description  string         `json:"description" gorm:"column:description"`
	tgorm.Model
}

// TableName 返回驱动层事件表名
func (DriverEvent) TableName() string {
	return "t_dac_driver_event"
}

// AlarmType 告警类型
type AlarmType int

// AlarmStateType 告警状态类型
type AlarmStateType int

// DriverAlarm 驱动层告警数据库模型
type DriverAlarm struct {
	ID           IDType         `json:"id" gorm:"primaryKey;autoIncrement"`
	ControllerID IDType         `json:"controller_id" gorm:"column:controller_id;uniqueIndex:driver_idx_channel_id_index_timestamp_door_number_type"`
	ChannelID    string         `json:"channel_id" gorm:"column:chid;size:128;uniqueIndex:driver_idx_channel_id_index_timestamp_door_number_type"`
	Index        int            `json:"index" gorm:"column:index;uniqueIndex:driver_idx_channel_id_index_timestamp_door_number_type"`
	Timestamp    int64          `json:"timestamp" gorm:"column:timestamp;uniqueIndex:driver_idx_channel_id_index_timestamp_door_number_type"` // 秒级时间戳
	DoorNumber   DoorNumberType `json:"door_number" gorm:"column:door_number;uniqueIndex:driver_idx_channel_id_index_timestamp_door_number_type"`
	Type         AlarmType      `json:"type" gorm:"column:type;uniqueIndex:driver_idx_channel_id_index_timestamp_door_number_type"`
	State        AlarmStateType `json:"state" gorm:"column:state"`
	Description  string         `json:"description" gorm:"column:channel_id"`
	tgorm.Model
}

// TableName 返回驱动层告警表名
func (DriverAlarm) TableName() string {
	return "t_dac_driver_alarm"
}
