package model

import (
	"fmt"
	"time"
)

const (
	CollectorTypeTbox            = 1 // TBOX采集器
	CollectorTypeTboxSubDevice   = 2 //  TBOX下子设备
	CollectorTypeVendorBox       = 3 // 厂商采集器
	CollectorTypeVendorSubDevice = 4 // 厂商下子设备
)

// CollectorDevice  采集设备
type CollectorDevice struct {
	Id                 int64     `gorm:"column:id;type:bigint(20);comment:主键ID;primaryKey;not null;" json:"id"`                                         // 主键ID
	DeviceGid          string    `gorm:"column:device_gid;type:varchar(255);comment:设备GID;not null;" json:"device_gid"`                                 // 设备GID
	DeviceNumber       string    `gorm:"column:device_number;type:varchar(255);comment:设备编码;not null;" json:"device_number"`                            // 设备编码
	DeviceSn           string    `gorm:"column:device_sn;type:varchar(127);comment:设备SN;not null;" json:"device_sn"`                                    // 设备SN
	DeviceCode         string    `gorm:"column:device_code;type:varchar(127);comment:设备代号;not null;" json:"device_code"`                                // 设备代号
	DeviceName         string    `gorm:"column:device_name;type:varchar(127);comment:设备名称;not null;" json:"device_name"`                                // 设备名称
	DeviceTypeEn       string    `gorm:"column:device_type_en;type:varchar(127);comment:采集设备类型英文;not null;" json:"device_type_en"`                      // 采集设备类型英文
	DeviceTypeZh       string    `gorm:"column:device_type_zh;type:varchar(127);comment:采集设备类型中文;not null;" json:"device_type_zh"`                      // 采集设备类型中文
	CollectorType      int32     `gorm:"column:collector_type;type:tinyint(4);comment:采集类型;not null;default:0;" json:"collector_type"`                  // 采集类型,1:Tbox,2: Tbox下子设备，3：厂商采集器，4：厂商采集器子设备
	ChannelType        string    `gorm:"column:channel_type;type:varchar(16);comment:通道类型;" json:"channel_type"`                                        // 通道类型
	ChannelId          string    `gorm:"column:channel_id;type:varchar(127);comment:通道地址;" json:"channel_id"`                                           // 通道地址
	ChannelLink        string    `gorm:"column:channel_link;type:text;comment:通道详细信息;" json:"channel_link"`                                             // 通道详细信息
	TemplateName       string    `gorm:"column:template_name;type:varchar(127);comment:模版名称;" json:"template_name"`                                     // 模版名称
	TemplateInfo       string    `gorm:"column:template_info;type:varchar(255);comment:模版信息;not null;" json:"template_info"`                            // 模版信息
	ParentDeviceNumber string    `gorm:"column:parent_device_number;type:varchar(255);comment:父级设备编号;" json:"parent_device_number"`                     // 父级设备编号
	MozuId             int32     `gorm:"column:mozu_id;type:int(11);comment:所属模组ID;not null;default:0;" json:"mozu_id"`                                 // 所属模组ID
	Extend             string    `gorm:"column:extend;type:text;comment:扩展字段;" json:"extend"`                                                           // 扩展字段
	CreateAt           time.Time `gorm:"<-:false;column:create_at;type:datetime;comment:记录创建时间;not null;default:CURRENT_TIMESTAMP;" json:"create_at"`   // 记录创建时间
	UpdateAt           time.Time `gorm:"<-:false;column:update_at;type:datetime;comment:记录上次更新时间;not null;default:CURRENT_TIMESTAMP;" json:"update_at"` // 记录上次更新时间
}

// TableName 采集设备表名称
func (CollectorDevice) TableName() string {
	return "t_collector_device"
}

// CalcUniqueKey 采集设备唯一标识
func (t CollectorDevice) CalcUniqueKey() string {
	return fmt.Sprintf("%s|%s", t.DeviceGid, t.TemplateName)
}

// ChannelLink 通道连接信息
type ChannelLink struct {
	ChannelType   string `json:"chtype,omitempty"`
	ChannelId     string `json:"chid,omitempty"`
	ChannelParams string `json:"chparams,omitempty"`
	Addr          string `json:"addr,omitempty"`
	WaitTime      string `json:"waitTime,omitempty"`
	CmdInterval   string `json:"cmdInterval,omitempty"`
	Timeout       string `json:"timeout,omitempty"`
	MaxFailCount  string `json:"maxFailCount,omitempty"`
	MaxFailTime   string `json:"maxFailTime,omitempty"`
}

// TemplateInfo 采集模版信息
type TemplateInfo struct {
	TemplateName string `json:"tplnm"`
	TemplatePath string `json:"tplpath"`
}
