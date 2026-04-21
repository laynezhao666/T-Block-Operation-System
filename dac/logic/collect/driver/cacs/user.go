// Package cacs 实现CACS门禁控制器协议的驱动层。
package cacs

import (
	"dac/entity/model/driver"
	"fmt"
)

// AddUser CACS协议不支持添加用户操作
func (c *Controller) AddUser(user driver.CardWithStaffInfo) error {
	return fmt.Errorf("cacs协议不支持添加用户操作")
}

// DeleteUser CACS协议不支持删除卡用户操作
func (c *Controller) DeleteUser(user driver.UserID) error {
	return fmt.Errorf("cacs协议不支持删除卡用户操作")
}
