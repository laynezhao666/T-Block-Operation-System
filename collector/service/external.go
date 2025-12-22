package service

import (
	"context"

	"collector/logic/bus/data/external"

	pb "trpcprotocol/collector"

	"google.golang.org/protobuf/types/known/emptypb"
)

// ExternalPlatformServiceImpl 外部平台对接服务
type ExternalPlatformServiceImpl struct{}

// Data 接口实现，接收外部平台测点数据上报
func (c *ExternalPlatformServiceImpl) Data(ctx context.Context, req *pb.ExternalData) (*emptypb.Empty, error) {
	return external.DataHandle(ctx, req)
}
