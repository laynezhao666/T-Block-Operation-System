package rtask

import (
	"fmt"

	"alarm-compute/entity/epoint"
	"alarm-compute/logic/pointeval"
)

func (at *AlarmTask) getPointDelayMap(out map[int][]string, pm map[string]struct{}) {
	at.Exp.GetPointDelayMap(out, pm)
}

func (at *AlarmTask) getPointDurationMap(out map[int][]string, pm map[string]struct{}) {
	at.Exp.GetPointDurationMap(out, pm)
}

func (at *AlarmTask) getPointRangeMap(out map[int]map[int][]string, pm map[string]struct{}) {
	at.Exp.GetPointRangeMap(out, pm)
}

// GetDelayPoints GetDelayPoints
func (at *AlarmTask) GetDelayPoints(m *epoint.DelayPointMap) {
	pl := at.GetPointList()

	pm := make(map[string]struct{}, len(pl))
	for _, item := range pl {
		pm[item] = struct{}{}
	}

	at.getPointDelayMap(m.HPointMap, pm)

	at.getPointDurationMap(m.HDPointMap, pm)

	at.getPointRangeMap(m.HRPointMap, pm)
}

func (at *AlarmTask) checkDelayPointValue(pointValueMap epoint.HistoryValueMap) (
	epoint.HistoryValueMap, error) {

	historyPV := epoint.HistoryValueMap{}

	for p := range at.Exp.PointFetchList {

		// 返回单个测点全量的数据，有函数的表达式需要全量的数据，而不是单个点
		if _, ok := pointValueMap[p]; ok {
			historyPV[p] = pointValueMap[p]
		}
	}

	return historyPV, nil
}

// StartDelayTimeAnalyze StartDelayTimeAnalyze
func (at *AlarmTask) StartDelayTimeAnalyze(pointValueMap epoint.HistoryValueMap,
	ts int64) (epoint.HistoryValueMap, bool, error) {
	// 获取当前需要的历史数据
	historyPV, err := at.checkDelayPointValue(pointValueMap)
	if err != nil {
		err = fmt.Errorf("checkPointValue failed; %w", err)
		return nil, false, err
	}

	var bRet bool
	result, err := at.EvalDelayTime(historyPV, ts)
	if err == nil {
		if b, ok := result.(bool); !ok {
			err = fmt.Errorf("EvalExtension not bool type, result: <%T, %v>", result, result)
		} else {
			bRet = b
		}
	}
	if err != nil {
		return nil, false, err
	}
	if !bRet {
		return nil, false, nil
	}
	// TODO 告警处理
	return historyPV, true, nil
}

// CheckMissDelayPointList CheckMissDelayPointList
func (at *AlarmTask) CheckMissDelayPointList(ts int64, pointValue epoint.HistoryValueMap) (missVarList, missPointList []string) {
	for sym, points := range at.Exp.PMap {
		for _, p := range points {
			val := pointValue[p]
			missList := checkMissDelayPoints(ts, at.Exp.PointFetchList, p, val)
			if len(missList) > 0 {
				missVarList = append(missVarList, sym)
				missPointList = append(missPointList, missList...)
			}
		}
	}
	return
}

func checkMissDelayPoints(ts int64, pointFetchList map[string][]pointeval.PointFetchInfo,
	p string, val map[int]float64) (
	miss []string) {
	delayList, ok := pointFetchList[p]
	if !ok {
		return
	}
	// 跳变测点优先判断RangeDelay
	// 如果是durtaion测点，查询duration处的值是否存在
	for _, delayItem := range delayList {
		var ok bool
		var checkInterval int
		if delayItem.RangeDelay > 0 {
			checkInterval = delayItem.RangeDelay
		} else if delayItem.Duration > 0 {
			checkInterval = delayItem.Duration
		} else {
			checkInterval = delayItem.Interval
		}
		_, ok = val[checkInterval]
		if !ok {
			miss = append(miss, fmt.Sprintf("%s:%v", p, ts-int64(checkInterval)))
		}
	}
	return
}

// EvalDelayTime EvalDelayTime
func (at *AlarmTask) EvalDelayTime(vMap epoint.HistoryValueMap, ts int64) (ret interface{}, err error) {
	pt := at.geneAnalyzePointTypeMap(ts)
	return pt.EvalWithIntervalPointData(vMap)
}
