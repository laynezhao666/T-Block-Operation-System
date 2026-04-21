// Package cacs 实现CACS门禁控制器协议的驱动层。
package cacs

// Clean 清空门控器中的所有卡和时间组数据。
// 依次执行删除所有卡和删除所有时间组操作。
func (c *Controller) Clean() error {
	if _, err := c.checkConnection(); err != nil {
		return err
	}
	// 先删除所有卡信息
	if err := c.DeleteAllCards(); err != nil {
		return err
	}
	// 再删除所有时间组
	if err := c.DeleteAllTimeGroups(); err != nil {
		return err
	}
	return nil
}
