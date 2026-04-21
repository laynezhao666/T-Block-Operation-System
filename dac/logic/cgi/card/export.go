// Package card 提供门禁卡的CGI业务逻辑层，封装卡的增删改查操作。
package card

import (
	"context"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/entity/utils/excel"
	"dac/entity/utils/ttime"
	"dac/repo/dac"
	"fmt"
	"github.com/tealeg/xlsx/v3"
	"strings"
)

// TypeLongTerm 长期卡类型标签
// TypeTemporary 临时卡类型标签
// FlagEnable 启用状态标签
// FlagDisable 禁用状态标签
const (
	ht            = 14.0
	TypeLongTerm  = "长期卡"
	TypeTemporary = "临时卡"
	FlagEnable    = "启用"
	FlagDisable   = "禁用"
)

// titles 卡导出Excel表头
var (
	titles = []string{"卡号", "卡类型", "卡状态", "有效期", "人员名称", "人员编号", "权限组", "模组ID"}
)

// writeExcel 将人员信息数据写入excel
func writeExcel(cards []db.Card, staffMap map[db.IDType]db.Staff, cardGroupMap map[string][]db.IDType, groupNameMap map[db.IDType]string) (*xlsx.File, error) {
	f := xlsx.NewFile()
	s, err := f.AddSheet(sheetName)
	if err != nil {
		return nil, err
	}

	if _, err = excel.AddStringRow(s, ht, titles...); err != nil {
		return nil, err
	}

	for i := range cards {
		r, err := excel.AddStringRow(s, ht)
		if err != nil {
			return nil, err
		}

		card := &cards[i]
		var validTime, cardType, cardFlag, staffName, access string

		if card.CardType == db.CardTypeLongTerm {
			cardType = TypeLongTerm
		} else {
			cardType = TypeTemporary
			validTime = ttime.Format(card.ValidTime)
		}

		if card.CardFlag == db.CardFlagEnable {
			cardFlag = FlagEnable
		} else {
			cardFlag = FlagDisable
		}

		if staff, ok := staffMap[card.StaffID]; ok {
			staffName = staff.Name
		}
		if groupIDs, ok := cardGroupMap[card.CardNo]; ok && len(groupIDs) > 0 {
			groupNames := make([]string, 0, len(groupIDs))
			for _, gid := range groupIDs {
				if name, exists := groupNameMap[gid]; exists {
					groupNames = append(groupNames, name)
				}
			}
			access = strings.Join(groupNames, ", ") // 使用逗号+空格分隔
		}

		if r.WriteStruct(&rt.CardItem{
			Number:    card.CardNo,
			Type:      cardType,
			Flag:      cardFlag,
			ValidTime: validTime,
			StaffName: staffName,
			StaffID:   fmt.Sprintf("%d", card.StaffID),
			Access:    access,
			MozuID:    card.MozuID,
		}, -1) < 0 {
			return nil, fmt.Errorf("write %+v error", *card)
		}
	}

	return f, nil
}

// Export 导出模组下所有门禁卡信息到Excel
func Export(ctx context.Context, mozuID string) (*xlsx.File, error) {
	cards, staffMap, cardGroupMap, groupNameMap, err := dac.GetRW().GetAllCardsWithStaffAndAccessGroup(ctx, mozuID)
	if err != nil {
		return nil, fmt.Errorf("查询当前模组%v的人员信息失败：%v", mozuID, err)
	}

	return writeExcel(cards, staffMap, cardGroupMap, groupNameMap)
}
