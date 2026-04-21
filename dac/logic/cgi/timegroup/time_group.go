// Package timegroup 提供时间组的查询、更新和同步功能。
package timegroup

import (
	"context"
	"dac/entity/consts"
	"fmt"

	"dac/entity/model/cgi"
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/utils"
	"dac/logic/timegroup"
	"dac/repo/dac"

	"dac/entity/utils/ttime"
	"gorm.io/gorm"
)

// GetAll 获取所有时间组并转换为CGI格式
func GetAll(ctx context.Context) ([]cgi.TimeGroup, error) {
	records, err := dac.GetRW().GetAllTimeGroups(ctx)
	if err != nil {
		return nil, err
	}

	timeGroups := make([]cgi.TimeGroup, 0)
	for i := range records {
		tg, err := utils.ConvertTimeGroupDBToDriver(records[i])
		if err != nil {
			continue
		}

		timeGroups = append(timeGroups, cgi.TimeGroup{
			TimeGroup: tg,
			GroupName: records[i].GroupName,
		})
	}

	return timeGroups, nil
}

// isTimezoneEqual 比较两个时区列表是否相等（忽略顺序）
func isTimezoneEqual(lhs, rhs []driver.TimeZone) bool {
	l1 := len(lhs)
	l2 := len(rhs)
	if l1 != l2 {
		return false
	}

	lhsMap := make(map[driver.TimeZone]struct{}, l1)
	for _, x := range lhs {
		lhsMap[x] = struct{}{}
	}
	for _, x := range rhs {
		if _, ok := lhsMap[x]; !ok {
			return false
		}
	}
	return true
}

// isWeekDayEqual 比较两个星期列表是否相等（忽略顺序）
func isWeekDayEqual(lhs, rhs []int) bool {
	l1 := len(lhs)
	l2 := len(rhs)
	if l1 != l2 {
		return false
	}

	lhsMap := make(map[int]struct{}, l1)
	for _, x := range lhs {
		lhsMap[x] = struct{}{}
	}
	for _, x := range rhs {
		if _, ok := lhsMap[x]; !ok {
			return false
		}
	}
	return true
}

// Update 更新时间组，变更时同步到门禁控制器
func Update(ctx context.Context, groupNo int, timeGroup cgi.TimeGroup) error {
	dbTimeGroup, err := utils.ConvertTimeGroupDriverToDB(timeGroup.TimeGroup)
	if err != nil {
		return err
	}
	dbTimeGroup.UpdateTime = ttime.GetNowUTC().Unix()
	dbTimeGroup.GroupName = timeGroup.GroupName

	var oldTg driver.TimeGroup
	return dac.GetRW().UpdateTimeGroup(ctx, dbTimeGroup, func(tx *gorm.DB) error {
		oldTimeGroup, err := dac.GetTimeGroup(tx, groupNo)
		if err != nil {
			return err
		}
		if oldTg, err = utils.ConvertTimeGroupDBToDriver(oldTimeGroup); err != nil {
			return err
		}
		return nil
	}, func(tx *gorm.DB) error {
		needUpdateInController := false
		if (len(timeGroup.Week) > 0 && !isWeekDayEqual(timeGroup.Week, oldTg.Week)) ||
			(len(timeGroup.TimeZone) > 0 && !isTimezoneEqual(timeGroup.TimeZone, oldTg.TimeZone)) {
			needUpdateInController = true
		}
		if !needUpdateInController {
			return nil
		}

		return timegroup.UpdateInController(tx, timeGroup.TimeGroup)
	})
}

// Sync 将所有时间组同步到模组下的所有门禁控制器
func Sync(ctx context.Context, mozuID string) error {
	groupsRecords, err := dac.GetRW().GetAllTimeGroups(ctx)
	if err != nil {
		return fmt.Errorf("get time groups error: %w", err)
	}
	groups := make([]driver.TimeGroup, 0, len(groupsRecords))
	for i := range groupsRecords {
		t, err := utils.ConvertTimeGroupDBToDriver(groupsRecords[i])
		if err != nil {
			continue
		}

		groups = append(groups, t)
	}

	b, err := driver.Marshal(groups)
	if err != nil {
		return fmt.Errorf("json marshal %+v error: %w", groups, err)
	}

	controllers, err := dac.GetRW().GetAllDoorControllers(ctx, mozuID)
	if err != nil {
		return err
	}

	reqs := make([]db.Request, 0, len(controllers))
	for i := range controllers {
		reqs = append(reqs, db.Request{
			ControllerID: controllers[i].ID,
			Method:       driver.MethodSetTimeGroup,
			Payload:      b,
			MozuID:       controllers[i].MozuID,
			State:        consts.StateToBeExecuted,
		})
	}

	return dac.GetRW().AddRequests(ctx, reqs)
}
