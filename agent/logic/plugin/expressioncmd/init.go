// Package expressioncmd 表达式命令
package expressioncmd

import (
	"agent/logic/plugin"

	"trpc.group/trpc-go/trpc-go/log"
)

// Init 初始化
func Init() {
	pluginInstance := Plugin()
	if err := plugin.Manager().Register("expression_cmd", pluginInstance); err != nil {
		log.Fatalf("Failed to register expression_cmd plugin: %v", err)
	}

	if err := plugin.Manager().Subscribe(pluginInstance, plugin.EventCollectConfigChange); err != nil {
		log.Errorf("Failed to subscribe config change: %v", err)
	}
}
