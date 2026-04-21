// Package door 提供门禁门信息管理功能。
package door

import (
	"context"
	"encoding/json"
	"fmt"

	"dac/entity/consts"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/repo/dac"

	"gorm.io/gorm"
)

// getDefaultDoorName 获取门名称，为空时使用默认格式 "doorN"。
func getDefaultDoorName(name string, number int) string {
	if len(name) > 0 {
		return name
	}
	return fmt.Sprintf("door%v", number)
}

// addDoors 为门控器初始化门信息
// XBrother 和 CHD 协议都不提供获取门信息的接口，需要在导入映射时手动添加
func addDoors(tx *gorm.DB, c *db.DoorController, doorItems []rt.DoorWithCodeItem) error {
	if (c.Protocol.Name != consts.ProtocolXBrother && c.Protocol.Name != consts.ProtocolChd806d4) || len(doorItems) == 0 {
		return nil
	}

	driverDoors := make([]db.DriverDoorParameter, 0, len(doorItems))
	doors := make([]db.Door, 0, len(doorItems))
	for i := range doorItems {
		d := &doorItems[i]
		number := d.Number
		name := getDefaultDoorName(d.Name, number)

		driverDoors = append(driverDoors, db.NewDriverDoorParameter(c.ID, c.Channel.ID, number, name))
		doors = append(doors, db.NewDoor(c.ID, number, name, d.Code))
	}

	var err error
	if err = dac.SetDriverDoorParameters(tx, c.ID, c.Channel.ID, driverDoors); err != nil {
		return fmt.Errorf("add driver door parameters error: %w", err)
	}
	if err = dac.SetDoors(tx, c.ID, doors, nil); err != nil {
		return fmt.Errorf("set doors error: %w", err)
	}
	return nil
}

// getControllerDoors 按控制器IP分组门信息。
func getControllerDoors(doorItems []rt.DoorWithCodeItem) map[string][]rt.DoorWithCodeItem {
	controllerDoors := make(map[string][]rt.DoorWithCodeItem)
	for _, item := range doorItems {
		controllerDoors[item.ControllerIP] = append(controllerDoors[item.ControllerIP], item)
	}
	return controllerDoors
}

// UpdateCode 批量更新门的编码和扩展信息。
// 根据控制器IP和门编号匹配已有门记录，更新编码、名称和扩展字段。
func UpdateCode(ctx context.Context, doorItems []rt.DoorWithCodeItem) error {
	doorMap := make(map[db.IDType]*db.Door)

	controllerDoors := getControllerDoors(doorItems)

	return dac.GetRW().UpdateDoors(ctx, doorMap, func(tx *gorm.DB) error {
		// 不需要筛选模组
		controllers, err := dac.GetAllControllerRecords(tx, "")
		if err != nil {
			return fmt.Errorf("get controllers error: %w", err)
		}

		ipToControllers := make(map[string]db.DoorController)
		for i := range controllers {
			c := &controllers[i]
			channelID := c.Channel.ID
			_, ok := ipToControllers[channelID]
			if ok {
				return fmt.Errorf("ip %v already exist: %+v", channelID, *c)
			}
			ipToControllers[channelID] = *c

			if err = addDoors(tx, c, controllerDoors[channelID]); err != nil {
				return fmt.Errorf("add doors of controller %v error: %w", c, err)
			}
		}

		existedDoors, err := dac.GetAllDoors(tx)
		if err != nil {
			return fmt.Errorf("get all doors error: %w", err)
		}

		for i := range doorItems {
			d := &doorItems[i]
			channelID := d.ControllerIP
			c, ok := ipToControllers[channelID]
			if !ok {
				return fmt.Errorf("ip %v not found", channelID)
			}

			find := false
			searchDoors := existedDoors[c.ID]
			for j := range searchDoors {
				if searchDoors[j].Number != d.Number {
					continue
				}

				find = true

				dd, ok := doorMap[searchDoors[j].ID]
				if !ok {
					dd = new(db.Door)
					dd.ID = searchDoors[j].ID
					doorMap[dd.ID] = dd
				}

				if searchDoors[j].GetIDCDBCode() != d.Code {
					dd.SetIDCDBCode(d.Code)
				}

				dd.Name = d.Name

				dd.Extend = searchDoors[j].Extend
				if dd.Extend == nil {
					dd.Extend = make(map[string]interface{})
				}

				if len(d.Extend) == 0 {
					continue
				}
				var temp map[string]interface{}
				if err = json.Unmarshal([]byte(d.Extend), &temp); err != nil {
					return fmt.Errorf("extend of %+v invalid, unmarshal error: %w", *d, err)
				}
				for k, v := range temp {
					dd.Extend[k] = v
				}

			}
			if !find {
				return fmt.Errorf("door %+v not found", doorItems[i])
			}
		}

		return nil
	}, nil)
}
