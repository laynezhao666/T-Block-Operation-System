// Package alarm 提供门禁告警记录的查询和导出功能。
package alarm

import (
	"context"
	"dac/entity/consts"
	"dac/entity/model/cgi"
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/repo/dac"
	"fmt"

	"dac/entity/utils/excel"
	"dac/entity/utils/ttime"
	"github.com/tealeg/xlsx/v3"
	"gorm.io/gorm"
)

// ht 默认Excel行高
const (
	ht = 14.0
)

// titles 告警导出Excel表头
var (
	titles = []string{"时间", "状态", "告警描述", "门名称", "门编号", "门禁控制器名称", "门禁控制器ID"}
)

// getState 将告警状态码转换为中文描述
func getState(s int) string {
	if s == int(driver.AlarmStateAlarming) {
		return "告警产生"
	}
	return "告警恢复"
}

// getCGIAlarms 将数据库告警记录转换为CGI告警结构体
func getCGIAlarms(as []db.Alarm) []cgi.Alarm {
	alarms := make([]cgi.Alarm, 0, len(as))
	for i := range as {
		e := &as[i]

		cgiA := cgi.Alarm{
			ControllerID:   e.ControllerID,
			ControllerName: e.ControllerName,
			Index:          e.Index,
			Time:           ttime.Format(e.Timestamp),
			DoorNumber:     e.DoorNumber,
			DoorName:       e.DoorName,
			Type:           e.Type,
			State:          e.State,
			StateDesc:      getState(e.State),
			Description:    e.Description,
		}

		alarms = append(alarms, cgiA)
	}
	return alarms
}

// Get 分页查询告警记录
func Get(ctx context.Context, mozuID string, controllerIDs []int,
	offset, limit int, beginTime, endTime int64,
) (int64, []cgi.Alarm, error) {
	n, as, err := dac.GetRW().GetAlarms(
		ctx, mozuID, controllerIDs,
		offset, limit, beginTime, endTime, nil,
	)
	if err != nil {
		return 0, nil, err
	}

	return n, getCGIAlarms(as), nil
}

// writeExcel 将告警记录写入Excel文件
func writeExcel(alarms []db.Alarm) (*xlsx.File, error) {
	f := xlsx.NewFile()
	s, err := f.AddSheet("告警记录")
	if err != nil {
		return nil, err
	}

	if _, err = excel.AddStringRow(s, ht, titles...); err != nil {
		return nil, err
	}

	for i := range alarms {
		a := &alarms[i]
		if _, err = excel.AddRow(
			s, ht, ttime.Format(a.Timestamp),
			getState(a.State), a.Description,
			a.DoorName, a.DoorNumber,
			a.ControllerName, a.ControllerID,
		); err != nil {
			return nil, err
		}
	}

	return f, nil
}

// Export 导出告警记录到Excel文件，限制最大导出条数
func Export(ctx context.Context, mozuID string,
	controllerIDs []int, condition cgi.TimeCondition,
) (*xlsx.File, error) {
	var (
		err    error
		alarms []db.Alarm
		e      error
	)
	beginTime := condition.BeginTime
	endTime := condition.EndTime
	if _, err = dac.GetRW().GetAlarmsNumber(
		ctx, mozuID, controllerIDs,
		beginTime, endTime,
		func(tx *gorm.DB, total int64) error {
			if total > consts.MaxRecord {
				return fmt.Errorf("最多导出 %v 条记录，当前记录数: %v", consts.MaxRecord, total)
			}
			alarms, e = dac.GetAlarms(tx, mozuID, controllerIDs, condition.BeginTime, condition.EndTime)
			return e
		}); err != nil {
		return nil, err
	}

	return writeExcel(alarms)
}
