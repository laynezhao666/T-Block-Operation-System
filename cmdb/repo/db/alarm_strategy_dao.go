package db

import (
	"cmdb/util/collutil"
	"common/entity/consts"
	"common/entity/model"
	"context"
	tgorm "etrpc-go/client/gorm"
	"etrpc-go/util/arrayutil"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// IAlarmStrategyDao 设备告警策略表相关操作接口
type IAlarmStrategyDao interface {
	// QueryByMozuId 按模组ID查询所有相关告警策略
	QueryByMozuId(ctx context.Context, mozuId int32) ([]*model.AlarmStrategy, error)
	// BatchUpdate 按模组ID批量更新模组相关的告警策略
	BatchUpdate(ctx context.Context, mozuId int32, data []*model.AlarmStrategy) error
}

// NewAlarmStrategyDao 创建设备告警策略表相关操作类对象
func NewAlarmStrategyDao() IAlarmStrategyDao {
	return &alarmStrategyDaoImpl{
		db: tgorm.GetDB(consts.TbosMysqlName),
	}
}

// alarmStrategyDaoImpl 设备告警策略表相关操作具体实现
type alarmStrategyDaoImpl struct {
	db *gorm.DB
}

func (t *alarmStrategyDaoImpl) QueryByMozuId(ctx context.Context, mozuId int32) ([]*model.AlarmStrategy, error) {
	session := t.db.WithContext(ctx)
	var results []*model.AlarmStrategy
	if err := session.Where("mozu_id = ?", mozuId).Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func (t *alarmStrategyDaoImpl) BatchUpdate(ctx context.Context, mozuId int32, data []*model.AlarmStrategy) error {
	dbAlarmStrategies, err := t.QueryByMozuId(ctx, mozuId)
	if err != nil {
		return errors.Wrapf(err, "query exist alarm strategy fail")
	}
	// 2、查找原来的采集模版测点列表和新从采集模版测点列表之间的差异
	addList, delList, _ := collutil.FindDiff(dbAlarmStrategies, data, func(item *model.AlarmStrategy) string {
		return item.CalcUniqueKey()
	}, true)
	delIds := arrayutil.Map(delList, func(t *model.AlarmStrategy) int64 {
		return t.Id
	})
	// 3、执行数据更新操作,采用事务进行操作
	if err := TransactionUpdate(t.db.WithContext(ctx), addList, delIds, "alarm strategy"); err != nil {
		return errors.Wrapf(err, "transaction update alarm strategy failed")
	}
	return nil
}
