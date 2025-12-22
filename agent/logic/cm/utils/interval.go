package utils

import (
	"fmt"
	"sort"
)

// Interval 表示一个 [Begin, End) 的区间
type Interval struct {
	Begin int
	End   int
}
// String 打印区间
func (i Interval) String() string {
	return fmt.Sprintf("[%v, %v)", i.Begin, i.End)
}

// verify 校验区间，不能出现起始端点与结束端点重合或者在结束端点之后的清空
func (i Interval) verify() error {
	if i.Begin >= i.End {
		return fmt.Errorf("error interval: %+v", i)
	}
	return nil
}

func (i Interval) exceedLength(maxRange int) error {
	if i.End-i.Begin > maxRange {
		return fmt.Errorf("%v 长度超过阈值 %v", i, maxRange)
	}
	return nil
}

// MergeResult 合并区间结果
type MergeResult struct {
	IndexMap  map[int]int // 原始区间的索引到合并后的区间索引的映射
	Intervals Intervals   // 合并后的区间
}

// Intervals 表示多个区间的集合
type Intervals []Interval

type intervalWithIndex struct {
	Interval Interval
	Index    int // 原始索引
}

func verifyIntervals(intervals Intervals) error {
	var err error
	for _, v := range intervals {
		if err = v.verify(); err != nil {
			return err
		}
	}
	return nil
}

// GenerateIntervals 合并区间，区间长度不超过 maxRange，结束与开始端点的距离超过 maxGap 的两个区间不会合并
func GenerateIntervals(intervals Intervals, maxRange, maxGap int) (*MergeResult, error) {
	if err := verifyIntervals(intervals); err != nil {
		return nil, err
	}

	newIntervals := make([]intervalWithIndex, len(intervals))
	for i := range intervals {
		newIntervals[i].Interval = intervals[i]
		newIntervals[i].Index = i
	}
	sort.Slice(newIntervals, func(i, j int) bool {
		if newIntervals[i].Interval.Begin < newIntervals[j].Interval.Begin {
			return true
		}
		return newIntervals[i].Interval.End < newIntervals[j].Interval.End
	})

	r, err := generateIntervals(newIntervals, maxRange, maxGap)
	return r, err
}

// generateIntervals 合并区间
func generateIntervals(intervals []intervalWithIndex, maxRange, maxGap int) (*MergeResult, error) {
	l := len(intervals)

	result := &MergeResult{
		IndexMap:  make(map[int]int),
		Intervals: nil,
	}
	if l == 0 {
		return result, nil
	}

	currentResultIndex := 0
	result.Intervals = []Interval{{Begin: intervals[0].Interval.Begin, End: intervals[0].Interval.End}}
	result.IndexMap[intervals[0].Index] = currentResultIndex
	if l == 1 {
		if err := intervals[0].Interval.exceedLength(maxRange); err != nil {
			return result, err
		}
		return result, nil
	}

	for i := 1; i < l; i++ {
		resultInterval := result.Intervals[currentResultIndex]
		// 若两区间距离超过 maxGap 或合并后区间长度超过 maxRange
		// 则不合并
		if intervals[i].Interval.Begin >= resultInterval.End+maxGap ||
			intervals[i].Interval.End > resultInterval.Begin+maxRange {
			if err := intervals[i].Interval.exceedLength(maxRange); err != nil {
				return result, err
			}
			result.Intervals = append(result.Intervals, intervals[i].Interval)
			currentResultIndex++
		} else {
			// 若合并后区间变长，则更新结束端点
			if result.Intervals[currentResultIndex].End < intervals[i].Interval.End {
				result.Intervals[currentResultIndex].End = intervals[i].Interval.End
			}
		}
		result.IndexMap[intervals[i].Index] = currentResultIndex
	}

	return result, nil
}
