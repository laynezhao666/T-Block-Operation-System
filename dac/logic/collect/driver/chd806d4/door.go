// Package chd806d4 实现CHD806D4门禁控制器协议的驱动层。
package chd806d4

import "fmt"

// GetDoors 获取门列表（根据配置生成，不发送网络请求）
func (c *Controller) GetDoors() (interface{}, error) {
	// 从配置中获取门数量，默认4门
	doorNum := 4
	if num, ok := c.chanInfo.Extend["door_num"].(int); ok && num > 0 {
		doorNum = num
	}

	// 直接根据配置生成门列表（不发送网络请求，与 xbrother 实现方式一致）
	type doorInfo struct {
		DoorNo   int    `json:"door_no"`
		DoorName string `json:"door_name"`
	}

	doorInfos := make([]doorInfo, 0, doorNum)
	for i := 1; i <= doorNum; i++ {
		doorInfos = append(doorInfos, doorInfo{
			DoorNo:   i,
			DoorName: fmt.Sprintf("门%d", i),
		})
	}

	return doorInfos, nil
}
