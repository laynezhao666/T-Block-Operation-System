package pointeval

import (
	"fmt"

	"gonum.org/v1/gonum/floats"

	"alarm-compute/entity/epoint"
	"alarm-compute/utils/tnql"
)

// [dcops告警表达运算符]
// (https://doc.weixin.qq.com/sheet/e3_AK0AugZ1ACcGXMlyLY0Tx0XPnLmrd?scode=AJEAIQdfAAokPeqFYYAK0AugZ1ACc&tab=BB08J2)

const (
	intervalArgsNum int = 3

	comparatorArgsNumMin int = 3
	comparatorArgsNumMax int = 4

	// aggregation operators 参考 [Aggregation operators | Prometheus]
	// (https://prometheus.io/docs/prometheus/latest/querying/operators/#aggregation-operators)
	aggregationArgsNumMin int = 2
	aggregationArgsNumMax int = 3

	// 所有测点数据都从外部传入，外部批量获取提高性能
	// fromGetter    bool = true
	fromParameter bool = false
)

type extractArgsFunc func(args ...interface{}) (string, int, int, error)

type extractArgvArgsFunc func(args ...interface{}) (string, float64, int, error)

type extractSingleArgsFunc func(args ...interface{}) (string, float64, error)

type extractArgsListFunc func(args ...interface{}) ([]string, int, error)

// GetFuncMapWithIntervalPoints 该部分为比较函数（返回 bool 值），使用指定的时间间隔测点，一般为两个测点
func (pt *PointTypeMap) GetFuncMapWithIntervalPoints() map[string]tnql.ExpressionFunction {
	return map[string]tnql.ExpressionFunction{
		tnql.ExprFuncJP: pt.expressFuncJp,
		// tnql.ExprFuncNJP: pt.expressFuncNjp,

		tnql.ExprFuncDelayEQ:  pt.expressFuncDelayEQ,
		tnql.ExprFuncDelayNEQ: pt.expressFuncDelayNEQ,
		tnql.ExprFuncDelayLT:  pt.expressFuncDelayLT,
		tnql.ExprFuncDelayLTE: pt.expressFuncDelayLTE,
		tnql.ExprFuncDelayGT:  pt.expressFuncDelayGT,
		tnql.ExprFuncDelayGTE: pt.expressFuncDelayGTE,
	}
}

// GetAggreMapWithDurationPoints 该部分为聚合函数（返回测点聚合后的浮点数），使用持续一段时间的测点
func (pt *PointTypeMap) GetAggreMapWithDurationPoints() map[string]tnql.ExpressionFunction {
	return map[string]tnql.ExpressionFunction{
		tnql.ExprFuncSum: pt.expressFuncSum,
		tnql.ExprFuncAvg: pt.expressFuncAvg,
		tnql.ExprFuncMin: pt.expressFuncMin,
		tnql.ExprFuncMax: pt.expressFuncMax,
	}
}

