package db

import (
	"cmdb/entity/cond"
	"cmdb/util/collutil"
	"common/entity/consts"
	"common/entity/model"
	"context"
	tgorm "etrpc-go/client/gorm"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

// ICollectorTemplateDao 采集模版Dao
type ICollectorTemplateDao interface {
	// QueryCollectorTemplate 查找采集设备模版
	QueryCollectorTemplate(ctx context.Context, templateNames []string) ([]*model.CollectorTemplate, error)
	// BatchUpdate 用最新的采集模版列表更新BD中已有的采集模版列表
	BatchUpdate(ctx context.Context, mozuId int32, data []*model.CollectorTemplate) error
	// GetList 获取采集模版列表
	GetList(ctx context.Context, cond *cond.ListCollectorTemplateCond) ([]*model.CollectorTemplate, int64, error)
}

type collectorTemplateDaoImpl struct {
	db *gorm.DB
}

// NewCollectorTemplateDao 创建采集模版Dao
func NewCollectorTemplateDao() ICollectorTemplateDao {
	return &collectorTemplateDaoImpl{
		db: tgorm.GetDB(consts.TbosMysqlName),
	}
}

// BatchUpdate 用最新的采集模版列表更新BD中已有的采集模版列表
func (t *collectorTemplateDaoImpl) BatchUpdate(ctx context.Context, mozuId int32, data []*model.CollectorTemplate) error {
	// 1、查询出所有原来的采集模版列表
	dbTemplates := make([]*model.CollectorTemplate, 0)
	if err := t.db.WithContext(ctx).Where("mozu_id = ?", mozuId).Find(&dbTemplates).Error; err != nil {
		return errors.Wrapf(err, "query exist collector template fail")
	}
	// 2、查找原来的采集模版列表和新从采集模版列表之间的差异
	addList, delList, _ := collutil.FindDiff(dbTemplates, data, func(item *model.CollectorTemplate) string {
		return item.CalcUniqueKey()
	}, false)
	delIds := lo.Map(delList, func(item *model.CollectorTemplate, index int) int64 {
		return item.Id
	})
	// 3、执行数据更新操作,采用事务进行操作
	if err := TransactionUpdate(t.db.WithContext(ctx), addList, delIds, "collector template"); err != nil {
		return errors.Wrapf(err, "transaction update collector template failed")
	}
	return nil
}

func (t *collectorTemplateDaoImpl) QueryCollectorTemplate(ctx context.Context, templateNames []string) ([]*model.CollectorTemplate, error) {
	templates := make([]*model.CollectorTemplate, 0)
	if len(templateNames) == 0 {
		return templates, nil
	}
	if err := t.db.WithContext(ctx).Where("template_name in ?", templateNames).Find(&templates).Error; err != nil {
		return nil, errors.Wrapf(err, "query collector template fail")
	}
	return templates, nil
}

func (t *collectorTemplateDaoImpl) GetList(ctx context.Context, cond *cond.ListCollectorTemplateCond) ([]*model.CollectorTemplate, int64, error) {
	sql := t.db.WithContext(ctx).Model(&model.CollectorTemplate{})
	if len(cond.MozuId) > 0 {
		sql = sql.Where("mozu_id in ?", cond.MozuId)
	}
	if len(cond.TemplateName) > 0 {
		sql = sql.Where("template_name in ?", cond.TemplateName)
	}
	if len(cond.ProtocolType) > 0 {
		sql = sql.Where("protocol_type in ?", cond.ProtocolType)
	}
	// 处理分页
	var total int64
	if cond.Page > 0 && cond.Size > 0 {
		if err := sql.Count(&total).Error; err != nil {
			return nil, 0, errors.Wrapf(err, "count collector template fail")
		}
		sql.Offset((cond.Page - 1) * cond.Size).Limit(cond.Size)
	}
	res := make([]*model.CollectorTemplate, 0)
	if err := sql.Find(&res).Error; err != nil {
		return nil, 0, err
	}
	return res, lo.Max([]int64{total, int64(len(res))}), nil
}
