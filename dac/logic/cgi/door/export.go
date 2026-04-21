// Package door 提供门的查询、更新、状态控制和导入导出功能。
package door

import (
	"context"
	"encoding/json"

	"dac/entity/model/rt"
	"dac/repo/dac"

	"dac/entity/utils/excel"
	"github.com/tealeg/xlsx/v3"
)

// ht 默认Excel行高
const (
	ht = 14.0
)

// titles 门导出Excel表头
var (
	titles = []string{"门名称", "门编号", "门禁控制器名称", "门禁控制器 IP", "IDCDB 编号", "扩展属性"}
)

// writeExcel 将门控制器和门数据写入Excel文件
func writeExcel(controllers []rt.DoorController) (*xlsx.File, error) {
	f := xlsx.NewFile()

	s, err := f.AddSheet("门")
	if err != nil {
		return nil, err
	}

	if _, err = excel.AddStringRow(s, ht, titles...); err != nil {
		return nil, err
	}

	for i := range controllers {
		c := &controllers[i]
		doors := controllers[i].Doors

		for j := range doors {
			d := &doors[j]

			extend := "{}"
			if len(d.Extend) > 0 {
				var temp []byte
				if temp, err = json.Marshal(d.Extend); err != nil {
					return nil, err
				}
				extend = string(temp)
			}

			if _, err = excel.AddRow(s, ht, d.GetName(), d.Number, c.Name, c.Channel.ID, d.GetIDCDBCode(), extend); err != nil {
				return nil, err
			}
		}
	}

	return f, nil
}

// Export 导出模组下所有门信息到Excel
func Export(ctx context.Context, mozuID string) (*xlsx.File, error) {
	controllers, doors, err := dac.GetRW().GetAllDoorControllersAndDoors(ctx, mozuID)
	if err != nil {
		return nil, err
	}

	cs := make([]rt.DoorController, len(controllers))
	for i := range controllers {
		cs[i].DoorController = controllers[i]
		cs[i].Doors = doors[cs[i].ID]
	}

	return writeExcel(cs)
}
