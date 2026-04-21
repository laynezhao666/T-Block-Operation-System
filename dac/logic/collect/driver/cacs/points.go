// Package cacs 实现CACS门禁控制器协议的驱动层。
package cacs

import (
	"dac/entity/model/rt"
	"dac/entity/utils"
	consts2 "dac/logic/collect/driver/cacs/consts"
)

// GetDoorPoints 获取门测点数据（包含门状态和当前告警信息），
// 仅将门开超时类型的告警作为有效告警。
func (c *Controller) GetDoorPoints(doors []int) (map[string]map[int]*rt.Point, error) {
	if _, err := c.checkConnection(); err != nil {
		return nil, err
	}
	states, err := c.GetDoorState(doors)
	if err != nil {
		return nil, err
	}

	alarmData, err := c.GetCurrentAlarm()
	if err != nil {
		return nil, err
	}

	return utils.BuildDoorPoints(c.baseInfo.ID, doors, states, alarmData, func(alarmType int) bool {
		return alarmType == int(consts2.KAlarmDoorOpenTimeout)
	}), nil
}
