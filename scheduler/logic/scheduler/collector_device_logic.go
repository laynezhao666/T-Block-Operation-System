package scheduler

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"scheduler/entity/dbmodel"
	"scheduler/entity/model"
	"scheduler/repo/db"
	"trpc.group/trpc-go/trpc-go/client"
	pb "trpcprotocol/agent"
)

// NewCollectorDeviceLogic 创建告警策略下发调度任务
func NewCollectorDeviceLogic(unitCfg *model.TaskConfig) ISchedulerLogic[*dbmodel.CollectorDevice, *pb.RecvCollectTaskReq] {
	obj := &collectorDeviceLogic{}
	obj.DefaultSchedulerLogic = DefaultSchedulerLogic[*dbmodel.CollectorDevice, *pb.RecvCollectTaskReq]{
		ISchedulerLogic: obj,
		Dao:             db.NewCollectorSchedulerDao(unitCfg),
		UnitCfg:         unitCfg,
	}
	return obj
}

// collectorDeviceLogic 采集设备调度任务具体实现
type collectorDeviceLogic struct {
	DefaultSchedulerLogic[*dbmodel.CollectorDevice, *pb.RecvCollectTaskReq]
}

// PartitionData 使用默认的任务拆分方式即可
func (c *collectorDeviceLogic) PartitionData(data []*model.TaskItem[*dbmodel.CollectorDevice], workerMap map[string]*model.WorkerInfo, lastAssignResult map[string]string) error {
	return c.DefaultPartitionData(data, workerMap, true, lastAssignResult)
}

// ConvertToReq 转换为下发请求参数
func (c *collectorDeviceLogic) ConvertToReq(addData []*dbmodel.CollectorDevice, delData []string, fullPublish bool, verMark string) *pb.RecvCollectTaskReq {
	var publishType int32
	if fullPublish {
		publishType = 1
	}
	delTask := make([]*pb.RecvCollectTaskReq_CollectDeviceInfo, 0, len(delData))
	for _, taskKey := range delData {
		delTask = append(delTask, &pb.RecvCollectTaskReq_CollectDeviceInfo{
			DeviceGid: taskKey,
		})
	}
	// 转化为采集端任务类型
	return &pb.RecvCollectTaskReq{
		VersionMark: verMark,
		PublishType: publishType,
		DelDevices:  delTask,
		AddDevices: lo.Map(addData, func(item *dbmodel.CollectorDevice, index int) *pb.RecvCollectTaskReq_CollectDeviceInfo {
			return &pb.RecvCollectTaskReq_CollectDeviceInfo{
				Id:                 item.Id,
				DeviceGid:          item.DeviceGid,
				DeviceNumber:       item.DeviceNumber,
				DeviceType:         item.CollectorType,
				ChannelType:        item.ChannelType,
				ChannelId:          item.ChannelId,
				ChannelLink:        item.ChannelLink,
				Profile:            item.Profile,
				ActiveStatus:       item.ActiveStatus,
				TemplateName:       item.TemplateName,
				ParentDeviceNumber: item.ParentDeviceNumber,
				MozuId:             item.MozuId,
			}
		})}
}

func (c *collectorDeviceLogic) CallPublish(ctx context.Context, worker *model.WorkerInfo, req *pb.RecvCollectTaskReq) (err error) {
	// 生成proxy客户端并发送请求
	cli := pb.NewCollectTaskClientProxy(client.WithProtocol(worker.WorkerProtocol))
	if _, err = cli.RecvCollectTask(ctx, req, client.WithTarget(fmt.Sprintf("ip://%s", worker.GetAddr()))); err != nil {
		return err
	}
	return nil
}
