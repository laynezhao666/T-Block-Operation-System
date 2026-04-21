// Package event 提供门禁事件记录的查询和导出功能。
package event

import (
	"context"
	"dac/entity/consts"
	"dac/entity/model/cgi"
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/utils/excel"
	"dac/entity/utils/ttime"
	"dac/repo/dac"
	"fmt"
	"github.com/tealeg/xlsx/v3"
	"gorm.io/gorm"
	"trpc.group/trpc-go/trpc-go/log"
)

// ht 默认Excel行高
const (
	ht = 14.0
)

// titles 事件导出Excel表头
var (
	titles = []string{"时间", "事件描述", "进出方向", "门名称", "门编号", "门禁控制器名称", "门禁控制器ID", "卡号", "人员", "单位"}
)

// getDirection 将进出方向码转换为中文描述
func getDirection(d int, t int) string {
	if t == int(driver.EventTypeRemoteOpen) {
		return "未知"
	}
	if d == int(driver.DirectionEnter) {
		return "进门"
	}
	return "出门"
}

// getCGIEvents 将数据库事件记录转换为CGI事件结构体
func getCGIEvents(es []db.Event) []cgi.Event {
	events := make([]cgi.Event, 0, len(es))
	for i := range es {
		e := &es[i]

		cgiE := cgi.Event{
			ControllerID:   e.ControllerID,
			ControllerName: e.ControllerName,
			Index:          e.Index,
			Time:           ttime.Format(e.Timestamp),
			CardNumber:     e.CardNumber,
			Username:       e.Username,
			DoorNumber:     e.DoorNumber,
			DoorName:       e.DoorName,
			Company:        e.Company,
			Direction:      getDirection(e.Direction, e.Type),
			Type:           e.Type,
			Description:    e.Description,
		}

		events = append(events, cgiE)
	}
	return events
}

// GetByDoors 按门查询事件记录
func GetByDoors(ctx context.Context,
	controllerDoors map[int][]int, condition cgi.QueryCondition,
) (int64, []cgi.Event, error) {
	n, es, err := dac.GetRW().GetEventsByDoors(
		ctx, controllerDoors,
		condition.Offset, condition.Limit,
		condition.BeginTime, condition.EndTime, nil,
	)
	if err != nil {
		return 0, nil, err
	}

	return n, getCGIEvents(es), nil
}

// Get 分页查询事件记录，支持按控制器、关键字、门名称过滤
func Get(ctx context.Context, mozuID string,
	controllerIDs []int, query string, doorName string,
	condition cgi.QueryCondition,
) (int64, []cgi.Event, error) {
	n, es, err := dac.GetRW().GetEvents(
		ctx, mozuID, controllerIDs, query, doorName,
		condition.Offset, condition.Limit,
		condition.BeginTime, condition.EndTime, nil,
	)
	if err != nil {
		return 0, nil, err
	}

	return n, getCGIEvents(es), nil
}

// writeExcel 将事件记录写入Excel文件
func writeExcel(events []db.Event) (*xlsx.File, error) {
	f := xlsx.NewFile()
	s, err := f.AddSheet("进出记录")
	if err != nil {
		return nil, err
	}

	if _, err = excel.AddStringRow(s, ht, titles...); err != nil {
		return nil, err
	}

	for i := range events {
		e := &events[i]
		if _, err = excel.AddRow(
			s, ht, ttime.Format(e.Timestamp),
			e.Description, getDirection(e.Direction, e.Type),
			e.DoorName, e.DoorNumber,
			e.ControllerName, e.ControllerID,
			e.CardNumber, e.Username, e.Company,
		); err != nil {
			return nil, err
		}
	}

	return f, nil
}

// Export 导出事件记录到Excel文件，限制最大导出条数
func Export(ctx context.Context, mozuID string, doorName string,
	controllerDoors map[int][]int, condition cgi.TimeCondition,
	query string,
) (*xlsx.File, error) {
	var (
		err    error
		events []db.Event
		e      error
	)

	beginTime := condition.BeginTime
	endTime := condition.EndTime
	if _, err = dac.GetRW().GetEventsNumber(
		ctx, mozuID, doorName, controllerDoors,
		beginTime, endTime, query,
		func(tx *gorm.DB, total int64) error {
			if total > consts.MaxRecord {
				return fmt.Errorf(
					"最多导出 %v 条记录，当前记录数: %v",
					consts.MaxRecord, total,
				)
			}
			log.Infof("查询到该门：%s的刷卡记录共有：%s条。",
				doorName, total)
			events, e = dac.GetEvents(
				tx, mozuID, doorName, controllerDoors,
				beginTime, endTime, query,
			)
			log.Infof("导出到该门：%s的刷卡记录共有：%s条。",
				doorName, len(events))
			return e
		}); err != nil {
		return nil, err
	}

	return writeExcel(events)
}
