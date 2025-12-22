// Package kafka kafka转发
package kafka

import (
	"context"
	"agent/utils"

	"agent/entity/definition"
)

func forwardMessages(writers []*definition.KafkaWriterType, messages []utils.KafkaMessage) {
	if len(writers) == 0 {
		return
	}
	if len(messages) == 0 {
		return
	}
	ctx := context.Background()
	for _, w := range writers {
		go func(writer *definition.KafkaWriterType) {
			_ = utils.WriteData(ctx, writer, messages...)
		}(w)
	}
}
