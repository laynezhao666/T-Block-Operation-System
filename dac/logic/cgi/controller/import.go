// Package controller 提供门禁控制器的增删改查、导入导出和远程控制功能。
package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"

	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/model/db"
	"dac/entity/model/rt"

	"dac/entity/model/tbox"
	"dac/entity/utils/excel"
	"dac/entity/utils/set"
	"github.com/tealeg/xlsx/v3"
)

// sheetName 控制器导入Excel的Sheet名称
const (
	sheetName = "门禁列表"
)

// convert 转换为不带启用字段的rt.DoorController结构体
func convert(cs []rt.DoorControllerItemWithEnable) ([]rt.DoorController, error) {
	results := make([]rt.DoorController, 0, len(cs))

	for i := range cs {
		c := &cs[i]

		cc := db.DoorController{
			Name: c.Name,
			Profile: db.Profile{
				Vendor: c.Vendor,
				Model:  c.Model,
				SN:     c.SN,
			},
			Position: tbox.Position{
				Room:  c.Room,
				Block: c.Block,
				No:    c.No,
			},
			Channel: tbox.ChannelRaw{
				ID:              c.ChannelID,
				RequestTimeout:  c.Timeout,
				CommandInterval: c.CommandInterval,
			},
			Protocol: db.Protocol{
				Name:    c.ProtocolName,
				Version: c.ProtocolVersion,
			},
			Extend:   make(map[string]interface{}),
			Account:  c.Account,
			Password: c.Password,
		}

		// FetchInterval 即使没有从excel中读取到，也有初值(int)
		// 事件告警同步时间间隔json解析
		if len(c.Extend) > 0 {
			var fi rt.FetchInterval
			if err := json.Unmarshal([]byte(c.Extend), &fi); err != nil {
				config.Log.Warnf("Error unmarshalling JSON: %v", err)
				continue
			}

			cc.Extend[consts.KeyFetchEventInterval] = fi.FetchEventInterval
			cc.Extend[consts.KeyFetchLoopEventInterval] = fi.FetchLoopEventInterval
			cc.Extend[consts.KeyFetchAlarmInterval] = fi.FetchAlarmInterval
			cc.Extend[consts.KeyFetchLoopAlarmInterval] = fi.FetchLoopAlarmInterval

			var temp map[string]interface{}
			if err := json.Unmarshal([]byte(c.Extend), &temp); err != nil {
				config.Log.Warnf("Error unmarshalling JSON: %v", err)
				continue
			}
			for k, v := range temp {
				cc.Extend[k] = v
			}
		}

		if len(c.DoorNum) > 0 {
			doorNum, err := strconv.Atoi(c.DoorNum)
			if err != nil {
				return nil, fmt.Errorf("parse door number error: %w", err)
			}
			cc.Extend[consts.KeyDoorNum] = doorNum
		}

		// 解析 URL 模式
		if len(c.URLMode) > 0 {
			// Excel 中填写 "默认" 或 "特殊"，存储为 "0" 或 "1"
			switch c.URLMode {
			case "特殊":
				cc.Extend[consts.KeyURLMode] = "1"
			case "默认":
				cc.Extend[consts.KeyURLMode] = "0"
			default:
				// 兼容直接填写 0/1 的情况
				cc.Extend[consts.KeyURLMode] = c.URLMode
			}
		}

		results = append(results, rt.DoorController{
			DoorController: cc,
			Doors:          make([]db.Door, 0),
		})
	}

	return results, nil
}

// verifyController 校验单个控制器数据的有效性
func verifyController(c *rt.DoorControllerItemWithEnable) error {
	if len(c.ChannelID) == 0 {
		return fmt.Errorf("%+v channel id is empty", *c)
	}
	return nil
}

// verifyControllersItem 校验控制器列表中是否有重复的通道ID
func verifyControllersItem(cs []rt.DoorControllerItemWithEnable) error {
	cache := set.NewStringSet()
	for i := range cs {
		c := &cs[i]
		if !cache.AddWithCheck(c.ChannelID) {
			return fmt.Errorf("channel %v 重复", c.ChannelID)
		}
	}
	return nil
}

// parseSheet 解析excel中的门禁控制器数据
func parseSheet(file *xlsx.File) ([]rt.DoorController, error) {
	s := file.Sheet[sheetName]
	if s == nil {
		return nil, fmt.Errorf("not found sheet %v", sheetName)
	}

	var (
		c   rt.DoorControllerItemWithEnable
		err error
	)
	cs := make([]rt.DoorControllerItemWithEnable, 0, s.MaxRow)
	if err = excel.ForEachRow(s, 1, &c, func(r *xlsx.Row, index int) error {
		if strings.Index(c.Enable, enableString) < 0 {
			return nil
		}
		if err = verifyController(&c); err != nil {
			return err
		}
		cs = append(cs, c)
		return nil
	}); err != nil {
		return nil, err
	}

	if err = verifyControllersItem(cs); err != nil {
		return nil, err
	}

	return convert(cs)
}

// Import 从Excel文件导入门禁控制器数据
func Import(ctx context.Context, mozuID string, file *multipart.FileHeader) error {
	xf, err := excel.OpenFile(file)
	if err != nil {
		return err
	}

	records, err := parseSheet(xf)
	if err != nil {
		return err
	}
	return BatchCreate(ctx, mozuID, records, true)
}
