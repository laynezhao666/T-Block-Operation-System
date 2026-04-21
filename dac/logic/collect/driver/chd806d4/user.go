// Package chd806d4 实现CHD806D4门禁控制器协议的驱动层。
package chd806d4

import (
	"dac/entity/model/driver"
	"fmt"
)

// ============ 用户管理 ============

// AddUser 添加用户（CHD806D4协议不支持此操作）
func (c *Controller) AddUser(user driver.CardWithStaffInfo) error {
	return fmt.Errorf("chd806d4协议不支持添加用户操作")
}

// DeleteUser 删除用户（CHD806D4协议不支持此操作）
func (c *Controller) DeleteUser(user driver.UserID) error {
	return fmt.Errorf("chd806d4协议不支持删除卡用户操作")
}
