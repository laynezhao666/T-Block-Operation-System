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

// IDevicePointDao 设备测点表相关操作接口
type IDevicePointDao interface {
	// BatchUpdate 批量更新某个模组的设备测点信息
	BatchUpdate(ctx context.Context, mozuId int32, data []*model.DevicePoint) error
	//GetList 条件查询测点
	GetList(ctx context.Context, cond *cond.ListDevicePointCond) ([]*model.DevicePoint, int64, error)
	// GetDeviceNumberByCollector 根据采集器查询相关标准设备编号列表
	GetDeviceNumberByCollector(ctx context.Context, collectorNumber string) ([]string, error)
	// GetCollectorDataVer 获取每个采集器下标准点的最新数据版本
	GetCollectorDataVer(ctx context.Context, deviceNumber []string) (map[string]int64, error)
}

// NewDevicePointDao 创建设备测点表相关操作类实例
func NewDevicePointDao() IDevicePointDao {
	return &devicePointDaoImpl{
		db: tgorm.GetDB(consts.TbosMysqlName),
	}
}

// devicePointDaoImpl 设备测点相关操作具体实现
type devicePointDaoImpl struct {
	db *gorm.DB
}

func (t *devicePointDaoImpl) BatchUpdate(ctx context.Context, mozuId int32, data []*model.DevicePoint) error {
	// 1、按模组查找DB出所有的设备测点
	dbDevicePoints, _, err := t.GetList(ctx, &cond.ListDevicePointCond{MozuId: []int32{mozuId}})
	if err != nil {
		return errors.Wrapf(err, "query exist device point failed, mozuId: %d", mozuId)
	}
	// 2、查找需要新增和删除的设备测点
	addList, delList, _ := collutil.FindDiff(dbDevicePoints, data, func(item *model.DevicePoint) string {
		return item.CalcUniqueKey()
	}, true)
	delIdList := arrayutil.Map(delList, func(t *model.DevicePoint) int64 {
		return t.Id
	})
	// 3、执行更新操作，采用事务执行
	if err := TransactionUpdate(t.db.WithContext(ctx), addList, delIdList, "device point"); err != nil {
		return errors.Wrapf(err, "transaction update device point failed")
	}
	return nil
}

func (t *devicePointDaoImpl) GetList(ctx context.Context, cond *cond.ListDevicePointCond) ([]*model.DevicePoint, int64, error) {
	sql := t.db.WithContext(ctx).Model(&model.DevicePoint{})
	if len(cond.DeviceGid) > 0 {
		sql = sql.Where("device_gid in ?", cond.DeviceGid)
	}
	if len(cond.DeviceNumber) > 0 {
		sql = sql.Where("device_number in ?", cond.DeviceNumber)
	}
	if len(cond.BelongCollector) > 0 {
		sql = sql.Where("belong_collector in ?", cond.BelongCollector)
	}
	if len(cond.PointNameEn) > 0 {
		sql = sql.Where("point_name_en in ?", cond.PointNameEn)
	}
	if len(cond.PointNameZh) > 0 {
		sql = sql.Where("point_name_zh in ?", cond.PointNameZh)
	}
	if len(cond.PointKey) > 0 {
		sql = sql.Where("point_key in ?", cond.PointKey)
	}
	if len(cond.PointType) > 0 {
		sql = sql.Where("point_type in ?", cond.PointType)
	}
	if len(cond.PointRw) > 0 {
		sql = sql.Where("point_rw in ?", cond.PointRw)
	}
	if len(cond.MozuId) > 0 {
		sql = sql.Where("mozu_id in ?", cond.MozuId)
	}
	if len(cond.PointLevel) > 0 {
		sql = sql.Where("point_level in ?", cond.PointLevel)
	}
	var total int64
	if cond.Page > 0 && cond.Size > 0 {
		if err := sql.Count(&total).Error; err != nil {
			return nil, 0, errors.Wrapf(err, "count device point fail")
		}
		sql.Offset(int((cond.Page - 1) * cond.Size)).Limit(int(cond.Size))
	}
	res := make([]*model.DevicePoint, 0)
	if err := sql.Find(&res).Error; err != nil {
		return nil, 0, err
	}
	return res, lo.Max([]int64{total, int64(len(res))}), nil
}

func (t *devicePointDaoImpl) GetDeviceNumberByCollector(ctx context.Context, belongCollector string) ([]string, error) {
	sql := t.db.WithContext(ctx).Model(&model.DevicePoint{}).
		Select("distinct device_number").
		Where("belong_collector = ?", belongCollector)
	res := make([]string, 0)
	if err := sql.Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (t *devicePointDaoImpl) GetCollectorDataVer(ctx context.Context, deviceNumber []string) (map[string]int64, error) {
	sql := t.db.WithContext(ctx).Model(&model.DevicePoint{}).
		Select("belong_collector", "max(update_at) as update_at").
		Where("belong_collector != '' ").
		Group("belong_collector")
	if len(deviceNumber) > 0 {
		sql.Where("belong_collector in ? ", deviceNumber)
	}
	res := make([]*model.DevicePoint, 0)
	if err := sql.Find(&res).Error; err != nil {
		return nil, err
	}
	return lo.SliceToMap(res, func(item *model.DevicePoint) (string, int64) {
		return item.BelongCollector, item.UpdateAt.Unix()
	}), nil
}
