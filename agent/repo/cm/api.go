package cm

import (
	"agent/entity/model"
	"agent/repo/cm/localfile"
	"agent/repo/cm/taskserver"
	"agent/repo/cm/tlink"
	"sync"

	"agent/repo/cm/backup"
)

// 配置来源
const LocalFileConfigModName = "local"
const TaskServerModName = "task_server"
const TLinkModName = "tlink"
const BackupModName = "backup"

var chConfigChanged chan bool
var stdConfigChanged chan bool
var deviceConfigChanged chan bool
var (
	configVersionChangeChanLock sync.RWMutex
	configVersionChangeChan     chan *model.ConfigChangeEvent
)

// Reader 配置读取接口
type Reader interface {
	GetDevices(devices []string) ([]model.Device, []model.Device, map[string]any, error)
	GetTemplate(name string) (*model.TemplateData, error)
	GetTemplates(list []string) (map[string]any, error)
	GetStdData(configVersion map[string]*model.ConfigVersion, devices []string) (*model.StdData, error)
	GetCmdbVersion() (map[string]*model.ConfigVersion, error)
	GetStdDevice(map[string]bool, []string) (*model.StdDeviceData, error)
}

// NewReader 新建配置读取接口
func NewReader(name string) Reader {
	switch name {
	case LocalFileConfigModName:
		return localfile.NewReadImpl(chConfigChanged)
	case TaskServerModName:
		return taskserver.NewReadImpl(chConfigChanged)
	case TLinkModName:
		return tlink.NewReadImpl(chConfigChanged)
	case BackupModName:
		return backup.NewReadImpl(chConfigChanged)
	default:
		return localfile.NewReadImpl(chConfigChanged)
	}
}

// ConfigChangedChan 配置变更通知
func ConfigChangedChan() chan bool {
	return chConfigChanged
}

// StdConfigChangedChan 标准配置变更通知
func StdConfigChangedChan() chan bool {
	return stdConfigChanged
}

// DeviceConfigChangedChan 设备配置变更通知
func DeviceConfigChangedChan() chan bool {
	return deviceConfigChanged
}
func ConfigVersionChangedChan() chan *model.ConfigChangeEvent {
	return configVersionChangeChan
}

func init() {
	chConfigChanged = make(chan bool)
	stdConfigChanged = make(chan bool)
	deviceConfigChanged = make(chan bool)
	configVersionChangeChan = make(chan *model.ConfigChangeEvent)
}