// GetFuncMapWithDurationPoints 该部分为比较函数（返回 bool 值），使用持续一段时间的测点
func (pt *PointTypeMap) GetFuncMapWithDurationPoints() map[string]tnql.ExpressionFunction {
	return map[string]tnql.ExpressionFunction{
		tnql.ExprFuncNJP: pt.expressFuncNjp, // 相当于恒等于

		tnql.ExprFuncAEQ:  pt.expressFuncAEQ,
		tnql.ExprFuncANEQ: pt.expressFuncANEQ,
		tnql.ExprFuncAGT:  pt.expressFuncAGT,
		tnql.ExprFuncAGTE: pt.expressFuncAGTE,
		tnql.ExprFuncALT:  pt.expressFuncALT,
		tnql.ExprFuncALTE: pt.expressFuncALTE,

		tnql.ExprFuncSumEQ:  pt.expressFuncSumEQ,
		tnql.ExprFuncSumNEQ: pt.expressFuncSumNEQ,
		tnql.ExprFuncSumGT:  pt.expressFuncSumGT,
		tnql.ExprFuncSumGTE: pt.expressFuncSumGTE,
		tnql.ExprFuncSumLT:  pt.expressFuncSumLT,
		tnql.ExprFuncSumLTE: pt.expressFuncSumLTE,
		tnql.ExprFuncAvgEQ:  pt.expressFuncAvgEQ,
		tnql.ExprFuncAvgNEQ: pt.expressFuncAvgNEQ,
		tnql.ExprFuncAvgGT:  pt.expressFuncAvgGT,
		tnql.ExprFuncAvgGTE: pt.expressFuncAvgGTE,
		tnql.ExprFuncAvgLT:  pt.expressFuncAvgLT,
		tnql.ExprFuncAvgLTE: pt.expressFuncAvgLTE,
		tnql.ExprFuncMinEQ:  pt.expressFuncMinEQ,
		tnql.ExprFuncMinNEQ: pt.expressFuncMinNEQ,
		tnql.ExprFuncMinGT:  pt.expressFuncMinGT,
		tnql.ExprFuncMinGTE: pt.expressFuncMinGTE,
		tnql.ExprFuncMinLT:  pt.expressFuncMinLT,
		tnql.ExprFuncMinLTE: pt.expressFuncMinLTE,
		tnql.ExprFuncMaxEQ:  pt.expressFuncMaxEQ,
		tnql.ExprFuncMaxNEQ: pt.expressFuncMaxNEQ,
		tnql.ExprFuncMaxGT:  pt.expressFuncMaxGT,
		tnql.ExprFuncMaxGTE: pt.expressFuncMaxGTE,
		tnql.ExprFuncMaxLT:  pt.expressFuncMaxLT,
		tnql.ExprFuncMaxLTE: pt.expressFuncMaxLTE,

		tnql.ExprFuncDeltaGT:  pt.expressFuncDeltaGT,
		tnql.ExprFuncDeltaGTE: pt.expressFuncDeltaGTE,
		tnql.ExprFuncDeltaEQ:  pt.expressFuncDeltaEQ,
		tnql.ExprFuncDeltaNEQ: pt.expressFuncDeltaNEQ,
		tnql.ExprFuncDeltaLT:  pt.expressFuncDeltaLT,
		tnql.ExprFuncDeltaLTE: pt.expressFuncDeltaLTE,

		tnql.ExprFuncCountRateGT:  pt.expressFuncCountRateGT,
		tnql.ExprFuncCountRateGTE: pt.expressFuncCountRateGTE,
		tnql.ExprFuncCountRateLT:  pt.expressFuncCountRateLT,
		tnql.ExprFuncCountRateLTE: pt.expressFuncCountRateLTE,
		tnql.ExprFuncCountRateEQ:  pt.expressFuncCountRateEQ,
		tnql.ExprFuncCountRateNEQ: pt.expressFuncCountRateNEQ,

		tnql.ExprFuncCountGT:  pt.expressFuncCountGT,
		tnql.ExprFuncCountGTE: pt.expressFuncCountGTE,
		tnql.ExprFuncCountLT:  pt.expressFuncCountLT,
		tnql.ExprFuncCountLTE: pt.expressFuncCountLTE,
		tnql.ExprFuncCountEQ:  pt.expressFuncCountEQ,
		tnql.ExprFuncCountNEQ: pt.expressFuncCountNEQ,

		tnql.ExprFuncAbsValue: pt.expressFuncAbsValue,
		tnql.ExprFuncAvgValue: pt.expressFuncAvgValue,
		tnql.ExprFuncMaxValue: pt.expressFuncMaxValue,
		tnql.ExprFuncMinValue: pt.expressFuncMinValue,
		tnql.ExprFuncSumValue: pt.expressFuncSumValue,
	}
}

// GetFuncMap 返回全量的函数
func (pt *PointTypeMap) GetFuncMap() map[string]tnql.ExpressionFunction {
	all := map[string]tnql.ExpressionFunction{}

	for k, v := range pt.GetFuncMapWithIntervalPoints() {
		all[k] = v
	}

	for k, v := range pt.GetFuncMapWithDurationPoints() {
		all[k] = v
	}

	for k, v := range pt.GetAggreMapWithDurationPoints() {
		all[k] = v
	}

	return all
}

// 以下为 aggregation functions 聚合函数，返回 float64 值

func (pt *PointTypeMap) expressFuncSum(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalAggregationFunc(func(f1 []float64) float64 {
		return floats.Sum(f1)
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncAvg(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalAggregationFunc(func(f1 []float64) float64 {
		return f64Avg(f1)
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncMin(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalAggregationFunc(func(f1 []float64) float64 {
		return floats.Min(f1)
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncMax(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalAggregationFunc(func(f1 []float64) float64 {
		return floats.Max(f1)
	}, param, fromParameter, args...)
}

// 以下为 compartor functions 比较函数，返回 bool 值，所有测点进行判断

// ----- Sum

func (pt *PointTypeMap) expressFuncSumEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Sum(f1) == f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncSumNEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Sum(f1) != f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncSumGT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Sum(f1) > f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncSumGTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Sum(f1) >= f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncSumLT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Sum(f1) < f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncSumLTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Sum(f1) <= f2
	}, param, fromParameter, args...)
}

// ----- Average

func f64Avg(f []float64) float64 {
	return floats.Sum(f) / float64(len(f))
}

func (pt *PointTypeMap) expressFuncAvgEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return f64Avg(f1) == f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncAvgNEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return f64Avg(f1) != f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncAvgGT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return f64Avg(f1) > f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncAvgGTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return f64Avg(f1) >= f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncAvgLT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return f64Avg(f1) < f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncAvgLTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return f64Avg(f1) <= f2
	}, param, fromParameter, args...)
}

// ----- Min

func (pt *PointTypeMap) expressFuncMinEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Min(f1) == f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncMinNEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Min(f1) != f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncMinGT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Min(f1) > f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncMinGTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Min(f1) >= f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncMinLT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Min(f1) < f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncMinLTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Min(f1) <= f2
	}, param, fromParameter, args...)
}

