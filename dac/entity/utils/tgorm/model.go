// Package tgorm 提供GORM数据库模型的通用基础结构。
package tgorm

// Model 通用数据库模型基类，包含创建和更新时间戳
type Model struct {
	UpdatedAt int64 // 更新时间（Unix时间戳）
	CreatedAt int64 // 创建时间（Unix时间戳）
}
