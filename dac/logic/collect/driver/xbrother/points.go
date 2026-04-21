// Package xbrother 实现XBrother门禁控制器协议的驱动层。
package xbrother

import (
	"dac/entity/model/rt"
	"dac/entity/utils"
)

// GetDoorPoints 获取门测点数据（包含门状态和当前告警信息）
func (c *Controller) GetDoorPoints(doors []int) (map[string]map[int]*rt.Point, error) {
	states, err := c.GetDoorState(doors)
	if err != nil {
		return nil, err
	}

	alarmData, err := c.GetCurrentAlarm()
	if err != nil {
		return nil, err
	}

	return utils.BuildDoorPoints(c.baseInfo.ID, doors, states, alarmData, nil), nil
}
