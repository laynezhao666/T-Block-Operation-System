// Package pointeval  expr dimension point eval
package pointeval

import (
	"fmt"

	"etrpc-go/log"

	"alarm-compute/entity/epoint"
	"alarm-compute/logic/point"
	"alarm-compute/utils/common"
	"alarm-compute/utils/tnql"
)

// PointFetchInfo PointFetchInfo  delay point time info
type PointFetchInfo struct {
	// 获取测点的时长
	Duration int `json:"duration,omitempty"`
	// 获取测点的间隔
	Interval int `json:"interval,omitempty"`

	// 跳变延迟信息
	RangeDelay int `json:"jpDelay,omitempty"`
}

// PointTypeMap PointTypeMap
type PointTypeMap struct {
	// 采集测点类型表达式(A+B)
	Express string `json:"express,omitempty"`
	// 变量映射 例如：{"A":["1153029394454814720.GecFireAlarm"],"B":["1153029394454814720.中压市电供电中断"]}
	// 一个变量映射为 gid.pointName列表
	// gid 在不同模组中是全局唯一的
	PMap map[string][]string `json:"pMap,omitempty"`
	// 计算表达式使用的引擎
	Engine string `json:"engine,omitempty"`
	// 拉取测点信息
	PointFetchList map[string][]PointFetchInfo `json:"-"`
	// 跳变函数拉取一段时间的配置
	JPRangeSec int `json:"-"`
}

func (pt *PointTypeMap) getGidPoint(token string) ([]string, bool) {
	points, ok := pt.PMap[token]
	return points, ok
}

// GetTokenMap GetTokenMap
func (pt *PointTypeMap) GetTokenMap() map[string]interface{} {
	tokenMap := make(map[string]interface{}, len(pt.PMap))
	for k := range pt.PMap {
		// 只需要 token，示例 {A: A}
		tokenMap[k] = k
	}

	return tokenMap
}

// UpdatePointFetchList UpdatePointFetchList
func (pt *PointTypeMap) UpdatePointFetchList() (err error) {
	expr, err := tnql.NewEvaluableExpressionWithFunctions(pt.Express, pt.GetFuncMap())
	if err != nil {
		return fmt.Errorf("new express failed; %w", err)
	}

	expr.ChecksTypes = false

	tokenMap := pt.GetTokenMap()
	expr.CustomData = map[string]interface{}{
		tnql.ExprCustomDataKeyDryRun: true,
	}
	_, err = expr.Evaluate(tokenMap)
	if err != nil {
		return fmt.Errorf("evaluate dryrun failed; %w", err)
	}

	// 处理完成 reset，后续的操作不是 dryrun
	expr.CustomData = nil

	return nil
}

// GetPointDelayMap 延迟测点 过去某个interval的测点map
// OutPut： map[interval][gid.pointName, ...]
func (pt *PointTypeMap) GetPointDelayMap(out map[int][]string, pm map[string]struct{}) {
	pointDelayMap := make(map[string][]int)
	for p, list := range pt.PointFetchList {
		if _, ok := pm[p]; !ok {
			continue
		}
		for _, info := range list {
			pointDelayMap[p] = append(pointDelayMap[p], info.Interval)
			if info.Duration != 0 {
				pointDelayMap[p] = append(pointDelayMap[p], info.Duration+info.RangeDelay)
			}
		}
	}
	for pointName, intervalList := range pointDelayMap {
		for _, interval := range intervalList {
			out[interval] = append(out[interval], pointName)
		}
	}
}

// GetPointDurationMap 延迟测点 一段持续时间的测点map
// OutPut： map[duration][gid.pointName, ...]
func (pt *PointTypeMap) GetPointDurationMap(out map[int][]string, pm map[string]struct{}) {
	pointDurationMap := make(map[string]int)
	for p, list := range pt.PointFetchList {
		if _, ok := pm[p]; !ok {
			continue
		}
		for _, info := range list {
			if info.Duration == 0 {
				// 只获取 duration 的测点
				continue
			}
			// 保留最长的 duration，减少请求次数
			if d, ok := pointDurationMap[p]; ok {
				// 已有，且当前的 info.Duration 大于旧的则更新，否则不操作
				if info.Duration > d {
					pointDurationMap[p] = info.Duration
				}
			} else {
				pointDurationMap[p] = info.Duration
			}
		}
	}
	for p, d := range pointDurationMap {
		out[d] = append(out[d], p)
	}
}

