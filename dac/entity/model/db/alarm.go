// Package db 定义门禁系统的数据库模型和表结构。
package db

// AlarmKey 告警记录的唯一标识组合键
type AlarmKey struct {
	ControllerID IDType // 控制器ID
	Timestamp    int64  // 时间戳
	DoorNumber   int    // 门编号
	Type         int    // 告警类型
	State        int    // 告警状态
}

// Alarm 告警记录数据库模型
type Alarm struct {
	ID             int64  `gorm:"primaryKey;autoIncrement"`
	ControllerID   IDType `gorm:"column:controller_id;index:controller_index;uniqueIndex:alarm_index"`
	ControllerName string `gorm:"column:controller_name"`
	Index          int    `gorm:"column:index;uniqueIndex:alarm_index"`
	Timestamp      int64  `gorm:"column:timestamp;index:timestamp_index;uniqueIndex:alarm_index"`
	DoorName       string `gorm:"column:door_name"`
	DoorNumber     int    `gorm:"column:door_number"`
	Type           int    `gorm:"column:type"`
	State          int    `gorm:"column:state"`
	StateDesc      int    `gorm:"column:state_desc"`
	Description    string `gorm:"column:description"`
	MozuID         string `json:"mozu_id" gorm:"column:mozu_id"`
}

// TableName 返回告警记录表名
func (*Alarm) TableName() string {
	return "t_dac_alarm"
}

// GetKey 获取告警记录的唯一标识键
func (a *Alarm) GetKey() AlarmKey {
	return AlarmKey{
		ControllerID: a.ControllerID,
		Timestamp:    a.Timestamp,
		DoorNumber:   a.DoorNumber,
		Type:         a.Type,
		State:        a.State,
	}
}

// AlarmIndexRecord 告警索引记录，用于增量拉取
type AlarmIndexRecord struct {
	ControllerID IDType `gorm:"primaryKey;column:controller_id"`
	Index        int    `gorm:"column:index"`
	Last         int    `gorm:"column:last"`
	UpdateTime   int64  `gorm:"column:update_time"`
}

// TableName 返回告警索引表名
func (*AlarmIndexRecord) TableName() string {
	return "t_dac_alarm_index"
}

// SetIndex 设置当前索引位置
func (a *AlarmIndexRecord) SetIndex(index int) {
	a.Index = index
}

// SetLast 设置最后索引位置
func (a *AlarmIndexRecord) SetLast(last int) {
	a.Last = last
}

// SetControllerID 设置控制器ID
func (a *AlarmIndexRecord) SetControllerID(controllerID IDType) {
	a.ControllerID = controllerID
}

// AlarmTimestampIndexRecord 基于时间戳的告警索引记录
type AlarmTimestampIndexRecord struct {
	TimestampIndexRecord
}

// TableName 返回告警时间戳索引表名
func (*AlarmTimestampIndexRecord) TableName() string {
	return "t_dac_alarm_timestamp_index"
}
