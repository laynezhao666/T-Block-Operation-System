package db

import (
	"context"
	tgorm "etrpc-go/client/gorm"
	tredis "etrpc-go/client/redis"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"scheduler/entity/dbmodel"
	"scheduler/entity/model"
)

// NewCollectorSchedulerDao 创建采集策略相关数据操作接口的实例化对象
func NewCollectorSchedulerDao(unitCfg *model.TaskConfig) ISchedulerDao[*dbmodel.CollectorDevice] {
	obj := &collectorDeviceDao{}
	obj.DefaultSchedulerDao = DefaultSchedulerDao[*dbmodel.CollectorDevice]{
		ISchedulerDao: obj,
		Cache:         tredis.GetRedis(unitCfg.RedisName),
		Db:            tgorm.GetDB(unitCfg.MysqlName),
		Cfg:           unitCfg,
	}
	return obj
}

type collectorDeviceDao struct {
	DefaultSchedulerDao[*dbmodel.CollectorDevice]
	cache []*model.TaskItem[*dbmodel.CollectorDevice]
}

type collectorCount struct {
	DeviceNumber string `gorm:"column:device_number;"`
	Cnt          int64  `gorm:"column:cnt;"`
}

func (c *collectorDeviceDao) GetPublishData(ctx context.Context, verNoChanged bool) ([]*model.TaskItem[*dbmodel.CollectorDevice], error) {
	if verNoChanged && c.cache != nil {
		return c.cache, nil
	}
	// 1、查询出所有的关联的设备
	res := make([]*dbmodel.CollectorDevice, 0)
	deviceSql := c.Db.WithContext(ctx).Where("collector_type = 3")
	// 按模组过滤数据
	if len(c.Cfg.FilterMozu) > 0 {
		deviceSql = deviceSql.Where("mozu_id in ?", c.Cfg.FilterMozu)
	}
	if err := deviceSql.Find(&res).Error; err != nil {
		return nil, errors.Wrapf(err, "query collector device list fail")
	}

	// 2、计算每个采集器下面的采集测点数
	collectorCountRes := make([]*collectorCount, 0)
	collectorPointCntSql := c.Db.Table("t_collector_device c " +
		"left join t_collector_template_point p on c.template_name = p.template_name").
		Select("c.parent_device_number as device_number, count(1) as cnt").
		Where("c.collector_type = 3").
		Group("c.parent_device_number")
	if len(c.Cfg.FilterMozu) > 0 {
		collectorPointCntSql = collectorPointCntSql.Where("c.mozu_id in ?", c.Cfg.FilterMozu)
	}
	if err := collectorPointCntSql.Find(&collectorCountRes).Error; err != nil {
		return nil, errors.Wrapf(err, "query collector device collect point cnt fail")
	}
	collectorCntMap := lo.SliceToMap(collectorCountRes, func(item *collectorCount) (string, int64) {
		return item.DeviceNumber, item.Cnt
	})

	// 3、计算每个采集器下面的标准测点数
	stdPointCntRes := make([]*collectorCount, 0)
	stdPointCntSql := c.Db.Table("t_device_point").
		Select("belong_collector as device_number, count(1) as cnt").
		Where("belong_collector !='' ").
		Group("belong_collector")
	if len(c.Cfg.FilterMozu) > 0 {
		stdPointCntSql = stdPointCntSql.Where("mozu_id in ?", c.Cfg.FilterMozu)
	}
	if err := stdPointCntSql.Find(&stdPointCntRes).Error; err != nil {
		return nil, errors.Wrapf(err, "query collector device std point cnt fail")
	}
	stdPointCntMap := lo.SliceToMap(stdPointCntRes, func(item *collectorCount) (string, int64) {
		return item.DeviceNumber, item.Cnt
	})

	// 4、为每个采集器设置计算量,并按计算量从大到小排序
	data := lo.Map(res, func(item *dbmodel.CollectorDevice, index int) *model.TaskItem[*dbmodel.CollectorDevice] {
		// 计算量 = 采集测点数 + 标准测点数
		return &model.TaskItem[*dbmodel.CollectorDevice]{
			TaskData:    item,
			TaskKey:     item.CalcUniqueKey(),
			ComputeCost: collectorCntMap[item.DeviceNumber] + stdPointCntMap[item.DeviceNumber],
		}
	})
	c.cache = data
	return data, nil
}
