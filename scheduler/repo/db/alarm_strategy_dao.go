// Package db store database related method, such as Db query function
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

// NewAlarmSchedulerDao 创建告警策略相关数据操作接口的实例化对象
func NewAlarmSchedulerDao(unitCfg *model.TaskConfig) ISchedulerDao[*dbmodel.AlarmStrategy] {
	obj := &alarmStrategyDao{}
	obj.DefaultSchedulerDao = DefaultSchedulerDao[*dbmodel.AlarmStrategy]{
		ISchedulerDao: obj,
		Cache:         tredis.GetRedis(unitCfg.RedisName),
		Db:            tgorm.GetDB(unitCfg.MysqlName),
		Cfg:           unitCfg,
	}
	return obj
}

type alarmStrategyDao struct {
	DefaultSchedulerDao[*dbmodel.AlarmStrategy]
	cache      []*model.TaskItem[*dbmodel.AlarmStrategy]
	lastVerStr string
}

func (d *alarmStrategyDao) GetPublishData(ctx context.Context, verNoChanged bool) ([]*model.TaskItem[*dbmodel.AlarmStrategy], error) {
	if verNoChanged && d.cache != nil {
		return d.cache, nil
	}
	sql := d.Db.WithContext(ctx)
	// 按模组过滤数据
	if len(d.Cfg.FilterMozu) > 0 {
		sql = sql.Where("mozu_id in ?", d.Cfg.FilterMozu)
	}
	res := make([]*dbmodel.AlarmStrategy, 0)
	batchRes := make([]*dbmodel.AlarmStrategy, 0)
	// 分批读取
	if err := sql.FindInBatches(&batchRes, consts.MySQLFetchBatchSize, func(tx *gorm.DB, batch int) error {
		res = append(res, batchRes...)
		return nil
	}).Error; err != nil {
		return nil, err
	}
	// 转化为任务对象
	data := lo.Map(res, func(item *dbmodel.AlarmStrategy, index int) *model.TaskItem[*dbmodel.AlarmStrategy] {
		return &model.TaskItem[*dbmodel.AlarmStrategy]{
			TaskData:    item,
			TaskKey:     item.CalcCombineKey(),
			ComputeCost: 1, // 默认计算耗时为1
		}
	})
	d.cache = data
	return data, nil
}
