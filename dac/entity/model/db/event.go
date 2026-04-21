// Package db 定义门禁系统的数据库模型和表结构。
package db

// EventKey 事件记录的唯一标识组合键
type EventKey struct {
	ControllerID IDType // 控制器ID
	Timestamp    int64  // 时间戳
	CardNumber   string // 卡号
	DoorNumber   int    // 门编号
	Direction    int    // 进出方向
	Type         int    // 事件类型
}

// Event 事件记录数据库模型
type Event struct {
	ID             int64  `gorm:"primaryKey;autoIncrement"`
	ControllerID   IDType `gorm:"column:controller_id;index:controller_index;uniqueIndex:event_index"`
	ControllerName string `gorm:"column:controller_name"`
	Index          int    `gorm:"column:index;uniqueIndex:event_index"`
	Timestamp      int64  `gorm:"column:timestamp;uniqueIndex:event_index;index:timestamp_index"` // 秒级时间戳
	CardNumber     string `gorm:"index;column:card_number"`
	Username       string `gorm:"index;column:username"`
	Company        string
	DoorName       string `gorm:"column:door_name"`
	DoorNumber     int    `gorm:"column:door_number"`
	Direction      int
	Type           int
	Description    string
	MozuID         string `json:"mozu_id" gorm:"column:mozu_id"`
}

// TableName 返回事件记录表名
func (*Event) TableName() string {
	return "t_dac_event"
}

// GetKey 获取事件记录的唯一标识键
func (e *Event) GetKey() EventKey {
	return EventKey{
		ControllerID: e.ControllerID,
		Timestamp:    e.Timestamp,
		CardNumber:   e.CardNumber,
		DoorNumber:   e.DoorNumber,
		Direction:    e.Direction,
		Type:         e.Type,
	}
}

// EventIndexRecord 事件索引记录，用于增量拉取
type EventIndexRecord struct {
	ControllerID IDType `gorm:"primaryKey;column:controller_id"` // 控制器ID
	Index        int    `gorm:"column:index"`                    // 当前索引
	Last         int    `gorm:"column:last"`                     // 最后索引
	UpdateTime   int64  `gorm:"column:update_time"`              // 更新时间
}

// TableName 返回事件索引表名
func (*EventIndexRecord) TableName() string {
	return "t_dac_event_index"
}

// SetIndex 设置当前索引位置
func (e *EventIndexRecord) SetIndex(index int) {
	e.Index = index
}

// SetLast 设置最后索引位置
func (e *EventIndexRecord) SetLast(last int) {
	e.Last = last
}

// SetControllerID 设置控制器ID
func (e *EventIndexRecord) SetControllerID(controllerID IDType) {
	e.ControllerID = controllerID
}

// EventTimestampIndexRecord 基于时间戳的事件索引记录
type EventTimestampIndexRecord struct {
	TimestampIndexRecord
}

// TableName 返回事件时间戳索引表名
func (*EventTimestampIndexRecord) TableName() string {
	return "t_dac_event_timestamp_index"
}
