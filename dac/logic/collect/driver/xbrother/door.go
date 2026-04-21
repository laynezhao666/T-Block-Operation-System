// Package xbrother 实现XBrother门禁控制器协议的驱动层。
package xbrother

import (
	"fmt"
)

// doorInfo 门信息结构，包含门编号和名称
type doorInfo struct {
	DoorNo   int    `json:"door_no"`   // 门编号
	DoorName string `json:"door_name"` // 门名称
}

// GetDoors 获取控制器的所有门信息列表
func (c *Controller) GetDoors() (interface{}, error) {
	doorInfos := make([]doorInfo, 0)
	for i := 0; i < c.doorNum; i++ {
		doorInfos = append(doorInfos, doorInfo{DoorNo: i + 1, DoorName: fmt.Sprintf("door%d", i+1)})
	}
	return doorInfos, nil
}
