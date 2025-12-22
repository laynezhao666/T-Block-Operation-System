package sysdio

import (
	"context"
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/device/driver/drivers/sysdio/controller"
	"agent/logic/collector/device/model"
	"agent/utils"
	"agent/utils/osal"

	"trpc.group/trpc-go/trpc-go/log"

	entityModel "agent/entity/model"
	rtdbModel "agent/logic/collector/rtdb/model"
	"strconv"
)

const DefaultTotalCalInterval = 5

// SysdioDev Sysdio device
type SysdioDev struct {
	Data           entityModel.IDeviceData
	isConnected    bool
	gpioController controller.IGPIOController
}

// NewSysdioDevice returns a new instance of SysdioDev
func NewSysdioDevice(gid definition.DeviceGidType, name string) *SysdioDev {
	return &SysdioDev{
		Data: entityModel.IDeviceData{
			Gid:  gid,
			Name: name,
		},
		isConnected: false,
	}
}

// Open opens the device
func (d *SysdioDev) Open(chanInfo model.ChannelInfo, packets model.ListCollectPackets) consts.Quality {
	d.gpioController = controller.GetController(chanInfo.ProtocolVer) // Assume GetController is defined elsewhere
	if d.gpioController == nil {
		return consts.QualityConfigError
	}
	return consts.QualityOk
}

// Close closes the device
func (d *SysdioDev) Close() consts.Quality {
	return consts.QualityOk
}

// Request queries the device
func (d *SysdioDev) Request(ctx context.Context, packet *model.CollectProtocolPacket) (consts.Quality,
	entityModel.MessageStatistics) {
	if packet == nil || d.gpioController == nil {
		return consts.QualityUncertain, entityModel.MessageStatistics{}
	}

	currentTime := utils.GetNowUTCTimeStamp()
	for _, point := range packet.Points {
		valParser := point.Attr.ValParser.(*SysdioValParser)
		d.ParseValue(valParser, &point.RtVal, currentTime)
	}
	return consts.QualityOk, entityModel.MessageStatistics{}
}

// RequestPing queries the device
func (d *SysdioDev) RequestPing(ctx context.Context, packet model.CollectProtocolPacket) consts.Quality {
	qua, _ := d.Request(ctx, &packet)
	return qua
}

// Control controls the device
func (d *SysdioDev) Control(packet *model.ControlProtocolPacket, val string) consts.Quality {
	if packet == nil || d.gpioController == nil {
		return consts.QualityUncertain
	}
	point := packet.Point
	if point == nil {
		return consts.QualityUncertain
	}
	valParser := point.Attr.ValParser.(*SysdioValParser) // Assume SysdioValParser is defined elsewhere
	if valParser == nil {
		return consts.QualityUncertain
	}

	value, err := strconv.Atoi(val)
	if err != nil {
		log.Errorf("control %s error, setting value is: %s", valParser.Pin, val)
		return consts.QualityConfigError
	}

	if w := d.gpioController.Write(valParser.Pin, value); w < 0 {
		log.Errorf("write %s error, value: %s", valParser.Pin, val)
		return consts.QualityCmdRespError
	}

	return consts.QualityOk
}

// ParseValue parses the value
func (d *SysdioDev) ParseValue(valParser *SysdioValParser, value *rtdbModel.RTValue, currentTime int64) {
	r := d.gpioController.Read(valParser.Pin)
	if r < 0 {
		value.Qua = consts.QualityValueAbnormal
		value.Pv = osal.NewVariant()
		return
	}

	val := utils.Unary(valParser.UnaryFun, r)

	value.Pv = osal.NewVariantWithValue(val)
	value.Qua = consts.QualityOk
	value.Tms = currentTime
}
