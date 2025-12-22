package model

import (
	"fmt"
	"time"
)

const (
	MarkDownRestoreTemplate = `告警**恢复**通知:
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
	**【恢复时间】**%s
	**【状态】**<font color="green">%s</font>`
)

// AlarmHistory 历史告警
type AlarmHistory struct {
	ID                   int64     `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	AlarmID              int64     `gorm:"column:alarm_id" json:"alarm_id"`
	Level                string    `gorm:"column:level" json:"level"`
	OccurTime            time.Time `gorm:"column:occur_time" json:"occur_time"`
	Rid                  int64     `gorm:"column:rid" json:"rid"`
	MozuId               int64     `gorm:"column:mozu_id" json:"mozu_id"`
	AlarmName            string    `gorm:"column:alarm_name" json:"alarm_name"`
	Content              string    `gorm:"column:content" json:"content"`
	AnalyzeResult        string    `gorm:"column:analyze_result" json:"analyze_result"`
	FingerPrint          string    `gorm:"column:fingerprint" json:"finger_print"`
	DeviceGid            string    `gorm:"column:device_gid" json:"device_gid"`
	DeviceNumber         string    `gorm:"column:device_number" json:"device_number"`
	BoxName              string    `gorm:"column:box_name" json:"box_name"`
	RoomName             string    `gorm:"column:room_name" json:"room_name"`
	RestoreTime          time.Time `gorm:"column:restore_time" json:"restore_time"`
	RestoreAnalyzeResult string    `gorm:"column:restore_analyze_result" json:"restore_analyze_result"`
	CreateAt             time.Time `gorm:"column:create_at" json:"create_at"`
	ActiveCreateAt       time.Time `gorm:"column:active_create_at" json:"active_create_at"`
	DeviceTypeZh         string    `gorm:"column:device_type_zh" json:"device_type_zh"`
	OpUser               string    `gorm:"column:op_user" json:"op_user"`
	OpReason             string    `gorm:"column:op_reason" json:"op_reason"`

	MozuName     string `gorm:"-" json:"mozu_name"`
	DeviceName   string `gorm:"-" json:"device_name"`
	DeviceTypeEn string `gorm:"-" json:"device_type_en"`
}

// TableName 历史告警表名称
func (a *AlarmHistory) TableName() string {
	return "t_alarm_history"
}

// ActiveAlert2History 活动告警转历史告警
func ActiveAlert2History(active *AlarmActive) *AlarmHistory {
	history := &AlarmHistory{
		AlarmID:        int64(active.AlarmID),
		OccurTime:      active.OccurTime,
		Content:        active.Content,
		FingerPrint:    active.FingerPrint,
		AnalyzeResult:  active.AnalyzeResult,
		MozuId:         active.MozuId,
		Rid:            active.Rid,
		DeviceGid:      active.DeviceGid,
		DeviceNumber:   active.DeviceNumber,
		Level:          active.Level,
		RoomName:       active.RoomName,
		BoxName:        active.BoxName,
		AlarmName:      active.AlarmName,
		ActiveCreateAt: active.CreateAt,
		DeviceTypeZh:   active.DeviceTypeZh,
		DeviceName:     active.DeviceName,
		MozuName:       active.MozuName,
		DeviceTypeEn:   active.DeviceTypeEn,
	}
	return history
}

// ConvertToRestoreMsg 转化为Markdown字符串
func (a *AlarmHistory) ConvertToRestoreMsg() string {
	return fmt.Sprintf(MarkDownRestoreTemplate,
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
		a.RestoreTime.Format("2006-01-02 15:04:05"),
		"告警已恢复",
	)
}
