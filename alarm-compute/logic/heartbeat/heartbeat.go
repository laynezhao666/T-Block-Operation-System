// Package heartbeat heathbeat
package heartbeat

import (
	"context"
	"sync"
	"time"

	"etrpc-go/config"
	"etrpc-go/log"

	scheduler_pb "trpcprotocol/scheduler"

	"trpc.group/trpc-go/trpc-go"

	"alarm-compute/conf"
)

var (
	heartAgent *HeartManager
	once       sync.Once
)

// HeartManager 心跳管理器
type HeartManager struct {
	mu          sync.RWMutex
	worker      *scheduler_pb.WorkerInfo         // 心跳上报请求参数
	timeStamp   int64                            // 版本号对应的时间戳
	taskVerMark string                           // 下发任务的数据版本标识,在接收到下发后可进行更新
	cli         scheduler_pb.RegisterClientProxy // 心跳上报客户端
}

// GetHeartAgent 获取心跳管理器
func GetHeartAgent() *HeartManager {
	once.Do(func() {
		heartAgent = &HeartManager{
			worker: &scheduler_pb.WorkerInfo{},
			cli:    scheduler_pb.NewRegisterClientProxy(),
		}
		heartAgent.worker.Ip = trpc.GlobalConfig().Global.LocalIP                 // 设置IP
		heartAgent.worker.Port = config.GetInt32OrDefault("heartbeat.port", 8080) // 设置端口,不传默认8080,trpc默认8081
		heartAgent.worker.StartTime = time.Now().Unix()                           // 设置程序启动时间
		heartAgent.worker.WorkerType = scheduler_pb.WorkerInfo_ALARM              // 设置Worker的任务类型
		heartAgent.worker.WorkerSet = trpc.GlobalConfig().Global.FullSetName      // 设置Set名称
		heartAgent.worker.WorkerProtocol = scheduler_pb.WorkerInfo_HTTP           // 不传默认http,可调整为trpc
	})
	return heartAgent
}

// SetTaskVerMask 设置任务版本标识
func (h *HeartManager) SetTaskVerMask(ctx context.Context, timeStamp int64, taskVerMask string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.timeStamp < timeStamp {
		h.taskVerMark = taskVerMask
		h.timeStamp = timeStamp
	}
}

func (h *HeartManager) register(ctx context.Context) {
	heartAgent.worker.MaxProcessCap = config.GetInt32OrDefault("heartbeat.max_process_cap", 0) // 动态获取最大处理能力
	heartAgent.worker.WorkerStatus = scheduler_pb.WorkerInfo_HEALTHY                           // 设置Worker状态,可根据实际情况调整
	heartAgent.worker.TaskVerMark = h.taskVerMark
	if _, err := heartAgent.cli.Heartbeat(ctx, heartAgent.worker); err != nil {
		log.AlarmContext(ctx, "register worker failed, err:", err.Error())
	}
}

func (h *HeartManager) unregister(ctx context.Context) {
	heartAgent.worker.WorkerStatus = scheduler_pb.WorkerInfo_SHUTDOWN // 状态设置为关闭
	if _, err := heartAgent.cli.Heartbeat(ctx, heartAgent.worker); err != nil {
		log.AlarmContext(ctx, "unregister worker failed, err:", err.Error())
	}
}

// ReportHeartbeat 上报当前服务心跳
func ReportHeartbeat(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	// 初始化instance基本信息
	interval := time.Duration(conf.ServerConf.HeartBeatConfig.HeartBeatInterval) * time.Second
	for {
		select {
		case <-ctx.Done():
			GetHeartAgent().unregister(trpc.BackgroundContext())
			return
		case <-time.After(interval):
			GetHeartAgent().register(ctx)
		}
	}
}
