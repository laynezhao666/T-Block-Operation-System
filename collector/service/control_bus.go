package service

import (
	"collector/entity/errcode"
	"collector/logic/bus/control"
	"context"

	monitorPb "trpcprotocol/tboxmonitor"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/errs"
)

// ControlBusServiceImpl 控制总线
type ControlBusServiceImpl struct{}

// Heartbeat 接口实现
func (c *ControlBusServiceImpl) Heartbeat(ctx context.Context, req *monitorPb.RequestHeartbeat) (*monitorPb.ResponseHeartbeat, error) {
	err := trpc.Go(ctx, defaultSendDataTimeout, func(ctx context.Context) {
		control.HeartbeatHandle(ctx, req)
	})
	if err != nil {
		return nil, errs.New(errcode.ErrHeartbeatFail, err.Error())
	}
	return &monitorPb.ResponseHeartbeat{}, nil
}
