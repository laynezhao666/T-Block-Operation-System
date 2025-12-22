package service

import (
	"cgi/entity/errcode"
	"context"

	"trpc.group/trpc-go/trpc-go/errs"

	"cgi/logic/cmdb"

	pb "trpcprotocol/cgi"
)

// NewCmdbService 创建一个Cmdb接口服务对象
func NewCmdbService() pb.CmdbService {
	return &cmdbService{
		cmdbApi: cmdb.NewCmdbApi(),
	}
}

type cmdbService struct {
	cmdbApi cmdb.ICmdbApi
}

func (c cmdbService) GetSubTree(ctx context.Context, req *pb.ReqGetSubTree) (*pb.RspGetSubTree, error) {
	if req.MozuId == 0 {
		return nil, errs.New(errcode.RequestParamError, "mozu_id required")
	}
	return c.cmdbApi.GetSubTree(ctx, req)
}

func (c cmdbService) GetDeviceEntity(ctx context.Context, req *pb.ReqGetDeviceEntity) (*pb.RspGetDeviceEntity, error) {
	if req.MozuId == 0 {
		return nil, errs.New(errcode.RequestParamError, "mozu_id required")
	}
	return c.cmdbApi.GetDeviceEntity(ctx, req)
}

func (c cmdbService) GetDeviceTree(ctx context.Context, req *pb.ReqGetDeviceTree) (*pb.RspGetDeviceTree, error) {
	if req.MozuId == 0 {
		return nil, errs.New(errcode.RequestParamError, "mozu_id required")
	}
	return c.cmdbApi.GetDeviceTree(ctx, req)
}

func (c cmdbService) GetDevicePoint(ctx context.Context, req *pb.ReqGetDevicePoint) (*pb.RspGetDevicePoint, error) {
	if req.MozuId == 0 {
		return nil, errs.New(errcode.RequestParamError, "mozu_id required")
	}
	return c.cmdbApi.GetDevicePoint(ctx, req)
}

func (c cmdbService) GetSubTreeFieldDic(ctx context.Context, req *pb.ReqGetSubTreeFieldDic) (*pb.RspCommonGetKeyDict, error) {
	if req.MozuId == 0 {
		return nil, errs.New(errcode.RequestParamError, "mozu_id required")
	}
	if req.DeviceGid == "" && req.DeviceNumber == "" {
		return nil, errs.New(errcode.RequestParamError, "device_gid, device_number required any")
	}
	return c.cmdbApi.GetSubTreeFieldDic(ctx, req)
}

func (c cmdbService) GetMozuInfo(ctx context.Context, req *pb.ReqGetMozuInfo) (*pb.RspGetMozuInfo, error) {
	return c.cmdbApi.GetMozuInfo(ctx, req)
}

func (c cmdbService) GetCollectorStatusTree(ctx context.Context, req *pb.ReqGetCollectorStatusTree) (*pb.RspGetCollectorStatusTree, error) {
	if req.MozuId == 0 {
		return nil, errs.New(errcode.RequestParamError, "mozu_id required")
	}
	return c.cmdbApi.GetCollectorStatusTree(ctx, req)
}

func (c cmdbService) GetCollectorInfo(ctx context.Context, req *pb.ReqGetCollectorInfo) (*pb.RspGetCollectorInfo, error) {
	if req.MozuId == 0 || req.DeviceGid == "" {
		return nil, errs.New(errcode.RequestParamError, "mozu_id and  device_gid required")
	}
	return c.cmdbApi.GetCollectorInfo(ctx, req)
}

func (c cmdbService) GetCollectorPoint(ctx context.Context, req *pb.ReqGetCollectorPoint) (*pb.RspGetCollectorPoint, error) {
	if req.MozuId == 0 || req.DeviceGid == "" {
		return nil, errs.New(errcode.RequestParamError, "mozu_id and  device_gid required")
	}
	return c.cmdbApi.GetCollectorPoint(ctx, req)
}
