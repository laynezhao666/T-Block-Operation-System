// Package card 提供门禁卡的CGI业务逻辑层，封装卡的增删改查操作。
package card

import (
	"context"

	"dac/logic/card"
)

// DeleteCards 批量删除门禁卡
func DeleteCards(ctx context.Context, cards []string, mozuID string) error {
	return card.Delete(ctx, cards, mozuID)
}
