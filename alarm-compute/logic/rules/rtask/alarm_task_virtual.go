package rtask

import (
	"fmt"
	"strconv"

	"alarm-compute/conf"
	"alarm-compute/entity/epoint"
)

// StartVirtualAnalyze StartVirtualAnalyze
func (at *AlarmTask) StartVirtualAnalyze(ts int64, pointValueMap epoint.HistoryValueMap) (float64, error) {
	var expRes float64
	// 获取当前需要的历史数据
	historyPV, err := at.checkDelayPointValue(pointValueMap)
	if err != nil {
		err = fmt.Errorf("vt checkPointValue failed; %w", err)
		return expRes, err
	}
	pt := at.geneAnalyzePointTypeMap(ts)
	result, err := pt.EvalWithIntervalPointData(historyPV)
	if err != nil {
		err = fmt.Errorf("vt eval exp failed; %w", err)
		return expRes, err
	}
	switch result := result.(type) {
	case bool:
		if result {
			expRes = 1
		} else {
			expRes = 0
		}
	case float64:
		expRes, err = strconv.ParseFloat(
			fmt.Sprintf("%.*f", conf.ServerConf.VirtualConfig.RoundPrecision, result), 64)
		if err != nil {
			err = fmt.Errorf("vt get float value failed, result: %#v", result)
			return expRes, err
		}
	default:
		err = fmt.Errorf("vt Unsupported type, result: %#v", result)
		return expRes, err
	}
	return expRes, nil
}
