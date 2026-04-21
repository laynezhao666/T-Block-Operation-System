// Package xbrother 实现XBrother门禁控制器协议的驱动层。
package xbrother

import (
	"dac/entity/model/driver/xbrother"
	"dac/logic/collect/driver/xbrother/consts"
)

// Clean 清除控制器数据（恢复出厂设置）
func (c *Controller) Clean() error {
	_, err := c.clean(xbrother.CleanReq{}, 0)
	return err
}

// clean 发送清除数据命令到控制器
func (c *Controller) clean(req xbrother.CleanReq, doorNo uint8) (xbrother.CommonResp, error) {
	return c.sendRequest(req, doorNo, consts.GetRRPCClean(c.chanInfo.ChannelID), consts.CommandClean)
}
