package model

import (
	"fmt"
	"time"
)

const (
	MarkDownAlarmTemplate = `**告警通知**:
	**【模组Id】**%d
	**【模组名称】**%s
	**【告警名称】**%s
	**【告警级别】**%s
	**【方舱名称】**%s
	**【房间名称】**%s
	**【设备Gid】**%s
	**【设备编号】**%s
	**【设备名称】**%s
	**【告警内容】**%s
	**【产生时间】**%s
	**【状态】**<font color="warning">%s</font>`

	// 告警状态 0:未挂起活动告警 1:挂起告警
	ActiveAlarmCode = 0
	HangupAlarmCode = 1
)

// AlarmActive 活动告警
type AlarmActive struct {
	ID            int64     `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	AlarmID       int64     `gorm:"column:alarm_id" json:"alarm_id"`
	OccurTime     time.Time `gorm:"column:occur_time" json:"occur_time"`
	Level         string    `gorm:"column:level" json:"level"`
	MozuId        int64     `gorm:"column:mozu_id" json:"mozu_id"`
	Rid           int64     `gorm:"column:rid" json:"rid"`
	AlarmName     string    `gorm:"column:alarm_name" json:"alarm_name"`
	Content       string    `gorm:"column:content" json:"content"`
	FingerPrint   string    `gorm:"column:fingerprint" json:"finger_print"`
	AnalyzeResult string    `gorm:"column:analyze_result" json:"analyze_result"`
	DeviceGid     string    `gorm:"column:device_gid" json:"device_gid"`
	DeviceNumber  string    `gorm:"column:device_number" json:"device_number"`
	BoxName       string    `gorm:"column:box_name" json:"box_name"`
	RoomName      string    `gorm:"column:room_name" json:"room_name"`
	CreateAt      time.Time `gorm:"column:create_at" json:"create_at"`
	UpdateTime    time.Time `gorm:"column:update_time" json:"update_time"`
	DeviceTypeZh  string    `gorm:"column:device_type_zh" json:"device_type_zh"`
	Status        int64     `gorm:"column:status" json:"status"`
	EventStatus   int64     `gorm:"column:event_status" json:"event_status"`
	OpUser        string    `gorm:"column:op_user" json:"op_user"`
	OpReason      string    `gorm:"column:op_reason" json:"op_reason"`

	DeviceName   string `gorm:"-" json:"device_name"`
	MozuName     string `gorm:"-" json:"mozu_name"`
	DeviceTypeEn string `gorm:"-" json:"device_type_en"`
}

// TableName 活动告警表名称
func (a *AlarmActive) TableName() string {
	return "t_alarm_active"
}

// AlarmCnt ...
type AlarmCnt struct {
	Cnt int64 `gorm:"column:cnt" json:"cnt"`
}

// AlarmStatisticsGroup ...
type AlarmStatisticsGroup struct {
	Name string `gorm:"column:name" json:"name"`
	Cnt  int64  `gorm:"column:cnt" json:"cnt"`
}

// ConvertToAlarmMsg 转化为Markdown字符串
func (a *AlarmActive) ConvertToAlarmMsg() string {
	return fmt.Sprintf(MarkDownAlarmTemplate,
		a.MozuId,
		a.MozuName,
		a.AlarmName,
		a.Level,
		a.BoxName,
		a.RoomName,
		a.DeviceGid,
		a.DeviceNumber,
		a.DeviceName,
		a.Content,
		a.OccurTime.Format("2006-01-02 15:04:05"),
		"告警持续中",
	)
}
