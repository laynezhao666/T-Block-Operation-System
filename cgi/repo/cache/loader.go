// Package cache 数据加载器接口定义及一些默认的数据加载器
package cache

import (
	"gorm.io/gorm"
)

// IDataLoader 数据加载器接口,根据模组ID加载数据
type IDataLoader[T any] interface {
	Load(mozuId int32) (T, error)
}

type dbTableLoader[R any, T []R] struct {
	db *gorm.DB
}

// NewDbTableLoader 创建一个从数据库加载数据的加载器
func NewDbTableLoader[R any](db *gorm.DB) IDataLoader[[]R] {
	return &dbTableLoader[R, []R]{
		db: db,
	}
}

func (d dbTableLoader[R, T]) Load(mozuId int32) (T, error) {
	all := make([]R, 0)
	res := make([]R, 0)
	if err := d.db.Where("mozu_id = ?", mozuId).FindInBatches(&res, 30000,
		func(tx *gorm.DB, batch int) error {
			all = append(all, res...)
			return nil
		}).Error; err != nil {
		return all, err
	}
	return all, nil
}