// ----- Max

func (pt *PointTypeMap) expressFuncMaxEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Max(f1) == f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncMaxNEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Max(f1) != f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncMaxGT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Max(f1) > f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncMaxGTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Max(f1) >= f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncMaxLT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Max(f1) < f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncMaxLTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointListFunc(func(f1 []float64, f2 float64) bool {
		return floats.Max(f1) <= f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncDeltaGT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalSortPointListFunc(func(f1 []float64, f2 float64) bool {
		// 这里就只用理会首尾（Td-T0测点值的差）确保f1的测点已排序，下标0表示T0测点
		fLen := len(f1)
		if fLen == 0 {
			return false
		}
		return f1[0]-f1[fLen-1] > f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncDeltaGTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalSortPointListFunc(func(f1 []float64, f2 float64) bool {
		// 这里就只用理会首尾（Td-T0测点值的差）, 确保f1的测点已排序，下标0表示T0测点
		fLen := len(f1)
		if fLen == 0 {
			return false
		}
		return f1[0]-f1[fLen-1] >= f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncDeltaEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalSortPointListFunc(func(f1 []float64, f2 float64) bool {
		// 这里就只用理会首尾（Td-T0测点值的差）, 确保f1的测点已排序，下标0表示T0测点
		fLen := len(f1)
		if fLen == 0 {
			return false
		}
		return f1[0]-f1[fLen-1] == f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncDeltaNEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalSortPointListFunc(func(f1 []float64, f2 float64) bool {
		// 这里就只用理会首尾（Td-T0测点值的差）, 确保f1的测点已排序，下标0表示T0测点
		fLen := len(f1)
		if fLen == 0 {
			return false
		}
		return f1[0]-f1[fLen-1] != f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncDeltaLT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalSortPointListFunc(func(f1 []float64, f2 float64) bool {
		// 这里就只用理会首尾（Td-T0测点值的差）, 确保f1的测点已排序，下标0表示T0测点
		fLen := len(f1)
		if fLen == 0 {
			return false
		}
		return f1[0]-f1[fLen-1] < f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncDeltaLTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalSortPointListFunc(func(f1 []float64, f2 float64) bool {
		// 这里就只用理会首尾（Td-T0测点值的差）, 确保f1的测点已排序，下标0表示T0测点
		fLen := len(f1)
		if fLen == 0 {
			return false
		}
		return f1[0]-f1[fLen-1] <= f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncCountRateGTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalArgvPointListFunc(func(f1 []float64, f2 float64) float64 {
		okCount, totalCount := 0, 0
		for _, fVal := range f1 {
			if fVal >= f2 {
				okCount++
			}
			totalCount++
		}
		return float64(okCount) / float64(totalCount)
	}, param, args...)
}

func (pt *PointTypeMap) expressFuncCountRateGT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalArgvPointListFunc(func(f1 []float64, f2 float64) float64 {
		okCount, totalCount := 0, 0
		for _, fVal := range f1 {
			if fVal > f2 {
				okCount++
			}
			totalCount++
		}
		return float64(okCount) / float64(totalCount)
	}, param, args...)
}

func (pt *PointTypeMap) expressFuncCountRateLTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalArgvPointListFunc(func(f1 []float64, f2 float64) float64 {
		okCount, totalCount := 0, 0
		for _, fVal := range f1 {
			if fVal <= f2 {
				okCount++
			}
			totalCount++
		}
		return float64(okCount) / float64(totalCount)
	}, param, args...)
}

func (pt *PointTypeMap) expressFuncCountRateLT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalArgvPointListFunc(func(f1 []float64, f2 float64) float64 {
		okCount, totalCount := 0, 0
		for _, fVal := range f1 {
			if fVal < f2 {
				okCount++
			}
			totalCount++
		}
		return float64(okCount) / float64(totalCount)
	}, param, args...)
}

func (pt *PointTypeMap) expressFuncCountRateEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalArgvPointListFunc(func(f1 []float64, f2 float64) float64 {
		okCount, totalCount := 0, 0
		for _, fVal := range f1 {
			if fVal == f2 {
				okCount++
			}
			totalCount++
		}
		return float64(okCount) / float64(totalCount)
	}, param, args...)
}

func (pt *PointTypeMap) expressFuncCountRateNEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalArgvPointListFunc(func(f1 []float64, f2 float64) float64 {
		okCount, totalCount := 0, 0
		for _, fVal := range f1 {
			if fVal != f2 {
				okCount++
			}
			totalCount++
		}
		return float64(okCount) / float64(totalCount)
	}, param, args...)
}

func (pt *PointTypeMap) expressFuncCountGTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalArgvPointListFunc(func(f1 []float64, f2 float64) float64 {
		okCount := 0
		for _, fVal := range f1 {
			if fVal >= f2 {
				okCount++
			}
		}
		// 为了统一evalArgvPointListFunc，数量使用float64类型
		return float64(okCount)
	}, param, args...)
}

func (pt *PointTypeMap) expressFuncCountGT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalArgvPointListFunc(func(f1 []float64, f2 float64) float64 {
		okCount := 0
		for _, fVal := range f1 {
			if fVal > f2 {
				okCount++
			}
		}
		// 为了统一evalArgvPointListFunc，数量使用float64类型
		return float64(okCount)
	}, param, args...)
}

func (pt *PointTypeMap) expressFuncCountLTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalArgvPointListFunc(func(f1 []float64, f2 float64) float64 {
		okCount := 0
		for _, fVal := range f1 {
			if fVal <= f2 {
				okCount++
			}
		}
		// 为了统一evalArgvPointListFunc，数量使用float64类型
		return float64(okCount)
	}, param, args...)
}

func (pt *PointTypeMap) expressFuncCountLT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalArgvPointListFunc(func(f1 []float64, f2 float64) float64 {
		okCount := 0
		for _, fVal := range f1 {
			if fVal < f2 {
				okCount++
			}
		}
		// 为了统一evalArgvPointListFunc，数量使用float64类型
		return float64(okCount)
	}, param, args...)
}

func (pt *PointTypeMap) expressFuncCountEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalArgvPointListFunc(func(f1 []float64, f2 float64) float64 {
		okCount := 0
		for _, fVal := range f1 {
			if fVal == f2 {
				okCount++
			}
		}
		// 为了统一evalArgvPointListFunc，数量使用float64类型
		return float64(okCount)
	}, param, args...)
}

func (pt *PointTypeMap) expressFuncCountNEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalArgvPointListFunc(func(f1 []float64, f2 float64) float64 {
		okCount := 0
		for _, fVal := range f1 {
			if fVal != f2 {
				okCount++
			}
		}
		// 为了统一evalArgvPointListFunc，数量使用float64类型
		return float64(okCount)
	}, param, args...)
}

func (pt *PointTypeMap) expressFuncAbsValue(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalSingleFloatFunc(func(f1 float64) float64 {
		absVal := f1
		if absVal < 0 {
			absVal = -absVal
		}
		return absVal
	}, param, args...)
}

func (pt *PointTypeMap) expressFuncAvgValue(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalValueAggregationFunc(func(f1 []float64) float64 {
		return f64Avg(f1)
	}, param, args...)
}

func (pt *PointTypeMap) expressFuncMaxValue(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalValueAggregationFunc(func(f1 []float64) float64 {
		return floats.Max(f1)
	}, param, args...)
}

func (pt *PointTypeMap) expressFuncMinValue(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalValueAggregationFunc(func(f1 []float64) float64 {
		return floats.Min(f1)
	}, param, args...)
}

func (pt *PointTypeMap) expressFuncSumValue(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalValueAggregationFunc(func(f1 []float64) float64 {
		return floats.Sum(f1)
	}, param, args...)
}

// ----- Check All Points

// expressFuncAEQ 恒等于 all equal， aeq(point, value, duration, interval)
// point - 测点； value - 测点值； duration - 持续时间； interval - 采样间隔（可选），默认 duration <= 60 为 1，duration > 60 为 20
func (pt *PointTypeMap) expressFuncAEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalAllFunc(func(f1, f2 float64) bool {
		return f1 == f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncANEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalAllFunc(func(f1, f2 float64) bool {
		return f1 != f2
	}, param, fromParameter, args...)
}

func (pt *PointTypeMap) expressFuncAGT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalAllFunc(func(f1, f2 float64) bool {
		return f1 > f2
	}, param, fromParameter, args...)
}
func (pt *PointTypeMap) expressFuncAGTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalAllFunc(func(f1, f2 float64) bool {
		return f1 >= f2
	}, param, fromParameter, args...)
}
func (pt *PointTypeMap) expressFuncALT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalAllFunc(func(f1, f2 float64) bool {
		return f1 < f2
	}, param, fromParameter, args...)
}
func (pt *PointTypeMap) expressFuncALTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalAllFunc(func(f1, f2 float64) bool {
		return f1 <= f2
	}, param, fromParameter, args...)
}

