// Package door 提供门参数的保存和同步功能。
package door

import (
	"context"

	"dac/entity/config"
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/utils"
	"dac/logic/mapping"
	"dac/repo/dac"

	"gorm.io/gorm"
)

// SaveDoorParameters 保存门参数到数据库，同时更新GID映射
func SaveDoorParameters(ctx context.Context,
	controllerID db.IDType,
	parameters []driver.DoorParameter,
	mozuID string,
) error {
	doors := make([]db.Door, 0, len(parameters))
	for i := range parameters {
		p := &parameters[i]

		d := db.Door{
			Number:       int(p.Number),
			GroupID:      db.DefaultGroupID,
			ControllerID: controllerID,
			Parameters:   utils.ConvertDBDoorParameter(p),
		}

		doors = append(doors, d)
	}

	var codeGIDs map[string]db.GIDType
	if !config.C.IgnoreGID(mozuID) {
		// 先获取 controller 记录（使用独立查询，不在事务中）
		c, err := dac.GetRW().GetDoorControllerRecord(ctx, controllerID)
		if err != nil {
			return err
		}

		code := c.GetCollectCode()
		if len(code) > 0 {
			codes := make([]string, 0, len(doors))
			for i := range doors {
				codes = append(codes, doors[i].GetCollectCode(code))
			}

			codeGIDs, err = mapping.GetWorker().FetchGIDs(codes)
			if err != nil {
				return err
			}
		}
	}

	return dac.GetRW().SetDoors(ctx, controllerID, doors, func(tx *gorm.DB) error {
		if len(codeGIDs) == 0 {
			return nil
		}

		return dac.UpdateControllerAndDoorGIDsByCode(tx, codeGIDs)
	})
}
