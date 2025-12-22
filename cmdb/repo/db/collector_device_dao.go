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

// ICollectorDeviceDao 设备采集策略表相关操作接口
type ICollectorDeviceDao interface {
	// BatchUpdate 批量更新某个模组的采集设备信息
	BatchUpdate(ctx context.Context, mozuId int32, data []*model.CollectorDevice) error
	// GetList 获取采集设备列表
	GetList(ctx context.Context, cond *cond.ListCollectorDeviceCond) ([]*model.CollectorDevice, int64, error)
	// GetCollectorDataVer 获取Collector的数据版本
	GetCollectorDataVer(ctx context.Context, deviceNumber []string) (map[string]int64, error)
}

type collectorDeviceDaoImpl struct {
	db *gorm.DB
}

// NewCollectorDeviceDao 创建采集策略表相关操作实现类对象
func NewCollectorDeviceDao() ICollectorDeviceDao {
	return &collectorDeviceDaoImpl{
		db: tgorm.GetDB(consts.TbosMysqlName),
	}
}

func (t collectorDeviceDaoImpl) BatchUpdate(ctx context.Context, mozuId int32, data []*model.CollectorDevice) error {
	// 1、按模组查询出DB中原来所有的采集策略配置
	dbCollectorDevices, _, err := t.GetList(ctx, &cond.ListCollectorDeviceCond{MozuId: []int32{mozuId}})
	if err != nil {
		return errors.Wrapf(err, "query exist collector device of mozu_id [%d] fail", mozuId)
	}
	// 2、查找原来的采集模版列表和新从采集模版列表之间的差异
	addList, delList, _ := collutil.FindDiff(dbCollectorDevices, data, func(item *model.CollectorDevice) string {
		return item.CalcUniqueKey()
	}, true)
	delIds := arrayutil.Map(delList, func(t *model.CollectorDevice) int64 {
		return t.Id
	})
	// 3、以事务进行更新数据
	if err = TransactionUpdate(t.db.WithContext(ctx), addList, delIds, "collector device"); err != nil {
		return errors.Wrapf(err, "transaction update collector device failed")
	}
	return nil
}

func (t collectorDeviceDaoImpl) GetList(ctx context.Context, cond *cond.ListCollectorDeviceCond) ([]*model.CollectorDevice, int64, error) {
	sql := t.db.WithContext(ctx).Model(model.CollectorDevice{})
	// 处理查询条件
	if len(cond.DeviceGid) > 0 {
		sql = sql.Where("device_gid in ?", cond.DeviceGid)
	}
	if len(cond.DeviceNumber) > 0 {
		sql = sql.Where("device_number in ?", cond.DeviceNumber)
	}
	if len(cond.DeviceSn) > 0 {
		sql = sql.Where("device_sn in ?", cond.DeviceSn)
	}
	if len(cond.ParentDeviceNumber) > 0 {
		sql = sql.Where("parent_device_number in ?", cond.ParentDeviceNumber)
	}
	if len(cond.CollectorType) > 0 {
		sql = sql.Where("collector_type in ?", cond.CollectorType)
	}
	if len(cond.DeviceTypeEn) > 0 {
		sql = sql.Where("device_type_en in ?", cond.DeviceTypeEn)
	}
	if len(cond.TemplateName) > 0 {
		sql = sql.Where("template_name in ?", cond.TemplateName)
	}
	if len(cond.MozuId) > 0 {
		sql = sql.Where("mozu_id in ?", cond.MozuId)
	}
	// 处理分页
	var total int64
	if cond.Page > 0 && cond.Size > 0 {
		if err := sql.Count(&total).Error; err != nil {
			return nil, 0, errors.Wrapf(err, "count collector device fail")
		}
		sql.Offset(int((cond.Page - 1) * cond.Size)).Limit(int(cond.Size))
	}
	res := make([]*model.CollectorDevice, 0)
	if err := sql.Find(&res).Error; err != nil {
		return nil, 0, err
	}
	return res, lo.Max([]int64{total, int64(len(res))}), nil
}

func (t collectorDeviceDaoImpl) GetCollectorDataVer(ctx context.Context, deviceNumber []string) (map[string]int64, error) {
	// 查询每个采集器的数据版本(更新时间)
	majorSql := t.db.WithContext(ctx).Model(model.CollectorDevice{}).
		Select("device_number", "update_at").
		Where("collector_type in (1, 3)")
	if len(deviceNumber) > 0 {
		majorSql.Where("device_number in ?", deviceNumber)
	}
	majorRes := make([]*model.CollectorDevice, 0)
	if err := majorSql.Find(&majorRes).Error; err != nil {
		return nil, err
	}
	// 查询每个采集器下子设备的数据版本(更新时间)
	subSql := t.db.WithContext(ctx).Model(model.CollectorDevice{}).
		Select("parent_device_number", "max(update_at) as update_at").
		Where("collector_type in (2, 4)").
		Group("parent_device_number")

	if len(deviceNumber) > 0 {
		subSql.Where("parent_device_number in ?", deviceNumber)
	}
	subRes := make([]*model.CollectorDevice, 0)
	if err := subSql.Find(&subRes).Error; err != nil {
		return nil, err
	}
	// 获取采集器和所有子设备的最大更新时间作为版本号
	verMap := make(map[string]int64, len(majorRes))
	for _, item := range majorRes {
		verMap[item.DeviceNumber] = item.UpdateAt.Unix()
	}
	for _, item := range subRes {
		verMap[item.ParentDeviceNumber] = lo.Max([]int64{verMap[item.ParentDeviceNumber], item.UpdateAt.Unix()})
	}
	return verMap, nil
}
