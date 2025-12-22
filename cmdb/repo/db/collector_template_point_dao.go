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

// ICollectorTemplatePointDao 设备采集策略表相关操作接口
type ICollectorTemplatePointDao interface {
	// BatchUpdate 用最新的采集模版测点列表更新BD中已有的采集模版测点列表
	BatchUpdate(ctx context.Context, mozuId int32, data []*model.CollectorTemplatePoint) error
	// QueryCollectorTemplatePoint 查找设备模版对应的测点
	QueryCollectorTemplatePoint(ctx context.Context, templateNames []string) ([]*model.CollectorTemplatePoint, error)
}

type collectorTemplatePointDaoImpl struct {
	db *gorm.DB
}

// NewCollectorTemplatePointDao 创建采集模版测点相关操作接口
func NewCollectorTemplatePointDao() ICollectorTemplatePointDao {
	return &collectorTemplatePointDaoImpl{
		db: tgorm.GetDB(consts.TbosMysqlName),
	}
}

// BatchUpdate 用最新的采集模版测点列表更新BD中已有的采集模版测点列表
func (t *collectorTemplatePointDaoImpl) BatchUpdate(ctx context.Context, mozuId int32, data []*model.CollectorTemplatePoint) error {
	// 1、查询出DB中所有的采集模版测点列表
	dbTemplatePoints := make([]*model.CollectorTemplatePoint, 0)
	if err := t.db.WithContext(ctx).Where("mozu_id = ?", mozuId).Find(&dbTemplatePoints).Error; err != nil {
		return errors.Wrapf(err, "query exist collector template point fail")
	}
	// 2、查找原来的采集模版测点列表和新从采集模版测点列表之间的差异
	addList, delList, _ := collutil.FindDiff(dbTemplatePoints, data, func(item *model.CollectorTemplatePoint) string {
		return item.CalcUniqueKey()
	}, false)
	delIds := arrayutil.Map(delList, func(t *model.CollectorTemplatePoint) int64 {
		return t.Id
	})
	// 3、执行更新操作，采用事务执行
	if err := TransactionUpdate(t.db.WithContext(ctx), addList, delIds, "collector point"); err != nil {
		return errors.Wrapf(err, "transaction update collector point failed")
	}
	return nil
}

func (t *collectorTemplatePointDaoImpl) QueryCollectorTemplatePoint(ctx context.Context, templateNames []string) ([]*model.CollectorTemplatePoint, error) {
	points := make([]*model.CollectorTemplatePoint, 0)
	if len(templateNames) == 0 {
		return points, nil
	}
	if err := t.db.WithContext(ctx).Where("template_name in ?", templateNames).Find(&points).Error; err != nil {
		return nil, errors.Wrapf(err, "query collector template point fail")
	}
	return points, nil
}
