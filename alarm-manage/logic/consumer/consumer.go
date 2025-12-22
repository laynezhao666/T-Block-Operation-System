// Package consumer consumer
package consumer

import (
	"context"

	pb "trpcprotocol/alarm-manage"

	"etrpc-go/log"

	"github.com/IBM/sarama"
	"google.golang.org/protobuf/proto"

	"alarm-manage/logic/manager"
	"alarm-manage/utils/modcall"
)

// Consumer 消费者结构
type Consumer struct{}

// Handle 消费回调方法
func (Consumer) Handle(ctx context.Context, msg *sarama.ConsumerMessage) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return HandleMessage(ctx, msg)
	}
}

// HandleMessage 处理告警/恢复消息
func HandleMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	alertMsg := &pb.AlarmMsgPb{}
	err := proto.Unmarshal(msg.Value, alertMsg)
	if err != nil {
		log.Error("HandleMessage Unmarshal error: ", err)
		return err
	}
	defer func() {
		modcall.RecordConsumeAlertCnt(int(alertMsg.MozuId), 1)
	}()
	if alertMsg.EndAt > 0 {
		// 恢复告警
		manager.GetGlobalManager().AddRestoreToCh(alertMsg)
	} else {
		manager.GetGlobalManager().AddAlertToCh(alertMsg)
	}

	return nil
}
