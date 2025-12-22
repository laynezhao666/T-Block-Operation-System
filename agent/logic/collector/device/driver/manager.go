package driver

import (
	"errors"
	"sync"
)

var (
	manager *Manager = nil
	once    sync.Once
)

var (
	ErrorDriverNULL      = errors.New("driver is nil")        // ErrorDriverNULL 驱动为空
	ErrorDriverDuplicate = errors.New("driver is duplicated") // ErrorDriverDuplicate 驱动重复
)

func init() {
	once.Do(func() {
		manager = &Manager{
			drivers: make(MapDriver),
		}
	})
}

// MapDriver 驱动映射
type MapDriver map[string]IDriver

// Manager 管理已注册的驱动
type Manager struct {
	drivers MapDriver
	sync.RWMutex
}

// DriverManager 返回 Manager 单例
func DriverManager() *Manager {
	return manager
}

// Close 关闭所有驱动
func (m *Manager) Close() {
	if m == nil {
		return
	}
	for _, driver := range m.drivers {
		driver.UnInit()
	}
	m.drivers = nil
}

// GetDriver 获取 name 对应的驱动
func (m *Manager) GetDriver(name string) (IDriver, bool) {
	if m == nil {
		return nil, false
	}
	m.RLock()
	defer m.RUnlock()

	d, ok := m.drivers[name]
	return d, ok
}

// Register 注册名称为 name 的驱动 driver
func Register(name string, driver IDriver) error {
	if driver == nil {
		return ErrorDriverNULL
	}
	manager.Lock()
	defer manager.Unlock()

	if _, ok := manager.drivers[name]; ok {
		return ErrorDriverDuplicate
	}
	manager.drivers[name] = driver
	return nil
}
