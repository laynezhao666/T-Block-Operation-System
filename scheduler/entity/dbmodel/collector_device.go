// Package dbmodel 采集设备表实体定义
package dbmodel

import (
	"time"
)

// CollectorDevice  采集设备表对应的数据库实体
type CollectorDevice struct {
	Id                 int64     `gorm:"column:id;type:bigint(20);comment:主键ID;primaryKey;not null;" json:"id"`                                         // 主键ID
	DeviceGid          string    `gorm:"column:device_gid;type:varchar(255);comment:设备GID;not null;" json:"device_gid"`                                 // 设备GID
	DeviceNumber       string    `gorm:"column:device_number;type:varchar(255);comment:设备编码;not null;" json:"device_number"`                            // 设备编码
	CollectorType      int32     `gorm:"column:device_type;type:tinyint(4);not null;default:0;" json:"device_type"`                                     // 设备类型,1:Tbox,2: Tbox下子设备，3：厂商采集器，4：厂商采集器子设备
	ChannelType        string    `gorm:"column:channel_type;type:varchar(16);comment:通道类型;default:NULL;" json:"channel_type"`                           // 通道类型
	ChannelId          string    `gorm:"column:channel_id;type:varchar(127);comment:通道地址;default:NULL;" json:"channel_id"`                              // 通道地址
	ChannelLink        string    `gorm:"column:channel_link;type:text;comment:通道详细信息;" json:"channel_link"`                                             // 通道详细信息
	Profile            string    `gorm:"column:profile;type:text;comment:模版详细配置;" json:"profile"`                                                       // 模版详细配置
	ActiveStatus       int32     `gorm:"column:active_status;type:tinyint(4);comment:激活状态;default:NULL;" json:"active_status"`                          // 激活状态
	TemplateName       string    `gorm:"column:template_name;type:varchar(127);comment:模版名称;default:NULL;" json:"template_name"`                        // 模版名称
	ParentDeviceNumber string    `gorm:"column:parent_device_number;type:varchar(255);comment:父级设备编号;default:NULL;" json:"parent_device_number"`        // 父级设备编号
	MozuId             int32     `gorm:"column:mozu_id;type:int(11);comment:所属模组ID;not null;default:0;" json:"mozu_id"`                                 // 所属模组ID
	ComputeCost        int64     `gorm:"-"`                                                                                                             // 涉及的测点数量,采集测点+标准测点
	UpdateAt           time.Time `gorm:"<-:false;column:update_at;type:datetime;comment:记录上次更新时间;not null;default:CURRENT_TIMESTAMP;" json:"update_at"` // 记录上次更新时间
}

// TableName 采集配置表名称
func (c *CollectorDevice) TableName() string {
	return "t_collector_device"
}

// CalcUniqueKey 采集器唯一标识
func (c *CollectorDevice) CalcUniqueKey() string {
	return c.DeviceNumber
}
