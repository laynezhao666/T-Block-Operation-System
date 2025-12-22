// Package service 采集相关配置查询接口
package service

import (
	"cmdb/logic/query"
	"context"
	"fmt"
	"trpcprotocol/cmdb"

	"google.golang.org/protobuf/types/known/emptypb"
)

// NewConfigQueryService 创建配置查询服务
func NewConfigQueryService() cmdb.ConfigQueryService {
	return &configQueryServiceImpl{
		configQueryLogic: query.GetConfigQueryApi(),
	}
}

// configQueryServiceImpl 配置查询相关服务
type configQueryServiceImpl struct {
	configQueryLogic query.IConfigQueryApi
}

// GetCollectorDevice 获取采集器下面的子设备
func (c configQueryServiceImpl) GetCollectorDevice(ctx context.Context, req *cmdb.ReqGetCollectorDevice) (*cmdb.RspGetCollectorDevice, error) {
	if len(req.DeviceNumbers) == 0 {
		return nil, fmt.Errorf("device_numbers is require")
	}
	return c.configQueryLogic.GetCollectorDevice(ctx, req.DeviceNumbers)
}

// GetCollectorPoint 获取采集器下面的测点信息
func (c configQueryServiceImpl) GetCollectorPoint(ctx context.Context, req *cmdb.ReqGetCollectorPoint) (*cmdb.RspGetCollectorPoint, error) {
	if len(req.DeviceNumbers) == 0 {
		return nil, fmt.Errorf("device_numbers is require")
	}
	return c.configQueryLogic.GetCollectorPoint(ctx, req.DeviceNumbers)
}

// GetCollectorTemplate 获取采集模版信息
func (c configQueryServiceImpl) GetCollectorTemplate(ctx context.Context, req *cmdb.ReqGetCollectorTemplate) (*cmdb.RspGetCollectorTemplate, error) {
	if len(req.TemplateNames) == 0 {
		return nil, fmt.Errorf("template_names is require")
	}
	return c.configQueryLogic.GetCollectorTemplate(ctx, req.TemplateNames)
}

// ExportCollectorConfig 导出采集器配置
func (c configQueryServiceImpl) ExportCollectorConfig(ctx context.Context, req *cmdb.ReqExportCollectorConfig) (*emptypb.Empty, error) {
	if req.CollectorType > 0 {
		if req.CollectorType != 1 && req.CollectorType != 3 {
			return nil, fmt.Errorf("device_type can only be [1: tbox, 3: vendor_box]")
		}
	}
	return c.configQueryLogic.ExportCollectorConfig(ctx, req)
}

// GetDeviceEntity 获取设备实体信息
func (c configQueryServiceImpl) GetDeviceEntity(ctx context.Context, req *cmdb.ReqGetDeviceEntity) (*cmdb.RspGetDeviceEntity, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10000
	}
	return c.configQueryLogic.GetDeviceEntity(ctx, req)
}

// GetDevicePoint 获取设备测点信息
func (c configQueryServiceImpl) GetDevicePoint(ctx context.Context, req *cmdb.ReqGetDevicePoint) (*cmdb.RspGetDevicePoint, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10000
	}
	return c.configQueryLogic.GetDevicePoint(ctx, req)
}

// ListCollectorDevice 获取采集器列表
func (c configQueryServiceImpl) ListCollectorDevice(ctx context.Context, req *cmdb.ReqListCollectorDevice) (*cmdb.RspListCollectorDevice, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10000
	}
	return c.configQueryLogic.ListCollectorDevice(ctx, req)
}

// GetMozuInfo 获取模组信息
func (c configQueryServiceImpl) GetMozuInfo(ctx context.Context, req *cmdb.ReqGetMozuInfo) (*cmdb.RspGetMozuInfo, error) {
	return c.configQueryLogic.GetMozuInfo(ctx, req)
}

// GetConfigModifyTime 获取采集器配置修改时间
func (c configQueryServiceImpl) GetConfigModifyTime(ctx context.Context, req *cmdb.ReqGetCollectorDevice) (*cmdb.RspGetConfigModifyTime, error) {
	return c.configQueryLogic.GetCollectorDataVer(ctx, req.DeviceNumbers)
}
