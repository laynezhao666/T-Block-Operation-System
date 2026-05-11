package service

import (
	"agent/entity/config"
	cmWorker "agent/logic/cm"
	"agent/logic/network/utils"
	"agent/utils/osal"
	"context"
	"encoding/json"

	emptypb "google.golang.org/protobuf/types/known/emptypb"

	"agent/logic/cgi"

	pb "trpcprotocol/agent"
)

// CgiServiceImpl cgi服务
type CgiServiceImpl struct{}

// Profile 采集profile
func (c *CgiServiceImpl) Profile(ctx context.Context, req *emptypb.Empty) (*pb.ProfileRsp, error) {
	return cgi.ProfileHandle()
}

// Debug 调试
func (c *CgiServiceImpl) Debug(ctx context.Context, req *pb.ReqDebug) (*pb.RspDebug, error) {
	return cgi.DebugHandle(ctx, req)
}

// Devices 设备列表
func (c *CgiServiceImpl) Devices(ctx context.Context, req *emptypb.Empty) (*pb.RspDevices, error) {
	return cgi.DevicesHandle(ctx)
}

// IntervalPoints 采集间隔点
func (c *CgiServiceImpl) IntervalPoints(ctx context.Context, req *emptypb.Empty) (*pb.RspIntervalPoints, error) {
	return cgi.IntervalPointsHandle(ctx)
}

// Rtd rtd
func (c *CgiServiceImpl) Rtd(ctx context.Context, req *pb.ReqRtd) (*pb.RspRtd, error) {
	return cgi.GetRtdHandle(ctx, req)
}

// RtdById rtd by id
func (c *CgiServiceImpl) RtdById(ctx context.Context, req *pb.ReqRtd) (*pb.RspRtd, error) {
	return cgi.GetRtdByIdHandle(ctx, req)
}

// StartupProbe 启动探测
func (c *CgiServiceImpl) StartupProbe(ctx context.Context, req *emptypb.Empty) (*pb.RspStartupProbe, error) {
	return cgi.StartupProbeHandle(ctx)
}

// Qua qua
func (c *CgiServiceImpl) Qua(ctx context.Context, req *pb.ReqQua) (*pb.RspQua, error) {
	return cgi.QuaHandle(ctx, req)
}

// ExprValidate 表达式校验
func (c *CgiServiceImpl) ExprValidate(ctx context.Context, req *pb.ReqExprValidate) (*pb.RspExprValidate, error) {
	return cgi.ExprValidateHandle(ctx, req)
}

// DevicesCommste 设备通讯状态
func (c *CgiServiceImpl) DevicesCommste(ctx context.Context, req *emptypb.Empty) (*pb.DevicesCommsteRsp, error) {
	return cgi.DevicesCommsteHandle(ctx, req)
}

// SetRtdById 设置rtd by id
func (c *CgiServiceImpl) SetRtdById(ctx context.Context, req *pb.SetRtdByIdReq) (*emptypb.Empty, error) {
	return cgi.SetRtdByIdHandle(ctx, req)
}

// SetEnv 设置环境变量
func (c *CgiServiceImpl) SetEnv(ctx context.Context, req *pb.SetEnvReq) (*emptypb.Empty, error) {
	var chList []string
	for _, ori := range req.Ids {
		if comConfig, ok := config.GetRB().Collector.Modbus.SerialsMap.COMs[ori]; ok {
			chList = append(chList, comConfig.Dev)
		} else {
			// 网口直接加入
			chList = append(chList, ori)
		}
	}

	osal.Instance().BatchSet(req.Name, chList, 1)
	return &emptypb.Empty{}, nil
}

func (c *CgiServiceImpl) SetSnBinding(ctx context.Context, req *pb.SetSnBindingReq) (*emptypb.Empty, error) {
	return cgi.SetSnBindingHandle(ctx, req)
}

// GetBasicInfo 获取基本信息
func (c *CgiServiceImpl) GetBasicInfo(ctx context.Context, req *emptypb.Empty) (*pb.GetBasicInfoRsp, error) {
	return cgi.GetBasicInfoHandle(ctx, req)
}

// SetBasicInfo 设置基本信息
func (c *CgiServiceImpl) SetBasicInfo(ctx context.Context, req *pb.SetBasicInfoReq) (*emptypb.Empty, error) {
	return cgi.SetBasicInfoHandle(ctx, req)
}

// Version 版本信息
func (c *CgiServiceImpl) Version(ctx context.Context, req *emptypb.Empty) (*pb.VersionRsp, error) {
	version, ext := utils.GetVersion()

	// 添加配置源信息
	ext["source"] = config.GetRB().Project.Source

	// 从内存缓存获取配置版本信息（无网络调用）
	versionMap := cmWorker.Worker().GetConfigVersion()
	if len(versionMap) > 0 {
		versionJSON, _ := json.Marshal(versionMap)
		ext["configVersion"] = string(versionJSON)
	}

	return &pb.VersionRsp{Version: version, Ext: ext}, nil
}

// ChangeMode 切换运营态
func (c *CgiServiceImpl) ChangeMode(ctx context.Context, req *pb.ChangeModeReq) (*emptypb.Empty, error) {
	return cgi.ChangeModeHandle(ctx, req)
}
