package tnql

import (
	"errors"
)

// Parameters TODO
/*
	Parameters is a collection of named parameters that can be used by an EvaluableExpression to retrieve parameters
	when an expression tries to use them.
*/
type Parameters interface {

	// Get TODO
	/*
		Get gets the parameter of the given name, or an error if the parameter is unavailable.
		Failure to find the given parameter should be indicated by returning an error.
	*/
	Get(name string) (interface{}, error)

	// GetExpression TODO
	// return current EvaluableExpression
	GetExpression() *EvaluableExpression
}

// MapParameters TODO
type MapParameters map[string]interface{}

// Get TODO
func (p MapParameters) Get(name string) (interface{}, error) {

	value, found := p[name]

	if !found {
		errorMessage := "No parameter '" + name + "' found."
		return nil, errors.New(errorMessage)
	}

	return value, nil
}

// GetExpression MapParameters no need expresion, just an empty implementation
func (p MapParameters) GetExpression() *EvaluableExpression {
	// empty implementation
	return nil
}

// IsDryrun 用于获取函数参数，不进行计算
func IsDryrun(param Parameters) bool {
	if param == nil || param.GetExpression() == nil {
		return false
	}

	customData := param.GetExpression().CustomData
	if isDryrun, ok := customData[ExprCustomDataKeyDryRun]; ok && isDryrun.(bool) {
		return true
	}

	return false
}

// IsSimpleExpr 是否简单表达式，即没有函数，只使用 kv 值进行计算的数据（不使用历史数据），等同于原来的实时测点计算
func IsSimpleExpr(param Parameters) bool {
	if param == nil || param.GetExpression() == nil {
		// param 为 nil，有两种情况，主要关注第一种
		// 1. 初始化 NewEvaluableExpressionXXX，如果有 A>10*0.9 这种，会把 `10*0.9` 进行计算，提前校验，走默认的即可
		// 2. param 确实没填，这种内部会直接报错
		// 统一返回 true，当作简单表达式处理
		return true
	}

	customData := param.GetExpression().CustomData
	if isSimple, ok := customData[ExprCustomDataKeySimple]; ok && isSimple.(bool) {
		return true
	}

	return false
}
