// Package query provides basic query functions.
package query

import (
	"context"
	"data-query/repo/read"
	"trpcprotocol/data-query"
)

// DataPointChangeHandler 查询指定测点在限定时间内是否发生过变化
func DataPointChangeHandler(ctx context.Context, req *data_query.PointChangeRequest) (*data_query.PointChangeResponse, error) {
	// 通过读取插件来查询实际数据
	res, err := read.BatchReadChangedPoint(ctx, req.PointList, req.Begin, req.End)
	if err != nil {
		return nil, err
	}
	var resPoint []string
	// 判断最近一次变化的时间是否超出给定范围
	for pointName, t := range res {
		if t >= req.Begin && t <= req.End {
			resPoint = append(resPoint, pointName)
		}
	}
	return &data_query.PointChangeResponse{PointList: resPoint}, nil
}
