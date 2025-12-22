package db

import (
	"cmdb/entity/cond"
	"common/entity/consts"
	"common/entity/model"
	"context"
	tgorm "etrpc-go/client/gorm"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"gorm.io/gorm"
)

// IMozuInfoDao 模组信息查询接口
type IMozuInfoDao interface {
	// List 查询模组列表
	List(ctx context.Context, cond *cond.ListMozuInfoCond) ([]*model.MozuInfo, int64, error)
	// Save 插入或更新模组信息
	Save(ctx context.Context, data *model.MozuInfo) error
	// Delete 删除模组信息
	Delete(ctx context.Context, mozuId []int32) error
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

func (m mozuInfoDaoImpl) Save(ctx context.Context, data *model.MozuInfo) error {
	if data.MozuId <= 0 {
		return nil
	}
	existMozuInfo := &model.MozuInfo{}
	// 存在则更新，不存在则插入
	return m.db.WithContext(ctx).Where("mozu_id = ?", data.MozuId).Assign(data).FirstOrCreate(&existMozuInfo).Error
}

func (m mozuInfoDaoImpl) List(ctx context.Context, cond *cond.ListMozuInfoCond) ([]*model.MozuInfo, int64, error) {
	sql := m.db.WithContext(ctx).Model(&model.MozuInfo{})
	// 一些基本的查询条件
	if len(cond.MozuId) > 0 {
		sql = sql.Where("mozu_id in ?", cond.MozuId)
	}
	if len(cond.MozuName) > 0 {
		sql = sql.Where("mozu_name like ?", "%"+cond.MozuName+"%")
	}
	if len(cond.MozuCode) > 0 {
		sql = sql.Where("mozu_code like ?", "%"+cond.MozuCode+"%")
	}
	if len(cond.MozuType) > 0 {
		sql = sql.Where("mozu_type in ?", cond.MozuType)
	}
	if len(cond.BelongCampus) > 0 {
		sql = sql.Where("belong_campus in ?", cond.BelongCampus)
	}
	if len(cond.BelongCampusCode) > 0 {
		sql = sql.Where("belong_campus_code in ?", cond.BelongCampusCode)
	}
	// 处理分页信息
	var total int64
	if cond.Page > 0 && cond.Size > 0 {
		if err := sql.Count(&total).Error; err != nil {
			return nil, 0, errors.Wrap(err, "count total fail")
		}
		sql = sql.Offset(int(cond.Page-1) * int(cond.Size)).Limit(int(cond.Size))
	}
	res := make([]*model.MozuInfo, 0)
	if err := sql.Find(&res).Error; err != nil {
		return nil, 0, errors.Wrap(err, "query db fail")
	}
	return res, lo.Max([]int64{total, int64(len(res))}), nil
}

func (m mozuInfoDaoImpl) Delete(ctx context.Context, mozuId []int32) error {
	sql := m.db.WithContext(ctx).Model(&model.MozuInfo{}).Where("mozu_id in ?", mozuId)
	return sql.Delete(&model.MozuInfo{}).Error
}
