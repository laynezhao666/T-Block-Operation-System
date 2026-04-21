// Package card 提供门禁卡的CGI业务逻辑层，封装卡的增删改查操作。
package card

import (
	"context"

	"dac/logic/card"
)

// AddCard 添加门禁卡，关联人员和权限组
func AddCard(ctx context.Context, mozuID, number string,
	flag, t int, validTime int64, staffID int, accessGroupIDs []int,
) error {
	return card.AddCard(
		ctx, mozuID, number, flag, t,
		validTime, staffID, accessGroupIDs,
	)
}
