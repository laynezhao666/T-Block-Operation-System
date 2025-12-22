// Package alarm provides ...
// @author: xincili
// -------------------------------------------
package alarm

import (
	"context"
)

// Alarmer 告警接口
type Alarmer interface {
	// Alarm 发送告警
	Alarm(ctx context.Context, msgs string) error
}

// GetAlarmClient 获取告警对象
func GetAlarmClient(name string) Alarmer {
	switch name {
	case "default":
		return NewDefaultAlarm()
	}
	return nil
}
