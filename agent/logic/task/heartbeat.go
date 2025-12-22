// Package task 任务相关
package task

import (
	"agent/entity/config"
	"context"
	"encoding/json"
	"net"
	"sync"
	"time"

	"trpc.group/trpc-go/trpc-go"

	tredis "etrpc-go/client/redis"
	"etrpc-go/log"

	"github.com/redis/go-redis/v9"
)

// instanceInfo 实例信息
type instanceInfo struct {
	Ip            string // 注册IP
	Port          int64  // 注册端口
	StartTime     int64  // 启动时间
	MaxProcessCap int64  // 最大处理能力
}

const (
	TBosRedisName = "trpc.redis.tbos" // Redis实例类常量

	CacheKeyRegisterCollectorWorkerList = "cache_key_register_collector_worker_list" // 保存当前注册的所有采集worker的key

	heartBeatInterval = 5 * time.Second // 心跳间隔
)

var (
	instanceStr string
)

func initInstanceInfo() {
	stdProcessCap := config.GetRB().Task.Schedule.MaxProcessCap
	instance := instanceInfo{
		Ip:            getIDCNetIP(),
		Port:          int64(trpc.GlobalConfig().Server.Service[0].Port),
		StartTime:     time.Now().Unix(),
		MaxProcessCap: stdProcessCap,
	}
	if instanceBytes, err := json.Marshal(instance); err != nil {
		panic("marshal instance info fail")
	} else {
		instanceStr = string(instanceBytes)
	}
}

// ReportHeartbeat 上报心跳
func ReportHeartbeat(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	// 初始化instance基本信息
	initInstanceInfo()
	for {
		select {
		case <-ctx.Done():
			unregister(context.Background())
			return
		case <-time.After(heartBeatInterval):
			register(ctx)
		}
	}
}

func register(ctx context.Context) {
	curTime := float64(time.Now().Unix())
	if err := tredis.GetRedis(TBosRedisName).ZAdd(ctx, CacheKeyRegisterCollectorWorkerList,
		redis.Z{Score: curTime, Member: instanceStr}).Err(); err != nil {
		log.AlarmContext(ctx, "register worker failed, err:", err.Error())
	}
}

func unregister(ctx context.Context) {
	if err := tredis.GetRedis(TBosRedisName).ZRem(ctx, CacheKeyRegisterCollectorWorkerList,
		instanceStr).Err(); err != nil {
		log.AlarmContext(ctx, "unregister worker failed, err:", err.Error())
	}
	log.Infof("unregister worker success:%v", instanceStr)
}

func getIDCNetIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, address := range addrs {
		// todo review 心跳仅在gateway mode使用，此模式下当前与调度服务不在一个集群，故使用非局域网地址IsPrivate
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsPrivate() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
