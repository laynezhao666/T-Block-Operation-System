// Package cacs 实现CACS门禁控制器协议的驱动层。
package cacs

import (
	"fmt"

	"dac/entity/model/driver/cacs"
	"dac/logic/collect/driver/cacs/consts"
)

// 远程控制模式常量
var (
	RemoteControlOpenDoorOnce    uint8 = 0 // 开门一次
	RemoteControlOpenDoorAlways  uint8 = 1 // 常开
	RemoteControlCloseDoorAlways uint8 = 2 // 常闭
	RemoteControlOpenDoorAuto    uint8 = 3 // 恢复自动控制
)

// Reset 重置所有门为自动控制模式
func (c *Controller) Reset() error {
	if _, err := c.checkConnection(); err != nil {
		return err
	}
	resp, err := c.GetDoors()
	if err != nil {
		c.Errorf("Get doors failed: %s", err)
		return err
	}
	getDoorInfos, ok := resp.([]getDoorInfo)
	if !ok {
		c.Errorf("GetDoors response type error, expect []getDoorInfo")
		return fmt.Errorf("GetDoors response type error, expect []getDoorInfo")
	}
	for i := range getDoorInfos {
		_, ok, packetRtn, _, err := c.RemoteControl(cacs.DoorControlReq{
			ControlMode: RemoteControlOpenDoorAuto,
			Id:          getDoorInfos[i].DoorNo,
		})
		if !ok {
			return fmt.Errorf("remote control failed, err: %s", err.Error())
		}
		if packetRtn != consts.KRtnNormal {
			return fmt.Errorf(consts.RtnInfoMap[packetRtn])
		}
	}
	return nil
}
