package cgi

import (
	"errors"
	"agent/entity/consts"
	"agent/entity/definition"
	rtdbModel "agent/logic/collector/rtdb/model"
	"agent/utils"
	"agent/utils/osal"
	pb "trpcprotocol/agent"
)

// OnlinePushPointHandle 接收测点推送
func OnlinePushPointHandle(req *pb.OnlineStrategyPushReq) error {
	value := make(map[string]rtdbModel.RTValue)
	for _, device := range req.Objs {
		for _, point := range device.Points {
			ID := device.Guid + consts.DefaultIDSep + point.Tag
			rtValue := rtdbModel.RTValue{
				Pv:  osal.NewVariantWithValue(point.Value),
				Qua: consts.Quality(point.Qua),
				Tms: point.Timestamp,
			}
			value[ID] = rtValue
		}

		// 添加设备级别的通讯状态
		deviceCommKey := device.Guid + consts.DefaultIDSep + definition.CommID
		value[deviceCommKey] = rtdbModel.RTValue{
			Pv:  osal.NewVariantWithValue(device.Status),
			Qua: consts.QualityOk,
			Tms: utils.GetNowUTCTimeStamp(),
		}
	}

	if len(value) == 0 {
		return errors.New("push point empty")
	}

	//httpsub.Cache().SetPointsValue(value)
	return nil
}
