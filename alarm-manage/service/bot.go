package service

import (
	"context"

	"etrpc-go/log"

	"google.golang.org/protobuf/types/known/emptypb"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/codec"

	"alarm-manage/logic/notification/robot"
)

// ManageService 告警管理服务接口
type ManageService struct{}

// Bot 机器人点击屏蔽按钮的回调接口
//
//	@receiver o
//	@param ctx
//	@param req
//	@return *emptypb.Empty
//	@return error
func (o *ManageService) Bot(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	msg := trpc.Message(ctx)
	msg.WithSerializationType(codec.SerializationTypeUnsupported)
	hasMsg, botMsg := robot.DecodeMsg(ctx)
	log.Infof("[Bot] hasMsg: %v, botMsg: %+v", hasMsg, botMsg)
	if !hasMsg {
		return &emptypb.Empty{}, nil
	}
	log.Infof("机器人解码后的消息为: %+v", botMsg)
	return &emptypb.Empty{}, nil
}
