// Package db 定义门禁系统的数据库模型和表结构。
package db

import (
	"fmt"

	"dac/entity/consts"
)

// 门相关默认值常量
const (
	DefaultGroupID  = IDType(-1) // 默认门组ID（未分组）
	DoorCollectType = "GSM"      // 门采集类型标识
)

// DoorParameter 门参数配置
type DoorParameter struct {
	Name           string `json:"name"`             // 门名称
	Password       string `json:"password"`         // 门密码
	KeepOpenTime   int    `json:"keep_open_time"`   // 门开保持时间（秒）
	OpenTimeout    int    `json:"open_timeout"`     // 门开超时时间（秒）
	LockCount      int    `json:"lock_count"`       // 非法卡允许的最长失败次数
	LockTime       int    `json:"lock_time"`        // 非法卡的封锁时间（秒）
	VerifyInterval int    `json:"verify_interval"`  // 非法卡刷卡间隔（秒）
	OpenMode       int    `json:"open_mode"`        // 开门模式
	FireSignalMode int    `json:"fire_signal_mode"` // 火警信号模式
}

// DoorBaseInfo 门基本信息
type DoorBaseInfo struct {
	ID   IDType `json:"id" gorm:"primaryKey;autoIncrement"` // 门ID
	Name string `json:"name" `                              // 门名称
}

// Door 门完整信息
type Door struct {
	DoorBaseInfo
	Number       int                    `json:"number" gorm:"uniqueIndex:door_index"`
	GroupID      IDType                 `json:"group_id" gorm:"column:group_id"`
	IDCDBCode    string                 `json:"code" gorm:"column:code"`
	GID          GIDType                `json:"gid" gorm:"column:gid"`
	ControllerID IDType                 `json:"controller_id" gorm:"column:controller_id;uniqueIndex:door_index"`
	Parameters   DoorParameter          `json:"parameters" gorm:"column:parameters;serializer:json"`
	Extend       map[string]interface{} `json:"extend"  gorm:"column:extend;serializer:json"`
}

// NewDoor 创建门实例（使用默认参数）
func NewDoor(controllerID IDType, number int, name string, idcdbCode string) Door {
	parm := DoorParameter{
		Name:           name,
		KeepOpenTime:   consts.DefaultDoorKeepOpenTime,
		OpenTimeout:    consts.DefaultDoorOpenTimeout,
		LockCount:      consts.DefaultDoorLockCount,
		LockTime:       consts.DefaultDoorLockTime,
		VerifyInterval: consts.DefaultDoorVerifyInterval,
		OpenMode:       consts.DefaultDoorOpenMode,
		FireSignalMode: consts.DefaultDoorFireSignalMode,
	}

	return Door{
		DoorBaseInfo: DoorBaseInfo{
			Name: name,
		},
		Number:       number,
		GroupID:      DefaultGroupID,
		IDCDBCode:    idcdbCode,
		ControllerID: controllerID,
		Parameters:   parm,
		Extend:       make(map[string]interface{}),
	}
}

// GetIDCDBCode 获取门的IDCDB编号
func (d *Door) GetIDCDBCode() string {
	if d == nil {
		return ""
	}
	return d.IDCDBCode
}

// GetCollectCode 获取门的采集编码
func (d *Door) GetCollectCode(controllerCode string) string {
	if d == nil {
		return ""
	}
	return fmt.Sprintf("%v.%v_%v", controllerCode, DoorCollectType, d.Number)
}

// SetIDCDBCode 设置门的IDCDB编号
func (d *Door) SetIDCDBCode(idcdbCode string) {
	if d == nil {
		return
	}

	d.IDCDBCode = idcdbCode
}

// GetName 获取门名称（优先使用Name字段，为空则使用参数中的名称）
func (d *Door) GetName() string {
	if d == nil {
		return ""
	}
	if len(d.Name) == 0 {
		return d.Parameters.Name
	}
	return d.Name
}

// SetName 设置门名称
func (d *Door) SetName(name string) {
	if d == nil {
		return
	}
	d.Name = name
}

// TableName 返回门信息表名
func (*Door) TableName() string {
	return "t_dac_door"
}

// DoorGroup 门组信息
type DoorGroup struct {
	ID     IDType `json:"id" gorm:"primaryKey;autoIncrement"`
	Name   string `json:"name" gorm:"index:idx_name_mozu_id,unique"`
	MozuID string `json:"mozu_id" gorm:"column:mozu_id;type:varchar(10);index:idx_name_mozu_id,unique"`
}

// TableName 返回门组表名
func (DoorGroup) TableName() string {
	return "t_dac_door_group"
}
