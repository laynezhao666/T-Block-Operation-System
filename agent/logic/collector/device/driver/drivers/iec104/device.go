package iec104

import (
	"context"
	"fmt"
	"agent/entity/consts"
	"agent/logic/collector/device/model"
	"agent/utils"
	"agent/utils/osal"
	"sync"

	"trpc.group/trpc-go/trpc-go/log"

	model1 "agent/entity/model"

	"git.woa.com/andromeda/tengreen/iec104"
	"git.woa.com/andromeda/tengreen/tengreen/common/tool"
)

const DefaultTotalCalInterval = 5

// Device 设备
type Device struct {
	client *iec104.IEC104Client // iec 客户端
	cache  *sync.Map            // 数据缓存
}

// NewIEC104Device 创建iec104设备
func NewIEC104Device() *Device {
	d := &Device{
		client: iec104.NewIEC104Client(),
		cache:  new(sync.Map),
	}
	err := d.client.SetTimeoutTotalCall(DefaultTotalCalInterval)
	if err != nil {
		return nil
	}
	d.client.SetLogger(log.GetDefaultLogger())
	return d
}

// Open 建立连接
func (d *Device) Open(chanInfo model.ChannelInfo, _ model.ListCollectPackets) consts.Quality {
	d.client.SetAddress(chanInfo.Name, "")

	if err := d.client.Connect(context.Background()); err != nil {
		log.Errorf("iec104 driver connect error:%s", err.Error())
		return consts.QualityCannotOpen
	}

	d.client.SetDeal(d.deal) // 数据处理
	return consts.QualityOk
}

// Close 关闭连接
func (d *Device) Close() consts.Quality {
	err := d.client.Close()
	if err != nil {
		return consts.QualityCmdRespError
	}
	return consts.QualityOk
}

// Request Request
func (d *Device) Request(ctx context.Context, packet *model.CollectProtocolPacket) (consts.Quality, model1.MessageStatistics) {
	statics := model1.MessageStatistics{SendCount: 0, SuccessCount: 0}
	if packet == nil {
		return consts.QualityOk, statics
	}

	if !d.client.IsConnected() {
		return consts.QualityCommDisconnected, statics
	}

	if packet == nil {
		return consts.QualityOk, statics
	}
	statics.SendCount = uint64(len(packet.Points))

	for _, p := range packet.Points {
		valueParser, ok := p.Attr.ValParser.(*ValueParser)
		if !ok || valueParser == nil {
			return consts.QualityUncertain, statics
		}

		sTmp, ok := d.cache.Load(valueParser.Addr)
		if !ok {
			p.RtVal.Qua = consts.QualityUncollected
			continue
		}
		s, ok := sTmp.(*iec104.Signal)
		if !ok {
			p.RtVal.Qua = consts.QualityCmdRespError
			continue
		}
		if err := parseVariable(s, valueParser, &p.RtVal.Pv); err != nil {
			log.Errorf("iec104 parse variable error:%s", err.Error())
			p.RtVal.Qua = consts.QualityCmdRespError
			continue
		}
		p.RtVal.Qua = parseQua(s)
		p.RtVal.Tms = utils.GetNowUTCTimeStamp()
		statics.SuccessCount++
	}
	return consts.QualityOk, statics
}

// RequestPing 发送采集指令，最小化指令发送包
func (d *Device) RequestPing(ctx context.Context, packet model.CollectProtocolPacket) consts.Quality {
	qua, _ := d.Request(ctx, &packet)
	return qua
}

// Control 控制
func (d *Device) Control(packet *model.ControlProtocolPacket, val string) consts.Quality {
	return consts.QualityOk
}

// deal 数据处理任务
func (d *Device) deal(data *iec104.APDU) {
	for _, s := range data.Signals {
		d.cache.Store(s.Address, s)
	}
}

func parseVariable(s *iec104.Signal, valueParser *ValueParser, variant *osal.Variant) error {
	switch s.TypeID {
	case iec104.MSpNa1:
		v, ok := s.Value.(byte)
		if !ok {
			return fmt.Errorf("convert to byte error, pdu: %s, signal: %+v", tool.Strval(s.Value), s)
		}

		if invalid := v & 0x80; invalid > 0 {
			return fmt.Errorf("invalid byte, pdu: %s, signal: %+v", tool.Strval(s.Value), s)
		}

		value := 0
		if (v & 1) > 0 {
			value = 1
		}
		variant.SetValue(value)
	case iec104.MMeNc1, iec104.MItNa1:
		variant.SetValue(s.Value)
	default:
		return fmt.Errorf("not supported iec104 type, type=%v", s.TypeID)
	}
	return nil
}

func parseQua(s *iec104.Signal) consts.Quality {
	return consts.QualityOk
}
