package service

import (
	"context"
	"data-cache/logic/query"
	"errors"
	"time"
	pb "trpcprotocol/data-cache"
)

// pointServiceImpl 测点相关接口实现
type pointServiceImpl struct {
	defaultPointLogic query.IPointApi
}

// NewPointService 创建一个测点接口服务实现对象
func NewPointService() pb.PointService {
	return &pointServiceImpl{
		defaultPointLogic: query.NewDefaultPointLogic(),
	}
}

func (p *pointServiceImpl) QueryCachedPoint(ctx context.Context, req *pb.ReqQueryCachedPoint) (*pb.RspQueryCachedPoint, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}
	return p.defaultPointLogic.QueryCachedPoint(ctx, req)
}

// QueryData 查询测点数据
func (p *pointServiceImpl) QueryData(ctx context.Context, req *pb.ReqQueryData) (*pb.RspQueryData, error) {
	if req.End <= 0 {
		return nil, errors.New("end time can not <= 0")
	}
	if req.Begin > req.End {
		return nil, errors.New("begin time can not > end time")
	}
	return p.defaultPointLogic.QueryData(ctx, req)
}

// QueryChanged 查询变化测点数据
func (p *pointServiceImpl) QueryChanged(ctx context.Context, req *pb.ReqQueryChanged) (*pb.RspQueryChanged, error) {
	if req.Begin == 0 {
		req.Begin = time.Now().Add(-time.Minute * 10).Unix()
	}
	return p.defaultPointLogic.QueryChanged(ctx, req)
}
