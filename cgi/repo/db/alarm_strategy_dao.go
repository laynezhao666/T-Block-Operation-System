package db

import (
	"common/entity/consts"
	"common/entity/model"
	"context"
	tgorm "etrpc-go/client/gorm"

	"gorm.io/gorm"
)

// IAlarmStrategyDao 设备告警策略表相关操作接口
type IAlarmStrategyDao interface {
	// GetExpressionByMozuId 按模组ID查询所有告警策略的表达式
	GetExpressionByMozuId(ctx context.Context, mozuId int32) ([]*model.AlarmStrategy, error)
	StatStrategyLevel(ctx context.Context, mozuId int32) ([]*model.AlarmStrategyLevelStat, error)
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

func (t *alarmStrategyDaoImpl) GetExpressionByMozuId(ctx context.Context, mozuId int32) ([]*model.AlarmStrategy, error) {
	sql := t.db.WithContext(ctx).Model(&model.AlarmStrategy{}).Select("id, expression_map")
	var allRes []*model.AlarmStrategy
	var res []*model.AlarmStrategy
	if err := sql.Where("mozu_id = ?", mozuId).FindInBatches(&res, consts.MySQLReadBatchSize, func(tx *gorm.DB, batch int) error {
		allRes = append(allRes, res...)
		return nil
	}).Error; err != nil {
		return nil, err
	}
	return allRes, nil
}

func (t *alarmStrategyDaoImpl) StatStrategyLevel(ctx context.Context, mozuId int32) ([]*model.AlarmStrategyLevelStat, error) {
	sql := t.db.WithContext(ctx).Model(&model.AlarmStrategy{}).
		Select("rid, alarm_level, count(1) as cnt").
		Where("mozu_id = ?", mozuId).
		Group("rid, alarm_level")
	res := make([]*model.AlarmStrategyLevelStat, 0)
	if err := sql.Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}
