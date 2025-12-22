package consumer

import (
	"encoding/json"
	"fmt"
	"testing"

	pb "trpcprotocol/alarm-manage"

	"github.com/IBM/sarama"
	"google.golang.org/protobuf/proto"
	"trpc.group/trpc-go/trpc-database/kafka"
	"trpc.group/trpc-go/trpc-go"

	"alarm-manage/entity/message"
)

func TestAdd(t *testing.T) {

	ret := message.AlarmTaskRet{
		PointValueMap: map[string]float64{"测试Gid1.测试值": 2},
		PointMap:      map[string][]string{"A": {"测试Gid1.测试值"}},
		StartRunAt:    1726192000,
	}
	retJson, err := json.Marshal(ret)

	fireAlert := &pb.AlarmMsgPb{
		StartAt: 1726192345,
		// EndAt: 1726193900,
		Rid:           29447,
		Gid:           "12344567",
		Level:         "L4",
		AlarmName:     "测试告警名称3",
		Content:       "当前告警内容为1{{测试值}}",
		MozuId:        794,
		AnalyzeResult: string(retJson[:]),
	}
	alertMsg, err := proto.Marshal(fireAlert)
	// topic := "asyncUpdateTask"
	//alertMsg, err := json.Marshal(fireAlert)
	if err != nil {
		t.Error(err)
	}
	cli := kafka.NewClientProxy("ip?topic=asyncUpdateTask&compression=none")
	partition, offset, err := cli.SendSaramaMessage(trpc.BackgroundContext(), sarama.ProducerMessage{
		Topic: "asyncUpdateTask",
		Value: sarama.ByteEncoder(alertMsg),
	})
	if err != nil {
		fmt.Println(alertMsg, " ", err)
	} else {
		fmt.Println("partition: ", partition, " offset: ", offset)
	}
}
