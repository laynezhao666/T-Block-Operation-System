package distribution

import (
	"agent/logic/distribution/change"
	httpDt "agent/logic/distribution/distributor/http"
	"agent/logic/distribution/distributor/kafka"
	"agent/logic/distribution/interval"
)

// Init 初始化
func Init() error {
	var err error
	if err = kafka.Init(); err != nil {
		return err
	}
	if err = httpDt.Init(); err != nil {
		return err
	}

	interval.Init()
	change.Init()

	return nil
}

// UnInit 反初始化
func UnInit() {
	kafka.UnInit()
	interval.UnInit()
	change.UnInit()
}