// ----- Delay 延迟

func (pt *PointTypeMap) expressFuncDelayEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointIntervalListFunc(func(f1 []float64, f2 float64) bool {
		for _, fVal := range f1 {
			if fVal == f2 {
				continue
			} else {
				return false
			}
		}
		return true
	}, param, fromParameter, args...)
}
func (pt *PointTypeMap) expressFuncDelayNEQ(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointIntervalListFunc(func(f1 []float64, f2 float64) bool {
		for _, fVal := range f1 {
			if fVal != f2 {
				continue
			} else {
				return false
			}
		}
		return true
	}, param, fromParameter, args...)
}
func (pt *PointTypeMap) expressFuncDelayGT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointIntervalListFunc(func(f1 []float64, f2 float64) bool {
		for _, fVal := range f1 {
			if fVal > f2 {
				continue
			} else {
				return false
			}
		}
		return true
	}, param, fromParameter, args...)
}
func (pt *PointTypeMap) expressFuncDelayGTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointIntervalListFunc(func(f1 []float64, f2 float64) bool {
		for _, fVal := range f1 {
			if fVal >= f2 {
				continue
			} else {
				return false
			}
		}
		return true
	}, param, fromParameter, args...)
}
func (pt *PointTypeMap) expressFuncDelayLT(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointIntervalListFunc(func(f1 []float64, f2 float64) bool {
		for _, fVal := range f1 {
			if fVal < f2 {
				continue
			} else {
				return false
			}
		}
		return true
	}, param, fromParameter, args...)
}
func (pt *PointTypeMap) expressFuncDelayLTE(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.evalPointIntervalListFunc(func(f1 []float64, f2 float64) bool {
		for _, fVal := range f1 {
			if fVal <= f2 {
				continue
			} else {
				return false
			}
		}
		return true
	}, param, fromParameter, args...)
}

