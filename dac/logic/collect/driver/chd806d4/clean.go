// Package chd806d4 实现CHD806D4门禁控制器协议的驱动层。
package chd806d4

import "fmt"

// ============ 系统操作 ============

// Clean CHD806D4协议不支持清空操作
func (c *Controller) Clean() error {
	return fmt.Errorf("chd806d4协议不支持clean清空操作")
}
