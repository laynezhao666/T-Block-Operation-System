package db

import (
	"cgi/entity/dto"
	"cgi/repo/cache"
	"common/entity/consts"
	"common/entity/model"
	tgorm "etrpc-go/client/gorm"
	"etrpc-go/util/arrayutil"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

// IDevicePointDao 设备测点对象查询接口
type IDevicePointDao interface {
	// GetList 获取设备测点列表
	GetList(mozuId int32, cond *dto.CondGetDevicePointList) ([]*model.DevicePoint, int)
}

// NewDevicePointDao 创建设备测点对象查询接口实例
func NewDevicePointDao() IDevicePointDao {
	return &devicePointDaoImpl{
		db: tgorm.GetDB(consts.TbosMysqlName),
		cache: cache.NewMozuDataCache("device_point",
			cache.NewDbTableLoader[*model.DevicePoint](tgorm.GetDB(consts.TbosMysqlName)), 300),
	}
}

type devicePointDaoImpl struct {
	cache cache.IDataCache[[]*model.DevicePoint]
	db    *gorm.DB
}

func (t *devicePointDaoImpl) GetList(mozuId int32, cond *dto.CondGetDevicePointList) ([]*model.DevicePoint, int) {
	// 从缓存读取数据
	res, ok := t.cache.GetData(mozuId)
	if !ok {
		return nil, 0
	}
	if cond == nil {
		return res, len(res)
	}
	res = t.filterByMozu(res, cond)
	// 处理分页
	total := len(res)
	if cond.Page > 0 && cond.Size > 0 {
		chunkRes := lo.Chunk(res, cond.Size)
		if cond.Page > len(chunkRes) {
			res = make([]*model.DevicePoint, 0)
		} else {
			res = chunkRes[cond.Page-1]
		}
	}
	return res, total
}

func (t *devicePointDaoImpl) filterByMozu(data []*model.DevicePoint, cond *dto.CondGetDevicePointList) []*model.DevicePoint {
	if len(cond.DeviceGid) > 0 {
		condMap := arrayutil.ToMap(cond.DeviceGid)
		data = lo.Filter(data, func(item *model.DevicePoint, index int) bool {
			return lo.HasKey(condMap, item.DeviceGid)
		})
	}
	if len(cond.DeviceNumber) > 0 {
		condMap := arrayutil.ToMap(cond.DeviceNumber)
		data = lo.Filter(data, func(item *model.DevicePoint, index int) bool {
			return lo.HasKey(condMap, item.DeviceNumber)
		})
	}
	if len(cond.PointNameEn) > 0 {
		condMap := arrayutil.ToMap(cond.PointNameEn)
		data = lo.Filter(data, func(item *model.DevicePoint, index int) bool {
			return lo.HasKey(condMap, item.PointNameEn)
		})
	}
	if len(cond.PointNameZh) > 0 {
		condMap := arrayutil.ToMap(cond.PointNameZh)
		data = lo.Filter(data, func(item *model.DevicePoint, index int) bool {
			return lo.HasKey(condMap, item.PointNameZh)
		})
	}
	if len(cond.PointKey) > 0 {
		condMap := arrayutil.ToMap(cond.PointKey)
		data = lo.Filter(data, func(item *model.DevicePoint, index int) bool {
			return lo.HasKey(condMap, item.PointKey)
		})
	}
	if len(cond.PointRw) > 0 {
		condMap := arrayutil.ToMap(cond.PointRw)
		data = lo.Filter(data, func(item *model.DevicePoint, index int) bool {
			return lo.HasKey(condMap, item.PointRw)
		})
	}
	if len(cond.PointLevel) > 0 {
		condMap := arrayutil.ToMap(cond.PointLevel)
		data = lo.Filter(data, func(item *model.DevicePoint, index int) bool {
			return lo.HasKey(condMap, item.PointLevel)
		})
	}
	if len(cond.ValueType) > 0 {
		condMap := arrayutil.ToMap(cond.ValueType)
		data = lo.Filter(data, func(item *model.DevicePoint, index int) bool {
			return lo.HasKey(condMap, item.ValueType)
		})
	}
	return data
}
