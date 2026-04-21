package xbrother

import (
	"context"
	"fmt"

	"dac/entity/model/driver"
	"dac/entity/model/driver/xbrother"
	"dac/entity/utils"
	"dac/logic/collect/driver/xbrother/consts"
	"dac/repo/dac"
)

// SetDoorParameter 批量设置门参数（逐门下发到门控器并保存到数据库）
func (c *Controller) SetDoorParameter(params []driver.DoorParameter) error {
	for i := range params {
		if err := c.setDoorParameter(params[i]); err != nil {
			c.logger.Errorf("params[%d] error, err: %s", i, err.Error())
			return err
		}
	}
	return nil
}

// GetDoorParameter 获取所有门参数（从数据库读取并转换为驱动模型）
func (c *Controller) GetDoorParameter() ([]driver.DoorParameter, error) {
	res, err := dac.GetRW().GetDriverDoorParameters(context.Background(), c.baseInfo.ID, c.chanInfo.ChannelID)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("doors is empty")
	}
	return utils.ConvertDBDriverDoorParams(res), nil
}

// setDoorParameter 设置单门参数（下发到门控器并保存到数据库）
func (c *Controller) setDoorParameter(param driver.DoorParameter) error {
	doorNo := uint8(param.Number)
	req := xbrother.SetDoorParamsReq{
		OpenDoorTime:        uint16(param.KeepOpenTime),
		OpenDoorTimeout:     uint8(param.OpenTimeout),
		BidirectionalDetect: consts.DefaultDoorParamBidirectionalDetect, // 协议规定默认1
		LongTimeOpenAlarm:   consts.DefaultDoorParamLongTimeOpenAlarm,
		AlarmType:           consts.OpenDoorTimeoutAlarm, // 当前只关注超时报警
		AlarmTime:           consts.DefaultAlarmTime,
	}
	if _, err := c.setDoorParams(req, doorNo); err != nil {
		return err
	}

	resDoorParameter := utils.ConvertDriverDoorParamToDB(c.baseInfo.ID, c.chanInfo.ChannelID, param)
	return dac.GetRW().AddDriverDoorParameter(context.Background(), c.baseInfo.ID, c.chanInfo.ChannelID, resDoorParameter)
}

// setDoorParams 向门控器发送设置门参数的协议请求
func (c *Controller) setDoorParams(req xbrother.SetDoorParamsReq, doorNo uint8) (xbrother.CommonResp, error) {
	return c.sendRequest(req, doorNo, consts.GetRRPCSetDoorParams(c.chanInfo.ChannelID), consts.CommandSetDoorParams)
}
