package rpc

import (
	"context"
	"fmt"
	"sync"

	cPb "trpcprotocol/alarm-compute"
)

var (
	alarmComputeRpc *AlarmComputeRpc
	once            sync.Once
)

// AlarmComputeRpc ...
type AlarmComputeRpc struct {
	computeClientProxy cPb.AlarmComputeClientProxy
}

// NewAlarmComputeRpc 新建远程调用实例
func NewAlarmComputeRpc() *AlarmComputeRpc {
	once.Do(func() {
		alarmComputeRpc = &AlarmComputeRpc{
			computeClientProxy: cPb.NewAlarmComputeClientProxy(),
		}
	})
	return alarmComputeRpc
}

// ExpCompute 表达式计算
// @param express [][]string  每一项为 主表达式 子表达式
// @param pMap 参数列表 []map[string]string  每一项表示主表达式的变量映射关系，子表达式通用
// @param pv 测点值 map[string]map[int64]float64
func (a *AlarmComputeRpc) ExpCompute(ctx context.Context, beginTime, endTime int64, interval int32,
	express [][]string, pMap []map[string]string, pv map[string]map[int64]float64) (*cPb.RspExpCompute, error) {
	if len(express) == 0 || len(express) != len(pMap) {
		return nil, fmt.Errorf("ExpCompute express and pMap length not match")
	}
	req := &cPb.ReqExpCompute{
		BeginTime: beginTime,
		EndTime:   endTime,
		Interval:  interval,
	}
	var withValue bool = len(pv) != 0
	req.WithValue = withValue
	for index, exps := range express {
		for _, e := range exps {
			item := &cPb.ReqExpCompute_Item{
				Express: e,
				PMap:    pMap[index],
			}
			if withValue {
				item.Pv = make(map[string]*cPb.ReqExpCompute_Item_PointValue)
				for k, v := range pv {
					item.Pv[k] = &cPb.ReqExpCompute_Item_PointValue{
						Tv: v,
					}
				}
			}
			req.List = append(req.List, item)
		}
	}
	rsp, err := a.computeClientProxy.ExpCompute(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ExpCompute 远程调用失败: %v", err)
	}
	return rsp, nil
}
