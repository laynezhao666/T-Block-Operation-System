package scheduler

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"scheduler/entity/dbmodel"
	"scheduler/entity/model"
	"scheduler/repo/db"
	"time"
	"trpc.group/trpc-go/trpc-go/client"
	pb "trpcprotocol/alarm-compute"
)

// NewAlarmStrategyLogic 创建告警策略下发调度任务
func NewAlarmStrategyLogic(unitCfg *model.TaskConfig) ISchedulerLogic[*dbmodel.AlarmStrategy, *pb.ReqStrategyRecv] {
	obj := &alarmStrategyLogic{}
	obj.DefaultSchedulerLogic = DefaultSchedulerLogic[*dbmodel.AlarmStrategy, *pb.ReqStrategyRecv]{
		ISchedulerLogic: obj,
		Dao:             db.NewAlarmSchedulerDao(unitCfg),
		UnitCfg:         unitCfg,
	}
	return obj
}

// alarmStrategyLogic 告警策略下发调度实现
type alarmStrategyLogic struct {
	DefaultSchedulerLogic[*dbmodel.AlarmStrategy, *pb.ReqStrategyRecv]
}

func (a *alarmStrategyLogic) PartitionData(data []*model.TaskItem[*dbmodel.AlarmStrategy], workerMap map[string]*model.WorkerInfo, lastAllocateMap map[string]string) error {
	// 先按ridType分组, 再按计算复杂度均分
	ridTypeDataMap := lo.GroupBy(data, func(item *model.TaskItem[*dbmodel.AlarmStrategy]) int64 {
		return item.TaskData.RidType
	})
	// 针对每一种ridType进行均分,第一批需要将已分配的计算复杂度重置为0
	idx := 0
	for _, strategies := range ridTypeDataMap {
		if err := a.DefaultSchedulerLogic.DefaultPartitionData(strategies, workerMap, idx == 0, lastAllocateMap); err != nil {
			return err
		}
		idx++
	}
	return nil
}

func (a *alarmStrategyLogic) ConvertToReq(addData []*dbmodel.AlarmStrategy, delData []string, fullPublish bool, verMark string) *pb.ReqStrategyRecv {
	publishType := pb.ReqStrategyRecv_INCREMENT
	if fullPublish {
		publishType = pb.ReqStrategyRecv_UPDATEALL
	}
	return &pb.ReqStrategyRecv{
		RecvVersion: verMark,
		PublishType: publishType,
		AddTask: lo.Map(addData, func(item *dbmodel.AlarmStrategy, index int) *pb.ReqStrategyRecv_AddItem {
			return &pb.ReqStrategyRecv_AddItem{
				MozuId:            item.MozuId,
				Rid:               item.Rid,
				Version:           item.RidVersion,
				RidType:           item.RidType,
				AlarmLevel:        item.AlarmLevel,
				AlarmName:         item.AlarmName,
				ContentTemplate:   item.ContentTemplate,
				Gid:               item.DeviceGid,
				AlarmExpression:   item.AlarmExpression,
				RestoreExpression: item.RestoreExpression,
				ExpressionMap:     item.ExpressionMap,
			}
		}),
		DelTaskKey: delData,
	}
}

// CallPublish 下发告警策略
func (a *alarmStrategyLogic) CallPublish(ctx context.Context, worker *model.WorkerInfo, req *pb.ReqStrategyRecv) (err error) {
	// 生成proxy客户端并发送请求
	cli := pb.NewAlarmComputeClientProxy(client.WithProtocol(worker.WorkerProtocol))
	req.RecvTimestamp = time.Now().UnixMilli()
	if _, err = cli.RecvTask(ctx, req, client.WithTarget(fmt.Sprintf("ip://%s", worker.GetAddr()))); err != nil {
		return err
	}
	return nil
}
