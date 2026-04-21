// Package xbrother 实现XBrother门禁控制器协议的驱动层。
package xbrother

import (
	"dac/entity/model/driver"
	"fmt"
)

// AddUser XBrother协议不支持添加用户操作
func (c *Controller) AddUser(user driver.CardWithStaffInfo) error {
	return fmt.Errorf("xbrother协议不支持添加用户操作")
}

// DeleteUser XBrother协议不支持删除卡用户操作
func (c *Controller) DeleteUser(user driver.UserID) error {
	return fmt.Errorf("xbrother协议不支持删除卡用户操作")
}
