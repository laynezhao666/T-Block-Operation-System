package std

import (
	"agent/logic/collector/rtdb"
)

// Init 初始化
func Init() error {
	cal := newCalManager()
	cal.start()

	rtdb.RegisterDataPointsUpdatedFun(callback, nil)

	return nil
}

// UnInit 反初始化
func UnInit() {
	rtdb.UnRegisterDataPointsUpdatedFun(callback)

	instance.stop()
}
