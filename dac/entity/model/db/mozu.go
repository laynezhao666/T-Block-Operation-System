// Package db 定义门禁系统的数据库模型和表结构。
package db

// Mozu 模组信息（关联楼栋）
type Mozu struct {
	ID          int    `gorm:"column:mozu_id"`      // 模组ID
	Name        string `gorm:"column:mozu_name"`    // 模组名称
	BuildingMID int    `gorm:"column:building_mid"` // 楼栋MID
}

// TableName 返回模组信息表名
func (*Mozu) TableName() string {
	return "datamodel_mozu_info"
}
