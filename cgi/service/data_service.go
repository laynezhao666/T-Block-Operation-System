package service

import (
	"cgi/entity/errcode"
	"cgi/logic/data"
	"context"
	pb "trpcprotocol/cgi"

	"trpc.group/trpc-go/trpc-go/errs"
)

type dataService struct {
	dataApi data.IDataApi
}

// NewDataService  创建一个Cmdb接口服务对象
func NewDataService() pb.DataService {
	return &dataService{
		dataApi: data.GetDataApi(),
	}
}

func (d dataService) Query(ctx context.Context, req *pb.ReqDataQuery) (*pb.RspDataQuery, error) {
	return d.dataApi.Query(ctx, req)
}

func (d dataService) TracePoint(ctx context.Context, req *pb.ReqTracePoint) (*pb.RspTracePoint, error) {
	if req.MozuId == 0 || req.PointKey == "" {
		return nil, errs.Newf(errcode.RequestParamError, "mozu_id and point_key required")
	}
	return d.dataApi.TracePoint(ctx, req.MozuId, req.PointKey)
}

func (d dataService) QueryLatest(ctx context.Context, req *pb.ReqQueryLatest) (*pb.RspQueryLatest, error) {
	return d.dataApi.QueryLatest(ctx, req)
}
