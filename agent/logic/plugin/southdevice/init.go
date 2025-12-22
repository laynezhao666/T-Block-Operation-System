package southdevice

import (
	"agent/entity/config"
	"agent/logic/plugin"

	"trpc.group/trpc-go/trpc-go/log"
)

// Init 初始化
func Init() {
	// 仅在gateway模式开启，因为根据comm异常设置测点异常功能在agent模式下有冲突
	if config.GetRB().IsGatewayMode() {
		if err := plugin.Manager().Register("southdevice", &southPlugin{}); err != nil {
			panic(err)
		}
		log.Info("south device plugin registered success")
	}
}
