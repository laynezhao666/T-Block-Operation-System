// Package card 提供门禁卡的CGI业务逻辑层，封装卡的增删改查操作。
package card

import (
	"context"

	"dac/repo/dac"
)

// UnbindCardStaff 解绑门禁卡与人员的关联关系
func UnbindCardStaff(ctx context.Context, card string, mozuID string) error {
	// 解绑人员时，只需要修改数据库，不影响下层门禁控制器
	return dac.GetRW().UnbindCards(ctx, []string{card}, mozuID)
}
