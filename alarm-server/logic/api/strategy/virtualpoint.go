package strategy

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"
	"trpc.group/trpc-go/trpc-go"

	"alarm-server/repo/dao/strategy"
	cmodel "common/entity/model"

	pb "trpcprotocol/alarm-server"
)

// GetVirtualPoint 获取虚拟测点
func (s *strategyLogicImpl) GetVirtualPoint(ctx context.Context, req *pb.ReqGetVirtualPoint) (*pb.RspGetVirtualPoint, error) {
	list, _, err := strategy.NewStrategyDao().GetStrategyList(trpc.BackgroundContext(), &strategy.StrategyFilter{
		MozuId:        req.MozuId,
		MozuAllowZero: true,
		RidType:       []int64{2},
		Gid:           req.DeviceGid,
		AlarmName:     req.PointId,
	})
	if err != nil {
		return nil, err
	}
	return &pb.RspGetVirtualPoint{
		List: lo.Map(list, func(item cmodel.AlarmStrategy, _ int) *pb.RspGetVirtualPoint_Item {
			return &pb.RspGetVirtualPoint_Item{
				MozuId:        int64(item.MozuId),
				PointName:     fmt.Sprintf("%s.%s", item.DeviceGid, item.AlarmName),
				Expression:    item.AlarmExpression,
				ExpressionMap: item.ExpressionMap,
				CreateAt:      item.CreateAt.Format(time.DateTime),
				UpdateAt:      item.UpdateAt.Format(time.DateTime),
			}
		}),
	}, nil
}
