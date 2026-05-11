package hotstandby

import (
	"agent/entity/config"
	"context"
	"fmt"
	"sync"
	"time"

	pb "trpcprotocol/agent"

	"google.golang.org/protobuf/types/known/emptypb"
	"trpc.group/trpc-go/trpc-go/client"
	"trpc.group/trpc-go/trpc-go/log"
)

// HostStandby 热备接口
type HostStandby interface {
	Start(ctx context.Context, notify chan bool)
	IsMaster(devNum string) bool
	GetDevStatusMap() map[string]bool
}

// GetHotStandbyManager 获取热备管理器
func GetHotStandbyManager() HostStandby {
	return g_hot_standby
}

var g_hot_standby = &hot_standby{
	collectDevIsMaster: make(map[string]bool),
}

type hot_standby struct {
	collectDevIsMaster      map[string]bool // 对应的采集器是否为主状态
	collectDevIsMasterMutex sync.RWMutex
	notify                  chan bool
}

// Start 启动热备
func (h *hot_standby) Start(ctx context.Context, notify chan bool) {
	if !config.GetRB().HotStandbyEnable() {
		return
	}
	log.Infof("hotstandby enable: %v", config.GetRB().Task.Local.HotStandby)
	h.notify = notify
	devStatus := make(map[string]bool)
	target2Dev := make(map[string][]string)
	for dev, info := range config.GetRB().Task.Local.HotStandby {
		devStatus[dev] = info.IsMaster
		// 由备设备探测主设备，本身为主的跳过
		if info.IsMaster {
			continue
		}
		target := fmt.Sprintf("ip://%s:%d", info.Ip, info.Port)
		if _, ok := target2Dev[target]; !ok {
			target2Dev[target] = []string{dev}
		} else {
			target2Dev[target] = append(target2Dev[target], dev)
		}
	}
	if len(target2Dev) == 0 {
		return
	}
	h.setDevStatusMap(devStatus)
	for target, devs := range target2Dev {
		go h.heartbeatLoop(ctx, target, devs)
	}
}

// IsMaster 判断设备是否为主设备
func (h *hot_standby) IsMaster(devNum string) bool {
	devIsMaster := h.GetDevStatusMap()
	if len(devIsMaster) == 0 {
		return true
	}
	if isMaster, ok := devIsMaster[devNum]; !ok {
		return true
	} else {
		return isMaster
	}
}

// GetDevStatusMap 获取采集设备的主从状态
func (h *hot_standby) GetDevStatusMap() map[string]bool {
	h.collectDevIsMasterMutex.RLock()
	defer h.collectDevIsMasterMutex.RUnlock()
	return h.collectDevIsMaster
}

// SetDevStatusMap 设置采集设备的主从状态
func (h *hot_standby) setDevStatusMap(devIsMaster map[string]bool) {
	h.collectDevIsMasterMutex.Lock()
	defer h.collectDevIsMasterMutex.Unlock()
	h.collectDevIsMaster = devIsMaster
}

// heartbeatLoop 热备心跳，由备设备定时探测主设备是否存活
func (h *hot_standby) heartbeatLoop(ctx context.Context, target string, devs []string) {
	failNum := 0
	inMaster := false
	maxFailNum := config.LoadIntOrDefault(config.GetRB().Task.Local.DetectFailNum, 3)
	intervalSecond := config.LoadIntOrDefault(config.GetRB().Task.Local.DetectInterval, 5)
	intervalSleep := time.Duration(intervalSecond) * time.Second
	for {

		select {
		case <-ctx.Done():
			log.Info("stop detect.")
			return
		default:
			// 探测主设备是否存活
			err := h.detect(target)
			if err != nil {
				failNum++
				if failNum >= maxFailNum {
					change := h.switchStatus(devs, true)
					inMaster = true
					if change {
						log.Errorf("master %+v unavailable, switch to master, failNum:%d, err:%v",
							devs, failNum, err)
					} else {
						log.Debugf("%+v still unavailable", devs)
					}
				}
			} else {
				// 如果主设备恢复，设备切换回备设备
				if inMaster {
					h.switchStatus(devs, false)
					inMaster = false
					log.Errorf("master %+v recover, switch to slave, detect failNum:%d, lasted:%ds",
						devs, failNum, failNum*intervalSecond)
				}
				failNum = 0
			}
		}
		time.Sleep(intervalSleep)
	}
}

func (h *hot_standby) switchStatus(devs []string, status bool) bool {
	hasSwitch := false
	dev2Status := h.GetDevStatusMap()
	newDev2Status := make(map[string]bool)
	for k, v := range dev2Status {
		newDev2Status[k] = v
	}
	for _, dev := range devs {
		// 已经为目标状态，不用处理
		if newDev2Status[dev] == status {
			continue
		} else {
			newDev2Status[dev] = status
			hasSwitch = true
		}
	}
	if hasSwitch {
		// 切换备设备
		h.setDevStatusMap(newDev2Status)
		// 通知任务变化
		h.notify <- true
	}
	return hasSwitch
}

func (h *hot_standby) detect(target string) error {
	timeout := time.Duration(config.LoadIntOrDefault(config.GetRB().Task.Local.DetectTimeout, 2)) * time.Second
	_, err := pb.NewBoxManagerClientProxy(client.WithTarget(target),
		client.WithTimeout(timeout),
		client.WithProtocol("http"),
		client.WithNetwork("tcp")).Heartbeat(context.Background(), &emptypb.Empty{})
	if err != nil {
		return err
	}
	return nil
}
