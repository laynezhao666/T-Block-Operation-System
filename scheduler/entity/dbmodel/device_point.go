// Package dbmodel 标准测点实体定义
package dbmodel

import (
	"fmt"
	"scheduler/entity/consts"
	"strings"
	"time"
)

// DevicePoint 设备测点表信息
type DevicePoint struct {
	Id            int64     `gorm:"column:id;type:bigint(20);comment:主键ID;primaryKey;not null;" json:"id"`                                         // 主键ID
	DeviceGid     string    `gorm:"column:device_gid;type:varchar(255);comment:设备GID;not null;" json:"device_gid"`                                 // 设备GID
	PointNameEn   string    `gorm:"column:point_name_en;type:varchar(255);comment:测点名称英文;not null;" json:"point_name_en"`                          // 测点名称英文
	Expression    string    `gorm:"column:expression;type:text;comment:测点表达式;" json:"expression"`                                                  // 测点表达式
	ExpressionMap string    `gorm:"column:expression_map;type:text;comment:测点映射(设备编号);" json:"expression_map"`                                     // 测点映射(设备编号)
	MozuId        int32     `gorm:"column:mozu_id;type:int(11);comment:所属模组ID;not null;default:0;" json:"mozu_id"`                                 // 所属模组ID
	UpdateAt      time.Time `gorm:"<-:false;column:update_at;type:datetime;comment:记录上次更新时间;not null;default:CURRENT_TIMESTAMP;" json:"update_at"` // 记录上次更新时间
}

// TableName 设备测点表名称
func (*DevicePoint) TableName() string {
	return "t_device_point"
}

// CalcUniqueKey 设备测点唯一标识
func (p *DevicePoint) CalcUniqueKey() string {
	return strings.Join([]string{p.DeviceGid, p.PointNameEn, fmt.Sprint(p.UpdateAt.UnixMilli())}, consts.CommonFieldSeq)
}
