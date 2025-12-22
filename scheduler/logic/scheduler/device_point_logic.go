package scheduler

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"scheduler/entity/consts"
	"scheduler/entity/dbmodel"
	"scheduler/entity/model"
	"scheduler/repo/db"
	"strings"
	"trpc.group/trpc-go/trpc-go/client"
	pb "trpcprotocol/data-compute"
)

// NewDevicePointLogic 创建设备测点计算调度任务
func NewDevicePointLogic(unitCfg *model.TaskConfig) ISchedulerLogic[*dbmodel.DevicePoint, *pb.ReqReceiveTask] {
	obj := &devicePointLogic{}
	obj.DefaultSchedulerLogic = DefaultSchedulerLogic[*dbmodel.DevicePoint, *pb.ReqReceiveTask]{
		ISchedulerLogic: obj,
		Dao:             db.NewPointSchedulerDao(unitCfg),
		UnitCfg:         unitCfg,
	}
	return obj
}

// devicePointLogic 设备测点调度任务具体实现
type devicePointLogic struct {
	DefaultSchedulerLogic[*dbmodel.DevicePoint, *pb.ReqReceiveTask]
}

// PartitionData 使用默认的任务拆分方式即可
func (p *devicePointLogic) PartitionData(data []*model.TaskItem[*dbmodel.DevicePoint], workerMap map[string]*model.WorkerInfo, lastAssignResult map[string]string) error {
	return p.DefaultPartitionData(data, workerMap, true, lastAssignResult)
}

// ConvertToReq 转换为下发请求参数
func (p *devicePointLogic) ConvertToReq(addData []*dbmodel.DevicePoint, delData []string, fullPublish bool, verMark string) *pb.ReqReceiveTask {
	var publishType int32
	if fullPublish {
		publishType = 1
	}
	delTask := make([]*pb.PointTask, 0, len(delData))
	for _, taskKey := range delData {
		fields := strings.Split(taskKey, consts.CommonFieldSeq)
		if len(fields) == 3 {
			delTask = append(delTask, &pb.PointTask{
				DeviceGid:   fields[0],
				PointNameEn: fields[1],
				Version:     fields[2],
			})
		}
	}
	return &pb.ReqReceiveTask{
		TaskVerMark: verMark,
		PublishType: publishType,
		DelTask:     delTask,
		AddTask: lo.Map(addData, func(item *dbmodel.DevicePoint, index int) *pb.PointTask {
			return &pb.PointTask{
				DeviceGid:     item.DeviceGid,
				PointNameEn:   item.PointNameEn,
				Expression:    item.Expression,
				ExpressionMap: item.ExpressionMap,
				Version:       fmt.Sprint(item.UpdateAt.UnixMilli()),
				MozuId:        item.MozuId,
			}
		})}
}

func (p *devicePointLogic) CallPublish(ctx context.Context, worker *model.WorkerInfo, req *pb.ReqReceiveTask) (err error) {
	// 生成proxy客户端并发送请求
	cli := pb.NewComputeClientProxy(client.WithProtocol(worker.WorkerProtocol))
	if _, err = cli.ReceiveTask(ctx, req, client.WithTarget(fmt.Sprintf("ip://%s", worker.GetAddr()))); err != nil {
		return err
	}
	return nil
}
