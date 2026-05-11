package std

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"sync"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"gonum.org/v1/gonum/floats"

	"agent/entity/consts"
)

var (
	commonParameters = map[string]any{
		"max": max,
		"MAX": max,
		"min": min,
		"MIN": min,
		"sum": sum,
		"SUM": sum,
		"avg": avg,
		"AVG": avg,
		"abs": abs,
		"ABS": abs,
	}
	programs   = make(map[string]*vm.Program)
	programsMu sync.RWMutex // 保护 programs 的读写锁
)

// RegisterCommonParameter 注册常用参数
//
//	@param name 参数名
//	@param parameter 参数值，可以是自定义计算函数或PI等常用量
func RegisterCommonParameter(name string, parameter any) {
	commonParameters[name] = parameter
}

func mergeParameters(dst, src map[string]any) {
	for k, v := range src {
		dst[k] = v
	}
}

// ExprEval 表达式计算
func ExprEval(expression string, parameters map[string]any, precision string) (any, consts.Quality, error) {
	// 表达式转换，耗cpu操作前置
	//expression = utils.TransformExpression(expression)
	var p *vm.Program
	var err error

	// 合并公共参数到当前参数
	mergeParameters(parameters, commonParameters)

	// 读锁检查缓存
	programsMu.RLock()
	p, ok := programs[expression]
	programsMu.RUnlock()

	if !ok {
		// 未命中缓存，获取写锁
		programsMu.Lock()
		// 双检避免重复编译
		p, ok = programs[expression]
		if !ok {
			// 编译表达式
			p, err = expr.Compile(expression, expr.Env(parameters))
			if err != nil {
				programsMu.Unlock()
				return nil, consts.QualityStdExprErr, err
			}
			// 存入缓存
			programs[expression] = p
		}
		programsMu.Unlock()
	}

	// 执行表达式
	e, err := expr.Run(p, parameters)
	if err != nil {
		return nil, consts.QualityStdEvalErr, err
	}
	if e == nil {
		return nil, consts.QualityStdValNilErr, nil
	}

	// 结果处理
	var result string
	switch v := e.(type) {
	case bool:
		result = "0"
		if v {
			result = "1"
		}
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, consts.QualityStdNaNInfErr, errors.New("NaN or Inf")
		}
		if len(precision) == 0 { // 不设置精度，则默认保留2位
			result = fmt.Sprintf("%.2f", v)
		} else {
			// 精度范围应该在1-6之间
			prec, err := strconv.Atoi(precision)
			if err != nil || prec < 1 || prec > 6 {
				prec = 2 // 默认精度2位
			}
			result = fmt.Sprintf("%.*f", prec, v)
		}
	case int:
		result = fmt.Sprintf("%d", v)
	default:
		result = fmt.Sprintf("%v", v)
		return result, consts.QualityStdValTypeErr, errors.New("val type err")
	}

	return result, consts.QualityOk, nil
}

func max(nums ...float64) (float64, error) {
	if len(nums) == 0 {
		return math.NaN(), fmt.Errorf("zero length nums")
	}
	return floats.Max(nums), nil
}

func min(nums ...float64) (float64, error) {
	if len(nums) == 0 {
		return math.NaN(), fmt.Errorf("zero length nums")
	}
	return floats.Min(nums), nil
}

func sum(nums ...float64) float64 {
	return floats.Sum(nums)
}

func avg(nums ...float64) (float64, error) {
	l := len(nums)
	if l == 0 {
		return math.NaN(), fmt.Errorf("zero length nums, divided by 0")
	}
	s := floats.Sum(nums)
	return s / float64(l), nil
}

func abs(num float64) float64 {
	return math.Abs(num)
}
