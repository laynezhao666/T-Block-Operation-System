// Package http 实现HTTP协议门禁控制器的驱动层。
package http

import (
	"dac/entity/consts"
	"dac/entity/model/rt"
	"dac/entity/utils"
)

// GetDoorPoints 获取指定门的测点数据，
// 包含门状态和当前告警信息。
func (c *Controller) GetDoorPoints(doors []int) (map[string]map[int]*rt.Point, error) {
	// 获取门状态
	states, err := c.GetDoorState(doors)
	if err != nil {
		return nil, err
	}

	// 获取当前告警
	alarmData, err := c.GetCurrentAlarm()
	if err != nil {
		return nil, err
	}

	return utils.BuildDoorPoints(c.baseInfo.ID, doors, states, alarmData, func(alarmType int) bool {
		return alarmType == consts.AlarmTypeOpenAlarm
	}), nil
}
