package service

import (
	"context"
	"data-compute/logic/compute"
	"data-compute/logic/register"
	"etrpc-go/log"
	"trpcprotocol/data-compute"

	"google.golang.org/protobuf/types/known/emptypb"
)

// computeServiceImpl 测点计算策略服务
type computeServiceImpl struct {
	computeApi compute.IComputeApi
}

// NewComputeService 新建一个测点计算策略服务
func NewComputeService() data_compute.ComputeService {
	return &computeServiceImpl{
		computeApi: compute.GetComputeApi(),
	}
}

// ReceiveTask 接收策略数据
func (s *computeServiceImpl) ReceiveTask(ctx context.Context, req *data_compute.ReqReceiveTask) (*emptypb.Empty, error) {
	log.InfoContextf(ctx, "recive publish task, task_ver_mark:[%s], publish_type:[%d], add_task cnt:[%d], "+
		"del_task cnt:[%d]", req.TaskVerMark, req.PublishType, len(req.AddTask), len(req.DelTask))
	register.TaskVerMark = req.TaskVerMark
	go s.computeApi.ReceiveTask(req)
	return &emptypb.Empty{}, nil
}

func (s *computeServiceImpl) ShowTask(ctx context.Context, req *data_compute.ReqQueryFilter) (*data_compute.RspShowTask, error) {
	return s.computeApi.ShowTask(ctx, req)
}

func (s *computeServiceImpl) ShowData(ctx context.Context, req *data_compute.ReqQueryFilter) (*data_compute.RspShowData, error) {
	return s.computeApi.ShowData(ctx, req)
}
