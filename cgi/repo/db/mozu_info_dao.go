// Package db 存放模组相关信息查询接口
package db

import (
	"cgi/entity/dto"
	"common/entity/consts"
	"common/entity/model"
	"context"

	tgorm "etrpc-go/client/gorm"

	"gorm.io/gorm"
)

// IMozuInfoDao 模组信息查询接口
type IMozuInfoDao interface {
	GetList(ctx context.Context, cond *dto.CondMozuInfoGetList) ([]*model.MozuInfo, error)
}

// NewMozuInfoDao 创建模组信息查询接口实现对象
func NewMozuInfoDao() IMozuInfoDao {
	return &mozuInfoDaoImpl{
		db: tgorm.GetDB(consts.TbosMysqlName),
	}
}

type mozuInfoDaoImpl struct {
	db *gorm.DB
}

// GetList 获取模组List
func (m mozuInfoDaoImpl) GetList(ctx context.Context, cond *dto.CondMozuInfoGetList) ([]*model.MozuInfo, error) {
	sql := m.db.WithContext(ctx).Model(&model.MozuInfo{})
	if len(cond.MozuId) > 0 {
		sql.Where("mozu_id in ?", cond.MozuId)
	}
	if len(cond.MozuName) > 0 {
		sql.Where("mozu_name in ?", cond.MozuName)
	}
	if len(cond.MozuCode) > 0 {
		sql.Where("mozu_code in ?", cond.MozuCode)
	}
	res := make([]*model.MozuInfo, 0)
	if err := sql.Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}
