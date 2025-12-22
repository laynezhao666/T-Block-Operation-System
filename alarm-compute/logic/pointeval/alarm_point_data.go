package pointeval

import (
	"fmt"
	"sort"

	"github.com/samber/lo"

	"alarm-compute/entity/epoint"
	"alarm-compute/utils/tnql"
)

// checkAndGetDurationPoints 支持一般 持续时间 函数的取值
// 适用于 单变量 映射 单gid的函数场景  A : [gid1.pointName]
// @param fn extractArgsFunc
// @param needSorted bool 是否需要对测点列表进行排序，若需要排序，则按照时间由近到远排序
// 即列表最左端为离当前时间最近测点，列表最右端为最远测点
func (pt *PointTypeMap) checkAndGetDurationPoints(fn extractArgsFunc, param tnql.Parameters,
	needSorted bool, args ...interface{}) (values []float64, err error) {
	token, durationArg, intervalArg, err := fn(args...)
	if err != nil {
		err = fmt.Errorf("checkAndGetDurationPoints extractArgs failed; %w", err)
		return
	}
	if tnql.IsDryrun(param) {
		info := PointFetchInfo{Duration: durationArg, Interval: intervalArg}
		err = pt.updatePointFetchList(param, token, info)
		if err != nil {
			err = fmt.Errorf("checkAndGetDurationPoints updatePointFetchList failed; %w", err)
			return
		}
		return
	}
	intervalMapList, err := pt.getDurationPointFromCustomData(param, token, 0, durationArg)
	if err != nil || len(intervalMapList) != 1 {
		err = fmt.Errorf(
			"checkAndGetDurationPoints failed, error gidName cnt sym map , param: %+v, args: %+v; %w",
			param, args, err)
		return nil, err
	}
	points := intervalMapList[0]
	values = make([]float64, 0, len(points))
	if needSorted {
		keyList := lo.Keys(points)
		sort.Ints(keyList)
		for _, key := range keyList {
			values = append(values, points[key])
		}
	} else {
		for _, v := range points {
			values = append(values, v)
		}
	}
	return
}

// checkAndGetIntervalPoints 支持 某些特定历史时间 函数的取值
// 适用于 单变量 映射 单gid的函数场景  A : [gid1.pointName]
func (pt *PointTypeMap) checkAndGetIntervalPoints(fn extractArgsFunc, param tnql.Parameters,
	args ...interface{}) (values []float64, err error) {
	token, durationArg, intervalArg, err := fn(args...)
	if err != nil {
		err = fmt.Errorf("checkAndGetIntervalPoints extractArgs failed; %w", err)
		return
	}
	if tnql.IsDryrun(param) {
		info := PointFetchInfo{Duration: durationArg, Interval: intervalArg}
		err = pt.updatePointFetchList(param, token, info)
		if err != nil {
			err = fmt.Errorf("checkAndGetIntervalPoints updatePointFetchList failed; %w", err)
			return
		}
		return
	}
	intervalList := []int{intervalArg}
	if intervalArg != tnql.ExprNoDelay {
		intervalList = append(intervalList, tnql.ExprNoDelay)
	}
	intervalMapList, err := pt.getIntervalPointFromCustomData(param, token, intervalList)
	if err != nil || len(intervalMapList) != 1 {
		err = fmt.Errorf(
			"checkAndGetIntervalPoints failed, error gidName cnt sym map , param: %+v, args: %+v; %w",
			param, args, err)
		return nil, err
	}
	points := intervalMapList[0]
	values = make([]float64, 0, len(points))
	for _, v := range points {
		values = append(values, v)
	}
	return
}

// checkAndGetArgvPoints 支持Count CountGT CountRateGt... 类函数的取值
// 适用于 单变量 映射 多gid的函数场景  A : [gid1.pointName, gid2.pointName, ...]
func (pt *PointTypeMap) checkAndGetArgvPoints(fn extractArgvArgsFunc, param tnql.Parameters,
	args ...interface{}) (values []float64, threshold float64, err error) {
	token, th, durationArg, err := fn(args...)
	if err != nil {
		err = fmt.Errorf("extractArgs failed; %w", err)
		return
	}
	threshold = th
	if tnql.IsDryrun(param) {
		info := PointFetchInfo{Duration: durationArg, Interval: 1}
		if err = pt.updatePointFetchList(param, token, info); err != nil {
			err = fmt.Errorf("checkAndGetArgvPoints updatePointFetchList failed; %w", err)
			return
		}
		return
	}
	intervalMapList, err := pt.getDurationPointFromCustomData(param, token, 0, durationArg)
	if err != nil {
		err = fmt.Errorf(
			"checkAndGetArgvPoints failed, error occur when getDurationPointFromCustomData , param: %+v, args: %+v; %w",
			param, args, err)
		return
	}
	values = make([]float64, 0)
	for _, intervalMap := range intervalMapList {
		for _, v := range intervalMap {
			values = append(values, v)
		}
	}
	return
}

// checkAndGetArgvPoints 支持ABS函数的取值
// 适用于 单变量 映射 单gid的函数场景  A : gid.pointName
func (pt *PointTypeMap) checkAndGetSinglePoint(fn extractSingleArgsFunc, param tnql.Parameters,
	args ...interface{}) (pointVal float64, val float64, err error) {
	token, compareVal, err := fn(args...)
	if err != nil {
		err = fmt.Errorf("extractArgs failed; %w", err)
		return
	}
	if tnql.IsDryrun(param) {
		info := PointFetchInfo{Duration: 0, Interval: 1}
		if err = pt.updatePointFetchList(param, token, info); err != nil {
			err = fmt.Errorf("checkAndGetSinglePoint updatePointFetchList err:%+v, param:%+v, args:%+v",
				err, param, args)
			return
		}
		return
	}
	intervalMapList, err := pt.getDurationPointFromCustomData(param, token, 0, 0)
	if err != nil || len(intervalMapList) != 1 {
		err = fmt.Errorf(
			"checkAndGetSinglePoint failed, error gidName cnt sym map , param: %+v, args: %+v; %w",
			param, args, err)
		return
	}
	intervalMap := intervalMapList[0]
	if err != nil {
		err = fmt.Errorf("checkAndGetSinglePoint getDurationPointFromCustomData err:%+v, param:%+v, args:%+v",
			err, param, args)
		return
	}
	pointVal, ok := intervalMap[0]
	if !ok {
		err = fmt.Errorf("checkAndGetSinglePoint pointVal not found,param:%v, token:%s", param, token)
		return
	}
	val = compareVal
	return
}

