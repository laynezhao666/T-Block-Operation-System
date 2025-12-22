// Package service 配置获取服务
package service

import (
	"collector/logic/bus/collectors_config"
	"context"

	pb "trpcprotocol/collector"
)

// ConfigBusServiceImpl 配置获取服务
type ConfigBusServiceImpl struct{}

// FetchConfig 服务接口实现
func (c *ConfigBusServiceImpl) FetchConfig(ctx context.Context, req *pb.ReqFetchConfig) (*pb.RspFetchConfig, error) {
	value, err := collectors_config.FetchConfigHandle(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.RspFetchConfig{
		Data: value,
	}, nil
}
