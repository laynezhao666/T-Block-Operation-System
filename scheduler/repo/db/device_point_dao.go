package db

import (
	"context"
	tgorm "etrpc-go/client/gorm"
	tredis "etrpc-go/client/redis"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"scheduler/entity/consts"
	"scheduler/entity/dbmodel"
	"scheduler/entity/model"
)

// NewPointSchedulerDao 创建标准点计算相关数据操作接口的实例化对象
func NewPointSchedulerDao(unitCfg *model.TaskConfig) ISchedulerDao[*dbmodel.DevicePoint] {
	obj := &devicePointDao{}
	obj.DefaultSchedulerDao = DefaultSchedulerDao[*dbmodel.DevicePoint]{
		ISchedulerDao: obj,
		Cache:         tredis.GetRedis(unitCfg.RedisName),
		Db:            tgorm.GetDB(unitCfg.MysqlName),
		Cfg:           unitCfg,
	}
	return obj
}

type devicePointDao struct {
	DefaultSchedulerDao[*dbmodel.DevicePoint]
	cache []*model.TaskItem[*dbmodel.DevicePoint]
}

// GetPublishData 获取需要下发的标准点数据
func (d *devicePointDao) GetPublishData(ctx context.Context, verNoChanged bool) ([]*model.TaskItem[*dbmodel.DevicePoint], error) {
	if verNoChanged && d.cache != nil {
		return d.cache, nil
	}
	sql := d.Db.WithContext(ctx).Where("belong_collector = ''")
	// 按模组过滤数据
	if len(d.Cfg.FilterMozu) > 0 {
		sql = sql.Where("mozu_id in ?", d.Cfg.FilterMozu)
	}
	// 分批读取
	res := make([]*dbmodel.DevicePoint, 0)
	batchRes := make([]*dbmodel.DevicePoint, 0)
	if err := sql.FindInBatches(&batchRes, consts.MySQLFetchBatchSize, func(tx *gorm.DB, batch int) error {
		res = append(res, batchRes...)
		return nil
	}).Error; err != nil {
		return nil, err
	}
	// 转化为任务对象
	data := lo.Map(res, func(item *dbmodel.DevicePoint, index int) *model.TaskItem[*dbmodel.DevicePoint] {
		return &model.TaskItem[*dbmodel.DevicePoint]{
			TaskData:    item,
			TaskKey:     item.CalcUniqueKey(),
			ComputeCost: 1,
		}
	})
	d.cache = data
	return data, nil
}