// checkAndGetValueListPoint 支持sum max min等聚合类函数
// 适用于 单变量 映射 多gid的函数场景  A : [gid1.pointName, gid2.pointName, ...]
func (pt *PointTypeMap) checkAndGetValueListPoint(fn extractArgsListFunc, param tnql.Parameters,
	args ...interface{}) (values []float64, err error) {
	tokenList, durationArg, err := fn(args...)
	if err != nil {
		err = fmt.Errorf("checkAndGetValueListPoint extractArgs failed; %w", err)
		return
	}
	if tnql.IsDryrun(param) {
		info := PointFetchInfo{Duration: durationArg, Interval: 1}
		for _, token := range tokenList {
			if err = pt.updatePointFetchList(param, token, info); err != nil {
				err = fmt.Errorf("checkAndGetValueListPoint updatePointFetchList err:%+v, param:%+v, args:%+v",
					err, param, args)
				return
			}
		}
		return
	}
	values = make([]float64, 0)
	for _, token := range tokenList {
		intervalMapList, getErr := pt.getDurationPointFromCustomData(param, token, durationArg, durationArg)
		if err != nil {
			err = fmt.Errorf("checkAndGetValueListPoint getDurationPointFromCustomData err:%+v, param:%+v, args:%+v",
				getErr, param, args)
			return
		}
		for _, intervalMap := range intervalMapList {
			for _, v := range intervalMap {
				values = append(values, v)
			}
		}
	}
	return
}

// 查询token一段时间内的测点值
// 适配单变量映射为多设备测点的应用场景
// A : [gid1.pointName, gid2.pointName, ...]
// return [{0: 1.9, 5: 2.0}, {}	...]
func (pt *PointTypeMap) getDurationPointFromCustomData(param tnql.Parameters, token string,
	start, end int) ([]epoint.IntervalMap, error) {
	var err error
	customData := param.GetExpression().CustomData
	pData, pDataOk := customData[tnql.ExprCustomDataKeyData]
	log := fmt.Sprintf("gidPnt token: %s, start: %d, end: %d", token, start, end)
	if !pDataOk {
		err = fmt.Errorf("customData not found, %s, customData: %+v", log, customData)
		return nil, err
	}
	tokenValList, tokenValOk := pData.(epoint.SymValueMapList)[token]
	if !tokenValOk {
		err = fmt.Errorf("tokenVal not found, %s, customData: %+v", log, customData)
		return nil, err
	}
	res := []epoint.IntervalMap{}
	for _, tokenValMap := range tokenValList {
		resItem := epoint.IntervalMap{}
		_, pntOk := tokenValMap[end]
		if !pntOk {
			err = fmt.Errorf("point end not found, %s, pntMap: %+v", log, tokenValMap)
			return nil, err
		}
		for interval, v := range tokenValMap {
			if interval >= start && interval <= end {
				resItem[interval] = v
			}
		}
		res = append(res, resItem)
	}
	return res, nil
}

// 查询token在指定间隔内的测点值
// 适配单变量映射为多设备测点的应用场景
// A : [gid1.pointName, gid2.pointName, ...]
// return [{0: 1.9, 5: 2.0}, {}	...]
func (pt *PointTypeMap) getIntervalPointFromCustomData(param tnql.Parameters, token string,
	intervalList []int) ([]epoint.IntervalMap, error) {
	var err error
	customData := param.GetExpression().CustomData
	pData, pDataOk := customData[tnql.ExprCustomDataKeyData]
	if !pDataOk {
		err = fmt.Errorf("customData not found, token: %+v, interval: %+v, customData: %+v",
			token, intervalList, customData)
		return nil, err
	}
	tokenValList, tokenValOk := pData.(epoint.SymValueMapList)[token]
	if !tokenValOk {
		err = fmt.Errorf("tokenVal not found, token:%s, customData: %+v", token, customData)
		return nil, err
	}
	res := []epoint.IntervalMap{}
	for _, tokenValMap := range tokenValList {
		resItem := epoint.IntervalMap{}
		for _, interval := range intervalList {
			v, valOk := tokenValMap[interval]
			if !valOk {
				err = fmt.Errorf("point not found, token: %+v, interval: %+v, customData: %+v",
					token, intervalList, customData)
				return nil, err
			}
			resItem[interval] = v
		}
		res = append(res, resItem)
	}
	return res, nil
}

func (pt *PointTypeMap) updatePointFetchList(param tnql.Parameters, token string, info PointFetchInfo) error {
	if info.Interval == 0 && info.Duration == 0 && info.RangeDelay == 0 {
		return nil
	}
	points, ok := pt.getGidPoint(token)
	if !ok {
		return fmt.Errorf("token not found, token: %v, pm: %v", token, pt.PMap)
	}
	if pt.PointFetchList == nil {
		pt.PointFetchList = map[string][]PointFetchInfo{}
	}
	for _, point := range points {
		pt.PointFetchList[point] = append(pt.PointFetchList[point], info)
	}
	return nil
}
