// Package controller 提供门禁控制器的增删改查、导入导出和远程控制功能。
package controller

import (
	"context"
	"fmt"

	"dac/entity/consts"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/repo/dac"

	"dac/entity/utils/excel"
	"github.com/tealeg/xlsx/v3"
)

// ht 默认Excel行高
const (
	ht = 14.0
)

// enableString 启用状态标签
// enableIndex 启用字段在Excel中的列索引
// commandIntervalIndex 命令间隔字段列索引
// doorNumIndex 门数量字段列索引
// urlModeIndex URL模式字段列索引
const (
	enableString         = "启用"
	enableIndex          = 13
	commandIntervalIndex = 14
	doorNumIndex         = 15
	urlModeIndex         = 16
)

// titles 控制器导出Excel表头
var (
	titles = []string{"名称", "厂商", "型号", "序列号", "房间", "方仓", "编号",
		"IP", "超时时间（毫秒）", "协议名称", "协议版本", "账号", "密码", "是否启用", "命令间隔（毫秒）",
		"门数量", "URL模式"}
)

// writeExcel 将rt.DoorController数据写入excel
func writeExcel(controllers []db.DoorController) (*xlsx.File, error) {
	f := xlsx.NewFile()
	s, err := f.AddSheet(sheetName)
	if err != nil {
		return nil, err
	}

	if _, err = excel.AddStringRow(s, ht, titles...); err != nil {
		return nil, err
	}

	for i := range controllers {
		r, err := excel.AddStringRow(s, ht)
		if err != nil {
			return nil, err
		}

		c := &controllers[i]

		// URL 模式：数据库存 0/1，导出为中文
		urlModeStr := ""
		if urlMode, ok := c.Extend[consts.KeyURLMode].(string); ok {
			switch urlMode {
			case "1":
				urlModeStr = "特殊"
			case "0":
				urlModeStr = "默认"
			default:
				urlModeStr = urlMode
			}
		}

		if r.WriteStruct(&rt.DoorControllerItemWithEnable{
			Name:            c.Name,
			Vendor:          c.Profile.Vendor,
			Model:           c.Profile.Model,
			SN:              c.Profile.SN,
			Room:            c.Position.Room,
			Block:           c.Position.Block,
			No:              c.Position.No,
			ChannelID:       c.Channel.ID,
			Timeout:         c.Channel.RequestTimeout,
			ProtocolName:    c.Protocol.Name,
			ProtocolVersion: c.Protocol.Version,
			Enable:          enableString,
			CommandInterval: c.Channel.CommandInterval,
			DoorNum:         fmt.Sprintf("%v", c.Extend[consts.KeyDoorNum]),
			URLMode:         urlModeStr,
		}, -1) < 0 {
			return nil, fmt.Errorf("write %+v error", *c)
		}
	}

	return f, nil
}

// Export 导出模组下所有控制器信息到Excel
func Export(ctx context.Context, mozuID string) (*xlsx.File, error) {
	controllers, err := dac.GetRW().GetAllDoorControllers(ctx, mozuID)
	if err != nil {
		return nil, err
	}

	return writeExcel(controllers)
}
