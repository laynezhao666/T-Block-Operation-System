package worker

import (
	"sync"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/definition"
	"agent/logic/cm"
	"agent/logic/collector/device"
)

var (
	t    *templateProtocolManager = nil
	once sync.Once
)

// TemplateProtocolManager 获取模板协议管理器
func TemplateProtocolManager() *templateProtocolManager {
	once.Do(
		func() {
			t = &templateProtocolManager{}
		},
	)
	return t
}

// MapTemplateProtocl 模板协议
type MapTemplateProtocl map[string]*device.TemplateProtocol

type templateProtocolManager struct {
}

// Release 释放
func (t *templateProtocolManager) Release() {
	if t == nil {
		return
	}
}

// GetTemplateProtocol 获取模板协议
func (t *templateProtocolManager) GetTemplateProtocol(
	templateName string, deviceGiD definition.DeviceGidType,
) *device.TemplateProtocol {
	if t == nil {
		return nil
	}
	return t.createTemplateProtocol(templateName, deviceGiD)
}

func (t *templateProtocolManager) createTemplateProtocol(
	templateName string, deviceGiD definition.DeviceGidType,
) *device.TemplateProtocol {
	if t == nil {
		return nil
	}

	template := device.NewTemplateProtocol(templateName)
	data, ok := cm.Worker().GetDeviceTemplateByGid(deviceGiD)
	if !ok {
		log.Errorf("get template %v error: not found", templateName)
		return nil
	}

	if !template.Load(data) {
		template.Unload()
		template = nil
	}
	return template
}
