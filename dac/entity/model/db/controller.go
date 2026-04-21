// Package db 定义门禁系统的数据库模型和表结构。
package db

import (
	"dac/entity/consts"

	"dac/entity/model/tbox"
)

// DoorController 门禁控制器数据库模型
type DoorController struct {
	ID        IDType  `json:"id" gorm:"primaryKey;autoIncrement"` // 控制器ID
	IDCDBCode string  `json:"-" gorm:"column:code"`               // IDCDB编号
	GID       GIDType `json:"gid" gorm:"column:gid"`              // GID编号
	Name      string  `json:"name"`                               // 门禁名称

	Version int64 `json:"version"` // 版本号

	MozuID string `json:"mozu_id" gorm:"column:mozu_id;type:varchar(64);index"` // 模组ID

	Profile  Profile         `json:"profile" gorm:"serializer:json"`  // 设备信息
	Position tbox.Position   `json:"position" gorm:"serializer:json"` // 位置信息
	Channel  tbox.ChannelRaw `json:"channel" gorm:"serializer:json"`  // 通道信息
	Protocol Protocol        `json:"protocol" gorm:"serializer:json"` // 协议信息

	// Extend 扩展参数，包含key字段（调用接口的密钥，值为md5(account||password)）
	Extend map[string]interface{} `json:"extend" gorm:"serializer:json"`

	Account  string `json:"account"`  // 账号
	Password string `json:"password"` // 密码
}

// TableName 返回门禁控制器表名
func (*DoorController) TableName() string {
	return "t_dac_controller"
}

// GetCollectCode 获取控制器的采集编码
func (d *DoorController) GetCollectCode() string {
	return d.Name
}

// IsMDC 判断控制器是否为MDC版本
func (d *DoorController) IsMDC() bool {
	if d == nil {
		return false
	}
	return d.Protocol.Name == consts.ProtocolHTTP &&
		d.Protocol.Version == consts.MDCProtocolVersion
}
