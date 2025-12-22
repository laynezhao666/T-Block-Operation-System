// Package register 计算Worker注册相关逻辑
package register

import (
	"context"
	"etrpc-go/config"
	"etrpc-go/log"
	"sync"
	"time"

	"trpcprotocol/scheduler"

	"trpc.group/trpc-go/trpc-go"
)

var (
	worker      = &scheduler.WorkerInfo{}     // 心跳上报请求参数
	cli         scheduler.RegisterClientProxy // 心跳上报客户端
	TaskVerMark string                        // 下发任务的数据版本标识,在接收到下发后可进行更新
)

// ReportHeartbeat 上报当前服务心跳
func ReportHeartbeat(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	// 初始化instance基本信息
	initWorkerInfo()
	interval := 5 * time.Second
	for {
		select {
		case <-ctx.Done():
			unregister(trpc.BackgroundContext())
			return
		case <-time.After(interval):
			register(ctx)
		}
	}
}

func initWorkerInfo() {
	worker.Ip = trpc.GlobalConfig().Global.LocalIP                 // 设置IP
	worker.Port = config.GetInt32OrDefault("heartbeat.port", 8080) // 设置端口,不传默认8080,trpc默认8081
	worker.StartTime = time.Now().Unix()                           // 设置程序启动时间
	worker.WorkerType = scheduler.WorkerInfo_POINT                 // 设置Worker的任务类型
	worker.WorkerSet = trpc.GlobalConfig().Global.FullSetName      // 设置Set名称
	worker.WorkerProtocol = scheduler.WorkerInfo_HTTP              // 不传默认http,可调整为trpc
	cli = scheduler.NewRegisterClientProxy()
}

func register(ctx context.Context) {
	worker.MaxProcessCap = config.GetInt32OrDefault("heartbeat.max_process_cap", 0) // 动态获取最大处理能力
	worker.WorkerStatus = scheduler.WorkerInfo_HEALTHY                              // 设置Worker状态,可根据实际情况调整
	worker.TaskVerMark = TaskVerMark                                                // 设置下发任务的数据版本标识
	if _, err := cli.Heartbeat(ctx, worker); err != nil {
		log.AlarmContext(ctx, "register worker failed, err:", err.Error())
	}
}

func unregister(ctx context.Context) {
	worker.WorkerStatus = scheduler.WorkerInfo_SHUTDOWN // 状态设置为关闭
	if _, err := cli.Heartbeat(ctx, worker); err != nil {
		log.AlarmContext(ctx, "unregister worker failed, err:", err.Error())
	}
}
