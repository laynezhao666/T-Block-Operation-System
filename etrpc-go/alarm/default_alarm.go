package alarm

import (
	"context"
	"trpc.group/trpc-go/trpc-go/log"
)

// defaultAlarmClient 默认告警
type defaultAlarmClient struct {
}

// NewDefaultAlarm 创建默认告警
func NewDefaultAlarm() Alarmer {
	return &defaultAlarmClient{}
}

func (d defaultAlarmClient) Alarm(ctx context.Context, msgs string) error {
	log.ErrorContext(ctx, msgs)
	return nil
}
