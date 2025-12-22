package batch

import (
	"context"
	"time"

	"etrpc-go/log"
)

// GroupData GroupData
type GroupData struct {
	Seq  int
	Data []interface{}
}

// NewGroupData NewGroupData
func NewGroupData() *GroupData {
	item := new(GroupData)
	item.Data = make([]interface{}, 0)
	return item
}

// BatchChannel 从inChan中读取数据，以batchSize为单位，将数据写入outChan
func BatchChannel(ctx context.Context, inChan <-chan interface{}, batchSize int,
	tickerDuration time.Duration) chan *GroupData {
	// 创建无缓冲的通道
	outChan := make(chan *GroupData)
	ticker := time.NewTicker(tickerDuration)
	go func(ch chan *GroupData) {
		defer close(ch)
		comboGroup := NewGroupData()
		seq := 0
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-inChan:
				if !ok {
					return
				}
				comboGroup.Data = append(comboGroup.Data, event)
				seq++
				if seq == batchSize {
					comboGroup.Seq = seq
					ch <- comboGroup
					log.Debugf("batch send len:%s, seq:%d", len(comboGroup.Data), seq)
					comboGroup = NewGroupData()
					seq = 0
				}
			case <-ticker.C:
				if seq > 0 {
					log.Debugf("ticker send len:%d, seq:%d", len(comboGroup.Data), seq)
					comboGroup.Seq = seq
					ch <- comboGroup
					comboGroup = NewGroupData()
					seq = 0
				}
			}
		}
	}(outChan)
	return outChan
}
