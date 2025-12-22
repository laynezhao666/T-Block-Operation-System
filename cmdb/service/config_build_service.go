package service

import (
	"cmdb/logic/build"
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
	"trpcprotocol/cmdb"
)

// NewConfigBuildService 创建配置导入服务
func NewConfigBuildService() cmdb.ConfigBuildService {
	return &configBuildServiceImpl{
		configBuildLogic: build.GetConfigImportApi(),
	}
}

type configBuildServiceImpl struct {
	configBuildLogic build.IConfigBuildApi
}

// SaveMozu 保存模组
func (s *configBuildServiceImpl) SaveMozu(ctx context.Context, req *cmdb.ReqSaveMozu) (*emptypb.Empty, error) {
	if req.MozuId <= 0 {
		return nil, fmt.Errorf("mozu_id is required")
	}
	if err := s.configBuildLogic.SaveMozu(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ListMozu 查询模组
func (s *configBuildServiceImpl) ListMozu(ctx context.Context, req *cmdb.ReqListMozu) (*cmdb.RspListMozu, error) {
	return s.configBuildLogic.ListMozu(ctx, req)
}

// DeleteMozu 删除模组
func (s *configBuildServiceImpl) DeleteMozu(ctx context.Context, req *cmdb.ReqDeleteMozu) (*emptypb.Empty, error) {
	if len(req.MozuId) != 0 {
		if err := s.configBuildLogic.DeleteMozu(ctx, req); err != nil {
			return nil, err
		}
	}
	return &emptypb.Empty{}, nil
}

// ImportModel 导入模型
func (s *configBuildServiceImpl) ImportModel(ctx context.Context, req *cmdb.ReqImportModel) (*cmdb.RspImportModel, error) {
	// 参数校验, mozuId和版本号必须存在
	if req.MozuId <= 0 {
		return nil, fmt.Errorf("mozu_id is required")
	}
	if req.Version == "" {
		return nil, fmt.Errorf("version is required")
	}
	return s.configBuildLogic.ImportModel(ctx, req)
}
