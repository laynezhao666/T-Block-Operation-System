package expr

import (
	"common/entity/consts"
	"fmt"
	"math"
	"sync"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/expr-lang/expr/vm/runtime"
	"gonum.org/v1/gonum/floats"
)

var (
	commonParameters = map[string]any{
		"max": maxEval,
		"MAX": maxEval,
		"min": minEval,
		"MIN": minEval,
		"sum": sumEval,
		"SUM": sumEval,
		"avg": avgEval,
		"AVG": avgEval,
		"abs": absEval,
		"ABS": absEval,
		"eq":  eqEval,
		"EQ":  eqEval,
		"neq": neqEval,
		"NEQ": neqEval,
	}
	programs   = map[string]*vm.Program{}
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

func EvalStr(expression string, parameters map[string]any) (string, consts.Quality, error) {
	e, quality, err := Eval(expression, parameters)
	switch v := e.(type) {
	case int:
		return fmt.Sprintf("%v", v), quality, err
	case float64:
		return fmt.Sprintf("%.2f", v), quality, err
	default:
		return fmt.Sprintf("%v", v), quality, err
	}
}

func EvalFloat(expression string, parameters map[string]any) (float64, consts.Quality, error) {
	e, quality, err := Eval(expression, parameters)
	switch v := e.(type) {
	case int:
		return float64(v), quality, err
	case float64:
		return v, quality, err
	default:
		return 0, quality, err
	}
}

// Eval 表达式计算
func Eval(expression string, parameters map[string]any) (any, consts.Quality, error) {
	// 表达式转换
	expression = transformExpression(expression)
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
				return 0, consts.QualityStdExprErr, err
			}
			// 存入缓存
			programs[expression] = p
		}
		programsMu.Unlock()
	}

	// 执行表达式
	e, err := expr.Run(p, parameters)
	if err != nil {
		return 0, consts.QualityStdEvalErr, err
	}
	if e == nil {
		return 0, consts.QualityStdValNilErr, nil
	}

	switch v := e.(type) {
	case bool:
		if v {
			return 1, consts.QualityOk, nil
		} else {
			return 0, consts.QualityOk, nil
		}
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, consts.QualityStdNaNInfErr, fmt.Errorf("NaN or Inf")
		}
		return v, consts.QualityOk, nil
	case int:
		return v, consts.QualityOk, nil
	default:
		return v, consts.QualityStdValTypeErr, fmt.Errorf("invalid result type, expression=%s, result=%v", expression, v)
	}
}

// maxEval calc max value with validation
func maxEval(num ...float64) (float64, error) {
	if len(num) == 0 {
		return math.NaN(), fmt.Errorf("zero length num")
	}
	return floats.Max(num), nil
}

// minEval calc min value with validation
func minEval(num ...float64) (float64, error) {
	if len(num) == 0 {
		return math.NaN(), fmt.Errorf("zero length num")
	}
	return floats.Min(num), nil
}

// sumEval calc sum value with validation
func sumEval(num ...float64) float64 {
	return floats.Sum(num)
}

// avgEval calc avg value with validation
func avgEval(num ...float64) (float64, error) {
	l := len(num)
	if l == 0 {
		return math.NaN(), fmt.Errorf("zero length num, divided by 0")
	}
	s := floats.Sum(num)
	return s / float64(l), nil
}

// absEval calc abs value
func absEval(num float64) float64 {
	return math.Abs(num)
}

// parseBool 转换bool类型为0/1
func parseBool(val any) any {
	if v, ok := val.(bool); ok {
		if v {
			return 1
		} else {
			return 0
		}
	}
	return val
}

// eq 判断是否相等
func eqEval(left any, right any) bool {
	return runtime.Equal(parseBool(left), parseBool(right))
}

// neq 判断是否不相等
func neqEval(left any, right any) bool {
	return !eqEval(left, right)
}
