package simulator

import (
	"context"
	"agent/entity/consts"
	"agent/logic/collector/device/model"
	rtdbModel "agent/logic/collector/rtdb/model"
	"agent/utils"
	"time"

	entityModel "agent/entity/model"
)

type simulatorDevice struct {
	data entityModel.IDeviceData
}

// Open 打开通道
func (s *simulatorDevice) Open(chanInfo model.ChannelInfo, _ model.ListCollectPackets) consts.Quality {
	return consts.QualityOk
}

// Control 控制测点
func (s *simulatorDevice) Control(packet *model.ControlProtocolPacket, val string) consts.Quality {
	return consts.QualityOk
}

// Close 关闭通道
func (s *simulatorDevice) Close() consts.Quality {
	return consts.QualityOk
}

// Request 查询测点
func (s *simulatorDevice) Request(ctx context.Context, packet *model.CollectProtocolPacket) (consts.Quality,
	entityModel.MessageStatistics) {
	time.Sleep(200 * time.Millisecond)

	if packet == nil {
		return consts.QualityOk, entityModel.MessageStatistics{}
	}

	t := utils.GetNowUTCTimeStamp()
	for _, p := range packet.Points {
		if p == nil {
			continue
		}

		vp, ok := p.Attr.ValParser.(*ValueParser)
		if !ok {
			continue
		}

		s.ParseValue(vp, string(p.Attr.ID), &p.RtVal, t)
	}

	return consts.QualityOk, entityModel.MessageStatistics{}
}

// RequestPing 查询测点
func (s *simulatorDevice) RequestPing(ctx context.Context, packet model.CollectProtocolPacket) consts.Quality {
	qua, _ := s.Request(ctx, &packet)
	return qua
}

// ParseValue 解析测点数据
func (s *simulatorDevice) ParseValue(parser *ValueParser, id string, value *rtdbModel.RTValue, currentTime int64) {
	value.Pv.SetType(parser.DataType)
	value.Pv.SetFloat64(parser.Generator.Generate(nil))
	value.Qua = consts.QualityOk
	value.Tms = currentTime
}
