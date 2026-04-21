// Package cacs 实现CACS门禁控制器协议的驱动层。
package cacs

import (
	"fmt"

	"dac/entity/model/driver/cacs"
	"dac/logic/collect/driver/cacs/consts"
)

// getDoorInfo 门信息结构（用于GetDoors返回）
type getDoorInfo struct {
	DoorNo   uint32 `json:"door_no"`
	DoorName string `json:"door_name"`
}

// GetDoors 返回门列表
// 优先返回缓存，如果缓存为空则主动扫描门控器
func (c *Controller) GetDoors() (interface{}, error) {
	server, err := c.checkConnection()
	if err != nil {
		return nil, err
	}

	// 如果缓存已设置，直接返回缓存
	if c.doorCacheSet && len(c.doorCache) > 0 {
		return c.doorCache, nil
	}

	// 缓存未设置，扫描门控器获取门列表
	// 这里直接调用底层方法扫描，不依赖 GetDoorParameter 避免循环依赖
	doors, err := c.scanDoors()
	if err != nil {
		return nil, err
	}

	_ = server // 后续如有需要可使用
	return doors, nil
}

// scanDoors 扫描门控器，获取门列表并更新缓存
// 这是内部方法，不对外暴露
func (c *Controller) scanDoors() ([]getDoorInfo, error) {
	doors := make([]getDoorInfo, 0)
	for i := 0; i < consts.KSupportedDoorNum; i++ {
		doorNo := uint32(i + 1)
		// 使用 getDoorParams 来判断门是否存在
		_, ok, packetRtn, _, err := c.getDoorParams(cacs.GetDoorParamsReq{Id: doorNo})
		if !ok {
			// 请求失败，可能是连接问题，返回错误
			return nil, fmt.Errorf("get door params failed, err: %s", err.Error())
		}
		if packetRtn != consts.KRtnNormal {
			// 返回码不正常，说明该门不存在，跳过
			continue
		}
		doors = append(doors, getDoorInfo{
			DoorNo:   doorNo,
			DoorName: fmt.Sprintf("门%d", doorNo),
		})
	}
	if len(doors) == 0 {
		return nil, fmt.Errorf(consts.RtnInfoMap[consts.KRtnInternalServerError])
	}

	// 更新缓存
	c.doorCache = doors
	c.doorCacheSet = true

	return doors, nil
}

// updateDoorCacheFromParams 从门参数中更新门缓存
// 由 GetDoorParameter 调用，避免重复扫描
func (c *Controller) updateDoorCacheFromParams(doorNos []uint32) {
	doors := make([]getDoorInfo, 0, len(doorNos))
	for _, doorNo := range doorNos {
		doors = append(doors, getDoorInfo{
			DoorNo:   doorNo,
			DoorName: fmt.Sprintf("门%d", doorNo),
		})
	}
	c.doorCache = doors
	c.doorCacheSet = true
}

// RefreshDoors 刷新门列表缓存（当门配置发生变化时调用）
func (c *Controller) RefreshDoors() {
	c.doorCacheSet = false
	c.doorCache = nil
}
