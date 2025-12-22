// Package cache 存放和Redis通信的一些通用接口
package cache

import (
	"context"
	"encoding/json"
	"etrpc-go/util/copyutil"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"math"
	"scheduler/entity/consts"
	"scheduler/entity/model"
	"scheduler/util/timezset"
	"time"
	"trpc.group/trpc-go/trpc-go/log"
)

// OldWorkerInfo 旧的Worker信息
type OldWorkerInfo struct {
	Ip            string // IP
	Port          int32  // 端口
	StartTime     int64  // 启动版本号
	MaxProcessCap int64  // 最大处理能力
}

// GetOldRegisterWorkerList 获取注册的worker列表
func GetOldRegisterWorkerList(ctx context.Context, cli redis.UniversalClient, key string) ([]*model.WorkerInfo, error) {
	// worker每5s上报一次心跳,如果3次未收到心跳,则认为worker挂掉了,这里取16秒是为了鲁棒性
	timeZSet := timezset.New(cli, key, time.Second*16)
	// 获取当前注册的所有worker
	workers, err := timeZSet.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	// 移除过期的worker
	if err := timeZSet.RemoveExpire(ctx); err != nil {
		return nil, err
	}
	// 返回当前注册的worker,每一项值为ip:port
	instances := make([]*model.WorkerInfo, 0, len(workers))
	for _, worker := range workers {
		oldWorker := &OldWorkerInfo{}
		if err = json.Unmarshal([]byte(worker.Data.(string)), oldWorker); err == nil {
			newWorker := &model.WorkerInfo{}
			_ = copyutil.Copy(oldWorker, newWorker)
			newWorker.ReportTime = worker.Score.Unix()
			newWorker.WorkerProtocol = "http"
			if newWorker.MaxProcessCap == 0 {
				newWorker.MaxProcessCap = math.MaxInt32
			}
			instances = append(instances, newWorker)
		}
	}
	return instances, nil
}

// GetRegisterWorkerList 获取注册的Worker列表
func GetRegisterWorkerList(ctx context.Context, cli redis.UniversalClient, key string) ([]*model.WorkerInfo, error) {
	res, err := cli.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	// worker每5s上报一次心跳,如果3次未收到心跳,则认为worker挂掉了,这里取17秒是为了鲁棒性
	expireTs := time.Now().Add(-time.Second * 17).Unix()
	workerMap := make(map[string]*model.WorkerInfo, len(res))
	invalidWorkers := make([]string, 0)
	for workerKey, workerInfoStr := range res {
		worker := new(model.WorkerInfo)
		// 格式错误的worker,需要移除掉
		if err = json.Unmarshal([]byte(workerInfoStr), worker); err != nil {
			invalidWorkers = append(invalidWorkers, workerKey)
			log.WarnContextf(ctx, "bad register worker info, key:[%s], worker:[%s]", key, workerInfoStr)
			continue
		}
		// 过期的worker, 也需要移除掉
		if worker.ReportTime <= expireTs {
			invalidWorkers = append(invalidWorkers, workerKey)
			continue
		}
		// 部分情况下,Worker自身panic后会造成重启,但IP不会变,此时旧的注册信息未移除,新的已上报,就会出现IP/Port重复
		// 判断是否有重复的IP/Port,有的话保留最新的
		workerAddr := worker.GetAddr()
		if exist, ok := workerMap[workerAddr]; ok {
			if exist.StartTime < worker.StartTime {
				// 移除启动时间较早的worker
				invalidWorkers = append(invalidWorkers, exist.GetWorkerKey())
				workerMap[workerAddr] = worker
			}
		} else {
			workerMap[workerAddr] = worker
		}
	}
	// 无效的worker移除掉
	if err = RemoveRegisterWorker(ctx, cli, key, invalidWorkers); err != nil {
		return nil, err
	}
	return lo.Values(workerMap), nil
}

// RemoveRegisterWorker 移除无效的Worker列表
func RemoveRegisterWorker(ctx context.Context, cli redis.UniversalClient, key string, workers []string) error {
	if len(workers) <= 0 {
		return nil
	}
	return cli.HDel(ctx, key, workers...).Err()
}

// GetCacheObj 根据key从redis获取value为v的数据,使用json反序列化
func GetCacheObj(ctx context.Context, cli redis.UniversalClient, key string, value any) error {
	result, err := cli.Get(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	if len(result) == 0 {
		return nil
	}
	if err := json.Unmarshal([]byte(result), value); err != nil {
		return err
	}
	return nil
}

// SetCacheObj 使用key缓存v到redis,使用json序列化v, 永不过期
func SetCacheObj(ctx context.Context, cli redis.UniversalClient, key string, value any) error {
	jsonStr, err := json.Marshal(value)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal value")
	}
	if _, err := cli.Set(ctx, key, jsonStr, 0).Result(); err != nil {
		return err
	}
	return nil
}

// SetCacheObjShard 使用key分片缓存v到redis,使用json序列化v, 永不过期,key为 key_0, key_1, ...
func SetCacheObjShard(ctx context.Context, cli redis.UniversalClient, key string, values []any) error {
	if len(values) == 0 {
		return nil
	}
	args := make([]string, 0, len(values)*2)
	for idx, val := range values {
		if jsonStr, err := json.Marshal(val); err != nil {
			return errors.Wrapf(err, "failed to marshal values at [%d]", idx)
		} else {
			args = append(args, fmt.Sprintf("%s%s%d", key, consts.RedisJoinFieldSep, idx))
			args = append(args, string(jsonStr))
		}
	}
	if _, err := cli.MSet(ctx, args).Result(); err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	return nil
}

// GetCacheObjShard 根据key从redis获取值为v的数据,使用json反序列化，注意values里面需为类型指针
func GetCacheObjShard(ctx context.Context, cli redis.UniversalClient, key string, values []map[string]string) error {
	if len(values) == 0 {
		return nil
	}
	keys := make([]string, 0, len(values))
	for idx := range values {
		keys = append(keys, fmt.Sprintf("%s%s%d", key, consts.RedisJoinFieldSep, idx))
	}
	if results, err := cli.MGet(ctx, keys...).Result(); err != nil && !errors.Is(err, redis.Nil) {
		return err
	} else {
		for idx, res := range results {
			if res == nil || res == "" {
				continue
			}
			if err := json.Unmarshal([]byte(res.(string)), &values[idx]); err != nil {
				return errors.Wrapf(err, "failed to unmarshal values at [%d]", idx)
			}
		}
		return nil
	}
}