// ----- Jump 跳变

func (pt *PointTypeMap) expressFuncNjp(param tnql.Parameters, args ...interface{}) (interface{}, error) {
	return pt.expressFuncAEQ(param, args...)
}

func (pt *PointTypeMap) extractJPArgs(args ...interface{}) (
	point string, judgement float64, delaySec int, err error) {
	if len(args) < tnql.JpArgsNum {
		err = fmt.Errorf("bad args num, args: %+v", args)
		return
	}

	point, judgement, delaySec = args[0].(string), args[1].(float64), int(args[2].(float64))
	return
}

/*
跳变逻辑优化1
优化前：每5秒执行一次，没有执行的5秒期间如果有跳变，会错过
优化后：每5秒执行一次，获取30s前的数据加入计算，没有执行的5秒期间如果有跳变，也可以覆盖

跳变逻辑优化2
优化前：虚拟软点为按需实时计算数据，用于跳变时，会请求30秒数据（每秒一次），一个设备测点请求30次，
假设有100个设备，则总请求次数 30 * 100 = 3000次
优化后：
1. 从外部合并请求，批量获取测点后传递进跳变函数
2. 50个设备测点请求合并成一个请求，100个设备测点请求2次，跳变请求30秒数据，则总请求次数 30 * 2 = 60次
ps: 注意，每隔5秒执行一次，一分钟执行12次，如果一次是 3000 个请求，则一分钟是 3000 * 12 = 36000 个请求。
目前清城52正常请求量一分钟不超过 2W 个
*/

