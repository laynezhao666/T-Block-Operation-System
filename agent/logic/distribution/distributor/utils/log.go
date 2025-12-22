package utils

import (
	"sync"
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/definition"

	"agent/utils/flog"
)

var (
	manager *loggerManager
)

type KeyType = definition.DeviceGidType

func init() {
	manager = &loggerManager{
		logger: make(map[KeyType]*flog.Filter),
	}
	go manager.refresh()
}

type loggerManager struct {
	sync.Mutex
	logger map[definition.DeviceGidType]*flog.Filter
}

// GetLogger 获取日志
func GetLogger(key KeyType) *flog.Filter {
	return manager.GetLogger(key)
}

// GetLogger 获取日志
func (m *loggerManager) GetLogger(key KeyType) *flog.Filter {
	if m == nil {
		return nil
	}

	m.Lock()
	defer m.Unlock()

	l, ok := m.logger[key]
	if ok {
		return l
	}
	l = flog.NewFilterLogger(10*time.Minute, log.GetDefaultLogger())
	m.logger[key] = l
	return l
}

func (m *loggerManager) clear() {
	for _, l := range m.logger {
		l.Stop()
	}

	m.logger = make(map[KeyType]*flog.Filter)
}

func (m *loggerManager) refresh() {
	for {
		time.Sleep(time.Hour)
		m.Lock()
		m.clear()
		m.Unlock()
	}
}
