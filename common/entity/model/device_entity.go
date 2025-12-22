package model

import (
	"fmt"
	"time"
)

// DeviceEntity 设备实体
type DeviceEntity struct {
	Id                      int64     `gorm:"column:id;type:bigint(20);comment:主键ID;primaryKey;not null;" json:"id"`                                          // 主键ID
	DeviceGid               string    `gorm:"column:device_gid;type:varchar(127);comment:设备GID;not null;" json:"device_gid"`                                  // 设备GID
	DeviceNumber            string    `gorm:"column:device_number;type:varchar(255);comment:设备编码;not null;" json:"device_number"`                             // 设备编码
	DeviceNumberRoute       string    `gorm:"column:device_number_route;type:varchar(255);comment:路由编码;not null;" json:"device_number_route"`                 // 路由编码
	DeviceNumberShow        string    `gorm:"column:device_number_show;type:varchar(255);comment:展示编码;not null;" json:"device_number_show"`                   // 展示编码
	DeviceName              string    `gorm:"column:device_name;type:varchar(127);comment:设备名称;not null;" json:"device_name"`                                 // 设备名称
	MozuId                  int32     `gorm:"column:mozu_id;type:int(11);comment:所属模组ID;not null;default:0;" json:"mozu_id"`                                  // 所属模组ID
	MozuName                string    `gorm:"column:mozu_name;type:varchar(64);comment:模组名称;not null;" json:"mozu_name"`                                      // 模组名称
	IdcArea                 string    `gorm:"column:idc_area;type:varchar(64);comment:机房区域;not null;" json:"idc_area"`                                        // 机房区域
	FuncRoom                string    `gorm:"column:func_room;type:varchar(64);comment:方仓/功能间;not null;" json:"func_room"`                                    // 方仓/功能间
	ParentDeviceNumber      string    `gorm:"column:parent_device_number;type:varchar(255);comment:父级设备编码;not null;" json:"parent_device_number"`             // 父级设备编码
	DeviceTypeEn            string    `gorm:"column:device_type_en;type:varchar(127);comment:设备种类英文;not null;" json:"device_type_en"`                         // 设备种类英文
	DeviceTypeZh            string    `gorm:"column:device_type_zh;type:varchar(127);comment:设备种类中文;not null;" json:"device_type_zh"`                         // 设备种类中文
	ApplicationTypeEn       string    `gorm:"column:application_type_en;type:varchar(127);comment:应用类型英文;not null;" json:"application_type_en"`               // 应用类型英文
	ApplicationTypeZh       string    `gorm:"column:application_type_zh;type:varchar(127);comment:应用类型中文;not null;" json:"application_type_zh"`               // 应用类型中文
	BelongApplicationTypeEn string    `gorm:"column:belong_application_type_en;type:varchar(127);comment:所属应用类型;not null;" json:"belong_application_type_en"` // 所属应用类型
	CreateAt                time.Time `gorm:"<-:false;column:create_at;type:datetime;comment:记录创建时间;not null;default:CURRENT_TIMESTAMP;" json:"create_at"`    // 记录创建时间
	UpdateAt                time.Time `gorm:"<-:false;column:update_at;type:datetime;comment:记录上次更新时间;not null;default:CURRENT_TIMESTAMP;" json:"update_at"`  // 记录上次更新时间
}

// TableName 设备实体表名称
func (d *DeviceEntity) TableName() string {
	return "t_device_entity"
}

// CalcUniqueKey 设备实体唯一标识
func (d *DeviceEntity) CalcUniqueKey() string {
	return fmt.Sprintf("%s.%d", d.DeviceGid, d.MozuId)
}