// expressFuncJp 跳变
// jp(dvcTypeEn, delay, value)
//
//	跳变：取T0与Td秒前、Td+1秒前3个时刻的数据判断
//	1. T0和Td判断是否和 value 【相等】
//	2. Td+1判断是否和 value 【不相等】
//	对于函数的秒级告警，实现的方案是秒级周期性执行（通常为5秒）。在这种情况下，跳变会遇到一个问题
//	假设测点P从 0~5 秒为 1，6~30秒 为 0，配置了 P->0:10 的跳变
//	Td+1	Td	    T0  	时间
//	1(0)	1(1)	0(11)	第11秒（括号内为时间） T0 == 0，但是 Td != 0 不触发告警
//	1(5)	0(6)	0(16)	第16秒（括号内为时间） T0 == 0 Td == 0 且 Td+1 != 0 触发告警
//	上述为理想情况，如果是从第12秒开始，那就会错过，导致永远不会触发告警
//	Td+1	Td	    T0  	时间
//	1(1)	1(2)	0(12)	第12秒（括号内为时间） T0 == 0 但是 Td != 0 不触发告警
//	0(6)	0(7)	0(17)	第16秒（括号内为时间） T0 == 0 Td == 0 但是 Td+1 == 0 ，不触发告警
//	如果是分钟级，可以在一分钟内多次判断（如设置为20秒一次），但对于秒级就会错过告警
//	解决方案：
//	对于秒级告警，由于是秒级周期性执行（通常为5秒），可以拉取一分钟的数据，对60秒的每一秒数据进行执行该判断，避免错过。
//	具体实现为
//	1. 获取每个时间点(T0, Td, Td+1)的数据，改为获取每个时间段(T0+1min, Td+1min, Td+1+1min)的数据
//	1. 按时间点循环判断跳变的逻辑
//	存在的问题：
//	由于多拉取一分钟数据，如果满足了跳变的规则，则最长会在一分钟内一直是跳变的状态。
//	如果触发表达式配置了跳变，可以及时触发告警，且最长会在一分钟内持续告警，恢复会延缓。
//	如果恢复表达式配置了跳变，可以及时恢复告警，且最长会在一分钟内持续恢复。
//	由于判断告警的逻辑是判断触发，再判断恢复，此时已进入恢复逻辑，表明触发不满足，不会延缓触发告警
func (pt *PointTypeMap) expressFuncJp(param tnql.Parameters, args ...interface{}) (ret interface{}, err error) {
	jpParams, err := pt.checkExpressFuncJp(param, args...)
	if err != nil {
		err = fmt.Errorf("checkExpressFuncJp failed; %w", err)
		return
	}
	if jpParams.IsDryrun {
		return true, nil
	}
	pntToken, gidPntName, judgement, delaySec, jpRange :=
		jpParams.PntToken, jpParams.GidPntName, jpParams.Judgement, jpParams.DelaySec, jpParams.JpRange
	t0DataList, err := pt.getDurationPointFromCustomData(param, pntToken, 0, jpRange)
	if err != nil || len(t0DataList) != 1 {
		err = fmt.Errorf("getDurationPointFromCustomData failed, start: 0, param: %+v, args: %+v; %w", param, args, err)
		return nil, err
	}
	t0Data := t0DataList[0]
	tdDataList, err := pt.getDurationPointFromCustomData(param, pntToken, delaySec, jpRange+delaySec)
	if err != nil || len(tdDataList) != 1 {
		err = fmt.Errorf("getDurationPointFromCustomData failed, start: %d, param: %+v, args: %+v; %w",
			delaySec, param, args, err)
		return nil, err
	}
	tdData := tdDataList[0]
	// 按时间点循环判断跳变的逻辑，触发时间是按 EvalTime 来记录，需要从最近的时间开始判断
	fire, allFailed := false, true
	for t := 0; t < jpRange; t++ {
		fire, err = pt.evalJP(t, int(delaySec), judgement, t0Data, tdData)
		if err != nil {
			// 由于是一段时间的点，有某个点获取不到，直接 continue，避免第一个点有问题直接返回 err
			continue
		}
		allFailed = false
		if fire {
			break
		}
	}
	if allFailed {
		err = fmt.Errorf("evalJP all failed, gidPntName: %v; %w", gidPntName, err)
		return
	}
	return fire, nil
}

// JPFuncParams JPFuncParams
type JPFuncParams struct {
	PntToken   string
	GidPntName string
	Judgement  float64
	DelaySec   int
	JpRange    int
	IsDryrun   bool
}

func (pt *PointTypeMap) checkExpressFuncJp(param tnql.Parameters, args ...interface{}) (
	jpParams JPFuncParams, err error) {
	jpRange := pt.JPRangeSec
	if jpRange == 0 {
		jpRange = 30
	}
	pntToken, judgement, delaySec, err := pt.extractJPArgs(args...)
	if err != nil {
		err = fmt.Errorf("extractJPArgs failed, args: %+v; %w", args, err)
		return
	}
	if tnql.IsDryrun(param) {
		jpParams.IsDryrun = true
		info := PointFetchInfo{Duration: jpRange, RangeDelay: int(delaySec)}
		e := pt.updatePointFetchList(param, pntToken, info)
		if e != nil {
			err = fmt.Errorf("updatePointFetchList failed; %w", e)
			return
		}

		return
	}
	// pt.PMap已经是把deviceType映射成具体gid了，参考alarm_task_extension_test.go中的PMap数据格式
	gidPnts, ok := pt.getGidPoint(pntToken)
	if !ok {
		err = fmt.Errorf("pntToken not found, pntToken: %v, args: %+v", pntToken, args)
		return
	}

	jpParams = JPFuncParams{
		PntToken:   pntToken,
		GidPntName: gidPnts[0],
		Judgement:  judgement,
		DelaySec:   delaySec,
		JpRange:    jpRange,
	}

	return
}

