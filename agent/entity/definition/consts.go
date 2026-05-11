package definition

import (
	"time"
	"unsafe"
)

const (
	// FloatTypeSize 浮点类型大小
	FloatTypeSize = int(unsafe.Sizeof(FloatType(0)))
)

const (
	KafkaMaxBatchBytes = 100 * (1 << 20)

	TemplateCacheDataDuration = 5 * time.Minute // 模板缓存刷新时间

	DefaultInterval = 60 // 推送数据的默认间隔
	ChangeInterval  = 1

	LogInterval = time.Hour

	DeviceRefreshTime = DefaultInterval * time.Second
	DeviceQueryTime   = time.Second

	StdPointsRefreshTime = DefaultInterval * time.Second

	PointNumberPerMessage = 5000

	OfflineValue = "-99998"

	StdDevice     = "std"
	CollectDevice = "collect"
	AllPoints     = "all"

	TaskModeSchedule = "schedule"
	TaskModeLocal    = "local"
	TaskModeTLink    = "tlink"

	CollectorDeviceTypeTBox      = 1
	CollectorDeviceTypeTBoxSub   = 2
	CollectorDeviceTypeVendor    = 3
	CollectorDeviceTypeVendorSub = 4
	CollectorDeviceTypeTOne      = 7

	ChannelTypeSerial = "serial"

	KafkaDataTypeCollector = 1
	KafkaDataTypeStd       = 2
	KafkaDataTypeVirtual   = 3
)
