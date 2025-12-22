package cgi

import (
	"common/util/expr"
	"context"
	"fmt"
	"strconv"
	"strings"

	pb "trpcprotocol/agent"

	"trpc.group/trpc-go/trpc-go/errs"
)

const (
	paramsListSplitChar string = ","
	paramSplitChar      string = "="
)

const (
	CodeOk = iota
	CodeParamInvalid
	CodeExprInvalid
	CodeCalcluateFailed
)

// ExprValidateHandle IDCDB-ExprValidate接口
func ExprValidateHandle(ctx context.Context, req *pb.ReqExprValidate) (*pb.RspExprValidate, error) {
	express := req.GetExpr()
	if express == "" {
		return nil, errs.New(CodeExprInvalid, "expression is empty")
	}
	params, err := getParams(req.Params)
	if err != nil {
		return nil, errs.New(CodeParamInvalid, err.Error())
	}
	result, quality, err := expr.EvalStr(express, params)
	return &pb.RspExprValidate{
		Value: fmt.Sprintf("v=%v, q=%v, err=%v", result, quality, err),
	}, nil
}

// param字符串的样式"A=point1;B=point2;C=point3"
func getParams(paramsStr string) (map[string]any, error) {
	paramsList := strings.Split(paramsStr, paramsListSplitChar)
	params := make(map[string]any)
	for _, param := range paramsList {
		pair := strings.Split(param, paramSplitChar)
		if len(pair) != 2 {
			return nil, fmt.Errorf("invalid param: %v", param)
		}
		switch pair[1] {
		case "true":
			params[pair[0]] = true
		case "false":
			params[pair[0]] = false
		default:
			value, err := strconv.ParseFloat(pair[1], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid param: %v, parse to float64 failed", param)
			}
			params[pair[0]] = value
		}
	}
	return params, nil
}