// t0 的数据是 [0, duration]，31个数据，td 的数据是 [delay+1, delay + duration]，30个数据
// 例如 delay 是 70，duration 是 30，则获取的是 [71, 100]，需要 +1 才能对上
// 实际上 t0 获取的数据应该为 [0, duration - 1]，这样 t0 和 td 都是30个数据，才能对得上
// 跳变本身是往前获取了 duration 的时间来覆盖，所以少一个不影响
func (pt *PointTypeMap) evalJP(i int, delaySec int, judgement float64, t0Data, tdData epoint.IntervalMap) (
	fire bool, err error) {
	t0 := i
	td := i + delaySec
	td1 := i + delaySec + 1

	t0Value, t0OK := t0Data[t0]
	tdValue, tdOK := tdData[td]
	td1Value, td1OK := tdData[td1]
	if !t0OK || !tdOK || !td1OK {
		err = fmt.Errorf("not point get point data, t0: %v, td: %v, td1: %v",
			t0, td, td1)
		return false, err
	}

	if t0Value == judgement && tdValue == judgement && td1Value != judgement {
		// 有一个满足，则返回
		fire = true
	}

	return
}

func (pt *PointTypeMap) extractIntervalArgs(args ...interface{}) (point string, duration, interval int, err error) {
	if len(args) != intervalArgsNum {
		err = fmt.Errorf("bad args num, args: %+v", args)
		return
	}

	// 没有 duration
	point, interval = args[0].(string), int(args[2].(float64))

	return
}

func (pt *PointTypeMap) extractComparatorArgs(args ...interface{}) (point string, duration, interval int, err error) {
	if len(args) < comparatorArgsNumMin {
		err = fmt.Errorf("bad args num, args: %+v", args)
		return
	}

	point, duration = args[0].(string), int(args[2].(float64))

	if len(args) == comparatorArgsNumMax {
		interval = int(args[comparatorArgsNumMax-1].(float64))
	}

	return
}

func (pt *PointTypeMap) extractComparatorArgvArgs(args ...interface{}) (point string,
	threshold float64, duration int, err error) {
	argsLen := len(args)
	if argsLen == 0 || argsLen < comparatorArgsNumMin {
		err = fmt.Errorf("bad args num, args: %+v", args)
		return
	}
	// argsLen-1是duration；argsLen-2是threshold
	point = args[0].(string)
	if fThreshold, fOk := args[argsLen-2].(float64); fOk {
		threshold = fThreshold
	} else if iThreshold, iOk := args[argsLen-2].(int); iOk {
		threshold = float64(iThreshold)
	}
	if iDuration, iOk := args[argsLen-1].(int); iOk {
		duration = iDuration
	} else if fDuration, fOk := args[argsLen-1].(float64); fOk {
		duration = int(fDuration)
	}
	return
}
func (pt *PointTypeMap) extractComparatorSingleArgs(args ...interface{}) (point string, value float64, err error) {
	argsLen := len(args)
	if argsLen == 0 || argsLen < aggregationArgsNumMin {
		err = fmt.Errorf("bad args num, args: %+v", args)
		return
	}

	point = args[0].(string)
	value = args[argsLen-1].(float64)

	return
}

func (pt *PointTypeMap) extractDefaultArgs(args ...interface{}) (point string, duration, interval int, err error) {
	if len(args) < aggregationArgsNumMin {
		err = fmt.Errorf("bad args num, args: %+v", args)
		return
	}

	point, duration = args[0].(string), int(args[1].(float64))

	if len(args) == aggregationArgsNumMax {
		interval = int(args[aggregationArgsNumMax-1].(float64))
	}

	return
}

func (pt *PointTypeMap) extractValueArgList(args ...interface{}) (points []string, durationArg int, err error) {
	if len(args) < aggregationArgsNumMin {
		err = fmt.Errorf("extractValueArgList, args:%+v", args)
		return
	}

	argLen := len(args)
	for i, arg := range args {
		if i == argLen-1 {
			durationArg = int(arg.(float64))
		} else {
			points = append(points, arg.(string))
		}
	}

	return
}

func (pt *PointTypeMap) evalAllFunc(fn func(float64, float64) bool, param tnql.Parameters,
	isFromGetter bool, args ...interface{}) (interface{}, error) {
	values, err := pt.checkAndGetDurationPoints(pt.extractComparatorArgs, param, false, args...)
	if err != nil {
		err := fmt.Errorf("checkAndGetDurationPoints failed; %w", err)
		return false, err
	}

	if tnql.IsDryrun(param) {
		return true, nil
	}

	valueArg := args[1].(float64)

	for _, v := range values {
		if fn(v, valueArg) {
			continue
		}

		return false, nil
	}

	return true, nil
}

