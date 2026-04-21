package driver

import (
	"errors"
	"sync"

	"dac/entity/consts"
	"dac/entity/model/db"
	"dac/entity/model/driver"
)

var (
	manager *Manager = nil
	once    sync.Once
)

var (
	ErrorDriverNULL      = errors.New("driver is nil")        // ErrorDriverNULL 驱动为空
	ErrorDriverDuplicate = errors.New("driver is duplicated") // ErrorDriverDuplicate 驱动重复
)

// Driver 驱动对象
type Driver interface {
	Init() consts.Quality
	UnInit() consts.Quality
	CreateController(id db.IDType, name string) driver.Controller
}

type MapDriver map[string]Driver

// Manager 管理已注册的驱动
type Manager struct {
	drivers MapDriver
	sync.RWMutex
}

func init() {
	once.Do(func() {
		manager = &Manager{
			drivers: make(MapDriver),
		}
	})
}

// GetManager 返回 Manager 单例
func GetManager() *Manager {
	return manager
}

// GetDriver 获取 name 对应的驱动
func (m *Manager) GetDriver(name string) (Driver, bool) {
	if m == nil {
		return nil, false
	}
	m.RLock()
	defer m.RUnlock()

	d, ok := m.drivers[name]
	return d, ok
}

// Register 注册名称为 name 的驱动 driver
func Register(name string, driver Driver) error {
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
