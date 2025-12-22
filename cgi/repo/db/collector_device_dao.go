package db

import (
	"cgi/entity/dto"
	"cgi/repo/cache"
	"common/entity/consts"
	"common/entity/model"
	"context"
	tgorm "etrpc-go/client/gorm"
	"etrpc-go/util/arrayutil"
	"fmt"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

// ICollectorDao 获取采集设备接口
type ICollectorDao interface {
	GetDeviceList(mozuId int32, cond *dto.CondCollectorGetDeviceList) ([]*model.CollectorDevice, int)
	GetTemplatePointByTriple(ctx context.Context, tripleCond [][]string) ([]*model.CollectorTemplatePoint, error)
	GetTemplatePoint(ctx context.Context, templateName, subDevice string) ([]*model.CollectorTemplatePoint, error)
	GetCollectPointCntByMozuId(ctx context.Context, mozuId int32) (int32, error)
}

// NewCollectorDao 获取采集设备查询接口
func NewCollectorDao() ICollectorDao {
	return &collectorDaoImpl{
		cache: cache.NewMozuDataCache("collector_device",
			cache.NewDbTableLoader[*model.CollectorDevice](tgorm.GetDB(consts.TbosMysqlName)), 300),
		db: tgorm.GetDB(consts.TbosMysqlName),
	}
}

type collectorDaoImpl struct {
	cache cache.IDataCache[[]*model.CollectorDevice]
	db    *gorm.DB
}

func (c *collectorDaoImpl) GetDeviceList(mozuId int32, cond *dto.CondCollectorGetDeviceList) ([]*model.CollectorDevice, int) {
	// 从缓存读取数据
	res, ok := c.cache.GetData(mozuId)
	if !ok {
		return nil, 0
	}
	if cond == nil {
		return res, len(res)
	}
	res = c.filterByMozu(res, cond)
	// 处理分页
	total := len(res)
	if cond.Page > 0 && cond.Size > 0 {
		chunkRes := lo.Chunk(res, cond.Size)
		if cond.Page > len(chunkRes) {
			res = make([]*model.CollectorDevice, 0)
		} else {
			res = chunkRes[cond.Page-1]
		}
	}
	return res, total
}

func (c *collectorDaoImpl) filterByMozu(data []*model.CollectorDevice, cond *dto.CondCollectorGetDeviceList) []*model.CollectorDevice {
	if len(cond.DeviceGid) > 0 {
		data = lo.Filter(data, func(item *model.CollectorDevice, index int) bool {
			return lo.Contains(cond.DeviceGid, item.DeviceGid)
		})
	}
	if len(cond.DeviceNumber) > 0 {
		condMap := arrayutil.ToMap(cond.DeviceNumber)
		data = lo.Filter(data, func(item *model.CollectorDevice, index int) bool {
			return lo.HasKey(condMap, item.DeviceNumber)
		})
	}
	if len(cond.ParentDeviceNumber) > 0 {
		condMap := arrayutil.ToMap(cond.ParentDeviceNumber)
		data = lo.Filter(data, func(item *model.CollectorDevice, index int) bool {
			return lo.HasKey(condMap, item.ParentDeviceNumber)
		})
	}
	if len(cond.CollectorType) > 0 {
		condMap := arrayutil.ToMap(cond.CollectorType)
		data = lo.Filter(data, func(item *model.CollectorDevice, index int) bool {
			return lo.HasKey(condMap, item.CollectorType)
		})
	}
	return data
}

func (c *collectorDaoImpl) GetTemplatePointByTriple(ctx context.Context, tripleCond [][]string) ([]*model.CollectorTemplatePoint, error) {
	res := make([]*model.CollectorTemplatePoint, len(tripleCond))
	if len(tripleCond) == 0 {
		return res, nil
	}
	sql := c.db.WithContext(ctx).Model(&model.CollectorTemplatePoint{})
	for idx, cond := range tripleCond {
		if len(cond) != 3 {
			return nil, fmt.Errorf("bad condition, cond [%d] have [%d] elem, should be 3", idx, len(cond))
		}
		sql = sql.Or("(template_name = ? and sub_device = ?  and point_name_en = ?)", cond[0], cond[1], cond[2])
	}
	if err := sql.Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (c *collectorDaoImpl) GetTemplatePoint(ctx context.Context, templateName, subDevice string) ([]*model.CollectorTemplatePoint, error) {
	sql := c.db.WithContext(ctx).Where("template_name = ? ", templateName)
	if len(subDevice) > 0 {
		sql = sql.Where("sub_device = ?", subDevice)
	}
	res := make([]*model.CollectorTemplatePoint, 0)
	if err := sql.Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (c *collectorDaoImpl) GetCollectPointCntByMozuId(ctx context.Context, mozuId int32) (int32, error) {
	sql := "SELECT ifnull(sum(tmp.cnt), 0) from t_collector_device cd " +
		"left join (SELECT template_name ,count(1) as cnt from t_collector_template_point where mozu_id=? group by template_name) tmp " +
		"on cd.template_name=tmp.template_name where cd.mozu_id=?"
	var res int32
	if err := c.db.WithContext(ctx).Raw(sql, mozuId, mozuId).Scan(&res).Error; err != nil {
		return 0, err
	}
	return res, nil
}
