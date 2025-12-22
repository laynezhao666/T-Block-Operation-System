package service

import (
	"context"
	"agent/logic/task"
	"agent/repo/cm"

	pb "trpcprotocol/agent"

	"trpc.group/trpc-go/trpc-go/log"
)

// ScheduleServiceImpl 任务接收服务
type ScheduleServiceImpl struct{}

var (
	DeviceNumber []string
)

// RecvTask 接收任务
func (t *ScheduleServiceImpl) RecvTask(ctx context.Context, req *pb.TaskReq) (*pb.TaskRsp, error) {
	log.Infof("TaskReceive %v", req)
	log.Infof("TaskReceive collector device count:%v", len(req.Data))

	collectorList := []string{}
	for _, c := range req.Data {
		collectorList = append(collectorList, c.DeviceNumber)
	}

	task.GetInstance().UpdateTasks(collectorList)

	// 通知任务变化
	cm.ConfigChangedChan() <- true

	rsp := pb.TaskRsp{
		Msg: "ok",
	}
	return &rsp, nil
}
