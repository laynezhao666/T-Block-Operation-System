// Package card 提供门禁卡的CGI业务逻辑层，封装卡的增删改查操作。
package card

import (
	"context"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/entity/utils/excel"
	"dac/entity/utils/ttime"
	"dac/logic/card"
	"dac/repo/dac"
	"fmt"
	"github.com/tealeg/xlsx/v3"
	"mime/multipart"
	"strconv"
	"strings"
)

// sheetName 门禁卡导入Excel的Sheet名称
const (
	sheetName = "门禁卡信息"
)

// convert 转换为数据库Card结构体并收集权限组信息
func convert(ctx context.Context, item rt.CardItem, mozuID string) (db.Card, []string, error) {
	// 处理卡号：删除空格和前导零
	cardNumber := strings.ReplaceAll(item.Number, " ", "")
	cardNumber = strings.TrimLeft(cardNumber, "0")
	if cardNumber == "" {
		return db.Card{}, nil, fmt.Errorf("卡号不能为空")
	}

	// 验证卡号是否为数字
	if _, err := strconv.ParseInt(cardNumber, 10, 64); err != nil {
		return db.Card{}, nil, fmt.Errorf("卡号 %s 不是有效的数字", item.Number)
	}

	// 解析卡类型和有效期
	cardType, validTime, err := parseTypeAndValidTime(item.Type, item.ValidTime)
	if err != nil {
		return db.Card{}, nil, err
	}

	// 解析卡状态
	cardFlag := parseCardFlag(item.Flag)

	// 解析人员ID
	staffID, err := parseStaffID(item.StaffName, item.StaffID, mozuID)

	// 解析权限组名称列表
	accessGroupNames := parseAccessGroups(item.Access)

	card := db.Card{
		CardNo:    cardNumber,
		CardType:  cardType,
		CardFlag:  cardFlag,
		ValidTime: validTime,
		StaffID:   staffID,
		MozuID:    mozuID,
	}

	return card, accessGroupNames, nil

}

// parseTypeAndValidTime 解析卡类型与有效时间
func parseTypeAndValidTime(cardType string, validTime string) (db.CardType, int64, error) {
	var (
		ct db.CardType
		vt int64
	)

	if cardType == TypeLongTerm {
		ct = db.CardTypeLongTerm
		vt = 0
	} else if cardType == TypeTemporary {
		t, err := ttime.Parse(validTime)
		if err != nil {
			return 0, 0, fmt.Errorf("无效的日期格式: %s", validTime)
		}
		ct = db.CardTypeTemporary
		vt = t.Unix()
	} else {
		return 0, 0, fmt.Errorf("无效的卡类型: %s", cardType)
	}

	return ct, vt, nil
}

// parseCardFlag 解析卡状态
func parseCardFlag(flagStr string) db.CardFlagType {
	flagStr = strings.TrimSpace(flagStr)
	if flagStr == FlagEnable {
		return db.CardFlagEnable
	} else {
		return db.CardFlagDisable
	}
}

// parseStaffID 解析人员ID
func parseStaffID(staffName string, staffID, mozuID string) (db.IDType, error) {
	staffID = strings.TrimSpace(staffID)
	staffName = strings.TrimSpace(staffName)

	// 情况1：人员编号为空，返回初始值
	if staffID == "" {
		return 0, nil
	}
	// 情况2：人员编号格式错误
	id, err := strconv.ParseInt(staffID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("员工编号格式错误: %s，必须是数字", staffID)
	}
	return db.IDType(id), nil
}

// parseAccessGroups 解析权限组
func parseAccessGroups(accessStr string) []string {
	accessStr = strings.TrimSpace(accessStr)
	if accessStr == "" {
		return []string{}
	}

	// 先按逗号分割（不管有没有空格）
	groups := strings.Split(accessStr, ",")

	// 过滤空字符串并去除每个权限组名称的首尾空格
	result := make([]string, 0, len(groups))
	for _, g := range groups {
		g = strings.TrimSpace(g)
		if g != "" {
			result = append(result, g)
		}
	}

	return result
}

// parseSheet 解析excel中的卡数据
func parseSheet(file *xlsx.File) ([]rt.CardItem, error) {
	return excel.ParseFirstSheet(file, func(item rt.CardItem) bool {
		return isEmptyCardItem(item)
	})
}

// isEmptyCardItem 判断门禁卡记录是否为空
func isEmptyCardItem(item rt.CardItem) bool {
	// 检查关键字段是否都为空
	return item.Number == "" || item.Type == "" || item.Flag == ""
}

// Import 从Excel文件批量导入门禁卡数据
func Import(ctx context.Context, mozuID string, file *multipart.FileHeader) error {
	xf, err := excel.OpenFile(file)
	if err != nil {
		return err
	}

	records, err := parseSheet(xf)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return fmt.Errorf("excel文件中没有有效的门禁卡数据")
	}

	// 转换为数据库结构体并收集权限组信息
	type cardWithAccess struct {
		card             db.Card
		accessGroupNames []string
	}
	cardItems := make([]cardWithAccess, 0, len(records))
	cardNoSet := make(map[string]int) // 用于检查Excel内部重复，值为行号

	for i, item := range records {
		card, accessGroupNames, err := convert(ctx, item, mozuID)
		if err != nil {
			return fmt.Errorf("第 %d 行数据转换失败: %w", i+2, err)
		}

		// 检查Excel内部是否有重复卡号
		if existRow, exists := cardNoSet[card.CardNo]; exists {
			return fmt.Errorf("第 %d 行: 卡号 %s 与第 %d 行重复", i+2, card.CardNo, existRow)
		}
		cardNoSet[card.CardNo] = i + 2

		cardItems = append(cardItems, cardWithAccess{
			card:             card,
			accessGroupNames: accessGroupNames,
		})
	}

	// 查询所有权限组，建立名称到ID的映射
	accessGroups, err := dac.GetRW().GetAllAccessGroups(ctx, mozuID)
	if err != nil {
		return fmt.Errorf("查询权限组失败: %w", err)
	}
	accessGroupMap := make(map[string]db.IDType)
	for _, ag := range accessGroups {
		accessGroupMap[ag.Name] = ag.ID
	}

	// 批量导入卡信息
	for i, item := range cardItems {
		// 转换权限组名称为ID
		accessGroupIDs := make([]int, 0, len(item.accessGroupNames))
		for _, name := range item.accessGroupNames {
			if name == "" {
				continue
			}
			if agID, ok := accessGroupMap[name]; ok {
				accessGroupIDs = append(accessGroupIDs, agID)
			} else {
				return fmt.Errorf("第 %d 行: 权限组 '%s' 不存在", i+2, name)
			}
		}

		// 调用添加卡的逻辑
		if err := card.AddCard(ctx, mozuID, item.card.CardNo, int(item.card.CardFlag),
			int(item.card.CardType), item.card.ValidTime, item.card.StaffID, accessGroupIDs); err != nil {
			return fmt.Errorf("第 %d 行: 添加卡 %s 失败: %w", i+2, item.card.CardNo, err)
		}
	}
	return nil
}
