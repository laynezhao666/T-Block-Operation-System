// Package db 存放MySQL,Redis相关的数据库操作逻辑
package db

import (
	"context"
	"encoding/json"
	"errors"
	"etrpc-go/log"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"scheduler/entity/consts"
	"scheduler/entity/dbmodel"
	"scheduler/entity/model"
	"scheduler/repo/cache"
	"strings"
)

// ISchedulerDao 调度器操作DB数据的接口
type ISchedulerDao[T any] interface {
	// GetCurVersionStr 获取最新版本标识
	GetCurVersionStr(ctx context.Context) (string, error)
	// GetLastVerStr 获取上次下发使用的版本标识
	GetLastVerStr(ctx context.Context) (string, error)
	// SetLastVerStr  设置本次下发使用的版本标识
	SetLastVerStr(ctx context.Context, value string) error

	// GetRegisterWorkerList 获取实时注册的worker列表
	GetRegisterWorkerList(ctx context.Context) ([]*model.WorkerInfo, error)
	// GetLastWorkerList  获取上次下发使用的worker列表
	GetLastWorkerList(ctx context.Context) ([]*model.WorkerInfo, error)
	// SetLastWorkerList  设置本次下发使用的worker列表
	SetLastWorkerList(ctx context.Context, workers []*model.WorkerInfo) error

	// GetLastAssignResult 获取上次的分配关系
	GetLastAssignResult(ctx context.Context) (map[string]string, error)
	// SetLastAssignResult  设置上次的分配关系
	SetLastAssignResult(ctx context.Context, workerMap map[string]string) error

	// GetPublishData 获取需要下发的数据列表,verNoChanged版本是否变化
	GetPublishData(ctx context.Context, verNoChanged bool) ([]*model.TaskItem[T], error)
}

// DefaultSchedulerDao 默认的调度器操作DB的实现类
type DefaultSchedulerDao[T any] struct {
	ISchedulerDao[T]
	Cache redis.UniversalClient
	Cfg   *model.TaskConfig
	Db    *gorm.DB
}

// GetCurVersionStr 从DB获取最新版本标识
func (s *DefaultSchedulerDao[T]) GetCurVersionStr(ctx context.Context) (string, error) {
	// 按mozu_id排个序,方便后面进行比对
	sql := s.Db.WithContext(ctx).Model(&dbmodel.MozuInfo{}).Order("mozu_id")
	// 需要过滤出具体的模组
	if len(s.Cfg.FilterMozu) > 0 {
		sql.Where("mozu_id in ?", s.Cfg.FilterMozu)
	}
	res := make([]*dbmodel.MozuInfo, 0)
	if err := sql.Find(&res).Error; err != nil {
		return "", err
	}
	// 取出所有的模组和版本信息,拼接成字符串
	mozuVerStrList := lo.Map(res, func(item *dbmodel.MozuInfo, index int) string {
		version := item.PublishVersion
		if s.Cfg.Type == model.TaskTypeAlarm {
			version = item.AlarmVersion
		}
		return fmt.Sprintf("%d;%s", item.MozuId, version)
	})
	return strings.Join(mozuVerStrList, consts.CommonFieldSeq), nil
}

// GetLastVerStr 从Redis缓存获取上次的下发版本标识
func (s *DefaultSchedulerDao[T]) GetLastVerStr(ctx context.Context) (string, error) {
	res, err := s.Cache.Get(ctx, s.Cfg.LastVerStrKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", err
	}
	return res, nil
}

// SetLastVerStr 将上次下发用的版本标识缓存到Redis
func (s *DefaultSchedulerDao[T]) SetLastVerStr(ctx context.Context, value string) error {
	return s.Cache.Set(ctx, s.Cfg.LastVerStrKey, value, 0).Err()
}

// GetRegisterWorkerList 获取实时注册的worker列表
func (s *DefaultSchedulerDao[T]) GetRegisterWorkerList(ctx context.Context) ([]*model.WorkerInfo, error) {
	if s.Cfg.OldVer {
		return cache.GetOldRegisterWorkerList(ctx, s.Cache, s.Cfg.RegisterWorkerKey)
	}
	return cache.GetRegisterWorkerList(ctx, s.Cache, s.Cfg.RegisterWorkerKey)
}

// GetLastWorkerList 从Redis获取上次下发使用的worker列表
func (s *DefaultSchedulerDao[T]) GetLastWorkerList(ctx context.Context) ([]*model.WorkerInfo, error) {
	workersStr := make([]string, 0)
	if err := cache.GetCacheObj(ctx, s.Cache, s.Cfg.LastRegisterWorkerKey, &workersStr); err != nil {
		return nil, err
	}
	workers := make([]*model.WorkerInfo, 0, len(workersStr))
	for _, workerStr := range workersStr {
		worker := &model.WorkerInfo{}
		if err := json.Unmarshal([]byte(workerStr), worker); err != nil {
			log.WarnContextf(ctx, "bad last worker str:[%s], unmarshal err:[%v]", workerStr, err)
			continue
		}
		workers = append(workers, worker)
	}
	return workers, nil
}

// SetLastWorkerList 将本次下发用的worker列表缓存到Redis
func (s *DefaultSchedulerDao[T]) SetLastWorkerList(ctx context.Context, workers []*model.WorkerInfo) error {
	workersStr := lo.Map(workers, func(item *model.WorkerInfo, index int) string {
		return item.ToJsonString()
	})
	return cache.SetCacheObj(ctx, s.Cache, s.Cfg.LastRegisterWorkerKey, workersStr)
}

// GetLastAssignResult 获取上次下发的分配关系
func (s *DefaultSchedulerDao[T]) GetLastAssignResult(ctx context.Context) (map[string]string, error) {
	strategyWorkerMap := make(map[string]string)
	// 分配信息需要分片存储，需要分片读取数据
	if s.Cfg.LastAllocateShardCnt > 1 {
		shardAnyData := make([]map[string]string, s.Cfg.LastAllocateShardCnt)
		for idx := range shardAnyData {
			shardAnyData[idx] = make(map[string]string)
		}
		err := cache.GetCacheObjShard(ctx, s.Cache, s.Cfg.LastAssignResultKey, shardAnyData)
		if err != nil {
			return nil, err
		}
		for _, data := range shardAnyData {
			for k, v := range data {
				strategyWorkerMap[k] = v
			}
		}
	} else {
		if err := cache.GetCacheObj(ctx, s.Cache, s.Cfg.LastAssignResultKey, &strategyWorkerMap); err != nil {
			return nil, err
		}
	}
	return strategyWorkerMap, nil
}

// SetLastAssignResult 设置上次的分配关系
func (s *DefaultSchedulerDao[T]) SetLastAssignResult(ctx context.Context, workerMap map[string]string) error {
	// 分配信息需要分片存储，避免Redis产生big key
	if s.Cfg.LastAllocateShardCnt > 1 {
		shardData := make([]map[string]string, s.Cfg.LastAllocateShardCnt)
		shardAnyData := make([]any, s.Cfg.LastAllocateShardCnt)
		for idx := range shardData {
			shardData[idx] = make(map[string]string)
			shardAnyData[idx] = shardData[idx]
		}
		idx := 0
		for k, v := range workerMap {
			shardData[idx%s.Cfg.LastAllocateShardCnt][k] = v
			idx++
		}
		return cache.SetCacheObjShard(ctx, s.Cache, s.Cfg.LastAssignResultKey, shardAnyData)
	} else {
		return cache.SetCacheObj(ctx, s.Cache, s.Cfg.LastAssignResultKey, workerMap)
	}
}
