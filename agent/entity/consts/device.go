package consts

const (
	DeviceIDVirtual = "0"
	DeviceIDDelta   = "1"
)

// DeviceIDEDC 采集器自身监控虚拟设备 ID
const DeviceIDEDC = "EDC_1"

const (
	ValueBigEndian    = "big"
	ValueLittleEndian = "little"
	ValueBigSwap      = "big-swap"
	ValueLittleSwap   = "little-swap"
)

const (
	LittleEndian     = "LittleEndian"
	BigEndian        = "BigEndian"
	LittleEndianSwap = "LittleEndianSwap"
	BigEndianSwap    = "BigEndianSwap"
)

const (
	DefaultSerialDeviceCmdIntervalMs = 100
	DefaultNetDeviceCmdIntervalMs    = 10
	DefaultChannelDeviceWaitMs       = 100
)

const (
	ChannelLogDir = "/tmp/tbox/log/channel"
)
