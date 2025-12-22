package plugin

import "fmt"

// EventType 插件通知事件类型
type EventType int

const (
	EventCollectConfigChange EventType = iota
)

// Subscribe 订阅事件
func (m *manager) Subscribe(p Plugin, event EventType) error {
	if m == nil || p == nil {
		return fmt.Errorf("nil manager or nil plugin")
	}

	m.Lock()
	defer m.Unlock()

	set, exists := m.subscription[event]
	if !exists {
		set = make(map[Plugin]struct{})
		m.subscription[event] = set
	}

	// 检查插件是否已订阅
	if _, duplicate := set[p]; duplicate {
		return fmt.Errorf("plugin already subscribed")
	}

	// 添加订阅
	set[p] = struct{}{}
	return nil
}

// Notify 通知事件
func (m *manager) Notify(event EventType) {
	m.Lock()
	defer m.Unlock()

	plugins, ok := m.subscription[event]
	if !ok {
		return
	}
	for p := range plugins {
		p.Notify(event)
	}
}
