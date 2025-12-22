package collector

import (
	"agent/logic/cm"
	"agent/logic/collector/device"
	"agent/logic/collector/device/driver"
	"agent/logic/collector/device/driver/drivers"
	"agent/logic/collector/device/model"
	"agent/logic/collector/device/virtualpoints"
	"agent/logic/collector/dispatcher"
	"agent/logic/plugin"
	"agent/logic/std"
	"runtime"
	"sync"

	"trpc.group/trpc-go/trpc-go/log"
)

var ch chan bool

var updateMux sync.Mutex

// Init 初始化
func Init() error {
	ch = make(chan bool)

	model.Init()
	device.Init()
	virtualpoints.Init()
	var err error
	if err = drivers.Init(); err != nil {
		return err
	}
	if err = dispatcher.Dispatcher().Load(); err != nil {
		return err
	}

	// 监听配置的变化，重新加载
	go func() {
		defer func() {
			if r := recover(); r != nil {
				stack := make([]byte, 102400)
				length := runtime.Stack(stack, true)
				log.Errorf("panic:%v,stack:%s",
					r, string(stack[:length]))
			}
		}()

		for {
			select {
			case <-cm.WatchDeviceChanged():
				log.Info("watch task changed!!!")
				ReloadAll()
			case <-cm.WatchStdConfigChangedChan():
				log.Warn("watch std changed!!!")
				ReloadStd()
			case <-cm.WatchDeviceConfigChangedChan():
				log.Warn("watch collect device changed!!!")
				ReloadCollect()
			case <-ch:
				return
			}
		}
	}()

	return nil
}

// ReloadAll 重新加载所有配置
func ReloadAll() {
	updateMux.Lock()
	defer updateMux.Unlock()
	// 重新加载配置
	if err := cm.ReInitWorker(); err != nil {
		log.Warnf("ReInit cm worker failed, %v", err)
		return
	}
	// 重新调度标准点计算
	if err := std.GetCalManager().Reload(); err != nil {
		log.Errorf("Reload std failed, %v", err)
	}
	// 重新调度采集
	if err := dispatcher.Dispatcher().Reload(); err != nil {
		log.Errorf("Reload devices failed, %v", err)
		return
	}
	plugin.Manager().Notify(plugin.EventCollectConfigChange)
	log.Warnf("all version change, reInit done!")
}

// ReloadStd 重新加载标准点配置
func ReloadStd() {
	updateMux.Lock()
	defer updateMux.Unlock()
	// 重新加载配置 todo 这里可以拆
	if err := cm.ReInitWorker(); err != nil {
		log.Warnf("ReInit cm worker failed, %v", err)
		return
	}
	// 重新调度标准点计算
	if err := std.GetCalManager().Reload(); err != nil {
		log.Errorf("Reload std failed, %v", err)
		return
	}
	log.Warnf("std version change, reInit std done!")
}

// ReloadCollect 重新加载采集配置
func ReloadCollect() {
	updateMux.Lock()
	defer updateMux.Unlock()
	// 重新加载配置 todo 这里可以拆
	if err := cm.ReInitWorker(); err != nil {
		log.Warnf("ReInit cm worker failed, %v", err)
		return
	}
	// 重新调度采集
	if err := dispatcher.Dispatcher().Reload(); err != nil {
		log.Errorf("Reload devices failed, %v", err)
		return
	}
	log.Warnf("collect device version change, reInit device and tpls done!")
}

// UnInit 清理
func UnInit() {
	ch <- true
	dispatcher.Dispatcher().Unload()
	driver.DriverManager().Close()
}
