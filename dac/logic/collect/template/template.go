// Package template 提供门禁驱动模板管理，缓存已创建的驱动实例。
package template

import (
	"sync"

	"dac/logic/collect/driver"
)

// m 全局模板管理器单例
var (
	m = &Manager{
		templates: make(map[string]*Template),
	}
)

// Template 驱动模板，封装驱动实例
type Template struct {
	driver driver.Driver
}

// GetDriver 获取模板关联的驱动实例
func (t *Template) GetDriver() driver.Driver {
	return t.driver
}

// Manager 驱动模板管理器，按协议名缓存模板
type Manager struct {
	templates map[string]*Template
	sync.RWMutex
}

// GetManager 获取全局模板管理器实例
func GetManager() *Manager {
	return m
}

// GetTemplate 根据协议名获取驱动模板，不存在则自动创建
func (m *Manager) GetTemplate(name string) *Template {
	if m == nil {
		return nil
	}

	m.Lock()
	defer m.Unlock()

	t, ok := m.templates[name]
	if !ok {
		d, ok := driver.GetManager().GetDriver(name)
		if !ok {
			return nil
		}

		t = new(Template)
		t.driver = d
		m.templates[name] = t
	}

	return t
}
