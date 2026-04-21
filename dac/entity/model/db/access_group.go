// Package db 定义门禁系统的数据库模型和表结构。
package db

// AccessGroupBaseInfo 权限组基础信息
type AccessGroupBaseInfo struct {
	ID   IDType `gorm:"primaryKey;index;unique;autoIncrement" json:"id"` // 主键
	Name string `json:"name"`                                            // 权限组名称
}

// AccessGroup 权限组数据库模型
type AccessGroup struct {
	AccessGroupBaseInfo
	Label       string `json:"label"`                         // 标签
	TimeGroupNo int    `json:"time_group_no"`                 // 关联时间组编号
	Comment     string `json:"comment"`                       // 备注
	MozuID      string `json:"mozu_id" gorm:"column:mozu_id"` // 模组ID
}

// TableName 返回权限组表名
func (AccessGroup) TableName() string {
	return "t_dac_access_group"
}

// AccessGroupRelation 权限组与门的关联关系
type AccessGroupRelation struct {
	ID            IDType `gorm:"primaryKey;autoIncrement" json:"id"`                  // 主键
	AccessGroupID IDType `gorm:"index;column:access_group_id" json:"access_group_id"` // 权限组ID
	DoorID        IDType `gorm:"column:door_id" json:"door_id"`                       // 门ID
}

// TableName 返回权限组关联表名
func (AccessGroupRelation) TableName() string {
	return "t_dac_access_group_relation"
}

// AccessGroupInfoWrapper 权限组信息包装器，包含关联的门和卡列表。
// 用于创建或更新权限组时传递完整信息。
type AccessGroupInfoWrapper struct {
	AccessGroup
	Doors []IDType `json:"doors"` // 关联的门ID列表
	Cards []string `json:"cards"` // 关联的卡号列表
}
