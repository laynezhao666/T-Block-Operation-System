package plugin

import (
	"context"
	"fmt"
	"agent/entity/definition"
	"agent/logic/collector/rtdb/model"
	"sync"

	"github.com/robfig/cron/v3"
	"trpc.group/trpc-go/trpc-go/log"
)

var (
	p = manager{
		plugins:      make(map[string]Plugin),
		subscription: make(map[EventType]map[Plugin]struct{}),
	}
)

// Plugin 插件接口
type Plugin interface {
	// Do 执行
	Do(arg interface{})
	// ProcessRtd 处理rtd
	ProcessRtd(deviceID definition.DeviceGidType, points model.DataPoints, ignore bool)
	// Notify 通知
	Notify(event EventType)
	// GetInterval 获取间隔
	GetInterval() int
}

type manager struct {
	subscription map[EventType]map[Plugin]struct{}
	plugins      map[string]Plugin
	cron         *cron.Cron
	sync.RWMutex
}

// Manager 插件管理器
func Manager() *manager {
	return &p
}

// Register 注册
func (m *manager) Register(name string, p Plugin) error {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.plugins[name]; ok {
		return fmt.Errorf("plugin %v already exists", name)
	}
	m.plugins[name] = p
	log.Infof("Plugin registered: %s", name)
	return nil
}

// Start 启动
func (m *manager) Start(ctx context.Context) {
	m.Lock()
	defer m.Unlock()

	if m.cron != nil {
		log.Warn("Plugin manager already started")
		return
	}

	m.cron = cron.New(cron.WithSeconds())
	log.Info("Initializing plugin schedules...")

	for rawName, rawPlugin := range m.plugins {
		name := rawName
		plugin := rawPlugin
		interval := plugin.GetInterval()
		if interval <= 0 {
			interval = 10 // 默认间隔
			log.Warnf("Invalid interval for %s, using default %ds", name, interval)
		}

		_, err := m.cron.AddFunc(fmt.Sprintf("0/%d * * * * *", interval), func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("Plugin %s panic: %v", name, r)
				}
			}()
			plugin.Do(nil)
		})

		if err != nil {
			log.Errorf("Failed to schedule plugin %s: %v", name, err)
		} else {
			log.Warnf("Scheduled %s with interval %ds", name, interval)
		}

	}

	m.cron.Start()
	go func() {
		<-ctx.Done()
		m.cron.Stop()
		log.Info("All plugin schedules stopped")
	}()
}

// Unregister 注销
func (m *manager) Unregister(name string) {
	if m == nil {
		return
	}
	m.Lock()
	defer m.Unlock()

	delete(m.plugins, name)
}

func (m *manager) do(arg interface{}) {
	for _, p := range m.plugins {
		go func(plugin Plugin) {
			plugin.Do(arg)
		}(p)
	}
}

func (m *manager) processRtd(deviceID definition.DeviceGidType, points model.DataPoints, ignore bool) {
	for _, p := range m.plugins {
		p.ProcessRtd(deviceID, points, ignore)
	}
}

// Do 执行
func (m *manager) Do(arg interface{}) {
	if m == nil {
		return
	}
	m.RLock()
	defer m.RUnlock()

	m.do(arg)
}

// func (m *manager) Notify(arg interface{}, names ...string) {
// 	if m == nil {
// 		return
// 	}
// 	m.RLock()
// 	defer m.RUnlock()

// 	for _, name := range names {
// 		if p, ok := m.plugins[name]; ok {
// 			go p.Do(arg)
// 		}
// 	}
// }
