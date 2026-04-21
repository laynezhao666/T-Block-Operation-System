package dac

import (
	"gorm.io/gorm"
)

// impl 是 RW 接口的实现，封装了 GORM 数据库连接。
type impl struct {
	db *gorm.DB
}
