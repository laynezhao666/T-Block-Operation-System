// Package query provides basic query functions.
package query

import (
	"context"
	"data-query/repo/read"
	"time"
	"trpcprotocol/data-query"
)

// DataChangeHandler  查询目标测点最近一次变化的时间
func DataChangeHandler(ctx context.Context, req *data_query.ChangeRequest) (*data_query.ChangeResponse, error) {
	// 通过读取插件来查询实际数据
	res, err := read.BatchReadChangedPoint(ctx, req.PointList, req.Begin, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	return &data_query.ChangeResponse{ChangedPointMap: res}, nil
}