// GetPointRangeMap 跳变测点map
func (pt *PointTypeMap) GetPointRangeMap(out map[int]map[int][]string, pm map[string]struct{}) {
	// 每个测点都有 duration 和 delay
	// 同一个测点可能会有不同的 duration 和 delay，分别用于不同的策略
	for p, list := range pt.PointFetchList {
		if _, ok := pm[p]; !ok {
			continue
		}
		for _, info := range list {
			if info.Duration == 0 || info.RangeDelay == 0 {
				// 过滤不需要 duration 或 delay 为 0 的测点
				// 跳变 delay 为 0 的数据直接通过 getPointDurationMap 获取
				continue
			}
			if _, ok := out[info.RangeDelay]; !ok {
				out[info.RangeDelay] = map[int][]string{}
			}
			out[info.RangeDelay][info.Duration] = append(out[info.RangeDelay][info.Duration], p)
		}
	}
}

// @param pointValueMap { 测点: { 间隔1: 测点值 } }, { point: { interval: value } } pointName: map[interval]val
// @param pvm  v1.0 varName: map[interval]val
// v2.0 适配单变量映射为多设备测点的应用场景 varName: []map[interval]val
func (pt *PointTypeMap) transformPointValueToToken(vars []string, pointValueMap epoint.HistoryValueMap) (
	pvm epoint.SymValueMapList, err error) {
	pvm = epoint.SymValueMapList{}
	hasAnalyzeTokenSet := map[string]struct{}{}
	for _, token := range vars {
		if _, ok := hasAnalyzeTokenSet[token]; ok {
			continue
		}
		hasAnalyzeTokenSet[token] = struct{}{}
		points, ok := pt.PMap[token]
		if !ok {
			err = fmt.Errorf("token not existed, express: %s, token: %s, pm: %v", pt.Express, token, pt.PMap)
			return
		}
		for _, p := range points {
			_, vExist := pointValueMap[p]
			if !vExist {
				err = fmt.Errorf("value not existed, express: %s, point:%s, pm: %v", pt.Express, token, pointValueMap)
				return
			}
			dValue := pointValueMap[p]
			for d, v := range dValue {
				if valid, _ := point.GetPointManager().AlarmPointValueValidate(v); !valid {
					err = fmt.Errorf("value invalid, express: %v, v: %v p: %v, d: %v", pt.Express, v, p, d)
					return
				}
			}
			if len(pvm[token]) == 0 {
				pvm[token] = []epoint.IntervalMap{dValue}
			} else {
				pvm[token] = append(pvm[token], dValue)
			}
		}
	}
	return
}

// EvalWithIntervalPointData EvalWithHIntervalPointData
func (pt *PointTypeMap) EvalWithIntervalPointData(intervalPointValueMap map[string]epoint.IntervalMap) (
	result interface{}, err error) {
	defer common.CatchPanicCb(func(i interface{}) {
		err = fmt.Errorf("express: %v, intervalPointValueMap: %v, panic info: %v",
			pt.Express, intervalPointValueMap, i)
		log.Errorf(err.Error())
	})
	expr, err := tnql.NewEvaluableExpressionWithFunctions(pt.Express, pt.GetFuncMap())
	if err != nil {
		err = fmt.Errorf("NewEval failed %w", err)
		return
	}

	// 关闭校验
	expr.ChecksTypes = false

	vars := expr.Vars()
	valueMap, err := pt.transformPointValueToToken(vars, intervalPointValueMap)
	if err != nil {
		err = fmt.Errorf("transformPointValueToToken failed; %w", err)
		return
	}

	expr.CustomData = map[string]interface{}{
		tnql.ExprCustomDataKeyData: valueMap,
	}
	// tokenMap := pt.getZeroDelayPointValue(valueMap)
	tokenMap := pt.GetTokenMap()
	result, err = expr.Evaluate(tokenMap)
	if err != nil {
		err = fmt.Errorf("evaluate failed, intervalPointValueMap: %v; %w", intervalPointValueMap, err)
		return
	}

	return
}
