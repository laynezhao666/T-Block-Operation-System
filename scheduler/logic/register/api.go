// Package register Worker注册相关接口逻辑
package register

import (
	"context"
	tredis "etrpc-go/client/redis"
	"github.com/redis/go-redis/v9"
	"scheduler/entity/consts"
	"scheduler/entity/model"
	"strings"
	"time"
	"trpcprotocol/scheduler"
)

// IRegisterApi Worker注册接口
type IRegisterApi interface {
	Heartbeat(ctx context.Context, req *scheduler.WorkerInfo) error
}

// NewRegisterApi 创建Worker注册接口实现类
func NewRegisterApi() IRegisterApi {
	return &registerApiImpl{
		cache: tredis.GetRedis(consts.TBosRedisName),
	}
}

type registerApiImpl struct {
	cache redis.UniversalClient
}

func (r *registerApiImpl) Heartbeat(ctx context.Context, req *scheduler.WorkerInfo) error {
	// 构建worker信息
	workerInfo := &model.WorkerInfo{
		Ip:             req.Ip,
		Port:           req.Port,
		StartTime:      req.StartTime,
		MaxProcessCap:  int64(req.MaxProcessCap),
		TaskVerMark:    req.TaskVerMark,
		WorkerProtocol: strings.ToLower(req.WorkerProtocol.String()),
		ReportTime:     time.Now().Unix(),
	}
	redisKey := strings.Join([]string{model.DefaultRegisterWorkKey, strings.ToLower(req.WorkerType.String()),
		req.WorkerSet}, consts.RedisJoinFieldSep)
	if req.WorkerStatus == scheduler.WorkerInfo_SHUTDOWN {
		// 取消注册
		return r.cache.HDel(ctx, redisKey, workerInfo.GetWorkerKey()).Err()
	} else {
		// 执行注册
		return r.cache.HSet(ctx, redisKey, workerInfo.GetWorkerKey(), workerInfo.ToJsonString()).Err()
	}
}
