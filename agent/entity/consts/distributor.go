package consts

// Distributor 配置 key
const (
	DistKeyTlink       = "tlink"
	DistKeyBypass      = "bypass"
	DistKeyDeviceModel = "deviceModel"
)

// Distributor Cfg 字段名
const (
	DistCfgTarget         = "target"
	DistCfgClientID       = "client_id"
	DistCfgBroker         = "broker"
	DistCfgQos            = "qos"
	DistCfgRetain         = "retain"
	DistCfgTimeoutConnect = "timeout_connect"
	DistCfgTimeoutRW      = "timeout_rw"
)

// Distributor 默认值
const (
	DistDefaultClientID     = "tbos_agent"
	DistDefaultBypassTarget = "http://127.0.0.1:13000/api/dcos/tiot/tbos"
)

// DeviceModel 默认配置
const (
	DistDefaultDeviceModelBroker         = "ws://127.0.0.1:9001"
	DistDefaultDeviceModelClientID       = "client_tbos"
	DistDefaultDeviceModelQos            = "1"
	DistDefaultDeviceModelRetain         = "true"
	DistDefaultDeviceModelTimeoutConnect = "2"
	DistDefaultDeviceModelTimeoutRW      = "2"
)
