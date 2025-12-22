// Package control 控制总线处理
package control

import (
	"collector/repo/report"
	"collector/utils"
	"context"
	"time"

	"etrpc-go/log"

	"trpc.group/trpc-go/trpc-go/metrics"

	monitorPb "trpcprotocol/tboxmonitor"
)

var (
	controlbusProxy = monitorPb.NewMonitorClientProxy()
)

var (
	fixedDimensions = []*metrics.Dimension{
		{
			Name:  report.HandleTypeDimension,
			Value: "control",
		},
	}
)

// HeartbeatHandle 处理心跳
func HeartbeatHandle(ctx context.Context, req *monitorPb.RequestHeartbeat) {
	defer utils.HandlePanic("heartbeat")
	startTime := time.Now().UnixMilli()
	handleCnt := 1
	handleFailCnt := 0
	defer func() {
		endTime := time.Now().UnixMilli()
		t := float64(endTime - startTime)
		additionalDimensions := []*metrics.Dimension{
			{
				Name:  report.UpstreamIpDimension,
				Value: utils.GetUpstreamIp(ctx),
			},
		}
		reportHandleMetric(float64(handleCnt), float64(handleFailCnt), t, additionalDimensions)
	}()
	_, err := controlbusProxy.Heartbeat(ctx, req)
	if err != nil {
		handleFailCnt += 1
		log.ErrorContextf(ctx, "send heartbeat <%v> fail: <%v>", req, err)
		return
	}
	log.InfoContextf(ctx, "send heartbeat <%v> success", req)
}

func reportHandleMetric(handleCnt, handleFailCnt, latency float64, additionalDimensions []*metrics.Dimension) {
	dimensions := append(fixedDimensions, additionalDimensions...)
	report.HandleCnt(dimensions, handleCnt)
	report.HandleFailCnt(dimensions, handleFailCnt)
	report.HandleLatency(dimensions, latency)
}
