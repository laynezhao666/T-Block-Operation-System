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
	"strings"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

// IDeviceEntityDao 设备实体对象查询接口
type IDeviceEntityDao interface {
	GetList(mozuId int32, cond *dto.CondDeviceEntityGetList) ([]*model.DeviceEntity, int)
	KeyDicFunc(keyFunc func(item *model.DeviceEntity) string) GetKeyDicFunc
	KvDicFunc(keyFunc, valFunc func(item *model.DeviceEntity) string) GetKvDicFunc
	CalcAlarmCoverDeviceCnt(ctx context.Context, mozuId int32) (int32, error)
}

// NewDeviceEntityDao 创建设备实体对象查询接口实例
func NewDeviceEntityDao() IDeviceEntityDao {
	return &deviceEntityDaoImpl{
		db: tgorm.GetDB(consts.TbosMysqlName),
		cache: cache.NewMozuDataCache("device_entity",
			cache.NewDbTableLoader[*model.DeviceEntity](tgorm.GetDB(consts.TbosMysqlName)), 300),
	}
}

type deviceEntityDaoImpl struct {
	db    *gorm.DB
	cache cache.IDataCache[[]*model.DeviceEntity]
}

func (t *deviceEntityDaoImpl) GetList(mozuId int32, cond *dto.CondDeviceEntityGetList) ([]*model.DeviceEntity, int) {
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
			res = make([]*model.DeviceEntity, 0)
		} else {
			res = chunkRes[cond.Page-1]
		}
	}
	return res, total
}

func (t *deviceEntityDaoImpl) KvDicFunc(keyFunc, valFunc func(item *model.DeviceEntity) string) GetKvDicFunc {
	return func(ctx context.Context, filter string, mozuId int32) (map[string]string, error) {
		if mozuId == 0 {
			return nil, fmt.Errorf("mozuId is required")
		}
		res := make(map[string]string)
		if data, ok := t.cache.GetData(mozuId); ok {
			if filter == "" {
				for _, item := range data {
					res[keyFunc(item)] = valFunc(item)
				}
			} else {
				for _, item := range data {
					key := keyFunc(item)
					val := valFunc(item)
					if strings.Contains(key, filter) || strings.Contains(val, filter) {
						res[key] = val
					}
				}
			}
		}
		return res, nil
	}
}

func (t *deviceEntityDaoImpl) KeyDicFunc(keyFunc func(item *model.DeviceEntity) string) GetKeyDicFunc {
	return func(ctx context.Context, filter string, mozuId int32) ([]string, error) {
		if mozuId == 0 {
			return nil, fmt.Errorf("mozuId is required")
		}
		if data, ok := t.cache.GetData(mozuId); ok {
			res := make([]string, 0)
			if filter == "" {
				for _, item := range data {
					res = append(res, keyFunc(item))
				}
			} else {
				for _, item := range data {
					s := keyFunc(item)
					if strings.Contains(s, filter) {
						res = append(res, s)
					}
				}
			}
			return res, nil
		} else {
			return make([]string, 0), nil
		}
	}
}

func (t *deviceEntityDaoImpl) filterByMozu(data []*model.DeviceEntity, cond *dto.CondDeviceEntityGetList) []*model.DeviceEntity {
	if len(cond.DeviceGid) > 0 {
		data = lo.Filter(data, func(item *model.DeviceEntity, index int) bool {
			return lo.Contains(cond.DeviceGid, item.DeviceGid)
		})
	}
	if len(cond.DeviceNumber) > 0 {
		condMap := arrayutil.ToMap(cond.DeviceNumber)
		data = lo.Filter(data, func(item *model.DeviceEntity, index int) bool {
			return lo.HasKey(condMap, item.DeviceNumber)
		})
	}
	if len(cond.ParentDeviceNumber) > 0 {
		condMap := arrayutil.ToMap(cond.ParentDeviceNumber)
		data = lo.Filter(data, func(item *model.DeviceEntity, index int) bool {
			return lo.HasKey(condMap, item.ParentDeviceNumber)
		})
	}

	if len(cond.ApplicationTypeEn) > 0 {
		condMap := arrayutil.ToMap(cond.ApplicationTypeEn)
		data = lo.Filter(data, func(item *model.DeviceEntity, index int) bool {
			return lo.HasKey(condMap, item.ApplicationTypeEn)
		})
	}
	if len(cond.ApplicationTypeZh) > 0 {
		condMap := arrayutil.ToMap(cond.ApplicationTypeZh)
		data = lo.Filter(data, func(item *model.DeviceEntity, index int) bool {
			return lo.HasKey(condMap, item.ApplicationTypeZh)
		})
	}
	return data
}

func (t *deviceEntityDaoImpl) CalcAlarmCoverDeviceCnt(ctx context.Context, mozuId int32) (int32, error) {
	sql := t.db.WithContext(ctx).Select("count(distinct(device_gid))").Table("t_alarm_strategy").
		Where("mozu_id = ? and device_gid in (select distinct device_gid from t_device_entity where mozu_id=?)",
			mozuId, mozuId)
	var coverCnt int32
	if err := sql.Find(&coverCnt).Error; err != nil {
		return 0, err
	}
	return coverCnt, nil
}