// evalPointIntervalListFunc 通过函数按测点间隔列表计算数值
func (pt *PointTypeMap) evalPointIntervalListFunc(fn func([]float64, float64) bool, param tnql.Parameters,
	isFromGetter bool, args ...interface{}) (interface{}, error) {
	values, err := pt.checkAndGetIntervalPoints(pt.extractIntervalArgs, param, args...)
	if err != nil {
		err := fmt.Errorf("checkAndGetDurationPoints failed; %w", err)
		return false, err
	}

	if tnql.IsDryrun(param) {
		return true, nil
	}

	valueArg := args[1].(float64)
	return fn(values, valueArg), nil
}

// evalPointListFunc 通过函数按测点列表计算数值
func (pt *PointTypeMap) evalPointListFunc(fn func([]float64, float64) bool, param tnql.Parameters,
	isFromGetter bool, args ...interface{}) (interface{}, error) {
	values, err := pt.checkAndGetDurationPoints(pt.extractComparatorArgs, param, false, args...)
	if err != nil {
		err := fmt.Errorf("evalPointListFunc checkAndGetDurationPoints failed; %w", err)
		return false, err
	}
	if tnql.IsDryrun(param) {
		return true, nil
	}
	valueArg := args[1].(float64)
	return fn(values, valueArg), nil
}

// evalSortPointListFunc 对测点列表进行排序，[0, 1, 2, 3]，其中，左端点为低延时测点值（eg：当前值）。
func (pt *PointTypeMap) evalSortPointListFunc(fn func([]float64, float64) bool, param tnql.Parameters,
	isFromGetter bool, args ...interface{}) (interface{}, error) {
	values, err := pt.checkAndGetDurationPoints(pt.extractComparatorArgs, param, true, args...)
	if err != nil {
		err := fmt.Errorf("evalSortPointListFunc checkAndGetDurationPoints failed; %w", err)
		return false, err
	}
	if tnql.IsDryrun(param) {
		return true, nil
	}
	valueArg := args[1].(float64)
	return fn(values, valueArg), nil
}

// evalArgvPointListFunc evalArgvPointListFunc
func (pt *PointTypeMap) evalArgvPointListFunc(fn func([]float64, float64) float64, param tnql.Parameters,
	args ...interface{}) (interface{}, error) {
	values, threshold, err := pt.checkAndGetArgvPoints(pt.extractComparatorArgvArgs, param, args...)
	if err != nil {
		err := fmt.Errorf("checkAndGetDurationPoints failed; %w", err)
		return false, err
	}

	if tnql.IsDryrun(param) {
		return true, nil
	}

	return fn(values, threshold), nil
}

// evalSingleFloatFunc 通过函数按测点列表计算数值
func (pt *PointTypeMap) evalSingleFloatFunc(fn func(float64) float64, param tnql.Parameters,
	args ...interface{}) (interface{}, error) {
	tokenVal, _, err := pt.checkAndGetSinglePoint(pt.extractComparatorSingleArgs, param, args...)
	if err != nil {
		err := fmt.Errorf("checkAndGetDurationPoints failed; %w", err)
		return false, err
	}

	if tnql.IsDryrun(param) {
		return true, nil
	}

	return fn(tokenVal), nil
}

// evalAggregationFunc 计算聚合函数，示例 sum(A, duration, [interval]), interval 可选
// 时间维度
func (pt *PointTypeMap) evalAggregationFunc(fn func([]float64) float64, param tnql.Parameters,
	isFromGetter bool, args ...interface{}) (interface{}, error) {
	values, err := pt.checkAndGetDurationPoints(pt.extractDefaultArgs, param, false, args...)
	if err != nil {
		err := fmt.Errorf("alawysFuncPoints failed; %w", err)
		return false, err
	}

	if tnql.IsDryrun(param) {
		return true, nil
	}

	return fn(values), nil
}

// evalValueAggregationFunc 计算数值的聚合函数，示例 sumValue(A,B,C), 最终得出一个值
// 多测点 多时间维度
func (pt *PointTypeMap) evalValueAggregationFunc(fn func([]float64) float64, param tnql.Parameters,
	args ...interface{}) (interface{}, error) {
	values, err := pt.checkAndGetValueListPoint(pt.extractValueArgList, param, args...)
	if err != nil {
		err := fmt.Errorf("alawysFuncPoints failed; %w", err)
		return false, err
	}

	if tnql.IsDryrun(param) {
		return true, nil
	}

	return fn(values), nil
}
