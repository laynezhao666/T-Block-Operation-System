package change

import (
	"agent/entity/config"
	"agent/logic/collector/rtdb"
)
// Init 初始化
func Init() error {
	if config.GetRB().IsCollectReportEnable() {
		collectInstance := newCollectChangeManager()
		collectInstance.start()
		rtdb.RegisterDataPointsUpdatedFun(callback, nil)
	}
	return nil
}
// UnInit 反初始化
func UnInit() {
	if config.GetRB().IsCollectReportEnable() {
		rtdb.UnRegisterDataPointsUpdatedFun(callback)
		collectInstance.stop()
	}

}
