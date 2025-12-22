package cgi

import (
	"context"

	"agent/logic/distribution/interval"

	pb "trpcprotocol/agent"
)

// IntervalPointsHandle 采集间隔点数据
func IntervalPointsHandle(ctx context.Context) (*pb.RspIntervalPoints, error) {
	tmp := interval.CollectProcessorManager().GetPoints()
	data := make(map[int32]*pb.DataPointIDs)
	for k, v := range tmp {
		ids := make([]string, len(v))
		for i, id := range v {
			ids[i] = string(id)
		}
		data[int32(k)] = &pb.DataPointIDs{
			DataPointId: ids,
		}
	}
	return &pb.RspIntervalPoints{
		Data: data,
	}, nil
}
