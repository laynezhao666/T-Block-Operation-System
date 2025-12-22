package service

import (
	"context"
	tredis "etrpc-go/client/redis"
	"etrpc-go/util/copyutil"
	"fmt"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/emptypb"
	"scheduler/entity/model"
	"time"
	"trpc.group/trpc-go/trpc-database/goredis/redlock"
	"trpcprotocol/scheduler"
)

// NewAdminService 创建一个调度管理服务
func NewAdminService() scheduler.AdminService {
	return &adminServiceImpl{}
}

type adminServiceImpl struct {
}

func (a *adminServiceImpl) ResetScheduler(ctx context.Context, req *scheduler.ReqResetScheduler) (*emptypb.Empty, error) {
	// 查找到对应的调度任务
	cfg := GetSchedulerService().GetConfig()
	filterTask := lo.Filter(cfg.Scheduler, func(item *model.TaskConfig, index int) bool {
		return item.Type == req.Type && item.SetGroup == req.SetGroup
	})
	if len(filterTask) != 1 {
		return nil, fmt.Errorf("found [%d] relate schduler task, please check request param", len(filterTask))
	}
	taskConfig := filterTask[0]
	redisCli := tredis.GetRedis(taskConfig.RedisName)

	// 任务可能正在执行，所以这里等待抢占任务分布式锁，直到成功或者超时
	lockCli, _ := redlock.New(redisCli)
	lock, err := lockCli.Lock(ctx, taskConfig.LockKey, redlock.WithLockInterval(time.Second),
		redlock.WithLockTimeout(time.Minute), redlock.WithKeyExpiration(time.Second*5))
	if err != nil {
		return nil, fmt.Errorf("get task lock fail, task may is running please try later, err:%v", err)
	}
	// 最后释放锁
	defer func(lock redlock.Mutex, ctx context.Context) {
		_ = lock.Unlock(ctx)
	}(lock, ctx)
	// 删除任务相关的key信息
	err = redisCli.Del(ctx, taskConfig.CalcAllHistoryKey()...).Err()
	if err != nil {
		return nil, fmt.Errorf("reset type:[%s], setGroup:[%s] of scheduler task fail, err:[%v]", req.Type, req.SetGroup, err)
	}
	return &emptypb.Empty{}, nil
}

// ShowAllScheduler 展示所有的调度任务
func (a *adminServiceImpl) ShowAllScheduler(ctx context.Context, req *emptypb.Empty) (*scheduler.RspShowAllScheduler, error) {
	// 读取配置获取所有调度任务
	cfg := GetSchedulerService().GetConfig()
	taskItems := make([]*scheduler.RspShowAllScheduler_TaskItem, 0, len(cfg.Scheduler))
	for _, item := range cfg.Scheduler {
		rspItem := &scheduler.RspShowAllScheduler_TaskItem{}
		_ = copyutil.Copy(item, rspItem)
		taskItems = append(taskItems, rspItem)
	}
	return &scheduler.RspShowAllScheduler{
		Task: taskItems,
	}, nil
}
