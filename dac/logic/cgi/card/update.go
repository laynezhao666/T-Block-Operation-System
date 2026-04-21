// Package card 提供门禁卡的CGI业务逻辑层，封装卡的增删改查操作。
package card

import (
	"context"

	"dac/entity/model/db"
	"dac/logic/card"
)

// UpdateCardsFlag 批量更新门禁卡启用/禁用状态
func UpdateCardsFlag(ctx context.Context, cards []string, mozuID string, flag int) error {
	return card.UpdateFlag(ctx, cards, mozuID, db.CardFlagType(flag))
}

// UpdateCardsType 批量更新门禁卡类型（长期卡/临时卡）
func UpdateCardsType(ctx context.Context, cards []string, mozuID string, cardType int) error {
	return card.UpdateType(ctx, cards, mozuID, db.CardType(cardType))
}

// UpdateCardValidTime 批量更新门禁卡有效期
func UpdateCardValidTime(ctx context.Context, cardsNo []string, mozuID string, validTime int64) error {
	return card.UpdateValidTime(ctx, cardsNo, mozuID, validTime)
}

// UpdateCardsAccessGroups 批量更新门禁卡所属权限组
func UpdateCardsAccessGroups(ctx context.Context, cards []string, groups []int, mozuID string) error {
	return card.UpdateToAccessGroups(ctx, cards, groups, mozuID)
}

// UpdateCardStaff 更新门禁卡绑定的人员
func UpdateCardStaff(ctx context.Context, cardNumber string, staffID int, mozuID string) error {
	return card.UpdateStaff(ctx, []string{cardNumber}, staffID, mozuID)
}
