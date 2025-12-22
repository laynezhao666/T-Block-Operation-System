package zero

import (
	"strconv"
	"sync"
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/definition"
	"agent/logic/collector/processor/iprocessor"

	"agent/utils/flog"
)

const (
	KeyThreshold = "zero_threshold"
)

var (
	logOnce   sync.Once
	filterLog *flog.Filter
)

type deviceType map[definition.DeviceGidType]*Processor

type manager struct {
	devices deviceType
	mutex   sync.RWMutex
}

// NewManager 新建一个处理器管理器
func NewManager() *manager {
	p := new(manager)
	p.devices = make(deviceType)

	return p
}

// GetProcessor 获取处理器
func (m *manager) GetProcessor(deviceGiD definition.DeviceGidType, extends map[string]interface{}) iprocessor.Processor {
	logOnce.Do(func() {
		filterLog = flog.NewFilterLogger(time.Minute*10, log.GetDefaultLogger())
	})

	m.mutex.RLock()
	d, ok := m.devices[deviceGiD]
	m.mutex.RUnlock()
	if ok {
		return d
	}

	threshold := 0
	tempValue, ok := extends[KeyThreshold]
	if !ok {
		// 若未设置阈值，默认为 100
		threshold = 100
	} else {
		// 若显式设置该字段且解析成功，则使用该值
		s, ok := tempValue.(string)
		if ok && len(s) > 0 {
			if t, err := strconv.Atoi(s); err == nil {
				threshold = t
			}
		}
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	if d, ok = m.devices[deviceGiD]; ok {
		return d
	}
	d = NewProcessor(threshold)
	m.devices[deviceGiD] = d
	return d
}
