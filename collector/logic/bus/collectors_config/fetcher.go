// Package collectors_config 采集配置获取
package collectors_config

// IConfigFetcher 采集器配置获取器接口
type IConfigFetcher interface {
	// Name 获取器的名称
	Name() string
	// FetchCollectDevices 获取采集设备配置
	FetchCollectDevices(deviceNumbers []string) ([]byte, error)
	// FetchCollectTemplates 获取采集模板配置
	FetchCollectTemplates(templateNames []string) ([]byte, error)
	// FetchStdPoints 获取采集设备相关标准点配置
	FetchStdPoints(deviceNumbers []string) ([]byte, error)
	// FetchConfigModifyTime 获取采集设备配置修改时间
	FetchConfigModifyTime(deviceNumbers []string) ([]byte, error)
	// FetchStdDevices 获取采集设备对应标准设备的配置
	FetchStdDevices(collectDeviceNumbers []string) ([]byte, error)
}
