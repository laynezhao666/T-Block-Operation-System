// Package tgorm 提供GORM数据库操作的工具函数。
package tgorm

import (
	"gorm.io/gorm"
)

// DBType GORM数据库类型别名
type DBType = gorm.DB

// HandlerFunc 事务处理函数类型
type HandlerFunc func(tx *DBType, args ...interface{}) error

// DoTransaction 在数据库事务中执行handler函数
func DoTransaction(db *DBType, handler HandlerFunc, args ...interface{}) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return handler(tx, args...)
	})
}
