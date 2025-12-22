package service

import (
	"cgi/entity/errcode"
	"context"

	"trpc.group/trpc-go/trpc-go/errs"

	"google.golang.org/protobuf/types/known/emptypb"

	"cgi/logic/common"

	pb "trpcprotocol/cgi"
)

type commonService struct {
	commonApi common.ICommonApi
}

// NewCommonService 创建一个公共的服务
func NewCommonService() pb.CommonService {
	return &commonService{
		commonApi: common.NewCommonApi(),
	}
}

func (c commonService) ExportData(ctx context.Context, req *pb.ReqCommonExportData) (*emptypb.Empty, error) {
	if len(req.Fields) == 0 {
		return nil, errs.New(errcode.RequestParamError, "fields list can not be empty")
	}
	return c.commonApi.ExportData(ctx, req)
}

func (c commonService) GetKeyDict(ctx context.Context, req *pb.ReqCommonGetDict) (*pb.RspCommonGetKeyDict, error) {
	return c.commonApi.GetKeyDict(ctx, req)
}

func (c commonService) GetKvDict(ctx context.Context, req *pb.ReqCommonGetDict) (*pb.RspCommonGetKvDict, error) {
	return c.commonApi.GetKvDict(ctx, req)
}
