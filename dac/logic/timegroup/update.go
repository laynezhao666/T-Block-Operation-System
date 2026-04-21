// Package timegroup 提供时间组的更新和同步功能。
package timegroup

import (
	"dac/entity/consts"
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/logic/cache"
	"dac/repo/dac"

	"gorm.io/gorm"
)

// UpdateInController 将时间组更新同步到所有在线控制器
func UpdateInController(tx *gorm.DB, timeGroup driver.TimeGroup) error {
	groups := []driver.TimeGroup{timeGroup}
	b, err := driver.Marshal(groups)
	if err != nil {
		return err
	}

	controllers := cache.Get().GetAllControllers()

	reqs := make([]db.Request, 0, len(controllers))
	for controllerID, c := range controllers {
		reqs = append(reqs, db.Request{
			ControllerID: controllerID,
			Method:       driver.MethodSetTimeGroup,
			Payload:      b,
			MozuID:       c.MozuID,
			State:        consts.StateToBeExecuted,
		})
	}

	return dac.AddRequests(tx, reqs)
}
