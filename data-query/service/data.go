// Package service  store api service definition, basic implementation of proto
package service

import (
	"context"
	logic "data-query/logic/query"
	"trpcprotocol/data-query"
)

// DataServiceImpl 数据服务接口
type DataServiceImpl struct {
}

// DataChange 数据变化查询接口
func (d DataServiceImpl) DataChange(ctx context.Context, req *data_query.ChangeRequest) (*data_query.ChangeResponse, error) {
	return logic.DataChangeHandler(ctx, req)
}

// DataQuery 数据查询接口
func (d DataServiceImpl) DataQuery(ctx context.Context, req *data_query.QueryRequest) (*data_query.QueryResponse, error) {
	return logic.DataQueryHandler(ctx, req)
}

// DataPointChange 变化测点查询接口
func (d DataServiceImpl) DataPointChange(ctx context.Context, req *data_query.PointChangeRequest) (*data_query.PointChangeResponse, error) {
	return logic.DataPointChangeHandler(ctx, req)
}
