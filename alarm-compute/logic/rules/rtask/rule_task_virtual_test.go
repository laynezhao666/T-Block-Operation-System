package rtask

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"trpc.group/trpc-go/trpc-database/kafka"
	"trpc.group/trpc-go/trpc-go"

	"alarm-compute/logic/collector/vtpoint"
)

func TestSendVirtualData(t *testing.T) {
	Key := &vtpoint.KafkaMessageKey{
		T:     time.Now().Unix(),
		D:     1,
		PubMs: time.Now().UnixMilli(),
	}
	newPointMsg := &vtpoint.VirtualPointMsg{
		Interval: 1,
		BoxID:    "alarmVirtualPoints",
		Points: []vtpoint.VirtualPoint{
			{
				I: "1234567.testVirtualPoint",
				V: "10000000019.02",
				Q: "0",
				T: strconv.FormatInt(time.Now().Unix(), 10),
			},
		},
	}
	key, err := json.Marshal(Key)
	if err != nil {
		fmt.Println(err)
		return
	}
	data, err := json.Marshal(newPointMsg)
	if err != nil {
		fmt.Println(err)
		return
	}
	cli := kafka.NewClientProxy("***")
	partition, offset, err := cli.SendSaramaMessage(trpc.BackgroundContext(), sarama.ProducerMessage{
		Topic: "virtual_test",
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(data),
	})
	if err != nil {
		fmt.Println(data, " ", err)
	} else {
		fmt.Println("partition: ", partition, " offset: ", offset)
	}
}

type MapTest struct {
	connMap []int32 // conn mozuId
}

type infoKey struct {
}

type testS struct{}
type connectInfo struct {
	mozuId string
}

func TestPerformance(t *testing.T) {
	// consts (
	// 	ip   = ""
	// 	port = 8082
	// )
	// dial := ws.Dialer{
	// 	Header: ws.HandshakeHeaderHTTP{
	// 		"mozu_id": []string{"421"},
	// 	},
	// }

	// conn, _, _, err := dial.Dial(trpc.BackgroundContext(), fmt.Sprintf("ws://%s:%d", ip, port))
	// if err != nil {
	// 	log.Fatalf("err: %v", err)
	// } else {
	// 	fmt.Println(conn)
	// }
}
