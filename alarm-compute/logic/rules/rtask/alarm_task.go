package rtask

import (
	"fmt"
	"strings"

	"alarm-compute/conf"
	"alarm-compute/entity"
	"alarm-compute/entity/epoint"
	"alarm-compute/logic/pointeval"
)

const (
	AlarmTaskType = iota
	RestoreTaskType
)

const (
	// gid分隔符
	gidSeparator = ";"
	// gid和测点分隔符
	gidPointSeparator = "."
	// gid和测点拼接模版
	gidPointTemplate = "%s.%s"
)

// AlarmTask 告警/恢复计算任务
type AlarmTask struct {
	RuleTask *RuleTask
	Type     int
	Exp      *pointeval.PointTypeMap
}

// NewAlarmTask NewAlarmTask
func NewAlarmTask(taskType int) *AlarmTask {
	return &AlarmTask{
		Type: taskType,
	}
}

// ServiceName ServiceName
func (at *AlarmTask) ServiceName() string {
	return "AlarmTask"
}

// analyzeInstancePoint 解析实例测点
// 将传入的 "gid1;gid2;gid3.pointName"
// 解析为gid1.pointName, gid2.pointName, gid3.pointName
func (at *AlarmTask) analyzeInstancePoint(instance string) ([]string, error) {
	if len(instance) == 0 {
		return nil, fmt.Errorf("var map gids cannot be empty")
	}
	dvcGidsPnt := strings.Split(instance, gidPointSeparator)
	if len(dvcGidsPnt) <= 1 {
		err := fmt.Errorf("gidsPoint malform, gidsPoint:%+v", instance)
		return nil, err
	}
	res := []string{}
	gidStr, pointName := dvcGidsPnt[0], dvcGidsPnt[1]
	gids := strings.Split(gidStr, gidSeparator)
	// 去重
	analyzeGidSet := map[string]struct{}{}
	for _, gid := range gids {
		if _, ok := analyzeGidSet[gid]; ok {
			continue
		}
		analyzeGidSet[gid] = struct{}{}
		res = append(res, fmt.Sprintf(gidPointTemplate, gid, pointName))
	}
	return res, nil
}

// SetExp SetAlarmExp
func (at *AlarmTask) SetExp(expression string, varGidMap *entity.VariableGidMap) error {
	if len(expression) == 0 || varGidMap == nil {
		at.Exp = nil
		return nil
	}
	pm := &pointeval.PointTypeMap{
		Express: expression,
		PMap:    map[string][]string{},
		Engine:  varGidMap.Engine,
		// TODO 决定该字段的取值
		JPRangeSec: int(conf.ServerConf.DelayTimeConfig.JPRangeSec),
	}
	pm.PointFetchList = map[string][]pointeval.PointFetchInfo{}
	for k, instancePoint := range varGidMap.ExprMap {
		points, err := at.analyzeInstancePoint(instancePoint)
		if err != nil {
			return fmt.Errorf("analyzeInstancePoint failed; %w", err)
		}
		pm.PMap[k] = points
		for _, p := range points {
			pm.PointFetchList[p] = []pointeval.PointFetchInfo{{
				Duration:   0,
				Interval:   0,
				RangeDelay: 0,
			}}
		}
	}
	at.Exp = pm
	return nil
}

// UpdatePointFetchList 检查表达式是否正常，并更新测点延时时间
func (at *AlarmTask) UpdatePointFetchList() error {
	// 指针引用，否则是直接拷贝
	return at.Exp.UpdatePointFetchList()
}

// GetMaxPointDelayMap 获取测点最大延迟秒数
func (at *AlarmTask) GetMaxPointDelayMap() map[string]int32 {
	resMap := map[string]int32{}
	for p, delayList := range at.Exp.PointFetchList {
		var pMax int
		for _, item := range delayList {
			if item.Duration+item.RangeDelay > pMax {
				pMax = item.Duration + item.RangeDelay
			}
			if item.Interval > pMax {
				pMax = item.Interval
			}
		}
		if _, ok := resMap[p]; !ok {
			resMap[p] = 0
		}
		resMap[p] = int32(pMax)
	}
	return resMap
}

// GetPointList GetPointList
func (at *AlarmTask) GetPointList() []string {
	var pointTypeList []string
	for _, points := range at.Exp.PMap {
		pointTypeList = append(pointTypeList, points...)
	}
	return pointTypeList
}

// CheckMissRealPointList CheckMissRealPointList
func (at *AlarmTask) CheckMissRealPointList(pointValue map[string]float64) (missVarList, missPointList []string) {
	for sym, points := range at.Exp.PMap {
		for _, p := range points {
			if _, ok := pointValue[p]; !ok {
				missVarList = append(missVarList, sym)
				missPointList = append(missPointList, p)
			}
		}
	}
	return
}

// StartRealTimeAnalyze StartRealTimeAnalyze
func (at *AlarmTask) StartRealTimeAnalyze(pointValueMap map[string]float64, ts int64) (map[string]float64, bool, error) {
	// 获取当前需要的测点数据
	execPV, err := at.checkRealPointValue(pointValueMap)
	if err != nil {
		err = fmt.Errorf("checkPointValue failed; %w", err)
		return nil, false, err
	}
	var bRet bool
	result, err := at.EvalRealTime(execPV, ts)
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
	return execPV, true, nil
}

// CheckMissRealPointList CheckMissRealPointList
func (at *AlarmTask) checkRealPointValue(pointValue map[string]float64) (map[string]float64, error) {
	execPV := map[string]float64{}
	for _, points := range at.Exp.PMap {
		for _, p := range points {
			if _, ok := pointValue[p]; ok {
				execPV[p] = pointValue[p]
			} else {
				return nil, fmt.Errorf("point %s not found", p)
			}
		}
	}
	return execPV, nil
}

func (at *AlarmTask) geneAnalyzePointTypeMap(ts int64) *pointeval.PointTypeMap {
	pt := &pointeval.PointTypeMap{
		Express:        at.Exp.Express,
		PMap:           at.Exp.PMap,
		Engine:         at.Exp.Engine,
		PointFetchList: at.Exp.PointFetchList,
		JPRangeSec:     at.Exp.JPRangeSec,
	}
	return pt
}

// EvalRealTime EvalRealTime
func (at *AlarmTask) EvalRealTime(pointValueMap map[string]float64, ts int64) (ret interface{}, err error) {
	pt := at.geneAnalyzePointTypeMap(ts)
	vMap := map[string]epoint.IntervalMap{}
	for p, v := range pointValueMap {
		vMap[p] = epoint.IntervalMap{
			0: v,
		}
	}
	return pt.EvalWithIntervalPointData(vMap)
}
