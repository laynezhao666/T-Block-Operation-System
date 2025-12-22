package db

import (
	"cmdb/entity/cond"
	"cmdb/util/collutil"
	"common/entity/consts"
	"common/entity/model"
	"context"
	tgorm "etrpc-go/client/gorm"
	"etrpc-go/util/arrayutil"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

// IDeviceEntityDao 设备实体表操作接口
type IDeviceEntityDao interface {
	// BatchUpdate 批量更新设备实体
	BatchUpdate(ctx context.Context, mozuId int32, values []*model.DeviceEntity) error
	// GetList 获取设备实体列表
	GetList(ctx context.Context, cond *cond.ListDeviceEntityCond) ([]*model.DeviceEntity, int64, error)
}

// NewDeviceEntityDao 获取设备设备实体表操作类对象
func NewDeviceEntityDao() IDeviceEntityDao {
	return &deviceEntityDaoImpl{
		db: tgorm.GetDB(consts.TbosMysqlName),
	}
}

// deviceEntityDaoImpl 设备实体表操作实现
type deviceEntityDaoImpl struct {
	db *gorm.DB
}

// BatchUpdate 批量更新设备实体
func (t *deviceEntityDaoImpl) BatchUpdate(ctx context.Context, mozuId int32, data []*model.DeviceEntity) error {
	// 1、按模组查找DB出所有的设备实体
	dbDeviceEntities, _, err := t.GetList(ctx, &cond.ListDeviceEntityCond{MozuId: []int32{mozuId}})
	if err != nil {
		return errors.Wrapf(err, "query exist device entity failed, mozuId: %d", mozuId)
	}
	// 2、查找需要新增和删除的设备实体
	addList, delList, _ := collutil.FindDiff(dbDeviceEntities, data, func(item *model.DeviceEntity) string {
		return item.CalcUniqueKey()
	}, true)
	delIdList := arrayutil.Map(delList, func(t *model.DeviceEntity) int64 {
		return t.Id
	})
	// 3、执行更新操作，采用事务执行
	if err := TransactionUpdate(t.db.WithContext(ctx), addList, delIdList, "device entity"); err != nil {
		return errors.Wrapf(err, "transaction update device entity failed")
	}
	return nil
}

// GetList 获取设备实体列表
func (t *deviceEntityDaoImpl) GetList(ctx context.Context, cond *cond.ListDeviceEntityCond) ([]*model.DeviceEntity, int64, error) {
	sql := t.db.WithContext(ctx).Model(&model.DeviceEntity{})
	// 一些基本的查询条件
	if len(cond.DeviceGid) > 0 {
		sql = sql.Where("device_gid in ?", cond.DeviceGid)
	}
	if len(cond.DeviceNumber) > 0 {
		sql = sql.Where("device_number in ?", cond.DeviceNumber)
	}
	if len(cond.MozuId) > 0 {
		sql = sql.Where("mozu_id in ?", cond.MozuId)
	}
	if len(cond.EnableStatus) > 0 {
		sql = sql.Where("enable_status in ?", cond.EnableStatus)
	}
	if len(cond.ParentDeviceNumber) > 0 {
		sql = sql.Where("parent_device_number in ?", cond.ParentDeviceNumber)
	}
	if len(cond.DeviceTypeEn) > 0 {
		sql = sql.Where("device_type_en in ?", cond.DeviceTypeEn)
	}
	if len(cond.DeviceTypeZh) > 0 {
		sql = sql.Where("device_type_zh in ?", cond.DeviceTypeZh)
	}
	if len(cond.ApplicationTypeEn) > 0 {
		sql = sql.Where("application_type_en in ?", cond.ApplicationTypeEn)
	}
	if len(cond.ApplicationTypeZh) > 0 {
		sql = sql.Where("application_type_zh in ?", cond.ApplicationTypeZh)
	}
	if len(cond.BelongApplicationTypeEn) > 0 {
		sql = sql.Where("belong_application_type_en in ?", cond.BelongApplicationTypeEn)
	}
	// 处理分页信息
	var total int64
	if cond.Page > 0 && cond.Size > 0 {
		if err := sql.Count(&total).Error; err != nil {
			return nil, 0, errors.Wrapf(err, "count device entity fail")
		}
		sql.Offset(int((cond.Page - 1) * cond.Size)).Limit(int(cond.Size))
	}
	res := make([]*model.DeviceEntity, 0)
	if err := sql.Find(&res).Error; err != nil {
		return nil, 0, err
	}
	return res, lo.Max([]int64{total, int64(len(res))}), nil
}
